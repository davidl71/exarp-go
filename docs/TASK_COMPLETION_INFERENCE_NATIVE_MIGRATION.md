# Task Completion Inference: Native Go Migration Plan

Task completion inference is core functionality for exarp-go. This document plans the migration from the current Python/agentic-tools MCP path to a **native Go** implementation so `make check-tasks` and `make update-completed-tasks` work without external MCP or Python.

---

## 1. Current Behavior (Python / agentic-tools)

| Aspect | Detail |
|--------|--------|
| **Entry** | `make check-tasks` (dry run), `make update-completed-tasks` (apply) |
| **Implementation** | `project_management_automation/tools/auto_update_task_status.py` → `agentic_tools_client.infer_task_progress_mcp()` |
| **Backend** | agentic-tools MCP server, tool `infer_task_progress` |
| **Inputs** | `workingDirectory` (required), `projectId`, `scanDepth` (1–5), `fileExtensions`, `autoUpdateTasks`, `confidenceThreshold` (default 0.7) |
| **Outputs** | `total_tasks_analyzed`, `inferences_made`, `tasks_updated`, `inferred_results`: `[{ task_id, confidence, evidence[] }]` |
| **Report** | Optional markdown at `docs/TASK_COMPLETION_CHECK.md` or `docs/TASK_COMPLETION_UPDATE.md` |

Behavior: analyze **Todo** and **In Progress** tasks; for each, infer from codebase (files, changes, implementation evidence) whether the task is actually done; optionally set status to **Done** when confidence ≥ threshold.

---

## 2. Target Behavior (Native Go)

- **Same contract**: same logical inputs/outputs and report shape so callers (Makefile, scripts, MCP clients) can switch without changing behavior.
- **No Python**: implementation entirely in Go under `internal/tools`.
- **No agentic-tools MCP**: no dependency on external MCP server for this flow.
- **Database-first**: load Todo/In Progress via `database.ListTasks` (or `LoadTodo2Tasks` + filter); updates via `database.UpdateTask`.
- **Optional FM**: use `DefaultFMProvider()` when available for richer inference; when not, use heuristics-only so the tool still runs everywhere.

---

## 3. Design Overview

### 3.1 New MCP tool: `infer_task_progress`

- **Name**: `infer_task_progress` (parity with agentic-tools).
- **Handler**: e.g. `handleInferTaskProgress` in `internal/tools`; registered in `registry.go` like `task_analysis` / `task_workflow`.
- **Parameters** (aligned with current Python/agentic-tools):
  - `project_root` or derived from `FindProjectRoot()`.
  - `scan_depth` (1–5, default 3).
  - `file_extensions` (default: `.go`, `.py`, `.ts`, `.tsx`, `.js`, `.jsx`, etc.).
  - `confidence_threshold` (0–1, default 0.7).
  - `auto_update_tasks` (bool).
  - `dry_run` (bool; if true, do not update).
  - `output_path` (optional, e.g. `docs/TASK_COMPLETION_CHECK.md`).

### 3.2 Pipeline

1. **Load candidates**  
   Todo + In Progress tasks: `database.GetTasksByStatus(ctx, "Todo")` and `GetTasksByStatus(ctx, "In Progress")`, or single query with status IN; fallback to `LoadTodo2Tasks(projectRoot)` + filter if DB unavailable.

2. **Gather evidence**  
   For the repo at `project_root`:
   - **File tree**: walk up to `scan_depth`; restrict by `file_extensions` (reuse or mirror task_discovery-style patterns).
   - **Content signals**: path names, file contents (e.g. first N lines or full for small files), optional git log (e.g. recent commits touching paths). No need to replicate full agentic-tools logic; start with path + content snippets.

3. **Score each task**  
   - **Heuristics (always)**:
     - Keyword/string match: task `Content` / `LongDescription` vs file paths and content snippets (e.g. “Add X” vs file `x.go` or comment “X implemented”).
     - Simple rules: e.g. “Remove Python fallback for Y” → presence of only Go in a previously mixed area.
   - **FM (when available)**:
     - `DefaultFMProvider().Supported()`; if true, for each task (or batched) prompt: “Given the following codebase context and task description, is this task complete? Return JSON: {\"task_id\": \"T-1\", \"complete\": true|false, \"confidence\": 0.0-1.0, \"evidence\": [\"...\"]}.”
     - Parse JSON; merge into overall confidence (e.g. max(heuristic, fm) or average).
   - **Confidence**: combine heuristic + FM into a single 0–1 score per task.

4. **Filter and optionally update**  
   Tasks with `confidence >= confidence_threshold` are “inferred complete.”
   - If `dry_run` or `!auto_update_tasks`: return results only; no DB writes.
   - Else: for each inferred-complete task, set `Status = "Done"`, `CompletedAt = now`, persist via `database.UpdateTask`, and increment `tasks_updated`.

5. **Response and report**  
   - Return JSON: `success`, `total_tasks_analyzed`, `inferences_made`, `tasks_updated`, `inferred_results` (same shape as current).
   - If `output_path` set, write markdown report (same structure as current Python report).

### 3.3 Reuse and boundaries

- **database**: `ListTasks` / `GetTasksByStatus`, `UpdateTask`; reuse existing status constants (`StatusTodo`, `StatusInProgress`, `StatusDone`).
- **project root**: `FindProjectRoot()`.
- **FM**: `DefaultFMProvider()`, `FMAvailable()`; same pattern as `task_analysis` hierarchy / `task_workflow` clarify.
- **File walking**: either reuse helpers from `task_discovery` (e.g. `scanCommentsBasic`-style glob/walk) or add a small `infer_task_progress_evidence` package that does a shallow file walk + extension filter; keep scanning logic in one place.
- **Normalization**: reuse `normalizeStatus`, `IsCompletedStatus` where relevant.
- **Optional**: later, add `task_workflow` action `approve`-style batch (e.g. “mark these IDs Done”) so a single approval flow can be used; for MVP, direct `UpdateTask` is enough.

---

## 4. Phases and Tasks

### Phase 1: Core inference (heuristics only, no FM, no update)

- **1.1** Add `internal/tools/infer_task_progress.go` (or split: `infer_task_progress.go` + `infer_task_progress_evidence.go`).
- **1.2** Implement: load Todo + In Progress (DB-first, fallback LoadTodo2Tasks); gather evidence (file walk, extensions, path/content snippets); heuristic scoring (keyword/match rules); return `inferred_results` with confidence and evidence; no `UpdateTask`, no FM.
- **1.3** Register `infer_task_progress` in `registry.go` with schema (params above); add handler in `handlers.go`.
- **1.4** CLI: ensure `exarp-go -tool infer_task_progress -args '{"dry_run": true, ...}'` works.
- **1.5** Unit tests: mock tasks + minimal fs; assert confidence 0/1 and evidence shape.

**Exit criterion**: `make check-tasks` can be switched to call `exarp-go -tool infer_task_progress` with `dry_run: true` and get a valid report (heuristics-only).

### Phase 2: FM-based scoring (optional enhancement)

- **2.1** When `FMAvailable()`: for each task (or batch), call `DefaultFMProvider().Generate(ctx, prompt, maxTokens, temp)` with a structured prompt; parse JSON completion answer.
- **2.2** Combine heuristic score and FM score (e.g. max or weighted average); cap at 1.0.
- **2.3** Handle FM errors gracefully: fall back to heuristics-only for that task; never fail entire run.
- **2.4** Tests: with FM mock, assert combined confidence and evidence; without FM, assert heuristics-only path unchanged.

**Exit criterion**: When FM is available, inference quality improves; when not, behavior equals Phase 1.

### Phase 3: Auto-update and report

- **3.1** If `auto_update_tasks && !dry_run`: for each task in `inferred_results` with `confidence >= confidence_threshold`, load task, set `Status = "Done"`, `CompletedAt = now`, call `database.UpdateTask`.
- **3.2** Respect DB vs file fallback: if using `LoadTodo2Tasks`/`SaveTodo2Tasks`, apply same status update there so JSON and DB stay in sync where both are used.
- **3.3** Write markdown report to `output_path` when provided (same format as current Python).
- **3.4** Tests: dry run vs apply; assert `tasks_updated` and DB state.

**Exit criterion**: `make update-completed-tasks` can call `exarp-go -tool infer_task_progress` with `auto_update_tasks: true` and see tasks move to Done and report written.

### Phase 4: Makefile and deprecation

- **4.1** Change `make check-tasks` to run `$(BINARY_PATH) -tool infer_task_progress -args '{"dry_run": true, "output_path": "docs/TASK_COMPLETION_CHECK.md"}'` (and equivalent for `update-completed-tasks` with `auto_update_tasks: true`, `output_path`: `docs/TASK_COMPLETION_UPDATE.md`).
- **4.2** Keep Python path as fallback behind a flag or env (e.g. `USE_PYTHON_TASK_CHECK=1`) for one release, then remove.
- **4.3** Update `docs/TASK_COMPLETION_CHECK_NATIVE_GO.md`: state native Go is the default; document Python deprecation.
- **4.4** Update `.cursor/rules/go-development.mdc` and any other references to `check-tasks` / `update-completed-tasks` to describe native implementation.

**Exit criterion**: Default `make check-tasks` and `make update-completed-tasks` use native Go; no dependency on agentic-tools MCP or Python for this flow.

---

## 5. File and Package Layout

| File | Purpose |
|------|--------|
| `internal/tools/infer_task_progress.go` | Handler, params, load tasks, score (heuristic + FM), optional update, response + report write. |
| `internal/tools/infer_task_progress_evidence.go` | File walk, extension filter, path/content snippet collection (or reuse task_discovery helpers). |
| `internal/tools/handlers.go` | Wire `handleInferTaskProgress`. |
| `internal/tools/registry.go` | Register `infer_task_progress` tool and schema. |
| `internal/tools/protobuf_helpers.go` | Optional: `ParseInferTaskProgressRequest`, response mapping if we add proto. |
| `proto/tools.proto` | Optional: `InferTaskProgressRequest` / `InferTaskProgressResponse` for future. |
| `Makefile` | `check-tasks` / `update-completed-tasks` invoke exarp-go tool. |
| `docs/TASK_COMPLETION_CHECK_NATIVE_GO.md` | Update with migration status and usage. |

No new top-level package required; keep everything under `internal/tools` for consistency with `task_analysis`, `task_workflow`, `task_discovery`.

---

## 6. Config and Defaults

- **confidence_threshold**: default 0.7; overridable via tool args; later via `internal/config` if desired.
- **scan_depth**: default 3; max 5.
- **file_extensions**: default `[".go", ".py", ".ts", ".tsx", ".js", ".jsx", ".java", ".cs", ".rs"]` to match current Python.

---

## 7. Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Heuristics too weak | Phase 1 still delivers value (list + evidence); Phase 2 FM improves accuracy; we can add more rules (e.g. “file X exists” → task “Add X” complete). |
| FM latency/cost | Batch tasks in one prompt where possible; make FM optional; timeouts in context. |
| DB vs JSON drift | Prefer database when available; when updating, use same path as task_workflow (DB or SaveTodo2Tasks) so one source of truth. |
| Breaking callers | Keep JSON response shape and report format; Makefile change is the only caller change. |

---

## 8. Success Criteria

- **Functional**: `make check-tasks` and `make update-completed-tasks` produce the same logical outcome as today (dry-run report; optional status updates) without Python or agentic-tools MCP.
- **Performance**: Completion check runs in a few seconds for hundreds of tasks (heuristics); with FM, tens of seconds is acceptable.
- **Maintainability**: Single codebase in Go; no external MCP for this flow; tests for heuristic and FM paths and for dry_run vs apply.

---

## 9. Optional Follow-ups

- Proto types for `infer_task_progress` for consistency with other tools.
- `task_workflow` action that delegates to infer_task_progress (e.g. `action=infer_completion`) for MCP-only callers.
- Tuning: more evidence sources (e.g. git blame, last-modified), or configurable confidence per tag/priority.

This plan treats task completion inference as core and migrates it to native Go in four phases with clear exit criteria and minimal breaking change for existing usage.
