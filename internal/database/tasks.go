package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
)

// Todo2Task is an alias for models.Todo2Task (for convenience)
type Todo2Task = models.Todo2Task

// TaskFilters represents filters for querying tasks
type TaskFilters struct {
	Status   *string
	Priority *string
	Tag      *string
	ProjectID *string
}

// CreateTask creates a new task in the database
// Uses a transaction to atomically insert task, tags, and dependencies
func CreateTask(task *Todo2Task) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Marshal metadata to JSON
	var metadataJSON string
	if task.Metadata != nil && len(task.Metadata) > 0 {
		metadataBytes, err := json.Marshal(task.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	// Convert completed boolean to integer (0 or 1)
	completedInt := 0
	if task.Completed {
		completedInt = 1
	}

	// Insert task
	now := time.Now().Format(time.RFC3339)
	_, err = tx.Exec(`
		INSERT INTO tasks (
			id, name, content, long_description, status, priority, completed,
			created, last_modified, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%s', 'now'), strftime('%s', 'now'))
	`,
		task.ID,
		"", // name - TODO: add to Todo2Task struct if needed
		task.Content,
		task.LongDescription,
		task.Status,
		task.Priority,
		completedInt,
		now, // created
		now, // last_modified
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	// Insert tags
	for _, tag := range task.Tags {
		_, err = tx.Exec(`
			INSERT INTO task_tags (task_id, tag) VALUES (?, ?)
		`, task.ID, tag)
		if err != nil {
			return fmt.Errorf("failed to insert tag %s: %w", tag, err)
		}
	}

	// Insert dependencies
	for _, depID := range task.Dependencies {
		_, err = tx.Exec(`
			INSERT INTO task_dependencies (task_id, depends_on_id) VALUES (?, ?)
		`, task.ID, depID)
		if err != nil {
			return fmt.Errorf("failed to insert dependency %s: %w", depID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTask retrieves a task by ID with all related data (tags, dependencies)
func GetTask(id string) (*Todo2Task, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// Query task
	var task Todo2Task
	var metadataJSON sql.NullString
	var completedInt int

	err = db.QueryRow(`
		SELECT id, name, content, long_description, status, priority, completed,
		       created, last_modified, metadata
		FROM tasks
		WHERE id = ?
	`, id).Scan(
		&task.ID,
		nil, // name - skip for now
		&task.Content,
		&task.LongDescription,
		&task.Status,
		&task.Priority,
		&completedInt,
		nil, // created - skip for now
		nil, // last_modified - skip for now
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}

	task.Completed = completedInt == 1

	// Unmarshal metadata
	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &task.Metadata); err != nil {
			// Log but don't fail - metadata is optional
			task.Metadata = nil
		}
	}

	// Load tags
	tagRows, err := db.Query(`
		SELECT tag FROM task_tags WHERE task_id = ? ORDER BY tag
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer tagRows.Close()

	var tags []string
	for tagRows.Next() {
		var tag string
		if err := tagRows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	task.Tags = tags

	// Load dependencies
	depRows, err := db.Query(`
		SELECT depends_on_id FROM task_dependencies WHERE task_id = ? ORDER BY depends_on_id
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer depRows.Close()

	var dependencies []string
	for depRows.Next() {
		var depID string
		if err := depRows.Scan(&depID); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		dependencies = append(dependencies, depID)
	}
	task.Dependencies = dependencies

	return &task, nil
}

// UpdateTask updates an existing task
// Uses a transaction to atomically update task, tags, and dependencies
func UpdateTask(task *Todo2Task) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Marshal metadata to JSON
	var metadataJSON string
	if task.Metadata != nil && len(task.Metadata) > 0 {
		metadataBytes, err := json.Marshal(task.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	// Convert completed boolean to integer
	completedInt := 0
	if task.Completed {
		completedInt = 1
	}

	// Update task
	now := time.Now().Format(time.RFC3339)
	result, err := tx.Exec(`
		UPDATE tasks SET
			content = ?,
			long_description = ?,
			status = ?,
			priority = ?,
			completed = ?,
			last_modified = ?,
			metadata = ?,
			updated_at = strftime('%s', 'now')
		WHERE id = ?
	`,
		task.Content,
		task.LongDescription,
		task.Status,
		task.Priority,
		completedInt,
		now,
		metadataJSON,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task %s not found", task.ID)
	}

	// Delete existing tags
	_, err = tx.Exec(`DELETE FROM task_tags WHERE task_id = ?`, task.ID)
	if err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}

	// Insert new tags
	for _, tag := range task.Tags {
		_, err = tx.Exec(`
			INSERT INTO task_tags (task_id, tag) VALUES (?, ?)
		`, task.ID, tag)
		if err != nil {
			return fmt.Errorf("failed to insert tag %s: %w", tag, err)
		}
	}

	// Delete existing dependencies
	_, err = tx.Exec(`DELETE FROM task_dependencies WHERE task_id = ?`, task.ID)
	if err != nil {
		return fmt.Errorf("failed to delete dependencies: %w", err)
	}

	// Insert new dependencies
	for _, depID := range task.Dependencies {
		_, err = tx.Exec(`
			INSERT INTO task_dependencies (task_id, depends_on_id) VALUES (?, ?)
		`, task.ID, depID)
		if err != nil {
			return fmt.Errorf("failed to insert dependency %s: %w", depID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteTask deletes a task and all related data (tags, dependencies cascade)
func DeleteTask(id string) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	result, err := db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	// Tags and dependencies are cascade deleted by foreign key constraints
	return nil
}

// ListTasks retrieves tasks with optional filtering
func ListTasks(filters *TaskFilters) ([]*Todo2Task, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// Build query with filters
	query := `
		SELECT DISTINCT t.id, t.content, t.long_description, t.status, t.priority, t.completed, t.metadata
		FROM tasks t
	`
	var args []interface{}
	var conditions []string

	if filters != nil {
		if filters.Status != nil {
			conditions = append(conditions, "t.status = ?")
			args = append(args, *filters.Status)
		}
		if filters.Priority != nil {
			conditions = append(conditions, "t.priority = ?")
			args = append(args, *filters.Priority)
		}
		if filters.Tag != nil {
			query += ` INNER JOIN task_tags tt ON t.id = tt.task_id `
			conditions = append(conditions, "tt.tag = ?")
			args = append(args, *filters.Tag)
		}
		if filters.ProjectID != nil {
			conditions = append(conditions, "t.project_id = ?")
			args = append(args, *filters.ProjectID)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY t.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Todo2Task
	for rows.Next() {
		var task Todo2Task
		var metadataJSON sql.NullString
		var completedInt int

		if err := rows.Scan(
			&task.ID,
			&task.Content,
			&task.LongDescription,
			&task.Status,
			&task.Priority,
			&completedInt,
			&metadataJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		task.Completed = completedInt == 1

		// Unmarshal metadata
		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &task.Metadata); err != nil {
				task.Metadata = nil
			}
		}

		// Load tags and dependencies (could be optimized with batch queries)
		tags, err := GetTagsForTask(task.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags for task %s: %w", task.ID, err)
		}
		task.Tags = tags

		deps, err := GetDependencies(task.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get dependencies for task %s: %w", task.ID, err)
		}
		task.Dependencies = deps

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tasks, nil
}

// GetTasksByStatus retrieves all tasks with the specified status
func GetTasksByStatus(status string) ([]*Todo2Task, error) {
	filters := &TaskFilters{Status: &status}
	return ListTasks(filters)
}

// GetTasksByTag retrieves all tasks with the specified tag
func GetTasksByTag(tag string) ([]*Todo2Task, error) {
	filters := &TaskFilters{Tag: &tag}
	return ListTasks(filters)
}

// GetTasksByPriority retrieves all tasks with the specified priority
func GetTasksByPriority(priority string) ([]*Todo2Task, error) {
	filters := &TaskFilters{Priority: &priority}
	return ListTasks(filters)
}

// GetDependencies retrieves all task IDs that the specified task depends on
func GetDependencies(taskID string) ([]string, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	rows, err := db.Query(`
		SELECT depends_on_id FROM task_dependencies WHERE task_id = ? ORDER BY depends_on_id
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	var dependencies []string
	for rows.Next() {
		var depID string
		if err := rows.Scan(&depID); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		dependencies = append(dependencies, depID)
	}

	return dependencies, rows.Err()
}

// GetDependents retrieves all task IDs that depend on the specified task
func GetDependents(taskID string) ([]string, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	rows, err := db.Query(`
		SELECT task_id FROM task_dependencies WHERE depends_on_id = ? ORDER BY task_id
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependents: %w", err)
	}
	defer rows.Close()

	var dependents []string
	for rows.Next() {
		var dependentID string
		if err := rows.Scan(&dependentID); err != nil {
			return nil, fmt.Errorf("failed to scan dependent: %w", err)
		}
		dependents = append(dependents, dependentID)
	}

	return dependents, rows.Err()
}

// GetTagsForTask is a helper function to retrieve tags for a task
func GetTagsForTask(taskID string) ([]string, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	rows, err := db.Query(`
		SELECT tag FROM task_tags WHERE task_id = ? ORDER BY tag
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

