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
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
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

	outputPath, _ := params["output_path"].(string)
	if outputFormat == "json" {
		if outputPath != "" {
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create output dir: %w", err)
			}
		}

		resultJSON, _ := json.Marshal(result)
		resp := &proto.TaskAnalysisResponse{Action: "dependencies", OutputPath: outputPath, ResultJson: string(resultJSON)}

		return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
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

// ─── handleTaskAnalysisExecutionPlan ────────────────────────────────────────
// handleTaskAnalysisExecutionPlan handles execution plan: backlog (Todo + In Progress) in dependency order.
func handleTaskAnalysisExecutionPlan(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
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

	result := map[string]interface{}{
		"success":          true,
		"method":           "native_go",
		"backlog_count":    len(orderedIDs),
		"ordered_task_ids": orderedIDs,
		"waves":            waves,
		"details":          details,
	}

	outputFormat := "json"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	outputPath, _ := params["output_path"].(string)

	// subagents_plan: write parallel-execution-subagents.plan.md using wave detection
	if outputFormat == "subagents_plan" {
		wavesCopy := waves
		if max := config.MaxTasksPerWave(); max > 0 {
			wavesCopy = LimitWavesByMaxTasks(wavesCopy, max)
		}

		if len(wavesCopy) == 0 {
			return nil, fmt.Errorf("no waves (empty backlog or no Todo/In Progress tasks)")
		}

		planTitle, _ := params["plan_title"].(string)
		if planTitle == "" {
			planTitle = filepath.Base(projectRoot)

			if info, err := getProjectInfo(projectRoot); err == nil {
				if name, ok := info["name"].(string); ok && name != "" {
					planTitle = name
				}
			}
		}

		if outputPath == "" {
			outputPath = filepath.Join(projectRoot, ".cursor", "plans", "parallel-execution-subagents.plan.md")
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

		return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
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

	return sb.String()
}

// ─── fmtTime ────────────────────────────────────────────────────────────────
func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// handleTaskAnalysisComplexity classifies task complexity (simple/medium/complex) using heuristic rules.
