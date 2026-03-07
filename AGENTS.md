# exarp-go Agent Guide (Codex)

Use these rules when editing this repository with Codex.

## Core Rules

1. Use `make` targets for build/test/lint workflows when available.
2. Prefer `rg`/`rg --files` for search and discovery.
3. Do not edit `.todo2/state.todo2.json` or `.todo2/todo2.db` directly. Use `exarp-go task ...` or MCP tools.
4. Keep changes scoped; do not revert unrelated local modifications.
5. Prefer updating docs/examples when changing config behavior.

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
