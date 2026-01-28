# Python Fallbacks Safe to Remove

**Date:** 2026-01-27  
**Scope:** Which Python bridge fallbacks in exarp-go can be removed so tools are fully native Go (no bridge call, no bridge branch in `execute_tool.py`).

---

## Summary

**Safely removable (5 tools):** `setup_hooks`, `check_attribution`, `session`, `memory_maint`, `analyze_alignment`.

For each, the Go handler already implements all supported actions. The Python fallback is only used when the native path returns an error. Removing it means those errors are surfaced to the user instead of being hidden by a Python run. No behavior is lost for success paths.

**Status (2026-01-27):** These five are **fully native** — handlers call only native code; the bridge does not route them (see comment in `bridge/execute_tool.py`). Automation uses `runDailyTask(ctx, "memory_maint", ...)` and `runDailyTask(ctx, "analyze_alignment", ...)`. Migration plan and regression tests updated.

---

## Tool-by-tool analysis

### ✅ Safe to remove

| Tool | Native coverage | Bridge used when | Extra change |
|------|-----------------|------------------|--------------|
| **setup_hooks** | `handleSetupHooksNative`: git + patterns | Native returns error (e.g. not a git repo) | None |
| **check_attribution** | `handleCheckAttributionNative`: full flow | Native returns error | None |
| **session** | `handleSessionNative`: prime, handoff, prompts, assignee | Native returns error | None |
| **memory_maint** | `handleMemoryMaintNative`: health, gc, prune, consolidate, dream | Native returns error | In `automation_native.go`, use `runDailyTask(ctx, "memory_maint", ...)` (done). |
| **analyze_alignment** | `handleAnalyzeAlignmentNative`: todo2 + prd (persona alignment) | — | Removed fallback 2026-01-27; bridge no longer routes it. |

### ❌ Not safe to remove (bridge still required)

| Tool | Why bridge is still needed |
|------|----------------------------|
| **memory** | Fallback used for semantic search / when native fails. |
| **report** | Briefing and scorecard no longer fall back: briefing is native-only (error if wisdom engine unavailable); non-Go scorecard returns clear error. Overview/prd still fall back to Python on native failure. |
| **security** | Fallback when native scan/alerts/report fails. |
| **task_analysis** | Removed: fully native; no Python fallback (FM provider abstraction; hierarchy returns clear error when FM not supported). |
| **task_discovery** | Fallback when native fails. |
| **task_workflow** | (1) Handler fallback for Apple-FM–related errors. (2) **Required:** `task_workflow_common.handleTaskWorkflowSync` calls the bridge when `external=true` (agentic-tools sync). |
| **testing** | `suggest` / `generate` are Python-only. |
| **lint** | Non-Go linters (e.g. ruff) use bridge by design. |
| **estimation** | Native only; no fallback (2026-01-27). |
| **ollama** | Actions like docs/quality/summary are Python-only. |
| **mlx** | No native impl; bridge-only. |
| **context** | Summarize without Apple FM, and some actions, use bridge. |
| **recommend** | Fallback when native model/workflow/advisor fails. |

---

## Implementation checklist (when removing the 4 tools)

For **setup_hooks**, **check_attribution**, **session**, **memory_maint**:

1. **Go handlers** (`internal/tools/handlers.go`)
   - Remove the `bridge.ExecutePythonTool(ctx, "<tool>", params)` block.
   - Return the error from the native call instead (e.g. `return nil, err` after the native attempt).

2. **Bridge** (`bridge/execute_tool.py`)
   - Remove the `tool_name == "<tool>"` branch and any imports used only by that branch.

3. **memory_maint only:** `internal/tools/automation_native.go`
   - Replace `runDailyTaskPython(ctx, "memory_maint", ...)` with `runDailyTask(ctx, "memory_maint", ...)` where it’s used (e.g. sprint daily, ~line 167).

4. **Tests**
   - `regression_test.go` and similar call the bridge for “session” (and possibly others) to compare native vs bridge. After removing these tools from the bridge, either:
     - Stop comparing against the bridge for those tools, or
     - Drop the bridge-invocation part of the test for them.

5. **Docs**
   - Update `NATIVE_GO_MIGRATION_PLAN.md` and any migration/audit docs to list these four as fully native with no Python fallback.

### Implementation status (2026-01-27)

**All checklist items are done for the four tools:**

- **Handlers:** No bridge calls for setup_hooks, check_attribution, session, memory_maint; each uses only its native handler.
- **Bridge:** `execute_tool.py` does not route these four (comment: "setup_hooks, check_attribution, session, memory_maint removed - fully native Go with no Python fallback").
- **Automation:** Sprint daily uses `runDailyTask(ctx, "memory_maint", ...)` at line 167.
- **Tests:** `toolsWithNoBridge` includes all four; `TestRegressionFallbackBehavior` skips bridge comparison for them; session case renamed to "session with invalid action (native only, no fallback)" with `expectFallback: false`.

---

## References

- Handlers and fallback logic: `internal/tools/handlers.go`
- Bridge routing: `bridge/execute_tool.py`
- Automation use of bridge: `internal/tools/automation_native.go` (`runDailyTaskPython`), `internal/tools/task_workflow_common.go` (task_workflow when `external=true`)
- Migration plan: `docs/NATIVE_GO_MIGRATION_PLAN.md`
