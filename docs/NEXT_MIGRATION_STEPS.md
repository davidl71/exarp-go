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

---

**Order (current):** 1) Testing/validation checklist ✅; 2) report overview in automation ✅; 3) analyze_alignment prd ✅; 4) report briefing/scorecard + estimation shrink ✅ (2026-01-28); 5) task_analysis fully native ✅ (2026-01-28); 6) task_discovery CGO parity + scope doc ✅ (create_tasks/output_path shared; NEXT_MIGRATION_STEPS scope). **Next:** testing/validation (unit tests for native impls); optional: remove task_discovery Python fallback for native-only.

---

## Scope: analyze_alignment (native, no Python)

**Actions (both native Go):**

| Action | Purpose | Implementation |
|--------|---------|----------------|
| **todo2** | Align Todo2 tasks with PROJECT_GOALS.md; report goals coverage and gaps | `handleAlignmentTodo2` in `internal/tools/alignment_analysis.go` |
| **prd** | Task-to-persona alignment using `docs/PRD.md` and persona keywords; outputs alignment_by_persona, unaligned_tasks, recommendations | `handleAlignmentPRD` in `internal/tools/alignment_analysis.go` |

**Usage:** Automation daily workflow calls `analyze_alignment` with `action=todo2`. No Python bridge; handler uses `handleAnalyzeAlignmentNative` only. Bridge does not route this tool.

---

## Scope: task_discovery (native first, bridge fallback on error)

**Handler:** Tries `handleTaskDiscoveryNative` first; on error falls back to `bridge.ExecutePythonTool(ctx, "task_discovery", params)`.

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

**Automation:** `automation_native.go` discover step uses native only (calls `handleTaskDiscoveryNative` via handler; no direct bridge call). If native fails, the handler falls back to the Python bridge.

**Bridge:** `bridge/execute_tool.py` still routes `task_discovery` when the handler invokes the fallback (e.g. native error on a given platform). Removing the fallback would make the tool native-only; current design keeps fallback for robustness.
