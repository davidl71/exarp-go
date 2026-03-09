## Codex Quickstart

This is the shortest useful entrypoint for Codex or other coding agents working in `exarp-go`.

### Start Here

1. Read [AGENTS.md](../AGENTS.md) for repo-specific working rules.
2. Use `make codex-smoke` for a fast verification pass after changes.
3. Prefer `make` targets over ad hoc build/test/lint commands when available.

### What To Ignore

Avoid spending context on local state, generated output, and archives unless the task is specifically about them:

- `.todo2/`
- `.exarp/`
- `.cache/`
- `bin/`
- `dist/`
- `vendor/`
- `docs/archive/`
- `.cursor/plans/`

These are already excluded from Cursor indexing via `.cursorignore`, and most are also gitignored.

### Artifact Layout

- `bin/` for local developer binaries such as `bin/exarp-go`
- `dist/` for packaged or release-style artifacts
- `out/` for generated one-off outputs and reports
- `out/` is gitignored, but agents should read it when recent tool output is relevant to the current task
- Do not leave compiled binaries at repo root; `server`, `migrate`, and similar root-level artifacts are legacy leftovers

### Preferred Commands

```bash
make codex-smoke
make test
make lint
make fmt
make go-build
```

### exarp-go Shortcuts

Use the convenience CLI first for common task operations:

```bash
exarp-go task list --status "Todo"
exarp-go task show T-123
exarp-go task update T-123 --new-status "Done"
```

For richer or project-aware operations, prefer exarp-go MCP tools/resources:

- `report` with `action=overview|scorecard|briefing`
- `session` with `action=prime`
- `task_workflow` for advanced task operations
- `stdio://tools` for the full tool catalog
- `stdio://suggested-tasks` for dependency-ready work

### Do Not

- Do not edit `.todo2/state.todo2.json` or `.todo2/todo2.db` directly.
- Do not use `exarp-go --help` to discover capabilities; prefer `make help`, `stdio://tools`, or `tool_catalog`.
- Do not revert unrelated local changes.

### When To Read More

- [README.md](../README.md) for repo overview and install/run commands
- [docs/README.md](./README.md) for the docs index
- [skills/README.md](../skills/README.md) for bundled Codex/Cursor skills
