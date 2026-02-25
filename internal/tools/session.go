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
	mcpframework "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
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
			if t.Status == models.StatusTodo {
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

	// include_cli_command defaults to false so interactive chat does not get the runnable agent command;
	// only suggested_next_action (text) is returned. Set true for CLI/TUI/scripts that may execute it.
	includeCliCommand := cast.ToBool(params["include_cli_command"])

	var suggestedNext []map[string]interface{}
	if includeTasks && tasksErr == nil {
		suggestedNext = getSuggestedNextTasksFromTasks(tasks, 10)
		if len(suggestedNext) > 0 {
			if includeCliCommand {
				if cmd := buildCursorCliSuggestion(suggestedNext[0]); cmd != "" {
					pb.CursorCliSuggestion = cmd
				}
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

	AddTokenEstimateToResult(result)
	compact := cast.ToBool(params["compact"])
	return FormatResultOptionalCompact(result, "", compact)
}

// handleSessionHandoff handles handoff actions (end, resume, latest, list, sync, export).
