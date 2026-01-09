//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	fm "github.com/blacktop/go-foundationmodels"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/platform"
)

// handleTaskAnalysisNative handles task_analysis with native Go and Apple FM
func handleTaskAnalysisNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "duplicates"
	}

	switch action {
	case "hierarchy":
		return handleTaskAnalysisHierarchy(ctx, params)
	case "duplicates":
		return handleTaskAnalysisDuplicates(ctx, params)
	case "tags":
		return handleTaskAnalysisTags(ctx, params)
	case "dependencies":
		return handleTaskAnalysisDependencies(ctx, params)
	case "parallelization":
		return handleTaskAnalysisParallelization(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// handleTaskAnalysisHierarchy handles hierarchy analysis with Apple FM
func handleTaskAnalysisHierarchy(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Check platform support
	support := platform.CheckAppleFoundationModelsSupport()
	if !support.Supported {
		return nil, fmt.Errorf("Apple Foundation Models not supported: %s", support.Reason)
	}

	// Load tasks
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Filter pending tasks
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

	// Use Apple FM to classify tasks into hierarchy levels
	sess := fm.NewSession()
	defer sess.Release()

	// Build prompt for classification
	taskDescriptions := make([]string, 0, len(pendingTasks))
	for i, task := range pendingTasks {
		if i >= 20 { // Limit to first 20 for prompt size
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

	// Use classify action with custom categories
	result := sess.RespondWithOptions(prompt, 2000, 0.2)

	// Parse result
	var classifications []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &classifications); err != nil {
		// If JSON parsing fails, try to extract JSON from text
		jsonStart := strings.Index(result, "[")
		jsonEnd := strings.LastIndex(result, "]")
		if jsonStart >= 0 && jsonEnd > jsonStart {
			if err := json.Unmarshal([]byte(result[jsonStart:jsonEnd+1]), &classifications); err != nil {
				return nil, fmt.Errorf("failed to parse Apple FM classification: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse Apple FM classification: %w", err)
		}
	}

	// Build analysis result
	analysis := map[string]interface{}{
		"success":                   true,
		"method":                    "apple_foundation_models",
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

	var output string
	if outputFormat == "text" {
		output = formatHierarchyAnalysisText(analysis)
	} else {
		outputBytes, _ := json.MarshalIndent(analysis, "", "  ")
		output = string(outputBytes)
	}

	return []framework.TextContent{
		{Type: "text", Text: output},
	}, nil
}

// buildHierarchyRecommendations builds recommendations from classifications
func buildHierarchyRecommendations(classifications []map[string]interface{}, tasks []Todo2Task) []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	// Group by component
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

// formatHierarchyAnalysisText formats analysis as text
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
