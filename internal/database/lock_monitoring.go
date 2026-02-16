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

// DetectStaleLocks finds all locks that are expired or near expiration
// Returns detailed information about lock status.
func DetectStaleLocks(ctx context.Context, nearExpiryThreshold time.Duration) (*StaleLockInfo, error) {
	ctx = ensureContext(ctx)

	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	info := &StaleLockInfo{
		Locks: []LockStatus{},
	}

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		now := time.Now().Unix()
		nearExpiryTime := now + int64(nearExpiryThreshold.Seconds())

		// Find all locked tasks (assignee is not NULL)
		rows, err := db.QueryContext(queryCtx, `
			SELECT 
				id,
				assignee,
				assigned_at,
				lock_until
			FROM tasks
			WHERE assignee IS NOT NULL
			  AND lock_until IS NOT NULL
			ORDER BY lock_until ASC
		`)
		if err != nil {
			return fmt.Errorf("failed to query locked tasks: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var taskID, assignee string

			var assignedAt, lockUntil sql.NullInt64

			if err := rows.Scan(&taskID, &assignee, &assignedAt, &lockUntil); err != nil {
				return fmt.Errorf("failed to scan lock status: %w", err)
			}

			if !lockUntil.Valid {
				continue
			}

			lockTime := time.Unix(lockUntil.Int64, 0)
			isExpired := lockUntil.Int64 < now
			isNearExpiry := !isExpired && lockUntil.Int64 < nearExpiryTime
			isStale := isExpired && (now-lockUntil.Int64) > 300 // Expired for > 5 minutes

			var timeRemaining, timeExpired time.Duration
			if isExpired {
				timeExpired = time.Since(lockTime)
			} else {
				timeRemaining = time.Until(lockTime)
			}

			status := LockStatus{
				TaskID:        taskID,
				Assignee:      assignee,
				AssignedAt:    time.Unix(assignedAt.Int64, 0),
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

		return rows.Err()
	})

	return info, err
}

// GetLockStatus returns the current lock status for a specific task.
func GetLockStatus(ctx context.Context, taskID string) (*LockStatus, error) {
	ctx = ensureContext(ctx)

	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var status *LockStatus

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		var assignee sql.NullString

		var assignedAt, lockUntil sql.NullInt64

		err = db.QueryRowContext(queryCtx, `
			SELECT assignee, assigned_at, lock_until
			FROM tasks
			WHERE id = ?
		`, taskID).Scan(&assignee, &assignedAt, &lockUntil)

		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("task %s not found", taskID)
		}

		if err != nil {
			return fmt.Errorf("failed to query lock status: %w", err)
		}

		if !assignee.Valid || !lockUntil.Valid {
			// Task is not locked
			return nil
		}

		now := time.Now().Unix()
		lockTime := time.Unix(lockUntil.Int64, 0)
		isExpired := lockUntil.Int64 < now
		isStale := isExpired && (now-lockUntil.Int64) > 300 // Expired for > 5 minutes

		var timeRemaining, timeExpired time.Duration
		if isExpired {
			timeExpired = time.Since(lockTime)
		} else {
			timeRemaining = time.Until(lockTime)
		}

		status = &LockStatus{
			TaskID:        taskID,
			Assignee:      assignee.String,
			AssignedAt:    time.Unix(assignedAt.Int64, 0),
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

// CleanupExpiredLocksWithReport releases locks that have expired and returns detailed report
// Returns number of locks cleaned up and list of cleaned task IDs.
func CleanupExpiredLocksWithReport(ctx context.Context, maxAge time.Duration) (int, []string, error) {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	var cleaned int

	var cleanedTaskIDs []string

	err := retryWithBackoff(ctx, func() error {
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
