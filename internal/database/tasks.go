package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
)

// Todo2Task is an alias for models.Todo2Task (for convenience)
type Todo2Task = models.Todo2Task

// TaskFilters represents filters for querying tasks
type TaskFilters struct {
	Status    *string
	Priority  *string
	Tag       *string
	ProjectID *string
}

// CreateTask creates a new task in the database
// Uses a transaction to atomically insert task, tags, and dependencies
// Supports context for timeout and cancellation
func CreateTask(ctx context.Context, task *Todo2Task) error {
	ctx = ensureContext(ctx)
	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	return retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		tx, err := db.BeginTx(txCtx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			}
		}()

		// Convert completed boolean to integer (0 or 1)
		completedInt := 0
		if task.Completed {
			completedInt = 1
		}

		// Serialize task to protobuf (preferred format)
		var metadataProtobuf []byte
		var metadataFormat string = "protobuf"
		protobufData, err := models.SerializeTaskToProtobuf(task)
		if err != nil {
			// Fall back to JSON if protobuf serialization fails
			metadataFormat = "json"
		} else {
			metadataProtobuf = protobufData
		}

		// Also store JSON for backward compatibility
		var metadataJSON string
		if task.Metadata != nil && len(task.Metadata) > 0 {
			metadataBytes, jsonErr := json.Marshal(task.Metadata)
			if jsonErr != nil {
				return fmt.Errorf("failed to marshal metadata: %w", jsonErr)
			}
			metadataJSON = string(metadataBytes)
		}

		// Insert task with protobuf data (if available) and JSON (for compatibility)
		now := time.Now().Format(time.RFC3339)
		_, err = tx.ExecContext(txCtx, `
			INSERT INTO tasks (
				id, name, content, long_description, status, priority, completed,
				created, last_modified, metadata, metadata_protobuf, metadata_format, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%s', 'now'), strftime('%s', 'now'))
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
			metadataJSON,      // JSON for backward compatibility
			metadataProtobuf, // Protobuf binary data (nil if serialization failed)
			metadataFormat,    // Format indicator
		)
		if err != nil {
			return fmt.Errorf("failed to insert task: %w", err)
		}

		// Insert tags (batch insert for better performance)
		if len(task.Tags) > 0 {
			placeholders := make([]string, len(task.Tags))
			args := make([]interface{}, len(task.Tags)*2)
			for i, tag := range task.Tags {
				placeholders[i] = "(?, ?)"
				args[i*2] = task.ID
				args[i*2+1] = tag
			}
			_, err = tx.ExecContext(txCtx, `
				INSERT INTO task_tags (task_id, tag) VALUES `+strings.Join(placeholders, ", "),
				args...)
			if err != nil {
				return fmt.Errorf("failed to insert tags: %w", err)
			}
		}

		// Insert dependencies (batch insert for better performance)
		if len(task.Dependencies) > 0 {
			placeholders := make([]string, len(task.Dependencies))
			args := make([]interface{}, len(task.Dependencies)*2)
			for i, depID := range task.Dependencies {
				placeholders[i] = "(?, ?)"
				args[i*2] = task.ID
				args[i*2+1] = depID
			}
			_, err = tx.ExecContext(txCtx, `
				INSERT INTO task_dependencies (task_id, depends_on_id) VALUES `+strings.Join(placeholders, ", "),
				args...)
			if err != nil {
				return fmt.Errorf("failed to insert dependencies: %w", err)
			}
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
}

// GetTask retrieves a task by ID with all related data (tags, dependencies)
// Supports context for timeout and cancellation
func GetTask(ctx context.Context, id string) (*Todo2Task, error) {
	ctx = ensureContext(ctx)
	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var task *Todo2Task
	err := retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		// Query task (include protobuf columns if they exist)
		var taskData Todo2Task
		var metadataJSON sql.NullString
		var metadataProtobuf []byte // BLOB column
		var metadataFormat sql.NullString
		var completedInt int
		var name sql.NullString         // name field (not used in Todo2Task struct yet)
		var created sql.NullString      // created field
		var lastModified sql.NullString // last_modified field

		// Try to query with protobuf columns first (new schema)
		err = db.QueryRowContext(queryCtx, `
			SELECT id, name, content, long_description, status, priority, completed,
			       created, last_modified, metadata, metadata_protobuf, metadata_format
			FROM tasks
			WHERE id = ?
		`, id).Scan(
			&taskData.ID,
			&name, // name - scan but don't use (field not in Todo2Task struct yet)
			&taskData.Content,
			&taskData.LongDescription,
			&taskData.Status,
			&taskData.Priority,
			&completedInt,
			&created,      // created - scan but don't use
			&lastModified, // last_modified - scan but don't use
			&metadataJSON,
			&metadataProtobuf,
			&metadataFormat,
		)

		// If protobuf columns don't exist, fall back to old schema
		if err != nil && strings.Contains(err.Error(), "no such column") {
			err = db.QueryRowContext(queryCtx, `
				SELECT id, name, content, long_description, status, priority, completed,
				       created, last_modified, metadata
				FROM tasks
				WHERE id = ?
			`, id).Scan(
				&taskData.ID,
				&name,
				&taskData.Content,
				&taskData.LongDescription,
				&taskData.Status,
				&taskData.Priority,
				&completedInt,
				&created,
				&lastModified,
				&metadataJSON,
			)
		}

		if err == sql.ErrNoRows {
			return fmt.Errorf("task %s not found", id)
		}
		if err != nil {
			return fmt.Errorf("failed to query task: %w", err)
		}

		taskData.Completed = completedInt == 1

		// Deserialize metadata: prefer protobuf if available, fall back to JSON
		if metadataFormat.Valid && metadataFormat.String == "protobuf" && len(metadataProtobuf) > 0 {
			// Deserialize from protobuf
			deserializedTask, err := models.DeserializeTaskFromProtobuf(metadataProtobuf)
			if err == nil {
				// Use metadata from protobuf deserialization
				taskData.Metadata = deserializedTask.Metadata
			} else {
				// Protobuf deserialization failed, fall back to JSON
				if metadataJSON.Valid && metadataJSON.String != "" {
					if err := json.Unmarshal([]byte(metadataJSON.String), &taskData.Metadata); err != nil {
						taskData.Metadata = nil
					}
				}
			}
		} else if metadataJSON.Valid && metadataJSON.String != "" {
			// Use JSON format (legacy or fallback)
			if err := json.Unmarshal([]byte(metadataJSON.String), &taskData.Metadata); err != nil {
				// Log but don't fail - metadata is optional
				taskData.Metadata = nil
			}
		}

		// Load tags
		tags, err := loadTaskTags(ctx, queryCtx, db, id)
		if err != nil {
			return err
		}
		taskData.Tags = tags

		// Load dependencies
		dependencies, err := loadTaskDependencies(ctx, queryCtx, db, id)
		if err != nil {
			return err
		}
		taskData.Dependencies = dependencies

		task = &taskData
		return nil
	})

	if err != nil {
		return nil, err
	}
	return task, nil
}

// loadTaskTags loads tags for a single task
// Works with both *sql.DB (via QueryContext) and *sql.Tx (via QueryContext)
func loadTaskTags(ctx context.Context, queryCtx context.Context, querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, taskID string) ([]string, error) {
	rows, err := querier.QueryContext(queryCtx, `
		SELECT tag FROM task_tags WHERE task_id = ? ORDER BY tag
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - this is cleanup
		}
	}()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag rows: %w", err)
	}

	return tags, nil
}

// loadTaskDependencies loads dependencies for a single task
// Works with both *sql.DB (via QueryContext) and *sql.Tx (via QueryContext)
func loadTaskDependencies(ctx context.Context, queryCtx context.Context, querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, taskID string) ([]string, error) {
	rows, err := querier.QueryContext(queryCtx, `
		SELECT depends_on_id FROM task_dependencies WHERE task_id = ? ORDER BY depends_on_id
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - this is cleanup
		}
	}()

	var dependencies []string
	for rows.Next() {
		var depID string
		if err := rows.Scan(&depID); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		dependencies = append(dependencies, depID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dependency rows: %w", err)
	}

	return dependencies, nil
}

// UpdateTask updates an existing task
// Uses a transaction to atomically update task, tags, and dependencies
// Supports context for timeout and cancellation
func UpdateTask(ctx context.Context, task *Todo2Task) error {
	ctx = ensureContext(ctx)
	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	return retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		tx, err := db.BeginTx(txCtx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			}
		}()

		// Convert completed boolean to integer
		completedInt := 0
		if task.Completed {
			completedInt = 1
		}

		// Serialize task to protobuf (preferred format)
		var metadataProtobuf []byte
		var metadataFormat string = "protobuf"
		protobufData, err := models.SerializeTaskToProtobuf(task)
		if err != nil {
			// Fall back to JSON if protobuf serialization fails
			metadataFormat = "json"
		} else {
			metadataProtobuf = protobufData
		}

		// Also store JSON for backward compatibility
		var metadataJSON string
		if task.Metadata != nil && len(task.Metadata) > 0 {
			metadataBytes, jsonErr := json.Marshal(task.Metadata)
			if jsonErr != nil {
				return fmt.Errorf("failed to marshal metadata: %w", jsonErr)
			}
			metadataJSON = string(metadataBytes)
		}

		// Get current version for optimistic locking
		var currentVersion int64
		err = tx.QueryRowContext(txCtx, `SELECT version FROM tasks WHERE id = ?`, task.ID).Scan(&currentVersion)
		if err == sql.ErrNoRows {
			return fmt.Errorf("task %s not found", task.ID)
		}
		if err != nil {
			return fmt.Errorf("failed to query task version: %w", err)
		}

		// Update task with optimistic locking (version check)
		// Include protobuf columns if they exist in schema
		now := time.Now().Format(time.RFC3339)
		result, err := tx.ExecContext(txCtx, `
			UPDATE tasks SET
				content = ?,
				long_description = ?,
				status = ?,
				priority = ?,
				completed = ?,
				last_modified = ?,
				metadata = ?,
				metadata_protobuf = ?,
				metadata_format = ?,
				version = version + 1,
				updated_at = strftime('%s', 'now')
			WHERE id = ? AND version = ?
		`,
			task.Content,
			task.LongDescription,
			task.Status,
			task.Priority,
			completedInt,
			now,
			metadataJSON,      // JSON for backward compatibility
			metadataProtobuf, // Protobuf binary data
			metadataFormat,    // Format indicator
			task.ID,
			currentVersion,
		)
		// If protobuf columns don't exist, fall back to old schema
		if err != nil && strings.Contains(err.Error(), "no such column") {
			result, err = tx.ExecContext(txCtx, `
				UPDATE tasks SET
					content = ?,
					long_description = ?,
					status = ?,
					priority = ?,
					completed = ?,
					last_modified = ?,
					metadata = ?,
					version = version + 1,
					updated_at = strftime('%s', 'now')
				WHERE id = ? AND version = ?
			`,
				task.Content,
				task.LongDescription,
				task.Status,
				task.Priority,
				completedInt,
				now,
				metadataJSON,
				task.ID,
				currentVersion,
			)
		}
		if err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("task %s not found or was modified by another agent (version mismatch)", task.ID)
		}

		// Delete existing tags
		_, err = tx.ExecContext(txCtx, `DELETE FROM task_tags WHERE task_id = ?`, task.ID)
		if err != nil {
			return fmt.Errorf("failed to delete tags: %w", err)
		}

		// Insert new tags (batch insert for better performance)
		if len(task.Tags) > 0 {
			placeholders := make([]string, len(task.Tags))
			args := make([]interface{}, len(task.Tags)*2)
			for i, tag := range task.Tags {
				placeholders[i] = "(?, ?)"
				args[i*2] = task.ID
				args[i*2+1] = tag
			}
			_, err = tx.ExecContext(txCtx, `
				INSERT INTO task_tags (task_id, tag) VALUES `+strings.Join(placeholders, ", "),
				args...)
			if err != nil {
				return fmt.Errorf("failed to insert tags: %w", err)
			}
		}

		// Delete existing dependencies
		_, err = tx.ExecContext(txCtx, `DELETE FROM task_dependencies WHERE task_id = ?`, task.ID)
		if err != nil {
			return fmt.Errorf("failed to delete dependencies: %w", err)
		}

		// Insert new dependencies (batch insert for better performance)
		if len(task.Dependencies) > 0 {
			placeholders := make([]string, len(task.Dependencies))
			args := make([]interface{}, len(task.Dependencies)*2)
			for i, depID := range task.Dependencies {
				placeholders[i] = "(?, ?)"
				args[i*2] = task.ID
				args[i*2+1] = depID
			}
			_, err = tx.ExecContext(txCtx, `
				INSERT INTO task_dependencies (task_id, depends_on_id) VALUES `+strings.Join(placeholders, ", "),
				args...)
			if err != nil {
				return fmt.Errorf("failed to insert dependencies: %w", err)
			}
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
}

// DeleteTask deletes a task and all related data (tags, dependencies cascade)
// Supports context for timeout and cancellation
func DeleteTask(ctx context.Context, id string) error {
	ctx = ensureContext(ctx)
	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	return retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		result, err := db.ExecContext(queryCtx, `DELETE FROM tasks WHERE id = ?`, id)
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
	})
}

// ListTasks retrieves tasks with optional filtering
// Supports context for timeout and cancellation
func ListTasks(ctx context.Context, filters *TaskFilters) ([]*Todo2Task, error) {
	ctx = ensureContext(ctx)
	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var tasks []*Todo2Task
	err := retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		// Build query with filters (using strings.Builder for better performance)
		// Include protobuf columns if they exist (for new schema)
		var queryBuilder strings.Builder
		queryBuilder.WriteString(`
			SELECT DISTINCT t.id, t.content, t.long_description, t.status, t.priority, t.completed, t.metadata, t.metadata_protobuf, t.metadata_format
			FROM tasks t
		`)
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
				queryBuilder.WriteString(` INNER JOIN task_tags tt ON t.id = tt.task_id `)
				conditions = append(conditions, "tt.tag = ?")
				args = append(args, *filters.Tag)
			}
			if filters.ProjectID != nil {
				conditions = append(conditions, "t.project_id = ?")
				args = append(args, *filters.ProjectID)
			}
		}

		if len(conditions) > 0 {
			queryBuilder.WriteString(" WHERE " + conditions[0])
			for i := 1; i < len(conditions); i++ {
				queryBuilder.WriteString(" AND " + conditions[i])
			}
		}

		queryBuilder.WriteString(" ORDER BY t.created_at DESC")
		query := queryBuilder.String()

		// Try to query with protobuf columns first
		rows, err := db.QueryContext(queryCtx, query, args...)
		hasProtobufColumns := true
		if err != nil && strings.Contains(err.Error(), "no such column") {
			// Protobuf columns don't exist, use old schema
			hasProtobufColumns = false
			queryBuilderOld := strings.Builder{}
			queryBuilderOld.WriteString(`
				SELECT DISTINCT t.id, t.content, t.long_description, t.status, t.priority, t.completed, t.metadata
				FROM tasks t
			`)
			if len(conditions) > 0 {
				queryBuilderOld.WriteString(" WHERE " + conditions[0])
				for i := 1; i < len(conditions); i++ {
					queryBuilderOld.WriteString(" AND " + conditions[i])
				}
			}
			queryBuilderOld.WriteString(" ORDER BY t.created_at DESC")
			rows, err = db.QueryContext(queryCtx, queryBuilderOld.String(), args...)
			if err != nil {
				return fmt.Errorf("failed to query tasks: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to query tasks: %w", err)
		}
		defer rows.Close()

		var taskList []*Todo2Task
		var taskIDs []string
		taskMap := make(map[string]*Todo2Task)

		// First pass: collect all tasks and their IDs
		for rows.Next() {
			var task Todo2Task
			var metadataJSON sql.NullString
			var metadataProtobuf []byte // BLOB column
			var metadataFormat sql.NullString
			var completedInt int

			// Scan based on whether protobuf columns exist
			var scanErr error
			if hasProtobufColumns {
				scanErr = rows.Scan(
					&task.ID,
					&task.Content,
					&task.LongDescription,
					&task.Status,
					&task.Priority,
					&completedInt,
					&metadataJSON,
					&metadataProtobuf,
					&metadataFormat,
				)
			} else {
				scanErr = rows.Scan(
					&task.ID,
					&task.Content,
					&task.LongDescription,
					&task.Status,
					&task.Priority,
					&completedInt,
					&metadataJSON,
				)
			}

			if scanErr != nil {
				return fmt.Errorf("failed to scan task: %w", scanErr)
			}

			task.Completed = completedInt == 1

			// Deserialize metadata: prefer protobuf if available, fall back to JSON
			if metadataFormat.Valid && metadataFormat.String == "protobuf" && len(metadataProtobuf) > 0 {
				// Deserialize from protobuf
				deserializedTask, err := models.DeserializeTaskFromProtobuf(metadataProtobuf)
				if err == nil {
					// Use metadata from protobuf deserialization
					task.Metadata = deserializedTask.Metadata
				} else {
					// Protobuf deserialization failed, fall back to JSON
					if metadataJSON.Valid && metadataJSON.String != "" {
						if err := json.Unmarshal([]byte(metadataJSON.String), &task.Metadata); err != nil {
							task.Metadata = nil
						}
					}
				}
			} else if metadataJSON.Valid && metadataJSON.String != "" {
				// Use JSON format (legacy or fallback)
				if err := json.Unmarshal([]byte(metadataJSON.String), &task.Metadata); err != nil {
					task.Metadata = nil
				}
			}

			taskIDs = append(taskIDs, task.ID)
			taskMap[task.ID] = &task
			taskList = append(taskList, &task)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating rows: %w", err)
		}

		// Batch load all tags and dependencies in 2 queries instead of N*2 queries
		if len(taskIDs) > 0 {
			// Batch load tags
			placeholders := make([]string, len(taskIDs))
			tagArgs := make([]interface{}, len(taskIDs))
			for i, id := range taskIDs {
				placeholders[i] = "?"
				tagArgs[i] = id
			}
			tagRows, err := db.QueryContext(queryCtx, `
				SELECT task_id, tag FROM task_tags 
				WHERE task_id IN (`+strings.Join(placeholders, ", ")+`) 
				ORDER BY task_id, tag
			`, tagArgs...)
			if err != nil {
				return fmt.Errorf("failed to batch query tags: %w", err)
			}
			defer tagRows.Close()

			for tagRows.Next() {
				var taskID, tag string
				if err := tagRows.Scan(&taskID, &tag); err != nil {
					return fmt.Errorf("failed to scan tag: %w", err)
				}
				if task, ok := taskMap[taskID]; ok {
					task.Tags = append(task.Tags, tag)
				}
			}
			if err = tagRows.Err(); err != nil {
				return fmt.Errorf("error iterating tag rows: %w", err)
			}

			// Batch load dependencies
			depRows, err := db.QueryContext(queryCtx, `
				SELECT task_id, depends_on_id FROM task_dependencies 
				WHERE task_id IN (`+strings.Join(placeholders, ", ")+`) 
				ORDER BY task_id, depends_on_id
			`, tagArgs...)
			if err != nil {
				return fmt.Errorf("failed to batch query dependencies: %w", err)
			}
			defer depRows.Close()

			for depRows.Next() {
				var taskID, depID string
				if err := depRows.Scan(&taskID, &depID); err != nil {
					return fmt.Errorf("failed to scan dependency: %w", err)
				}
				if task, ok := taskMap[taskID]; ok {
					task.Dependencies = append(task.Dependencies, depID)
				}
			}
			if err = depRows.Err(); err != nil {
				return fmt.Errorf("error iterating dependency rows: %w", err)
			}
		}

		tasks = taskList
		return nil
	})

	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTasksByStatus retrieves all tasks with the specified status
func GetTasksByStatus(ctx context.Context, status string) ([]*Todo2Task, error) {
	filters := &TaskFilters{Status: &status}
	return ListTasks(ctx, filters)
}

// GetTasksByTag retrieves all tasks with the specified tag
func GetTasksByTag(ctx context.Context, tag string) ([]*Todo2Task, error) {
	filters := &TaskFilters{Tag: &tag}
	return ListTasks(ctx, filters)
}

// GetTasksByPriority retrieves all tasks with the specified priority
func GetTasksByPriority(ctx context.Context, priority string) ([]*Todo2Task, error) {
	filters := &TaskFilters{Priority: &priority}
	return ListTasks(ctx, filters)
}

// GetDependencies retrieves all task IDs that the specified task depends on
func GetDependencies(taskID string) ([]string, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	ctx := context.Background()
	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	return loadTaskDependencies(ctx, queryCtx, db, taskID)
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

	ctx := context.Background()
	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	return loadTaskTags(ctx, queryCtx, db, taskID)
}
