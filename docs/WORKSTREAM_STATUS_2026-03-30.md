# Workstream status (2026-03-30)

Snapshot of **where related work stands** after the MCP JSON / protobuf discussion and the journal vs concurrency question. For the execution-cockpit product direction, see **`EXARP_EXECUTION_COCKPIT_GAPS.md`** (human-oriented gaps and priorities).

## Journal and concurrency (exarp / Todo2)

- **SQLite:** WAL is already enabled for the task store (see `internal/database/driver_sqlite.go` and health checks that report `journal_mode`).
- **An append-only application journal** (separate table or event log) is **optional** for audit, replay, or projections. It **does not replace** leases, optimistic versioning, or claim semantics for multi-agent correctness.
- **No dedicated “journal epic”** was scheduled in this slice; treat the above as the design stance until a concrete audit/replay requirement appears.

## MCP: JSON on the wire (mcp-go-core + exarp-go)

| Area | Status |
|------|--------|
| **Normalize stdio tool results to JSON-shaped text** | Implemented in **`mcp-go-core`**: `pkg/mcp/framework/adapters/gosdk/mcp_json_result.go` + tests; wired from `adapter.go` via middleware after `WrapToolHandler`. |
| **Protobuf** | Still fine **inside** handlers/adapters; **MCP content** is normalized to JSON strings for clients that expect structured parsing. |
| **exarp-go defaults** | `task_workflow` and related tool schemas default **`output_format: json`** where applicable (`registry_core.go`, list/maintenance paths, task_analysis deps/ownership). |
| **CLI / in-process CallTool** | Callers that invoke the handler **directly** (not through the stdio stack) may still see **raw** text when they pass `output_format: text`; only the MCP stdio path is wrapped by the new middleware. |

**Repo state (uncommitted):** see `git status` under `mcp-go-core` (adapter + new files) and `exarp-go` (tool/registry + database + normalization edits).

## exarp-go database and tools (same session)

Local changes **not yet committed** (check `git status` for exact set) include:

- **Lock cleanup:** batched `UPDATE`s where possible (`lock_monitoring.go` and related).
- **GetTask / claim path:** fewer round-trips for tags and dependencies (`tasks_list.go`, `tasks_crud.go`, helpers).
- **Comments:** multi-row insert in one statement inside the existing transaction (`comments.go`).
- **Models / normalization:** optional version on `Todo2Task` and plumbing for skipping redundant pre-update reads where applicable.
- **Task tooling:** `output_format` defaults, normalization tests, task_analysis lookup/helper refactors, optional benches.

**Todo2 note:** If you completed work that matches **database performance** tasks still marked `Todo` in the DB, run **`task sync`** (or equivalent) so SQLite and `.todo2/state.todo2.json` stay aligned.

## Execution cockpit (product, not infra)

exarp-go remains strong as a **backlog and planning** tool. Gaps for **active coding sessions** (session focus, structured progress, fast child tasks, execution logs) are documented in **`EXARP_EXECUTION_COCKPIT_GAPS.md`**. No single “cockpit v1” shipped in this slice; that doc is still the source of truth for next product steps.

## Aether / multi-repo backlog (high level)

When `PROJECT_ROOT` points at the Aether repo, **task_workflow** lists show a large **Todo** surface: TUI UX foundation (modes, discoverability, feedback), composed **market/operations** workspaces, **NATS health** metrics in snapshot/TUI, **yield curve** pipeline (mock provider, optional Polygon/Yahoo), **CSV** import/export, **disk pressure** / target prune workflows, **charts** (`risk_metrics` MVP), **RUSTSEC** hygiene tasks, etc. Those items are **outstanding** unless individually marked Done in Todo2.

Treat this section as **pointer-only**; authoritative status is always the current Todo2 store for that project root.

## Suggested next verification

1. From **`mcp-go-core`:** run **`make test`** (or project test target) after JSON middleware changes.
2. From **`exarp-go`:** run **`make test-go`** / **`make lint`** after DB and tool changes.
3. Commit **mcp-go-core** first or keep **`replace`** in `exarp-go/go.mod` until both land.
4. Run **`./scripts/run_exarp_go.sh task sync`** in each repo that uses Todo2 after bulk status edits.
