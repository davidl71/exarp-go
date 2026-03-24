// tasks_misc.go — FixTaskDates, GetDependencies, GetDependents, GetTagsForTask.
// Same package as tasks.go; uses loadTaskTags, loadTaskDependencies.
package database

import (
	"context"
	"fmt"
	"strings"
)

// DatabaseStatus describes explicit database health/maintenance state.
type DatabaseStatus struct {
	Driver             string `json:"driver"`
	JournalMode        string `json:"journal_mode,omitempty"`
	AutoVacuum         string `json:"auto_vacuum,omitempty"`
	PageSize           int64  `json:"page_size,omitempty"`
	PageCount          int64  `json:"page_count,omitempty"`
	FreelistCount      int64  `json:"freelist_count,omitempty"`
	BusyTimeoutMS      int64  `json:"busy_timeout_ms,omitempty"`
	WALAutocheckpoint  int64  `json:"wal_autocheckpoint,omitempty"`
	EstimatedDBBytes   int64  `json:"estimated_db_bytes,omitempty"`
	EstimatedFreeBytes int64  `json:"estimated_free_bytes,omitempty"`
}

// CheckpointResult describes a WAL checkpoint run.
type CheckpointResult struct {
	Mode         string `json:"mode"`
	Busy         int64  `json:"busy"`
	LogFrames    int64  `json:"log_frames"`
	Checkpointed int64  `json:"checkpointed_frames"`
}

// CurrentDriverType returns the active driver type, if initialized.
func CurrentDriverType() DriverType {
	if currentDriver == nil {
		return ""
	}
	return currentDriver.Type()
}

// FixTaskDates backfills created and last_modified from created_at/updated_at (Unix epoch)
// for rows where created or last_modified is empty or 1970-01-01. Returns the number of rows updated.
func FixTaskDates(ctx context.Context) (int64, error) {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	var rowsAffected int64

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
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
	db, err := GetDBx()
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
	db, err := GetDBx()
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

// GetDatabaseStatus returns explicit runtime database status.
func GetDatabaseStatus(ctx context.Context) (*DatabaseStatus, error) {
	db, err := GetDBx()
	if err != nil {
		return nil, err
	}

	status := &DatabaseStatus{
		Driver: string(CurrentDriverType()),
	}
	if status.Driver == "" {
		status.Driver = string(DriverSQLite)
	}
	if status.Driver != string(DriverSQLite) {
		return status, nil
	}

	queryCtx, cancel := withQueryTimeout(ensureContext(ctx))
	defer cancel()

	if err := db.QueryRowxContext(queryCtx, "PRAGMA journal_mode").Scan(&status.JournalMode); err != nil {
		return nil, fmt.Errorf("pragma journal_mode: %w", err)
	}
	if err := db.QueryRowxContext(queryCtx, "PRAGMA auto_vacuum").Scan(&status.AutoVacuum); err != nil {
		return nil, fmt.Errorf("pragma auto_vacuum: %w", err)
	}
	if err := db.QueryRowxContext(queryCtx, "PRAGMA page_size").Scan(&status.PageSize); err != nil {
		return nil, fmt.Errorf("pragma page_size: %w", err)
	}
	if err := db.QueryRowxContext(queryCtx, "PRAGMA page_count").Scan(&status.PageCount); err != nil {
		return nil, fmt.Errorf("pragma page_count: %w", err)
	}
	if err := db.QueryRowxContext(queryCtx, "PRAGMA freelist_count").Scan(&status.FreelistCount); err != nil {
		return nil, fmt.Errorf("pragma freelist_count: %w", err)
	}
	if err := db.QueryRowxContext(queryCtx, "PRAGMA busy_timeout").Scan(&status.BusyTimeoutMS); err != nil {
		return nil, fmt.Errorf("pragma busy_timeout: %w", err)
	}
	if err := db.QueryRowxContext(queryCtx, "PRAGMA wal_autocheckpoint").Scan(&status.WALAutocheckpoint); err != nil {
		return nil, fmt.Errorf("pragma wal_autocheckpoint: %w", err)
	}

	status.EstimatedDBBytes = status.PageCount * status.PageSize
	status.EstimatedFreeBytes = status.FreelistCount * status.PageSize

	return status, nil
}

// RunCheckpoint performs an explicit SQLite WAL checkpoint.
func RunCheckpoint(ctx context.Context, mode string) (*CheckpointResult, error) {
	db, err := GetDBx()
	if err != nil {
		return nil, err
	}
	if CurrentDriverType() != "" && CurrentDriverType() != DriverSQLite {
		return nil, fmt.Errorf("checkpoint is only supported for sqlite")
	}

	mode = strings.ToUpper(strings.TrimSpace(mode))
	if mode == "" {
		mode = "PASSIVE"
	}
	switch mode {
	case "PASSIVE", "FULL", "RESTART", "TRUNCATE":
	default:
		return nil, fmt.Errorf("unsupported checkpoint mode %q", mode)
	}

	queryCtx, cancel := withQueryTimeout(ensureContext(ctx))
	defer cancel()

	result := &CheckpointResult{Mode: mode}
	if err := db.QueryRowxContext(queryCtx, fmt.Sprintf("PRAGMA wal_checkpoint(%s)", mode)).Scan(&result.Busy, &result.LogFrames, &result.Checkpointed); err != nil {
		return nil, fmt.Errorf("wal_checkpoint(%s): %w", mode, err)
	}

	return result, nil
}

// VacuumDatabase runs an explicit VACUUM.
func VacuumDatabase(ctx context.Context) error {
	db, err := GetDBx()
	if err != nil {
		return err
	}
	if CurrentDriverType() != "" && CurrentDriverType() != DriverSQLite {
		return fmt.Errorf("vacuum is only supported for sqlite")
	}

	queryCtx, cancel := withQueryTimeout(ensureContext(ctx))
	defer cancel()

	if _, err := db.ExecContext(queryCtx, "VACUUM"); err != nil {
		return fmt.Errorf("vacuum: %w", err)
	}

	return nil
}

// AnalyzeDatabase runs explicit ANALYZE.
func AnalyzeDatabase(ctx context.Context) error {
	db, err := GetDBx()
	if err != nil {
		return err
	}

	queryCtx, cancel := withQueryTimeout(ensureContext(ctx))
	defer cancel()

	if _, err := db.ExecContext(queryCtx, "ANALYZE"); err != nil {
		return fmt.Errorf("analyze: %w", err)
	}

	return nil
}
