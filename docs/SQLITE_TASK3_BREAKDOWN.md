# Task 3: CRUD Operations - Parallel Subtask Breakdown

**Parent Task:** sqlite-migration-3  
**Status:** Planning  
**Estimated Time (Sequential):** 5-6 days  
**Estimated Time (Parallel):** 3-4 days  
**Time Savings:** 2-3 days

## Overview

Task 3 involves implementing all Todo2 task CRUD operations using SQLite. This task has the **highest parallelization potential** with 20+ independent functions that can be implemented simultaneously.

## Subtask Groups

### Group 1: Core CRUD Operations (4 subtasks)

**Dependencies:** sqlite-migration-2 (Database Infrastructure)  
**Can Run In Parallel:** ✅ Yes (all independent)

#### T3.1: CreateTask Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `CreateTask(task *Todo2Task) error`
- **Complexity:** Medium
- **Estimated Time:** 0.5 days
- **Dependencies:** None (within group)
- **Description:**
  - Insert task into `tasks` table
  - Insert tags into `task_tags` table
  - Insert dependencies into `task_dependencies` table
  - Handle transaction rollback on error
  - Return task ID

#### T3.2: GetTask Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `GetTask(id string) (*Todo2Task, error)`
- **Complexity:** Medium
- **Estimated Time:** 0.5 days
- **Dependencies:** None (within group)
- **Description:**
  - Query task from `tasks` table
  - Join with `task_tags` to get tags
  - Join with `task_dependencies` to get dependencies
  - Parse metadata JSON
  - Return complete task struct

#### T3.3: UpdateTask Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `UpdateTask(task *Todo2Task) error`
- **Complexity:** High
- **Estimated Time:** 1 day
- **Dependencies:** None (within group)
- **Description:**
  - Update task in `tasks` table
  - Update tags (delete old, insert new)
  - Update dependencies (delete old, insert new)
  - Update `last_modified` timestamp
  - Handle transaction for atomicity

#### T3.4: DeleteTask Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `DeleteTask(id string) error`
- **Complexity:** Low
- **Estimated Time:** 0.25 days
- **Dependencies:** None (within group)
- **Description:**
  - Delete task (cascades to tags, dependencies via foreign keys)
  - Verify task exists before deletion
  - Return error if task not found

**Group 1 Total (Parallel):** 1 day (vs 2.25 days sequential)

---

### Group 2: Query Operations (6 subtasks)

**Dependencies:** sqlite-migration-2, T3.2 (GetTask for reference)  
**Can Run In Parallel:** ✅ Yes (all independent)

#### T3.5: ListTasks Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `ListTasks(filters *TaskFilters) ([]*Todo2Task, error)`
- **Complexity:** High
- **Estimated Time:** 1 day
- **Dependencies:** T3.2 (for task struct reference)
- **Description:**
  - Query all tasks with optional filters
  - Support filtering by status, priority, tags, project_id
  - Join with tags and dependencies
  - Support pagination (limit/offset)
  - Return sorted list

#### T3.6: GetTasksByStatus Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `GetTasksByStatus(status string) ([]*Todo2Task, error)`
- **Complexity:** Low
- **Estimated Time:** 0.25 days
- **Dependencies:** T3.5 (can reuse query logic)
- **Description:**
  - Query tasks filtered by status
  - Use index on `status` column
  - Return sorted by created_at

#### T3.7: GetTasksByTag Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `GetTasksByTag(tag string) ([]*Todo2Task, error)`
- **Complexity:** Medium
- **Estimated Time:** 0.5 days
- **Dependencies:** T3.5 (can reuse query logic)
- **Description:**
  - Join `tasks` with `task_tags` table
  - Filter by tag name
  - Use index on `task_tags.tag`
  - Return sorted list

#### T3.8: GetTasksByPriority Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `GetTasksByPriority(priority string) ([]*Todo2Task, error)`
- **Complexity:** Low
- **Estimated Time:** 0.25 days
- **Dependencies:** T3.5 (can reuse query logic)
- **Description:**
  - Query tasks filtered by priority
  - Use index on `priority` column
  - Return sorted list

#### T3.9: GetTasksByProject Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `GetTasksByProject(projectID string) ([]*Todo2Task, error)`
- **Complexity:** Low
- **Estimated Time:** 0.25 days
- **Dependencies:** T3.5 (can reuse query logic)
- **Description:**
  - Query tasks filtered by project_id
  - Use index on `project_id` column
  - Return sorted list

#### T3.10: SearchTasks Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `SearchTasks(query string) ([]*Todo2Task, error)`
- **Complexity:** Medium
- **Estimated Time:** 0.5 days
- **Dependencies:** T3.5 (can reuse query logic)
- **Description:**
  - Full-text search on task content/description
  - Use SQLite FTS5 (optional, can use LIKE for MVP)
  - Return ranked results

**Group 2 Total (Parallel):** 1.5 days (vs 2.75 days sequential)

---

### Group 3: Relationship Operations (4 subtasks)

**Dependencies:** sqlite-migration-2, T3.1 (CreateTask for reference)  
**Can Run In Parallel:** ✅ Yes (all independent)

#### T3.11: Tag Operations - Add/Remove Tags
- **File:** `internal/database/tasks.go`
- **Functions:** 
  - `AddTag(taskID, tag string) error`
  - `RemoveTag(taskID, tag string) error`
  - `GetTags(taskID string) ([]string, error)`
- **Complexity:** Low
- **Estimated Time:** 0.5 days
- **Dependencies:** None (within group)
- **Description:**
  - Insert/delete from `task_tags` table
  - Validate tag doesn't already exist (for add)
  - Validate tag exists (for remove)
  - Return all tags for a task

#### T3.12: Dependency Operations - Add/Remove Dependencies
- **File:** `internal/database/tasks.go`
- **Functions:**
  - `AddDependency(taskID, dependsOnID string) error`
  - `RemoveDependency(taskID, dependsOnID string) error`
  - `GetDependencies(taskID string) ([]string, error)`
- **Complexity:** Medium
- **Estimated Time:** 0.75 days
- **Dependencies:** None (within group)
- **Description:**
  - Insert/delete from `task_dependencies` table
  - Validate no circular dependencies
  - Validate both tasks exist
  - Prevent self-dependencies

#### T3.13: GetDependents Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `GetDependents(taskID string) ([]string, error)`
- **Complexity:** Low
- **Estimated Time:** 0.25 days
- **Dependencies:** T3.12 (for reference)
- **Description:**
  - Query reverse dependencies (tasks that depend on this task)
  - Use index on `depends_on_id` column
  - Return list of task IDs

#### T3.14: ValidateDependencyChain Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `ValidateDependencyChain(taskID string) (bool, error)`
- **Complexity:** High
- **Estimated Time:** 0.75 days
- **Dependencies:** T3.12, T3.13
- **Description:**
  - Check for circular dependencies
  - Traverse dependency graph
  - Return true if valid, false if circular
  - Return error if task not found

**Group 3 Total (Parallel):** 1.25 days (vs 2.25 days sequential)

---

### Group 4: History & Comments Operations (4 subtasks)

**Dependencies:** sqlite-migration-2  
**Can Run In Parallel:** ✅ Yes (all independent)

#### T3.15: Change Tracking - Record Changes
- **File:** `internal/database/tasks.go`
- **Function:** `RecordTaskChange(taskID, field, oldValue, newValue string) error`
- **Complexity:** Low
- **Estimated Time:** 0.25 days
- **Dependencies:** None (within group)
- **Description:**
  - Insert into `task_changes` table
  - Store JSON-encoded old/new values
  - Set timestamp automatically
  - Return change ID

#### T3.16: GetTaskHistory Implementation
- **File:** `internal/database/tasks.go`
- **Function:** `GetTaskHistory(taskID string) ([]*TaskChange, error)`
- **Complexity:** Low
- **Estimated Time:** 0.25 days
- **Dependencies:** T3.15 (for struct reference)
- **Description:**
  - Query `task_changes` table
  - Filter by task_id
  - Order by timestamp DESC
  - Parse JSON values
  - Return change list

#### T3.17: Comment Operations - Add/Get Comments
- **File:** `internal/database/tasks.go`
- **Functions:**
  - `AddComment(taskID, commentType, content string) (string, error)`
  - `GetComments(taskID string) ([]*TaskComment, error)`
  - `GetCommentsByType(taskID, commentType string) ([]*TaskComment, error)`
- **Complexity:** Medium
- **Estimated Time:** 0.5 days
- **Dependencies:** None (within group)
- **Description:**
  - Insert into `task_comments` table
  - Generate comment ID
  - Query comments with filters
  - Return comment structs

#### T3.18: Activity Log Operations
- **File:** `internal/database/tasks.go`
- **Functions:**
  - `RecordActivity(activityType, taskID, taskName string, details map[string]interface{}) error`
  - `GetActivityLog(taskID string, limit int) ([]*TaskActivity, error)`
- **Complexity:** Low
- **Estimated Time:** 0.5 days
- **Dependencies:** None (within group)
- **Description:**
  - Insert into `task_activities` table
  - Store details as JSON
  - Query recent activities
  - Return activity list

**Group 4 Total (Parallel):** 0.75 days (vs 1.5 days sequential)

---

### Group 5: Utility & Helper Functions (3 subtasks)

**Dependencies:** sqlite-migration-2, T3.1-T3.4 (for struct references)  
**Can Run In Parallel:** ✅ Yes (all independent)

#### T3.19: Task Filters Struct & Builder
- **File:** `internal/database/tasks.go`
- **Type:** `TaskFilters struct`
- **Function:** `NewTaskFilters() *TaskFilters`
- **Complexity:** Low
- **Estimated Time:** 0.25 days
- **Dependencies:** None (within group)
- **Description:**
  - Define filter struct with optional fields
  - Builder pattern for chaining filters
  - Convert to SQL WHERE clause
  - Support status, priority, tags, project_id filters

#### T3.20: Task Conversion Helpers
- **File:** `internal/database/tasks.go`
- **Functions:**
  - `taskToRow(task *Todo2Task) ([]interface{}, error)`
  - `rowToTask(rows *sql.Rows) (*Todo2Task, error)`
  - `scanTask(rows *sql.Rows) (*Todo2Task, error)`
- **Complexity:** Medium
- **Estimated Time:** 0.5 days
- **Dependencies:** T3.1-T3.4 (for struct reference)
- **Description:**
  - Convert task struct to database row
  - Convert database row to task struct
  - Handle JSON encoding/decoding
  - Handle tag/dependency arrays

#### T3.21: Prepared Statements Initialization
- **File:** `internal/database/tasks.go`
- **Function:** `initPreparedStatements(db *sql.DB) error`
- **Complexity:** Medium
- **Estimated Time:** 0.5 days
- **Dependencies:** All previous functions (for SQL queries)
- **Description:**
  - Create prepared statements for common queries
  - Store in package-level variables
  - Optimize for performance
  - Handle statement errors

**Group 5 Total (Parallel):** 0.75 days (vs 1.25 days sequential)

---

### Group 6: Testing (Can run in parallel with all groups)

**Dependencies:** Each test depends on its corresponding implementation  
**Can Run In Parallel:** ✅ Yes (write tests as functions are implemented)

#### T3.22: Core CRUD Tests
- **File:** `internal/database/tasks_test.go`
- **Tests:** CreateTask, GetTask, UpdateTask, DeleteTask
- **Complexity:** Medium
- **Estimated Time:** 1 day
- **Dependencies:** T3.1-T3.4
- **Description:**
  - Unit tests for each CRUD operation
  - Test error cases
  - Test transaction rollback
  - Test data integrity

#### T3.23: Query Operation Tests
- **File:** `internal/database/tasks_test.go`
- **Tests:** All query functions (ListTasks, GetByStatus, etc.)
- **Complexity:** Medium
- **Estimated Time:** 1 day
- **Dependencies:** T3.5-T3.10
- **Description:**
  - Test each query function
  - Test filters
  - Test sorting
  - Test pagination

#### T3.24: Relationship Operation Tests
- **File:** `internal/database/tasks_test.go`
- **Tests:** Tag ops, dependency ops, validation
- **Complexity:** High
- **Estimated Time:** 1 day
- **Dependencies:** T3.11-T3.14
- **Description:**
  - Test tag add/remove
  - Test dependency add/remove
  - Test circular dependency detection
  - Test dependency chain validation

#### T3.25: History & Comment Tests
- **File:** `internal/database/tasks_test.go`
- **Tests:** Change tracking, comments, activities
- **Complexity:** Low
- **Estimated Time:** 0.5 days
- **Dependencies:** T3.15-T3.18
- **Description:**
  - Test change recording
  - Test history retrieval
  - Test comment operations
  - Test activity logging

#### T3.26: Integration Tests
- **File:** `internal/database/tasks_integration_test.go`
- **Tests:** End-to-end workflows
- **Complexity:** High
- **Estimated Time:** 1 day
- **Dependencies:** All implementation tasks
- **Description:**
  - Test complete workflows (create → update → delete)
  - Test complex queries
  - Test concurrent operations
  - Test transaction handling

**Group 6 Total (Parallel with implementation):** 1.5 days (overlaps with implementation)

---

## Parallel Execution Plan

### Day 1: Core CRUD (Group 1)
**Morning:**
- T3.1: CreateTask (0.5 days)
- T3.2: GetTask (0.5 days)

**Afternoon:**
- T3.3: UpdateTask (1 day) - continues into Day 2 morning
- T3.4: DeleteTask (0.25 days) - can start after T3.1/T3.2

**Evening:**
- T3.22: Write tests for T3.1-T3.4

### Day 2: Queries & Relationships
**Morning:**
- Finish T3.3: UpdateTask
- T3.5: ListTasks (1 day) - starts morning, continues afternoon

**Afternoon:**
- T3.6-T3.10: Query operations (parallel) - 1.5 days total
  - T3.6: GetTasksByStatus (0.25 days)
  - T3.7: GetTasksByTag (0.5 days)
  - T3.8: GetTasksByPriority (0.25 days)
  - T3.9: GetTasksByProject (0.25 days)
  - T3.10: SearchTasks (0.5 days)

**Evening:**
- T3.23: Write tests for query operations

### Day 3: Relationships & History
**Morning:**
- T3.11: Tag operations (0.5 days)
- T3.12: Dependency operations (0.75 days)

**Afternoon:**
- T3.13: GetDependents (0.25 days)
- T3.14: ValidateDependencyChain (0.75 days)
- T3.15: Change tracking (0.25 days)
- T3.16: GetTaskHistory (0.25 days)

**Evening:**
- T3.17: Comment operations (0.5 days)
- T3.18: Activity log (0.5 days)
- T3.24: Write relationship tests
- T3.25: Write history/comment tests

### Day 4: Utilities & Final Testing
**Morning:**
- T3.19: Task filters (0.25 days)
- T3.20: Conversion helpers (0.5 days)
- T3.21: Prepared statements (0.5 days)

**Afternoon:**
- T3.26: Integration tests (1 day)

**Evening:**
- Code review
- Performance optimization
- Documentation

## Time Summary

| Approach | Time |
|----------|------|
| **Sequential** | 5-6 days |
| **Parallel (this plan)** | 3-4 days |
| **Savings** | 2-3 days |

## Dependencies Summary

### Critical Path
```
T3.1 (CreateTask) → T3.2 (GetTask) → T3.5 (ListTasks) → T3.20 (Conversion) → T3.21 (Prepared Stmts)
```

### Independent Parallel Streams
1. **Stream 1:** T3.1, T3.2, T3.3, T3.4 (Core CRUD)
2. **Stream 2:** T3.6, T3.7, T3.8, T3.9, T3.10 (Query ops) - after T3.5
3. **Stream 3:** T3.11, T3.12, T3.13, T3.14 (Relationships)
4. **Stream 4:** T3.15, T3.16, T3.17, T3.18 (History/Comments)
5. **Stream 5:** T3.19, T3.20, T3.21 (Utilities)
6. **Stream 6:** T3.22-T3.26 (Testing) - parallel with implementation

## Recommendations

1. **Start with Core CRUD (T3.1-T3.4)** - Foundation for everything else
2. **Implement ListTasks (T3.5) early** - Other queries can reuse its patterns
3. **Write tests immediately** - Don't wait until end, write tests as you implement
4. **Use TDD for complex functions** - T3.3 (UpdateTask), T3.14 (ValidateDependencyChain)
5. **Batch similar operations** - All "GetByX" queries together, all tag ops together

## Risk Mitigation

1. **Interface Changes:** Define interfaces early, get review before implementation
2. **Integration Issues:** Test integration after each group completes
3. **Performance:** Benchmark prepared statements vs ad-hoc queries
4. **Data Integrity:** Test foreign key constraints and cascades thoroughly

