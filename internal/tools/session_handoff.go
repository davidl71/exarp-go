// session_handoff.go — session handoff handlers.
package tools

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/utils"
	mcpresponse "github.com/davidl71/mcp-go-core/pkg/mcp/response"
	"github.com/spf13/cast"
)

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
		return handleSessionResume(ctx, params, projectRoot)
	case "latest":
		return handleSessionLatest(params, projectRoot)
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
				if task.Status == models.StatusInProgress {
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

	// Add suggested_next_action; add cursor_cli_suggestion only when include_cli_command is true (default false).
	includeCliCommand := cast.ToBool(params["include_cli_command"])
	if suggested := GetSuggestedNextTasks(projectRoot, 1); len(suggested) > 0 {
		t := suggested[0]

		taskMap := map[string]interface{}{"id": t.ID, "content": t.Content}
		if hint := buildSuggestedNextAction(taskMap); hint != "" {
			result["suggested_next_action"] = hint
		}

		if includeCliCommand {
			if cmd := buildCursorCliSuggestion(taskMap); cmd != "" {
				result["cursor_cli_suggestion"] = cmd
			}
		}
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionResume resumes a session by reviewing latest handoff.
func handleSessionResume(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
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

	// Add suggested_next_action; add cursor_cli_suggestion only when include_cli_command is true (default false).
	includeCliCommand := cast.ToBool(params["include_cli_command"])
	if suggested := GetSuggestedNextTasks(projectRoot, 1); len(suggested) > 0 {
		t := suggested[0]

		taskMap := map[string]interface{}{"id": t.ID, "content": t.Content}
		if hint := buildSuggestedNextAction(taskMap); hint != "" {
			result["suggested_next_action"] = hint
		}

		if includeCliCommand {
			if cmd := buildCursorCliSuggestion(taskMap); cmd != "" {
				result["cursor_cli_suggestion"] = cmd
			}
		}
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionLatest gets the most recent handoff note.
func handleSessionLatest(params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
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

