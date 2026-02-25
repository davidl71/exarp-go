// task_analysis_deps_analysis.go — Task analysis: complexity, parallelization, fix-deps, validate, and noise handlers.
// See also: task_analysis_deps.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/taskanalysis"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleTaskAnalysisComplexity
//   handleTaskAnalysisParallelization — handleTaskAnalysisParallelization handles parallelization analysis.
//   handleTaskAnalysisFixMissingDeps — handleTaskAnalysisFixMissingDeps removes invalid dependency refs from tasks and saves.
//   handleTaskAnalysisValidate — handleTaskAnalysisValidate reports missing dependency IDs and optionally hierarchy parse warnings.
//   handleTaskAnalysisNoise — handleTaskAnalysisNoise identifies likely noise tasks (sentence fragments, junk) using heuristics.
//   findNoiseTasks — findNoiseTasks applies heuristics to detect sentence fragments and junk tasks.
// ────────────────────────────────────────────────────────────────────────────

// ─── handleTaskAnalysisComplexity ───────────────────────────────────────────
func handleTaskAnalysisComplexity(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Optional: filter to backlog only
	statusFilter := params["status_filter"]

	var tasks []Todo2Task

	if sf, ok := statusFilter.(string); ok && sf != "" {
		for _, t := range list {
			if t != nil && t.Status == sf {
				tasks = append(tasks, *t)
			}
		}
	} else {
		tasks = tasksFromPtrs(list)
	}

	classifications := make([]map[string]interface{}, 0, len(tasks))

	for _, t := range tasks {
		taskPtr := &t
		r := taskanalysis.AnalyzeTask(taskPtr)
		classifications = append(classifications, map[string]interface{}{
			"id":               t.ID,
			"content":          t.Content,
			"complexity":       string(r.Complexity),
			"can_auto_execute": r.CanAutoExecute,
			"needs_breakdown":  r.NeedsBreakdown,
			"reason":           r.Reason,
		})
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"classifications": classifications,
		"total":           len(classifications),
	}

	outputPath, _ := params["output_path"].(string)
	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "complexity", OutputPath: outputPath, ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// ─── handleTaskAnalysisParallelization ──────────────────────────────────────
// handleTaskAnalysisParallelization handles parallelization analysis.
func handleTaskAnalysisParallelization(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	durationWeight := 0.3
	if weight, ok := params["duration_weight"].(float64); ok {
		durationWeight = weight
	}

	// Find parallelizable tasks
	parallelGroups := findParallelizableTasks(tasks, durationWeight)

	outputFormat := "json"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"total_tasks":     len(tasks),
		"parallel_groups": parallelGroups,
		"duration_weight": durationWeight,
		"recommendations": buildParallelizationRecommendations(parallelGroups),
	}

	// Include human-readable report in JSON for CLI/consumers
	result["report"] = formatParallelizationAnalysisText(result)

	outputPath, _ := params["output_path"].(string)
	if outputFormat == "json" {
		if outputPath != "" {
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create output dir: %w", err)
			}
		}

		resultJSON, _ := json.Marshal(result)
		resp := &proto.TaskAnalysisResponse{Action: "parallelization", OutputPath: outputPath, ResultJson: string(resultJSON)}

		return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	output := formatParallelizationAnalysisText(result)

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

// ─── handleTaskAnalysisFixMissingDeps ───────────────────────────────────────
// handleTaskAnalysisFixMissingDeps removes invalid dependency refs from tasks and saves.
// Use once to fix tasks that depend on non-existent IDs (e.g. T-45, T-5 depending on T-4).
func handleTaskAnalysisFixMissingDeps(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	missing := findMissingDependencies(tasks, tg)
	if len(missing) == 0 {
		out := map[string]interface{}{
			"success":     true,
			"message":     "No missing dependency refs to fix",
			"total_tasks": len(tasks),
		}
		resultJSON, _ := json.Marshal(out)
		resp := &proto.TaskAnalysisResponse{Action: "fix_missing_deps", ResultJson: string(resultJSON)}

		return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
	}

	// Per task, collect missing dep IDs to remove
	missingDepsByTask := make(map[string]map[string]struct{})

	for _, m := range missing {
		tid, _ := m["task_id"].(string)
		dep, _ := m["missing_dep"].(string)

		if tid == "" || dep == "" {
			continue
		}

		if missingDepsByTask[tid] == nil {
			missingDepsByTask[tid] = make(map[string]struct{})
		}

		missingDepsByTask[tid][dep] = struct{}{}
	}

	// Remove all invalid deps from each task in one pass
	fixed := 0

	for i := range tasks {
		t := &tasks[i]

		toRemove := missingDepsByTask[t.ID]
		if len(toRemove) == 0 {
			continue
		}

		newDeps := make([]string, 0, len(t.Dependencies))

		for _, d := range t.Dependencies {
			if _, remove := toRemove[d]; remove {
				fixed++
			} else {
				newDeps = append(newDeps, d)
			}
		}

		t.Dependencies = newDeps
	}

	for _, t := range tasks {
		if missingDepsByTask[t.ID] != nil {
			taskPtr := &t
			if err := store.UpdateTask(ctx, taskPtr); err != nil {
				return nil, fmt.Errorf("failed to save task %s after fix: %w", t.ID, err)
			}
		}
	}

	result := map[string]interface{}{
		"success":      true,
		"total_tasks":  len(tasks),
		"missing_refs": len(missing),
		"removed":      fixed,
		"message":      fmt.Sprintf("Removed %d invalid dependency ref(s) and saved", fixed),
	}
	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "fix_missing_deps", ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// ─── handleTaskAnalysisValidate ─────────────────────────────────────────────
// handleTaskAnalysisValidate reports missing dependency IDs and optionally hierarchy parse warnings.
// Returns JSON with missing_deps and optional hierarchy_warning (when FM returns non-JSON).
func handleTaskAnalysisValidate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	missing := findMissingDependencies(tasks, tg)
	result := map[string]interface{}{
		"success":       true,
		"method":        "native_go",
		"total_tasks":   len(tasks),
		"missing_deps":  missing,
		"missing_count": len(missing),
	}

	includeHierarchy := false
	if h, ok := params["include_hierarchy"].(bool); ok {
		includeHierarchy = h
	}

	if includeHierarchy && FMAvailable() {
		hierResult, err := handleTaskAnalysisHierarchy(ctx, params)
		if err == nil && len(hierResult) > 0 {
			var hierData map[string]interface{}
			if json.Unmarshal([]byte(hierResult[0].Text), &hierData) == nil {
				if skipped, _ := hierData["hierarchy_skipped"].(string); skipped != "" {
					result["hierarchy_warning"] = map[string]interface{}{
						"skipped":          skipped,
						"parse_error":      hierData["parse_error"],
						"response_snippet": hierData["response_snippet"],
					}
				}
			}
		}
	}

	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "validate", ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// ─── handleTaskAnalysisNoise ────────────────────────────────────────────────
// handleTaskAnalysisNoise identifies likely noise tasks (sentence fragments, junk) using heuristics.
// By default filters to tasks tagged #discovered; pass filter_tag="" to scan all tasks.
func handleTaskAnalysisNoise(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	// Default filter_tag to "discovered" for noise action
	filterTag, _ := params["filter_tag"].(string)
	if filterTag == "" {
		filterTag = "discovered"
	}

	// Filter to tasks with the tag
	if filterTag != "" {
		filtered := make([]Todo2Task, 0, len(tasks))
		tagLower := strings.ToLower(filterTag)
		for _, t := range tasks {
			for _, tag := range t.Tags {
				if strings.ToLower(tag) == tagLower {
					filtered = append(filtered, t)
					break
				}
			}
		}
		tasks = filtered
	}

	noiseCandidates := findNoiseTasks(tasks)
	result := map[string]interface{}{
		"success":          true,
		"method":           "native_go",
		"filter_tag":       filterTag,
		"tasks_scanned":    len(tasks),
		"noise_count":      len(noiseCandidates),
		"noise_candidates": noiseCandidates,
	}

	outputPath, _ := params["output_path"].(string)
	resultJSON, _ := json.Marshal(result)
	resp := &proto.TaskAnalysisResponse{Action: "noise", OutputPath: outputPath, ResultJson: string(resultJSON)}

	return response.FormatResult(TaskAnalysisResponseToMap(resp), resp.GetOutputPath())
}

// ─── findNoiseTasks ─────────────────────────────────────────────────────────
// findNoiseTasks applies heuristics to detect sentence fragments and junk tasks.
func findNoiseTasks(tasks []Todo2Task) []map[string]interface{} {
	// Fragment starters: mid-sentence phrases, not actionable
	fragmentStarters := []string{
		"that ", "and ", "which ", "the ", "this ", "it ", "these ", "those ",
		"of ", "in ", "for ", "to ", "by ", "is ", "are ", "was ", "were ",
	}
	// Action verbs that suggest real tasks
	actionVerbs := []string{
		"add", "fix", "implement", "create", "update", "migrate", "document",
		"test", "refactor", "remove", "replace", "extract", "integrate",
		"verify", "validate", "support", "handle", "allow", "enable",
		"ensure", "improve", "optimize", "complete", "run", "build",
	}

	var candidates []map[string]interface{}
	for _, t := range tasks {
		content := strings.TrimSpace(t.Content + " " + t.LongDescription)
		contentLower := strings.ToLower(content)
		var reason string

		if len(content) < 20 {
			reason = "very short content"
		} else if strings.Contains(content, "...") && len(content) < 80 {
			reason = "truncated/ellipsis"
		} else {
			firstWord := strings.Fields(contentLower)
			if len(firstWord) > 0 {
				first := firstWord[0]
				for _, starter := range fragmentStarters {
					if strings.HasPrefix(contentLower, starter) {
						reason = "sentence fragment (starts with \"" + starter + "\")"
						break
					}
				}
				if reason == "" {
					hasAction := false
					for _, verb := range actionVerbs {
						if strings.HasPrefix(first, verb) || strings.Contains(contentLower, " "+verb+" ") {
							hasAction = true
							break
						}
					}
					if !hasAction && len(content) < 60 {
						reason = "no action verb, short content"
					}
				}
			}
		}

		if reason != "" {
			candidates = append(candidates, map[string]interface{}{
				"id":      t.ID,
				"content": truncateStr(t.Content, 80),
				"reason":  reason,
			})
		}
	}
	return candidates
}

