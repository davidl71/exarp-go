# Task-Related Code Simplification Analysis

**Task:** T-1768321828916  
**Date:** 2026-02-02

## 1. Overview

Analysis of task-related code in `handlers.go`, `task_workflow_common.go`, `task_workflow_native.go`, `cli/task.go`, `database/tasks.go`, and related files to identify simplification opportunities.

---

## 2. Duplicate Code Patterns

### 2.1 Database-then-JSON Fallback (Repeated ~26 files)

**Pattern:** Many handlers use:

```go
db, dbErr := database.GetDB()
useDB := dbErr == nil && db != nil
if useDB {
    // database path
} else {
    // JSON/file path
}
```

**Locations:** `task_workflow_native.go` (clarify resolve, sync_approvals, apply_approval_result), `task_workflow_common.go` (approve, link_planning, create), `plan_sync.go`, `infer_task_progress.go`.

**Recommendation:** Extract `WithTaskStore(ctx, projectRoot, fn func(store TaskStore) error)` or a simpler `GetTaskByID(ctx, projectRoot, id)` that encapsulates DB-first, JSON-fallback logic. Consolidate in `todo2_utils.go`.

### 2.2 Load-Search-Update-Save for Single Task

**Pattern:** Load all tasks, find by ID, update, save:

```go
tasks, err := LoadTodo2Tasks(projectRoot)
// ... find task ...
tasks[i].Status = newStatus
SaveTodo2Tasks(projectRoot, tasks)
```

vs. direct DB path:

```go
task, err := database.GetTask(ctx, taskID)
task.Status = newStatus
database.UpdateTask(ctx, task)
```

**Recommendation:** Add `UpdateTaskStatus(ctx, projectRoot, taskID, newStatus)` in `todo2_utils.go` that uses DB when available, else LoadTodo2Tasks → update in slice → SaveTodo2Tasks. Reduces ~10 call sites to one helper.

### 2.3 Task ID Parsing (task_id vs task_ids)

**Pattern:** Multiple handlers parse `task_id` (string) and `task_ids` (comma-separated or JSON array):

```go
var ids []string
if taskID, ok := params["task_id"].(string); ok && taskID != "" {
    ids = []string{taskID}
} else if idsRaw, ok := params["task_ids"]; ok {
    // parse comma or JSON array
}
```

**Locations:** `task_workflow_common.go` (approve, link_planning), `task_workflow_native.go` (delete, apply_approval_result).

**Recommendation:** Extract `ParseTaskIDsFromParams(params map[string]interface{}) ([]string, error)` in `task_workflow_common.go` or a shared params helper. Reuse in all handlers.

---

## 3. Database/File Fallback Patterns

### 3.1 Current Abstraction (todo2_utils.go)

| Function | Behavior |
|----------|----------|
| `LoadTodo2Tasks(projectRoot)` | DB first; on failure, JSON |
| `SaveTodo2Tasks(projectRoot, tasks)` | DB first; on success, also write JSON (both in sync) |
| `SyncTodo2Tasks(projectRoot)` | Load both, merge (DB wins), save to both |

**Observation:** SaveTodo2Tasks already keeps both in sync (per TASK_WORKFLOW_ABSTRACTION_ANALYSIS). The pattern is consistent.

### 3.2 Simplification Opportunity

Handlers that do "get task → update → sync" can be simplified:

- **Before:** LoadTodo2Tasks, loop to find, mutate, SaveTodo2Tasks, optionally SyncTodo2Tasks.
- **After:** Prefer `database.GetTask` + `database.UpdateTask` when DB available; otherwise a single `LoadTodo2Tasks` → mutate → `SaveTodo2Tasks` path. Avoid duplicated "try DB, fallback file" in each handler.

---

## 4. Protobuf/JSON Handling

### 4.1 Current State

- **database/tasks.go:** CreateTask/UpdateTask support both protobuf and JSON metadata; fallback on "no such column" for older schema.
- **todo2_json.go:** `ParseTasksFromJSON`, `MarshalTasksToStateJSON` handle JSON format; metadata sanitization for invalid JSON.
- **Models:** `Todo2Task` uses `map[string]interface{}` for Metadata; protobuf serialization in `models` package.

### 4.2 Consolidation Opportunities

1. **Metadata normalization:** Several places convert Metadata to/from JSON for DB. Consider a single `SerializeTaskMetadata(task)` / `DeserializeTaskMetadata(raw)` that handles both protobuf and JSON.
2. **Schema version check:** The CreateTask fallback checks "no such column" to detect old schema. A helper `HasProtobufColumns(db)` could make this explicit and reusable.

---

## 5. Refactoring Recommendations

### 5.1 High Impact, Low Risk

1. **Extract `ParseTaskIDsFromParams`**  
   - File: `internal/tools/task_workflow_common.go`  
   - Use in: approve, link_planning, delete, apply_approval_result  
   - ~50 lines eliminated across handlers

2. **Add `UpdateTaskStatus(ctx, projectRoot, taskID, newStatus)`**  
   - File: `internal/tools/todo2_utils.go`  
   - Encapsulates DB vs file update logic  
   - Use in: apply_approval_result, approve (single-task path), clarify resolve

### 5.2 Medium Impact

3. **Consolidate "get task by ID" logic**  
   - Single `GetTaskByID(ctx, projectRoot, id) (*Todo2Task, error)`  
   - Use in: request_approval, apply_approval_result, resolveTaskClarification

4. **Unify metadata serialization**  
   - `SerializeTaskMetadata` / `DeserializeTaskMetadata` in `todo2_json.go` or `models`  
   - Reduces repeated marshaling in database/tasks.go and todo2_json.go

### 5.3 Lower Priority

5. **Extract TaskStore interface**  
   - `GetTask(id)`, `UpdateTask(task)`, `ListTasks(filters)`  
   - Would allow testing with mock store; larger refactor

---

## 6. Files to Modify (Summary)

| File | Changes |
|------|---------|
| `internal/tools/task_workflow_common.go` | Add ParseTaskIDsFromParams; use in approve, link_planning |
| `internal/tools/task_workflow_native.go` | Use ParseTaskIDsFromParams in delete; use GetTaskByID in request_approval, apply_approval_result |
| `internal/tools/todo2_utils.go` | Add UpdateTaskStatus, GetTaskByID |
| `internal/tools/plan_sync.go` | Use GetTaskByID if applicable (minor) |
| `internal/tools/infer_task_progress.go` | Use UpdateTaskStatus for apply path |

---

## 7. References

- [TASK_WORKFLOW_ABSTRACTION_ANALYSIS.md](TASK_WORKFLOW_ABSTRACTION_ANALYSIS.md) — SaveTodo2Tasks / SyncTodo2Tasks behavior
- [internal/tools/todo2_utils.go](../internal/tools/todo2_utils.go) — LoadTodo2Tasks, SaveTodo2Tasks, SyncTodo2Tasks
- [internal/database/tasks.go](../internal/database/tasks.go) — CreateTask, UpdateTask, GetTask
