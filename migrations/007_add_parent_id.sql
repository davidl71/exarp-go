-- SQLite Migration: Add parent_id to tasks
-- Version: 7
-- Description: Add parent_id column for hierarchical/epic relationships (separate from dependencies)

ALTER TABLE tasks ADD COLUMN parent_id TEXT;

CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
