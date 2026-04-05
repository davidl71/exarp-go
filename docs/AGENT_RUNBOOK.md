# Agent runbook (exarp-go)

Short operational reference for humans and coding agents working **with** exarp-go—CLI, MCP, and common failure modes.

## Build the stdio server

```bash
go build -o bin/exarp-go ./cmd/server
```

Running `go build .` at the repository root builds the wrong `main` package. Prefer `make go-build` or the line above.

## Environment check

```bash
./bin/exarp-go doctor
# or, from another repo with exarp on PATH:
exarp-go doctor
```

Expect: `project_root`, `.todo2/todo2.db` stat, optional `parallel-execution-waves.json`, `EXARP_MIGRATIONS_DIR`, binary path, optional centralized config, and `task_rows_in_db` when the DB initializes.

## MCP / `PROJECT_ROOT`

The server resolves tasks and Todo2 state under **`PROJECT_ROOT`** (env in Cursor, Codex, OpenCode, etc.). If task lists look empty or wrong, verify **`exarp-go doctor`** matches the repo you have open.

## `task_workflow` (MCP tool)

| Situation | What to do |
|-----------|------------|
| List tasks as JSON | `action=list`, `output_format=json` — each task includes **`priority_rank`**, **`dependencies`**, **`version`**. |
| Update “succeeds” but nothing changes | Read **`update_issues`**; **`success`** may be false with `no tasks updated; see update_issues`. |
| Set priority to zero | **`priority_rank: 0`** is valid—do not strip zero numerics when serializing JSON. |
| Discovery | Use **stdio://tools** or **`tool_catalog`** — not `exarp-go --help`. |

## CLI: parallel waves

Requires **`.cursor/plans/parallel-execution-waves.json`** under the project root (see `task wave` help).

```bash
exarp-go task wave ids 0
exarp-go task wave remaining 0 --batch 15 --json
```

`remaining` intersects wave IDs with **open** statuses (Todo, In Progress, Blocked per `OpenStatuses()`).

## Verification

```bash
go test ./...
# or
make test
```

Long-running automation subtests use extended timeouts; use **`go test -short ./...`** to skip some of them.

## Further reading

- Repo root **[AGENTS.md](../AGENTS.md)** — contributor rules.
- **[docs/EXARP_ABILITIES_AUDIT.md](EXARP_ABILITIES_AUDIT.md)** — tool surface (if present in tree).
