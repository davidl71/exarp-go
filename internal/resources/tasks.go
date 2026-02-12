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
// Returns all tasks with pagination (limit 50)
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
// Returns a single task by ID
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

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal task: %w", err)
	}

	return jsonData, "application/json", nil
}

// formatTasksForResource formats a slice of tasks for resource output
func formatTasksForResource(tasks []*database.Todo2Task) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tasks))
	for i, t := range tasks {
		result[i] = formatTaskForResource(t)
	}
	return result
}

// handleTasksByStatus handles the stdio://tasks/status/{status} resource
// Returns tasks filtered by status (Todo, In Progress, Done)
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
// Returns tasks filtered by priority (low, medium, high, critical)
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
// Returns tasks filtered by tag (any tag value)
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
// Returns task statistics and overview
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
		taskMaps[i] = map[string]interface{}{
			"id":       d.ID,
			"content":  d.Content,
			"priority": d.Priority,
			"status":   d.Status,
			"level":    d.Level,
			"tags":     d.Tags,
		}
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

// normalizeStatus normalizes status values to Title Case.
// This is a wrapper around tools.NormalizeStatusToTitleCase for backward compatibility.
func normalizeStatus(status string) string {
	return tools.NormalizeStatusToTitleCase(status)
}

// formatTaskForResource formats a single task for resource output
// Includes assignee information with task ID context when available
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
		"tags":             task.Tags,
		"dependencies":     task.Dependencies,
		"completed":        task.Completed,
		"metadata":         task.Metadata,
	}

	// Query assignee from database if available
	ctx := context.Background()
	db, err := database.GetDB()
	if err == nil {
		var assignee sql.NullString
		err := db.QueryRowContext(ctx, `
			SELECT assignee FROM tasks WHERE id = ?
		`, task.ID).Scan(&assignee)
		if err == nil && assignee.Valid && assignee.String != "" {
			// Format assignee with task ID context
			result["assignee"] = fmt.Sprintf("%s (task: %s)", assignee.String, task.ID)
		}
	}

	return result
}
