# Next Native Go Migration Steps

**Date:** 2026-01-27  
**Source:** NATIVE_GO_MIGRATION_PLAN.md, PYTHON_FALLBACKS_SAFE_TO_REMOVE.md

---

## 1. High priority: testing and validation

- Unit tests for all native implementations
- Integration tests for tools/resources via MCP
- Regression tests for hybrid paths (native first, then bridge)
- Migration testing checklist in docs

## 2. Reduce automation Python use

| Where | Status |
|-------|--------|
| `automation_native.go` ~327 (sprint) | Done: uses `runDailyTask(ctx, "report", {"action": "overview"})` (native). |

## 3. Medium: shrink Python fallback surface

| Tool | Next step |
|------|-----------|
| **report** | Done: overview, briefing (native only, no fallback), scorecard (Go only; non-Go returns clear error). |
| **analyze_alignment** | Done: native `action=prd` (persona alignment). See scope below. |
| **estimation** | Done: native only, no Python fallback. |
| **task_analysis** | Done: fully native (FM provider abstraction; no Python fallback). |
| **task_discovery** | Done: native actions + create_tasks/output_path (CGO and nocgo). See scope below. |

## 4. Deferred / intentional

- **task_workflow** `external=true` → bridge required (agentic-tools)
- **mlx** → bridge-only by design
- **lint** → hybrid by design (Go native, others bridge)
- memory semantic search, testing suggest/generate, security non-Go, recommend advisor

## 5. Docs

- **Migration testing checklist:** `docs/MIGRATION_TESTING_CHECKLIST.md` (when to add tests, when to update toolsWithNoBridge)
- Hybrid pattern guide (when native-only vs native+fallback)
- Bridge call map (handlers, automation, task_workflow_common)
- **2026-01-28:** `NATIVE_GO_HANDLER_STATUS.md` updated with full-native/hybrid/bridge-only lists and explicit **4 removed Python fallbacks** (`setup_hooks`, `check_attribution`, `session`, `memory_maint`). See `PYTHON_FALLBACKS_SAFE_TO_REMOVE.md`.

---

**Order (current):** 1) Testing/validation checklist ✅; 2) report overview in automation ✅; 3) analyze_alignment prd ✅; 4) report briefing/scorecard + estimation shrink ✅ (2026-01-28); 5) task_analysis fully native ✅ (2026-01-28); 6) task_discovery native-only ✅ (scope doc updated 2026-01-28; bridge fallback removed). **4 removed Python fallbacks (2026-01-27):** `setup_hooks`, `check_attribution`, `session`, `memory_maint`. **Logging consolidation (2026-01-29):** Single facade `internal/logging` wrapping mcp-go-core (slog); main, database, CLI use shared logger. **Next:** testing/validation (unit tests for native impls).

---

## Scope: analyze_alignment (native, no Python)

**Actions (both native Go):**

| Action | Purpose | Implementation |
|--------|---------|----------------|
| **todo2** | Align Todo2 tasks with PROJECT_GOALS.md; report goals coverage and gaps | `handleAlignmentTodo2` in `internal/tools/alignment_analysis.go` |
| **prd** | Task-to-persona alignment using `docs/PRD.md` and persona keywords; outputs alignment_by_persona, unaligned_tasks, recommendations | `handleAlignmentPRD` in `internal/tools/alignment_analysis.go` |

**Usage:** Automation daily workflow calls `analyze_alignment` with `action=todo2`. No Python bridge; handler uses `handleAnalyzeAlignmentNative` only. Bridge does not route this tool.

---

## Pre-push hook and non-JSON stdin (fixed 2026-01-28, hardened 2026-01-29)

Git’s pre-push hook runs with refs on stdin (e.g. `refs/heads/main ...`). If exarp-go is invoked as `exarp-go analyze_alignment action=todo2` without a TTY, it used to start in **MCP server mode** and try to read stdin as JSON-RPC, causing: `invalid character 'r' looking for beginning of value`.

**Fixes:** (1) **`cmd/server/main.go`:** When the first argument is a single word (not `task`/`config`/`tui`/`tui3270` and not starting with `-`), it is treated as a **tool name** and the rest as `key=value` args. `os.Args` is rewritten to `-tool <name> -args <json>`, so the process runs in **CLI mode** and never parses stdin as JSON. (2) **2026-01-29:** When `GIT_HOOK=1` and no CLI flags are present, exit with a usage message instead of starting MCP. (3) **2026-01-29:** Installed hook scripts invoke exarp-go with `</dev/null` so stdin is never git refs.

**If the hook still fails:** Ensure it uses the **rebuilt** binary (e.g. `make build` then run `git push`, or set the hook to call `$(git rev-parse --show-toplevel)/bin/exarp-go` explicitly). Re-run `setup_hooks action=git` to get the latest hook scripts with `</dev/null`.

---

## Scope: task_discovery (native only, no Python fallback)

**Handler:** Uses `handleTaskDiscoveryNative` only. On error returns the error; **no** `bridge.ExecutePythonTool` call. Bridge does not route `task_discovery` (removed 2026-01-28).

**Native actions (both CGO and nocgo builds):**

| Action | Purpose | Implementation |
|--------|---------|----------------|
| **comments** | Scan code for TODO/FIXME/XXX/HACK/NOTE | CGO: `scanComments` (Apple FM optional); nocgo: `scanCommentsBasic` |
| **markdown** | Scan markdown for task-like items | CGO: `scanMarkdown`; nocgo: `scanMarkdownBasic` |
| **orphans** | Find tasks in docs not in Todo2 | CGO: `findOrphanTasks`; nocgo: `findOrphanTasksBasic` |
| **git_json** | Scan git for JSON task files | Shared: `scanGitJSON` |
| **planning_links** | Scan planning docs for task/epic links | CGO: `scanPlanningDocs`; nocgo: `scanPlanningDocsBasic` |
| **all** | Run all of the above | Both builds |

**Parameters (native):** `action`, `file_patterns`, `include_fixme`, `doc_path`, `json_pattern`, `create_tasks`, `output_path`.

- **create_tasks:** If true, native creates Todo2 tasks from discoveries via shared `createTasksFromDiscoveries` (in `task_discovery_common.go`). Both CGO and nocgo builds support it.
- **output_path:** If set, native writes full result JSON to the path (under project root). Both CGO and nocgo support it.

**Automation:** `automation_native.go` discover step uses native only (calls `handleTaskDiscoveryNative`). No bridge fallback.
