package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/logging"
	"github.com/davidl71/exarp-go/internal/models"
)

// ErrVersionMismatch is returned when an update fails because the task was modified by another agent.
var ErrVersionMismatch = errors.New("task version mismatch")

// Todo2Task is an alias for models.Todo2Task (for convenience).
type Todo2Task = models.Todo2Task

// TaskFilters represents filters for querying tasks.
type TaskFilters struct {
	Status    *string
	Priority  *string
	Tag       *string
	ProjectID *string
}

// SanitizeTaskMetadata parses JSON metadata; on failure returns a map with "raw" key and logs.
// Callers never see invalid character parse errors—invalid JSON is coerced to {"raw": "..."}.
// Use when loading tasks from DB or JSON so code that expects metadata as a map never fails.
func SanitizeTaskMetadata(s string) map[string]interface{} {
	if s == "" {
		return nil
	}

	var out map[string]interface{}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		logging.Warn("database: invalid task metadata JSON, coercing to raw: %v", err)
		return map[string]interface{}{"raw": s}
	}

	return out
}

// unmarshalTaskMetadata is an alias for SanitizeTaskMetadata (internal use).
func unmarshalTaskMetadata(s string) map[string]interface{} {
	return SanitizeTaskMetadata(s)
}

// SerializeTaskMetadata serializes task metadata for DB storage: JSON string, optional protobuf bytes, and format.
// Returns (metadataJSON, metadataProtobuf, metadataFormat, nil) or error if JSON marshal fails.
func SerializeTaskMetadata(task *Todo2Task) (metadataJSON string, metadataProtobuf []byte, metadataFormat string, err error) {
	metadataFormat = "protobuf"
	protobufData, err := models.SerializeTaskToProtobuf(task)

	if err != nil {
		metadataFormat = "json"
	} else {
		metadataProtobuf = protobufData
	}

	if len(task.Metadata) > 0 {
		sanitized := SanitizeMetadataForWrite(task.Metadata)

		metadataBytes, jsonErr := json.Marshal(sanitized)
		if jsonErr != nil {
			return "", nil, "", fmt.Errorf("failed to marshal metadata: %w", jsonErr)
		}

		metadataJSON = string(metadataBytes)
	}

	return metadataJSON, metadataProtobuf, metadataFormat, nil
}

// DeserializeTaskMetadata deserializes metadata from DB: prefers protobuf when format is "protobuf", else JSON.
// metadataJSON and metadataFormat may be empty (e.g. from old schema); invalid JSON is coerced via SanitizeTaskMetadata.
func DeserializeTaskMetadata(metadataJSON string, metadataProtobuf []byte, metadataFormat string) map[string]interface{} {
	if metadataFormat == "protobuf" && len(metadataProtobuf) > 0 {
		deserialized, err := models.DeserializeTaskFromProtobuf(metadataProtobuf)
		if err == nil {
			return deserialized.Metadata
		}
	}

	if metadataJSON != "" {
		return SanitizeTaskMetadata(metadataJSON)
	}

	return nil
}

// SanitizeMetadataForWrite returns a copy of metadata with all values coerced to types
// that encoding/json can marshal (string, float64, int, int64, bool, nil, []interface{},
// map[string]interface{}). Use when writing DB or state JSON so non-scalar values never break marshaling.
func SanitizeMetadataForWrite(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return nil
	}

	out := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		out[k] = jsonSafeValue(v)
	}

	return out
}

func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{String: s, Valid: true}
}

func jsonSafeValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch x := v.(type) {
	case string:
		return x
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case bool:
		return x
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, e := range x {
			out[i] = jsonSafeValue(e)
		}

		return out
	case map[string]interface{}:
		out := make(map[string]interface{}, len(x))
		for k2, val := range x {
			out[k2] = jsonSafeValue(val)
		}

		return out
	default:
		return fmt.Sprint(x)
	}
}

// IsValidTaskID returns true if id is a valid task ID (T-<digits>).
// Rejects empty, "T-NaN", "T-undefined", and any non-numeric suffix.
func IsValidTaskID(id string) bool {
	if id == "" || id == "T-NaN" || id == "T-undefined" {
		return false
	}

	if !strings.HasPrefix(id, "T-") || len(id) <= 2 {
		return false
	}

	for _, c := range id[2:] {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// GenerateTaskID returns a new task ID in the form T-<epoch_milliseconds>.
func GenerateTaskID() string {
	return fmt.Sprintf("T-%d", time.Now().UnixMilli())
}

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

		// Insert task with protobuf data (if available) and JSON (for compatibility)
		now := time.Now().Format(time.RFC3339)
		parentID := sqlNullString(task.ParentID)
		_, err = tx.ExecContext(txCtx, `
			INSERT INTO tasks (
				id, name, content, long_description, status, priority, completed,
				created, last_modified, metadata, metadata_protobuf, metadata_format, parent_id, version, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, strftime('%s', 'now'), strftime('%s', 'now'))
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
		)
		// If protobuf or parent_id columns don't exist, fall back to old schema
		if err != nil && strings.Contains(err.Error(), "no such column") {
			_, err = tx.ExecContext(txCtx, `
				INSERT INTO tasks (
					id, name, content, long_description, status, priority, completed,
					created, last_modified, metadata, version, created_at, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, strftime('%s', 'now'), strftime('%s', 'now'))
			`,
				task.ID,
				"", // name - TODO: add to Todo2Task struct if needed
				task.Content,
				task.LongDescription,
				task.Status,
				task.Priority,
				completedInt,
				now,          // created
				now,          // last_modified
				metadataJSON, // JSON only
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

		var parentID sql.NullString

		// Try to query with protobuf columns and parent_id first (new schema)
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

		// If protobuf, completed_at, or parent_id column don't exist, fall back to old schema
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

// loadTaskTags loads tags for a single task
// Works with both *sql.DB (via QueryContext) and *sql.Tx (via QueryContext).
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
// Works with both *sql.DB (via QueryContext) and *sql.Tx (via QueryContext).
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
			metadataJSON,     // JSON for backward compatibility
			metadataProtobuf, // Protobuf binary data
			metadataFormat,   // Format indicator
			sqlNullString(task.ParentID),
			task.ID,
			currentVersion,
		)
		// If protobuf, completed_at, or parent_id columns don't exist, fall back to old schema
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

// ListTasks retrieves tasks with optional filtering
// Supports context for timeout and cancellation.
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
		// Include protobuf columns and date columns if they exist (for new schema)
		var queryBuilder strings.Builder

		queryBuilder.WriteString(`
			SELECT DISTINCT t.id, t.content, t.long_description, t.status, t.priority, t.completed, t.created, t.last_modified, t.completed_at, t.metadata, t.metadata_protobuf, t.metadata_format, t.parent_id
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
			// Protobuf or date columns don't exist, use old schema
			hasProtobufColumns = false
			queryBuilderOld := strings.Builder{}
			queryBuilderOld.WriteString(`
				SELECT DISTINCT t.id, t.content, t.long_description, t.status, t.priority, t.completed, t.created, t.last_modified, t.completed_at, t.metadata
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

			var created, lastMod, completedAt, parentID sql.NullString

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
					&created,
					&lastMod,
					&completedAt,
					&metadataJSON,
					&metadataProtobuf,
					&metadataFormat,
					&parentID,
				)
				if scanErr == nil && parentID.Valid {
					task.ParentID = parentID.String
				}
			} else {
				scanErr = rows.Scan(
					&task.ID,
					&task.Content,
					&task.LongDescription,
					&task.Status,
					&task.Priority,
					&completedInt,
					&created,
					&lastMod,
					&completedAt,
					&metadataJSON,
				)
			}

			if scanErr != nil {
				return fmt.Errorf("failed to scan task: %w", scanErr)
			}

			task.Completed = completedInt == 1
			if created.Valid {
				task.CreatedAt = created.String
			}

			if lastMod.Valid {
				task.LastModified = lastMod.String
			}

			if completedAt.Valid {
				task.CompletedAt = completedAt.String
			}

			task.NormalizeEpochDates()

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
						task.Metadata = unmarshalTaskMetadata(metadataJSON.String)
					}
				}
			} else if metadataJSON.Valid && metadataJSON.String != "" {
				// Use JSON format (legacy or fallback)
				task.Metadata = unmarshalTaskMetadata(metadataJSON.String)
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

// TaskForEstimation holds Done task fields needed for estimation/historical analysis.
// Used by the estimation tool to load completed tasks from DB without full Todo2Task.
type TaskForEstimation struct {
	ID              string
	Content         string
	LongDescription string
	Status          string
	Priority        string
	Created         string
	LastModified    string
	CompletedAt     string
	EstimatedHours  float64
	ActualHours     float64
	Tags            []string
}

// GetDoneTasksForEstimation returns Done tasks with estimation-relevant columns
// (created, last_modified, completed_at, estimated_hours, actual_hours).
// Used by estimation tool for DB-first historical loading; falls back to JSON in tools layer.
func GetDoneTasksForEstimation(ctx context.Context) ([]*TaskForEstimation, error) {
	ctx = ensureContext(ctx)

	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var result []*TaskForEstimation

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		// Schema 001 has created, last_modified, completed_at, estimated_hours, actual_hours
		rows, err := db.QueryContext(queryCtx, `
			SELECT id, content, long_description, status, priority,
			       created, last_modified, completed_at, estimated_hours, actual_hours
			FROM tasks
			WHERE status = ?
			ORDER BY created_at DESC
		`, StatusDone)
		if err != nil {
			return fmt.Errorf("failed to query Done tasks: %w", err)
		}
		defer rows.Close()

		var list []*TaskForEstimation

		var taskIDs []string

		taskMap := make(map[string]*TaskForEstimation)

		for rows.Next() {
			var t TaskForEstimation

			var created, lastMod, completedAt sql.NullString

			var estHours, actHours sql.NullFloat64

			if err := rows.Scan(
				&t.ID,
				&t.Content,
				&t.LongDescription,
				&t.Status,
				&t.Priority,
				&created,
				&lastMod,
				&completedAt,
				&estHours,
				&actHours,
			); err != nil {
				return fmt.Errorf("failed to scan task: %w", err)
			}

			if created.Valid {
				t.Created = created.String
			}

			if lastMod.Valid {
				t.LastModified = lastMod.String
			}

			if completedAt.Valid {
				t.CompletedAt = completedAt.String
			}

			if estHours.Valid {
				t.EstimatedHours = estHours.Float64
			}

			if actHours.Valid {
				t.ActualHours = actHours.Float64
			}

			list = append(list, &t)
			taskIDs = append(taskIDs, t.ID)
			taskMap[t.ID] = &t
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating rows: %w", err)
		}

		// Batch load tags
		if len(taskIDs) > 0 {
			placeholders := make([]string, len(taskIDs))
			args := make([]interface{}, len(taskIDs))

			for i, id := range taskIDs {
				placeholders[i] = "?"
				args[i] = id
			}

			tagRows, err := db.QueryContext(queryCtx, `
				SELECT task_id, tag FROM task_tags
				WHERE task_id IN (`+strings.Join(placeholders, ", ")+`)
				ORDER BY task_id, tag
			`, args...)
			if err != nil {
				return fmt.Errorf("failed to batch query tags: %w", err)
			}

			defer tagRows.Close()

			for tagRows.Next() {
				var taskID, tag string
				if err := tagRows.Scan(&taskID, &tag); err != nil {
					return fmt.Errorf("failed to scan tag: %w", err)
				}

				if t, ok := taskMap[taskID]; ok {
					t.Tags = append(t.Tags, tag)
				}
			}

			if err = tagRows.Err(); err != nil {
				return fmt.Errorf("error iterating tag rows: %w", err)
			}
		}

		result = list

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetTasksByStatus retrieves all tasks with the specified status.
func GetTasksByStatus(ctx context.Context, status string) ([]*Todo2Task, error) {
	filters := &TaskFilters{Status: &status}
	return ListTasks(ctx, filters)
}

// GetTasksByTag retrieves all tasks with the specified tag.
func GetTasksByTag(ctx context.Context, tag string) ([]*Todo2Task, error) {
	filters := &TaskFilters{Tag: &tag}
	return ListTasks(ctx, filters)
}

// GetTasksByPriority retrieves all tasks with the specified priority.
func GetTasksByPriority(ctx context.Context, priority string) ([]*Todo2Task, error) {
	filters := &TaskFilters{Priority: &priority}
	return ListTasks(ctx, filters)
}

// FixTaskDates backfills created and last_modified from created_at/updated_at (Unix epoch)
// for rows where created or last_modified is empty or 1970-01-01. Returns the number of rows updated.
func FixTaskDates(ctx context.Context) (int64, error) {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	var rowsAffected int64

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}
		// Backfill created and last_modified from integer created_at/updated_at
		res, err := db.ExecContext(txCtx, `
			UPDATE tasks SET
				created = strftime('%Y-%m-%dT%H:%M:%SZ', datetime(created_at, 'unixepoch')),
				last_modified = strftime('%Y-%m-%dT%H:%M:%SZ', datetime(updated_at, 'unixepoch'))
			WHERE created = '' OR created LIKE '1970%'
			   OR last_modified IS NULL OR last_modified = '' OR last_modified LIKE '1970%'
		`)
		if err != nil {
			return fmt.Errorf("failed to fix task dates: %w", err)
		}

		rowsAffected, err = res.RowsAffected()
		if err != nil {
			return err
		}
		// For Done tasks, backfill completed_at from updated_at if missing or epoch
		_, err = db.ExecContext(txCtx, `
			UPDATE tasks SET
				completed_at = strftime('%Y-%m-%dT%H:%M:%SZ', datetime(updated_at, 'unixepoch'))
			WHERE status = 'Done' AND (completed_at IS NULL OR completed_at = '' OR completed_at LIKE '1970%')
		`)
		if err != nil {
			// completed_at column might not exist in older schema
			if !strings.Contains(err.Error(), "no such column") {
				return fmt.Errorf("failed to fix completed_at: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// GetDependencies retrieves all task IDs that the specified task depends on.
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

// GetDependents retrieves all task IDs that depend on the specified task.
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

// GetTagsForTask is a helper function to retrieve tags for a task.
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
