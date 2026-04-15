// tasks_helpers.go — Helpers for loading task tags and dependencies.
// Used by GetTask, ListTasks, GetDependencies, GetTagsForTask. Same package as tasks.go.
package database

import (
	"context"
	"database/sql"
	"fmt"
)

// querier interface for both *sql.DB and *sql.Tx
type querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// loadStrings loads string values from a query
func loadStrings(ctx context.Context, queryCtx context.Context, querier querier, query string, args ...interface{}) ([]string, error) {
	rows, err := querier.QueryContext(queryCtx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - this is cleanup
		}
	}()

	var results []string

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		results = append(results, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// loadTaskTags loads tags for a single task
func loadTaskTags(ctx context.Context, queryCtx context.Context, querier querier, taskID string) ([]string, error) {
	return loadStrings(ctx, queryCtx, querier,
		`SELECT tag FROM task_tags WHERE task_id = ? ORDER BY tag`,
		taskID)
}

// loadTaskDependencies loads dependencies for a single task
func loadTaskDependencies(ctx context.Context, queryCtx context.Context, querier querier, taskID string) ([]string, error) {
	return loadStrings(ctx, queryCtx, querier,
		`SELECT depends_on_id FROM task_dependencies WHERE task_id = ? ORDER BY depends_on_id`,
		taskID)
}
