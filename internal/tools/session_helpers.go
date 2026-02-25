// session_helpers.go — Session helpers: types, agent/mode detection, hints, task summaries, and session status.
// See also: session_helpers_handoff.go
package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/models"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   AgentInfo
//   AgentContext
//   TimeSuggestion
//   detectAgentType — detectAgentType detects the current agent type.
//   getAgentContext — getAgentContext returns context for an agent type.
//   suggestModeByTime — suggestModeByTime suggests workflow mode based on time of day.
//   getWorkflowModeDescription — getWorkflowModeDescription returns description for a workflow mode.
//   getPlanModeContext — getPlanModeContext returns (plan_path, plan_mode_hint) for session prime.
//   getCurrentPlanPath — getCurrentPlanPath returns the relative path to the primary .plan.md file if one exists.
//   shouldSuggestPlanMode — shouldSuggestPlanMode returns true when backlog suggests complex or multi-step work.
//   getHintsForMode — getHintsForMode returns tool hints for a mode (simplified).
//   getTasksSummaryFromTasks — getTasksSummaryFromTasks returns a summary from already-loaded tasks (avoids duplicate load).
//   getSuggestedNextTasksFromTasks — getSuggestedNextTasksFromTasks returns first N backlog tasks in dependency order (avoids duplicate load).
//   GetSessionStatus — GetSessionStatus returns current session context for UI status display (e.g. Cursor when it supports dynamic labels).
// ────────────────────────────────────────────────────────────────────────────

// ─── AgentInfo ──────────────────────────────────────────────────────────────
type AgentInfo struct {
	Agent  string
	Source string
	Config map[string]interface{}
}

// ─── AgentContext ───────────────────────────────────────────────────────────
type AgentContext struct {
	Agent           string
	RecommendedMode string
	FocusAreas      []string
	RelevantTools   []string
}

// ─── TimeSuggestion ─────────────────────────────────────────────────────────
type TimeSuggestion struct {
	Mode       string
	Reason     string
	Confidence float64
}

// ─── detectAgentType ────────────────────────────────────────────────────────
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

// ─── getAgentContext ────────────────────────────────────────────────────────
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

// ─── suggestModeByTime ──────────────────────────────────────────────────────
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

// ─── getWorkflowModeDescription ─────────────────────────────────────────────
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

// ─── getPlanModeContext ─────────────────────────────────────────────────────
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

// ─── getCurrentPlanPath ─────────────────────────────────────────────────────
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

// ─── shouldSuggestPlanMode ──────────────────────────────────────────────────
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

		if t.Priority == models.PriorityHigh || t.Priority == models.PriorityCritical {
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

// ─── getHintsForMode ────────────────────────────────────────────────────────
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

// ─── getTasksSummaryFromTasks ───────────────────────────────────────────────
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

// ─── getSuggestedNextTasksFromTasks ─────────────────────────────────────────
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

// ─── GetSessionStatus ───────────────────────────────────────────────────────
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
