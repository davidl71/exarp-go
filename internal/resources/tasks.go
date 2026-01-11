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

	// Try database first
	var allTasks []*database.Todo2Task
	tasks, err := database.ListTasks(ctx, nil)
	if err == nil && len(tasks) > 0 {
		// Database available and has tasks
		allTasks = tasks
	} else {
		// Fallback to JSON
		jsonTasks, err := tools.LoadTodo2Tasks(projectRoot)
		if err != nil {
			return nil, "", fmt.Errorf("failed to load tasks: %w", err)
		}

		// Convert to pointer slice
		allTasks = make([]*database.Todo2Task, len(jsonTasks))
		for i := range jsonTasks {
			allTasks[i] = &jsonTasks[i]
		}
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

	// Try database first
	task, err := database.GetTask(ctx, taskID)
	if err == nil && task != nil {
		// Database available and task found
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

	// Fallback to JSON
	tasks, err := tools.LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load tasks: %w", err)
	}

	// Find task by ID
	for _, t := range tasks {
		if t.ID == taskID {
			result := map[string]interface{}{
				"task":      formatTaskForResource(&t),
				"timestamp": time.Now().Format(time.RFC3339),
			}

			jsonData, err := json.Marshal(result)
			if err != nil {
				return nil, "", fmt.Errorf("failed to marshal task: %w", err)
			}

			return jsonData, "application/json", nil
		}
	}

	return nil, "", fmt.Errorf("task %s not found", taskID)
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

	// Try database first
	tasks, err := database.GetTasksByStatus(ctx, status)
	if err == nil && tasks != nil {
		// Database available and has results (or empty results)
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

	// Fallback to JSON
	allTasks, err := tools.LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load tasks: %w", err)
	}

	// Filter by status in memory
	filtered := []*database.Todo2Task{}
	for i := range allTasks {
		if normalizeStatus(allTasks[i].Status) == status {
			filtered = append(filtered, &allTasks[i])
		}
	}

	limit := 50
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	result := map[string]interface{}{
		"status":    status,
		"tasks":     formatTasksForResource(filtered),
		"total":     len(filtered),
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

	// Try database first
	tasks, err := database.GetTasksByPriority(ctx, priority)
	if err == nil && tasks != nil {
		// Database available and has results (or empty results)
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

	// Fallback to JSON
	allTasks, err := tools.LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load tasks: %w", err)
	}

	// Filter by priority in memory
	filtered := []*database.Todo2Task{}
	for i := range allTasks {
		if strings.ToLower(allTasks[i].Priority) == priority {
			filtered = append(filtered, &allTasks[i])
		}
	}

	limit := 50
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	result := map[string]interface{}{
		"priority":  priority,
		"tasks":     formatTasksForResource(filtered),
		"total":     len(filtered),
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

	// Try database first
	tasks, err := database.GetTasksByTag(ctx, tag)
	if err == nil && tasks != nil {
		// Database available and has results (or empty results)
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

	// Fallback to JSON
	allTasks, err := tools.LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load tasks: %w", err)
	}

	// Filter by tag in memory (tasks can have multiple tags, match any)
	filtered := []*database.Todo2Task{}
	for i := range allTasks {
		for _, taskTag := range allTasks[i].Tags {
			if taskTag == tag {
				filtered = append(filtered, &allTasks[i])
				break // Found match, no need to check other tags
			}
		}
	}

	limit := 50
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	result := map[string]interface{}{
		"tag":       tag,
		"tasks":     formatTasksForResource(filtered),
		"total":     len(filtered),
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

	// Load all tasks (try database first)
	var allTasks []*database.Todo2Task
	tasks, err := database.ListTasks(ctx, nil)
	if err == nil && tasks != nil {
		// Database available
		allTasks = tasks
	} else {
		// Fallback to JSON
		jsonTasks, err := tools.LoadTodo2Tasks(projectRoot)
		if err != nil {
			return nil, "", fmt.Errorf("failed to load tasks: %w", err)
		}

		// Convert to pointer slice
		allTasks = make([]*database.Todo2Task, len(jsonTasks))
		for i := range jsonTasks {
			allTasks[i] = &jsonTasks[i]
		}
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

// normalizeStatus normalizes status values (case-insensitive, proper case)
// Returns: "Todo", "In Progress", "Done" or original if not recognized
func normalizeStatus(status string) string {
	lower := strings.ToLower(status)
	switch lower {
	case "todo":
		return "Todo"
	case "in progress", "inprogress":
		return "In Progress"
	case "done":
		return "Done"
	default:
		// Return original if not recognized (allow custom statuses)
		return status
	}
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
