// task_analysis_shared.go — MCP "task_analysis" tool dispatcher and core handlers.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	action := ParamString(params, "action")
	if action == "" {
		action = "duplicates"
	}

	switch action {
	case "next_batch":
		return handleTaskAnalysisNextBatch(ctx, params)
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
	case "suggest_dependencies", "suggest_deps":
		return handleTaskAnalysisSuggestDependencies(ctx, params)
	case "noise":
		return handleTaskAnalysisNoise(ctx, params)
	case "infer_ownership":
		return handleTaskAnalysisInferOwnership(ctx, params)
	case "hotspots":
		return handleTaskAnalysisHotspots(ctx, params)
	case "stale":
		return handleTaskAnalysisStale(ctx, params)
	case "completable":
		return handleTaskAnalysisCompletable(ctx, params)
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

	taskOverlaps := DetectTaskOverlapConflicts(list)
	fileConflicts := DetectFileConflicts(list)
	hasConflict := len(taskOverlaps) > 0 || len(fileConflicts) > 0
	overlapping := make([]string, 0)

	if hasConflict {
		seen := make(map[string]bool)
		for _, c := range taskOverlaps {
			if !seen[c.DepTaskID] {
				seen[c.DepTaskID] = true

				overlapping = append(overlapping, c.DepTaskID)
			}

			if !seen[c.TaskID] {
				seen[c.TaskID] = true

				overlapping = append(overlapping, c.TaskID)
			}
		}
		for _, c := range fileConflicts {
			for _, id := range c.TaskIDs {
				if !seen[id] {
					seen[id] = true
					overlapping = append(overlapping, id)
				}
			}
		}
	}

	out := map[string]interface{}{
		"conflict":       hasConflict,
		"conflicts":      taskOverlaps,
		"file_conflicts": fileConflicts,
		"overlapping":    overlapping,
	}

	if hasConflict {
		reasons := make([]string, 0, len(taskOverlaps)+len(fileConflicts))
		for _, c := range taskOverlaps {
			reasons = append(reasons, c.Reason)
		}
		for _, c := range fileConflicts {
			reasons = append(reasons, "File conflict: tasks "+strings.Join(c.TaskIDs, ", ")+" share "+strings.Join(c.Files, ", "))
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
	if v, ok := ParamFloat64OK(params, "similarity_threshold"); ok {
		similarityThreshold = v
	}

	autoFix := ParamBool(params, "auto_fix", false)

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

	projectRoot, err := GetProjectRootWithFallback()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project root: %w", err)
	}
	outputPath := DefaultReportOutputPath(projectRoot, "TASK_ANALYSIS_DUPLICATES.md", params)
	if outputPath != "" {
		if err := EnsureParentDir(outputPath); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}
	}

	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "duplicates", OutputPath: outputPath, ResultJson: string(resultJSON)}

	return framework.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// handleTaskAnalysisStale surfaces stale-tag / metadata-flagged backlog tasks using the same
// heuristics as task_workflow cleanup (always dry_run; no deletions).
func handleTaskAnalysisStale(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	cleanupParams := map[string]interface{}{
		"dry_run":                 true,
		"include_legacy":         ParamBool(params, "include_legacy", false),
		"stale_threshold_hours":  ParamFloat64(params, "stale_threshold_hours", 2.0),
	}

	contents, err := handleTaskWorkflowCleanup(ctx, cleanupParams)
	if err != nil {
		return nil, fmt.Errorf("stale analysis: %w", err)
	}

	if len(contents) == 0 {
		return nil, fmt.Errorf("stale analysis: empty result")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(contents[0].Text), &payload); err != nil {
		return nil, fmt.Errorf("stale analysis: parse cleanup result: %w", err)
	}

	payload["action"] = "stale"

	return framework.FormatResult(payload, ParamString(params, "output_path"))
}

// handleTaskAnalysisCompletable runs infer_task_progress heuristics and labels the payload for task_analysis clients.
func handleTaskAnalysisCompletable(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	inferParams := make(map[string]interface{}, len(params))
	for k, v := range params {
		if k == "action" {
			continue
		}

		inferParams[k] = v
	}

	if _, ok := inferParams["dry_run"]; !ok {
		inferParams["dry_run"] = true
	}

	contents, err := handleInferTaskProgressNative(ctx, inferParams)
	if err != nil {
		return nil, fmt.Errorf("completable analysis: %w", err)
	}

	if len(contents) == 0 {
		return nil, fmt.Errorf("completable analysis: empty result")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(contents[0].Text), &payload); err != nil {
		return nil, fmt.Errorf("completable analysis: parse infer result: %w", err)
	}

	payload["action"] = "completable"

	return framework.FormatResult(payload, ParamString(params, "output_path"))
}

// CanonicalTagRules returns default tag consolidation rules aligned with scorecard dimensions.
// Categories: testing, docs, security, build, performance, bug, feature, refactor, migration, config, cli, mcp, llm, database, workflow, planning, linting.
