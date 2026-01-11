package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
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
		// Prompts action depends on prompt discovery resource - fall back to Python bridge
		return nil, fmt.Errorf("prompts action requires prompt discovery resource, falling back to Python bridge")
	case "assignee":
		// Assignee action depends on task assignee logic - fall back to Python bridge for now
		return nil, fmt.Errorf("assignee action requires task assignee logic, falling back to Python bridge")
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
		"auto_primed":  true,
		"method":       "native_go",
		"timestamp":    time.Now().Format(time.RFC3339),
		"duration_ms":  time.Since(startTime).Milliseconds(),
		"detection": map[string]interface{}{
			"agent":       agentInfo.Agent,
			"agent_source": agentInfo.Source,
			"mode":        mode,
			"mode_source": modeSource,
			"time_of_day": time.Now().Format("15:04"),
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

	// 4. Add tasks summary if requested
	if includeTasks {
		tasksSummary := getTasksSummary(projectRoot)
		result["tasks"] = tasksSummary
	}

	// 5. Add hints if requested (simplified)
	if includeHints {
		hints := getHintsForMode(mode)
		result["hints"] = hints
		result["hints_count"] = len(hints)
	}

	// 6. Check for handoff notes if requested
	if includeHandoff, ok := params["include_handoff"].(bool); !ok || includeHandoff {
		if handoffAlert := checkHandoffAlert(projectRoot); handoffAlert != nil {
			result["handoff_alert"] = handoffAlert
			result["action_required"] = "📋 Review handoff from previous developer before starting work"
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleSessionHandoff handles handoff actions (end, resume, latest, list, sync)
func handleSessionHandoff(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	// Note: The action parameter for handoff might be nested - check both
	if action == "" {
		// Check if this is called from session tool with action="handoff"
		if summary, ok := params["summary"].(string); ok && summary != "" {
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
	default:
		return nil, fmt.Errorf("unknown handoff action: %s (use 'end', 'resume', 'latest', 'list', or 'sync')", action)
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
		"id":        fmt.Sprintf("handoff-%d", time.Now().Unix()),
		"timestamp": time.Now().Format(time.RFC3339),
		"host":      hostname,
		"summary":   summary,
		"blockers":  blockers,
		"next_steps": nextSteps,
		"git_status": gitStatus,
		"tasks_in_progress": tasksInProgress,
	}

	if !dryRun {
		// Save handoff
		if err := saveHandoff(projectRoot, handoff); err != nil {
			return nil, fmt.Errorf("failed to save handoff: %w", err)
		}

		// Unassign tasks if requested
		if unassignMyTasks {
			// TODO: Implement task unassignment using Todo2 utilities
			// For now, just log that it should be done
		}
	}

	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"dry_run":     dryRun,
		"handoff":     handoff,
		"message":     "Session ended. Handoff note created.",
	}

	if dryRun {
		result["message"] = "Dry run: Would create handoff note"
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleSessionResume resumes a session by reviewing latest handoff
func handleSessionResume(ctx context.Context, projectRoot string) ([]framework.TextContent, error) {
	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")

	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		result := map[string]interface{}{
			"success":      true,
			"method":       "native_go",
			"has_handoff":  false,
			"message":      "No handoff notes found. Starting fresh session.",
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}

	// Load handoff history
	data, err := os.ReadFile(handoffFile)
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
		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
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
		"success":     true,
		"method":      "native_go",
		"has_handoff": true,
		"handoff":     handoffMap,
		"from_same_host": handoffHost == hostname,
		"message":     fmt.Sprintf("Resuming session. Latest handoff from %s", handoffHost),
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
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
		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}

	data, err := os.ReadFile(handoffFile)
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
		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}

	latestHandoff := handoffs[len(handoffs)-1]

	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"has_handoff": true,
		"handoff":     latestHandoff,
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
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
		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}

	data, err := os.ReadFile(handoffFile)
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

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
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
		"success":  true,
		"method":   "native_go",
		"direction": direction,
		"dry_run":  dryRun,
		"message":  "Todo2 state sync via Git (simplified implementation)",
		"note":     "Full sync implementation requires agentic-tools MCP integration",
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

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// Helper types and functions

type AgentInfo struct {
	Agent  string
	Source string
	Config map[string]interface{}
}

type AgentContext struct {
	Agent          string
	RecommendedMode string
	FocusAreas     []string
	RelevantTools  []string
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
	agentConfigPath := filepath.Join(projectRoot, "cursor-agent.json")
	if data, err := os.ReadFile(agentConfigPath); err == nil {
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
		hints["lint"] = "Use lint tool for code quality"
	case "task_management":
		hints["task_workflow"] = "Use task_workflow for task operations"
	}

	return hints
}

// getTasksSummary returns a summary of tasks
func getTasksSummary(projectRoot string) map[string]interface{} {
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return map[string]interface{}{
			"error": "Failed to load tasks",
		}
	}

	byStatus := make(map[string]int)
	for _, task := range tasks {
		byStatus[task.Status]++
	}

	// Get recent tasks
	recentTasks := []map[string]interface{}{}
	for i, task := range tasks {
		if i >= 10 { // Limit to 10 most recent
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
		"total":      len(tasks),
		"by_status":  byStatus,
		"recent":     recentTasks,
	}
}

// checkHandoffAlert checks for handoff notes from other developers
func checkHandoffAlert(projectRoot string) map[string]interface{} {
	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")
	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(handoffFile)
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
			"from_host": handoffMap["host"],
			"timestamp": handoffMap["timestamp"],
			"summary":   truncateString(fmt.Sprintf("%v", handoffMap["summary"]), 100),
			"blockers":  handoffMap["blockers"],
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
	handoffs := []interface{}{}
	if data, err := os.ReadFile(handoffFile); err == nil {
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
