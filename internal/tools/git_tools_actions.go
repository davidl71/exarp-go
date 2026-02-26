// git_tools_actions.go — Git tools: action handlers (branches, tasks, diff, graph, merge, set-branch).
// See also: git_tools.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidl71/exarp-go/internal/database"
	"sort"
	"time"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleBranchesAction
//   handleTasksAction — handleTasksAction handles the tasks action.
//   handleDiffAction — handleDiffAction returns task change history (commits for the task) as a diff-style list.
//   handleGraphAction — handleGraphAction returns a minimal branch/task graph: branches with task counts and task IDs per branch.
//   handleMergeAction — handleMergeAction handles the merge action.
//   handleSetBranchAction — handleSetBranchAction handles the set_branch action.
//   previewMerge — previewMerge previews what would happen if merging branches.
//   mergeBranches — mergeBranches merges tasks from source branch to target branch.
//   tasksEqual — Helper functions
//   stringsEqual
//   taskToMap
//   trackTaskUpdate — trackTaskUpdate tracks a task update as a commit.
// ────────────────────────────────────────────────────────────────────────────

// ─── handleBranchesAction ───────────────────────────────────────────────────
func handleBranchesAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return "", err
	}

	tasks := tasksFromPtrs(list)

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

// ─── handleTasksAction ──────────────────────────────────────────────────────
// handleTasksAction handles the tasks action.
func handleTasksAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	if params.Branch == "" {
		return "", fmt.Errorf("branch parameter required for tasks action")
	}

	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return "", err
	}

	tasks := tasksFromPtrs(list)

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

// ─── handleDiffAction ───────────────────────────────────────────────────────
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

// ─── handleGraphAction ──────────────────────────────────────────────────────
// handleGraphAction returns a minimal branch/task graph: branches with task counts and task IDs per branch.
func handleGraphAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return "", err
	}

	tasks := tasksFromPtrs(list)

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

// ─── handleMergeAction ──────────────────────────────────────────────────────
// handleMergeAction handles the merge action.
func handleMergeAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	if params.SourceBranch == "" || params.TargetBranch == "" {
		return "", fmt.Errorf("source_branch and target_branch required for merge action")
	}

	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return "", err
	}

	tasks := tasksFromPtrs(list)

	conflictStrategy := params.ConflictStrategy
	if conflictStrategy == "" {
		conflictStrategy = "newer"
	}

	if params.DryRun {
		return previewMerge(tasks, params.SourceBranch, params.TargetBranch)
	}

	return mergeBranches(ctx, store, tasks, params.SourceBranch, params.TargetBranch, conflictStrategy, params.Author)
}

// ─── handleSetBranchAction ──────────────────────────────────────────────────
// handleSetBranchAction handles the set_branch action.
func handleSetBranchAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	if params.TaskID == "" || params.Branch == "" {
		return "", fmt.Errorf("task_id and branch required for set_branch action")
	}

	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return "", err
	}

	tasks := tasksFromPtrs(list)

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

	// Save task via store
	if err := store.UpdateTask(ctx, &tasks[taskIndex]); err != nil {
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

// ─── previewMerge ───────────────────────────────────────────────────────────
// previewMerge previews what would happen if merging branches.
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

// ─── mergeBranches ──────────────────────────────────────────────────────────
// mergeBranches merges tasks from source branch to target branch.
func mergeBranches(ctx context.Context, store database.TaskStore, tasks []Todo2Task, sourceBranch, targetBranch, conflictStrategy, author string) (string, error) {
	targetTaskMap := make(map[string]int)

	for i, task := range tasks {
		if getTaskBranch(task) == targetBranch {
			targetTaskMap[task.ID] = i
		}
	}

	mergedCount := 0
	conflictCount := 0
	newCount := 0

	var modified []*Todo2Task

	// Update tasks in place
	for i := range tasks {
		task := &tasks[i]
		currentBranch := getTaskBranch(*task)

		// If task is in source branch
		if currentBranch == sourceBranch {
			if targetIndex, exists := targetTaskMap[task.ID]; exists {
				// Conflict - apply strategy
				_ = tasks[targetIndex] // Reference target task for conflict resolution

				switch conflictStrategy {
				case "newer":
					// Use task with newer updated_at timestamp (if available)
					// For now, prefer source
					setTaskBranch(task, targetBranch)
					modified = append(modified, task)
					mergedCount++
				case "source":
					setTaskBranch(task, targetBranch)
					modified = append(modified, task)
					mergedCount++
				case "target":
					conflictCount++ // Keep target version
				}
			} else {
				// New task - move to target branch
				setTaskBranch(task, targetBranch)
				modified = append(modified, task)
				newCount++
				mergedCount++
			}
		}
	}

	// Save updated tasks via store
	for _, t := range modified {
		if err := store.UpdateTask(ctx, t); err != nil {
			return "", fmt.Errorf("failed to save task %s: %w", t.ID, err)
		}
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

// ─── tasksEqual ─────────────────────────────────────────────────────────────
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

// ─── stringsEqual ───────────────────────────────────────────────────────────
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

// ─── taskToMap ──────────────────────────────────────────────────────────────
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

// ─── trackTaskUpdate ────────────────────────────────────────────────────────
// trackTaskUpdate tracks a task update as a commit.
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
