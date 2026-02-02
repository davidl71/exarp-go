package tools

import (
	"context"
	"database/sql"
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
	mcpframework "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	mcpresponse "github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// handleSessionNative handles the session tool with native Go implementation
func handleSessionNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
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

// handleSessionPrime handles the prime action - auto-prime AI context at session start
func handleSessionPrime(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()

	includeHints := true
	if hints, ok := params["include_hints"].(bool); ok {
		includeHints = hints
	}

	includeTasks := true
	if tasks, ok := params["include_tasks"].(bool); ok {
		includeTasks = tasks
	}

	// Optional MCP Elicitation: ask user for prime preferences when ask_preferences is true.
	// Use a short timeout so prime never blocks indefinitely if the client is slow or doesn't respond.
	const elicitationTimeout = 5 * time.Second
	var elicitationOutcome string
	if ask, _ := params["ask_preferences"].(bool); ask {
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

	overrideMode, _ := params["override_mode"].(string)

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
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

	// 3. Build result
	result := map[string]interface{}{
		"auto_primed": true,
		"method":      "native_go",
		"timestamp":   time.Now().Format(time.RFC3339),
		"duration_ms": time.Since(startTime).Milliseconds(),
		"detection": map[string]interface{}{
			"agent":        agentInfo.Agent,
			"agent_source": agentInfo.Source,
			"mode":         mode,
			"mode_source":  modeSource,
			"time_of_day":  time.Now().Format("15:04"),
		},
		"agent_context": map[string]interface{}{
			"focus_areas":      agentContext.FocusAreas,
			"relevant_tools":   agentContext.RelevantTools,
			"recommended_mode": agentContext.RecommendedMode,
		},
		"workflow": map[string]interface{}{
			"mode":        mode,
			"description": getWorkflowModeDescription(mode),
		},
	}
	if elicitationOutcome != "" {
		result["elicitation"] = elicitationOutcome
	}

	// 4. Load tasks when needed (summary, suggested_next, or plan mode context)
	var tasks []Todo2Task
	var tasksErr error
	// Load tasks if we need them for task summary, hints, or plan mode context
	if includeTasks || includeHints {
		tasks, tasksErr = LoadTodo2Tasks(projectRoot)
	}
	if includeTasks {
		if tasksErr != nil {
			result["tasks"] = map[string]interface{}{"error": "Failed to load tasks"}
		} else {
			result["tasks"] = getTasksSummaryFromTasks(tasks)
			suggestedNext := getSuggestedNextTasksFromTasks(tasks, 10)
			if len(suggestedNext) > 0 {
				result["suggested_next"] = suggestedNext
			}
		}
	}

	// 5. Add hints if requested (simplified)
	if includeHints {
		hints := getHintsForMode(mode)
		// Add Plan Mode hint when backlog suggests complex/multi-step work
		planPath, planModeHint := getPlanModeContext(projectRoot, tasks)
		if planPath != "" {
			result["plan_path"] = planPath
		}
		if planModeHint != "" {
			hints["plan_mode"] = planModeHint
		}
		result["hints"] = hints
		result["hints_count"] = len(hints)
	}

	// 6. Add plan path when tasks included but hints not requested
	// Note: tasks are already loaded when includeTasks is true (see step 4)
	if includeTasks && !includeHints {
		if planPath, _ := getPlanModeContext(projectRoot, tasks); planPath != "" {
			result["plan_path"] = planPath
		}
	}

	// 7. Check for handoff notes if requested
	if includeHandoff, ok := params["include_handoff"].(bool); !ok || includeHandoff {
		if handoffAlert := checkHandoffAlert(projectRoot); handoffAlert != nil {
			result["handoff_alert"] = handoffAlert
			result["action_required"] = "📋 Review handoff from previous developer before starting work"
		}
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionHandoff handles handoff actions (end, resume, latest, list, sync, export)
func handleSessionHandoff(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	// Note: The action parameter for handoff might be nested - check both
	// Also check for sub_action parameter (for nested actions like export)
	if action == "" || action == "handoff" {
		// Check for sub_action first (for explicit sub-actions)
		if subAction, ok := params["sub_action"].(string); ok && subAction != "" {
			action = subAction
		} else if summary, ok := params["summary"].(string); ok && summary != "" {
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
	default:
		return nil, fmt.Errorf("unknown handoff action: %s (use 'end', 'resume', 'latest', 'list', 'sync', or 'export')", action)
	}
}

// handleSessionEnd ends a session and creates handoff note
func handleSessionEnd(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	summary, _ := params["summary"].(string)
	blockersRaw, _ := params["blockers"]
	nextStepsRaw, _ := params["next_steps"]

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
	if unassign, ok := params["unassign_my_tasks"].(bool); ok {
		unassignMyTasks = unassign
	}

	includeGitStatus := true
	if gitStatus, ok := params["include_git_status"].(bool); ok {
		includeGitStatus = gitStatus
	}

	dryRun := false
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	// Get hostname
	hostname, _ := os.Hostname()

	// Get Git status if requested
	var gitStatus map[string]interface{}
	if includeGitStatus {
		gitStatus = getGitStatus(ctx, projectRoot)
	}

	// Get current tasks in progress
	var tasksInProgress []map[string]interface{}
	if includeTasks, ok := params["include_tasks"].(bool); !ok || includeTasks {
		tasks, err := LoadTodo2Tasks(projectRoot)
		if err == nil {
			for _, task := range tasks {
				if task.Status == "In Progress" {
					tasksInProgress = append(tasksInProgress, map[string]interface{}{
						"id":      task.ID,
						"content": task.Content,
						"status":  task.Status,
					})
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

	return mcpresponse.FormatResult(result, "")
}

// handleSessionResume resumes a session by reviewing latest handoff
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

	return mcpresponse.FormatResult(result, "")
}

// handleSessionLatest gets the most recent handoff note
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

// handleSessionList lists recent handoff notes
func handleSessionList(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	limit := 5
	if l, ok := params["limit"].(float64); ok {
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

// handleSessionSync syncs Todo2 state across agents
func handleSessionSync(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	direction := "both"
	if d, ok := params["direction"].(string); ok && d != "" {
		direction = d
	}

	autoCommit := true
	if ac, ok := params["auto_commit"].(bool); ok {
		autoCommit = ac
	}

	dryRun := false
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
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

	// Basic Git sync implementation
	if !dryRun {
		if direction == "pull" || direction == "both" {
			// Pull latest changes
			cmd := exec.CommandContext(ctx, "git", "pull", "--no-edit")
			cmd.Dir = projectRoot
			if err := cmd.Run(); err == nil {
				result["pulled"] = true
			}
		}

		if (direction == "push" || direction == "both") && autoCommit {
			// Check if there are changes to commit
			cmd := exec.CommandContext(ctx, "git", "status", "--porcelain", ".todo2")
			cmd.Dir = projectRoot
			output, err := cmd.Output()
			if err == nil && len(output) > 0 {
				// Commit changes
				cmd = exec.CommandContext(ctx, "git", "add", ".todo2")
				cmd.Dir = projectRoot
				cmd.Run()

				cmd = exec.CommandContext(ctx, "git", "commit", "-m", "Auto-sync Todo2 state")
				cmd.Dir = projectRoot
				cmd.Run()

				result["committed"] = true
			}

			// Push changes
			cmd = exec.CommandContext(ctx, "git", "push")
			cmd.Dir = projectRoot
			if err := cmd.Run(); err == nil {
				result["pushed"] = true
			}
		}
	}

	return mcpresponse.FormatResult(result, "")
}

// handleSessionExport exports handoff data to a JSON file for sharing between agents
func handleSessionExport(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	outputPath, _ := params["output_path"].(string)
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
	if latest, ok := params["export_latest"].(bool); ok {
		exportLatest = latest
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

// detectAgentType detects the current agent type
func detectAgentType(projectRoot string) AgentInfo {
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

// getAgentContext returns context for an agent type
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
			RelevantTools:   []string{"run_tests", "generate_cursorignore"},
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

// suggestModeByTime suggests workflow mode based on time of day
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

// getWorkflowModeDescription returns description for a workflow mode
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
// plan_mode_hint: suggestion to use Cursor Plan Mode when backlog suggests complex/multi-step work.
func getPlanModeContext(projectRoot string, tasks []Todo2Task) (planPath, planModeHint string) {
	planPath = getCurrentPlanPath(projectRoot)
	if shouldSuggestPlanMode(tasks) {
		planModeHint = "Consider Cursor Plan Mode for multi-step work when backlog has many high-priority or dependency-heavy tasks"
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

// getHintsForMode returns tool hints for a mode (simplified)
func getHintsForMode(mode string) map[string]string {
	// Simplified hints - can be expanded
	hints := map[string]string{
		"session": "Use session tool for context priming and handoffs",
		"tasks":   "Use task_workflow tool for task management",
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
			out = append(out, map[string]interface{}{"id": id, "content": ""})
			continue
		}
		out = append(out, map[string]interface{}{
			"id":       d.ID,
			"content":  d.Content,
			"priority": d.Priority,
			"level":    d.Level,
		})
	}
	return out
}

// checkHandoffAlert checks for handoff notes from other developers
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

// saveHandoff saves a handoff note to the handoffs.json file
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

// getGitStatus gets current Git status
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

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// handleSessionPrompts handles the prompts action - lists available prompts
func handleSessionPrompts(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get optional filters
	mode, _ := params["mode"].(string)
	persona, _ := params["persona"].(string)
	category, _ := params["category"].(string)
	keywords, _ := params["keywords"].(string)
	limit := 50
	if l, ok := params["limit"].(float64); ok {
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

// handleSessionAssignee handles the assignee action - manages task assignments
func handleSessionAssignee(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	subAction, _ := params["sub_action"].(string)
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

// handleSessionAssigneeList lists tasks with their assignees
func handleSessionAssigneeList(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	includeUnassigned := true
	if unassigned, ok := params["include_unassigned"].(bool); ok {
		includeUnassigned = unassigned
	}

	statusFilter, _ := params["status_filter"].(string)

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

// handleSessionAssigneeAssign assigns a task to an agent/human/host
func handleSessionAssigneeAssign(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID, _ := params["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id parameter is required")
	}

	assigneeName, _ := params["assignee_name"].(string)
	if assigneeName == "" {
		return nil, fmt.Errorf("assignee_name parameter is required")
	}

	assigneeType := "agent"
	if t, ok := params["assignee_type"].(string); ok && t != "" {
		assigneeType = t
	}

	dryRun := false
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
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
		return nil, fmt.Errorf("task assignment failed: %v", result.Error)
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

// handleSessionAssigneeUnassign unassigns a task
func handleSessionAssigneeUnassign(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID, _ := params["task_id"].(string)
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
