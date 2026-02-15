# Stale Lock Detection and Handling

**Date**: 2026-01-10
**Status**: ✅ Implemented
**Purpose**: Detect and handle stale locks from dead agents or crashed processes

---

## 🎯 Problem Statement

When agents crash, disconnect, or fail:
- **Locks remain held indefinitely** - Tasks stay locked even though agent is dead
- **Other agents blocked** - Can't claim locked tasks
- **Tasks stuck** - Work stalls until manual intervention

---

## ✅ Solution: Multi-Layered Stale Lock Detection

### **Layer 1: Lease Expiration** (Primary)

**How it works:**
- Every lock has a `lock_until` timestamp (Unix epoch)
- When agent claims task, lease is set (default: 30 minutes)
- Lock automatically expires when `lock_until < now()`

**Detection:**
```go
// Automatic detection in ClaimTaskForAgent
if lockUntil.Valid && lockUntil.Int64 > now {
    // Lock is still valid
    return error("task already assigned")
}
// Lock expired - can reassign
```

**Cleanup:**
```go
// Cleanup expired locks
cleaned, err := database.CleanupExpiredLocks(ctx)
```

---

### **Layer 2: Stale Lock Detection** (Monitoring)

**Detection Functions:**

1. **`DetectStaleLocks()`** - Find all stale/expired locks
   ```go
   info, err := database.DetectStaleLocks(ctx, 5*time.Minute)
   // Returns:
   // - ExpiredCount: Number of expired locks
   // - NearExpiryCount: Locks expiring soon
   // - StaleCount: Locks expired > 5 minutes
   // - Locks: Detailed list of all locked tasks
   ```

2. **`GetLockStatus()`** - Check specific task lock status
   ```go
   status, err := database.GetLockStatus(ctx, "T-123")
   if status != nil {
       if status.IsExpired {
           log.Printf("Lock expired %v ago", status.TimeExpired)
       }
       if status.IsStale {
           log.Printf("Lock is stale (expired > 5min)")
       }
   }
   ```

3. **`CleanupExpiredLocksWithReport()`** - Cleanup with detailed report
   ```go
   cleaned, taskIDs, err := database.CleanupExpiredLocksWithReport(ctx, 1*time.Hour)
   // Returns:
   // - cleaned: Number of locks cleaned
   // - taskIDs: List of task IDs that were cleaned
   ```

---

### **Conflict detection** (optimistic locking)

When `UpdateTask` fails due to a concurrent update, use:

- **`database.IsVersionMismatchError(err)`** — reports whether an error is a version mismatch (another agent modified the task).
- **`database.CheckUpdateConflict(ctx, taskID, expectedVersion)`** — returns whether the task's current version differs from expected (e.g. to check before update or after failure). Returns `(hasConflict, currentVersion, err)`.
- **`database.ErrVersionMismatch`** — sentinel error; wrap with `%w` when returning version mismatch.

**Flow:** Reload task (e.g. `GetTask`), then retry update with fresh version.

---

### **Layer 3: Lease Renewal** (Prevention)

**How it works:**
- Agents renew leases before expiration
- Long-running tasks keep locks alive
- Missing renewals = dead agent

**Usage:**

One-shot renewal:
```go
err := database.RenewLease(ctx, taskID, agentID, 30*time.Minute)
```

Background renewal (for long-running work; stops when ctx is cancelled):
```go
// Renew every 20 min for a 30 min lease; goroutine stops on ctx.Done()
database.RunLeaseRenewal(ctx, taskID, agentID, 30*time.Minute, 20*time.Minute)
defer cancel() // when done, cancel to stop renewal
```

Manual ticker (alternative to RunLeaseRenewal):
```go
ticker := time.NewTicker(10 * time.Minute)
defer ticker.Stop()
for {
    select {
    case <-ticker.C:
        if err := database.RenewLease(ctx, taskID, agentID, 30*time.Minute); err != nil {
            log.Printf("Failed to renew lease: %v", err)
        }
    case <-ctx.Done():
        return
    }
}
```

---

## Process verification (T-319)

To verify that an agent holding a lock is still running (before treating a lock as stale):

- **`utils.ProcessExists(pid int) bool`** — returns true if the process with the given PID exists (Unix: `kill(pid, 0)`; Windows: no-op, returns true).
- **`database.ParsePIDFromAgentID(agentID string) (pid int, ok bool)`** — extracts PID from agent ID format `{type}-{hostname}-{pid}`.
- **`database.AgentProcessExists(agentID string) bool`** — returns true if the agent’s process is running (uses PID from agent ID). Returns false if agent ID has no PID (e.g. from GetAgentIDSimple).

Use in stale-lock logic: `CleanupDeadAgentLocks(ctx, staleThreshold)` does this automatically — it calls `DetectStaleLocks`, then for each expired lock skips when `AgentProcessExists(assignee)` is true, and releases the rest via `releaseLocksForTaskIDs`.

---

## 🔍 Detection Strategies

### **1. Expired Locks (Time-Based)**

**Definition:** `lock_until < now()`

**Detection:**
```sql
SELECT id, assignee, lock_until
FROM tasks
WHERE assignee IS NOT NULL
  AND lock_until IS NOT NULL
  AND lock_until < strftime('%s', 'now')
```

**Action:** Auto-cleanup when detected during claim attempts

---

### **2. Stale Locks (Expired + Age)**

**Definition:** Expired AND expired for > 5 minutes

**Detection:**
```sql
SELECT id, assignee, lock_until
FROM tasks
WHERE assignee IS NOT NULL
  AND lock_until IS NOT NULL
  AND lock_until < (strftime('%s', 'now') - 300)  -- 5 minutes ago
```

**Action:** Prioritize cleanup, alert if many stale locks

---

### **3. Near-Expiry Locks (Preventive)**

**Definition:** Expiring within threshold (e.g., 5 minutes)

**Detection:**
```sql
SELECT id, assignee, lock_until
FROM tasks
WHERE assignee IS NOT NULL
  AND lock_until IS NOT NULL
  AND lock_until < (strftime('%s', 'now') + 300)  -- Within 5 minutes
  AND lock_until > strftime('%s', 'now')           -- Not expired yet
```

**Action:** Warn agents to renew leases

---

### **4. Orphaned Locks (Agent Dead)**

**Definition:** Lock exists but agent process not running

**Detection:**
- Check if agent process exists (requires process monitoring)
- Compare `assigned_at` vs `updated_at` (no activity for > lease duration)
- Heartbeat mechanism (agents report alive)

**Action:** Cleanup if agent not responding

---

## 🛠️ Scheduled Cleanup

### **Automatic Cleanup on Claim**

**When:** Every time an agent tries to claim a task

**Action:**
```go
// In ClaimTaskForAgent()
if lockUntil.Valid && lockUntil.Int64 < now {
    // Lock expired - can reassign
    // Cleanup happens automatically during claim
}
```

**Benefits:**
- No background process needed
- Cleanup happens when needed
- Automatic recovery from dead agents

---

### **Periodic Background Cleanup**

**Schedule:** Run every 5-10 minutes (cron or background goroutine)

**Implementation:**
```go
// Background cleanup goroutine
func StartLockCleanupBackground(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            cleaned, err := database.CleanupExpiredLocks(ctx)
            if err != nil {
                log.Printf("Lock cleanup error: %v", err)
            } else if cleaned > 0 {
                log.Printf("Cleaned up %d expired locks", cleaned)
            }
        case <-ctx.Done():
            return
        }
    }
}
```

**Usage:**
```go
// Start in main() or server startup
go StartLockCleanupBackground(context.Background(), 5*time.Minute)
```

---

### **Scheduled Cleanup via Cron**

**Schedule:** Run every 5-10 minutes

**Script:**
```bash
#!/bin/bash
# scripts/cleanup_stale_locks.sh

cd /path/to/project
./bin/exarp-go cleanup-locks --max-age 1h --report
```

**Cron Entry:**
```bash
# Run every 5 minutes
*/5 * * * * /path/to/scripts/cleanup_stale_locks.sh >> /var/log/lock-cleanup.log 2>&1
```

---

## 📊 Monitoring and Alerting

### **Lock Status Monitoring**

**Function:** `DetectStaleLocks()`

**Metrics to track:**
- Total locked tasks
- Expired locks count
- Stale locks count (> 5 min expired)
- Near-expiry locks count

**Example:**
```go
info, err := database.DetectStaleLocks(ctx, 5*time.Minute)
if err != nil {
    log.Printf("Failed to detect stale locks: %v", err)
    return
}

// Alert if too many stale locks
if info.StaleCount > 10 {
    log.Printf("WARNING: %d stale locks detected (expired > 5min)", info.StaleCount)
    // Send alert notification
}

// Report lock status
log.Printf("Lock Status: %d total, %d expired, %d stale, %d near-expiry",
    len(info.Locks), info.ExpiredCount, info.StaleCount, info.NearExpiryCount)
```

---

### **Individual Task Monitoring**

**Function:** `GetLockStatus()`

**Use cases:**
- Check specific task before manual intervention
- Debug why task is stuck
- Verify lock state after agent crash

**Example:**
```go
status, err := database.GetLockStatus(ctx, "T-123")
if err != nil {
    log.Printf("Failed to get lock status: %v", err)
    return
}

if status == nil {
    log.Printf("Task T-123 is not locked")
    return
}

if status.IsExpired {
    log.Printf("Task T-123 lock expired %v ago (assigned to %s)",
        status.TimeExpired, status.Assignee)
}

if status.IsStale {
    log.Printf("Task T-123 has stale lock (expired > 5min)")
    // Cleanup automatically
    database.CleanupExpiredLocks(ctx)
}
```

---

## 🔧 Configuration

### **Lease Duration**

**Default:** 30 minutes

**When to adjust:**
- **Shorter leases** (5-10 min): Fast cleanup, but requires frequent renewal
- **Longer leases** (60+ min): Less renewal overhead, but slower cleanup

**Example:**
```go
// Short lease for quick tasks
leaseDuration := 10 * time.Minute
result, err := database.ClaimTaskForAgent(ctx, taskID, agentID, leaseDuration)

// Long lease for complex tasks
leaseDuration := 2 * time.Hour
result, err := database.ClaimTaskForAgent(ctx, taskID, agentID, leaseDuration)
```

---

### **Stale Threshold**

**Default:** 5 minutes (expired > 5min = stale)

**Configuration:**
```go
// Detect locks expired > 10 minutes
info, err := database.DetectStaleLocks(ctx, 10*time.Minute)

// Near-expiry threshold (warn if expiring within 5 min)
nearExpiry := 5 * time.Minute
info, err := database.DetectStaleLocks(ctx, nearExpiry)
```

---

### **Cleanup Max Age**

**Default:** No max age (cleanup all expired)

**Configuration:**
```go
// Only cleanup locks expired > 1 hour ago
cleaned, taskIDs, err := database.CleanupExpiredLocksWithReport(ctx, 1*time.Hour)
```

**Use case:** Prevent cleanup of recently expired locks (agent might still be working)

---

## 🎯 Best Practices

### **For Agents (Lock Claimers)**

1. **Always set reasonable lease duration**
   ```go
   // Good: Match lease to expected task duration
   leaseDuration := 30 * time.Minute  // For typical tasks
   ```

2. **Renew leases for long-running tasks**
   ```go
   // Renew every 10 minutes for long tasks
   ticker := time.NewTicker(10 * time.Minute)
   defer ticker.Stop()
   
   for {
       select {
       case <-ticker.C:
           if err := database.RenewLease(ctx, taskID, agentID, 30*time.Minute); err != nil {
               // Lost lock - need to re-claim
               break
           }
       case <-ctx.Done():
           return
       }
   }
   ```

3. **Release locks when done**
   ```go
   defer database.ReleaseTask(ctx, taskID, agentID)
   ```

4. **Handle lock failures gracefully**
   ```go
   result, err := database.ClaimTaskForAgent(ctx, taskID, agentID, leaseDuration)
   if err != nil {
       if result != nil && result.WasLocked {
           // Task already locked - try different task
           continue
       }
       return err
   }
   ```

---

### **For System Administrators**

1. **Monitor stale locks regularly**
   ```bash
   # Check lock status
   ./bin/exarp-go lock-status
   
   # Detect stale locks
   ./bin/exarp-go detect-stale-locks
   ```

2. **Set up periodic cleanup (scheduled via cron)**  
   The automation tool runs dead-agent lock cleanup as part of `action=daily`. Use system cron to run it from the project root:
   ```bash
   # Cron: daily at 02:00 (runs automation daily, which includes dead_agent_cleanup)
   0 2 * * * cd /path/to/exarp-go && ./bin/exarp-go -tool automation -args '{"action":"daily"}' >> /tmp/exarp-automation.log 2>&1
   ```
   For more frequent cleanup, run the same command every 10–15 minutes; `dead_agent_cleanup` releases expired locks only.

3. **Alert on high stale lock count**
   - Monitor `DetectStaleLocks()` output
   - Alert if `StaleCount > 10`
   - Investigate agent crashes if pattern emerges

---

## 🚨 Edge Cases

### **1. Agent Crashes Mid-Task**

**Scenario:** Agent crashes while task is "In Progress"

**Solution:**
- Lock expires automatically after lease duration
- Next agent can claim task
- Task status reverts to "Todo" (preserves work)

**Prevention:**
- Short lease durations (10-15 min) for quick cleanup
- Periodic renewal for long tasks

---

### **2. Network Partition**

**Scenario:** Agent still running but can't reach database

**Solution:**
- Lock expires after lease duration
- Agent can't renew lease
- Another agent can claim task

**Recovery:**
- Agent should check lock status after reconnection
- If lost lock, re-claim or abandon work

---

### **3. Clock Skew**

**Scenario:** Server and client clocks are out of sync

**Solution:**
- Always use server time (`strftime('%s', 'now')`)
- Don't rely on client-provided timestamps
- Lease expiration checked server-side

---

### **4. Rapid Claim/Release**

**Scenario:** Agent claims and releases very quickly

**Solution:**
- No minimum lease duration required
- Cleanup handles empty locks gracefully
- No performance impact

---

## 📝 Implementation Checklist

### **Phase 1: Basic Cleanup** ✅

- [x] `CleanupExpiredLocks()` - Basic cleanup function
- [x] Automatic cleanup in `ClaimTaskForAgent()` (expired locks)
- [x] Database migration with `lock_until` field

### **Phase 2: Monitoring** ✅

- [x] `DetectStaleLocks()` - Comprehensive detection
- [x] `GetLockStatus()` - Individual task status
- [x] `CleanupExpiredLocksWithReport()` - Detailed cleanup report

### **Phase 3: Background Processes**

- [ ] Background goroutine for periodic cleanup — **Future improvement**: exarp-go runs as STDIO (short-lived per request); a background goroutine only helps long-lived processes (HTTP/SSE). Defer until multi-agent/server deployment.
- [ ] CLI command for lock status monitoring
- [x] Scheduled cleanup via cron — Use system cron to run `exarp-go -tool automation -args '{"action":"daily"}'` from project root (see "Set up periodic cleanup" above). Automation runs `dead_agent_cleanup` which releases expired locks.

### **Phase 4: Advanced Features**

- [x] Process monitoring (verify agent process exists) — T-319 Done; `utils.ProcessExists`, `database.AgentProcessExists`
- [x] Dead agent cleanup job — T-76 Done; `database.CleanupDeadAgentLocks`, wired into automation (daily/nightly/sprint) and session prime
- [ ] Lock statistics and metrics
- [ ] Alerting system for stale lock patterns

---

## 🔗 Related Documentation

- `docs/NATIVE_GO_LOCKING_IMPLEMENTATION.md` - Locking implementation
- `docs/AGENT_LOCKING_STRATEGY.md` - Overall locking strategy
- `internal/database/tasks_lock.go` - Locking functions
- `internal/database/lock_monitoring.go` - Detection functions

---

## 🔮 Future Improvements

- **Background goroutine for periodic cleanup** — exarp-go runs as STDIO (short-lived per request). A background goroutine only helps long-lived processes (HTTP/SSE server). Defer until multi-agent/server deployment. Design: `StartLockCleanupBackground(ctx, 5*time.Minute)` calling `database.CleanupExpiredLocks()` on a ticker; see "Periodic Background Cleanup" section above.
- **Heartbeat mechanism (agents report alive)** — ~~T-318, T-79~~ *(removed)*. Agents must run continuously to send periodic heartbeats; not applicable with STDIO (short-lived per request). Defer until long-lived agent deployment.
- **Monitoring dashboard** — ~~T-232~~ *(removed)*. Implies a long-lived web server to serve the UI. Defer until HTTP/SSE server deployment. See `docs/MULTI_AGENT_PLAN.md`.

---

**Status**: ✅ Core functionality implemented; background cleanup and long-lived-only features documented as future improvements
