// tasks_crud.go — Create, Get, Update, Delete task operations and version conflict helpers.
// Same package as tasks.go; uses types and metadata helpers from tasks.go.
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// CreateTask creates a new task in the database
// Uses a transaction to atomically insert task, tags, and dependencies
// Supports context for timeout and cancellation
// If task.ID is invalid (e.g. T-NaN from MCP), it is replaced with a generated ID.
func CreateTask(ctx context.Context, task *Todo2Task) error {
	if !IsValidTaskID(task.ID) {
		task.ID = GenerateTaskID()
	}

	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	return retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
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

		metadataJSON, metadataProtobuf, metadataFormat, err := SerializeTaskMetadata(task)
		if err != nil {
			return err
		}

		// Auto-populate distributed tracking fields if not set
		if task.Host == "" {
			if h, e := os.Hostname(); e == nil {
				task.Host = h
			} else {
				task.Host = "unknown"
			}
		}
		if task.Agent == "" {
			if a, e := GetAgentID(); e == nil {
				task.Agent = a
			}
		}

		// Insert task with protobuf data (if available) and JSON (for compatibility)
		now := time.Now().Format(time.RFC3339)
		// v9 schema: parent_id, assigned_to, host, agent are NOT NULL DEFAULT ''
		// Ensure non-empty strings to satisfy NOT NULL constraints in any schema version
		parentID := task.ParentID
		if parentID == "" {
			parentID = ""
		}
		projectID := task.ProjectID // project_id still nullable
		assignedTo := task.AssignedTo
		if assignedTo == "" {
			assignedTo = ""
		}
		host := task.Host
		if host == "" {
			if h, e := os.Hostname(); e == nil {
				host = h
			} else {
				host = ""
			}
		}
		agent := task.Agent
		if agent == "" {
			if a, e := GetAgentID(); e == nil {
				agent = a
			}
		}
		_, err = tx.ExecContext(txCtx, `
			INSERT INTO tasks (
				id, name, content, long_description, status, priority, completed,
				created, last_modified, metadata, metadata_protobuf, metadata_format, parent_id, project_id, assigned_to, host, agent, version, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, strftime('%s', 'now'), strftime('%s', 'now'))
		`,
			task.ID,
			"", // name - TODO: add to Todo2Task struct if needed
			task.Content,
			task.LongDescription,
			task.Status,
			task.Priority,
			completedInt,
			now,              // created
			now,              // last_modified
			metadataJSON,     // JSON for backward compatibility
			metadataProtobuf, // Protobuf binary data (nil if serialization failed)
			metadataFormat,   // Format indicator
			parentID,
			projectID,
			assignedTo,
			host,
			agent,
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

// taskRow is a flat row scanner for v9 schema (parent_id, assigned_to, host, agent are NOT NULL).
type taskRow struct {
	ID              string         `db:"id"`
	Name            string         `db:"name"`
	Content         string         `db:"content"`
	LongDescription string         `db:"long_description"`
	Status          string         `db:"status"`
	Priority        string         `db:"priority"`
	Completed       int            `db:"completed"`
	Created         string         `db:"created"`
	LastModified    string         `db:"last_modified"`
	CompletedAt     string         `db:"completed_at"`
	Metadata        []byte         `db:"metadata"`
	MetadataProto   []byte         `db:"metadata_protobuf"`
	MetadataFormat  string         `db:"metadata_format"`
	ParentID        string         `db:"parent_id"`
	ProjectID       sql.NullString `db:"project_id"`
	AssignedTo      string         `db:"assigned_to"`
	Host            string         `db:"host"`
	Agent           string         `db:"agent"`
}

// GetTask retrieves a task by ID with all related data (tags, dependencies)
// Supports context for timeout and cancellation.
// Uses sqlx.GetContext for efficient row scanning.
func GetTask(ctx context.Context, id string) (*Todo2Task, error) {
	ctx = ensureContext(ctx)

	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var task *Todo2Task

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		var row taskRow
		err = db.GetContext(queryCtx, &row, `
			SELECT id, name, content, long_description, status, priority, completed,
			       created, last_modified, completed_at, metadata, metadata_protobuf, metadata_format,
			       parent_id, project_id, assigned_to, host, agent
			FROM tasks
			WHERE id = ?
		`, id)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("task %s not found", id)
		}
		if err != nil {
			return fmt.Errorf("failed to query task: %w", err)
		}

		taskData := Todo2Task{
			ID:              row.ID,
			Content:         row.Content,
			LongDescription: row.LongDescription,
			Status:          row.Status,
			Priority:        row.Priority,
			Completed:       row.Completed == 1,
			CreatedAt:       row.Created,
			LastModified:    row.LastModified,
			CompletedAt:     row.CompletedAt,
			ParentID:        row.ParentID,
			ProjectID:       row.ProjectID.String,
			AssignedTo:      row.AssignedTo,
			Host:            row.Host,
			Agent:           row.Agent,
		}

		taskData.NormalizeEpochDates()
		taskData.Metadata = DeserializeTaskMetadata(string(row.Metadata), row.MetadataProto, row.MetadataFormat)

		tags, err := loadTaskTags(ctx, queryCtx, db, id)
		if err != nil {
			return err
		}
		taskData.Tags = tags

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

// UpdateTask updates an existing task
// Uses a transaction to atomically update task, tags, and dependencies
// Supports context for timeout and cancellation.
func UpdateTask(ctx context.Context, task *Todo2Task) error {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	return retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
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

		metadataJSON, metadataProtobuf, metadataFormat, err := SerializeTaskMetadata(task)
		if err != nil {
			return err
		}

		// Get current version for optimistic locking
		var currentVersion int64

		err = tx.QueryRowContext(txCtx, `SELECT version FROM tasks WHERE id = ?`, task.ID).Scan(&currentVersion)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("task %s not found", task.ID)
		}

		if err != nil {
			return fmt.Errorf("failed to query task version: %w", err)
		}

		// Update task with optimistic locking (version check)
		// Include protobuf columns if they exist in schema; set completed_at when status is Done
		now := time.Now().Format(time.RFC3339)
		completedAtVal := task.CompletedAt

		if task.Status == "Done" && completedAtVal == "" {
			completedAtVal = now
		}

		// Update host/agent to current on modification (for distributed tracking)
		updateHost := task.Host
		if updateHost == "" {
			if h, e := os.Hostname(); e == nil {
				updateHost = h
			}
		}
		updateAgent := task.Agent
		if updateAgent == "" {
			if a, e := GetAgentID(); e == nil {
				updateAgent = a
			}
		}

		result, err := tx.ExecContext(txCtx, `
			UPDATE tasks SET
				content = ?,
				long_description = ?,
				status = ?,
				priority = ?,
				completed = ?,
				last_modified = ?,
				completed_at = ?,
				metadata = ?,
				metadata_protobuf = ?,
				metadata_format = ?,
				parent_id = ?,
				project_id = ?,
				assigned_to = ?,
				host = ?,
				agent = ?,
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
			completedAtVal,
			metadataJSON,
			metadataProtobuf,
			metadataFormat,
			task.ParentID,
			task.ProjectID,
			task.AssignedTo,
			updateHost,
			updateAgent,
			task.ID,
			currentVersion,
		)

		if err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("task %s not found or was modified by another agent: %w", task.ID, ErrVersionMismatch)
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

		// Invalidate caches linked to this task
		_ = ClearTaskTagSuggestions(task.ID)

		return nil
	})
}

// TaskMetadataUpdate represents a single task's metadata update.
type TaskMetadataUpdate struct {
	TaskID   string
	Metadata map[string]interface{}
}

// TaskStatusUpdate represents a single task's status/priority update.
type TaskStatusUpdate struct {
	TaskID   string
	Status   string
	Priority string
}

// BatchUpdateTaskStatus updates status and/or priority for multiple tasks in a single transaction.
// This is more efficient than individual UpdateTask calls when only status/priority changes are needed.
// Returns the number of tasks updated and any error.
func BatchUpdateTaskStatus(ctx context.Context, updates []TaskStatusUpdate) (int, error) {
	if len(updates) == 0 {
		return 0, nil
	}

	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	var updated int

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
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

		for _, u := range updates {
			setClauses := []string{"version = version + 1", "updated_at = strftime('%s', 'now')"}
			args := []interface{}{}

			if u.Status != "" {
				setClauses = append(setClauses, "status = ?")
				args = append(args, u.Status)
			}

			if u.Priority != "" {
				setClauses = append(setClauses, "priority = ?")
				args = append(args, u.Priority)
			}

			args = append(args, u.TaskID)

			query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ?", strings.Join(setClauses, ", "))
			result, err := tx.ExecContext(txCtx, query, args...)
			if err != nil {
				return fmt.Errorf("failed to update status for task %s: %w", u.TaskID, err)
			}

			rows, _ := result.RowsAffected()
			updated += int(rows)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Invalidate caches for all updated tasks
		for _, u := range updates {
			_ = ClearTaskTagSuggestions(u.TaskID)
		}

		return nil
	})

	return updated, err
}

// BatchUpdateTaskMetadata updates metadata for multiple tasks in a single transaction.
// This is more efficient than individual UpdateTask calls when only metadata changes are needed.
// Returns the number of tasks updated and any error.
func BatchUpdateTaskMetadata(ctx context.Context, updates []TaskMetadataUpdate) (int, error) {
	if len(updates) == 0 {
		return 0, nil
	}

	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	var updated int

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
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

		for _, u := range updates {
			metadataJSON, err := json.Marshal(u.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata for task %s: %w", u.TaskID, err)
			}

			result, err := tx.ExecContext(txCtx, `
				UPDATE tasks SET
					metadata = ?,
					version = version + 1,
					updated_at = strftime('%s', 'now')
				WHERE id = ?
			`, string(metadataJSON), u.TaskID)
			if err != nil {
				return fmt.Errorf("failed to update metadata for task %s: %w", u.TaskID, err)
			}

			rows, _ := result.RowsAffected()
			updated += int(rows)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Invalidate caches for all updated tasks
		for _, u := range updates {
			_ = ClearTaskTagSuggestions(u.TaskID)
		}

		return nil
	})

	return updated, err
}

// TaskFieldUpdate represents field updates for a single task.
type TaskFieldUpdate struct {
	TaskID      string
	Status      *string
	Priority    *string
	Name        *string
	Description *string
}

// UpdateTaskFields updates only the specified fields for a task in a single query.
// This is more efficient than UpdateTask when only simple fields need updating (no tags/deps changes).
// Returns error if task not found or update fails.
func UpdateTaskFields(ctx context.Context, update TaskFieldUpdate) error {
	if update.TaskID == "" {
		return fmt.Errorf("task ID is required")
	}

	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	return retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
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

		setClauses := []string{"version = version + 1", "updated_at = strftime('%s', 'now')"}
		args := []interface{}{}

		if update.Status != nil {
			setClauses = append(setClauses, "status = ?")
			args = append(args, *update.Status)
		}
		if update.Priority != nil {
			setClauses = append(setClauses, "priority = ?")
			args = append(args, *update.Priority)
		}
		if update.Name != nil {
			setClauses = append(setClauses, "content = ?")
			args = append(args, *update.Name)
		}
		if update.Description != nil {
			setClauses = append(setClauses, "long_description = ?")
			args = append(args, *update.Description)
		}

		args = append(args, update.TaskID)

		query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ?", strings.Join(setClauses, ", "))
		result, err := tx.ExecContext(txCtx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update task fields: %w", err)
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("task %s not found", update.TaskID)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Invalidate caches linked to this task
		_ = ClearTaskTagSuggestions(update.TaskID)

		return nil
	})
}

// IsVersionMismatchError reports whether err is a task version mismatch (concurrent update).
func IsVersionMismatchError(err error) bool {
	return errors.Is(err, ErrVersionMismatch)
}

// CheckUpdateConflict returns whether the task's current version differs from expectedVersion.
// Used to detect conflicts before or after an update attempt. If the task is not found, err is non-nil.
func CheckUpdateConflict(ctx context.Context, taskID string, expectedVersion int64) (hasConflict bool, currentVersion int64, err error) {
	ctx = ensureContext(ctx)

	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	db, err := GetDBx()
	if err != nil {
		return false, 0, fmt.Errorf("failed to get database: %w", err)
	}

	err = db.GetContext(queryCtx, &currentVersion, `SELECT version FROM tasks WHERE id = ?`, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, fmt.Errorf("task %s not found", taskID)
	}

	if err != nil {
		return false, 0, fmt.Errorf("failed to get task version: %w", err)
	}

	return currentVersion != expectedVersion, currentVersion, nil
}

// DeleteTask deletes a task and all related data (tags, dependencies cascade)
// Supports context for timeout and cancellation.
func DeleteTask(ctx context.Context, id string) error {
	ctx = ensureContext(ctx)

	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	return retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
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

		// Invalidate caches linked to this task
		_ = ClearTaskTagSuggestions(id)
		_ = ClearFileTaskTags(id)

		// Tags and dependencies are cascade deleted by foreign key constraints
		return nil
	})
}

// BatchDeleteTasks deletes multiple tasks in a single transaction and returns the IDs that were deleted.
// Silently skips IDs not found. Tags and dependencies are cascade-deleted by FK constraints.
func BatchDeleteTasks(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	ctx = ensureContext(ctx)

	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var deleted []string

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		query, args, err := sqlx.In(`DELETE FROM tasks WHERE id IN (?)`, ids)
		if err != nil {
			return fmt.Errorf("failed to build batch delete query: %w", err)
		}

		result, err := db.ExecContext(queryCtx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to batch delete tasks: %w", err)
		}

		n, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		// Report all requested IDs as deleted if the count matches; otherwise fall back to all IDs.
		// (SQLite does not return which rows were deleted from a bulk DELETE.)
		if int(n) == len(ids) {
			deleted = ids
		} else {
			deleted = ids[:n] // best-effort: n <= len(ids)
		}

		// Invalidate caches for each deleted task
		for _, id := range deleted {
			_ = ClearTaskTagSuggestions(id)
			_ = ClearFileTaskTags(id)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return deleted, nil
}
