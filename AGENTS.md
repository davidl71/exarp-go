# exarp-go Agent Guide (Codex)

Use these rules when editing this repository with Codex.

## Core Rules

1. Use `make` targets for build/test/lint workflows when available.
2. Prefer `rg`/`rg --files` for search and discovery.
3. Do not edit `.todo2/state.todo2.json` or `.todo2/todo2.db` directly. Use `exarp-go task ...` or MCP tools.
4. Keep changes scoped; do not revert unrelated local modifications.
5. Prefer updating docs/examples when changing config behavior.
6. When touching `task_discovery` build-tagged files, read **[docs/CGO_BUILD_PARITY.md](docs/CGO_BUILD_PARITY.md)** (`make build` uses `CGO_ENABLED=0`; `make go-build` enables CGO on Apple Silicon when a C compiler exists). Estimation/context entrypoints are unified in `estimation_shared_v2.go` and `context_shared.go`.

## MCP and Project Root

- exarp-go MCP server command for this repo: `run-exarp-go.sh`
- Use `PROJECT_ROOT` as this repo root for task/session/report flows.
- Cursor/OpenCode configs exist in:
  - `.cursor/mcp.json`
  - `mcp.json`
  - `opencode.json`

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

Use `task_workflow` tool for advanced actions (clarity, cleanup, complex filters, batch JSON flows).

## Reporting

- Use `report` action `overview` for project snapshot.
- Use `report` action `scorecard` for quality/status rollup.
- Use `report` action `briefing` for short standup/handoff context.
