# Database Patterns: New vs Old

## Overview

This document describes the database access patterns in exarp-go, distinguishing between the modern sqlx-based approach and legacy patterns.

## Patterns Comparison

| Aspect | Old Pattern | New Pattern |
|--------|-------------|-------------|
| **DB Access** | `database.GetDB()` → `*sql.DB` | `database.GetDBx()` → `*sqlx.DB` |
| **Query Method** | `db.Query()`, `db.Exec()` | `db.Select()`, `db.Get()` |
| **Scanning** | Manual `rows.Scan()` | Struct-based with `sqlx.StructScan()` |
| **Schema Fallback** | Yes (old schema detection) | No (schema 9 assumed) |
| **File Fallback** | Yes (DB + JSON) | Deprecated (DB-only) |

## File Usage Summary

### ✅ Modern (sqlx) - `internal/database/`

All functions in `internal/database/` use `GetDBx()`:

| File | Functions |
|------|-----------|
| **tasks_crud.go** | CreateTask, GetTask, UpdateTask, BatchUpdateTaskStatus, BatchUpdateTaskMetadata, UpdateTaskFields, DeleteTask, IsVersionMismatchError, CheckUpdateConflict |
| **tasks_list.go** | ListTasks, GetDoneTasksForEstimation, GetTasksByStatus, GetTaskCountByStatus, GetTasksByTag, GetTasksByPriority, FindNextClaimableTask |
| **tag_cache.go** | GetDiscoveredTagsForFile, GetDiscoveredTagsWithHash, SaveDiscoveredTags, UpdateTagFrequency, GetTagFrequencies, SaveFileTaskTag, ClearTaskTagSuggestions, ClearFileTaskTags, GetTaskTagSuggestions, GetTopTagFrequencies |
| **tasks_misc.go** | FixTaskDates, GetDependencies, GetDependents |
| **tasks_lock.go** | ClaimTaskForAgent, ReleaseTask, RenewLease, RunLeaseRenewal, CleanupExpiredLocks |
| **comments.go** | AddComments, GetComments, GetCommentsByType, DeleteComment, GetCommentsWithTypeFilter |
| **lock_monitoring.go** | DetectStaleLocks, GetLockStatus, GetLocksByAgent, CleanupExpiredLocksWithReport, CleanupDeadAgentLocks |

### ⚠️ Legacy (GetDB) - `internal/tools/`

These functions use `GetDB()` for dual DB/file fallback (intentional):

| File | Functions | Purpose |
|------|-----------|---------|
| **task_store.go** | TaskStoreDBExists, TaskStoreDBList, TaskStoreDBCreate, TaskStoreDBUpdate, TaskStoreDBDelete | TaskStore operations with JSON fallback |
| **todo2_db_adapter.go** | loadTodo2TasksFromDB, loadTodo2TasksFromJSON, SaveTodo2TasksToDB, GetDBForTodo2DBAdapter | DB ↔ JSON sync adapter |
| **task_workflow_actions.go** | getTaskStatusFromDB, getTaskMetadataFromDB, batchStatusUpdateFromDB | Task workflow handlers |
| **task_workflow_maintenance.go** | GetTaskForMaintenance, GetTasksWithMetadataForMaintenance | Maintenance with file backup |
| **session.go** | getSessionFromDB | Session state with file fallback |
| **todo2_utils.go** | getNextTasksFromDB | Legacy utility functions |
| **task_workflow_maintenance_helpers.go** | getMaintenanceTaskIDsFromDB | Maintenance helpers |
| **session_assignee.go** | getAgentTasksFromDB | Assignee tracking |
| **automation_discover.go** | checkAutomationStoreExists | Discovery with file fallback |
| **resources/tasks.go** | getTaskForResource | Task resources |

## Migration Status

### Completed (2026-03-24)

- [x] `GetTask` - sqlx with `taskRow` struct
- [x] `CreateTask` - sqlx, removed legacy fallbacks
- [x] `UpdateTask` - sqlx, removed legacy fallbacks
- [x] `ListTasks` - sqlx, removed legacy fallbacks
- [x] `GetDoneTasksForEstimation` - sqlx
- [x] `FindNextClaimableTask` - sqlx

### Remaining Legacy Usage

The `GetDB()` usage in `internal/tools/` is **intentional** for files that need:
1. JSON file fallback (when DB unavailable)
2. Sync between DB and JSON representations
3. Backward compatibility with existing code

These files are candidates for future migration to use TaskStore interface.

## Schema Version

- **Current**: Schema 9
- **Assumed**: Full schema (protobuf + distributed tracking columns)
- **Fallback removed**: No runtime schema detection needed

## Code Examples

### New Pattern (sqlx)

```go
// GetDBx returns *sqlx.DB
db, err := GetDBx()
if err != nil {
    return err
}

// Using sqlx.GetContext for single row
var task taskRow
err = db.GetContext(ctx, &task, `SELECT id, content, status, ... FROM tasks WHERE id = ?`, id)

// Using sqlx.Select for multiple rows
var tasks []taskRow
err = db.SelectContext(ctx, &tasks, `SELECT ... FROM tasks WHERE status = ?`, status)
```

### Old Pattern (legacy)

```go
// GetDB returns *sql.DB
db, err := GetDB()
if err != nil {
    return err
}

// Manual scanning
rows, err := db.Query(`SELECT ... FROM tasks WHERE ...`)
for rows.Next() {
    var task Todo2Task
    err := rows.Scan(&task.ID, &task.Content, ...)
}
```

## Recommendations

1. **New code**: Always use `GetDBx()` from `internal/database/`
2. **Tools package**: Consider using TaskStore interface instead of direct DB access
3. **JSON fallback**: Deprecated for new features; existing usage can remain
4. **Testing**: Use `database.GetDBx()` for consistent test patterns