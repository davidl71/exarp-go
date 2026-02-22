# Database Retry Pattern Consolidation

**Date:** 2026-02-22
**Status:** Research / Documentation
**Task:** T-1771572209680403000
**Tag hints:** `#database` `#refactor`

---

## Current State

`retryWithBackoff()` is defined in `internal/database/retry.go` and used across 7 files with 21 call sites:

| File | Calls | Purpose |
|------|-------|---------|
| `retry.go` | 1 (definition) | Central retry logic with config-driven backoff |
| `tasks_crud.go` | 4 | CreateTask, GetTask, UpdateTask, DeleteTask |
| `tasks_list.go` | 2 | ListTasks, ListTasksForEstimation |
| `tasks_lock.go` | 5 | ClaimTaskForAgent, ReleaseTask, RenewLease, CleanupExpiredLocks, BatchClaimTasks |
| `tasks_misc.go` | 1 | PurgeCompletedTasks |
| `lock_monitoring.go` | 4 | DetectStaleLocks, GetLockStatus, CleanupExpiredLocksWithReport, another cleanup variant |
| `comments.go` | 3 | AddComment, GetComments, DeleteComments |

## Current Pattern (Consistent)

Every call site follows the same structure:

```go
ctx = ensureContext(ctx)
ctx, cancel := withQueryTimeout(ctx)  // or withTransactionTimeout
defer cancel()

err := retryWithBackoff(ctx, func() error {
    db, err := GetDB()
    if err != nil {
        return fmt.Errorf("...: %w", err)
    }
    // ... SQL operation ...
    return nil
})
```

The retry function itself (`retry.go:87`) is well-implemented:
- Config-driven: attempts, initial delay, max delay, multiplier from `config.GetGlobalConfig()`
- Defaults: 3 retries, 100ms initial, 5s max, 2.0 multiplier
- Transient error detection via `isTransientError()` (SQLITE_BUSY, SQLITE_LOCKED)
- Context cancellation respected between retry attempts

## Identified Issues

### 1. Error Message Inconsistency

Error messages in `fmt.Errorf` wrapping vary across files:

```go
// tasks_crud.go
return fmt.Errorf("failed to get database: %w", err)

// comments.go
return fmt.Errorf("failed to get database: %w", err)

// tasks_lock.go
return fmt.Errorf("failed to get database: %w", err)
```

The "get database" message is consistent, but operation-specific error wrapping varies:

```go
// Some use "failed to X"
return fmt.Errorf("failed to create task: %w", err)

// Some use "X failed"  (less common but exists)
return fmt.Errorf("insert task failed: %w", err)
```

### 2. No Structured Logging on Retries

`retryWithBackoff` silently retries without logging. When debugging production issues, there is no visibility into:
- Which operations are being retried
- How many retries occurred before success/failure
- Which transient errors triggered retries

### 3. Timeout Selection

Call sites choose between `withQueryTimeout` (30s) and `withTransactionTimeout` (60s) manually. Some operations that could be transactions use the query timeout.

## Proposed Standardization

### Phase 1: Add Retry Logging (Low Risk)

Add optional structured logging to `retryWithBackoff`:

```go
func retryWithBackoff(ctx context.Context, fn func() error) error {
    // ... existing setup ...
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := fn()
        if err == nil {
            if attempt > 0 {
                slog.Debug("database operation succeeded after retry",
                    "attempts", attempt+1)
            }
            return nil
        }
        if isTransientError(err) && attempt < maxRetries-1 {
            slog.Warn("retrying database operation",
                "attempt", attempt+1,
                "max_retries", maxRetries,
                "error", err.Error())
        }
        // ... existing backoff logic ...
    }
}
```

### Phase 2: Standardize Error Messages (Medium Risk)

Create error helpers in `retry.go`:

```go
func dbGetError(op string, err error) error {
    return fmt.Errorf("%s: failed to get database: %w", op, err)
}
```

### Phase 3: Named Retry Wrapper (Low Risk)

Add an operation-name parameter for better diagnostics:

```go
func retryWithBackoffNamed(ctx context.Context, opName string, fn func() error) error
```

## Recommendation

1. **Phase 1 first** — Add `slog.Debug`/`slog.Warn` to the existing `retryWithBackoff`. No caller changes needed. Immediate observability improvement.
2. **Phase 2 deferred** — Error message standardization across all 21 call sites is a larger change; batch it with other database refactoring.
3. **Phase 3 optional** — The named wrapper is nice-to-have but requires updating all call sites.

**Estimated effort:** Phase 1: 30 min. Phase 2: 2h. Phase 3: 1h.

## References

- `internal/database/retry.go` — Central retry implementation
- `docs/FUTURE_REFACTORING_PLAN.md` — Item 3 (Database Retry Pattern Consolidation)
- Config: `config.GetGlobalConfig().Database.RetryAttempts` etc.
