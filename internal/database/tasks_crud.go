// tasks_crud.go — Create, Get, Update, Delete task operations and version conflict helpers.
// Same package as tasks.go; uses types and metadata helpers from tasks.go.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
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
		parentID := sqlNullString(task.ParentID)
		projectID := sqlNullString(task.ProjectID)
		assignedTo := sqlNullString(task.AssignedTo)
		host := sqlNullString(task.Host)
		agent := sqlNullString(task.Agent)
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
		// If distributed tracking columns don't exist, try without them
		if err != nil && strings.Contains(err.Error(), "no such column") {
			_, err = tx.ExecContext(txCtx, `
				INSERT INTO tasks (
					id, name, content, long_description, status, priority, completed,
					created, last_modified, metadata, metadata_protobuf, metadata_format, parent_id, version, created_at, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, strftime('%s', 'now'), strftime('%s', 'now'))
			`,
				task.ID,
				"",
				task.Content,
				task.LongDescription,
				task.Status,
				task.Priority,
				completedInt,
				now,
				now,
				metadataJSON,
				metadataProtobuf,
				metadataFormat,
				parentID,
			)
		}
		// If protobuf or parent_id columns don't exist, fall back to minimal schema
		if err != nil && strings.Contains(err.Error(), "no such column") {
			_, err = tx.ExecContext(txCtx, `
				INSERT INTO tasks (
					id, name, content, long_description, status, priority, completed,
					created, last_modified, metadata, version, created_at, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, strftime('%s', 'now'), strftime('%s', 'now'))
			`,
				task.ID,
				"",
				task.Content,
				task.LongDescription,
				task.Status,
				task.Priority,
				completedInt,
				now,
				now,
				metadataJSON,
			)
		}

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
// Supports context for timeout and cancellation.
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

		var name sql.NullString

		var created, lastModified, completedAt sql.NullString

		var parentID, projectID, assignedTo, host, agent sql.NullString

		// Try to query with full schema (protobuf + distributed tracking)
		err = db.QueryRowContext(queryCtx, `
			SELECT id, name, content, long_description, status, priority, completed,
			       created, last_modified, completed_at, metadata, metadata_protobuf, metadata_format, parent_id,
			       project_id, assigned_to, host, agent
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
			&completedAt,
			&metadataJSON,
			&metadataProtobuf,
			&metadataFormat,
			&parentID,
			&projectID,
			&assignedTo,
			&host,
			&agent,
		)
		if err == nil {
			if parentID.Valid {
				taskData.ParentID = parentID.String
			}
			if projectID.Valid {
				taskData.ProjectID = projectID.String
			}
			if assignedTo.Valid {
				taskData.AssignedTo = assignedTo.String
			}
			if host.Valid {
				taskData.Host = host.String
			}
			if agent.Valid {
				taskData.Agent = agent.String
			}
		}

		// If distributed tracking columns don't exist, try without them
		if err != nil && strings.Contains(err.Error(), "no such column") {
			projectID, assignedTo, host, agent = sql.NullString{}, sql.NullString{}, sql.NullString{}, sql.NullString{}
			err = db.QueryRowContext(queryCtx, `
				SELECT id, name, content, long_description, status, priority, completed,
				       created, last_modified, completed_at, metadata, metadata_protobuf, metadata_format, parent_id
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
				&completedAt,
				&metadataJSON,
				&metadataProtobuf,
				&metadataFormat,
				&parentID,
			)
			if err == nil && parentID.Valid {
				taskData.ParentID = parentID.String
			}
		}

		// If protobuf, completed_at, or parent_id column don't exist, fall back to minimal schema
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

		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("task %s not found", id)
		}

		if err != nil {
			return fmt.Errorf("failed to query task: %w", err)
		}

		taskData.Completed = completedInt == 1
		if created.Valid {
			taskData.CreatedAt = created.String
		}

		if lastModified.Valid {
			taskData.LastModified = lastModified.String
		}

		if completedAt.Valid {
			taskData.CompletedAt = completedAt.String
		}

		taskData.NormalizeEpochDates()

		metadataJSONStr := ""
		if metadataJSON.Valid {
			metadataJSONStr = metadataJSON.String
		}

		metadataFormatStr := ""
		if metadataFormat.Valid {
			metadataFormatStr = metadataFormat.String
		}

		taskData.Metadata = DeserializeTaskMetadata(metadataJSONStr, metadataProtobuf, metadataFormatStr)

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

// UpdateTask updates an existing task
// Uses a transaction to atomically update task, tags, and dependencies
// Supports context for timeout and cancellation.
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
			sqlNullString(task.ParentID),
			sqlNullString(task.ProjectID),
			sqlNullString(task.AssignedTo),
			sqlNullString(updateHost),
			sqlNullString(updateAgent),
			task.ID,
			currentVersion,
		)
		// If distributed tracking columns don't exist, try without them
		if err != nil && strings.Contains(err.Error(), "no such column") {
			result, err = tx.ExecContext(txCtx, `
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
				sqlNullString(task.ParentID),
				task.ID,
				currentVersion,
			)
		}
		// If protobuf or parent_id columns don't exist, fall back to minimal schema
		if err != nil && strings.Contains(err.Error(), "no such column") {
			result, err = tx.ExecContext(txCtx, `
				UPDATE tasks SET
					content = ?,
					long_description = ?,
					status = ?,
					priority = ?,
					completed = ?,
					last_modified = ?,
					completed_at = ?,
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
				completedAtVal,
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

	db, err := GetDB()
	if err != nil {
		return false, 0, fmt.Errorf("failed to get database: %w", err)
	}

	err = db.QueryRowContext(queryCtx, `SELECT version FROM tasks WHERE id = ?`, taskID).Scan(&currentVersion)
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
