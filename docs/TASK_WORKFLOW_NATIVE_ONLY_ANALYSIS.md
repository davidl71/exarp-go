# task_workflow Native-Only Analysis

**Question:** Can we make `task_workflow` native-only? What would we lose?

---

## Where the bridge is used today

1. **Handler fallback** (`handlers.go`): When native returns an error containing  
   `"Apple Foundation Models not supported"` / `"not available on this platform"` /  
   `"requires Apple Foundation Models"` → we fall back to the Python bridge.

2. **External sync** (`task_workflow_common.go`): When `external=true`, we **always** call the bridge for `action=sync` instead of the native SQLite↔JSON sync.

---

## What we’d lose if we go native-only

### 1. Clarify on non–Apple-FM platforms

- **Today:** On Linux, Windows, or nocgo builds, `action=clarify` fails natively  
  (`"clarify action requires Apple Foundation Models (not available on this platform)"`).  
  We then fall back to Python, which can use **Ollama** (or another LLM) for clarification.
- **Native-only:** No fallback. `clarify` would **always error** on those platforms.
- **Impact:** Users without Apple FM lose `task_workflow(action=clarify)` unless we add another native LLM path (e.g. Ollama) for clarify.

### 2. `sync` with `external=true` (agentic-tools, etc.)

- **Today:** We short‑circuit and call the bridge. The bridge invokes Python `task_workflow`, but it **does not pass `external`** in the args (see `bridge/execute_tool.py`). The Python implementation also has **no `external` parameter** and only does `sync_todo_tasks` (shared TODO table ↔ Todo2).
- **Reality:** “External sync” (e.g. agentic-tools, GitHub Issues, Google Tasks) is **not actually implemented** end‑to‑end via the current bridge. We route to the bridge, but the Python side ignores `external` and does normal sync.
- **Impact:** Going native-only would not remove a working external-sync path. We’d only stop routing `external=true` to a bridge that doesn’t use it. We could return a clear error, e.g.  
  `"external sync (agentic-tools) not supported; use sync without external or a future native implementation"`.

---

## What stays the same (native today)

These **already work** in native Go on all platforms:

- `sync` (without `external`): SQLite ↔ JSON via `handleTaskWorkflowSync` → `SyncTodo2Tasks` / `handleTaskWorkflowList`.
- `approve`, `create`, `clarity`, `cleanup`: Fully native.

On **darwin/arm64/cgo** (Apple FM), `clarify` is also native.

---

## Recommendation

- **Safe to make native-only from a “external sync” perspective:** Yes. The bridge never forwards `external`, and Python doesn’t support it. We can remove the `external=true` → bridge path and, for now, return an explicit “external sync not supported” error when `external=true`.
- **Clarify:** We **would** lose the Python fallback for `clarify` on non–Apple-FM platforms. Options:
  1. **Accept the loss:** `clarify` errors there; document that Apple FM (or a future native LLM) is required.
  2. **Add a native clarify path** using Ollama (or another provider) on non–Apple-FM platforms, then remove the bridge fallback.

---

## Concrete steps to go native-only

1. **Handlers**
   - In `handleTaskWorkflow`, remove the bridge fallback. Always `return nil, err` when native fails (including “Apple Foundation Models not supported”).
   - In `handleTaskWorkflowSync`, when `external=true`, return a clear error instead of calling the bridge, e.g.  
     `"external task sync (agentic-tools) not implemented in native mode; use sync without external"`.

2. **Bridge**
   - Remove the `task_workflow` branch from `bridge/execute_tool.py` (and any now‑unused imports).

3. **Tests / docs**
   - Update `toolsWithNoBridge`, regression tests, and migration docs to treat `task_workflow` as native-only.
   - Document that `clarify` requires Apple FM (or a future native LLM) and that `external=true` is unsupported.

4. **Optional later**
   - Implement external sync in Go via `internal/tasksync` (GitHub, Google Tasks, etc.) and support `external=true` natively.
   - Add a native `clarify` path using Ollama (or similar) for non–Apple-FM platforms.
