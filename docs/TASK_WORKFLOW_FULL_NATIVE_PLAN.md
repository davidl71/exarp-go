# task_workflow Full Native Plan

**Goal:** Make `task_workflow` fully native (no Python bridge).  
**Status:** Done. External sync **removed** as a supported feature; documented as **future nice-to-have**. If `external=true` is passed, it is ignored and SQLite↔JSON sync is performed (no error).  
**Previously:** Option A returned error when `external=true`; now we ignore the param and document external sync as future nice-to-have.

---

## Current State

### Handler flow

- **Entry:** `handleTaskWorkflow` (handlers.go) → `handleTaskWorkflowNative` (task_workflow_native.go). No top-level bridge fallback; native errors are returned as-is.

### Actions (all have native Go)

| Action   | Native implementation              | Bridge? |
|----------|------------------------------------|--------|
| clarify  | handleTaskWorkflowClarify (Apple FM chain) | No      |
| approve  | handleTaskWorkflowApprove          | No      |
| sync     | handleTaskWorkflowSync             | No — `external` ignored; future nice-to-have |
| clarity  | handleTaskWorkflowClarity          | No      |
| cleanup  | handleTaskWorkflowCleanup         | No      |
| create   | handleTaskWorkflowCreate          | No      |

### External sync: removed, documented as future nice-to-have

**File:** `internal/tools/task_workflow_common.go`  
**Function:** `handleTaskWorkflowSync`

**Current behavior:** The `external` param is **not** implemented. If `external=true` is passed, it is **ignored** and SQLite↔JSON sync is performed. A note is added to the result: `external_sync_note: "External sync is a future nice-to-have; performed SQLite↔JSON sync only."`

**Future nice-to-have:** External sync (e.g. run infer_task_progress in-process then sync) could be added later without any agentic-tools MCP. See "Do we need agentic-tools MCP?" below.

*Historical:* When **sync** was called with **`external=true`**, the code used to invoke the Python bridge. The comment says “syncs with external task sources (agentic-tools)”. The Python `task_workflow()` in `consolidated_automation.py` does **not** take an `external` argument; for `action == "sync"` it only calls `sync_todo_tasks(dry_run, output_path)` (TODO table ↔ Todo2). So today the bridge path runs the same Python sync logic; there is no separate “agentic-tools” sync implementation in the Python code checked.

### Clarify and FM

- **clarify** is already native: `handleTaskWorkflowClarify` uses the FM chain (Apple → Ollama → stub). If FM is not available, it returns `ErrFMNotSupported`; there is no bridge fallback for clarify.

---

## Options to Reach Full Native

### Option A: Return error when `external=true` (recommended, minimal change)

**Change:** In `handleTaskWorkflowSync`, when `external == true`, do **not** call the bridge. Return a clear error instead.

**Pros:**

- Removes the last bridge use for `task_workflow`; tool becomes fully native.
- Small, localized change in one function.
- No new dependencies or MCP client in Go.

**Cons:**

- Callers that pass `external=true` will get an error until/unless external sync is implemented in Go (Option B) or deprecated.

**Implementation:**

1. In `task_workflow_common.go`, replace the `if external { ... bridge ... }` block with:
   - Return an error, e.g.:  
     `"external sync (agentic-tools) is not implemented in native Go; omit external or use external=false for SQLite↔JSON sync"`.
2. Optionally extend the tool schema/docs to state that `external` is unsupported in the native implementation.
3. Update `docs/NATIVE_GO_HANDLER_STATUS.md`: move `task_workflow` to “Full Native”, remove from “Hybrid”.
4. If the bridge is no longer invoked for `task_workflow`, consider removing or narrowing the `task_workflow` branch in `bridge/execute_tool.py` (optional cleanup).

**Tests:**

- Unit test: `task_workflow(sync, external=true)` returns the new error and does not call the bridge.
- Existing tests for sync with `external=false` (or no `external`) continue to pass.

---

### Option B: Implement external sync in Go

**Idea:** Provide “external sync” (e.g. agentic-tools) inside the Go server so `external=true` no longer needs the bridge.

**Requirements (to be refined):**

1. **Define “external sync”:** e.g. call agentic-tools MCP (e.g. `infer_task_progress` or equivalent), then update Todo2 from that result. Align with how Python was intended to use agentic-tools (see `auto_update_task_status.py` / `agentic_tools_client`).
2. **MCP client in Go:** Call agentic-tools MCP from Go (stdio or HTTP, depending on how Cursor runs it). May reuse or mirror patterns from existing Go MCP usage in the repo.
3. **Orchestration:** From `handleTaskWorkflowSync(external=true)`:
   - Resolve project root, load Todo2 state.
   - Call agentic-tools (and any other external source) and map responses to task updates.
   - Apply updates (Todo2 DB / JSON) using existing native helpers.
   - Return a structured result (e.g. same shape as current sync result plus “external” metadata).

**Pros:**

- Keeps `external=true` working and makes it native.
- Single implementation for sync regardless of `external`.

**Cons:**

- Higher effort: MCP client, agentic-tools contract, error handling, tests.
- Depends on agentic-tools MCP being available and stable.

**Suggested steps:**

1. Document the exact agentic-tools sync contract (which tools, args, response format).
2. Add a small Go package or internal helper to call agentic-tools MCP (e.g. `internal/agentic` or under `internal/tools`).
3. Implement “external sync” in `handleTaskWorkflowSync` when `external=true` using that client and existing Todo2 write path.
4. Remove the bridge call for `task_workflow` and update NATIVE_GO_HANDLER_STATUS.md and tests as in Option A.

---

### Option C: Deprecate `external=true`

**Idea:** Treat external sync as deprecated; document that callers should use sync without `external` (or a future dedicated tool). Same code change as Option A (return error when `external=true`), plus:

- Schema/docs: mark `external` as deprecated or unsupported.
- Changelog/README: state that external sync via `task_workflow` is no longer supported; point to SQLite↔JSON sync or future external-sync design.

---

## Do we need agentic-tools MCP?

**No.** We do **not** need to call an external agentic-tools MCP.

- **"Agentic" in this repo** = **Agentic CI** (`.github/workflows/agentic-ci.yml`) — the workflow that runs exarp-go tools. Not an agentic-tools MCP server.
- **External sync** was never implemented in Python in this repo: `consolidated_automation.py` sync only does SQLite↔JSON; there is no `auto_update_task_status.py` or `agentic_tools_client` in exarp-go.
- **We already implement natively:** **infer_task_progress** (native Go: heuristics + optional FM, `auto_update_tasks` to mark completed tasks Done; used in automation daily; no Python, no agentic-tools MCP). **task_workflow** sync/approve/clarify/clarity/cleanup/create/list are all native.
- **If we want `external=true` to do something:** implement sync(external=true) as "run infer_task_progress (e.g. auto_update_tasks=true) in-process, then SQLite↔JSON sync" — all native, no MCP.

---

## Recommendation

- **Short term:** **Option A** — return a clear error when `external=true`. That makes `task_workflow` fully native with minimal code and no new dependencies.
- **Later:** If product needs “external sync” again, implement it in Go (**Option B**) or introduce a dedicated tool/flow and keep `task_workflow` sync as SQLite↔JSON only (**Option C**).

---

## Files to Touch (Option A)

| File | Change |
|------|--------|
| `internal/tools/task_workflow_common.go` | In `handleTaskWorkflowSync`, when `external==true` return error instead of calling bridge. |
| `docs/NATIVE_GO_HANDLER_STATUS.md` | Move `task_workflow` to Full Native; remove from Hybrid. |
| `internal/tools/regression_test.go` | Ensure `task_workflow` is documented as fully native (already says “Fully native” for clarify; add note that sync external returns error). |
| Optional: `bridge/execute_tool.py` | Remove or guard `task_workflow` branch if no longer needed. |
| Optional: Tool schema in `registry.go` | Document that `external` is unsupported in native implementation. |

---

## References

- Handler: `internal/tools/handlers.go` (`handleTaskWorkflow`)
- Sync logic: `internal/tools/task_workflow_common.go` (`handleTaskWorkflowSync`)
- Native router: `internal/tools/task_workflow_native.go` (`handleTaskWorkflowNative`)
- Python entry: `project_management_automation/tools/consolidated_automation.py` (`task_workflow`), `todo_sync.py` (`sync_todo_tasks`)
- (No agentic-tools MCP in this repo; `auto_update_task_status.py` / `agentic_tools_client` are not present — see "Do we need agentic-tools MCP?" above.)
