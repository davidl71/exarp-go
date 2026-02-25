// task_analysis_graph.go — task_analysis graph helpers.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

func buildLegacyGraphFormat(tg *TaskGraph) DependencyGraph {
	graph := make(DependencyGraph)

	nodes := tg.Graph.Nodes()
	for nodes.Next() {
		nodeID := nodes.Node().ID()
		taskID := tg.NodeIDMap[nodeID]

		deps := []string{}

		fromNodes := tg.Graph.To(nodeID)
		for fromNodes.Next() {
			fromNodeID := fromNodes.Node().ID()
			if depTaskID, ok := tg.NodeIDMap[fromNodeID]; ok {
				deps = append(deps, depTaskID)
			}
		}

		graph[taskID] = deps
	}

	return graph
}

func findMissingDependencies(tasks []Todo2Task, tg *TaskGraph) []map[string]interface{} {
	missing := []map[string]interface{}{}
	taskMap := make(map[string]bool)

	for _, task := range tasks {
		taskMap[task.ID] = true
	}

	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if !taskMap[dep] {
				missing = append(missing, map[string]interface{}{
					"task_id":     task.ID,
					"missing_dep": dep,
					"message":     fmt.Sprintf("Task %s depends on %s which doesn't exist", task.ID, dep),
				})
			}
		}
	}

	return missing
}

// GetDependencyAnalysisFromTasks returns cycles and missing dependencies for the given tasks.
// Used by task_discovery findOrphanTasks and handleTaskAnalysisDependencies to share graph logic.
func GetDependencyAnalysisFromTasks(tasks []Todo2Task) (cycles [][]string, missing []map[string]interface{}, err error) {
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	cycles = DetectCycles(tg)
	missing = findMissingDependencies(tasks, tg)

	return cycles, missing, nil
}

func buildDependencyRecommendations(graph DependencyGraph, cycles [][]string, missing []map[string]interface{}) []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	if len(cycles) > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":    "circular_dependency",
			"count":   len(cycles),
			"message": fmt.Sprintf("Found %d circular dependency chain(s)", len(cycles)),
		})
	}

	if len(missing) > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":    "missing_dependency",
			"count":   len(missing),
			"message": fmt.Sprintf("Found %d missing dependency reference(s)", len(missing)),
		})
	}

	return recommendations
}

func formatDependencyAnalysisText(result map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Dependency Analysis\n")
	sb.WriteString("==================\n\n")

	if total, ok := result["total_tasks"].(int); ok {
		sb.WriteString(fmt.Sprintf("Total Tasks: %d\n", total))
	}

	if maxLevel, ok := result["max_dependency_level"].(int); ok {
		sb.WriteString(fmt.Sprintf("Max Dependency Level: %d\n", maxLevel))
	}

	sb.WriteString("\n")

	// Critical Path
	if criticalPath, ok := result["critical_path"].([]string); ok && len(criticalPath) > 0 {
		sb.WriteString("Critical Path (Longest Dependency Chain):\n")
		sb.WriteString(fmt.Sprintf("  Length: %d tasks\n\n", len(criticalPath)))

		if details, ok := result["critical_path_details"].([]map[string]interface{}); ok {
			for i, detail := range details {
				taskID, _ := detail["id"].(string)
				content, _ := detail["content"].(string)

				sb.WriteString(fmt.Sprintf("  %d. %s", i+1, taskID))

				if content != "" {
					sb.WriteString(fmt.Sprintf(": %s", content))
				}

				sb.WriteString("\n")

				if deps, ok := detail["dependencies"].([]interface{}); ok && len(deps) > 0 {
					depStrs := make([]string, len(deps))

					for j, d := range deps {
						if depStr, ok := d.(string); ok {
							depStrs[j] = depStr
						}
					}

					if len(depStrs) > 0 {
						sb.WriteString(fmt.Sprintf("     Depends on: %s\n", strings.Join(depStrs, ", ")))
					}
				}

				if i < len(details)-1 {
					sb.WriteString("     ↓\n")
				}
			}
		} else {
			// Fallback to simple path
			sb.WriteString(fmt.Sprintf("  %s\n", strings.Join(criticalPath, " → ")))
		}

		sb.WriteString("\n")
	}

	if cycles, ok := result["circular_dependencies"].([][]string); ok && len(cycles) > 0 {
		sb.WriteString("Circular Dependencies:\n")

		for i, cycle := range cycles {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, strings.Join(cycle, " -> ")))
		}

		sb.WriteString("\n")
	}

	if missing, ok := result["missing_dependencies"].([]map[string]interface{}); ok && len(missing) > 0 {
		sb.WriteString("Missing Dependencies:\n")

		for _, m := range missing {
			if taskID, ok := m["task_id"].(string); ok {
				if dep, ok := m["missing_dep"].(string); ok {
					sb.WriteString(fmt.Sprintf("  - %s depends on %s (not found)\n", taskID, dep))
				}
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// Helper functions for parallelization analysis

type ParallelGroup struct {
	Tasks    []string `json:"tasks"`
	Priority string   `json:"priority"`
	Reason   string   `json:"reason"`
}

func findParallelizableTasks(tasks []Todo2Task, durationWeight float64) []ParallelGroup {
	groups := []ParallelGroup{}

	// Build dependency graph using gonum
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		// Fallback to simple approach if graph building fails
		return findParallelizableTasksSimple(tasks, durationWeight)
	}

	taskMap := make(map[string]*Todo2Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Filter to pending tasks only
	pendingTasks := []Todo2Task{}

	for _, task := range tasks {
		if IsPendingStatus(task.Status) {
			pendingTasks = append(pendingTasks, task)
		}
	}

	if len(pendingTasks) == 0 {
		return groups
	}

	// Use dependency levels to group parallelizable tasks
	levels := GetTaskLevels(tg)

	// Group tasks by dependency level (tasks at same level can run in parallel)
	byLevel := make(map[int][]string)

	for _, task := range pendingTasks {
		level := levels[task.ID]
		byLevel[level] = append(byLevel[level], task.ID)
	}

	// Build parallel groups from each level
	for level, taskIDs := range byLevel {
		if len(taskIDs) < 2 {
			continue // Skip levels with only one task
		}

		// Group by priority within each level
		byPriority := make(map[string][]string)

		for _, taskID := range taskIDs {
			task := taskMap[taskID]

			priority := task.Priority
			if priority == "" {
				priority = models.PriorityMedium
			}

			byPriority[priority] = append(byPriority[priority], taskID)
		}

		// Create groups for each priority within this level
		for priority, ids := range byPriority {
			if len(ids) > 1 {
				groups = append(groups, ParallelGroup{
					Tasks:    ids,
					Priority: priority,
					Reason:   fmt.Sprintf("%d tasks at dependency level %d can run in parallel", len(ids), level),
				})
			}
		}
	}

	// Also include tasks with no dependencies (level 0)
	if level0Tasks, ok := byLevel[0]; ok && len(level0Tasks) > 0 {
		byPriority := make(map[string][]string)

		for _, taskID := range level0Tasks {
			task := taskMap[taskID]

			priority := task.Priority
			if priority == "" {
				priority = models.PriorityMedium
			}

			byPriority[priority] = append(byPriority[priority], taskID)
		}

		for priority, ids := range byPriority {
			if len(ids) > 1 {
				// Check if we already added this group
				exists := false

				for _, g := range groups {
					if g.Priority == priority && len(g.Tasks) == len(ids) {
						exists = true
						break
					}
				}

				if !exists {
					groups = append(groups, ParallelGroup{
						Tasks:    ids,
						Priority: priority,
						Reason:   fmt.Sprintf("%d tasks with no dependencies can run in parallel", len(ids)),
					})
				}
			}
		}
	}

	// Sort groups by priority (high -> medium -> low)
	sort.Slice(groups, func(i, j int) bool {
		priorityOrder := map[string]int{models.PriorityHigh: 0, models.PriorityMedium: 1, models.PriorityLow: 2}
		return priorityOrder[groups[i].Priority] < priorityOrder[groups[j].Priority]
	})

	return groups
}

// findParallelizableTasksSimple is a fallback implementation without graph analysis.
func findParallelizableTasksSimple(tasks []Todo2Task, durationWeight float64) []ParallelGroup {
	groups := []ParallelGroup{}

	taskMap := make(map[string]*Todo2Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Find tasks with no dependencies (can run in parallel)
	readyTasks := []string{}

	for _, task := range tasks {
		if IsPendingStatus(task.Status) && len(task.Dependencies) == 0 {
			readyTasks = append(readyTasks, task.ID)
		}
	}

	if len(readyTasks) > 0 {
		// Group by priority
		byPriority := make(map[string][]string)

		for _, taskID := range readyTasks {
			task := taskMap[taskID]

			priority := task.Priority
			if priority == "" {
				priority = models.PriorityMedium
			}

			byPriority[priority] = append(byPriority[priority], taskID)
		}

		for priority, taskIDs := range byPriority {
			if len(taskIDs) > 1 {
				groups = append(groups, ParallelGroup{
					Tasks:    taskIDs,
					Priority: priority,
					Reason:   fmt.Sprintf("%d tasks with no dependencies can run in parallel", len(taskIDs)),
				})
			}
		}
	}

	// Sort groups by priority (high -> medium -> low)
	sort.Slice(groups, func(i, j int) bool {
		priorityOrder := map[string]int{models.PriorityHigh: 0, models.PriorityMedium: 1, models.PriorityLow: 2}
		return priorityOrder[groups[i].Priority] < priorityOrder[groups[j].Priority]
	})

	return groups
}

func buildParallelizationRecommendations(groups []ParallelGroup) []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	totalParallelizable := 0
	for _, group := range groups {
		totalParallelizable += len(group.Tasks)
	}

	if totalParallelizable > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":    "parallel_execution",
			"count":   totalParallelizable,
			"groups":  len(groups),
			"message": fmt.Sprintf("%d tasks can be executed in parallel across %d groups", totalParallelizable, len(groups)),
		})
	}

	return recommendations
}

func formatParallelizationAnalysisText(result map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Parallelization Analysis\n")
	sb.WriteString("========================\n\n")

	if total, ok := result["total_tasks"].(int); ok {
		sb.WriteString(fmt.Sprintf("Total Tasks: %d\n\n", total))
	}

	if groups, ok := result["parallel_groups"].([]ParallelGroup); ok && len(groups) > 0 {
		sb.WriteString("Parallel Execution Groups:\n\n")

		for i, group := range groups {
			sb.WriteString(fmt.Sprintf("Group %d (%s priority):\n", i+1, group.Priority))
			sb.WriteString(fmt.Sprintf("  Reason: %s\n", group.Reason))
			sb.WriteString("  Tasks:\n")

			for _, taskID := range group.Tasks {
				sb.WriteString(fmt.Sprintf("    - %s\n", taskID))
			}

			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("No parallel execution opportunities found.\n")
	}

	return sb.String()
}

// Helper function to save analysis results.
func saveAnalysisResult(outputPath string, result map[string]interface{}) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, output, 0644)
}

// handleTaskAnalysisHierarchy handles hierarchy analysis using the FM provider abstraction.
// When DefaultFMProvider() is available (e.g. Apple FM on darwin/arm64/cgo), it classifies tasks; otherwise returns ErrFMNotSupported.
func handleTaskAnalysisHierarchy(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	if !FMAvailable() {
		return nil, fmt.Errorf("hierarchy requires a foundation model: %w", ErrFMNotSupported)
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

	pendingTasks := []Todo2Task{}

	for _, task := range tasks {
		if IsPendingStatus(task.Status) {
			pendingTasks = append(pendingTasks, task)
		}
	}

	if len(pendingTasks) == 0 {
		return []framework.TextContent{
			{Type: "text", Text: `{"success": true, "message": "No pending tasks to analyze"}`},
		}, nil
	}

	taskDescriptions := make([]string, 0, len(pendingTasks))

	for i, task := range pendingTasks {
		if i >= 20 {
			break
		}

		desc := task.Content
		if task.LongDescription != "" {
			desc += " " + task.LongDescription
		}

		taskDescriptions = append(taskDescriptions, fmt.Sprintf("Task %s: %s", task.ID, desc))
	}

	prompt := fmt.Sprintf(`Analyze these tasks and classify them into hierarchy levels:

Tasks:
%s

Classify each task into one of these categories:
- "component" - Tasks that belong to a specific component/feature
- "epic" - High-level tasks that contain multiple subtasks
- "task" - Regular standalone tasks
- "subtask" - Tasks that are part of a larger task

Return JSON array with format: [{"task_id": "T-1", "level": "component", "component": "security", "reason": "..."}, ...]`,
		strings.Join(taskDescriptions, "\n"))

	result, err := DefaultFMProvider().Generate(ctx, prompt, 2000, 0.2)
	if err != nil {
		return nil, fmt.Errorf("foundation model classification: %w", err)
	}

	var classifications []map[string]interface{}

	candidate := result
	if err := json.Unmarshal([]byte(candidate), &classifications); err != nil {
		candidate = ExtractJSONArrayFromLLMResponse(result)
		if err = json.Unmarshal([]byte(candidate), &classifications); err != nil {
			// Graceful fallback: return success with empty classifications and a warning so
			// task_analysis doesn't fail hard when the FM returns plain text instead of JSON.
			snippet := result
			if len(snippet) > MaxLLMResponseSnippetLen {
				snippet = snippet[:MaxLLMResponseSnippetLen] + "..."
			}

			analysis := map[string]interface{}{
				"success":           true,
				"method":            "foundation_model",
				"total_tasks":       len(tasks),
				"pending_tasks":     len(pendingTasks),
				"classifications":   []map[string]interface{}{},
				"hierarchy_skipped": "fm_response_not_valid_json",
				"parse_error":       err.Error(),
				"response_snippet":  snippet,
			}
			resultJSON, _ := json.Marshal(analysis)
			resp := &proto.TaskAnalysisResponse{Action: "hierarchy", ResultJson: string(resultJSON)}

			return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
		}
	}

	analysis := map[string]interface{}{
		"success":                   true,
		"method":                    "foundation_model",
		"total_tasks":               len(tasks),
		"pending_tasks":             len(pendingTasks),
		"classifications":           classifications,
		"hierarchy_recommendations": buildHierarchyRecommendations(classifications, pendingTasks),
	}

	includeRecommendations := true
	if rec, ok := params["include_recommendations"].(bool); ok {
		includeRecommendations = rec
	}

	if !includeRecommendations {
		delete(analysis, "hierarchy_recommendations")
	}

	outputFormat := "json"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	if outputFormat == "text" {
		output := formatHierarchyAnalysisText(analysis)
		return []framework.TextContent{{Type: "text", Text: output}}, nil
	}

	resultJSON, _ := json.Marshal(analysis)
	resp := &proto.TaskAnalysisResponse{Action: "hierarchy", ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

func buildHierarchyRecommendations(classifications []map[string]interface{}, tasks []Todo2Task) []map[string]interface{} {
	recommendations := []map[string]interface{}{}
	componentGroups := make(map[string][]string)

	for _, cls := range classifications {
		if comp, ok := cls["component"].(string); ok && comp != "" {
			taskID, _ := cls["task_id"].(string)
			componentGroups[comp] = append(componentGroups[comp], taskID)
		}
	}

	for comp, taskIDs := range componentGroups {
		if len(taskIDs) >= 5 {
			recommendations = append(recommendations, map[string]interface{}{
				"component":        comp,
				"task_count":       len(taskIDs),
				"recommendation":   "use_hierarchy",
				"suggested_prefix": fmt.Sprintf("T-%s", strings.ToUpper(comp)),
				"task_ids":         taskIDs,
			})
		}
	}

	return recommendations
}

func formatHierarchyAnalysisText(analysis map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Task Hierarchy Analysis\n")
	sb.WriteString("=======================\n\n")

	if total, ok := analysis["total_tasks"].(int); ok {
		sb.WriteString(fmt.Sprintf("Total Tasks: %d\n", total))
	}

	if pending, ok := analysis["pending_tasks"].(int); ok {
		sb.WriteString(fmt.Sprintf("Pending Tasks: %d\n\n", pending))
	}

	if recs, ok := analysis["hierarchy_recommendations"].([]map[string]interface{}); ok {
		sb.WriteString("Recommendations:\n")

		for _, rec := range recs {
			if comp, ok := rec["component"].(string); ok {
				sb.WriteString(fmt.Sprintf("- Component: %s\n", comp))

				if count, ok := rec["task_count"].(int); ok {
					sb.WriteString(fmt.Sprintf("  Tasks: %d\n", count))
				}

				if prefix, ok := rec["suggested_prefix"].(string); ok {
					sb.WriteString(fmt.Sprintf("  Suggested Prefix: %s\n", prefix))
				}

				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}
