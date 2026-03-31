// tasks_list.go — ListTasks, GetDoneTasksForEstimation, GetTasksByStatus/ByTag/ByPriority.
// Same package as tasks.go; uses TaskFilters, Todo2Task, unmarshalTaskMetadata, loadTaskTags, loadTaskDependencies.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
	"github.com/jmoiron/sqlx"
)

// ListTasks retrieves tasks with optional filtering
// Supports context for timeout and cancellation.
func ListTasks(ctx context.Context, filters *TaskFilters) ([]*Todo2Task, error) {
	ctx = ensureContext(ctx)

	var tasks []*Todo2Task

	err := retryWithBackoff(ctx, func() error {
		queryCtx, cancel, db, err := QueryContextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()

		var queryBuilder strings.Builder

		queryBuilder.WriteString(`
			SELECT DISTINCT
			       t.id, t.name, t.content, t.long_description,
			       t.status, t.status_enum,
			       t.priority, t.priority_enum,
			       t.completed,
			       t.created, t.last_modified, t.completed_at,
			       t.created_ts, t.last_modified_ts, t.completed_at_ts,
			       t.metadata, t.metadata_protobuf, t.metadata_format,
			       t.parent_id, t.project_id, t.assigned_to, t.host, t.agent, t.version
			FROM tasks t
		`)

		var args []interface{}
		var conditions []string

		if filters != nil {
			if filters.Status != nil {
				conditions = append(conditions, "t.status = ?")
				args = append(args, *filters.Status)
			}

			if filters.StatusEnum != nil && filters.Status == nil {
				if title := (*filters.StatusEnum).TitleString(); title != "" {
					conditions = append(conditions, "t.status_enum = ?")
					args = append(args, taskStatusEnumInt(title))
				}
			}

			if len(filters.Statuses) > 0 {
				placeholders := make([]string, len(filters.Statuses))
				for i, s := range filters.Statuses {
					placeholders[i] = "?"
					args = append(args, s)
				}
				conditions = append(conditions, fmt.Sprintf("t.status IN (%s)", strings.Join(placeholders, ",")))
			}

			if len(filters.StatusEnums) > 0 && len(filters.Statuses) == 0 {
				var ints []int
				for _, s := range filters.StatusEnums {
					if title := s.TitleString(); title != "" {
						ints = append(ints, taskStatusEnumInt(title))
					}
				}
				if len(ints) > 0 {
					placeholders := make([]string, len(ints))
					for i, v := range ints {
						placeholders[i] = "?"
						args = append(args, v)
					}
					conditions = append(conditions, fmt.Sprintf("t.status_enum IN (%s)", strings.Join(placeholders, ",")))
				}
			}

			if filters.Priority != nil {
				conditions = append(conditions, "t.priority = ?")
				args = append(args, *filters.Priority)
			}

			if filters.PriorityEnum != nil && filters.Priority == nil {
				// priority can be empty meaning "unspecified".
				canon := (*filters.PriorityEnum).CanonicalString()
				conditions = append(conditions, "t.priority_enum = ?")
				args = append(args, taskPriorityEnumInt(canon))
			}

			if filters.Tag != nil {
				queryBuilder.WriteString(` INNER JOIN task_tags tt ON t.id = tt.task_id `)

				conditions = append(conditions, "tt.tag = ?")
				args = append(args, *filters.Tag)
			}

			if filters.ProjectID != nil {
				if filters.IncludeNullProjectID {
					conditions = append(conditions, "(t.project_id = ? OR t.project_id IS NULL OR t.project_id = '')")
				} else {
					conditions = append(conditions, "t.project_id = ?")
				}
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

		queryBuilder.WriteString(" ORDER BY t.created_ts DESC")
		query := queryBuilder.String()

		var rows []taskRow
		if err := db.SelectContext(queryCtx, &rows, query, args...); err != nil {
			return fmt.Errorf("failed to query tasks: %w", err)
		}

		var taskList []*Todo2Task
		var taskIDs []string
		taskMap := make(map[string]*Todo2Task, len(rows))

		for _, row := range rows {
			task := Todo2Task{
				ID:              row.ID,
				Name:            row.Name,
				Content:         row.Content,
				LongDescription: row.LongDescription,
				Status:          row.Status,
				StatusEnum:      taskStatusFromEnumInt(row.StatusEnumInt),
				Priority:        row.Priority,
				PriorityEnum:    taskPriorityFromEnumInt(row.PriorityEnumInt),
				Completed:       row.Completed == 1,
				CreatedAt:       row.Created,
				LastModified:    row.LastModified,
				CompletedAt:     row.CompletedAt,
				ParentID:        row.ParentID,
				AssignedTo:      row.AssignedTo,
				Host:            row.Host,
				Agent:           row.Agent,
				Version:         row.Version,
			}

			if row.ProjectID.Valid {
				task.ProjectID = row.ProjectID.String
			}

			includeMetadata := true
			if filters != nil && filters.IncludeMetadata != nil {
				includeMetadata = *filters.IncludeMetadata
			}
			if includeMetadata {
				if row.MetadataFormat == "protobuf" && len(row.MetadataProto) > 0 {
					deserializedTask, err := models.DeserializeTaskFromProtobuf(row.MetadataProto)
					if err == nil {
						task.Metadata = deserializedTask.Metadata
					} else if len(row.Metadata) > 0 {
						task.Metadata = unmarshalTaskMetadata(string(row.Metadata))
					}
				} else if len(row.Metadata) > 0 {
					task.Metadata = unmarshalTaskMetadata(string(row.Metadata))
				}
			}

			task.NormalizeEpochDates()
			task.EnsureName()

			rowCopy := task
			taskIDs = append(taskIDs, row.ID)
			taskMap[row.ID] = &rowCopy
			taskList = append(taskList, &rowCopy)
		}

		if len(taskIDs) > 0 {
			tagQuery, tagArgs, err := sqlx.In(`
				SELECT task_id, tag FROM task_tags
				WHERE task_id IN (?)
				ORDER BY task_id, tag
			`, taskIDs)
			if err != nil {
				return fmt.Errorf("failed to build tag query: %w", err)
			}

			var tagResults []struct {
				TaskID string `db:"task_id"`
				Tag    string `db:"tag"`
			}
			if err := db.SelectContext(queryCtx, &tagResults, tagQuery, tagArgs...); err != nil {
				return fmt.Errorf("failed to batch query tags: %w", err)
			}

			for _, tr := range tagResults {
				if task, ok := taskMap[tr.TaskID]; ok {
					task.Tags = append(task.Tags, tr.Tag)
				}
			}

			depQuery, depArgs, err := sqlx.In(`
				SELECT task_id, depends_on_id FROM task_dependencies
				WHERE task_id IN (?)
				ORDER BY task_id, depends_on_id
			`, taskIDs)
			if err != nil {
				return fmt.Errorf("failed to build dependency query: %w", err)
			}

			var depResults []struct {
				TaskID      string `db:"task_id"`
				DependsOnID string `db:"depends_on_id"`
			}
			if err := db.SelectContext(queryCtx, &depResults, depQuery, depArgs...); err != nil {
				return fmt.Errorf("failed to batch query dependencies: %w", err)
			}

			for _, dr := range depResults {
				if task, ok := taskMap[dr.TaskID]; ok {
					task.Dependencies = append(task.Dependencies, dr.DependsOnID)
				}
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

type taskForEstimationRow struct {
	ID              string          `db:"id"`
	Content         string          `db:"content"`
	LongDescription string          `db:"long_description"`
	Status          string          `db:"status"`
	Priority        string          `db:"priority"`
	Created         sql.NullString  `db:"created"`
	LastModified    sql.NullString  `db:"last_modified"`
	CompletedAt     sql.NullString  `db:"completed_at"`
	EstimatedHours  sql.NullFloat64 `db:"estimated_hours"`
	ActualHours     sql.NullFloat64 `db:"actual_hours"`
}

// GetDoneTasksForEstimation returns Done tasks with estimation-relevant columns
// (created, last_modified, completed_at, estimated_hours, actual_hours).
// Used by estimation tool for DB-first historical loading; falls back to JSON in tools layer.
func GetDoneTasksForEstimation(ctx context.Context) ([]*TaskForEstimation, error) {
	ctx = ensureContext(ctx)

	var result []*TaskForEstimation

	err := retryWithBackoff(ctx, func() error {
		queryCtx, cancel, db, err := QueryContextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()

		var rows []taskForEstimationRow
		if err := db.SelectContext(queryCtx, &rows, `
			SELECT id, content, long_description, status, priority,
			       created, last_modified, completed_at, estimated_hours, actual_hours
			FROM tasks
			WHERE status = ?
			ORDER BY created_at DESC
		`, StatusDone); err != nil {
			return fmt.Errorf("failed to query Done tasks: %w", err)
		}

		var list []*TaskForEstimation
		var taskIDs []string
		taskMap := make(map[string]*TaskForEstimation)

		for _, row := range rows {
			t := &TaskForEstimation{
				ID:              row.ID,
				Content:         row.Content,
				LongDescription: row.LongDescription,
				Status:          row.Status,
				Priority:        row.Priority,
			}

			if row.Created.Valid {
				t.Created = row.Created.String
			}
			if row.LastModified.Valid {
				t.LastModified = row.LastModified.String
			}
			if row.CompletedAt.Valid {
				t.CompletedAt = row.CompletedAt.String
			}
			if row.EstimatedHours.Valid {
				t.EstimatedHours = row.EstimatedHours.Float64
			}
			if row.ActualHours.Valid {
				t.ActualHours = row.ActualHours.Float64
			}

			list = append(list, t)
			taskIDs = append(taskIDs, t.ID)
			taskMap[t.ID] = t
		}

		// Batch load tags using sqlx.In + SelectContext
		if len(taskIDs) > 0 {
			tagQuery, tagArgs, err := sqlx.In(`
				SELECT task_id, tag FROM task_tags
				WHERE task_id IN (?)
				ORDER BY task_id, tag
			`, taskIDs)
			if err != nil {
				return fmt.Errorf("failed to build tag query: %w", err)
			}

			var tagResults []struct {
				TaskID string `db:"task_id"`
				Tag    string `db:"tag"`
			}
			if err := db.SelectContext(queryCtx, &tagResults, tagQuery, tagArgs...); err != nil {
				return fmt.Errorf("failed to batch query tags: %w", err)
			}

			for _, tr := range tagResults {
				if t, ok := taskMap[tr.TaskID]; ok {
					t.Tags = append(t.Tags, tr.Tag)
				}
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

// GetTaskCountByStatus returns the count of tasks with the specified status.
func GetTaskCountByStatus(ctx context.Context, status string) (int, error) {
	ctx = ensureContext(ctx)

	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var count int

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		err = db.GetContext(queryCtx, &count, `SELECT COUNT(*) FROM tasks WHERE status = ?`, status)
		if err != nil {
			return fmt.Errorf("failed to count tasks: %w", err)
		}
		return nil
	})

	return count, err
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

// FindNextClaimableTask returns the first unassigned Todo task ordered by priority
// (high → critical → medium). Returns nil if none found.
// This is more efficient than calling GetTasksByPriority multiple times.
func FindNextClaimableTask(ctx context.Context) (*Todo2Task, error) {
	ctx = ensureContext(ctx)

	var task *Todo2Task

	err := retryWithBackoff(ctx, func() error {
		queryCtx, cancel, db, err := QueryContextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()

		// Exclude tasks currently locked by an agent (assignee set + lock not yet expired).
		// assigned_to is persistent ownership; assignee+lock_until is the agent lock column.
		now := time.Now().Unix()
		var row taskRowWithAgg
		err = db.GetContext(queryCtx, &row, `
			SELECT t.id, t.content, t.long_description, t.status, t.priority, t.completed, t.created, t.last_modified,
			       t.completed_at, t.metadata, t.metadata_protobuf, t.metadata_format, t.parent_id, t.project_id,
			       t.assigned_to, t.host, t.agent, t.version`+sqlTaskAggJSON+`
			FROM tasks AS t
			WHERE t.status = ?
			  AND (t.assignee = '' OR t.lock_until = 0 OR t.lock_until < ?)
			ORDER BY
				CASE t.priority
					WHEN 'high' THEN 1
					WHEN 'critical' THEN 2
					WHEN 'medium' THEN 3
					WHEN 'low' THEN 4
					ELSE 5
				END
			LIMIT 1
		`, StatusTodo, now)
		if errors.Is(err, sql.ErrNoRows) {
			task = nil
			return nil // No task found, not an error
		}
		if err != nil {
			return fmt.Errorf("failed to query claimable task: %w", err)
		}

		task = &Todo2Task{
			ID:              row.ID,
			Content:         row.Content,
			LongDescription: row.LongDescription,
			Status:          row.Status,
			StatusEnum:      models.ParseTaskStatus(row.Status),
			Priority:        row.Priority,
			PriorityEnum:    models.ParseTaskPriority(row.Priority),
			Completed:       row.Completed == 1,
			CreatedAt:       row.Created,
			LastModified:    row.LastModified,
			CompletedAt:     row.CompletedAt,
			ParentID:        row.ParentID,
			AssignedTo:      row.AssignedTo,
			Host:            row.Host,
			Agent:           row.Agent,
			Version:         row.Version,
		}
		if row.ProjectID.Valid {
			task.ProjectID = row.ProjectID.String
		}

		task.Metadata = DeserializeTaskMetadata(string(row.Metadata), row.MetadataProto, row.MetadataFormat)

		task.NormalizeEpochDates()

		tags, err := parseJSONArrayToStrings(row.TagsJSON)
		if err != nil {
			return fmt.Errorf("failed to parse tags: %w", err)
		}
		task.Tags = tags

		deps, err := parseJSONArrayToStrings(row.DepsJSON)
		if err != nil {
			return fmt.Errorf("failed to parse dependencies: %w", err)
		}
		task.Dependencies = deps

		return nil
	})

	return task, err
}
