// session.go — MCP "session" tool: dispatcher, prime handler, and response helpers.
package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
)

// HandleSessionPrimeJSON returns session prime data as JSON bytes.
// Used by the prime://context resource handler for OpenCode session bootstrap.
func HandleSessionPrimeJSON(ctx context.Context) ([]byte, error) {
	params := map[string]interface{}{
		"include_hints": true,
		"include_tasks": true,
	}
	result, err := handleSessionPrime(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 || result[0].Text == "" {
		return nil, fmt.Errorf("session prime returned empty result")
	}
	return []byte(result[0].Text), nil
}

func isSessionHandoffSubAction(sub string) bool {
	switch sub {
	case "end", "resume", "latest", "list", "sync", "export", "close", "approve", "delete":
		return true
	default:
		return false
	}
}

// handleSessionNative handles the session tool with native Go implementation.
func handleSessionNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action := strings.TrimSpace(cast.ToString(params["action"]))
	// Allow handoff-only clients to pass sub_action (e.g. resume, list) without also setting action=handoff.
	if action == "" {
		if sub := strings.TrimSpace(cast.ToString(params["sub_action"])); sub != "" && isSessionHandoffSubAction(sub) {
			action = "handoff"
			params["action"] = "handoff"
		}
	}
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
	case "restore":
		return handleSessionRestore(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s (use 'prime', 'handoff', 'prompts', 'assignee', or 'restore')", action)
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
		if eliciter := framework.EliciterFromContext(ctx); eliciter != nil {
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

	// Resolve client identity: explicit param takes precedence over context injection.
	clientName := cast.ToString(params["client"])
	if clientName == "" {
		clientName = framework.ClientNameFromContext(ctx)
	}

	// Client-specific adjustments before building the result.
	// opencode: force compact=true to reduce token overhead.
	if clientName == "opencode" {
		if _, ok := params["compact"]; !ok {
			params["compact"] = true
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

		todoCount, _ := database.GetTaskCountByStatus(ctx, models.StatusTodo)
		if todoCount > 10 {
			hints["thinking_workflow"] = "For complex backlog analysis, sprint planning, or dependency enrichment: use the thinking-workflow skill (.cursor/skills/thinking-workflow/SKILL.md) — chain tractatus (structure) + sequential (process) + exarp-go MCP (execute)"
		}

		// Hint about ownership if tasks lack it
		if len(tasks) > 0 {
			missingOwnership := 0
			for _, task := range tasks {
				if IsPendingStatus(task.Status) && models.GetTaskOwnership(&task) == nil {
					missingOwnership++
				}
			}
			if missingOwnership > 2 {
				hints["add_ownership"] = fmt.Sprintf("⚠️ %d pending tasks lack file ownership. Add with: task_workflow update <id> owned_files=['...'] lane='...' — or run task_analysis action=infer_ownership", missingOwnership)
			}
		}
	} else if includeTasks {
		planPath, _ = getPlanModeContext(projectRoot, tasks)
	}

	handoffAlert := (map[string]interface{})(nil)
	// cursor client: suppress handoff alert (suppress noise); other clients follow include_handoff param.
	suppressHandoff := clientName == "cursor"
	if !suppressHandoff {
		if _, has := params["include_handoff"]; !has || cast.ToBool(params["include_handoff"]) {
			handoffAlert = checkHandoffAlert(projectRoot)
		}
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

	// include_cli_command defaults to false so interactive chat does not get the runnable agent command;
	// only suggested_next_action (text) is returned. Set true for CLI/TUI/scripts that may execute it.
	includeCliCommand := cast.ToBool(params["include_cli_command"])

	var suggestedNext []map[string]interface{}
	if includeTasks && tasksErr == nil {
		suggestedNext = getSuggestedNextTasksFromTasks(tasks, 5)
		if len(suggestedNext) > 0 {
			if includeCliCommand {
				if cmd := buildCursorCliSuggestion(suggestedNext[0]); cmd != "" {
					pb.CursorCliSuggestion = cmd
				}
			}
		}
	}

	result := SessionPrimeResultToMap(pb)

	activeLocks, _ := database.GetActiveLocks(ctx)
	if len(activeLocks) > 0 {
		lockMaps := make([]map[string]interface{}, 0, len(activeLocks))
		for _, lock := range activeLocks {
			lockMaps = append(lockMaps, lockToMap(lock))
		}
		result["active_claims"] = lockMaps
	}
	activeRuns, _ := database.ListTaskExecutionRuns(ctx, "", "running", 10)
	if len(activeRuns) > 0 {
		runMaps := make([]map[string]interface{}, 0, len(activeRuns))
		for i := range activeRuns {
			runMaps = append(runMaps, runToMap(&activeRuns[i]))
		}
		result["active_runs"] = runMaps
	}

	// generic client: minimal output — tasks + mode only, no hints, no handoff noise.
	if clientName == "generic" {
		includeHints = false
		handoffAlert = nil
	}

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
				if ln, ok := suggestedNext[0]["lane"].(string); ok && ln != "" {
					result["suggested_lane"] = ln
				}

				// Add ownership collision warnings for suggested tasks
				suggestedTaskIDs := make([]string, 0, len(suggestedNext))
				for _, st := range suggestedNext {
					if id, ok := st["id"].(string); ok {
						suggestedTaskIDs = append(suggestedTaskIDs, id)
					}
				}

				suggestedTaskObjs := make([]Todo2Task, 0, len(suggestedTaskIDs))
				for _, task := range tasks {
					for _, sid := range suggestedTaskIDs {
						if task.ID == sid {
							suggestedTaskObjs = append(suggestedTaskObjs, task)
							break
						}
					}
				}

				if ownershipHints := buildOwnershipHints(suggestedTaskObjs); len(ownershipHints) > 0 {
					result["ownership_warnings"] = ownershipHints
				}
			}

			// Add hotspot summary: files contested by multiple pending tasks
			hotspotSummary := buildHotspotSummary(tasks)
			if len(hotspotSummary) > 0 {
				result["hotspot_summary"] = hotspotSummary
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

	// Record resolved client identity in metadata.
	if clientName != "" {
		result["client"] = clientName
	}

	AddTokenEstimateToResult(result)

	// Auto-compaction ledger: if context_threshold_pct is set (1–100), compare current token
	// usage against budget and auto-write a CONTINUITY ledger when the threshold is exceeded.
	if thresholdPct := cast.ToFloat64(params["context_threshold_pct"]); thresholdPct > 0 {
		currentTokens := cast.ToInt(params["current_tokens"])
		if currentTokens == 0 {
			// Fall back to this response's own token estimate as a conservative proxy.
			currentTokens, _ = result["token_estimate"].(int)
		}
		budgetTokens := cast.ToInt(params["budget_tokens"])
		if budgetTokens == 0 {
			budgetTokens = config.DefaultContextBudget()
		}
		if currentTokens > 0 && budgetTokens > 0 {
			usagePct := float64(currentTokens) / float64(budgetTokens) * 100
			if usagePct >= thresholdPct {
				if ledgerPath, err := writeCompactionLedger(ctx, projectRoot, params); err == nil {
					result["ledger_written"] = true
					result["ledger_path"] = ledgerPath
					result["ledger_trigger_pct"] = usagePct
				}
			}
		}
	}

	// inject_ledger: include the latest compaction ledger in the prime result so the next
	// session can read it without a separate call. Useful after context compaction.
	if cast.ToBool(params["inject_ledger"]) {
		if content, ledgerPath := readLatestLedger(projectRoot); content != "" {
			result["latest_ledger"] = content
			result["latest_ledger_path"] = ledgerPath
		}
	}

	// Default compact=true for MCP callers to reduce token overhead; pass compact=false to opt out
	compact := ParamBool(params, "compact", true)
	return FormatResultOptionalCompact(result, "", compact)
}

// handleSessionHandoff handles handoff actions (end, resume, latest, list, sync, export).
