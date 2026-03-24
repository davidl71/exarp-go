// lock_monitoring.go — Stale lock detection, cleanup, and lease monitoring.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// LockStatus represents the status of a task lock.
type LockStatus struct {
	TaskID        string
	Assignee      string
	AssignedAt    time.Time
	LockUntil     time.Time
	IsExpired     bool
	IsStale       bool
	TimeRemaining time.Duration
	TimeExpired   time.Duration
}

// StaleLockInfo provides detailed information about stale locks.
type StaleLockInfo struct {
	ExpiredCount    int
	NearExpiryCount int // Within 5 minutes of expiry
	StaleCount      int // Expired for > 5 minutes
	Locks           []LockStatus
}

// GetActiveLocks returns all non-expired task locks ordered by expiry.
func GetActiveLocks(ctx context.Context) ([]LockStatus, error) {
	ctx = ensureContext(ctx)

	var locks []LockStatus

	err := retryWithBackoff(ctx, func() error {
		queryCtx, cancel, db, err := QueryContextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()

		now := time.Now().Unix()

		var rows []struct {
			TaskID     string        `db:"id"`
			Assignee   string        `db:"assignee"`
			AssignedAt sql.NullInt64 `db:"assigned_at"`
			LockUntil  sql.NullInt64 `db:"lock_until"`
		}

		if err := db.SelectContext(queryCtx, &rows, `
			SELECT id, assignee, assigned_at, lock_until
			FROM tasks
			WHERE assignee IS NOT NULL
			  AND lock_until IS NOT NULL
			  AND lock_until >= ?
			ORDER BY lock_until ASC
		`, now); err != nil {
			return fmt.Errorf("failed to query active locks: %w", err)
		}

		locks = nil
		for _, row := range rows {
			if !row.LockUntil.Valid {
				continue
			}
			lockTime := time.Unix(row.LockUntil.Int64, 0)
			locks = append(locks, LockStatus{
				TaskID:        row.TaskID,
				Assignee:      row.Assignee,
				AssignedAt:    time.Unix(row.AssignedAt.Int64, 0),
				LockUntil:     lockTime,
				TimeRemaining: time.Until(lockTime),
			})
		}
		return nil
	})

	return locks, err
}

// GetActiveLockMapForTasks returns active non-expired locks keyed by task ID.
func GetActiveLockMapForTasks(ctx context.Context, taskIDs []string) (map[string]LockStatus, error) {
	locks, err := GetActiveLocks(ctx)
	if err != nil {
		return nil, err
	}
	if len(taskIDs) == 0 {
		return map[string]LockStatus{}, nil
	}
	allowed := make(map[string]bool, len(taskIDs))
	for _, id := range taskIDs {
		allowed[id] = true
	}
	out := make(map[string]LockStatus)
	for _, lock := range locks {
		if allowed[lock.TaskID] {
			out[lock.TaskID] = lock
		}
	}
	return out, nil
}

// DetectStaleLocks finds all locks that are expired or near expiration
// Returns detailed information about lock status.
func DetectStaleLocks(ctx context.Context, nearExpiryThreshold time.Duration) (*StaleLockInfo, error) {
	ctx = ensureContext(ctx)

	info := &StaleLockInfo{
		Locks: []LockStatus{},
	}

	err := retryWithBackoff(ctx, func() error {
		queryCtx, cancel, db, err := QueryContextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()

		now := time.Now().Unix()
		nearExpiryTime := now + int64(nearExpiryThreshold.Seconds())

		var rows []struct {
			TaskID     string        `db:"id"`
			Assignee   string        `db:"assignee"`
			AssignedAt sql.NullInt64 `db:"assigned_at"`
			LockUntil  sql.NullInt64 `db:"lock_until"`
		}

		if err := db.SelectContext(queryCtx, &rows, `
			SELECT id, assignee, assigned_at, lock_until
			FROM tasks
			WHERE assignee IS NOT NULL
			  AND lock_until IS NOT NULL
			ORDER BY lock_until ASC
		`); err != nil {
			return fmt.Errorf("failed to query locked tasks: %w", err)
		}

		for _, row := range rows {
			if !row.LockUntil.Valid {
				continue
			}

			lockTime := time.Unix(row.LockUntil.Int64, 0)
			isExpired := row.LockUntil.Int64 < now
			isNearExpiry := !isExpired && row.LockUntil.Int64 < nearExpiryTime
			isStale := isExpired && (now-row.LockUntil.Int64) > 300

			var timeRemaining, timeExpired time.Duration
			if isExpired {
				timeExpired = time.Since(lockTime)
			} else {
				timeRemaining = time.Until(lockTime)
			}

			status := LockStatus{
				TaskID:        row.TaskID,
				Assignee:      row.Assignee,
				AssignedAt:    time.Unix(row.AssignedAt.Int64, 0),
				LockUntil:     lockTime,
				IsExpired:     isExpired,
				IsStale:       isStale,
				TimeRemaining: timeRemaining,
				TimeExpired:   timeExpired,
			}

			info.Locks = append(info.Locks, status)

			if isExpired {
				info.ExpiredCount++
				if isStale {
					info.StaleCount++
				}
			} else if isNearExpiry {
				info.NearExpiryCount++
			}
		}

		return nil
	})

	return info, err
}

// GetLockStatus returns the current lock status for a specific task.
func GetLockStatus(ctx context.Context, taskID string) (*LockStatus, error) {
	ctx = ensureContext(ctx)

	var status *LockStatus

	err := retryWithBackoff(ctx, func() error {
		queryCtx, cancel, db, err := QueryContextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()

		var row struct {
			Assignee   sql.NullString `db:"assignee"`
			AssignedAt sql.NullInt64  `db:"assigned_at"`
			LockUntil  sql.NullInt64  `db:"lock_until"`
		}

		err = db.GetContext(queryCtx, &row, `
			SELECT assignee, assigned_at, lock_until
			FROM tasks
			WHERE id = ?
		`, taskID)

		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("task %s not found", taskID)
		}

		if err != nil {
			return fmt.Errorf("failed to query lock status: %w", err)
		}

		if !row.Assignee.Valid || !row.LockUntil.Valid {
			return nil
		}

		now := time.Now().Unix()
		lockTime := time.Unix(row.LockUntil.Int64, 0)
		isExpired := row.LockUntil.Int64 < now
		isStale := isExpired && (now-row.LockUntil.Int64) > 300

		var timeRemaining, timeExpired time.Duration
		if isExpired {
			timeExpired = time.Since(lockTime)
		} else {
			timeRemaining = time.Until(lockTime)
		}

		status = &LockStatus{
			TaskID:        taskID,
			Assignee:      row.Assignee.String,
			AssignedAt:    time.Unix(row.AssignedAt.Int64, 0),
			LockUntil:     lockTime,
			IsExpired:     isExpired,
			IsStale:       isStale,
			TimeRemaining: timeRemaining,
			TimeExpired:   timeExpired,
		}

		return nil
	})

	return status, err
}

// GetLocksByAgent returns all active locks held by a specific agent.
func GetLocksByAgent(ctx context.Context, agentID string) ([]LockStatus, error) {
	ctx = ensureContext(ctx)

	var locks []LockStatus

	err := retryWithBackoff(ctx, func() error {
		queryCtx, cancel, db, err := QueryContextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()

		now := time.Now().Unix()

		var rows []struct {
			TaskID     string        `db:"id"`
			Assignee   string        `db:"assignee"`
			AssignedAt sql.NullInt64 `db:"assigned_at"`
			LockUntil  sql.NullInt64 `db:"lock_until"`
		}

		if err := db.SelectContext(queryCtx, &rows, `
			SELECT id, assignee, assigned_at, lock_until
			FROM tasks
			WHERE assignee = ? AND lock_until IS NOT NULL
			ORDER BY lock_until ASC
		`, agentID); err != nil {
			return fmt.Errorf("failed to query locks: %w", err)
		}

		locks = nil
		for _, row := range rows {
			if !row.LockUntil.Valid {
				continue
			}

			lockTime := time.Unix(row.LockUntil.Int64, 0)
			isExpired := row.LockUntil.Int64 < now
			isStale := isExpired && (now-row.LockUntil.Int64) > 300

			var timeRemaining, timeExpired time.Duration
			if isExpired {
				timeExpired = time.Since(lockTime)
			} else {
				timeRemaining = time.Until(lockTime)
			}

			locks = append(locks, LockStatus{
				TaskID:        row.TaskID,
				Assignee:      row.Assignee,
				AssignedAt:    time.Unix(row.AssignedAt.Int64, 0),
				LockUntil:     lockTime,
				IsExpired:     isExpired,
				IsStale:       isStale,
				TimeRemaining: timeRemaining,
				TimeExpired:   timeExpired,
			})
		}

		return nil
	})

	return locks, err
}

// CleanupExpiredLocksWithReport releases locks that have expired and returns detailed report
// Returns number of locks cleaned up and list of cleaned task IDs.
func CleanupExpiredLocksWithReport(ctx context.Context, maxAge time.Duration) (int, []string, error) {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	var cleaned int

	var cleanedTaskIDs []string

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
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

		now := time.Now().Unix()
		maxAgeUnix := now - int64(maxAge.Seconds())

		// Find expired locks (also consider locks older than maxAge even if not expired)
		rows, err := tx.QueryContext(txCtx, `
			SELECT id, assignee, lock_until
			FROM tasks
			WHERE assignee IS NOT NULL
			  AND lock_until IS NOT NULL
			  AND (lock_until < ? OR assigned_at < ?)
			ORDER BY lock_until ASC
		`, now, maxAgeUnix)
		if err != nil {
			return fmt.Errorf("failed to query expired locks: %w", err)
		}
		defer rows.Close()

		var taskIDsToClean []string

		for rows.Next() {
			var taskID, assignee string

			var lockUntil sql.NullInt64

			if err := rows.Scan(&taskID, &assignee, &lockUntil); err != nil {
				return fmt.Errorf("failed to scan expired lock: %w", err)
			}

			taskIDsToClean = append(taskIDsToClean, taskID)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("failed to iterate expired locks: %w", err)
		}

		// Release expired locks (one at a time for simplicity)
		if len(taskIDsToClean) > 0 {
			for _, taskID := range taskIDsToClean {
				result, err := tx.ExecContext(txCtx, `
					UPDATE tasks SET
						assignee = NULL,
						assigned_at = NULL,
						lock_until = NULL,
						status = CASE 
							WHEN status = 'In Progress' THEN 'Todo'
							ELSE status
						END,
						version = version + 1,
						updated_at = strftime('%s', 'now')
					WHERE id = ?
				`, taskID)
				if err != nil {
					return fmt.Errorf("failed to cleanup expired lock for task %s: %w", taskID, err)
				}

				rowsAffected, err := result.RowsAffected()
				if err != nil {
					return fmt.Errorf("failed to get rows affected: %w", err)
				}

				if rowsAffected > 0 {
					cleaned++

					cleanedTaskIDs = append(cleanedTaskIDs, taskID)
				}
			}
		}

		return tx.Commit()
	})

	return cleaned, cleanedTaskIDs, err
}

// CleanupDeadAgentLocks releases expired locks only when the assigning agent's process
// no longer exists (dead agent). Uses DetectStaleLocks and AgentProcessExists per
// docs/STALE_LOCK_DETECTION_AND_HANDLING.md. Returns cleaned count and task IDs.
func CleanupDeadAgentLocks(ctx context.Context, staleThreshold time.Duration) (int, []string, error) {
	info, err := DetectStaleLocks(ctx, staleThreshold)
	if err != nil {
		return 0, nil, fmt.Errorf("detect stale locks: %w", err)
	}

	// Collect expired locks where agent process does not exist
	var taskIDsToClean []string

	for _, lock := range info.Locks {
		if !lock.IsExpired {
			continue
		}
		// Skip if agent process is still running (might renew soon)
		if AgentProcessExists(lock.Assignee) {
			continue
		}

		taskIDsToClean = append(taskIDsToClean, lock.TaskID)
	}

	if len(taskIDsToClean) == 0 {
		return 0, nil, nil
	}

	return releaseLocksForTaskIDs(ctx, taskIDsToClean)
}

// releaseLocksForTaskIDs clears assignee/lock_until for the given task IDs
// (used by CleanupDeadAgentLocks). Returns cleaned count and cleaned task IDs.
func releaseLocksForTaskIDs(ctx context.Context, taskIDs []string) (int, []string, error) {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	var cleaned int

	var cleanedTaskIDs []string

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDBx()
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

		for _, taskID := range taskIDs {
			result, err := tx.ExecContext(txCtx, `
				UPDATE tasks SET
					assignee = NULL,
					assigned_at = NULL,
					lock_until = NULL,
					status = CASE 
						WHEN status = 'In Progress' THEN 'Todo'
						ELSE status
					END,
					version = version + 1,
					updated_at = strftime('%s', 'now')
				WHERE id = ?
			`, taskID)
			if err != nil {
				return fmt.Errorf("failed to release lock for task %s: %w", taskID, err)
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return fmt.Errorf("failed to get rows affected: %w", err)
			}

			if rowsAffected > 0 {
				cleaned++

				cleanedTaskIDs = append(cleanedTaskIDs, taskID)
			}
		}

		return tx.Commit()
	})

	return cleaned, cleanedTaskIDs, err
}
