# exarp-go Agent Guide (Cursor / Codex / OpenCode / Claude)

Use these rules when editing this repository with any AI agent that drives exarp-go.

## Core Rules

1. Use `make` targets for build/test/lint workflows when available.
2. Prefer `rg`/`rg --files` for search and discovery.
3. Do not edit `.todo2/state.todo2.json` or `.todo2/todo2.db` directly. Use `exarp-go task ...` or MCP tools.
4. Keep changes scoped; do not revert unrelated local modifications.
5. Prefer updating docs/examples when changing config behavior.
6. When touching `task_discovery` build-tagged files, read **[docs/CGO_BUILD_PARITY.md](docs/CGO_BUILD_PARITY.md)** (`make build` uses `CGO_ENABLED=0`; `make go-build` enables CGO on Apple Silicon when a C compiler exists). Estimation/context entrypoints are unified in `estimation_shared_v2.go` and `context_shared.go`.

## Build and MCP

- **Build the stdio server:** `go build -o bin/exarp-go ./cmd/server` (do **not** use `go build .` at repo root — wrong `main`).
- MCP command for this repo: **`bin/exarp-go`** or **`run_server.sh`** / **`start.sh`** (wrappers that build if needed).
- Use **`PROJECT_ROOT`** as this repo root (or the **target** app repo) for task/session/report flows.
- **Diagnostics:** `exarp-go doctor` — prints root, DB path, migrations hint, binary path, optional wave-plan file, task count.
- **Parallel waves (CLI):** `exarp-go task wave ids N` / `task wave remaining N` — reads `.cursor/plans/parallel-execution-waves.json` under `PROJECT_ROOT`.

Example configs: `.cursor/mcp.json`, `mcp.json`, `opencode.json`, `~/.codex/config.toml` (`mcp_servers.exarp-go`).

## Recommended Commands

```bash
make test
make lint
make fmt
make go-build
```

## Shell Commands

- Do NOT use `timeout` command (GNU-only, not available on Mac). Use `gtimeout` from coreutils if installed, or avoid timeouts in commands.
- Use `$(go run ./cmd/...)` instead of compiled binaries when testing changes.

## Task Workflow

Use convenience CLI first:

```bash
exarp-go task list --status "Todo"
exarp-go task show T-123
exarp-go task update T-123 --new-status "Done"
```

Use `task_workflow` MCP tool for advanced actions (clarity, cleanup, complex filters, batch JSON flows).

**`task_workflow` semantics:**

- **`action=list`** with `output_format=json`: each task always includes `priority_rank`, `dependencies` (array), `version`.
- **`action=update`:** if nothing updates, check **`update_issues`**; response may set `success: false` with `no tasks updated; see update_issues`. **`priority_rank`** may be **0** — preserve zero in JSON.

Tool discovery: **stdio://tools**, **tool_catalog** `action=help` — not `exarp-go --help`.

Operational quick reference: **[docs/AGENT_RUNBOOK.md](docs/AGENT_RUNBOOK.md)**.

## Reporting

- Use `report` action `overview` for project snapshot.
- Use `report` action `scorecard` for quality/status rollup.
- Use `report` action `briefing` for short standup/handoff context.
