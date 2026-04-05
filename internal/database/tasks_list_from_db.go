// tasks_list_from_db.go — List tasks using an explicit *sqlx.DB (e.g. read-only import DB).
package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
	"github.com/jmoiron/sqlx"
)

// listTasksFromDB runs the ListTasks query + tag/dependency hydration against db (not the global pool).
func listTasksFromDB(ctx context.Context, db *sqlx.DB, filters *TaskFilters) ([]*Todo2Task, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}

	var queryBuilder strings.Builder

	queryBuilder.WriteString(`
			SELECT DISTINCT
			       t.id, t.name, t.content, t.long_description,
			       t.status, t.status_enum,
			       t.priority, t.priority_enum, t.priority_rank,
			       t.completed,
			       t.created, t.last_modified, t.completed_at,
			       t.created_ts, t.last_modified_ts, t.completed_at_ts,
			       t.created_at, t.updated_at,
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

		if filters.NameContains != nil {
			if q := strings.TrimSpace(*filters.NameContains); q != "" {
				pat := strings.ToLower(likeContainsPattern(q))
				conditions = append(conditions,
					"(LOWER(COALESCE(t.name,'')) LIKE ? ESCAPE '\\' OR LOWER(COALESCE(t.content,'')) LIKE ? ESCAPE '\\')")
				args = append(args, pat, pat)
			}
		}
	}

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE " + conditions[0])
		for i := 1; i < len(conditions); i++ {
			queryBuilder.WriteString(" AND " + conditions[i])
		}
	}

	queryBuilder.WriteString(` ORDER BY
			CASE t.priority
				WHEN 'critical' THEN 0
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
				ELSE 4
			END ASC,
			t.priority_rank ASC,
			t.created_ts DESC`)
	query := queryBuilder.String()

	var rows []taskRow
	if err := db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
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

		task.FillRFC3339FromSQLiteTimes(row.CreatedTS, row.LastModifiedTS, row.CompletedAtTS,
			row.InternalCreatedUnix, row.InternalUpdatedUnix)

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
			return nil, fmt.Errorf("failed to build tag query: %w", err)
		}

		var tagResults []struct {
			TaskID string `db:"task_id"`
			Tag    string `db:"tag"`
		}
		if err := db.SelectContext(ctx, &tagResults, tagQuery, tagArgs...); err != nil {
			return nil, fmt.Errorf("failed to batch query tags: %w", err)
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
			return nil, fmt.Errorf("failed to build dependency query: %w", err)
		}

		var depResults []struct {
			TaskID      string `db:"task_id"`
			DependsOnID string `db:"depends_on_id"`
		}
		if err := db.SelectContext(ctx, &depResults, depQuery, depArgs...); err != nil {
			return nil, fmt.Errorf("failed to batch query dependencies: %w", err)
		}

		for _, dr := range depResults {
			if task, ok := taskMap[dr.TaskID]; ok {
				task.Dependencies = append(task.Dependencies, dr.DependsOnID)
			}
		}
	}

	return taskList, nil
}
