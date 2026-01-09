# SQLite Migration Plan for Todo2 Tasks

**Status:** Planning Phase  
**Created:** 2026-01-09  
**Target:** Migrate Todo2 task storage from JSON files to SQLite database

## Executive Summary

This document outlines the migration plan to replace JSON-based Todo2 task storage with SQLite, eliminating file locking complexity, improving query performance, and simplifying the codebase by removing ~2000+ lines of file I/O code.

## Current State Analysis

### Current Storage
- **Location:** `.todo2/state.todo2.json`
- **Size:** ~1580+ lines, ~95 tasks
- **Operations:** Load entire file for any operation
- **Concurrency:** Manual file locking (`file_lock.py`, 239 lines)
- **Caching:** Complex JSON cache manager (464 lines)
- **Queries:** In-memory filtering after loading entire file

### Current Task Structure

Based on analysis of `.todo2/state.todo2.json`, tasks contain:

**Core Fields:**
- `id` (string) - Unique identifier
- `name` (string, optional) - Task name
- `content` (string, optional) - Short description
- `long_description` (string, optional) - Detailed description
- `status` (string) - Task status (Todo, In Progress, Review, Done, Cancelled)
- `priority` (string, optional) - Priority level (high, medium, low, critical)
- `tags` (array of strings) - Task tags
- `dependencies` (array of strings) - Task IDs this depends on
- `completed` (boolean, optional) - Completion flag

**Timestamps:**
- `created` (ISO 8601 string) - Creation timestamp
- `lastModified` (ISO 8601 string, optional) - Last modification timestamp
- `completedAt` (ISO 8601 string, optional) - Completion timestamp

**Metadata:**
- `taskNumber` (integer, optional) - Task number
- `estimatedHours` (float, optional) - Estimated duration
- `actualHours` (float, optional) - Actual duration
- `changes` (array of change objects) - Change history
- `metadata` (object, optional) - Additional metadata

**Change Object Structure:**
- `field` (string) - Field name that changed
- `oldValue` (any) - Previous value
- `newValue` (any) - New value
- `timestamp` (ISO 8601 string) - Change timestamp

## Database Schema Design

### Main Tables

#### 1. `tasks` Table

Primary table for task data.

```sql
CREATE TABLE tasks (
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
    project_id TEXT,  -- Project identifier (e.g., "davidl71/exarp-go")
    metadata TEXT,  -- JSON blob for flexible metadata
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Indexes for common queries
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_created ON tasks(created_at);
CREATE INDEX idx_tasks_completed ON tasks(completed);
CREATE INDEX idx_tasks_last_modified ON tasks(last_modified);
CREATE INDEX idx_tasks_project ON tasks(project_id);
```

#### 2. `task_tags` Table

Many-to-many relationship for tags.

```sql
CREATE TABLE task_tags (
    task_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (task_id, tag),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_task_tags_task ON task_tags(task_id);
CREATE INDEX idx_task_tags_tag ON task_tags(tag);
```

#### 3. `task_dependencies` Table

Many-to-many relationship for dependencies.

```sql
CREATE TABLE task_dependencies (
    task_id TEXT NOT NULL,
    depends_on_id TEXT NOT NULL,
    PRIMARY KEY (task_id, depends_on_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CHECK (task_id != depends_on_id)  -- Prevent self-dependencies
);

CREATE INDEX idx_task_deps_task ON task_dependencies(task_id);
CREATE INDEX idx_task_deps_depends ON task_dependencies(depends_on_id);
```

#### 4. `task_changes` Table

Change history tracking.

```sql
CREATE TABLE task_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    field TEXT NOT NULL,
    old_value TEXT,  -- JSON-encoded value
    new_value TEXT,  -- JSON-encoded value
    timestamp TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_task_changes_task ON task_changes(task_id);
CREATE INDEX idx_task_changes_timestamp ON task_changes(timestamp);
```

#### 5. `task_comments` Table

Task comments (stored both per-task and at root level in JSON).

```sql
CREATE TABLE task_comments (
    id TEXT PRIMARY KEY,  -- Comment ID (e.g., "T-NaN-C-1")
    task_id TEXT NOT NULL,  -- todoId in JSON
    comment_type TEXT NOT NULL,  -- 'research_with_links', 'result', 'note', 'manualsetup'
    content TEXT NOT NULL,
    created TEXT NOT NULL,  -- ISO 8601 timestamp
    last_modified TEXT,  -- ISO 8601 timestamp
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_task_comments_task ON task_comments(task_id);
CREATE INDEX idx_task_comments_type ON task_comments(comment_type);
```

#### 6. `task_activities` Table

Activity log for task changes (from root-level `activities` array).

```sql
CREATE TABLE task_activities (
    id TEXT PRIMARY KEY,  -- Activity ID (e.g., "activity-1767912853265-mv2lrmp39")
    activity_type TEXT NOT NULL,  -- 'todo_created', 'comment_added', 'status_changed', etc.
    task_id TEXT NOT NULL,  -- todoId
    task_name TEXT,  -- todoName
    timestamp TEXT NOT NULL,  -- ISO 8601 timestamp
    details TEXT,  -- JSON blob for activity-specific details
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_task_activities_task ON task_activities(task_id);
CREATE INDEX idx_task_activities_type ON task_activities(activity_type);
CREATE INDEX idx_task_activities_timestamp ON task_activities(timestamp);
```

#### 7. `schema_migrations` Table

Track database schema versions.

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    description TEXT
);
```

### Full Schema SQL

See `migrations/001_initial_schema.sql` for complete schema definition.

## Migration Strategy

### Phase 1: Database Infrastructure (Task: sqlite-migration-2)

**Goal:** Create database package with schema and migration support.

**Deliverables:**
1. `internal/database/sqlite.go` - Database connection and initialization
2. `internal/database/schema.go` - Schema definitions
3. `internal/database/migrations.go` - Migration system
4. `migrations/001_initial_schema.sql` - Initial schema SQL

**Key Features:**
- Automatic database file creation in `.todo2/todo2.db`
- Migration version tracking
- Schema validation
- Connection pooling
- Transaction support

### Phase 2: CRUD Operations (Task: sqlite-migration-3)

**Goal:** Implement task operations using SQLite.

**Deliverables:**
1. `internal/database/tasks.go` - Task CRUD operations
2. Update `internal/tools/todo2_utils.go` - Use database instead of JSON
3. Update `internal/tools/task_workflow_native.go` - Use database

**Operations to Implement:**
- `CreateTask(task *Todo2Task) error`
- `GetTask(id string) (*Todo2Task, error)`
- `UpdateTask(task *Todo2Task) error`
- `DeleteTask(id string) error`
- `ListTasks(filters *TaskFilters) ([]*Todo2Task, error)`
- `GetTasksByStatus(status string) ([]*Todo2Task, error)`
- `GetTasksByTag(tag string) ([]*Todo2Task, error)`
- `GetTasksByPriority(priority string) ([]*Todo2Task, error)`

**Query Helpers:**
- `GetDependencies(taskID string) ([]string, error)`
- `GetDependents(taskID string) ([]string, error)`
- `GetTaskHistory(taskID string) ([]*TaskChange, error)`

### Phase 3: Data Migration (Task: sqlite-migration-4)

**Goal:** Migrate existing JSON data to SQLite.

**Deliverables:**
1. `internal/database/migrate_json.go` - JSON to SQLite migration
2. `cmd/migrate-todo2/main.go` - Migration CLI tool
3. Update `Makefile` - Add migration target

**Migration Process:**
1. **Backup:** Create backup of `.todo2/state.todo2.json`
2. **Validate:** Verify JSON structure is valid
3. **Migrate:** Convert all tasks to SQLite format
4. **Validate:** Compare counts and verify data integrity
5. **Rollback:** Support rollback if migration fails

**Migration Tool Features:**
- Dry-run mode (validate without migrating)
- Progress reporting
- Data validation
- Rollback support
- Backup creation

### Phase 4: Code Simplification (Task: sqlite-migration-5)

**Goal:** Remove file locking and JSON cache complexity.

**Deliverables:**
1. Remove `task_lock()` calls from task operations
2. Remove JSON cache usage for tasks
3. Update error handling to use database errors
4. Document changes in code comments

**Files to Update:**
- `internal/tools/todo2_utils.go` - Remove file I/O
- `internal/tools/task_workflow_native.go` - Use transactions
- `project_management_automation/utils/file_lock.py` - Document task locking removal
- `project_management_automation/utils/json_cache.py` - Remove task caching

## Implementation Details

### Database Location

- **Path:** `.todo2/todo2.db`
- **Backup:** `.todo2/todo2.db.backup` (before migration)
- **JSON Backup:** `.todo2/state.todo2.json.backup` (before migration)

### Go Dependencies

```go
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)
```

**Alternative:** Consider `modernc.org/sqlite` for pure Go (no CGO):
```go
import _ "modernc.org/sqlite"
```

### Transaction Handling

Use database transactions for atomic operations:

```go
func UpdateTaskWithHistory(task *Todo2Task, changes []TaskChange) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Update task
    if err := updateTaskTx(tx, task); err != nil {
        return err
    }
    
    // Insert changes
    for _, change := range changes {
        if err := insertChangeTx(tx, task.ID, change); err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

### Query Optimization

**Use Prepared Statements:**
```go
var getTaskStmt *sql.Stmt

func init() {
    getTaskStmt = db.Prepare(`
        SELECT id, name, content, long_description, status, priority, 
               completed, task_number, estimated_hours, actual_hours,
               created, last_modified, completed_at, metadata
        FROM tasks WHERE id = ?
    `)
}
```

**Use Indexes:**
- Status queries: `idx_tasks_status`
- Tag queries: `idx_task_tags_tag`
- Dependency queries: `idx_task_deps_task`
- Time-based queries: `idx_tasks_created`, `idx_tasks_last_modified`

### Error Handling

Handle SQLite-specific errors:

```go
import "github.com/mattn/go-sqlite3"

func handleDBError(err error) error {
    if sqliteErr, ok := err.(sqlite3.Error); ok {
        switch sqliteErr.Code {
        case sqlite3.ErrConstraint:
            return fmt.Errorf("constraint violation: %w", err)
        case sqlite3.ErrBusy:
            return fmt.Errorf("database busy, retry: %w", err)
        }
    }
    return err
}
```

## Backward Compatibility

### Dual-Mode Support (Optional)

During transition, support both JSON and SQLite:

```go
type TaskStore interface {
    LoadTasks() ([]*Todo2Task, error)
    SaveTask(task *Todo2Task) error
}

type JSONTaskStore struct { /* ... */ }
type SQLiteTaskStore struct { /* ... */ }

// Factory based on config
func NewTaskStore(useDB bool) TaskStore {
    if useDB {
        return NewSQLiteTaskStore()
    }
    return NewJSONTaskStore()
}
```

### Migration Rollback

If migration fails, restore from backup:

```bash
# Restore JSON backup
cp .todo2/state.todo2.json.backup .todo2/state.todo2.json

# Remove database
rm .todo2/todo2.db
```

## Testing Strategy

### Unit Tests

- Test all CRUD operations
- Test query methods
- Test transaction handling
- Test error cases

### Integration Tests

- Test migration from JSON
- Test concurrent access
- Test data integrity
- Test performance

### Performance Tests

- Compare query performance vs JSON
- Test with large datasets (1000+ tasks)
- Test concurrent operations
- Benchmark common queries

## Performance Expectations

### Query Performance

**Before (JSON):**
- Load all tasks: O(n) file read + O(n) JSON parse
- Filter by status: O(n) in-memory scan
- Find by ID: O(n) in-memory scan

**After (SQLite):**
- Load all tasks: O(n) with streaming
- Filter by status: O(log n) with index
- Find by ID: O(1) primary key lookup

### Concurrency

**Before:**
- Manual file locking required
- Lock contention issues
- Complex locking code

**After:**
- ACID transactions handle concurrency
- Automatic locking by SQLite
- No manual lock management

## Migration Checklist

### Pre-Migration

- [ ] Backup `.todo2/state.todo2.json`
- [ ] Verify JSON structure is valid
- [ ] Count tasks in JSON file
- [ ] Test database schema creation
- [ ] Review migration plan

### Migration

- [ ] Run migration tool (dry-run first)
- [ ] Verify all tasks migrated
- [ ] Validate data integrity
- [ ] Test query operations
- [ ] Test concurrent access

### Post-Migration

- [ ] Update all code to use database
- [ ] Remove file locking code
- [ ] Remove JSON cache for tasks
- [ ] Update tests
- [ ] Update documentation
- [ ] Monitor for issues

### Rollback (if needed)

- [ ] Restore JSON backup
- [ ] Remove database file
- [ ] Revert code changes
- [ ] Verify system works with JSON

## Future Enhancements

### Phase 5: Additional Features

1. **Full-Text Search:** Use SQLite FTS5 for task content search
2. **Analytics:** Add query methods for task statistics
3. **Relationships:** Better dependency graph queries
4. **History:** Enhanced change tracking and audit trail

### Phase 6: Other Data Migration

1. **Memories:** Migrate `.exarp/memories/*.json` to SQLite
2. **Commits:** Migrate `.todo2/commits.json` to SQLite
3. **Session State:** Consolidate session/workflow state

## Risks and Mitigation

### Risk 1: Data Loss During Migration

**Mitigation:**
- Create backups before migration
- Validate data after migration
- Support rollback mechanism
- Test migration on copy first

### Risk 2: Performance Regression

**Mitigation:**
- Benchmark before/after
- Use indexes appropriately
- Optimize queries
- Monitor performance

### Risk 3: Breaking Changes

**Mitigation:**
- Maintain backward compatibility during transition
- Gradual migration approach
- Comprehensive testing
- Clear documentation

### Risk 4: Concurrent Access Issues

**Mitigation:**
- Use SQLite WAL mode for better concurrency
- Test concurrent operations
- Handle database locked errors gracefully
- Use appropriate transaction isolation

## Success Metrics

### Code Simplification

- **Target:** Remove 2000+ lines of file I/O code
- **Measure:** Lines of code removed
- **Files:** `file_lock.py`, `json_cache.py` (task parts), `todo2_utils.go`

### Performance Improvement

- **Target:** 10x faster queries for common operations
- **Measure:** Benchmark query times
- **Operations:** Get by status, Get by ID, Filter by tags

### Reliability Improvement

- **Target:** Zero data loss, zero corruption
- **Measure:** Migration success rate, data integrity checks
- **Metrics:** Backup/restore success, validation passes

## Timeline Estimate

### Original Estimates
- **Phase 1 (Research & Schema):** 2-3 days → ✅ **1 hour (ACTUAL)**
- **Phase 2 (Database Infrastructure):** 4-6 days → **1-1.5 days (ADJUSTED)**
- **Phase 3 (CRUD Operations):** 10-15 days → **3-4 days (ADJUSTED)**
- **Phase 4 (Data Migration):** 4-6 days → **1.5-2 days (ADJUSTED)**
- **Phase 5 (Code Simplification):** 2-3 days → **1 day (ADJUSTED)**

**Original Total:** ~22-35 days (4-6 weeks)  
**Adjusted Total:** **5.5-7 days (1-1.5 weeks)**  
**Reduction:** 68-74% faster

### Adjusted Estimates (Post Task 1)

Based on Task 1 learnings (1 hour actual vs 1-1.5 days estimated):
- Well-defined tasks are 80-90% faster than estimated
- Code generation reduces time by 50-70%
- Tool estimates are more accurate for well-defined tasks

**Recommended Timeline:** **5.5-7 days (1-1.5 weeks)**

*See `docs/SQLITE_ESTIMATION_ADJUSTED.md` for detailed adjusted estimates.*

## References

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [Go database/sql Tutorial](https://go.dev/doc/database/)
- [go-sqlite3 Documentation](https://github.com/mattn/go-sqlite3)
- [SQLite Best Practices](https://www.sqlite.org/performance.html)

## Appendix: Example Queries

### Get All Tasks by Status

```sql
SELECT t.*, 
       GROUP_CONCAT(tt.tag) as tags,
       GROUP_CONCAT(td.depends_on_id) as dependencies
FROM tasks t
LEFT JOIN task_tags tt ON t.id = tt.task_id
LEFT JOIN task_dependencies td ON t.id = td.task_id
WHERE t.status = ?
GROUP BY t.id
ORDER BY t.created_at DESC;
```

### Get Task with Full Details

```sql
SELECT 
    t.*,
    GROUP_CONCAT(DISTINCT tt.tag) as tags,
    GROUP_CONCAT(DISTINCT td.depends_on_id) as dependencies,
    (SELECT COUNT(*) FROM task_dependencies WHERE depends_on_id = t.id) as dependents_count
FROM tasks t
LEFT JOIN task_tags tt ON t.id = tt.task_id
LEFT JOIN task_dependencies td ON t.id = td.task_id
WHERE t.id = ?
GROUP BY t.id;
```

### Get Task History

```sql
SELECT field, old_value, new_value, timestamp
FROM task_changes
WHERE task_id = ?
ORDER BY timestamp DESC;
```

### Search Tasks by Tag

```sql
SELECT DISTINCT t.*
FROM tasks t
INNER JOIN task_tags tt ON t.id = tt.task_id
WHERE tt.tag = ?
ORDER BY t.created_at DESC;
```

