-- SQLite Migration: Normalize nullable columns to NOT NULL DEFAULT ''
-- Version: 9
-- Description: Eliminate NULL values in optional text columns by using empty string as default.
-- This removes need for sql.NullString handling in application code (~300 lines saved).
-- Uses table recreation approach for maximum SQLite compatibility.

-- Step 1: Create new tasks table with NOT NULL columns
CREATE TABLE tasks_new (
    id TEXT PRIMARY KEY,
    name TEXT,
    content TEXT,
    long_description TEXT,
    status TEXT NOT NULL DEFAULT 'Todo',
    priority TEXT,
    completed INTEGER DEFAULT 0,
    task_number INTEGER,
    estimated_hours REAL,
    actual_hours REAL,
    created TEXT NOT NULL,
    last_modified TEXT NOT NULL DEFAULT '',
    completed_at TEXT NOT NULL DEFAULT '',
    project_id TEXT,
    metadata TEXT,
    metadata_protobuf BLOB,
    metadata_format TEXT DEFAULT 'json',
    version INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    parent_id TEXT NOT NULL DEFAULT '',
    assigned_to TEXT NOT NULL DEFAULT '',
    host TEXT NOT NULL DEFAULT '',
    agent TEXT NOT NULL DEFAULT ''
);

-- Step 2: Copy data from old table (NULLs become empty strings, protobuf/format preserved)
INSERT INTO tasks_new SELECT 
    id, name, content, long_description, status, priority, completed,
    task_number, estimated_hours, actual_hours, created,
    COALESCE(last_modified, ''),
    COALESCE(completed_at, ''),
    project_id, metadata, metadata_protobuf, metadata_format, version, created_at, updated_at,
    COALESCE(parent_id, ''),
    COALESCE(assigned_to, ''),
    COALESCE(host, ''),
    COALESCE(agent, '')
FROM tasks;

-- Step 3: Drop old table
DROP TABLE tasks;

-- Step 4: Rename new table to tasks
ALTER TABLE tasks_new RENAME TO tasks;

-- Step 5: Recreate indexes (original indexes plus new ones for NOT NULL columns)
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_tasks_created ON tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_completed ON tasks(completed);
CREATE INDEX IF NOT EXISTS idx_tasks_last_modified ON tasks(last_modified);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_metadata_format ON tasks(metadata_format);

-- assigned_to, host, agent indexes - now simpler since columns are NOT NULL
CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX IF NOT EXISTS idx_tasks_host ON tasks(host);
CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent);

-- Existing indexes from earlier migrations
CREATE INDEX IF NOT EXISTS idx_task_tags_task ON task_tags(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tags_tag ON task_tags(tag);
CREATE INDEX IF NOT EXISTS idx_task_deps_task ON task_dependencies(task_id);
CREATE INDEX IF NOT EXISTS idx_task_deps_depends ON task_dependencies(depends_on_id);
CREATE INDEX IF NOT EXISTS idx_task_changes_task ON task_changes(task_id);
CREATE INDEX IF NOT EXISTS idx_task_changes_timestamp ON task_changes(timestamp);
CREATE INDEX IF NOT EXISTS idx_task_comments_task ON task_comments(task_id);
CREATE INDEX IF NOT EXISTS idx_task_comments_type ON task_comments(comment_type);
CREATE INDEX IF NOT EXISTS idx_task_activities_task ON task_activities(task_id);
CREATE INDEX IF NOT EXISTS idx_task_activities_type ON task_activities(activity_type);
CREATE INDEX IF NOT EXISTS idx_task_activities_timestamp ON task_activities(timestamp);
