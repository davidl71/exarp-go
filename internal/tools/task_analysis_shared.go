package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
)

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
	case "dependencies":
		return handleTaskAnalysisDependencies(ctx, params)
	case "parallelization":
		return handleTaskAnalysisParallelization(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// handleTaskAnalysisDuplicates handles duplicates detection
func handleTaskAnalysisDuplicates(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

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
		if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
			return nil, fmt.Errorf("failed to save merged tasks: %w", err)
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

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		if err := saveAnalysisResult(outputPath, result); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}
		result["output_path"] = outputPath
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleTaskAnalysisTags handles tag analysis and consolidation
func handleTaskAnalysisTags(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	dryRun := true
	if run, ok := params["dry_run"].(bool); ok {
		dryRun = run
	}

	// Analyze tags
	tagAnalysis := analyzeTags(tasks)

	// Apply custom rules if provided
	customRulesJSON, _ := params["custom_rules"].(string)
	if customRulesJSON != "" {
		var rules map[string]string
		if err := json.Unmarshal([]byte(customRulesJSON), &rules); err == nil {
			tagAnalysis = applyTagRules(tasks, rules, tagAnalysis)
		}
	}

	// Remove tags if requested
	removeTagsJSON, _ := params["remove_tags"].(string)
	if removeTagsJSON != "" {
		var tagsToRemove []string
		if err := json.Unmarshal([]byte(removeTagsJSON), &tagsToRemove); err == nil {
			tagAnalysis = removeTags(tasks, tagsToRemove, tagAnalysis)
		}
	}

	// Apply changes if not dry run
	if !dryRun {
		tasks = applyTagChanges(tasks, tagAnalysis)
		if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
			return nil, fmt.Errorf("failed to save tasks: %w", err)
		}
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"dry_run":         dryRun,
		"tag_analysis":    tagAnalysis,
		"total_tasks":     len(tasks),
		"recommendations": buildTagRecommendations(tagAnalysis),
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		if err := saveAnalysisResult(outputPath, result); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}
		result["output_path"] = outputPath
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleTaskAnalysisDependencies handles dependency analysis
func handleTaskAnalysisDependencies(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Build dependency graph using gonum
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	// Detect cycles
	cycles := DetectCycles(tg)

	// Find missing dependencies
	missing := findMissingDependencies(tasks, tg)

	// Build legacy graph format for backward compatibility
	graph := buildLegacyGraphFormat(tg)

	// Calculate critical path if no cycles
	var criticalPath []string
	var criticalPathDetails []map[string]interface{}
	maxLevel := 0

	hasCycles, err := HasCycles(tg)
	if err == nil && !hasCycles {
		// Find critical path
		path, err := FindCriticalPath(tg)
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

		// Get max dependency level
		levels := GetTaskLevels(tg)
		for _, level := range levels {
			if level > maxLevel {
				maxLevel = level
			}
		}
	}

	outputFormat := "text"
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

	var output string
	if outputFormat == "json" {
		outputBytes, _ := json.MarshalIndent(result, "", "  ")
		output = string(outputBytes)
	} else {
		output = formatDependencyAnalysisText(result)
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}
		result["output_path"] = outputPath
	}

	return []framework.TextContent{
		{Type: "text", Text: output},
	}, nil
}

// handleTaskAnalysisParallelization handles parallelization analysis
func handleTaskAnalysisParallelization(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	durationWeight := 0.3
	if weight, ok := params["duration_weight"].(float64); ok {
		durationWeight = weight
	}

	// Find parallelizable tasks
	parallelGroups := findParallelizableTasks(tasks, durationWeight)

	outputFormat := "text"
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

	var output string
	if outputFormat == "json" {
		outputBytes, _ := json.MarshalIndent(result, "", "  ")
		output = string(outputBytes)
	} else {
		output = formatParallelizationAnalysisText(result)
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("failed to save result: %w", err)
		}
		result["output_path"] = outputPath
	}

	return []framework.TextContent{
		{Type: "text", Text: output},
	}, nil
}

// Helper functions for duplicates detection

func findDuplicateTasks(tasks []Todo2Task, threshold float64) [][]string {
	// For small datasets, use sequential approach (overhead of parallelization not worth it)
	if len(tasks) < 100 {
		return findDuplicateTasksSequential(tasks, threshold)
	}

	// For larger datasets, use parallel processing
	return findDuplicateTasksParallel(tasks, threshold)
}

// findDuplicateTasksSequential is the original sequential implementation
// Optimized for small datasets where parallel overhead isn't worth it
func findDuplicateTasksSequential(tasks []Todo2Task, threshold float64) [][]string {
	duplicates := [][]string{}
	processed := make(map[string]bool)

	for i, task1 := range tasks {
		if processed[task1.ID] {
			continue
		}

		group := []string{task1.ID}
		for j := i + 1; j < len(tasks); j++ {
			task2 := tasks[j]
			if processed[task2.ID] {
				continue
			}

			similarity := calculateSimilarity(task1, task2)
			if similarity >= threshold {
				group = append(group, task2.ID)
				processed[task2.ID] = true
			}
		}

		if len(group) > 1 {
			duplicates = append(duplicates, group)
			processed[task1.ID] = true
		}
	}

	return duplicates
}

// findDuplicateTasksParallel uses worker pool for parallel duplicate detection
// Optimized for large datasets (>100 tasks)
func findDuplicateTasksParallel(tasks []Todo2Task, threshold float64) [][]string {
	const numWorkers = 4 // Adjust based on CPU cores
	if len(tasks) == 0 {
		return [][]string{}
	}

	type taskPair struct {
		i     int
		task1 Todo2Task
	}

	type similarityResult struct {
		i          int
		j          int
		similarity float64
	}

	// Channel for task pairs to process
	taskPairs := make(chan taskPair, len(tasks))
	results := make(chan similarityResult, len(tasks)*len(tasks))

	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pair := range taskPairs {
				task1 := pair.task1
				for j := pair.i + 1; j < len(tasks); j++ {
					task2 := tasks[j]
					similarity := calculateSimilarity(task1, task2)
					if similarity >= threshold {
						results <- similarityResult{
							i:          pair.i,
							j:          j,
							similarity: similarity,
						}
					}
				}
			}
		}()
	}

	// Send all task pairs to workers
	go func() {
		defer close(taskPairs)
		for i, task1 := range tasks {
			taskPairs <- taskPair{i: i, task1: task1}
		}
	}()

	// Close results channel when all workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	processed := make(map[string]bool)
	duplicateMap := make(map[int][]int) // task index -> list of duplicate indices

	for result := range results {
		if !processed[tasks[result.i].ID] && !processed[tasks[result.j].ID] {
			if duplicateMap[result.i] == nil {
				duplicateMap[result.i] = []int{result.i}
			}
			duplicateMap[result.i] = append(duplicateMap[result.i], result.j)
			processed[tasks[result.j].ID] = true
		}
	}

	// Build duplicate groups
	duplicates := [][]string{}
	for i, group := range duplicateMap {
		if len(group) > 1 && !processed[tasks[i].ID] {
			groupIDs := make([]string, len(group))
			for idx, taskIdx := range group {
				groupIDs[idx] = tasks[taskIdx].ID
			}
			duplicates = append(duplicates, groupIDs)
			processed[tasks[i].ID] = true
		}
	}

	return duplicates
}

func calculateSimilarity(task1, task2 Todo2Task) float64 {
	// Simple word-based similarity (can be enhanced with Levenshtein)
	text1 := strings.ToLower(task1.Content + " " + task1.LongDescription)
	text2 := strings.ToLower(task2.Content + " " + task2.LongDescription)

	words1 := strings.Fields(text1)
	words2 := strings.Fields(text2)

	if len(words1) == 0 && len(words2) == 0 {
		return 1.0
	}
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Count common words
	wordSet1 := make(map[string]bool)
	for _, word := range words1 {
		wordSet1[word] = true
	}

	common := 0
	for _, word := range words2 {
		if wordSet1[word] {
			common++
		}
	}

	// Jaccard similarity
	union := len(wordSet1) + len(words2) - common
	if union == 0 {
		return 0.0
	}

	return float64(common) / float64(union)
}

func mergeDuplicateTasks(tasks []Todo2Task, duplicates [][]string) []Todo2Task {
	// Keep first task in each group, merge others into it
	keepMap := make(map[string]bool)
	removeMap := make(map[string]bool)

	for _, group := range duplicates {
		if len(group) == 0 {
			continue
		}
		keepMap[group[0]] = true
		for i := 1; i < len(group); i++ {
			removeMap[group[i]] = true
		}
	}

	// Merge task data
	taskMap := make(map[string]*Todo2Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	for _, group := range duplicates {
		if len(group) < 2 {
			continue
		}
		primary := taskMap[group[0]]
		for i := 1; i < len(group); i++ {
			secondary := taskMap[group[i]]
			// Merge tags
			tagSet := make(map[string]bool)
			for _, tag := range primary.Tags {
				tagSet[tag] = true
			}
			for _, tag := range secondary.Tags {
				if !tagSet[tag] {
					primary.Tags = append(primary.Tags, tag)
				}
			}
			// Merge dependencies
			depSet := make(map[string]bool)
			for _, dep := range primary.Dependencies {
				depSet[dep] = true
			}
			for _, dep := range secondary.Dependencies {
				if !depSet[dep] {
					primary.Dependencies = append(primary.Dependencies, dep)
				}
			}
		}
	}

	// Remove duplicate tasks
	result := []Todo2Task{}
	for _, task := range tasks {
		if !removeMap[task.ID] {
			result = append(result, task)
		}
	}

	return result
}

// Helper functions for tag analysis

type TagAnalysis struct {
	TagFrequency     map[string]int      `json:"tag_frequency"`
	TagSuggestions   map[string][]string `json:"tag_suggestions"`
	InconsistentTags map[string][]string `json:"inconsistent_tags"`
	UnusedTags       []string            `json:"unused_tags"`
}

func analyzeTags(tasks []Todo2Task) TagAnalysis {
	analysis := TagAnalysis{
		TagFrequency:     make(map[string]int),
		TagSuggestions:   make(map[string][]string),
		InconsistentTags: make(map[string][]string),
	}

	tagSet := make(map[string]bool)
	for _, task := range tasks {
		for _, tag := range task.Tags {
			analysis.TagFrequency[tag]++
			tagSet[tag] = true
		}
	}

	// Find tasks that might need tags
	for _, task := range tasks {
		if len(task.Tags) == 0 {
			// Suggest tags based on content
			suggestions := suggestTagsForTask(task)
			if len(suggestions) > 0 {
				analysis.TagSuggestions[task.ID] = suggestions
			}
		}
	}

	return analysis
}

func suggestTagsForTask(task Todo2Task) []string {
	suggestions := []string{}
	content := strings.ToLower(task.Content + " " + task.LongDescription)

	// Simple keyword-based suggestions
	keywords := map[string]string{
		"bug": "bug", "fix": "bug", "error": "bug",
		"feature": "feature", "add": "feature", "implement": "feature",
		"refactor": "refactor", "cleanup": "refactor", "improve": "refactor",
		"test": "testing", "testing": "testing",
		"doc": "documentation", "documentation": "documentation",
		"migration": "migration", "migrate": "migration",
	}

	for keyword, tag := range keywords {
		if strings.Contains(content, keyword) {
			suggestions = append(suggestions, tag)
		}
	}

	return suggestions
}

func applyTagRules(tasks []Todo2Task, rules map[string]string, analysis TagAnalysis) TagAnalysis {
	// Apply rename rules
	for oldTag, newTag := range rules {
		if count, ok := analysis.TagFrequency[oldTag]; ok {
			delete(analysis.TagFrequency, oldTag)
			analysis.TagFrequency[newTag] += count
		}
	}
	return analysis
}

func removeTags(tasks []Todo2Task, tagsToRemove []string, analysis TagAnalysis) TagAnalysis {
	removeSet := make(map[string]bool)
	for _, tag := range tagsToRemove {
		removeSet[tag] = true
		delete(analysis.TagFrequency, tag)
	}
	return analysis
}

func applyTagChanges(tasks []Todo2Task, analysis TagAnalysis) []Todo2Task {
	// Apply tag changes to tasks
	// This is a simplified version - full implementation would apply all changes
	return tasks
}

func buildTagRecommendations(analysis TagAnalysis) []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	// Recommend consolidating low-frequency tags
	for tag, count := range analysis.TagFrequency {
		if count < 2 {
			recommendations = append(recommendations, map[string]interface{}{
				"type":    "consolidate",
				"tag":     tag,
				"count":   count,
				"message": fmt.Sprintf("Tag '%s' is only used %d time(s), consider consolidating", tag, count),
			})
		}
	}

	return recommendations
}

// Helper functions for dependency analysis

// DependencyGraph is the legacy format for backward compatibility
type DependencyGraph map[string][]string

// buildLegacyGraphFormat converts TaskGraph to legacy map format for backward compatibility
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
				priority = "medium"
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
				priority = "medium"
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
		priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
		return priorityOrder[groups[i].Priority] < priorityOrder[groups[j].Priority]
	})

	return groups
}

// findParallelizableTasksSimple is a fallback implementation without graph analysis
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
				priority = "medium"
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
		priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
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

// Helper function to save analysis results
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

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

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
				"success":             true,
				"method":              "foundation_model",
				"total_tasks":         len(tasks),
				"pending_tasks":       len(pendingTasks),
				"classifications":     []map[string]interface{}{},
				"hierarchy_skipped":   "fm_response_not_valid_json",
				"parse_error":         err.Error(),
				"response_snippet":    snippet,
			}
			outputBytes, _ := json.MarshalIndent(analysis, "", "  ")
			return []framework.TextContent{
				{Type: "text", Text: string(outputBytes)},
			}, nil
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
