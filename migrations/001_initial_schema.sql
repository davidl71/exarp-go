-- SQLite Migration: Initial Schema for Todo2 Tasks
-- Version: 1
-- Created: 2026-01-09
-- Description: Initial schema migration from JSON to SQLite

-- Note: PRAGMA statements are set in Init() function, not in migrations
-- (PRAGMA cannot be run inside transactions)

-- ============================================================================
-- Main Tables
-- ============================================================================

-- Tasks table: Main task storage
CREATE TABLE IF NOT EXISTS tasks (
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
    last_modified TEXT,
    completed_at TEXT,
    project_id TEXT,
    metadata TEXT,  -- JSON blob for flexible metadata
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Task tags: Many-to-many relationship
CREATE TABLE IF NOT EXISTS task_tags (
    task_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (task_id, tag),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Task dependencies: Many-to-many relationship
CREATE TABLE IF NOT EXISTS task_dependencies (
    task_id TEXT NOT NULL,
    depends_on_id TEXT NOT NULL,
    PRIMARY KEY (task_id, depends_on_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CHECK (task_id != depends_on_id)  -- Prevent self-dependencies
);

-- Task changes: Change history tracking
CREATE TABLE IF NOT EXISTS task_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    field TEXT NOT NULL,
    old_value TEXT,  -- JSON-encoded value
    new_value TEXT,  -- JSON-encoded value
    timestamp TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Task comments: Task comments storage
CREATE TABLE IF NOT EXISTS task_comments (
    id TEXT PRIMARY KEY,  -- Comment ID (e.g., "T-NaN-C-1")
    task_id TEXT NOT NULL,  -- todoId in JSON
    comment_type TEXT NOT NULL,  -- 'research_with_links', 'result', 'note', 'manualsetup'
    content TEXT NOT NULL,
    created TEXT NOT NULL,  -- ISO 8601 timestamp
    last_modified TEXT,  -- ISO 8601 timestamp
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Task activities: Activity log
CREATE TABLE IF NOT EXISTS task_activities (
    id TEXT PRIMARY KEY,  -- Activity ID (e.g., "activity-1767912853265-mv2lrmp39")
    activity_type TEXT NOT NULL,  -- 'todo_created', 'comment_added', 'status_changed', etc.
    task_id TEXT NOT NULL,  -- todoId
    task_name TEXT,  -- todoName
    timestamp TEXT NOT NULL,  -- ISO 8601 timestamp
    details TEXT,  -- JSON blob for activity-specific details
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Schema migrations: Track database schema versions
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    description TEXT
);

-- ============================================================================
-- Indexes for Performance
-- ============================================================================

-- Tasks indexes
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_tasks_created ON tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_completed ON tasks(completed);
CREATE INDEX IF NOT EXISTS idx_tasks_last_modified ON tasks(last_modified);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);

-- Task tags indexes
CREATE INDEX IF NOT EXISTS idx_task_tags_task ON task_tags(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tags_tag ON task_tags(tag);

-- Task dependencies indexes
CREATE INDEX IF NOT EXISTS idx_task_deps_task ON task_dependencies(task_id);
CREATE INDEX IF NOT EXISTS idx_task_deps_depends ON task_dependencies(depends_on_id);

-- Task changes indexes
CREATE INDEX IF NOT EXISTS idx_task_changes_task ON task_changes(task_id);
CREATE INDEX IF NOT EXISTS idx_task_changes_timestamp ON task_changes(timestamp);

-- Task comments indexes
CREATE INDEX IF NOT EXISTS idx_task_comments_task ON task_comments(task_id);
CREATE INDEX IF NOT EXISTS idx_task_comments_type ON task_comments(comment_type);

-- Task activities indexes
CREATE INDEX IF NOT EXISTS idx_task_activities_task ON task_activities(task_id);
CREATE INDEX IF NOT EXISTS idx_task_activities_type ON task_activities(activity_type);
CREATE INDEX IF NOT EXISTS idx_task_activities_timestamp ON task_activities(timestamp);

-- ============================================================================
-- Initial Migration Record
-- ============================================================================

-- Note: Migration record is inserted by applyMigration() function, not in SQL

