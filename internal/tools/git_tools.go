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
)

// Branch tag prefix
const (
	branchTagPrefix = "branch:"
	mainBranch      = "main"
)

// GitToolsParams represents parameters for git_tools
type GitToolsParams struct {
	Action           string `json:"action"`
	TaskID           string `json:"task_id,omitempty"`
	Branch           string `json:"branch,omitempty"`
	Limit            int    `json:"limit,omitempty"`
	Commit1          string `json:"commit1,omitempty"`
	Commit2          string `json:"commit2,omitempty"`
	Time1            string `json:"time1,omitempty"`
	Time2            string `json:"time2,omitempty"`
	Format           string `json:"format,omitempty"`
	OutputPath       string `json:"output_path,omitempty"`
	MaxCommits       int    `json:"max_commits,omitempty"`
	SourceBranch     string `json:"source_branch,omitempty"`
	TargetBranch     string `json:"target_branch,omitempty"`
	ConflictStrategy string `json:"conflict_strategy,omitempty"`
	Author           string `json:"author,omitempty"`
	DryRun           bool   `json:"dry_run,omitempty"`
}

// TaskCommit represents a commit (change) to a task
type TaskCommit struct {
	ID        string                 `json:"id"`
	TaskID    string                 `json:"task_id,omitempty"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Author    string                 `json:"author"`
	Branch    string                 `json:"branch"`
	OldState  map[string]interface{} `json:"old_state,omitempty"`
	NewState  map[string]interface{} `json:"new_state,omitempty"`
}

// Branch utilities

// extractBranchFromTags extracts branch name from task tags
func extractBranchFromTags(tags []string) string {
	if tags == nil {
		return mainBranch
	}

	for _, tag := range tags {
		if strings.HasPrefix(tag, branchTagPrefix) {
			branch := strings.TrimSpace(strings.TrimPrefix(tag, branchTagPrefix))
			if branch != "" {
				return branch
			}
		}
	}

	return mainBranch
}

// getTaskBranch gets branch for a task from its tags
func getTaskBranch(task Todo2Task) string {
	branch := extractBranchFromTags(task.Tags)
	if branch == "" {
		return mainBranch
	}
	return branch
}

// setTaskBranch sets branch for a task by updating its tags
func setTaskBranch(task *Todo2Task, branch string) {
	// Remove existing branch tags
	newTags := make([]string, 0, len(task.Tags))
	for _, tag := range task.Tags {
		if !strings.HasPrefix(tag, branchTagPrefix) {
			newTags = append(newTags, tag)
		}
	}

	// Add new branch tag (skip if main branch)
	if branch != mainBranch && branch != "" {
		newTags = append(newTags, branchTagPrefix+branch)
	}

	task.Tags = newTags
}

// getAllBranches extracts all unique branches from tasks
func getAllBranches(tasks []Todo2Task) []string {
	branches := make(map[string]bool)
	branches[mainBranch] = true

	for _, task := range tasks {
		branch := getTaskBranch(task)
		branches[branch] = true
	}

	result := make([]string, 0, len(branches))
	for branch := range branches {
		result = append(result, branch)
	}
	sort.Strings(result)
	return result
}

// filterTasksByBranch filters tasks to only those in a specific branch
func filterTasksByBranch(tasks []Todo2Task, branch string) []Todo2Task {
	result := make([]Todo2Task, 0)
	for _, task := range tasks {
		if getTaskBranch(task) == branch {
			result = append(result, task)
		}
	}
	return result
}

// BranchStatistics represents statistics for a branch
type BranchStatistics struct {
	Branch         string         `json:"branch"`
	TaskCount      int            `json:"task_count"`
	ByStatus       map[string]int `json:"by_status"`
	CompletedCount int            `json:"completed_count"`
	CompletionRate float64        `json:"completion_rate"`
}

// getBranchStatistics gets statistics for a specific branch
func getBranchStatistics(tasks []Todo2Task, branch string) BranchStatistics {
	branchTasks := filterTasksByBranch(tasks, branch)

	statusCounts := make(map[string]int)
	completedCount := 0

	for _, task := range branchTasks {
		status := normalizeStatus(task.Status)
		statusCounts[status]++

		if IsCompletedStatus(status) {
			completedCount++
		}
	}

	total := len(branchTasks)
	completionRate := 0.0
	if total > 0 {
		completionRate = float64(completedCount) / float64(total) * 100
	}

	return BranchStatistics{
		Branch:         branch,
		TaskCount:      total,
		ByStatus:       statusCounts,
		CompletedCount: completedCount,
		CompletionRate: completionRate,
	}
}

// getAllBranchStatistics gets statistics for all branches
func getAllBranchStatistics(tasks []Todo2Task) map[string]BranchStatistics {
	branches := getAllBranches(tasks)
	result := make(map[string]BranchStatistics)

	for _, branch := range branches {
		result[branch] = getBranchStatistics(tasks, branch)
	}

	return result
}

// Commit tracking utilities

// loadCommits loads commits from .todo2/commits.json
func loadCommits(projectRoot string) ([]TaskCommit, error) {
	commitsPath := filepath.Join(projectRoot, ".todo2", "commits.json")

	data, err := os.ReadFile(commitsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []TaskCommit{}, nil
		}
		return nil, fmt.Errorf("failed to read commits file: %w", err)
	}

	var commitsData struct {
		Commits []TaskCommit `json:"commits"`
		Version string       `json:"version"`
	}

	if err := json.Unmarshal(data, &commitsData); err != nil {
		return nil, fmt.Errorf("failed to parse commits JSON: %w", err)
	}

	return commitsData.Commits, nil
}

// saveCommits saves commits to .todo2/commits.json
func saveCommits(projectRoot string, commits []TaskCommit) error {
	commitsPath := filepath.Join(projectRoot, ".todo2", "commits.json")

	// Ensure directory exists
	dir := filepath.Dir(commitsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create .todo2 directory: %w", err)
	}

	commitsData := struct {
		Commits []TaskCommit `json:"commits"`
		Version string       `json:"version"`
	}{
		Commits: commits,
		Version: "1.0",
	}

	data, err := json.MarshalIndent(commitsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commits: %w", err)
	}

	if err := os.WriteFile(commitsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write commits file: %w", err)
	}

	return nil
}

// HandleGitToolsNative handles git_tools using native Go implementation
func HandleGitToolsNative(ctx context.Context, params GitToolsParams) (string, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return "", fmt.Errorf("failed to find project root: %w", err)
	}

	action := params.Action
	if action == "" {
		action = "commits"
	}

	switch action {
	case "commits":
		return handleCommitsAction(ctx, projectRoot, params)
	case "branches":
		return handleBranchesAction(ctx, projectRoot, params)
	case "tasks":
		return handleTasksAction(ctx, projectRoot, params)
	case "diff":
		return handleDiffAction(ctx, projectRoot, params)
	case "graph":
		return handleGraphAction(ctx, projectRoot, params)
	case "merge":
		return handleMergeAction(ctx, projectRoot, params)
	case "set_branch":
		return handleSetBranchAction(ctx, projectRoot, params)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

// handleCommitsAction handles the commits action
func handleCommitsAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	commits, err := loadCommits(projectRoot)
	if err != nil {
		return "", err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	var filtered []TaskCommit

	if params.TaskID != "" {
		// Filter by task ID
		for _, commit := range commits {
			if commit.TaskID == params.TaskID {
				if params.Branch == "" || commit.Branch == params.Branch {
					filtered = append(filtered, commit)
				}
			}
		}
	} else if params.Branch != "" {
		// Filter by branch
		for _, commit := range commits {
			if commit.Branch == params.Branch {
				filtered = append(filtered, commit)
			}
		}
	} else {
		return "", fmt.Errorf("task_id or branch required for commits action")
	}

	// Sort by timestamp (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Limit results
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	result := map[string]interface{}{
		"task_id":       params.TaskID,
		"branch":        params.Branch,
		"total_commits": len(filtered),
		"commits":       filtered,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// handleBranchesAction handles the branches action
func handleBranchesAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return "", err
	}

	branches := getAllBranches(tasks)
	statistics := getAllBranchStatistics(tasks)

	result := map[string]interface{}{
		"branches":   branches,
		"statistics": statistics,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// handleTasksAction handles the tasks action
func handleTasksAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	if params.Branch == "" {
		return "", fmt.Errorf("branch parameter required for tasks action")
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return "", err
	}

	branchTasks := filterTasksByBranch(tasks, params.Branch)

	// Convert to JSON-compatible format
	taskList := make([]map[string]interface{}, len(branchTasks))
	for i, task := range branchTasks {
		taskMap := map[string]interface{}{
			"id":               task.ID,
			"content":          task.Content,
			"long_description": task.LongDescription,
			"status":           task.Status,
			"priority":         task.Priority,
			"tags":             task.Tags,
			"dependencies":     task.Dependencies,
			"completed":        task.Completed,
			"metadata":         task.Metadata,
		}
		taskList[i] = taskMap
	}

	result := map[string]interface{}{
		"branch":     params.Branch,
		"task_count": len(branchTasks),
		"tasks":      taskList,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// handleDiffAction returns task change history (commits for the task) as a diff-style list.
func handleDiffAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	if params.TaskID == "" {
		return "", fmt.Errorf("task_id parameter required for diff action")
	}

	commits, err := loadCommits(projectRoot)
	if err != nil {
		return "", err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	var filtered []TaskCommit
	for _, commit := range commits {
		if commit.TaskID == params.TaskID {
			if params.Branch == "" || commit.Branch == params.Branch {
				filtered = append(filtered, commit)
			}
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	result := map[string]interface{}{
		"task_id":       params.TaskID,
		"branch":        params.Branch,
		"total_changes": len(filtered),
		"changes":       filtered,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// handleGraphAction returns a minimal branch/task graph: branches with task counts and task IDs per branch.
func handleGraphAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return "", err
	}

	branches := getAllBranches(tasks)
	statistics := getAllBranchStatistics(tasks)

	// Optional: filter to one branch if requested
	if params.Branch != "" {
		if _, ok := statistics[params.Branch]; !ok {
			branches = []string{params.Branch}
			statistics = map[string]BranchStatistics{params.Branch: getBranchStatistics(tasks, params.Branch)}
		} else {
			branches = []string{params.Branch}
			statistics = map[string]BranchStatistics{params.Branch: statistics[params.Branch]}
		}
	}

	// Build graph nodes: branch -> task IDs (and stats)
	nodes := make([]map[string]interface{}, 0, len(branches))
	for _, branch := range branches {
		branchTasks := filterTasksByBranch(tasks, branch)
		taskIDs := make([]string, 0, len(branchTasks))
		for _, t := range branchTasks {
			taskIDs = append(taskIDs, t.ID)
		}
		stats := statistics[branch]
		nodes = append(nodes, map[string]interface{}{
			"branch":          branch,
			"task_count":      stats.TaskCount,
			"by_status":       stats.ByStatus,
			"completion_rate": stats.CompletionRate,
			"task_ids":        taskIDs,
		})
	}

	result := map[string]interface{}{
		"format":   params.Format,
		"branch":   params.Branch,
		"branches": branches,
		"nodes":    nodes,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// handleMergeAction handles the merge action
func handleMergeAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	if params.SourceBranch == "" || params.TargetBranch == "" {
		return "", fmt.Errorf("source_branch and target_branch required for merge action")
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return "", err
	}

	conflictStrategy := params.ConflictStrategy
	if conflictStrategy == "" {
		conflictStrategy = "newer"
	}

	if params.DryRun {
		return previewMerge(tasks, params.SourceBranch, params.TargetBranch)
	}

	return mergeBranches(ctx, projectRoot, tasks, params.SourceBranch, params.TargetBranch, conflictStrategy, params.Author)
}

// handleSetBranchAction handles the set_branch action
func handleSetBranchAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	if params.TaskID == "" || params.Branch == "" {
		return "", fmt.Errorf("task_id and branch required for set_branch action")
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return "", err
	}

	// Find task
	taskIndex := -1
	var oldTask Todo2Task
	for i, task := range tasks {
		if task.ID == params.TaskID {
			taskIndex = i
			oldTask = task
			break
		}
	}

	if taskIndex == -1 {
		return "", fmt.Errorf("task %s not found", params.TaskID)
	}

	oldBranch := getTaskBranch(oldTask)

	// Update task branch
	setTaskBranch(&tasks[taskIndex], params.Branch)

	// Save tasks
	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		return "", fmt.Errorf("failed to save tasks: %w", err)
	}

	// Track commit
	if err := trackTaskUpdate(projectRoot, params.TaskID, oldTask, tasks[taskIndex], params.Author, params.Branch); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to track commit: %v\n", err)
	}

	result := map[string]interface{}{
		"task_id":    params.TaskID,
		"old_branch": oldBranch,
		"new_branch": params.Branch,
		"success":    true,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// previewMerge previews what would happen if merging branches
func previewMerge(tasks []Todo2Task, sourceBranch, targetBranch string) (string, error) {
	sourceTasks := filterTasksByBranch(tasks, sourceBranch)
	targetTaskMap := make(map[string]Todo2Task)
	for _, task := range filterTasksByBranch(tasks, targetBranch) {
		targetTaskMap[task.ID] = task
	}

	conflicts := []map[string]interface{}{}
	newTasks := []map[string]interface{}{}

	for _, sourceTask := range sourceTasks {
		if targetTask, exists := targetTaskMap[sourceTask.ID]; exists {
			// Check if tasks differ
			if !tasksEqual(sourceTask, targetTask) {
				conflicts = append(conflicts, map[string]interface{}{
					"task_id": sourceTask.ID,
					"source":  sourceTask,
					"target":  targetTask,
				})
			}
		} else {
			// New task to be added
			taskMap := taskToMap(sourceTask)
			newTasks = append(newTasks, taskMap)
		}
	}

	result := map[string]interface{}{
		"source_branch":  sourceBranch,
		"target_branch":  targetBranch,
		"conflicts":      conflicts,
		"new_tasks":      newTasks,
		"conflict_count": len(conflicts),
		"new_task_count": len(newTasks),
		"dry_run":        true,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// mergeBranches merges tasks from source branch to target branch
func mergeBranches(ctx context.Context, projectRoot string, tasks []Todo2Task, sourceBranch, targetBranch, conflictStrategy, author string) (string, error) {
	targetTaskMap := make(map[string]int)
	for i, task := range tasks {
		if getTaskBranch(task) == targetBranch {
			targetTaskMap[task.ID] = i
		}
	}

	mergedCount := 0
	conflictCount := 0
	newCount := 0

	// Update tasks in place
	for i := range tasks {
		task := &tasks[i]
		currentBranch := getTaskBranch(*task)

		// If task is in source branch
		if currentBranch == sourceBranch {
			if targetIndex, exists := targetTaskMap[task.ID]; exists {
				// Conflict - apply strategy
				_ = tasks[targetIndex] // Reference target task for conflict resolution
				if conflictStrategy == "newer" {
					// Use task with newer updated_at timestamp (if available)
					// For now, prefer source
					setTaskBranch(task, targetBranch)
					mergedCount++
				} else if conflictStrategy == "source" {
					setTaskBranch(task, targetBranch)
					mergedCount++
				} else if conflictStrategy == "target" {
					conflictCount++ // Keep target version
				}
			} else {
				// New task - move to target branch
				setTaskBranch(task, targetBranch)
				newCount++
				mergedCount++
			}
		}
	}

	// Save updated tasks
	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		return "", fmt.Errorf("failed to save tasks: %w", err)
	}

	result := map[string]interface{}{
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"merged_count":  mergedCount,
		"new_tasks":     newCount,
		"conflicts":     conflictCount,
		"strategy":      conflictStrategy,
		"author":        author,
		"success":       true,
		"dry_run":       false,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// Helper functions

func tasksEqual(a, b Todo2Task) bool {
	return a.ID == b.ID &&
		a.Content == b.Content &&
		a.Status == b.Status &&
		a.Priority == b.Priority &&
		stringsEqual(a.Tags, b.Tags) &&
		stringsEqual(a.Dependencies, b.Dependencies) &&
		a.Completed == b.Completed
}

func stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func taskToMap(task Todo2Task) map[string]interface{} {
	return map[string]interface{}{
		"id":               task.ID,
		"content":          task.Content,
		"long_description": task.LongDescription,
		"status":           task.Status,
		"priority":         task.Priority,
		"tags":             task.Tags,
		"dependencies":     task.Dependencies,
		"completed":        task.Completed,
		"metadata":         task.Metadata,
	}
}

// trackTaskUpdate tracks a task update as a commit
func trackTaskUpdate(projectRoot string, taskID string, oldTask, newTask Todo2Task, author, branch string) error {
	commits, err := loadCommits(projectRoot)
	if err != nil {
		return err
	}

	// Create commit
	commit := TaskCommit{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		TaskID:    taskID,
		Message:   fmt.Sprintf("Updated task %s", taskID),
		Timestamp: time.Now(),
		Author:    author,
		Branch:    branch,
		OldState:  taskToMap(oldTask),
		NewState:  taskToMap(newTask),
	}

	commits = append(commits, commit)

	return saveCommits(projectRoot, commits)
}
