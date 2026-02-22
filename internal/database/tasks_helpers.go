// tasks_helpers.go — Helpers for loading task tags and dependencies.
// Used by GetTask, ListTasks, GetDependencies, GetTagsForTask. Same package as tasks.go.
package database

import (
	"context"
	"database/sql"
	"fmt"
)

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
