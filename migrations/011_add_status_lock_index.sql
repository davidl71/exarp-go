-- 011_add_status_lock_index.sql — Restore locking columns dropped by migration 009 and add index.
-- Migration 009 recreated the tasks table without assignee/assigned_at/lock_until
-- (those were added in migration 002). Restore them so agent locking works on fresh DBs.

ALTER TABLE tasks ADD COLUMN assignee TEXT;
ALTER TABLE tasks ADD COLUMN assigned_at INTEGER;
ALTER TABLE tasks ADD COLUMN lock_until INTEGER;

-- Composite index on (status, lock_until) for FindNextClaimableTask:
-- WHERE status = ? AND (assignee IS NULL OR lock_until IS NULL OR lock_until < ?)
CREATE INDEX IF NOT EXISTS idx_tasks_status_lock ON tasks(status, lock_until);

-- Restore the status+assignee index from migration 002 (also dropped by 009)
CREATE INDEX IF NOT EXISTS idx_tasks_status_assignee ON tasks(status, assignee);
CREATE INDEX IF NOT EXISTS idx_tasks_lock_until ON tasks(lock_until) WHERE lock_until > 0;
