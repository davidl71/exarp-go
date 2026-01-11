# Agent Locking Strategy for Parallel Execution

**Date**: 2026-01-10
**Status**: Proposed
**Purpose**: Prevent parallel agents from interfering with each other during concurrent task execution

---

## 🎯 Problem Statement

When multiple agents execute tasks in parallel, they can:
- **Conflict on task assignment**: Multiple agents grab the same task
- **Overwrite each other's changes**: Race conditions on task updates
- **Work on conflicting files**: Multiple agents modify same files simultaneously

---

## 📊 Current Infrastructure Analysis

### ✅ What Exists

1. **Python File-Based Locking** (`project_management_automation/utils/file_lock.py`):
   - OS-level file locks (fcntl/msvcrt)
   - Task-level locks (`.todo2/locks/task_{id}.lock`)
   - State file locks (`.todo2/state.todo2.json.lock`)
   - Works well for JSON-based state file

2. **Go Database Layer** (`internal/database/`):
   - SQLite with WAL mode (allows concurrent readers)
   - Retry logic with exponential backoff (handles BUSY/LOCKED errors)
   - Transactions for atomicity
   - Missing: Row-level locking, assignee tracking, optimistic locking

### ❌ Current Gaps

1. **No Coordination Between Python and Go**:
   - Python file locks don't protect Go database operations
   - Go database operations don't use file locks
   - Two separate locking mechanisms

2. **No Explicit Task-Level Database Locking**:
   - No `SELECT FOR UPDATE` for row-level locks
   - No assignee field to prevent duplicate assignments
   - No version field for optimistic locking

3. **No Atomic "Claim Task" Operation**:
   - Agents can both read "available" status
   - Both try to update to "In Progress"
   - Last write wins (race condition)

---

## 🎯 Recommended Multi-Layered Approach

### **Layer 1: Database-Level Locking (Primary)** ⭐ **RECOMMENDED**

**Why**: 
- Most robust for SQLite
- Works across all processes (Go, Python, any language)
- Leverages SQLite's built-in concurrency controls
- Better performance than file locks

#### **1.1 Pessimistic Locking with SELECT FOR UPDATE**

Use `SELECT FOR UPDATE` to lock a task row before updating:

```go
// internal/database/tasks_lock.go
func ClaimTaskForAgent(ctx context.Context, taskID string, agentID string) (*Todo2Task, error) {
    return retryWithBackoff(ctx, func() error {
        db, err := GetDB()
        if err != nil {
            return err
        }
        
        tx, err := db.BeginTx(ctx, nil)
        if err != nil {
            return err
        }
        defer tx.Rollback()
        
        // SELECT FOR UPDATE - locks the row
        var task Todo2Task
        err = tx.QueryRowContext(ctx, `
            SELECT id, status, assignee, version
            FROM tasks
            WHERE id = ? AND (status = ? OR status = ?)
            FOR UPDATE
        `, taskID, StatusTodo, StatusInProgress).Scan(
            &task.ID,
            &task.Status,
            &task.Assignee, // NEW field
            &task.Version,  // NEW field for optimistic locking
        )
        
        if err == sql.ErrNoRows {
            return fmt.Errorf("task not available (already assigned or wrong status)")
        }
        
        // Check if already assigned
        if task.Assignee != "" && task.Assignee != agentID {
            return fmt.Errorf("task already assigned to %s", task.Assignee)
        }
        
        // Update: assign + change status atomically
        _, err = tx.ExecContext(ctx, `
            UPDATE tasks SET
                assignee = ?,
                status = ?,
                assigned_at = strftime('%s', 'now'),
                updated_at = strftime('%s', 'now')
            WHERE id = ? AND version = ?
        `, agentID, StatusInProgress, taskID, task.Version)
        
        if err != nil {
            return err
        }
        
        return tx.Commit()
    })
}
```

**Benefits**:
- ✅ Atomic operation (no race conditions)
- ✅ Works across all processes
- ✅ SQLite handles locking automatically
- ✅ Retry logic handles transient errors

#### **1.2 Optimistic Locking with Version Field**

Add version field to prevent lost updates:

```sql
-- Migration: Add version and assignee fields
ALTER TABLE tasks ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE tasks ADD COLUMN assignee TEXT;
ALTER TABLE tasks ADD COLUMN assigned_at INTEGER;
```

```go
func UpdateTaskWithVersion(ctx context.Context, task *Todo2Task, expectedVersion int64) error {
    return retryWithBackoff(ctx, func() error {
        db, err := GetDB()
        if err != nil {
            return err
        }
        
        result, err := db.ExecContext(ctx, `
            UPDATE tasks SET
                content = ?,
                status = ?,
                version = version + 1,
                updated_at = strftime('%s', 'now')
            WHERE id = ? AND version = ?
        `, task.Content, task.Status, task.ID, expectedVersion)
        
        if err != nil {
            return err
        }
        
        rowsAffected, _ := result.RowsAffected()
        if rowsAffected == 0 {
            return fmt.Errorf("task was modified by another agent (version mismatch)")
        }
        
        return nil
    })
}
```

**Benefits**:
- ✅ Detects concurrent modifications
- ✅ Prevents lost updates
- ✅ Works well with high read/low write ratio

#### **1.3 Task Assignment Tracking**

Add `assignee` field to track which agent has claimed a task:

```sql
-- Add to schema
assignee TEXT,              -- Agent ID that claimed the task
assigned_at INTEGER,        -- Timestamp when assigned
lock_until INTEGER,         -- Optional: lease expiration (for dead agent cleanup)
```

```go
func ClaimTask(ctx context.Context, taskID string, agentID string, leaseDuration time.Duration) error {
    // Atomic claim: check status + assign + lock
    tx, err := db.BeginTx(ctx, nil)
    
    // SELECT FOR UPDATE
    var currentAssignee sql.NullString
    var lockUntil sql.NullInt64
    
    err = tx.QueryRowContext(ctx, `
        SELECT assignee, lock_until
        FROM tasks
        WHERE id = ? AND status = ?
        FOR UPDATE
    `, taskID, StatusTodo).Scan(&currentAssignee, &lockUntil)
    
    // Check if available (unassigned or lease expired)
    now := time.Now().Unix()
    if currentAssignee.Valid && currentAssignee.String != "" {
        if lockUntil.Valid && lockUntil.Int64 > now {
            return fmt.Errorf("task already assigned to %s", currentAssignee.String)
        }
        // Lease expired - can reassign
    }
    
    // Claim task
    lockUntilTime := now + int64(leaseDuration.Seconds())
    _, err = tx.ExecContext(ctx, `
        UPDATE tasks SET
            assignee = ?,
            assigned_at = ?,
            lock_until = ?,
            status = ?
        WHERE id = ?
    `, agentID, now, lockUntilTime, StatusInProgress, taskID)
    
    return tx.Commit()
}
```

---

### **Layer 2: File-Based Locking (Fallback)**

Keep existing Python file locks for:
- **JSON mode fallback** (when database unavailable)
- **File-level coordination** (for Git operations, file writes)
- **Cross-language coordination** (Python ↔ Go via files)

```python
# Enhanced: Coordinate with database
def claim_task_atomic(task_id: str, agent_id: str) -> bool:
    """Claim task using both file lock and database."""
    
    # 1. Acquire file lock (prevents other Python processes)
    with task_lock(task_id=task_id):
        # 2. Try database claim (coordinates with Go processes)
        try:
            # Use MCP tool to claim via database
            result = update_todos_mcp([{
                'id': task_id,
                'status': 'In Progress',
                'assignee': agent_id
            }])
            return result['success']
        except:
            # Fallback to JSON file (if database unavailable)
            return atomic_assign_task(task_id, agent_id, assignee_type="agent")
```

---

### **Layer 3: Status-Based Coordination**

Use task status as a coordination mechanism:

```go
// Status transitions enforce locking
const (
    StatusTodo       = "Todo"        // Available for assignment
    StatusInProgress = "In Progress" // Locked by agent
    StatusReview     = "Review"      // Locked, awaiting human approval
    StatusDone       = "Done"        // Completed (read-only)
)
```

**Rules**:
- ✅ Only `StatusTodo` tasks can be claimed
- ✅ `StatusInProgress` means "locked by agent"
- ✅ Status transition to `In Progress` must be atomic

---

## 🔧 Implementation Plan

### **Phase 1: Database Schema Updates** (High Priority)

1. **Add Locking Fields**:
   ```sql
   ALTER TABLE tasks ADD COLUMN version INTEGER DEFAULT 1;
   ALTER TABLE tasks ADD COLUMN assignee TEXT;
   ALTER TABLE tasks ADD COLUMN assigned_at INTEGER;
   ALTER TABLE tasks ADD COLUMN lock_until INTEGER;
   ```

2. **Add Indexes**:
   ```sql
   CREATE INDEX idx_tasks_status_assignee ON tasks(status, assignee);
   CREATE INDEX idx_tasks_lock_until ON tasks(lock_until) WHERE lock_until > 0;
   ```

### **Phase 2: Atomic Claim Operations** (High Priority)

1. **Go Implementation**:
   - `ClaimTaskForAgent()` - atomic claim with SELECT FOR UPDATE
   - `ReleaseTask()` - release lock (set assignee = NULL)
   - `RenewLease()` - extend lock duration

2. **Python Bridge**:
   - Wrap Go functions via MCP tools
   - Fallback to file locks for JSON mode

### **Phase 3: Optimistic Locking** (Medium Priority)

1. **Version-Based Updates**:
   - Check version before update
   - Increment version on update
   - Return error if version mismatch

2. **Retry Logic**:
   - Automatically retry on version conflicts
   - Max retries (3-5 attempts)
   - Exponential backoff

### **Phase 4: Dead Agent Cleanup** (Medium Priority)

1. **Lease Expiration**:
   - `lock_until` timestamp
   - Background cleanup job
   - Release tasks with expired leases

2. **Heartbeat Mechanism**:
   - Agents renew leases periodically
   - Missing heartbeats = dead agent
   - Auto-release after timeout

---

## 🎯 Recommended Approach: **Hybrid (Database + Status)**

**Best for your architecture**:

### **1. Database SELECT FOR UPDATE** (Primary)
```go
// Use for all task claims/updates
SELECT ... FOR UPDATE  // Locks row in transaction
UPDATE ... WHERE id = ? AND version = ?  // Optimistic locking
```

### **2. Status as Lock Indicator**
```go
// Status transitions are atomic
Todo → In Progress  // Claim (requires SELECT FOR UPDATE)
In Progress → Review  // Unlock (anyone can review)
Review → Done  // Complete (unlock)
```

### **3. File Locks for JSON Fallback**
```python
# Only when database unavailable
with task_lock(task_id):
    # Modify JSON file
```

---

## 📋 Implementation Checklist

### **Immediate (Phase 1)**

- [ ] Add `version`, `assignee`, `assigned_at`, `lock_until` fields to schema
- [ ] Create migration script
- [ ] Implement `ClaimTaskForAgent()` with SELECT FOR UPDATE
- [ ] Implement `ReleaseTask()`
- [ ] Update `UpdateTask()` to use version checks

### **Short Term (Phase 2)**

- [ ] Add Python wrapper for claim operations
- [ ] Implement lease renewal mechanism
- [ ] Add dead agent cleanup job
- [ ] Update task workflow tools to use atomic claims

### **Long Term (Phase 3)**

- [ ] Fine-grained file locking (for Git operations)
- [ ] Agent heartbeat system
- [ ] Lock timeout monitoring/alerts
- [ ] Performance metrics (lock contention, wait times)

---

## 🔍 Comparison of Approaches

| Approach | Pros | Cons | Best For |
|----------|------|------|----------|
| **SELECT FOR UPDATE** | ✅ Atomic, cross-process, robust | ⚠️ Requires transactions | **Primary choice** |
| **Optimistic Locking** | ✅ High performance, detects conflicts | ⚠️ Retry overhead | High read/low write |
| **File Locks** | ✅ Simple, works everywhere | ⚠️ Python-only, no DB coordination | JSON fallback |
| **Status-Based** | ✅ Simple, visible in UI | ⚠️ Not atomic alone | **Combine with DB locks** |
| **Distributed Lock (Redis)** | ✅ Scalable, distributed | ⚠️ Extra dependency | Multi-machine setup |

---

## 🎯 Recommendation

**Use Hybrid Approach: Database + Status**

1. **Primary**: `SELECT FOR UPDATE` for atomic claims (Go database layer)
2. **Secondary**: Status transitions as lock indicators (visible in UI)
3. **Fallback**: File locks for JSON mode (Python compatibility)
4. **Future**: Optimistic locking with version field (performance optimization)

**Implementation Order**:
1. ✅ Add database fields (version, assignee)
2. ✅ Implement `ClaimTaskForAgent()` with SELECT FOR UPDATE
3. ✅ Update task workflow to use atomic claims
4. ✅ Add lease expiration/cleanup
5. ⏳ Add optimistic locking for updates

---

## 📚 References

- SQLite WAL Mode: https://www.sqlite.org/wal.html
- SELECT FOR UPDATE: https://www.sqlite.org/lang_select.html
- File Locking (Python): `project_management_automation/utils/file_lock.py`
- Database Retry Logic: `internal/database/retry.go`
- Task Assignment: `project_management_automation/utils/task_locking.py`
