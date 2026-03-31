// tasks.go — Resource handlers for stdio://tasks/* (list, by-status, by-id, models).
package resources

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/tools"
)

// handleAllTasks handles the stdio://tasks resource
// Returns all tasks with pagination (limit 50).
func handleAllTasks(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	store := tools.NewDefaultTaskStore(projectRoot)

	allTasks, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load tasks: %w", err)
	}

	// Apply limit (50 tasks max)
	limit := 50
	tasksToReturn := allTasks

	if len(allTasks) > limit {
		tasksToReturn = allTasks[:limit]
	}

	result := map[string]interface{}{
		"tasks":     formatTasksForResource(tasksToReturn),
		"total":     len(allTasks),
		"returned":  len(tasksToReturn),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal tasks: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleTaskByID handles the stdio://tasks/{task_id} resource
// Returns a single task by ID.
func handleTaskByID(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse task_id from URI: stdio://tasks/{task_id}
	taskID, err := parseURIVariableByIndexWithValidation(uri, 3, "task ID", "stdio://tasks/{task_id}")
	if err != nil {
		return nil, "", err
	}

	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	store := tools.NewDefaultTaskStore(projectRoot)

	task, err := store.GetTask(ctx, taskID)
	if err != nil || task == nil {
		return nil, "", fmt.Errorf("task %s not found", taskID)
	}

	result := map[string]interface{}{
		"task":      formatTaskForResource(task),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	if workflowContract := tools.BuildWorkflowContract(task, nil, nil); len(workflowContract) > 0 {
		result["workflow_contract"] = workflowContract
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal task: %w", err)
	}

	return jsonData, "application/json", nil
}

// formatTasksForResource formats a slice of tasks for resource output.
func formatTasksForResource(tasks []*database.Todo2Task) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tasks))
	for i, t := range tasks {
		result[i] = formatTaskForResource(t)
	}

	return result
}

// handleTasksByStatus handles the stdio://tasks/status/{status} resource
// Returns tasks filtered by status (Todo, In Progress, Done).
func handleTasksByStatus(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse status from URI: stdio://tasks/status/{status}
	status, err := parseURIVariableByIndexWithValidation(uri, 3, "status", "stdio://tasks/status/{status}")
	if err != nil {
		return nil, "", err
	}

	// Normalize status (case-insensitive)
	status = normalizeStatus(status)

	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	store := tools.NewDefaultTaskStore(projectRoot)
	filters := &database.TaskFilters{Status: &status}

	tasks, err := store.ListTasks(ctx, filters)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load tasks: %w", err)
	}

	limit := 50
	if len(tasks) > limit {
		tasks = tasks[:limit]
	}

	result := map[string]interface{}{
		"status":    status,
		"tasks":     formatTasksForResource(tasks),
		"total":     len(tasks),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal tasks: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleTasksByPriority handles the stdio://tasks/priority/{priority} resource
// Returns tasks filtered by priority (low, medium, high, critical).
func handleTasksByPriority(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse priority from URI: stdio://tasks/priority/{priority}
	priority, err := parseURIVariableByIndexWithValidation(uri, 3, "priority", "stdio://tasks/priority/{priority}")
	if err != nil {
		return nil, "", err
	}

	// Normalize priority (case-insensitive, lowercase)
	priority = strings.ToLower(priority)

	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	store := tools.NewDefaultTaskStore(projectRoot)
	filters := &database.TaskFilters{Priority: &priority}

	tasks, err := store.ListTasks(ctx, filters)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load tasks: %w", err)
	}

	limit := 50
	if len(tasks) > limit {
		tasks = tasks[:limit]
	}

	result := map[string]interface{}{
		"priority":  priority,
		"tasks":     formatTasksForResource(tasks),
		"total":     len(tasks),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal tasks: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleTasksByTag handles the stdio://tasks/tag/{tag} resource
// Returns tasks filtered by tag (any tag value).
func handleTasksByTag(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse tag from URI: stdio://tasks/tag/{tag}
	tag, err := parseURIVariableByIndexWithValidation(uri, 3, "tag", "stdio://tasks/tag/{tag}")
	if err != nil {
		return nil, "", err
	}

	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	store := tools.NewDefaultTaskStore(projectRoot)
	filters := &database.TaskFilters{Tag: &tag}

	tasks, err := store.ListTasks(ctx, filters)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load tasks: %w", err)
	}

	limit := 50
	if len(tasks) > limit {
		tasks = tasks[:limit]
	}

	result := map[string]interface{}{
		"tag":       tag,
		"tasks":     formatTasksForResource(tasks),
		"total":     len(tasks),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal tasks: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleTasksSummary handles the stdio://tasks/summary resource
// Returns task statistics and overview.
func handleTasksSummary(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	store := tools.NewDefaultTaskStore(projectRoot)

	allTasks, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load tasks: %w", err)
	}

	// Calculate statistics
	total := len(allTasks)
	byStatus := make(map[string]int)
	byPriority := make(map[string]int)
	tagDistribution := make(map[string]int)

	for _, task := range allTasks {
		// Count by status (normalize for consistency)
		status := normalizeStatus(task.Status)
		byStatus[status]++

		// Count by priority (normalize to lowercase)
		priority := strings.ToLower(task.Priority)
		if priority == "" {
			priority = "none" // Tasks without priority
		}

		byPriority[priority]++

		// Count tags
		for _, tag := range task.Tags {
			if tag != "" {
				tagDistribution[tag]++
			}
		}
	}

	result := map[string]interface{}{
		"total":       total,
		"by_status":   byStatus,
		"by_priority": byPriority,
		"tags":        tagDistribution,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal summary: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleSuggestedTasks handles the stdio://suggested-tasks resource
// Returns dependency-ordered tasks ready to start (all dependencies Done), for MCP clients and Cursor.
func handleSuggestedTasks(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	suggested := tools.GetSuggestedNextTasks(projectRoot, 10)
	taskMaps := make([]map[string]interface{}, len(suggested))

	for i, d := range suggested {
		m := map[string]interface{}{
			"id":       d.ID,
			"content":  d.Content,
			"priority": d.Priority,
			"status":   d.Status,
			"level":    d.Level,
			"tags":     cloneStringSlice(d.Tags),
		}
		if d.Lane != "" {
			m["lane"] = d.Lane
		}
		taskMaps[i] = m
	}

	result := map[string]interface{}{
		"suggested_tasks": taskMaps,
		"count":           len(suggested),
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal suggested tasks: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleReadyTasks handles the stdio://tasks/ready resource.
// Returns dependency-ready tasks using the same payload as stdio://suggested-tasks.
func handleReadyTasks(ctx context.Context, uri string) ([]byte, string, error) {
	return handleSuggestedTasks(ctx, uri)
}

// handleTaskRuns handles the stdio://task-runs/{task_id} resource.
// Returns recent execution runs for a task.
func handleTaskRuns(ctx context.Context, uri string) ([]byte, string, error) {
	taskID, err := parseURIVariableByIndexWithValidation(uri, 3, "task ID", "stdio://task-runs/{task_id}")
	if err != nil {
		return nil, "", err
	}
	runs, err := database.ListTaskExecutionRuns(ctx, taskID, "", 10)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load task runs: %w", err)
	}
	items := make([]map[string]interface{}, 0, len(runs))
	for i := range runs {
		items = append(items, map[string]interface{}{
			"run_id":        runs[i].RunID,
			"task_id":       runs[i].TaskID,
			"agent_id":      runs[i].AgentID,
			"host":          runs[i].Host,
			"status":        runs[i].Status,
			"summary":       runs[i].Summary,
			"started_at":    runs[i].StartedAt.Format(time.RFC3339),
			"ended_at":      runs[i].EndedAt.Format(time.RFC3339),
			"files_touched": cloneStringSlice(runs[i].FilesTouched),
			"commands_run":  cloneStringSlice(runs[i].CommandsRun),
		})
	}
	result := map[string]interface{}{
		"task_id":   taskID,
		"runs":      items,
		"count":     len(items),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal task runs: %w", err)
	}
	return jsonData, "application/json", nil
}

// handleActiveWork handles the stdio://active-work resource.
// Returns active claims and active execution runs for execution-cockpit discovery.
func handleActiveWork(ctx context.Context, uri string) ([]byte, string, error) {
	activeLocks, err := database.GetActiveLocks(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load active claims: %w", err)
	}
	activeRuns, err := database.ListTaskExecutionRuns(ctx, "", "running", 25)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load active runs: %w", err)
	}

	claimMaps := make([]map[string]interface{}, 0, len(activeLocks))
	for _, lock := range activeLocks {
		claimMaps = append(claimMaps, map[string]interface{}{
			"task_id":        lock.TaskID,
			"assignee":       lock.Assignee,
			"assigned_at":    lock.AssignedAt.Format(time.RFC3339),
			"lock_until":     lock.LockUntil.Format(time.RFC3339),
			"time_remaining": lock.TimeRemaining.Round(time.Second).String(),
			"status":         "active",
		})
	}

	runMaps := make([]map[string]interface{}, 0, len(activeRuns))
	for i := range activeRuns {
		runMaps = append(runMaps, map[string]interface{}{
			"run_id":        activeRuns[i].RunID,
			"task_id":       activeRuns[i].TaskID,
			"agent_id":      activeRuns[i].AgentID,
			"host":          activeRuns[i].Host,
			"status":        activeRuns[i].Status,
			"summary":       activeRuns[i].Summary,
			"started_at":    activeRuns[i].StartedAt.Format(time.RFC3339),
			"files_touched": cloneStringSlice(activeRuns[i].FilesTouched),
			"commands_run":  cloneStringSlice(activeRuns[i].CommandsRun),
		})
	}

	result := map[string]interface{}{
		"active_claims": claimMaps,
		"active_runs":   runMaps,
		"timestamp":     time.Now().Format(time.RFC3339),
	}
	if projectRoot, err := tools.FindProjectRoot(); err == nil {
		if orchestration, err := tools.BuildExecutionOrchestrationSummary(ctx, projectRoot, activeLocks, activeRuns); err == nil {
			for key, value := range orchestration {
				result[key] = value
			}
		}
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal active work: %w", err)
	}

	return jsonData, "application/json", nil
}

// normalizeStatus normalizes status values to Title Case.
// This is a wrapper around tools.NormalizeStatusToTitleCase for backward compatibility.
func normalizeStatus(status string) string {
	return tools.NormalizeStatusToTitleCase(status)
}

// formatTaskForResource formats a single task for resource output
// Includes assignee information with task ID context when available.
func formatTaskForResource(task *database.Todo2Task) map[string]interface{} {
	if task == nil {
		return nil
	}

	result := map[string]interface{}{
		"id":               task.ID,
		"content":          task.Content,
		"long_description": task.LongDescription,
		"status":           task.Status,
		"priority":         task.Priority,
		"tags":             cloneStringSlice(task.Tags),
		"dependencies":     cloneStringSlice(task.Dependencies),
		"completed":        task.Completed,
		"metadata":         cloneMapStringAny(task.Metadata),
	}
	if rt := tools.GetRecommendedTools(task.Metadata); len(rt) > 0 {
		result["recommended_tools"] = cloneStringSlice(rt)
	}

	// Query assignee from database if available
	ctx := context.Background()

	db, err := database.GetDB()
	if err == nil {
		var assignee sql.NullString
		var lockUntil sql.NullInt64

		err := db.QueryRowContext(ctx, `
			SELECT assignee, lock_until FROM tasks WHERE id = ?
		`, task.ID).Scan(&assignee, &lockUntil)
		if err == nil && assignee.Valid && assignee.String != "" {
			result["assignee"] = fmt.Sprintf("%s (task: %s)", assignee.String, task.ID)
			if lockUntil.Valid {
				result["active_claim"] = map[string]interface{}{
					"assignee":   assignee.String,
					"lock_until": time.Unix(lockUntil.Int64, 0).Format(time.RFC3339),
					"status":     "active",
				}
			}
		}
	}

	if runs, err := database.ListTaskExecutionRuns(ctx, task.ID, "", 3); err == nil && len(runs) > 0 {
		runSummaries := make([]map[string]interface{}, 0, len(runs))
		for i := range runs {
			runSummaries = append(runSummaries, map[string]interface{}{
				"run_id":     runs[i].RunID,
				"status":     runs[i].Status,
				"summary":    runs[i].Summary,
				"started_at": runs[i].StartedAt.Format(time.RFC3339),
			})
		}
		result["recent_runs"] = runSummaries
	}

	return result
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	return append([]string(nil), in...)
}

func cloneMapStringAny(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
