// task_analysis_deps.go — Task analysis: dependency, summary, execution-plan handlers, formatters, and fmtTime.
// See also: task_analysis_deps_analysis.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleTaskAnalysisDependencies
//   handleTaskAnalysisDependenciesSummary — handleTaskAnalysisDependenciesSummary combines dependencies, parallelization, and execution_plan (T-227).
//   handleTaskAnalysisExecutionPlan — handleTaskAnalysisExecutionPlan handles execution plan: backlog (Todo + In Progress) in dependency order.
//   formatExecutionPlanText
//   formatExecutionPlanMarkdown
//   fmtTime
// ────────────────────────────────────────────────────────────────────────────

// ─── handleTaskAnalysisDependencies ─────────────────────────────────────────
func handleTaskAnalysisDependencies(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	cycles, missing, err := GetDependencyAnalysisFromTasks(tasks)
	if err != nil {
		return nil, err
	}

	// Build legacy graph format for backward compatibility
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	graph := buildLegacyGraphFormat(tg)

	// Calculate critical path from backlog only (exclude Done by default)
	var criticalPath []string

	var criticalPathDetails []map[string]interface{}

	maxLevel := 0

	tgBacklog, err := BuildTaskGraphBacklogOnly(tasks)
	if err == nil && tgBacklog.Graph.Nodes().Len() > 0 {
		hasCycles, err := HasCycles(tgBacklog)
		if err == nil && !hasCycles {
			// Find critical path among backlog tasks
			path, err := FindCriticalPath(tgBacklog)
			if err == nil {
				criticalPath = path

				// Build detailed path information
				for _, taskID := range path {
					for _, task := range tasks {
						if task.ID == taskID {
							criticalPathDetails = append(criticalPathDetails, map[string]interface{}{
								"id":                 task.ID,
								"content":            task.Content,
								"priority":           task.Priority,
								"status":             task.Status,
								"dependencies":       task.Dependencies,
								"dependencies_count": len(task.Dependencies),
							})

							break
						}
					}
				}
			}

			// Get max dependency level from backlog graph
			levels := GetTaskLevels(tgBacklog)
			for _, level := range levels {
				if level > maxLevel {
					maxLevel = level
				}
			}
		}
	}

	outputFormat := "json"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	result := map[string]interface{}{
		"success":               true,
		"method":                "native_go",
		"total_tasks":           len(tasks),
		"dependency_graph":      graph,
		"circular_dependencies": cycles,
		"missing_dependencies":  missing,
		"recommendations":       buildDependencyRecommendations(graph, cycles, missing),
	}

	// Add critical path information if available
	if len(criticalPath) > 0 {
		result["critical_path"] = criticalPath
		result["critical_path_length"] = len(criticalPath)
		result["critical_path_details"] = criticalPathDetails
		result["max_dependency_level"] = maxLevel
	}

	// Include human-readable report in JSON for CLI/consumers
	result["report"] = formatDependencyAnalysisText(result)

	projectRoot, err := GetProjectRootWithFallback()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project root: %w", err)
	}
	outputPath := DefaultReportOutputPath(projectRoot, "TASK_ANALYSIS_DEPENDENCIES.md", params)
	if outputFormat == "json" {
		if outputPath != "" {
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create output dir: %w", err)
			}
		}

		resultJSON, _ := json.Marshal(result)
		resp := &proto.TaskAnalysisResponse{Action: "dependencies", OutputPath: outputPath, ResultJson: string(resultJSON)}

		return framework.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	output := formatDependencyAnalysisText(result)

	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}

		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}

		output += fmt.Sprintf("\n\n[Saved to: %s]", outputPath)
	}

	return []framework.TextContent{{Type: "text", Text: output}}, nil
}

// ─── handleTaskAnalysisDependenciesSummary ──────────────────────────────────
// handleTaskAnalysisDependenciesSummary combines dependencies, parallelization, and execution_plan (T-227).
func handleTaskAnalysisDependenciesSummary(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	outputFormat := ParamString(params, "output_format")
	if outputFormat == "" {
		outputFormat = "text"
	}

	if outputFormat == "json" {
		jsonParams := make(map[string]interface{}, len(params)+1)
		for k, v := range params {
			jsonParams[k] = v
		}
		jsonParams["output_format"] = "json"

		deps, err := handleTaskAnalysisDependencies(ctx, jsonParams)
		if err != nil {
			return nil, err
		}
		par, err := handleTaskAnalysisParallelization(ctx, jsonParams)
		if err != nil {
			return nil, err
		}
		plan, err := handleTaskAnalysisExecutionPlan(ctx, jsonParams)
		if err != nil {
			return nil, err
		}

		depsData, err := parseTaskAnalysisJSONContent(deps)
		if err != nil {
			return nil, fmt.Errorf("dependencies_summary dependencies: %w", err)
		}
		parData, err := parseTaskAnalysisJSONContent(par)
		if err != nil {
			return nil, fmt.Errorf("dependencies_summary parallelization: %w", err)
		}
		planData, err := parseTaskAnalysisJSONContent(plan)
		if err != nil {
			return nil, fmt.Errorf("dependencies_summary execution_plan: %w", err)
		}

		reportParts := []string{}
		if report := ParamString(depsData, "report"); report != "" {
			reportParts = append(reportParts, "## Dependency Analysis\n"+report)
		}
		if report := ParamString(parData, "report"); report != "" {
			reportParts = append(reportParts, "## Parallelization\n"+report)
		}
		reportParts = append(reportParts, "## Execution Plan\n"+formatExecutionPlanText(planData))

		result := map[string]interface{}{
			"success":         true,
			"method":          "native_go",
			"action":          "dependencies_summary",
			"dependencies":    depsData,
			"parallelization": parData,
			"execution_plan":  planData,
			"report":          "# Task Dependencies Summary\n\n" + strings.Join(reportParts, "\n\n"),
		}

		projectRoot, err := GetProjectRootWithFallback()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project root: %w", err)
		}
		outputPath := DefaultReportOutputPath(projectRoot, "TASK_ANALYSIS_DEPENDENCIES_SUMMARY.json", params)
		return framework.FormatResult(result, outputPath)
	}

	var parts []string

	deps, err := handleTaskAnalysisDependencies(ctx, params)
	if err == nil && len(deps) > 0 {
		parts = append(parts, "## Dependency Analysis\n"+deps[0].Text)
	}

	par, err := handleTaskAnalysisParallelization(ctx, params)
	if err == nil && len(par) > 0 {
		parts = append(parts, "## Parallelization\n"+par[0].Text)
	}

	plan, err := handleTaskAnalysisExecutionPlan(ctx, params)
	if err == nil && len(plan) > 0 {
		parts = append(parts, "## Execution Plan\n"+plan[0].Text)
	}

	report := "# Task Dependencies Summary\n\n" + strings.Join(parts, "\n\n")

	return []framework.TextContent{{Type: "text", Text: report}}, nil
}

func parseTaskAnalysisJSONContent(contents []framework.TextContent) (map[string]interface{}, error) {
	if len(contents) == 0 {
		return nil, fmt.Errorf("empty tool result")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(contents[0].Text), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON result: %w", err)
	}

	return data, nil
}

// ─── handleTaskAnalysisExecutionPlan ────────────────────────────────────────
// handleTaskAnalysisExecutionPlan handles execution plan: backlog (Todo + In Progress) in dependency order.
func handleTaskAnalysisExecutionPlan(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := GetProjectRootWithFallback()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	// Optional tag filter: restrict backlog to tasks with filter_tag or any of filter_tags
	var backlogFilter map[string]bool
	if ft, ok := params["filter_tag"].(string); ok && ft != "" {
		backlogFilter = make(map[string]bool)

		for _, t := range tasks {
			if !IsBacklogStatus(t.Status) {
				continue
			}

			for _, tag := range t.Tags {
				if tag == ft {
					backlogFilter[t.ID] = true
					break
				}
			}
		}
	} else if fts, ok := params["filter_tags"].(string); ok && fts != "" {
		allowed := strings.Split(fts, ",")
		for i := range allowed {
			allowed[i] = strings.TrimSpace(allowed[i])
		}

		backlogFilter = make(map[string]bool)

		for _, t := range tasks {
			if !IsBacklogStatus(t.Status) {
				continue
			}

			for _, tag := range t.Tags {
				for _, a := range allowed {
					if a != "" && tag == a {
						backlogFilter[t.ID] = true
						break
					}
				}

				if backlogFilter[t.ID] {
					break
				}
			}
		}
	}

	orderedIDs, waves, details, err := BacklogExecutionOrder(tasks, backlogFilter)
	if err != nil {
		return nil, fmt.Errorf("execution order: %w", err)
	}

	// Optional limit (0 = all)
	limit := 0
	if l, ok := params["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	if limit > 0 && len(orderedIDs) > limit {
		orderedIDs = orderedIDs[:limit]
		// Trim details to match
		if len(details) > limit {
			details = details[:limit]
		}
	}

	topTasks := buildCodexTopTasks(details, waves)
	suggestedNextAction := buildExecutionPlanSuggestedNextAction(topTasks)
	agentHint := buildExecutionPlanAgentHint(suggestedNextAction)
	summary := buildExecutionPlanSummary(len(orderedIDs), topTasks)

	// Detect file collisions between tasks with ownership metadata
	collisions := DetectFileCollisions(tasks)
	collisionMap := BuildCollisionMap(collisions)

	result := map[string]interface{}{
		"success":               true,
		"method":                "native_go",
		"backlog_count":         len(orderedIDs),
		"ordered_task_ids":      orderedIDs,
		"waves":                 waves,
		"details":               details,
		"top_tasks":             topTasks,
		"suggested_next_action": suggestedNextAction,
		"agent_hint":            agentHint,
		"summary":               summary,
		"file_collisions":       collisions,
		"collision_map":         collisionMap,
		"has_collisions":        len(collisions) > 0,
	}

	outputFormat := "json"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	outputPath := cast.ToString(params["output_path"])

	// subagents_plan: write parallel-execution-subagents.plan.md using wave detection
	if outputFormat == "subagents_plan" {
		wavesCopy := waves
		if max := config.MaxTasksPerWave(); max > 0 {
			wavesCopy = LimitWavesByMaxTasks(wavesCopy, max)
		}

		if len(wavesCopy) == 0 {
			return nil, fmt.Errorf("no waves (empty backlog or no Todo/In Progress tasks)")
		}

		planTitle := cast.ToString(params["plan_title"])
		if planTitle == "" {
			planTitle = filepath.Base(projectRoot)

			if info, err := getProjectInfo(projectRoot); err == nil {
				if name, ok := info["name"].(string); ok && name != "" {
					planTitle = name
				}
			}
		}

		if outputPath == "" {
			outputPath = DefaultPlanOutputPath(projectRoot, "parallel-execution-subagents.plan.md", params)
		}

		if dir := filepath.Dir(outputPath); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create plan directory: %w", err)
			}
		}

		md := FormatWavesAsSubagentsPlanMarkdown(wavesCopy, planTitle)
		if err := os.WriteFile(outputPath, []byte(md), 0644); err != nil {
			return nil, fmt.Errorf("failed to write subagents plan: %w", err)
		}

		msg := fmt.Sprintf("Parallel execution subagents plan saved to: %s", outputPath)

		return []framework.TextContent{{Type: "text", Text: msg}}, nil
	}

	if outputFormat == "json" {
		if outputPath != "" {
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create output dir: %w", err)
			}
		}

		resultJSON, _ := json.Marshal(result)
		resp := &proto.TaskAnalysisResponse{Action: "execution_plan", OutputPath: outputPath, ResultJson: string(resultJSON)}

		return framework.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	output := formatExecutionPlanText(result)

	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}

		if strings.HasSuffix(strings.ToLower(outputPath), ".md") {
			if !strings.HasSuffix(strings.ToLower(outputPath), ".plan.md") {
				outputPath = outputPath[:len(outputPath)-3] + ".plan.md"
			}

			md := formatExecutionPlanMarkdown(result, projectRoot)
			if err := os.WriteFile(outputPath, []byte(md), 0644); err != nil {
				return nil, fmt.Errorf("failed to save markdown: %w", err)
			}
		} else {
			if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
				return nil, fmt.Errorf("failed to save result: %w", err)
			}
		}

		output += fmt.Sprintf("\n\n[Saved to: %s]", outputPath)
	}

	return []framework.TextContent{{Type: "text", Text: output}}, nil
}

// ─── formatExecutionPlanText ────────────────────────────────────────────────
func formatExecutionPlanText(result map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Backlog execution order\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	if summary := ParamString(result, "summary"); summary != "" {
		sb.WriteString(summary + "\n")
	}
	if next := ParamString(result, "suggested_next_action"); next != "" {
		sb.WriteString("Next: " + next + "\n")
	}

	// Collision warnings
	if hasCollisions, ok := result["has_collisions"].(bool); ok && hasCollisions {
		sb.WriteString("\n⚠️  File Collision Warnings:\n")
		if collisions, ok := result["file_collisions"].([]TaskCollision); ok {
			for _, c := range collisions {
				sb.WriteString(fmt.Sprintf("  - %s ↔ %s [%s]", c.TaskA, c.TaskB, c.Risk))
				if len(c.Files) > 0 {
					sb.WriteString(fmt.Sprintf(" (files: %s)", strings.Join(c.Files, ", ")))
				}
				if c.LaneA != "" && c.LaneB != "" && c.LaneA == c.LaneB {
					sb.WriteString(fmt.Sprintf(" (same lane: %s)", c.LaneA))
				}
				sb.WriteString("\n")
			}
		}
	}
	sb.WriteString("\n")

	if ids, ok := result["ordered_task_ids"].([]string); ok {
		for i, id := range ids {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, id))
		}
	}

	return sb.String()
}

// ─── formatExecutionPlanMarkdown ────────────────────────────────────────────
func formatExecutionPlanMarkdown(result map[string]interface{}, projectRoot string) string {
	var sb strings.Builder

	sb.WriteString("# Backlog Execution Plan\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", fmtTime(time.Now())))

	if count, ok := result["backlog_count"].(int); ok {
		sb.WriteString(fmt.Sprintf("**Backlog:** %d tasks (Todo + In Progress)\n\n", count))
	}

	if w, ok := result["waves"].(map[int][]string); ok && len(w) > 0 {
		sb.WriteString(fmt.Sprintf("**Waves:** %d dependency levels\n\n", len(w)))

		details, _ := result["details"].([]BacklogTaskDetail)

		levelOrder := make([]int, 0, len(w))
		for k := range w {
			levelOrder = append(levelOrder, k)
		}

		sort.Ints(levelOrder)

		for _, level := range levelOrder {
			ids := w[level]
			sb.WriteString(fmt.Sprintf("## Wave %d\n\n", level))
			sb.WriteString("| ID | Content | Priority | Tags |\n")
			sb.WriteString("|----|--------|----------|------|\n")

			for _, id := range ids {
				for _, d := range details {
					if d.ID == id {
						content := d.Content
						if len(content) > 60 {
							content = content[:57] + "..."
						}

						tagsStr := strings.Join(d.Tags, ", ")
						if tagsStr == "" {
							tagsStr = "-"
						}

						sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", d.ID, content, d.Priority, tagsStr))

						break
					}
				}
			}

			sb.WriteString("\n")
		}
	}

	sb.WriteString("## Full order\n\n")

	if ids, ok := result["ordered_task_ids"].([]string); ok {
		sb.WriteString(strings.Join(ids, ", "))
		sb.WriteString("\n")
	}

	// Collision warnings in markdown
	if hasCollisions, ok := result["has_collisions"].(bool); ok && hasCollisions {
		sb.WriteString("\n## ⚠️ File Collision Warnings\n\n")
		if collisions, ok := result["file_collisions"].([]TaskCollision); ok {
			sb.WriteString("| Task A | Task B | Risk | Files | Lane |\n")
			sb.WriteString("|--------|--------|------|-------|------|\n")
			for _, c := range collisions {
				files := strings.Join(c.Files, ", ")
				if files == "" {
					files = "-"
				}
				lane := "-"
				if c.LaneA != "" && c.LaneB != "" && c.LaneA == c.LaneB {
					lane = c.LaneA
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", c.TaskA, c.TaskB, c.Risk, files, lane))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ─── fmtTime ────────────────────────────────────────────────────────────────
func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func buildCodexTopTasks(details []BacklogTaskDetail, waves map[int][]string) []map[string]interface{} {
	if len(details) == 0 {
		return []map[string]interface{}{}
	}

	waveByTask := make(map[string]int, len(details))
	for level, ids := range waves {
		for _, id := range ids {
			waveByTask[id] = level
		}
	}

	limit := 3
	if len(details) < limit {
		limit = len(details)
	}

	topTasks := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		d := details[i]
		whyNow := fmt.Sprintf("wave_%d_priority_%s", waveByTask[d.ID], NormalizePriority(d.Priority))
		if waveByTask[d.ID] == 0 {
			whyNow = "wave_0_no_dependencies"
		}

		topTasks = append(topTasks, map[string]interface{}{
			"id":       d.ID,
			"content":  d.Content,
			"priority": d.Priority,
			"status":   d.Status,
			"level":    d.Level,
			"tags":     d.Tags,
			"why_now":  whyNow,
		})
	}

	return topTasks
}

func buildExecutionPlanSuggestedNextAction(topTasks []map[string]interface{}) string {
	if len(topTasks) == 0 {
		return "No actionable backlog tasks found."
	}

	first := topTasks[0]
	id := ParamString(first, "id")
	content := ParamString(first, "content")
	if id == "" {
		return "No actionable backlog tasks found."
	}
	if content == "" {
		return fmt.Sprintf("Work on %s.", id)
	}

	return fmt.Sprintf("Work on %s: %s", id, content)
}

func buildExecutionPlanAgentHint(suggestedNextAction string) string {
	if suggestedNextAction == "" || suggestedNextAction == "No actionable backlog tasks found." {
		return "Backlog is empty or filtered out; use task_workflow list/show to inspect remaining work."
	}

	return suggestedNextAction + " Then use task_workflow show or summarize on that task before implementation."
}

func buildExecutionPlanSummary(backlogCount int, topTasks []map[string]interface{}) string {
	if backlogCount == 0 {
		return "Backlog has no Todo or In Progress tasks."
	}

	if len(topTasks) == 0 {
		return fmt.Sprintf("Backlog has %d actionable tasks.", backlogCount)
	}

	return fmt.Sprintf("Backlog has %d actionable tasks. Start with %s.", backlogCount, ParamString(topTasks[0], "id"))
}

// TaskCollision represents a file collision between two tasks.
type TaskCollision struct {
	TaskA string   `json:"task_a"`
	TaskB string   `json:"task_b"`
	Files []string `json:"files"`  // Overlapping files
	LaneA string   `json:"lane_a"` // Lane of task A (if set)
	LaneB string   `json:"lane_b"` // Lane of task B (if set)
	Risk  string   `json:"risk"`   // "high" (same hotspot files) or "medium" (same lane)
}

// TaskOwnershipInfo holds ownership info for collision detection.
type TaskOwnershipInfo struct {
	TaskID         string
	OwnedFiles     []string
	OwnedGlobs     []string
	Lane           string
	ForbiddenFiles []string
}

// GetTaskOwnershipInfo extracts ownership info from a task.
func GetTaskOwnershipInfo(task *Todo2Task) *TaskOwnershipInfo {
	own := models.GetTaskOwnership(task)
	if own == nil {
		return nil
	}
	return &TaskOwnershipInfo{
		TaskID:         task.ID,
		OwnedFiles:     own.OwnedFiles,
		OwnedGlobs:     own.OwnedGlobs,
		Lane:           own.Lane,
		ForbiddenFiles: own.ForbiddenFiles,
	}
}

// DetectFileCollisions detects file collisions between tasks with ownership metadata.
// Returns collisions sorted by risk (high first).
func DetectFileCollisions(tasks []Todo2Task) []TaskCollision {
	// Build ownership map
	ownershipMap := make(map[string]*TaskOwnershipInfo)
	for i := range tasks {
		info := GetTaskOwnershipInfo(&tasks[i])
		if info != nil && (len(info.OwnedFiles) > 0 || len(info.OwnedGlobs) > 0 || info.Lane != "") {
			ownershipMap[info.TaskID] = info
		}
	}

	// Only check pending tasks (Todo + In Progress)
	pendingIDs := make(map[string]bool)
	for _, task := range tasks {
		if IsPendingStatus(task.Status) {
			pendingIDs[task.ID] = true
		}
	}

	var collisions []TaskCollision

	// Check each pair of pending tasks
	taskIDs := make([]string, 0, len(ownershipMap))
	for id := range ownershipMap {
		if pendingIDs[id] {
			taskIDs = append(taskIDs, id)
		}
	}
	sort.Strings(taskIDs)

	for i := 0; i < len(taskIDs); i++ {
		for j := i + 1; j < len(taskIDs); j++ {
			infoA := ownershipMap[taskIDs[i]]
			infoB := ownershipMap[taskIDs[j]]

			// Find overlapping files
			overlapping := findOverlappingFiles(infoA, infoB)

			// Check same lane
			sameLane := infoA.Lane != "" && infoB.Lane != "" && infoA.Lane == infoB.Lane

			if len(overlapping) > 0 || sameLane {
				risk := "medium"
				if len(overlapping) > 0 {
					risk = "high" // Direct file overlap is high risk
				}
				collisions = append(collisions, TaskCollision{
					TaskA: infoA.TaskID,
					TaskB: infoB.TaskID,
					Files: overlapping,
					LaneA: infoA.Lane,
					LaneB: infoB.Lane,
					Risk:  risk,
				})
			}
		}
	}

	// Sort: high risk first
	sort.Slice(collisions, func(i, j int) bool {
		if collisions[i].Risk != collisions[j].Risk {
			return collisions[i].Risk == "high"
		}
		return collisions[i].TaskA < collisions[j].TaskA
	})

	return collisions
}

// findOverlappingFiles finds files that appear in both tasks' owned_files.
func findOverlappingFiles(a, b *TaskOwnershipInfo) []string {
	fileSet := make(map[string]bool)
	for _, f := range a.OwnedFiles {
		fileSet[f] = true
	}

	var overlapping []string
	for _, f := range b.OwnedFiles {
		if fileSet[f] {
			overlapping = append(overlapping, f)
		}
	}
	sort.Strings(overlapping)
	return overlapping
}

// BuildCollisionMap returns a map of task ID -> list of task IDs it conflicts with.
func BuildCollisionMap(collisions []TaskCollision) map[string][]string {
	m := make(map[string][]string)
	for _, c := range collisions {
		m[c.TaskA] = append(m[c.TaskA], c.TaskB)
		m[c.TaskB] = append(m[c.TaskB], c.TaskA)
	}
	// Sort each list
	for k := range m {
		sort.Strings(m[k])
	}
	return m
}

// handleTaskAnalysisComplexity classifies task complexity (simple/medium/complex) using heuristic rules.
