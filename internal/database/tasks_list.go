// tasks_list.go — ListTasks, GetDoneTasksForEstimation, GetTasksByStatus/ByTag/ByPriority.
// Same package as tasks.go; uses TaskFilters, Todo2Task, unmarshalTaskMetadata, loadTaskTags, loadTaskDependencies.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
)

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
			SELECT DISTINCT t.id, t.content, t.long_description, t.status, t.priority, t.completed, t.created, t.last_modified, t.completed_at, t.metadata, t.metadata_protobuf, t.metadata_format, t.parent_id, t.project_id, t.assigned_to, t.host, t.agent
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
			if filters.AssignedTo != nil {
				conditions = append(conditions, "t.assigned_to = ?")
				args = append(args, *filters.AssignedTo)
			}
			if filters.Host != nil {
				conditions = append(conditions, "t.host = ?")
				args = append(args, *filters.Host)
			}
			if filters.Agent != nil {
				conditions = append(conditions, "t.agent = ?")
				args = append(args, *filters.Agent)
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

		// Try to query with full schema (protobuf + distributed tracking) first
		rows, err := db.QueryContext(queryCtx, query, args...)
		hasProtobufColumns := true
		hasDistributedColumns := true

		if err != nil && strings.Contains(err.Error(), "no such column") {
			hasDistributedColumns = false
			// Distributed tracking columns don't exist, try without them
			queryBuilderMid := strings.Builder{}
			queryBuilderMid.WriteString(`
				SELECT DISTINCT t.id, t.content, t.long_description, t.status, t.priority, t.completed, t.created, t.last_modified, t.completed_at, t.metadata, t.metadata_protobuf, t.metadata_format, t.parent_id
				FROM tasks t
			`)
			if len(conditions) > 0 {
				queryBuilderMid.WriteString(" WHERE " + conditions[0])
				for i := 1; i < len(conditions); i++ {
					queryBuilderMid.WriteString(" AND " + conditions[i])
				}
			}
			queryBuilderMid.WriteString(" ORDER BY t.created_at DESC")
			rows, err = db.QueryContext(queryCtx, queryBuilderMid.String(), args...)
		}
		if err != nil && strings.Contains(err.Error(), "no such column") {
			// Protobuf or date columns don't exist, use minimal schema
			hasProtobufColumns = false
			hasDistributedColumns = false
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

			var created, lastMod, completedAt, parentID, projectID, assignedTo, host, agent sql.NullString

			// Scan based on schema level (full, protobuf-only, or minimal)
			var scanErr error
			if hasDistributedColumns {
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
					&projectID,
					&assignedTo,
					&host,
					&agent,
				)
				if scanErr == nil {
					if parentID.Valid {
						task.ParentID = parentID.String
					}
					if projectID.Valid {
						task.ProjectID = projectID.String
					}
					if assignedTo.Valid {
						task.AssignedTo = assignedTo.String
					}
					if host.Valid {
						task.Host = host.String
					}
					if agent.Valid {
						task.Agent = agent.String
					}
				}
			} else if hasProtobufColumns {
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
