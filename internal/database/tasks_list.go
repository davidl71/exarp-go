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

// likeContainsPattern builds a LIKE pattern for case-insensitive substring match.
// %, _, and \ in the user input are escaped for use with "... LIKE ? ESCAPE '\'" (SQLite).
func likeContainsPattern(substr string) string {
	var b strings.Builder

	b.WriteByte('%')

	for _, r := range substr {
		switch r {
		case '\\', '%', '_':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}

	b.WriteByte('%')

	return b.String()
}

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

		var errList error
		tasks, errList = listTasksFromDB(queryCtx, db, filters)

		return errList
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
			SELECT t.id, t.content, t.long_description, t.status, t.priority, t.priority_rank, t.completed, t.created, t.last_modified,
			       t.completed_at, t.created_ts, t.last_modified_ts, t.completed_at_ts,
			       t.created_at, t.updated_at,
			       t.metadata, t.metadata_protobuf, t.metadata_format, t.parent_id, t.project_id,
			       t.assigned_to, t.host, t.agent, t.version`+sqlTaskAggJSON+`
			FROM tasks AS t
			WHERE t.status = ?
			  AND (t.assignee = '' OR t.lock_until = 0 OR t.lock_until < ?)
			ORDER BY
				CASE t.priority
					WHEN 'critical' THEN 0
					WHEN 'high' THEN 1
					WHEN 'medium' THEN 2
					WHEN 'low' THEN 3
					ELSE 4
				END,
				t.priority_rank ASC,
				t.created_ts ASC
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
			PriorityRank:    row.PriorityRank,
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

		task.FillRFC3339FromSQLiteTimes(row.CreatedTS, row.LastModifiedTS, row.CompletedAtTS,
			row.InternalCreatedUnix, row.InternalUpdatedUnix)
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
