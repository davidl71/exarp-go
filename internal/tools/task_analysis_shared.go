// task_analysis_shared.go — MCP "task_analysis" tool dispatcher and core handlers.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
)

// TaskAnalysisResponseToMap converts TaskAnalysisResponse to a map for response.FormatResult (unmarshals result_json into map).
func TaskAnalysisResponseToMap(resp *proto.TaskAnalysisResponse) map[string]interface{} {
	if resp == nil {
		return nil
	}

	out := map[string]interface{}{
		"action": resp.GetAction(),
	}
	if resp.GetOutputPath() != "" {
		out["output_path"] = resp.GetOutputPath()
	}

	if resp.GetResultJson() != "" {
		var payload map[string]interface{}
		if json.Unmarshal([]byte(resp.GetResultJson()), &payload) == nil {
			for k, v := range payload {
				out[k] = v
			}
		}
	}

	return out
}

// handleTaskAnalysisNative dispatches to the appropriate action (duplicates, tags, dependencies, parallelization, hierarchy).
// Hierarchy uses the FM abstraction (DefaultFMProvider()); when FM is not available, hierarchy returns a clear error (no Python fallback).
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
	case "discover_tags":
		return handleTaskAnalysisDiscoverTags(ctx, params)
	case "dependencies":
		return handleTaskAnalysisDependencies(ctx, params)
	case "parallelization":
		return handleTaskAnalysisParallelization(ctx, params)
	case "fix_missing_deps":
		return handleTaskAnalysisFixMissingDeps(ctx, params)
	case "validate":
		return handleTaskAnalysisValidate(ctx, params)
	case "execution_plan":
		return handleTaskAnalysisExecutionPlan(ctx, params)
	case "complexity":
		return handleTaskAnalysisComplexity(ctx, params)
	case "conflicts":
		return handleTaskAnalysisConflicts(ctx, params)
	case "dependencies_summary":
		return handleTaskAnalysisDependenciesSummary(ctx, params)
	case "suggest_dependencies":
		return handleTaskAnalysisSuggestDependencies(ctx, params)
	case "noise":
		return handleTaskAnalysisNoise(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// handleTaskAnalysisConflicts detects task-overlap conflicts (In Progress tasks with dependent also In Progress).
func handleTaskAnalysisConflicts(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	conflicts := DetectTaskOverlapConflicts(list)
	hasConflict := len(conflicts) > 0
	overlapping := make([]string, 0)

	if hasConflict {
		seen := make(map[string]bool)
		for _, c := range conflicts {
			if !seen[c.DepTaskID] {
				seen[c.DepTaskID] = true

				overlapping = append(overlapping, c.DepTaskID)
			}

			if !seen[c.TaskID] {
				seen[c.TaskID] = true

				overlapping = append(overlapping, c.TaskID)
			}
		}
	}

	out := map[string]interface{}{
		"conflict":    hasConflict,
		"conflicts":   conflicts,
		"overlapping": overlapping,
	}

	if hasConflict {
		reasons := make([]string, len(conflicts))
		for i, c := range conflicts {
			reasons[i] = c.Reason
		}

		out["reasons"] = reasons
	}

	resultJSON, _ := json.Marshal(out)
	resp := &proto.TaskAnalysisResponse{Action: "conflicts", ResultJson: string(resultJSON)}

	return framework.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// handleTaskAnalysisDuplicates handles duplicates detection.
func handleTaskAnalysisDuplicates(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	// Use config default, allow override from params
	similarityThreshold := config.SimilarityThreshold()
	if threshold, ok := params["similarity_threshold"].(float64); ok {
		similarityThreshold = threshold
	}

	autoFix := false
	if fix, ok := params["auto_fix"].(bool); ok {
		autoFix = fix
	}

	// Find duplicates
	duplicates := findDuplicateTasks(tasks, similarityThreshold)

	// Auto-fix if requested
	if autoFix && len(duplicates) > 0 {
		tasks = mergeDuplicateTasks(tasks, duplicates)
		// Delete removed task IDs (merge keeps first per group, removes group[1:])
		for _, grp := range duplicates {
			for i := 1; i < len(grp); i++ {
				_ = store.DeleteTask(ctx, grp[i])
			}
		}
		// Update kept/merged tasks
		for _, t := range tasks {
			taskPtr := &t
			if err := store.UpdateTask(ctx, taskPtr); err != nil {
				return nil, fmt.Errorf("failed to save merged task %s: %w", t.ID, err)
			}
		}
	}

	// Build result
	result := map[string]interface{}{
		"success":              true,
		"method":               "native_go",
		"total_tasks":          len(tasks),
		"duplicate_groups":     len(duplicates),
		"duplicates":           duplicates,
		"similarity_threshold": similarityThreshold,
		"auto_fix":             autoFix,
	}

	if autoFix {
		result["merged"] = true
		result["tasks_after_merge"] = len(tasks)
	}

	projectRoot, _ := FindProjectRoot()
	outputPath := DefaultReportOutputPath(projectRoot, "TASK_ANALYSIS_DUPLICATES.md", params)
	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}
	}

	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "duplicates", OutputPath: outputPath, ResultJson: string(resultJSON)}

	return framework.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// CanonicalTagRules returns default tag consolidation rules aligned with scorecard dimensions.
// Categories: testing, docs, security, build, performance, bug, feature, refactor, migration, config, cli, mcp, llm, database, workflow, planning, linting.
