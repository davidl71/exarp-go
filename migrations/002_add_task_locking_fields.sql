-- SQLite Migration: Add Task Locking Fields
-- Version: 2
-- Created: 2026-01-10
-- Description: Add version, assignee, and lease fields for agent locking

-- ============================================================================
-- Add Locking Fields to Tasks Table
-- ============================================================================

-- Note: version field is now in initial schema (001), but we add it here
-- for backward compatibility with databases created before version was added to 001.
-- SQLite doesn't support IF NOT EXISTS for ALTER TABLE ADD COLUMN, so we use
-- a workaround: attempt to add, ignore error if column already exists.

-- Version field for optimistic locking (detect concurrent modifications)
-- (May already exist if created with schema 001 that includes version)
-- SQLite will return error if column exists, which we handle in migration code
ALTER TABLE tasks ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- Assignee tracking (which agent/host has claimed the task)
ALTER TABLE tasks ADD COLUMN assignee TEXT;

-- Assignment timestamp (when task was assigned)
ALTER TABLE tasks ADD COLUMN assigned_at INTEGER;

-- Lease expiration timestamp (when lock expires, for dead agent cleanup)
ALTER TABLE tasks ADD COLUMN lock_until INTEGER;

-- ============================================================================
-- Indexes for Locking Operations
-- ============================================================================

-- Index for finding available tasks (status + assignee)
CREATE INDEX IF NOT EXISTS idx_tasks_status_assignee ON tasks(status, assignee);

-- Index for lease expiration cleanup (find expired locks)
CREATE INDEX IF NOT EXISTS idx_tasks_lock_until ON tasks(lock_until) WHERE lock_until > 0;

-- Index for version checks (optimistic locking)
CREATE INDEX IF NOT EXISTS idx_tasks_id_version ON tasks(id, version);
