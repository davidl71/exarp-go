-- 016_recreate_tasks_with_enum_and_timeints.sql — Radical: add integer enums + int timestamps, enforce consistency.
--
-- This migration recreates the tasks table to add:
-- - status_enum / priority_enum (INTEGER): internal enum-first filtering and indexing
-- - created_ts / last_modified_ts / completed_at_ts (INTEGER): fast sorting/range queries; aligns with WKT Timestamp usage
-- - CHECK constraints to keep legacy string fields consistent with enums (compat window)
--
-- Note: SQLite cannot add CHECK constraints via ALTER TABLE; table recreation is the most compatible approach.

-- Step 1: Create new tasks table with additional columns and constraints
CREATE TABLE tasks_new (
    id TEXT PRIMARY KEY,

    -- Compatibility fields (kept during client compatibility window)
    name TEXT,
    content TEXT,
    long_description TEXT,

    status TEXT NOT NULL DEFAULT 'Todo',
    status_enum INTEGER NOT NULL DEFAULT 1,

    priority TEXT NOT NULL DEFAULT '',
    priority_enum INTEGER NOT NULL DEFAULT 0,

    completed INTEGER DEFAULT 0,
    task_number INTEGER,
    estimated_hours REAL,
    actual_hours REAL,

    -- Human-readable timestamps (compat / display)
    created TEXT NOT NULL,
    last_modified TEXT NOT NULL DEFAULT '',
    completed_at TEXT NOT NULL DEFAULT '',

    -- Fast query timestamps (unix seconds)
    created_ts INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    last_modified_ts INTEGER NOT NULL DEFAULT 0,
    completed_at_ts INTEGER NOT NULL DEFAULT 0,

    project_id TEXT,

    -- Flexible metadata (JSON + optional protobuf blob)
    metadata TEXT,
    metadata_protobuf BLOB,
    metadata_format TEXT NOT NULL DEFAULT 'json',

    -- Optimistic locking + internal timestamps
    version INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),

    -- Hierarchy + distributed tracking
    parent_id TEXT NOT NULL DEFAULT '',
    assigned_to TEXT NOT NULL DEFAULT '',
    host TEXT NOT NULL DEFAULT '',
    agent TEXT NOT NULL DEFAULT '',

    -- Locking / leases (agent claim fields)
    assignee TEXT NOT NULL DEFAULT '',
    assigned_at INTEGER NOT NULL DEFAULT 0,
    lock_until INTEGER NOT NULL DEFAULT 0,

    -- Enforce canonical sets (no "enum drift")
    CHECK (status IN ('Todo', 'In Progress', 'Review', 'Done', 'Blocked', 'Cancelled')),
    CHECK (status_enum IN (1, 2, 3, 4, 5, 6)),
    CHECK (
      status_enum = CASE status
        WHEN 'Todo' THEN 1
        WHEN 'In Progress' THEN 2
        WHEN 'Review' THEN 3
        WHEN 'Done' THEN 4
        WHEN 'Blocked' THEN 5
        WHEN 'Cancelled' THEN 6
        ELSE 0
      END
    ),

    CHECK (priority IN ('', 'low', 'medium', 'high', 'critical')),
    CHECK (priority_enum IN (0, 1, 2, 3, 4)),
    CHECK (
      priority_enum = CASE priority
        WHEN '' THEN 0
        WHEN 'low' THEN 1
        WHEN 'medium' THEN 2
        WHEN 'high' THEN 3
        WHEN 'critical' THEN 4
        ELSE 0
      END
    )
);

-- Step 2: Copy + backfill from old tasks table
INSERT INTO tasks_new (
  id,
  name,
  content,
  long_description,
  status,
  status_enum,
  priority,
  priority_enum,
  completed,
  task_number,
  estimated_hours,
  actual_hours,
  created,
  last_modified,
  completed_at,
  created_ts,
  last_modified_ts,
  completed_at_ts,
  project_id,
  metadata,
  metadata_protobuf,
  metadata_format,
  version,
  created_at,
  updated_at,
  parent_id,
  assigned_to,
  host,
  agent,
  assignee,
  assigned_at,
  lock_until
)
SELECT
  id,
  name,
  content,
  long_description,

  -- status: keep as-is (should already be canonical Title Case)
  status,
  CASE status
    WHEN 'Todo' THEN 1
    WHEN 'In Progress' THEN 2
    WHEN 'Review' THEN 3
    WHEN 'Done' THEN 4
    WHEN 'Blocked' THEN 5
    WHEN 'Cancelled' THEN 6
    ELSE 1
  END AS status_enum,

  COALESCE(priority, ''),
  CASE COALESCE(priority, '')
    WHEN '' THEN 0
    WHEN 'low' THEN 1
    WHEN 'medium' THEN 2
    WHEN 'high' THEN 3
    WHEN 'critical' THEN 4
    ELSE 0
  END AS priority_enum,

  completed,
  task_number,
  estimated_hours,
  actual_hours,

  created,
  COALESCE(last_modified, ''),
  COALESCE(completed_at, ''),

  -- created_ts: prefer parsing created; fall back to created_at
  CASE
    WHEN created IS NULL OR created = '' THEN COALESCE(created_at, CAST(strftime('%s', 'now') AS INTEGER))
    ELSE CAST(strftime('%s', created) AS INTEGER)
  END AS created_ts,

  CASE
    WHEN last_modified IS NULL OR last_modified = '' THEN 0
    ELSE CAST(strftime('%s', last_modified) AS INTEGER)
  END AS last_modified_ts,

  CASE
    WHEN completed_at IS NULL OR completed_at = '' THEN 0
    ELSE CAST(strftime('%s', completed_at) AS INTEGER)
  END AS completed_at_ts,

  project_id,
  metadata,
  metadata_protobuf,
  COALESCE(metadata_format, 'json'),
  version,
  created_at,
  updated_at,
  COALESCE(parent_id, ''),
  COALESCE(assigned_to, ''),
  COALESCE(host, ''),
  COALESCE(agent, ''),
  COALESCE(assignee, ''),
  COALESCE(assigned_at, 0),
  COALESCE(lock_until, 0)
FROM tasks;

-- Step 3: Swap tables
DROP TABLE tasks;
ALTER TABLE tasks_new RENAME TO tasks;

-- Step 4: Recreate indexes (enum-first)
CREATE INDEX IF NOT EXISTS idx_tasks_status_enum ON tasks(status_enum);
CREATE INDEX IF NOT EXISTS idx_tasks_priority_enum ON tasks(priority_enum);
CREATE INDEX IF NOT EXISTS idx_tasks_created_ts ON tasks(created_ts);
CREATE INDEX IF NOT EXISTS idx_tasks_completed ON tasks(completed);
CREATE INDEX IF NOT EXISTS idx_tasks_last_modified_ts ON tasks(last_modified_ts);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_metadata_format ON tasks(metadata_format);

CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX IF NOT EXISTS idx_tasks_host ON tasks(host);
CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent);

-- Locking indexes (enum-first)
CREATE INDEX IF NOT EXISTS idx_tasks_id_version ON tasks(id, version);
CREATE INDEX IF NOT EXISTS idx_tasks_status_enum_lock ON tasks(status_enum, lock_until);
CREATE INDEX IF NOT EXISTS idx_tasks_status_enum_assignee ON tasks(status_enum, assignee);
CREATE INDEX IF NOT EXISTS idx_tasks_lock_until ON tasks(lock_until) WHERE lock_until > 0;

-- Backlog hot-path partial index (Todo + In Progress)
CREATE INDEX IF NOT EXISTS idx_tasks_backlog_hot
  ON tasks(priority_enum, created_ts)
  WHERE status_enum IN (1, 2);

