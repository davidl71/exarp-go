package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
)

// TaskClaimResult represents the result of a task claim operation.
type TaskClaimResult struct {
	Success   bool
	Task      *models.Todo2Task
	Error     error
	WasLocked bool   // True if task was already locked by another agent
	LockedBy  string // Agent ID that currently holds the lock
}

// ClaimTaskForAgent atomically claims a task for an agent using SELECT FOR UPDATE
// Uses pessimistic locking to prevent race conditions
// Returns TaskClaimResult with success status and task details.
func ClaimTaskForAgent(ctx context.Context, taskID string, agentID string, leaseDuration time.Duration) (*TaskClaimResult, error) {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	result := &TaskClaimResult{
		Success: false,
	}

	err := retryWithBackoff(ctx, func() error {
		db, err := GetDB()
		if err != nil {
			result.Error = fmt.Errorf("failed to get database: %w", err)
			return result.Error
		}

		// SQLite doesn't support row-level locking (SELECT FOR UPDATE is a no-op)
		// We use transactions + WHERE clause checks instead:
		// 1. Transaction provides atomicity
		// 2. Check assignee in WHERE clause of UPDATE
		// 3. Optimistic locking with version field
		tx, err := db.BeginTx(txCtx, nil)
		if err != nil {
			result.Error = fmt.Errorf("failed to begin transaction: %w", err)
			return result.Error
		}

		defer func() {
			if err != nil {
				_ = tx.Rollback()
			}
		}()

		var taskData models.Todo2Task

		var currentAssignee sql.NullString

		var lockUntil sql.NullInt64

		var version int64

		now := time.Now().Unix()
		leaseUntil := now + int64(leaseDuration.Seconds())

		// SQLite doesn't support SELECT FOR UPDATE, but transactions provide atomicity
		// We check assignee in the UPDATE WHERE clause instead
		err = tx.QueryRowContext(txCtx, `
			SELECT 
				id, content, long_description, status, priority, completed,
				assignee, lock_until, version
			FROM tasks
			WHERE id = ? AND (status = ? OR status = ?)
		`, taskID, StatusTodo, StatusInProgress).Scan(
			&taskData.ID,
			&taskData.Content,
			&taskData.LongDescription,
			&taskData.Status,
			&taskData.Priority,
			&taskData.Completed,
			&currentAssignee,
			&lockUntil,
			&version,
		)

		if errors.Is(err, sql.ErrNoRows) {
			result.Error = fmt.Errorf("task %s not found or not available (must be Todo or In Progress)", taskID)
			return result.Error
		}

		if err != nil {
			result.Error = fmt.Errorf("failed to query task: %w", err)
			return result.Error
		}

		// Check if task is already assigned to another agent
		if currentAssignee.Valid && currentAssignee.String != "" {
			// Check if lease is expired
			if lockUntil.Valid && lockUntil.Int64 > now {
				// Lock is still valid
				result.WasLocked = true
				result.LockedBy = currentAssignee.String
				result.Error = fmt.Errorf("task %s already assigned to %s (task: %s, lock expires at %d)", taskID, currentAssignee.String, taskID, lockUntil.Int64)

				return result.Error
			}
			// Lease expired - can reassign (checked in WHERE clause)
		}

		// Claim task atomically
		// SQLite: Check assignee in WHERE clause to prevent race conditions
		// If assignee is NULL or expired, this UPDATE will succeed
		// If assignee is set and not expired, this UPDATE will affect 0 rows
		updateResult, err := tx.ExecContext(txCtx, `
			UPDATE tasks SET
				assignee = ?,
				assigned_at = ?,
				lock_until = ?,
				status = ?,
				version = version + 1,
				updated_at = strftime('%s', 'now')
			WHERE id = ? 
			  AND version = ?
			  AND (assignee IS NULL OR lock_until IS NULL OR lock_until < ?)
		`, agentID, now, leaseUntil, StatusInProgress, taskID, version, now)

		if err != nil {
			result.Error = fmt.Errorf("failed to update task: %w", err)
			return result.Error
		}

		rowsAffected, err := updateResult.RowsAffected()
		if err != nil {
			result.Error = fmt.Errorf("failed to get rows affected: %w", err)
			return result.Error
		}

		if rowsAffected == 0 {
			// Version mismatch - task was modified concurrently
			result.Error = fmt.Errorf("task was modified by another agent (version mismatch)")
			return result.Error
		}

		// Load full task details (tags, dependencies)
		fullTask, err := loadTaskWithRelations(txCtx, tx, taskID)
		if err != nil {
			result.Error = fmt.Errorf("failed to load task details: %w", err)
			return result.Error
		}

		taskData.Tags = fullTask.Tags
		taskData.Dependencies = fullTask.Dependencies
		taskData.Status = StatusInProgress

		// Commit transaction (releases SELECT FOR UPDATE lock)
		if err = tx.Commit(); err != nil {
			result.Error = fmt.Errorf("failed to commit transaction: %w", err)
			return result.Error
		}

		result.Success = true
		result.Task = &taskData

		return nil
	})

	if err != nil && result.Error == nil {
		result.Error = err
	}

	// Return result.Error as the error if set, otherwise nil
	if result.Error != nil {
		return result, result.Error
	}

	return result, nil
}

// ReleaseTask releases a task lock (sets assignee to NULL)
// Only the current assignee can release their own lock.
func ReleaseTask(ctx context.Context, taskID string, agentID string) error {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	return retryWithBackoff(ctx, func() error {
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

		// Query current assignee (transaction provides atomicity)
		var currentAssignee sql.NullString
		err = tx.QueryRowContext(txCtx, `
			SELECT assignee
			FROM tasks
			WHERE id = ?
		`, taskID).Scan(&currentAssignee)

		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("task %s not found", taskID)
		}

		if err != nil {
			return fmt.Errorf("failed to query task: %w", err)
		}

		// Verify current assignee matches
		if !currentAssignee.Valid || currentAssignee.String != agentID {
			currentAssigneeStr := "none"
			if currentAssignee.Valid {
				currentAssigneeStr = fmt.Sprintf("%s (task: %s)", currentAssignee.String, taskID)
			}

			return fmt.Errorf("task %s not assigned to %s (task: %s) (current assignee: %s)", taskID, agentID, taskID, currentAssigneeStr)
		}

		// Release lock
		result, err := tx.ExecContext(txCtx, `
			UPDATE tasks SET
				assignee = NULL,
				assigned_at = NULL,
				lock_until = NULL,
				version = version + 1,
				updated_at = strftime('%s', 'now')
			WHERE id = ? AND assignee = ?
		`, taskID, agentID)

		if err != nil {
			return fmt.Errorf("failed to release lock: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("task %s not assigned to %s (task: %s)", taskID, agentID, taskID)
		}

		return tx.Commit()
	})
}

// RenewLease extends the lock duration for an already-claimed task
// Only the current assignee can renew their lease.
func RenewLease(ctx context.Context, taskID string, agentID string, leaseDuration time.Duration) error {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	return retryWithBackoff(ctx, func() error {
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

		// Query current assignee (transaction provides atomicity)
		var currentAssignee sql.NullString
		err = tx.QueryRowContext(txCtx, `
			SELECT assignee
			FROM tasks
			WHERE id = ?
		`, taskID).Scan(&currentAssignee)

		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("task %s not found", taskID)
		}

		if err != nil {
			return fmt.Errorf("failed to query task: %w", err)
		}

		// Verify current assignee matches
		if !currentAssignee.Valid || currentAssignee.String != agentID {
			currentAssigneeStr := "none"
			if currentAssignee.Valid {
				currentAssigneeStr = fmt.Sprintf("%s (task: %s)", currentAssignee.String, taskID)
			}

			return fmt.Errorf("task %s not assigned to %s (task: %s) (current assignee: %s)", taskID, agentID, taskID, currentAssigneeStr)
		}

		// Renew lease
		now := time.Now().Unix()
		leaseUntil := now + int64(leaseDuration.Seconds())

		result, err := tx.ExecContext(txCtx, `
			UPDATE tasks SET
				lock_until = ?,
				version = version + 1,
				updated_at = strftime('%s', 'now')
			WHERE id = ? AND assignee = ?
		`, leaseUntil, taskID, agentID)

		if err != nil {
			return fmt.Errorf("failed to renew lease: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("task %s not assigned to %s (task: %s)", taskID, agentID, taskID)
		}

		return tx.Commit()
	})
}

// RunLeaseRenewal starts a background goroutine that renews the task lease every renewInterval
// until ctx is cancelled. Call this when holding a task longer than leaseDuration (e.g. long-running work).
// renewInterval should be less than leaseDuration (e.g. renew every 20 min for a 30 min lease).
// The goroutine stops when ctx.Done(); release the task with ReleaseTask when work is done.
func RunLeaseRenewal(ctx context.Context, taskID, agentID string, leaseDuration, renewInterval time.Duration) {
	if renewInterval <= 0 || leaseDuration <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(renewInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = RenewLease(ctx, taskID, agentID, leaseDuration)
			}
		}
	}()
}

// CleanupExpiredLocks releases locks that have expired (for dead agent cleanup)
// Returns number of locks cleaned up.
func CleanupExpiredLocks(ctx context.Context) (int, error) {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	var cleaned int

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

		// Find and release expired locks
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
			WHERE assignee IS NOT NULL
			  AND lock_until IS NOT NULL
			  AND lock_until < ?
		`, now)

		if err != nil {
			return fmt.Errorf("failed to cleanup expired locks: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		cleaned = int(rowsAffected)

		return tx.Commit()
	})

	return cleaned, err
}

// BatchClaimTasks atomically claims multiple tasks (all or nothing)
// Uses state-level locking to ensure atomic batch assignment.
func BatchClaimTasks(ctx context.Context, taskIDs []string, agentID string, leaseDuration time.Duration) ([]string, []string, error) {
	ctx = ensureContext(ctx)

	txCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()

	var claimed []string

	var failed []string

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
		leaseUntil := now + int64(leaseDuration.Seconds())

		// Check all tasks are available (SELECT FOR UPDATE for each)
		for _, taskID := range taskIDs {
			var currentAssignee sql.NullString

			var lockUntil sql.NullInt64

			var version int64

			var status string

			err = tx.QueryRowContext(txCtx, `
				SELECT assignee, lock_until, version, status
				FROM tasks
				WHERE id = ?
			`, taskID).Scan(&currentAssignee, &lockUntil, &version, &status)

			if errors.Is(err, sql.ErrNoRows) {
				failed = append(failed, taskID)
				continue
			}

			if err != nil {
				return fmt.Errorf("failed to query task %s: %w", taskID, err)
			}

			// Check if available
			if currentAssignee.Valid && currentAssignee.String != "" {
				if lockUntil.Valid && lockUntil.Int64 > now {
					failed = append(failed, taskID)
					continue
				}
			}

			if status != StatusTodo && status != StatusInProgress {
				failed = append(failed, taskID)
				continue
			}

			// Claim task
			_, err = tx.ExecContext(txCtx, `
				UPDATE tasks SET
					assignee = ?,
					assigned_at = ?,
					lock_until = ?,
					status = ?,
					version = version + 1,
					updated_at = strftime('%s', 'now')
				WHERE id = ? AND version = ?
			`, agentID, now, leaseUntil, StatusInProgress, taskID, version)

			if err != nil {
				return fmt.Errorf("failed to claim task %s: %w", taskID, err)
			}

			claimed = append(claimed, taskID)
		}

		// Commit only if all tasks were claimed
		if len(failed) > 0 {
			return fmt.Errorf("failed to claim %d tasks", len(failed))
		}

		return tx.Commit()
	})

	return claimed, failed, err
}

// Helper function to load task with relations (tags, dependencies).
func loadTaskWithRelations(ctx context.Context, tx *sql.Tx, taskID string) (*models.Todo2Task, error) {
	task := &models.Todo2Task{ID: taskID}

	// Load tags
	tags, err := loadTaskTags(ctx, ctx, tx, taskID)
	if err != nil {
		return nil, err
	}

	task.Tags = tags

	// Load dependencies
	dependencies, err := loadTaskDependencies(ctx, ctx, tx, taskID)
	if err != nil {
		return nil, err
	}

	task.Dependencies = dependencies

	return task, nil
}
