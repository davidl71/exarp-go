package tools

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/utils"
	"github.com/davidl71/exarp-go/proto"
	mcpframework "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	mcpresponse "github.com/davidl71/mcp-go-core/pkg/mcp/response"
	"github.com/spf13/cast"
)

// handleSessionNative handles the session tool with native Go implementation.
func handleSessionNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action := strings.TrimSpace(cast.ToString(params["action"]))
	if action == "" {
		action = "prime"
	}

	switch action {
	case "prime":
		return handleSessionPrime(ctx, params)
	case "handoff":
		return handleSessionHandoff(ctx, params)
	case "prompts":
		return handleSessionPrompts(ctx, params)
	case "assignee":
		return handleSessionAssignee(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s (use 'prime', 'handoff', 'prompts', or 'assignee')", action)
	}
}

// SessionPrimeResultToMap converts SessionPrimeResult proto to map for FormatResult (stable JSON shape).
func SessionPrimeResultToMap(pb *proto.SessionPrimeResult) map[string]interface{} {
	if pb == nil {
		return nil
	}

	out := map[string]interface{}{
		"auto_primed": pb.AutoPrimed,
		"method":      pb.Method,
		"timestamp":   pb.Timestamp,
		"duration_ms": pb.DurationMs,
		"hints_count": pb.HintsCount,
	}
	if pb.Detection != nil {
		out["detection"] = map[string]interface{}{
			"agent":        pb.Detection.Agent,
			"agent_source": pb.Detection.AgentSource,
			"mode":         pb.Detection.Mode,
			"mode_source":  pb.Detection.ModeSource,
			"time_of_day":  pb.Detection.TimeOfDay,
		}
	}

	if pb.AgentContext != nil {
		out["agent_context"] = map[string]interface{}{
			"focus_areas":      pb.AgentContext.FocusAreas,
			"relevant_tools":   pb.AgentContext.RelevantTools,
			"recommended_mode": pb.AgentContext.RecommendedMode,
		}
	}

	if pb.Workflow != nil {
		out["workflow"] = map[string]interface{}{
			"mode":        pb.Workflow.Mode,
			"description": pb.Workflow.Description,
		}
	}

	if pb.Elicitation != "" {
		out["elicitation"] = pb.Elicitation
	}

	if pb.LockCleanup != nil && pb.LockCleanup.Cleaned > 0 {
		out["lock_cleanup"] = map[string]interface{}{
			"cleaned":  int(pb.LockCleanup.Cleaned),
			"task_ids": pb.LockCleanup.TaskIds,
		}
	}

	if pb.PlanPath != "" {
		out["plan_path"] = pb.PlanPath
	}

	if pb.ActionRequired != "" {
		out["action_required"] = pb.ActionRequired
	}

	if len(pb.ConflictHints) > 0 {
		out["conflict_hints"] = pb.ConflictHints
	}

	if pb.StatusLabel != "" {
		out["status_label"] = pb.StatusLabel
	}

	if pb.StatusContext != "" {
		out["status_context"] = pb.StatusContext
	}

	if pb.CursorCliSuggestion != "" {
		out["cursor_cli_suggestion"] = pb.CursorCliSuggestion
	}

	return out
}

// handleSessionPrime handles the prime action - auto-prime AI context at session start.
func handleSessionPrime(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()

	includeHints := true
	if _, ok := params["include_hints"]; ok {
		includeHints = cast.ToBool(params["include_hints"])
	}

	includeTasks := true
	if _, ok := params["include_tasks"]; ok {
		includeTasks = cast.ToBool(params["include_tasks"])
	}

	// Optional MCP Elicitation: ask user for prime preferences when ask_preferences is true.
	// Use a short timeout so prime never blocks indefinitely if the client is slow or doesn't respond.
	const elicitationTimeout = 5 * time.Second

	var elicitationOutcome string

	if cast.ToBool(params["ask_preferences"]) {
		if eliciter := mcpframework.EliciterFromContext(ctx); eliciter != nil {
			schema := map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"include_tasks": map[string]interface{}{"type": "boolean", "description": "Include task summary"},
					"include_hints": map[string]interface{}{"type": "boolean", "description": "Include tool hints"},
				},
			}

			elicitCtx, cancel := context.WithTimeout(ctx, elicitationTimeout)
			defer cancel()

			action, content, err := eliciter.ElicitForm(elicitCtx, "Session prime: include task summary and tool hints?", schema)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || (elicitCtx.Err() != nil && errors.Is(elicitCtx.Err(), context.DeadlineExceeded)) {
					elicitationOutcome = "timeout"
				} else {
					elicitationOutcome = "error"
				}
			} else if action == "accept" && content != nil {
				if v, ok := content["include_tasks"].(bool); ok {
					includeTasks = v
				}

				if v, ok := content["include_hints"].(bool); ok {
					includeHints = v
				}

				elicitationOutcome = "ok"
			} else {
				elicitationOutcome = "declined"
			}
		}
	}

	overrideMode := cast.ToString(params["override_mode"])

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	// 0. Dead agent lock cleanup (T-76) - quick cleanup before loading tasks
	var lockCleanupReport map[string]interface{}

	if db, dbErr := database.GetDB(); dbErr == nil && db != nil {
		staleThreshold := config.GetGlobalConfig().Timeouts.StaleLockThreshold
		if staleThreshold <= 0 {
			staleThreshold = 5 * time.Minute
		}

		if cleaned, taskIDs, cleanupErr := database.CleanupDeadAgentLocks(ctx, staleThreshold); cleanupErr == nil && cleaned > 0 {
			lockCleanupReport = map[string]interface{}{
				"cleaned":  cleaned,
				"task_ids": taskIDs,
			}
		}
	}

	// 1. Detect agent type
	agentInfo := detectAgentType(projectRoot)
	agentContext := getAgentContext(agentInfo.Agent)

	// 2. Determine mode
	var mode string

	var modeSource string

	if overrideMode != "" {
		mode = overrideMode
		modeSource = "override"
	} else {
		timeSuggestion := suggestModeByTime()
		agentMode := agentContext.RecommendedMode

		// Prefer agent-specific mode in working hours, time-based otherwise
		if timeSuggestion.Mode == "daily_checkin" {
			mode = timeSuggestion.Mode
			modeSource = "time_of_day"
		} else {
			mode = agentMode
			modeSource = "agent_type"
		}
	}

	// 3. Load tasks when needed (summary, suggested_next, or plan mode context)
	var tasks []Todo2Task

	var tasksErr error

	if includeTasks || includeHints {
		store := NewDefaultTaskStore(projectRoot)

		list, err := store.ListTasks(ctx, nil)
		if err != nil {
			tasksErr = err
		} else {
			tasks = tasksFromPtrs(list)
		}
	}

	// 4. Hints and plan path (needed for proto hints_count and plan_path)
	var planPath string

	hints := make(map[string]string)
	if includeHints {
		hints = getHintsForMode(mode)

		planPath, planModeHint := getPlanModeContext(projectRoot, tasks)
		if planPath != "" {
			// set below in proto
		}

		if planModeHint != "" {
			hints["plan_mode"] = planModeHint
		}

		todoCount := 0
		for _, t := range tasks {
			if t.Status == "Todo" {
				todoCount++
			}
		}
		if todoCount > 10 {
			hints["thinking_workflow"] = "For complex backlog analysis, sprint planning, or dependency enrichment: use the thinking-workflow skill (.cursor/skills/thinking-workflow/SKILL.md) — chain tractatus (structure) + sequential (process) + exarp-go MCP (execute)"
		}
	} else if includeTasks {
		planPath, _ = getPlanModeContext(projectRoot, tasks)
	}

	handoffAlert := (map[string]interface{})(nil)
	if _, has := params["include_handoff"]; !has || cast.ToBool(params["include_handoff"]) {
		handoffAlert = checkHandoffAlert(projectRoot)
	}

	actionRequired := ""
	if handoffAlert != nil {
		actionRequired = "📋 Review handoff from previous developer before starting work"
	}

	var conflictHints []string

	if taskOverlaps, fileConflicts, err := DetectConflicts(ctx, projectRoot); err == nil {
		for _, c := range taskOverlaps {
			conflictHints = append(conflictHints, "Task overlap: "+c.Reason)
		}

		for _, c := range fileConflicts {
			conflictHints = append(conflictHints, "File conflict: tasks "+strings.Join(c.TaskIDs, ", ")+" share file(s): "+strings.Join(c.Files, ", "))
		}
	}

	// 5. Build type-safe proto for prime result
	pb := &proto.SessionPrimeResult{
		AutoPrimed:     true,
		Method:         "native_go",
		Timestamp:      time.Now().Format(time.RFC3339),
		DurationMs:     time.Since(startTime).Milliseconds(),
		Detection:      &proto.SessionDetection{Agent: agentInfo.Agent, AgentSource: agentInfo.Source, Mode: mode, ModeSource: modeSource, TimeOfDay: time.Now().Format("15:04")},
		AgentContext:   &proto.SessionAgentContext{FocusAreas: agentContext.FocusAreas, RelevantTools: agentContext.RelevantTools, RecommendedMode: agentContext.RecommendedMode},
		Workflow:       &proto.SessionWorkflow{Mode: mode, Description: getWorkflowModeDescription(mode)},
		PlanPath:       planPath,
		HintsCount:     int32(len(hints)),
		ActionRequired: actionRequired,
		ConflictHints:  conflictHints,
	}
	if elicitationOutcome != "" {
		pb.Elicitation = elicitationOutcome
	}

	if lockCleanupReport != nil {
		if cleaned, ok := lockCleanupReport["cleaned"].(int); ok {
			var taskIDs []string
			if ids, ok := lockCleanupReport["task_ids"].([]string); ok {
				taskIDs = ids
			}

			pb.LockCleanup = &proto.LockCleanupReport{Cleaned: int32(cleaned), TaskIds: taskIDs}
		}
	}

	var suggestedNext []map[string]interface{}
	if includeTasks && tasksErr == nil {
		suggestedNext = getSuggestedNextTasksFromTasks(tasks, 10)
		if len(suggestedNext) > 0 {
			if cmd := buildCursorCliSuggestion(suggestedNext[0]); cmd != "" {
				pb.CursorCliSuggestion = cmd
			}
		}
	}

	result := SessionPrimeResultToMap(pb)

	if includeTasks {
		if tasksErr != nil {
			result["tasks"] = map[string]interface{}{"error": "Failed to load tasks"}
		} else {
			result["tasks"] = getTasksSummaryFromTasks(tasks)
			if len(suggestedNext) > 0 {
				result["suggested_next"] = suggestedNext
				if hint := buildSuggestedNextAction(suggestedNext[0]); hint != "" {
					result["suggested_next_action"] = hint
				}
			}
		}
	}

	if includeHints {
		result["hints"] = hints
	}

	if handoffAlert != nil {
		result["handoff_alert"] = handoffAlert
	}
	// Explicit status context: machine-readable enum (dashboard|handoff|task) and display label from single source of truth
	statusLabel, statusContext, _ := GetSessionStatus(projectRoot)
	result["status_label"] = statusLabel
	result["status_context"] = statusContext // enum: dashboard, handoff, or task
	pb.StatusLabel = statusLabel
	pb.StatusContext = statusContext

	compact := cast.ToBool(params["compact"])
	return FormatResultOptionalCompact(result, "", compact)
}

// handleSessionHandoff handles handoff actions (end, resume, latest, list, sync, export).
func handleSessionHandoff(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action := cast.ToString(params["action"])
	// Note: The action parameter for handoff might be nested - check both
	// Also check for sub_action parameter (for nested actions like export)
	if action == "" || action == "handoff" {
		// Check for sub_action first (for explicit sub-actions)
		if subAction := strings.TrimSpace(cast.ToString(params["sub_action"])); subAction != "" {
			action = subAction
		} else if summary := strings.TrimSpace(cast.ToString(params["summary"])); summary != "" {
			// Check if this is called from session tool with action="handoff"
			action = "end"
		} else {
			action = "resume"
		}
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	switch action {
	case "end":
		return handleSessionEnd(ctx, params, projectRoot)
	case "resume":
		return handleSessionResume(ctx, projectRoot)
	case "latest":
		return handleSessionLatest(projectRoot)
	case "list":
		return handleSessionList(ctx, params, projectRoot)
	case "sync":
		return handleSessionSync(ctx, params, projectRoot)
	case "export":
		return handleSessionExport(ctx, params, projectRoot)
	case "close", "approve":
		status := "closed"
		if action == "approve" {
			status = "approved"
		}

		return handleSessionHandoffStatus(ctx, params, projectRoot, status)
	case "delete":
		return handleSessionHandoffDelete(ctx, params, projectRoot)
	default:
		return nil, fmt.Errorf("unknown handoff action: %s (use 'end', 'resume', 'latest', 'list', 'sync', 'export', 'close', 'approve', or 'delete')", action)
	}
}

// handleSessionEnd ends a session and creates handoff note.
func handleSessionEnd(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	summary := cast.ToString(params["summary"])
	blockersRaw := params["blockers"]
	nextStepsRaw := params["next_steps"]

	var blockers []string

	if blockersRaw != nil {
		switch v := blockersRaw.(type) {
		case string:
			if v != "" {
				var blkList []string
				if err := json.Unmarshal([]byte(v), &blkList); err == nil {
					blockers = blkList
				} else {
					blockers = []string{v}
				}
			}
		case []interface{}:
			for _, b := range v {
				if str, ok := b.(string); ok {
					blockers = append(blockers, str)
				}
			}
		}
	}

	var nextSteps []string

	if nextStepsRaw != nil {
		switch v := nextStepsRaw.(type) {
		case string:
			if v != "" {
				var stepsList []string
				if err := json.Unmarshal([]byte(v), &stepsList); err == nil {
					nextSteps = stepsList
				} else {
					nextSteps = []string{v}
				}
			}
		case []interface{}:
			for _, s := range v {
				if str, ok := s.(string); ok {
					nextSteps = append(nextSteps, str)
				}
			}
		}
	}

	unassignMyTasks := true
	if _, ok := params["unassign_my_tasks"]; ok {
		unassignMyTasks = cast.ToBool(params["unassign_my_tasks"])
	}

	includeGitStatus := true
	if _, ok := params["include_git_status"]; ok {
		includeGitStatus = cast.ToBool(params["include_git_status"])
	}

	dryRun := false
	if _, ok := params["dry_run"]; ok {
		dryRun = cast.ToBool(params["dry_run"])
	}

	// Get hostname
	hostname, _ := os.Hostname()

	// Get Git status if requested
	var gitStatus map[string]interface{}
	if includeGitStatus {
		gitStatus = getGitStatus(ctx, projectRoot)
	}

	// Get current tasks in progress and optionally full task list for point-in-time snapshot
	var tasksInProgress []map[string]interface{}
	var allTasksForSnapshot []Todo2Task
	store := NewDefaultTaskStore(projectRoot)

	if _, has := params["include_tasks"]; !has || cast.ToBool(params["include_tasks"]) {
		list, err := store.ListTasks(ctx, nil)
		if err == nil {
			tasks := tasksFromPtrs(list)
			for _, task := range tasks {
				if task.Status == "In Progress" {
					tasksInProgress = append(tasksInProgress, map[string]interface{}{
						"id":      task.ID,
						"content": task.Content,
						"status":  task.Status,
					})
				}
			}
			allTasksForSnapshot = tasks
		}
	}

	includePointInTimeSnapshot := cast.ToBool(params["include_point_in_time_snapshot"])
	if includePointInTimeSnapshot && len(allTasksForSnapshot) == 0 {
		list, err := store.ListTasks(ctx, nil)
		if err == nil {
			allTasksForSnapshot = tasksFromPtrs(list)
		}
	}

	// Optional: task journal (modified tasks this session). Caller can pass modified_task_ids or task_journal.
	var taskJournal []map[string]interface{}
	if ids, ok := params["modified_task_ids"].([]interface{}); ok {
		for _, v := range ids {
			if id, ok := v.(string); ok && id != "" {
				taskJournal = append(taskJournal, map[string]interface{}{"id": id, "action": "modified"})
			}
		}
	}
	if journal, ok := params["task_journal"].([]interface{}); ok && len(taskJournal) == 0 {
		for _, v := range journal {
			if m, ok := v.(map[string]interface{}); ok {
				taskJournal = append(taskJournal, m)
			}
		}
	}
	if journalRaw, ok := params["task_journal"].(string); ok && journalRaw != "" && len(taskJournal) == 0 {
		var decoded []map[string]interface{}
		if err := json.Unmarshal([]byte(journalRaw), &decoded); err == nil {
			taskJournal = decoded
		}
	}

	// Point-in-time snapshot: full task list as gzip+base64 (state.todo2.json shape)
	var pointInTimeSnapshot, pointInTimeSnapshotFormat string
	if includePointInTimeSnapshot && len(allTasksForSnapshot) > 0 {
		raw, err := MarshalTasksToStateJSON(allTasksForSnapshot)
		if err == nil {
			var buf bytes.Buffer
			w := gzip.NewWriter(&buf)
			if _, err := w.Write(raw); err == nil {
				if err := w.Close(); err == nil {
					pointInTimeSnapshot = base64.StdEncoding.EncodeToString(buf.Bytes())
					pointInTimeSnapshotFormat = "gz+b64"
				}
			}
		}
	}

	// Create handoff note
	handoff := map[string]interface{}{
		"id":                fmt.Sprintf("handoff-%d", time.Now().Unix()),
		"timestamp":         time.Now().Format(time.RFC3339),
		"host":              hostname,
		"summary":           summary,
		"blockers":          blockers,
		"next_steps":        nextSteps,
		"git_status":        gitStatus,
		"tasks_in_progress": tasksInProgress,
	}
	if len(taskJournal) > 0 {
		handoff["task_journal"] = taskJournal
	}
	if pointInTimeSnapshot != "" {
		handoff["point_in_time_snapshot"] = pointInTimeSnapshot
		handoff["point_in_time_snapshot_format"] = pointInTimeSnapshotFormat
		handoff["point_in_time_snapshot_task_count"] = len(allTasksForSnapshot)
	}

	if !dryRun {
		// Save handoff
		if err := saveHandoff(projectRoot, handoff); err != nil {
			return nil, fmt.Errorf("failed to save handoff: %w", err)
		}

		// Unassign tasks if requested (release lock for current agent on in-progress tasks)
		if unassignMyTasks {
			agentID, err := database.GetAgentID()
			if err == nil {
				for _, m := range tasksInProgress {
					if id, ok := m["id"].(string); ok && id != "" {
						_ = database.ReleaseTask(ctx, id, agentID)
					}
				}
			}
		}
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"dry_run": dryRun,
		"handoff": handoff,
		"message": "Session ended. Handoff note created.",
	}

	if dryRun {
		result["message"] = "Dry run: Would create handoff note"
	}

	// Add suggested_next_action and cursor_cli_suggestion for the next suggested task
	if suggested := GetSuggestedNextTasks(projectRoot, 1); len(suggested) > 0 {
		t := suggested[0]

		taskMap := map[string]interface{}{"id": t.ID, "content": t.Content}
		if hint := buildSuggestedNextAction(taskMap); hint != "" {
			result["suggested_next_action"] = hint
		}

		if cmd := buildCursorCliSuggestion(taskMap); cmd != "" {
			result["cursor_cli_suggestion"] = cmd
		}
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionResume resumes a session by reviewing latest handoff.
func handleSessionResume(ctx context.Context, projectRoot string) ([]framework.TextContent, error) {
	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")

	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		result := map[string]interface{}{
			"success":     true,
			"method":      "native_go",
			"has_handoff": false,
			"message":     "No handoff notes found. Starting fresh session.",
		}

		return mcpresponse.FormatResult(result, "")
	}

	// Load handoff history (using file cache)
	fileCache := cache.GetGlobalFileCache()

	data, _, err := fileCache.ReadFile(handoffFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read handoff file: %w", err)
	}

	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return nil, fmt.Errorf("failed to parse handoff file: %w", err)
	}

	handoffs, _ := handoffData["handoffs"].([]interface{})
	if len(handoffs) == 0 {
		result := map[string]interface{}{
			"success":     true,
			"method":      "native_go",
			"has_handoff": false,
			"message":     "No handoff notes found. Starting fresh session.",
		}

		return mcpresponse.FormatResult(result, "")
	}

	// Get latest handoff (last in array)
	latestHandoff := handoffs[len(handoffs)-1]

	handoffMap, ok := latestHandoff.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid handoff format")
	}

	hostname, _ := os.Hostname()
	handoffHost, _ := handoffMap["host"].(string)

	result := map[string]interface{}{
		"success":        true,
		"method":         "native_go",
		"has_handoff":    true,
		"handoff":        handoffMap,
		"from_same_host": handoffHost == hostname,
		"message":        fmt.Sprintf("Resuming session. Latest handoff from %s", handoffHost),
	}

	// Add suggested_next_action and cursor_cli_suggestion for first suggested task (same as prime)
	if suggested := GetSuggestedNextTasks(projectRoot, 1); len(suggested) > 0 {
		t := suggested[0]

		taskMap := map[string]interface{}{"id": t.ID, "content": t.Content}
		if hint := buildSuggestedNextAction(taskMap); hint != "" {
			result["suggested_next_action"] = hint
		}

		if cmd := buildCursorCliSuggestion(taskMap); cmd != "" {
			result["cursor_cli_suggestion"] = cmd
		}
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionLatest gets the most recent handoff note.
func handleSessionLatest(projectRoot string) ([]framework.TextContent, error) {
	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")

	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		result := map[string]interface{}{
			"success":     false,
			"method":      "native_go",
			"has_handoff": false,
			"message":     "No handoff notes found",
		}

		return mcpresponse.FormatResult(result, "")
	}

	fileCache := cache.GetGlobalFileCache()

	data, _, err := fileCache.ReadFile(handoffFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read handoff file: %w", err)
	}

	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return nil, fmt.Errorf("failed to parse handoff file: %w", err)
	}

	handoffs, _ := handoffData["handoffs"].([]interface{})
	if len(handoffs) == 0 {
		result := map[string]interface{}{
			"success":     false,
			"method":      "native_go",
			"has_handoff": false,
			"message":     "No handoff notes found",
		}

		return mcpresponse.FormatResult(result, "")
	}

	latestHandoff := handoffs[len(handoffs)-1]

	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"has_handoff": true,
		"handoff":     latestHandoff,
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionList lists recent handoff notes.
func handleSessionList(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	limit := 5
	if l := cast.ToFloat64(params["limit"]); l > 0 {
		limit = int(l)
	}

	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")

	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		result := map[string]interface{}{
			"success":  true,
			"method":   "native_go",
			"handoffs": []interface{}{},
			"count":    0,
		}

		return mcpresponse.FormatResult(result, "")
	}

	fileCache := cache.GetGlobalFileCache()

	data, _, err := fileCache.ReadFile(handoffFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read handoff file: %w", err)
	}

	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return nil, fmt.Errorf("failed to parse handoff file: %w", err)
	}

	handoffs, _ := handoffData["handoffs"].([]interface{})

	// Filter closed/approved if include_closed is false (default)
	includeClosed := true
	if _, ok := params["include_closed"]; ok {
		includeClosed = cast.ToBool(params["include_closed"])
	}

	if !includeClosed {
		var open []interface{}

		for _, v := range handoffs {
			h, ok := v.(map[string]interface{})
			if !ok {
				continue
			}

			status, _ := h["status"].(string)
			if status != "closed" && status != "approved" {
				open = append(open, v)
			}
		}

		handoffs = open
	}

	// Get last N handoffs
	start := len(handoffs) - limit
	if start < 0 {
		start = 0
	}

	recentHandoffs := handoffs[start:]

	result := map[string]interface{}{
		"success":  true,
		"method":   "native_go",
		"handoffs": recentHandoffs,
		"count":    len(recentHandoffs),
		"total":    len(handoffs),
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionSync syncs Todo2 state across agents.
func handleSessionSync(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	direction := "both"
	if d := strings.TrimSpace(cast.ToString(params["direction"])); d != "" {
		direction = d
	}

	autoCommit := true
	if _, ok := params["auto_commit"]; ok {
		autoCommit = cast.ToBool(params["auto_commit"])
	}

	dryRun := false
	if _, ok := params["dry_run"]; ok {
		dryRun = cast.ToBool(params["dry_run"])
	}

	// For sync, we'll use Git operations to pull/push Todo2 state
	// This is a simplified implementation
	result := map[string]interface{}{
		"success":   true,
		"method":    "native_go",
		"direction": direction,
		"dry_run":   dryRun,
		"message":   "Todo2 state sync via Git (simplified implementation)",
		"note":      "Full sync implementation requires agentic-tools MCP integration",
	}

	// Basic Git sync implementation: all Git writes run under repo-level lock (T-78)
	if !dryRun {
		err := utils.WithGitLock(projectRoot, 0, func() error {
			if direction == "pull" || direction == "both" {
				cmd := exec.CommandContext(ctx, "git", "pull", "--no-edit")
				cmd.Dir = projectRoot

				if err := cmd.Run(); err == nil {
					result["pulled"] = true
				}
			}

			if (direction == "push" || direction == "both") && autoCommit {
				cmd := exec.CommandContext(ctx, "git", "status", "--porcelain", ".todo2")
				cmd.Dir = projectRoot
				output, err := cmd.Output()

				if err == nil && len(output) > 0 {
					cmd = exec.CommandContext(ctx, "git", "add", ".todo2")

					cmd.Dir = projectRoot
					if err := cmd.Run(); err != nil {
						return fmt.Errorf("git add .todo2: %w", err)
					}

					cmd = exec.CommandContext(ctx, "git", "commit", "-m", "Auto-sync Todo2 state")

					cmd.Dir = projectRoot
					if err := cmd.Run(); err != nil {
						return fmt.Errorf("git commit auto-sync Todo2 state: %w", err)
					}

					result["committed"] = true
				}

				cmd = exec.CommandContext(ctx, "git", "push")
				cmd.Dir = projectRoot

				if err := cmd.Run(); err == nil {
					result["pushed"] = true
				}
			}

			return nil
		})
		if err != nil {
			result["success"] = false
			result["error"] = err.Error()

			return mcpresponse.FormatResult(result, "")
		}
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionExport exports handoff data to a JSON file for sharing between agents.
func handleSessionExport(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	outputPath := cast.ToString(params["output_path"])
	if outputPath == "" {
		// Default to handoff-export-{timestamp}.json in project root
		outputPath = filepath.Join(projectRoot, fmt.Sprintf("handoff-export-%d.json", time.Now().Unix()))
	}

	// Make path absolute if relative
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(projectRoot, outputPath)
	}

	// Get which handoffs to export (latest or all)
	exportLatest := true
	if _, ok := params["export_latest"]; ok {
		exportLatest = cast.ToBool(params["export_latest"])
	}

	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")

	// Load handoff data
	var handoffData map[string]interface{}
	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		// No handoffs file - return empty export
		handoffData = map[string]interface{}{
			"handoffs":    []interface{}{},
			"exported_at": time.Now().Format(time.RFC3339),
			"export_type": "all",
			"count":       0,
		}
	} else {
		fileCache := cache.GetGlobalFileCache()

		data, _, err := fileCache.ReadFile(handoffFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read handoff file: %w", err)
		}

		if err := json.Unmarshal(data, &handoffData); err != nil {
			return nil, fmt.Errorf("failed to parse handoff file: %w", err)
		}

		handoffs, _ := handoffData["handoffs"].([]interface{})

		// If exporting latest only, keep only the last handoff
		if exportLatest && len(handoffs) > 0 {
			handoffData["handoffs"] = []interface{}{handoffs[len(handoffs)-1]}
			handoffData["export_type"] = "latest"
		} else {
			handoffData["export_type"] = "all"
		}

		handoffData["exported_at"] = time.Now().Format(time.RFC3339)
		handoffData["count"] = len(handoffData["handoffs"].([]interface{}))
	}

	// Write to output file
	// Use compact JSON (no indentation) for better performance and smaller file size
	// This is more efficient for machine-to-machine transfer
	data, err := json.Marshal(handoffData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal handoff data: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write export file: %w", err)
	}

	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"output_path": outputPath,
		"export_type": handoffData["export_type"],
		"count":       handoffData["count"],
		"message":     fmt.Sprintf("Handoff data exported to %s", outputPath),
	}

	return mcpresponse.FormatResult(result, "")
}

// DecodePointInTimeSnapshot decodes a gzip+base64 point-in-time snapshot from a handoff.
// Returns the raw JSON bytes (state.todo2.json shape with "todos" array). Use ParseTasksFromJSON to get []Todo2Task.
func DecodePointInTimeSnapshot(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, fmt.Errorf("empty snapshot")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	r, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer r.Close()
	var out bytes.Buffer
	if _, err := out.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("gzip read: %w", err)
	}
	return out.Bytes(), nil
}

// Helper types and functions

type AgentInfo struct {
	Agent  string
	Source string
	Config map[string]interface{}
}

type AgentContext struct {
	Agent           string
	RecommendedMode string
	FocusAreas      []string
	RelevantTools   []string
}

type TimeSuggestion struct {
	Mode       string
	Reason     string
	Confidence float64
}

// detectAgentType detects the current agent type.
func detectAgentType(projectRoot string) AgentInfo {
	// 0. Claude Code detection (check before EXARP_AGENT so it's always recognized)
	if os.Getenv("CLAUDE_CODE") != "" || os.Getenv("ANTHROPIC_CLI") != "" || os.Getenv("CLAUDE_CODE_VERSION") != "" {
		return AgentInfo{
			Agent:  "claude-code",
			Source: "environment",
			Config: map[string]interface{}{"client": "claude-code"},
		}
	}

	// 1. Environment variable
	if agent := os.Getenv("EXARP_AGENT"); agent != "" {
		return AgentInfo{
			Agent:  agent,
			Source: "environment",
			Config: map[string]interface{}{},
		}
	}

	// 2. cursor-agent.json
	fileCache := cache.GetGlobalFileCache()
	agentConfigPath := filepath.Join(projectRoot, "cursor-agent.json")

	if data, _, err := fileCache.ReadFile(agentConfigPath); err == nil {
		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err == nil {
			agent := ""
			if name, ok := config["name"].(string); ok {
				agent = name
			} else if agentVal, ok := config["agent"].(string); ok {
				agent = agentVal
			}

			if agent != "" {
				return AgentInfo{
					Agent:  agent,
					Source: "cursor-agent.json",
					Config: config,
				}
			}
		}
	}

	// 3. Check agents directory structure
	cwd, _ := os.Getwd()
	if strings.Contains(cwd, "agents") {
		parts := strings.Split(cwd, string(filepath.Separator))
		for i, part := range parts {
			if part == "agents" && i+1 < len(parts) {
				return AgentInfo{
					Agent:  parts[i+1],
					Source: "path",
					Config: map[string]interface{}{},
				}
			}
		}
	}

	// 4. Default
	return AgentInfo{
		Agent:  "general",
		Source: "default",
		Config: map[string]interface{}{},
	}
}

// getAgentContext returns context for an agent type.
func getAgentContext(agent string) AgentContext {
	agentLower := strings.ToLower(strings.ReplaceAll(agent, "-agent", ""))

	// Agent mode mapping
	agentModeMap := map[string]string{
		"backend":  "development",
		"web":      "development",
		"security": "security_review",
		"qa":       "code_review",
		"devops":   "sprint_planning",
		"pm":       "task_management",
	}

	mode := "development"
	if m, ok := agentModeMap[agentLower]; ok {
		mode = m
	}

	// Agent context (simplified)
	agentContextMap := map[string]AgentContext{
		"backend": {
			Agent:           agent,
			RecommendedMode: "development",
			FocusAreas:      []string{"API development", "database", "services"},
			RelevantTools:   []string{"run_tests", "scan_dependency_security", "analyze_test_coverage"},
		},
		"web": {
			Agent:           agent,
			RecommendedMode: "development",
			FocusAreas:      []string{"Frontend", "React", "TypeScript"},
			RelevantTools:   []string{"run_tests", "generate_config"},
		},
		"security": {
			Agent:           agent,
			RecommendedMode: "security_review",
			FocusAreas:      []string{"Vulnerability scanning", "Dependency audit", "Security review"},
			RelevantTools:   []string{"scan_dependency_security", "fetch_dependabot_alerts", "generate_security_report"},
		},
	}

	if ctx, ok := agentContextMap[agentLower]; ok {
		return ctx
	}

	// Default general context
	return AgentContext{
		Agent:           agent,
		RecommendedMode: mode,
		FocusAreas:      []string{"General development", "Project management"},
		RelevantTools:   []string{"project_scorecard", "check_documentation_health"},
	}
}

// suggestModeByTime suggests workflow mode based on time of day.
func suggestModeByTime() TimeSuggestion {
	hour := time.Now().Hour()

	switch {
	case hour >= 6 && hour < 9:
		return TimeSuggestion{
			Mode:       "daily_checkin",
			Reason:     "Morning hours - start with a health check",
			Confidence: 0.7,
		}
	case hour >= 9 && hour < 12:
		return TimeSuggestion{
			Mode:       "development",
			Reason:     "Prime working hours - development focus",
			Confidence: 0.8,
		}
	case hour >= 12 && hour < 13:
		return TimeSuggestion{
			Mode:       "task_management",
			Reason:     "Lunch time - good for task review",
			Confidence: 0.5,
		}
	case hour >= 13 && hour < 17:
		return TimeSuggestion{
			Mode:       "development",
			Reason:     "Afternoon - continued development",
			Confidence: 0.8,
		}
	case hour >= 17 && hour < 18:
		return TimeSuggestion{
			Mode:       "code_review",
			Reason:     "End of day - review and wrap up",
			Confidence: 0.6,
		}
	default:
		return TimeSuggestion{
			Mode:       "development",
			Reason:     "Evening/night - default development mode",
			Confidence: 0.5,
		}
	}
}

// getWorkflowModeDescription returns description for a workflow mode.
func getWorkflowModeDescription(mode string) string {
	descriptions := map[string]string{
		"daily_checkin":   "Daily check-in mode: Overview + health checks",
		"security_review": "Security review mode: Security-focused tools",
		"task_management": "Task management mode: Task tools only",
		"sprint_planning": "Sprint planning mode: Tasks + automation + PRD",
		"code_review":     "Code review mode: Testing + linting",
		"development":     "Development mode: Balanced set",
		"debugging":       "Debugging mode: Memory + testing",
		"all":             "All tools mode: Full tool access",
	}

	if desc, ok := descriptions[mode]; ok {
		return desc
	}

	return "Development mode: Balanced set"
}

// getPlanModeContext returns (plan_path, plan_mode_hint) for session prime.
// plan_path: relative path to current .plan.md if found.
// plan_mode_hint: suggestion to use structured planning when backlog is complex.
func getPlanModeContext(projectRoot string, tasks []Todo2Task) (planPath, planModeHint string) {
	planPath = getCurrentPlanPath(projectRoot)

	if shouldSuggestPlanMode(tasks) {
		planModeHint = "Complex backlog detected: consider structured planning — generate a .plan.md (report action=plan) and work through tasks in dependency order"
	}

	return planPath, planModeHint
}

// getCurrentPlanPath returns the relative path to the primary .plan.md file if one exists.
// Checks: .cursor/plans/{project-slug}.plan.md, {project-slug}.plan.md in root, then most recent in .cursor/plans/.
func getCurrentPlanPath(projectRoot string) string {
	slug := filepath.Base(projectRoot)
	if slug == "." || slug == "" {
		slug = "exarp-go" // fallback
	}

	// 1. .cursor/plans/{slug}.plan.md (canonical location)
	candidates := []string{
		filepath.Join(".cursor", "plans", slug+".plan.md"),
		slug + ".plan.md", // project root
	}
	for _, rel := range candidates {
		p := filepath.Join(projectRoot, rel)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return rel
		}
	}

	// 2. Most recently modified .plan.md in .cursor/plans/
	plansDir := filepath.Join(projectRoot, ".cursor", "plans")

	entries, err := os.ReadDir(plansDir)
	if err != nil {
		return ""
	}

	var bestPath string

	var bestMod time.Time

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if !strings.HasSuffix(e.Name(), ".plan.md") {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(bestMod) {
			bestMod = info.ModTime()
			bestPath = filepath.Join(".cursor", "plans", e.Name())
		}
	}

	return bestPath
}

// shouldSuggestPlanMode returns true when backlog suggests complex or multi-step work.
// Uses heuristics aligned with infer_session_mode: many high-priority items, large backlog, or high dependency ratio.
func shouldSuggestPlanMode(tasks []Todo2Task) bool {
	if len(tasks) == 0 {
		return false
	}

	backlog := 0
	highPriority := 0
	withDeps := 0

	for _, t := range tasks {
		if !IsBacklogStatus(t.Status) {
			continue
		}

		backlog++

		if t.Priority == "high" || t.Priority == "critical" {
			highPriority++
		}

		if len(t.Dependencies) > 0 {
			withDeps++
		}
	}

	if backlog == 0 {
		return false
	}

	total := len(tasks)
	depsRatio := float64(withDeps) / float64(total)

	// Suggest Plan Mode when: many backlog items, many high-priority, or high dependency ratio
	const backlogThreshold = 15

	const highPriorityThreshold = 5

	const depsRatioThreshold = 0.4

	return backlog >= backlogThreshold ||
		highPriority >= highPriorityThreshold ||
		depsRatio >= depsRatioThreshold
}

// getHintsForMode returns tool hints for a mode (simplified).
func getHintsForMode(mode string) map[string]string {
	// Simplified hints - can be expanded
	hints := map[string]string{
		"session":           "Use session tool for context priming and handoffs",
		"tasks":             "Use task_workflow tool for task management",
		"tractatus":         "Consider tractatus_thinking for logical decomposition of complex concepts (use operation=start)",
		"context_reduction": "When context is large: use compact=true on prime/task_workflow/report; call context(action=budget, items=[...], budget_tokens=N) for safe_to_summarize and agent_hint.",
	}

	switch mode {
	case "security_review":
		hints["security"] = "Use security tool for vulnerability scanning"
	case "code_review":
		hints["testing"] = "Use testing tool for test coverage"
		hints["lint"] = "Use lint tool for code quality; for docs/markdown includes broken link check (gomarklint)"
	case "task_management":
		hints["task_workflow"] = "Use task_workflow for task operations"
	}

	return hints
}

// getTasksSummaryFromTasks returns a summary from already-loaded tasks (avoids duplicate load).
func getTasksSummaryFromTasks(tasks []Todo2Task) map[string]interface{} {
	byStatus := make(map[string]int)
	for _, task := range tasks {
		byStatus[task.Status]++
	}

	recentTasks := []map[string]interface{}{}

	for i, task := range tasks {
		if i >= 10 {
			break
		}

		recentTasks = append(recentTasks, map[string]interface{}{
			"id":       task.ID,
			"content":  task.Content,
			"status":   task.Status,
			"priority": task.Priority,
		})
	}

	return map[string]interface{}{
		"total":     len(tasks),
		"by_status": byStatus,
		"recent":    recentTasks,
	}
}

// getSuggestedNextTasksFromTasks returns first N backlog tasks in dependency order (avoids duplicate load).
func getSuggestedNextTasksFromTasks(tasks []Todo2Task, limit int) []map[string]interface{} {
	if limit <= 0 {
		return nil
	}

	orderedIDs, _, details, err := BacklogExecutionOrder(tasks, nil)
	if err != nil || len(orderedIDs) == 0 {
		return nil
	}

	taskByID := make(map[string]Todo2Task)
	for _, t := range tasks {
		taskByID[t.ID] = t
	}

	out := make([]map[string]interface{}, 0, limit)

	detailMap := make(map[string]BacklogTaskDetail)
	for _, d := range details {
		detailMap[d.ID] = d
	}

	for i, id := range orderedIDs {
		if i >= limit {
			break
		}

		d, ok := detailMap[id]
		if !ok {
			m := map[string]interface{}{"id": id, "content": ""}
			if t, has := taskByID[id]; has {
				if rt := GetRecommendedTools(t.Metadata); len(rt) > 0 {
					m["recommended_tools"] = rt
				}
			}
			out = append(out, m)
			continue
		}

		m := map[string]interface{}{
			"id":       d.ID,
			"content":  d.Content,
			"priority": d.Priority,
			"level":    d.Level,
		}
		if t, has := taskByID[id]; has {
			if rt := GetRecommendedTools(t.Metadata); len(rt) > 0 {
				m["recommended_tools"] = rt
			}
		}
		out = append(out, m)
	}

	return out
}

// GetSessionStatus returns current session context for UI status display (e.g. Cursor when it supports dynamic labels).
// Returns label (short display text), contextType (handoff|task|dashboard), and details map.
// Reuses checkHandoffAlert and GetSuggestedNextTasks to avoid duplication.
func GetSessionStatus(projectRoot string) (label, contextType string, details map[string]interface{}) {
	details = make(map[string]interface{})

	if handoff := checkHandoffAlert(projectRoot); handoff != nil {
		label = "Handoff – review pending"
		contextType = "handoff"
		details["handoff"] = handoff
		details["action"] = "Review handoff from previous developer"

		return
	}

	suggested := GetSuggestedNextTasks(projectRoot, 1)
	if len(suggested) > 0 {
		t := suggested[0]
		label = t.ID + " – " + truncateString(t.Content, 40)
		contextType = "task"
		details["task_id"] = t.ID
		details["content"] = t.Content
		details["priority"] = t.Priority

		return
	}

	label = "Project dashboard"
	contextType = "dashboard"

	return
}

// checkHandoffAlert checks for handoff notes from other developers.
func checkHandoffAlert(projectRoot string) map[string]interface{} {
	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")
	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		return nil
	}

	fileCache := cache.GetGlobalFileCache()

	data, _, err := fileCache.ReadFile(handoffFile)
	if err != nil {
		return nil
	}

	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return nil
	}

	handoffs, _ := handoffData["handoffs"].([]interface{})
	if len(handoffs) == 0 {
		return nil
	}

	// Get latest handoff
	latestHandoff := handoffs[len(handoffs)-1]

	handoffMap, ok := latestHandoff.(map[string]interface{})
	if !ok {
		return nil
	}

	hostname, _ := os.Hostname()
	handoffHost, _ := handoffMap["host"].(string)

	// Only show if from different host
	if handoffHost != hostname {
		return map[string]interface{}{
			"from_host":  handoffMap["host"],
			"timestamp":  handoffMap["timestamp"],
			"summary":    truncateString(fmt.Sprintf("%v", handoffMap["summary"]), 100),
			"blockers":   handoffMap["blockers"],
			"next_steps": handoffMap["next_steps"],
		}
	}

	return nil
}

// saveHandoff saves a handoff note to the handoffs.json file.
func saveHandoff(projectRoot string, handoff map[string]interface{}) error {
	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")

	// Ensure .todo2 directory exists
	if err := os.MkdirAll(filepath.Dir(handoffFile), 0755); err != nil {
		return err
	}

	// Load existing handoffs
	fileCache := cache.GetGlobalFileCache()
	handoffs := []interface{}{}

	if data, _, err := fileCache.ReadFile(handoffFile); err == nil {
		var handoffData map[string]interface{}
		if err := json.Unmarshal(data, &handoffData); err == nil {
			if existing, ok := handoffData["handoffs"].([]interface{}); ok {
				handoffs = existing
			}
		}
	}

	// Add new handoff
	handoffs = append(handoffs, handoff)

	// Keep last 20 handoffs
	if len(handoffs) > 20 {
		handoffs = handoffs[len(handoffs)-20:]
	}

	// Save
	handoffData := map[string]interface{}{
		"handoffs": handoffs,
	}

	data, err := json.MarshalIndent(handoffData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(handoffFile, data, 0644)
}

// updateHandoffStatus sets status on handoffs by id and writes handoffs.json back.
func updateHandoffStatus(projectRoot string, handoffIDs []string, status string) error {
	if len(handoffIDs) == 0 {
		return nil
	}

	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")
	if err := os.MkdirAll(filepath.Dir(handoffFile), 0755); err != nil {
		return err
	}

	idsSet := make(map[string]struct{})

	for _, id := range handoffIDs {
		if id != "" {
			idsSet[id] = struct{}{}
		}
	}

	fileCache := cache.GetGlobalFileCache()

	data, _, err := fileCache.ReadFile(handoffFile)
	if err != nil {
		return err
	}

	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return err
	}

	handoffs, _ := handoffData["handoffs"].([]interface{})
	if len(handoffs) == 0 {
		return nil
	}

	updated := 0

	for _, v := range handoffs {
		h, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := h["id"].(string)
		if _, want := idsSet[id]; want {
			h["status"] = status
			updated++
		}
	}

	if updated == 0 {
		return nil
	}

	handoffData["handoffs"] = handoffs

	out, err := json.MarshalIndent(handoffData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(handoffFile, out, 0644)
}

// handleSessionHandoffStatus closes or approves handoffs by id.
func handleSessionHandoffStatus(ctx context.Context, params map[string]interface{}, projectRoot, status string) ([]framework.TextContent, error) {
	var ids []string
	if id := strings.TrimSpace(cast.ToString(params["handoff_id"])); id != "" {
		ids = []string{id}
	} else if raw, ok := params["handoff_ids"]; ok {
		switch v := raw.(type) {
		case []interface{}:
			for _, i := range v {
				if s, ok := i.(string); ok && s != "" {
					ids = append(ids, s)
				}
			}
		case string:
			if v != "" {
				var list []string
				if json.Unmarshal([]byte(v), &list) == nil {
					ids = list
				} else {
					ids = []string{v}
				}
			}
		}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("handoff_id or handoff_ids required for close/approve")
	}

	if err := updateHandoffStatus(projectRoot, ids, status); err != nil {
		return nil, fmt.Errorf("failed to update handoff status: %w", err)
	}

	label := "closed"
	if status == "approved" {
		label = "approved"
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"updated": len(ids),
		"status":  status,
		"message": fmt.Sprintf("%d handoff(s) %s", len(ids), label),
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionHandoffDelete removes handoffs by id from handoffs.json.
func handleSessionHandoffDelete(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	var ids []string
	if id := strings.TrimSpace(cast.ToString(params["handoff_id"])); id != "" {
		ids = []string{id}
	} else if raw, ok := params["handoff_ids"]; ok {
		switch v := raw.(type) {
		case []interface{}:
			for _, i := range v {
				if s, ok := i.(string); ok && s != "" {
					ids = append(ids, s)
				}
			}
		case string:
			if v != "" {
				var list []string
				if json.Unmarshal([]byte(v), &list) == nil {
					ids = list
				} else {
					ids = []string{v}
				}
			}
		}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("handoff_id or handoff_ids required for delete")
	}

	deleted, err := deleteHandoffs(projectRoot, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to delete handoffs: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"deleted": deleted,
		"message": fmt.Sprintf("%d handoff(s) deleted", deleted),
	}

	return mcpresponse.FormatResult(result, "")
}

// deleteHandoffs removes handoffs by id from handoffs.json. Returns count deleted.
func deleteHandoffs(projectRoot string, handoffIDs []string) (int, error) {
	if len(handoffIDs) == 0 {
		return 0, nil
	}

	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")
	if err := os.MkdirAll(filepath.Dir(handoffFile), 0755); err != nil {
		return 0, err
	}

	idsSet := make(map[string]struct{})

	for _, id := range handoffIDs {
		if id != "" {
			idsSet[id] = struct{}{}
		}
	}

	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		return 0, nil
	}

	data, err := os.ReadFile(handoffFile)
	if err != nil {
		return 0, err
	}

	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return 0, err
	}

	handoffs, _ := handoffData["handoffs"].([]interface{})

	var kept []interface{}

	deleted := 0

	for _, v := range handoffs {
		h, ok := v.(map[string]interface{})
		if !ok {
			kept = append(kept, v)
			continue
		}

		id, _ := h["id"].(string)
		if _, want := idsSet[id]; want {
			deleted++
			continue
		}

		kept = append(kept, v)
	}

	if deleted == 0 {
		return 0, nil
	}

	handoffData["handoffs"] = kept

	out, err := json.MarshalIndent(handoffData, "", "  ")
	if err != nil {
		return 0, err
	}

	if err := os.WriteFile(handoffFile, out, 0644); err != nil {
		return 0, err
	}

	return deleted, nil
}

// getGitStatus gets current Git status.
func getGitStatus(ctx context.Context, projectRoot string) map[string]interface{} {
	status := map[string]interface{}{}

	// Get branch
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectRoot

	if output, err := cmd.Output(); err == nil {
		status["branch"] = strings.TrimSpace(string(output))
	}

	// Get status
	cmd = exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = projectRoot

	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")

		var changedFiles []string

		for _, line := range lines {
			if line != "" {
				changedFiles = append(changedFiles, strings.TrimSpace(line))
			}
		}

		status["uncommitted_files"] = len(changedFiles)
		status["changed_files"] = changedFiles

		if len(changedFiles) > 10 {
			status["changed_files"] = changedFiles[:10]
		}
	}

	return status
}

// buildSuggestedNextAction builds a client-agnostic next-action hint from a suggested task map.
// Expects a map with "id" and "content" keys. Returns empty string if task info is missing.
// For Cursor: can be used as argument to `agent -p`. For Claude Code: descriptive action hint.
func buildSuggestedNextAction(task map[string]interface{}) string {
	id, _ := task["id"].(string)
	content, _ := task["content"].(string)

	if id == "" {
		return ""
	}

	if content != "" {
		return fmt.Sprintf("Work on %s: %s", id, truncateString(content, 80))
	}

	return fmt.Sprintf("Work on %s", id)
}

// buildCursorCliSuggestion builds a ready-to-run Cursor CLI command from the first suggested task.
// Returns e.g. `agent -p "Work on T-123: Task name" --mode=plan` for session prime/handoff JSON.
// See docs/CURSOR_API_AND_CLI_INTEGRATION.md §3.2.
func buildCursorCliSuggestion(task map[string]interface{}) string {
	action := buildSuggestedNextAction(task)
	if action == "" {
		return ""
	}

	return fmt.Sprintf("agent -p %q --mode=plan", action)
}

// truncateString truncates a string to max length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

// handleSessionPrompts handles the prompts action - lists available prompts.
func handleSessionPrompts(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get optional filters
	mode := cast.ToString(params["mode"])
	persona := cast.ToString(params["persona"])
	category := cast.ToString(params["category"])
	keywords := cast.ToString(params["keywords"])

	limit := 50
	if l := cast.ToFloat64(params["limit"]); l > 0 {
		limit = int(l)
	}

	// Get all registered prompts from the prompts registry
	// We'll build a simple list from the known prompts
	allPrompts := []map[string]interface{}{
		// Original prompts (8)
		{"name": "align", "description": "Analyze Todo2 task alignment with project goals.", "category": "tasks"},
		{"name": "discover", "description": "Discover tasks from TODO comments, markdown, and orphaned tasks.", "category": "tasks"},
		{"name": "config", "description": "Generate IDE configuration files.", "category": "config"},
		{"name": "scan", "description": "Scan project dependencies for security vulnerabilities. Supports all languages via tool parameter.", "category": "security"},
		{"name": "scorecard", "description": "Generate comprehensive project health scorecard with all metrics.", "category": "reports"},
		{"name": "overview", "description": "Generate one-page project overview for stakeholders.", "category": "reports"},
		{"name": "dashboard", "description": "Display comprehensive project dashboard with key metrics and status overview.", "category": "reports"},
		{"name": "remember", "description": "Use AI session memory to persist insights.", "category": "memory"},
		// High-value workflow prompts (7)
		{"name": "daily_checkin", "description": "Daily check-in workflow: server status, blockers, git health.", "category": "workflow", "mode": "daily_checkin"},
		{"name": "sprint_start", "description": "Sprint start workflow: clean backlog, align tasks, queue work.", "category": "workflow", "mode": "sprint_planning"},
		{"name": "sprint_end", "description": "Sprint end workflow: test coverage, docs, security check.", "category": "workflow", "mode": "sprint_planning"},
		{"name": "pre_sprint", "description": "Pre-sprint cleanup workflow: duplicates, alignment, documentation.", "category": "workflow", "mode": "sprint_planning"},
		{"name": "post_impl", "description": "Post-implementation review workflow: docs, security, automation.", "category": "workflow"},
		{"name": "sync", "description": "Synchronize tasks between shared TODO table and Todo2.", "category": "tasks"},
		{"name": "dups", "description": "Find and consolidate duplicate Todo2 tasks.", "category": "tasks"},
		// mcp-generic-tools prompts
		{"name": "context", "description": "Manage LLM context with summarization and budget tools.", "category": "workflow"},
		{"name": "mode", "description": "Suggest optimal Cursor IDE mode (Agent vs Ask) for a task.", "category": "workflow"},
		// Task management prompts
		{"name": "task_update", "description": "Update Todo2 task status using proper MCP tools - never edit JSON directly.", "category": "tasks"},
	}

	// Filter prompts
	filtered := []map[string]interface{}{}

	for _, prompt := range allPrompts {
		// Filter by mode
		if mode != "" {
			if promptMode, ok := prompt["mode"].(string); !ok || promptMode != mode {
				continue
			}
		}

		// Filter by category
		if category != "" {
			if promptCategory, ok := prompt["category"].(string); !ok || promptCategory != category {
				continue
			}
		}

		// Filter by keywords (search in name and description)
		if keywords != "" {
			keywordsLower := strings.ToLower(keywords)
			name, _ := prompt["name"].(string)
			desc, _ := prompt["description"].(string)

			if !strings.Contains(strings.ToLower(name), keywordsLower) &&
				!strings.Contains(strings.ToLower(desc), keywordsLower) {
				continue
			}
		}

		// Persona filtering would require persona mapping - skip for now
		// (would need to add persona field to prompts)

		filtered = append(filtered, prompt)
		if len(filtered) >= limit {
			break
		}
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"prompts": filtered,
		"total":   len(filtered),
		"filters": map[string]interface{}{
			"mode":     mode,
			"persona":  persona,
			"category": category,
			"keywords": keywords,
			"limit":    limit,
		},
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionAssignee handles the assignee action - manages task assignments.
func handleSessionAssignee(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	subAction := cast.ToString(params["sub_action"])
	if subAction == "" {
		subAction = "list"
	}

	switch subAction {
	case "list":
		return handleSessionAssigneeList(ctx, params)
	case "assign":
		return handleSessionAssigneeAssign(ctx, params)
	case "unassign":
		return handleSessionAssigneeUnassign(ctx, params)
	default:
		return nil, fmt.Errorf("unknown sub_action: %s (use 'list', 'assign', or 'unassign')", subAction)
	}
}

// handleSessionAssigneeList lists tasks with their assignees.
func handleSessionAssigneeList(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	includeUnassigned := true
	if _, ok := params["include_unassigned"]; ok {
		includeUnassigned = cast.ToBool(params["include_unassigned"])
	}

	statusFilter := cast.ToString(params["status_filter"])

	// Get tasks from database
	var tasks []*database.Todo2Task

	var err error

	if statusFilter != "" {
		filters := &database.TaskFilters{Status: &statusFilter}
		tasks, err = database.ListTasks(ctx, filters)
	} else {
		// Get all pending tasks
		tasks, err = database.ListTasks(ctx, &database.TaskFilters{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Get assignee information from database
	// Note: assignee is stored in tasks_lock table, we'll need to query it
	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	assignments := []map[string]interface{}{}

	for _, task := range tasks {
		// Query assignee from tasks table
		var assignee sql.NullString

		var assignedAt sql.NullInt64 // assigned_at is INTEGER (Unix timestamp)

		err := db.QueryRowContext(ctx, `
			SELECT assignee, assigned_at
			FROM tasks
			WHERE id = ?
		`, task.ID).Scan(&assignee, &assignedAt)

		hasAssignee := err == nil && assignee.Valid && assignee.String != ""

		if !includeUnassigned && !hasAssignee {
			continue
		}

		assignment := map[string]interface{}{
			"task_id":     task.ID,
			"content":     task.Content,
			"status":      task.Status,
			"assignee":    nil,
			"assigned_at": nil,
		}

		if hasAssignee {
			assignment["assignee"] = assignee.String
			if assignedAt.Valid {
				// Convert Unix timestamp to ISO 8601 string
				assignment["assigned_at"] = time.Unix(assignedAt.Int64, 0).Format(time.RFC3339)
			}
		}

		assignments = append(assignments, assignment)
	}

	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"assignments": assignments,
		"total":       len(assignments),
		"filters": map[string]interface{}{
			"status_filter":      statusFilter,
			"include_unassigned": includeUnassigned,
		},
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionAssigneeAssign assigns a task to an agent/human/host.
func handleSessionAssigneeAssign(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := cast.ToString(params["task_id"])
	if taskID == "" {
		return nil, fmt.Errorf("task_id parameter is required")
	}

	assigneeName := cast.ToString(params["assignee_name"])
	if assigneeName == "" {
		return nil, fmt.Errorf("assignee_name parameter is required")
	}

	assigneeType := "agent"
	if t := strings.TrimSpace(cast.ToString(params["assignee_type"])); t != "" {
		assigneeType = t
	}

	dryRun := false
	if _, ok := params["dry_run"]; ok {
		dryRun = cast.ToBool(params["dry_run"])
	}

	// Get agent ID (for agent type)
	if assigneeType == "agent" {
		agentID, err := database.GetAgentID()
		if err == nil {
			// Use agent ID format: {agent-type}-{hostname}-{pid}
			// But for assignment, we might want just the name
			// For now, use the provided assigneeName
		}

		_ = agentID // Use if needed
	}

	// Use database ClaimTaskForAgent for atomic assignment
	// This handles locking and prevents race conditions
	leaseDuration := config.TaskLockLease()

	result, err := database.ClaimTaskForAgent(ctx, taskID, assigneeName, leaseDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to assign task: %w", err)
	}

	if !result.Success {
		if result.WasLocked {
			response := map[string]interface{}{
				"success":   false,
				"method":    "native_go",
				"error":     fmt.Sprintf("Task already assigned to %s", result.LockedBy),
				"task_id":   taskID,
				"locked_by": result.LockedBy,
			}

			return mcpresponse.FormatResult(response, "")
		}

		return nil, fmt.Errorf("task assignment failed: %w", result.Error)
	}

	response := map[string]interface{}{
		"success":       true,
		"method":        "native_go",
		"task_id":       taskID,
		"assignee":      assigneeName,
		"assignee_type": assigneeType,
		"assigned_at":   time.Now().Format(time.RFC3339),
		"dry_run":       dryRun,
	}

	return mcpresponse.FormatResult(response, "")
}

// handleSessionAssigneeUnassign unassigns a task.
func handleSessionAssigneeUnassign(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := cast.ToString(params["task_id"])
	if taskID == "" {
		return nil, fmt.Errorf("task_id parameter is required")
	}

	agentID, err := database.GetAgentID()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent ID: %w", err)
	}

	// Release task lock (this unassigns the task)
	err = database.ReleaseTask(ctx, taskID, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to unassign task: %w", err)
	}

	response := map[string]interface{}{
		"success":       true,
		"method":        "native_go",
		"task_id":       taskID,
		"unassigned_at": time.Now().Format(time.RFC3339),
	}

	return mcpresponse.FormatResult(response, "")
}
