-- 011_add_status_lock_index.sql — Composite index on (status, lock_until) for FindNextClaimableTask.
-- The query filters on status = ? AND (assignee IS NULL OR lock_until IS NULL OR lock_until < ?).
-- idx_tasks_status_assignee covers status+assignee but not lock_until; this index closes the gap.

CREATE INDEX IF NOT EXISTS idx_tasks_status_lock ON tasks(status, lock_until);
