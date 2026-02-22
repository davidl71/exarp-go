// tasks_misc.go — FixTaskDates, GetDependencies, GetDependents, GetTagsForTask.
// Same package as tasks.go; uses loadTaskTags, loadTaskDependencies.
package database

import (
	"context"
	"fmt"
	"strings"
)

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
