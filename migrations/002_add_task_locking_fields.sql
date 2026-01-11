-- SQLite Migration: Add Task Locking Fields
-- Version: 2
-- Created: 2026-01-10
-- Description: Add version, assignee, and lease fields for agent locking

-- ============================================================================
-- Add Locking Fields to Tasks Table
-- ============================================================================

-- Version field for optimistic locking (detect concurrent modifications)
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
