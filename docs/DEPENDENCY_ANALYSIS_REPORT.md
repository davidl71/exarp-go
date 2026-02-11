# Todo2 Dependency Analysis Report

**Date:** 2026-02-11  
**Tool:** `exarp-go task_analysis` (dependencies, validate, parallelization, execution_plan)

## Summary

| Check | Result |
|-------|--------|
| Missing dependency refs | ✅ None (all dep IDs exist) |
| Circular dependencies | ✅ None detected |
| Invalid task IDs | ✅ None |
| Critical path length | 8 tasks (Config Epic) |

---

## Correctly Set Dependencies

### TaskStore migration chain
- **T-1770827786553**, T-1770827787404, T-1770827788823, T-1770827789804 all depend on **T-1769989774217** (Extract TaskStore) — Done ✅
- T-1770827788204 (session/report) — Done ✅
- Dependencies are valid and dependency T-1769989774217 exists and is complete.

---

## Recommended Dependency Additions

### 1. task_discovery orphans (T-1770828166659, T-1770828167551, T-1770828167906)

| Task | Current deps | Recommended addition | Rationale |
|------|--------------|----------------------|-----------|
| **T-1770828167551** (Update nocgo orphans) | none | T-1770828166659 | Nocgo variant applies the same refactor; must follow the main refactor. |
| **T-1770828167906** (Add tests) | none | T-1770828166659 | Tests verify the refactored behavior; refactor must exist first. |

**Suggested fix:**
```
T-1770828167551: dependencies = [T-1770828166659]
T-1770828167906: dependencies = [T-1770828166659]
```

### 2. Rate limiting tests (T-277, T-288, T-289, T-291, T-292)

| Task | Current deps | Recommended addition | Rationale |
|------|--------------|----------------------|-----------|
| **T-277** (Add tests for rate limiting) | none | T-274 (optional) | If integration tests: need middleware. If unit tests only: ratelimit_test.go exists, can stay independent. |
| **T-288** (Test request limit enforcement) | none | T-274 | Integration test requires middleware in request path. |
| **T-289** (Test sliding window accuracy) | none | T-274 | Integration test requires middleware. |
| **T-291** (Test concurrent requests) | none | T-274 | Integration test requires middleware. |
| **T-292** (Test limit reset behavior) | none | T-274 | Integration test requires middleware. |

**Note:** Per `docs/RATE_LIMITING_TASKS.md`, these are lower priority for stdio-only (T-290 per-tool was removed). If keeping these tasks, add T-274 as dependency.

### 3. Agent locking / cleanup (T-76)

| Task | Current deps | Recommended addition | Rationale |
|------|--------------|----------------------|-----------|
| **T-76** (Add dead agent cleanup job) | none | T-319 | Cleanup needs a way to detect dead agents; T-318 (heartbeat) removed as long-lived-only. Process check (T-319) remains. |

*(T-315 background goroutine removed; documented as future improvement in STALE_LOCK_DETECTION_AND_HANDLING.md)*

---

## Tasks with No Dependencies (OK)

Many tasks correctly have no dependencies (level 0) and can run in parallel:

- **High priority:** T-249, T-256, T-319, T-320, T-80
- **Security tests:** T-283–T-287 (path validation)
- **Adapter/helper extraction:** T-1768251816882, T-1768253988999, T-1768253990892
- **Agent/model tasks:** T-207, T-208, T-209, T-211–T-218, T-221, T-224–T-231 *(T-232 removed; monitoring dashboard documented as future improvement)*
- **Lock/heartbeat:** T-75, T-317 *(T-79, T-318 removed; documented as future improvement)*

---

## Parallelization Snapshot

- **Group 1 (level 0):** 6 high-priority tasks can run in parallel
- **Group 2 (level 2):** 7 TaskStore migrations + 3 extract helpers can run in parallel
- **Group 3 (level 1):** 3 adapter tasks
- **Group 4 (level 0):** 42 tasks (no deps)

---

## Action Items — Applied 2026-02-11

1. **Add dependencies for task_discovery orphans** ✅
   - T-1770828167551 → T-1770828166659
   - T-1770828167906 → T-1770828166659

2. **Add dependencies for rate limit tests** ✅
   - T-288, T-289, T-291, T-292 → T-274

3. **Add dependency for dead agent cleanup** ✅
   - T-76 → T-319 (process monitoring)

4. ~~Link T-315 → T-76~~ — T-315 removed; background goroutine documented as future improvement.

**Method:** Direct SQLite `task_dependencies` insert + sync. `task_workflow` update handler was extended to support `dependencies` param (code change in progress; build has pre-existing errors).
