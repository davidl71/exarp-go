# Enum and fixed-vocabulary candidates

This document indexes **stringly-typed tool and config options** that are good candidates for shared `ParseX() (T, error)` helpers, typed constants, and tests—similar to [`internal/conflictpolicy`](./CONFLICT_POLICY.md) for conflict modes.

Design choice (implementation phase): either a small **`internal/toolparams`** (or **`internal/enums`**) package with **no** import of `internal/tools` (to avoid cycles), or colocate per-domain packages. Prefer **one parse function per concept** plus unit tests over a single mega-enum type.

## Tier A — policy / mode strings (high leverage)

| Concept | Allowed values | Primary code | Registry / schema |
|--------|----------------|--------------|-------------------|
| **Import scan mode** | `none`, `immediate`, `recursive` | `resolveImportSQLitePaths` in [`internal/tools/task_workflow_import_sqlite.go`](../internal/tools/task_workflow_import_sqlite.go) | `import_scan_mode` in [`internal/tools/registry_core.go`](../internal/tools/registry_core.go) |
| **Split dependency layout** | `parallel`, `serial` (default `parallel`) | [`internal/tools/task_workflow_execution.go`](../internal/tools/task_workflow_execution.go) (`dependency_mode`) | task_workflow `split` parameter docs |
| **Handoff restore strategy** | `merge`, `replace` (default `merge`) | [`internal/tools/session_handoff.go`](../internal/tools/session_handoff.go) | session `restore` / handoff flow |
| **Comment type** | `result`, `note`, `research_with_links`, `manualsetup` | [`internal/tools/task_workflow_actions.go`](../internal/tools/task_workflow_actions.go) (`ParamEnum` on `comment_type`) | `comment_type` in `registry_core.go` |

**Overlap with conflict policy:** [`import_on_conflict`](./CONFLICT_POLICY.md) (`fail` \| `skip`) and git merge **`newer` \| `source` \| `target`** are covered by `internal/conflictpolicy` design, not this list’s primary scope—keep parsers composable so `ParseMode` can stay conflict-focused or be reused where values intersect.

## Tier B — proto-backed actions (duplication / drift risk)

Large **`action`** dispatch sets (e.g. `task_workflow`, `report`, `security`, `testing`) already have **protobuf enums** in places; remaining work is usually **one canonical list** for errors, MCP schema enums, and switches (avoid three divergent string slices).

## Tier C — mapping vocabulary

| Concept | Notes |
|--------|--------|
| **Plan file ↔ Todo2 status** | [`internal/tools/plan_sync.go`](../internal/tools/plan_sync.go): `cursorStatusToTodo2`, `todo2StatusToPlanStatus` — candidate for typed plan status + tests. |
| **Default status / priority (config)** | [`internal/cli/config.go`](../internal/cli/config.go) `validateTaskValue` vs task `ParamEnum` in create/update — **Blocked** is in task UI but omitted from config validator today; reconcile when introducing shared types. |

## Tier D — optional tighten-ups

- **`verify` `kind`** — descriptive, not a closed set in code; low priority unless you want a closed enum.
- **Lint `linter`** — fixed set in [`internal/tools/handlers_ai.go`](../internal/tools/handlers_ai.go); candidate if you standardize “parse tool string” patterns.
- **`workflow_mode` modes** — [`internal/tools/workflow_mode.go`](../internal/tools/workflow_mode.go): `workflowValidModes` map; optional `ParseWorkflowMode`.

## Implementation notes

- **CLI / MCP:** `task_workflow` args pass through **proto JSON**; avoid `#tag` in JSON strings (protojson). Use JSON **arrays** for `tags` in raw `-args` if needed.
- **Tests:** Each new `Parse*` should have table-driven tests: case insensitivity, whitespace, default vs invalid.

## References

- [`docs/CONFLICT_POLICY.md`](./CONFLICT_POLICY.md) — conflict modes and `internal/conflictpolicy`.
- [`docs/FACTORY_CANDIDATES.md`](./FACTORY_CANDIDATES.md) — `New*` / registry consolidation (LLM backends, agent registry, etc.).
- [`docs/tool-parameter-parsing.md`](./tool-parameter-parsing.md) — `ParamEnum` and handler patterns.

## Todo2 tracking (exarp-go project)

| Workstream | Task ID |
|-----------|---------|
| ImportScanMode + import_sqlite | `T-1775395233453527000` |
| SplitDependencyMode + split | `T-1775395233478087000` |
| HandoffRestoreStrategy + restore | `T-1775395233502101000` |
| CommentType + add_comment | `T-1775395233526678000` |
| Optional: plan status + config/task alignment | `T-1775395233551761000` |
