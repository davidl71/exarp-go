# exarp-go — Claude Code Guide

## Project

Go-based MCP server. 37 tools (38 with Apple FM), 36 prompts, 24 resources. Primary language: Go. Also: shell scripts (scripts/, ansible/), Ansible (YAML playbooks/roles in ansible/). SQLite-backed task system (Todo2). Apple Foundation Models + Ollama + MLX for local AI.

## MCP servers available in this session

| Server | Role |
|--------|------|
| `exarp-go` | Executor — task mgmt, reports, health, session, testing |
| `tractatus_thinking` | Analyst — logical decomposition |
| `context7` | Researcher — library docs |
| `devwisdom` | Guidance — wisdom, advisors |

## Session start (always)

Call `session` tool with `action=prime`, `include_hints=true`, `include_tasks=true` at the start of every session to get current task context, handoffs, and mode hints.

## Quick reference — common agent operations

These are the most frequent operations. Use CLI for terminal ops; use MCP tools when orchestrating via tool calls.

### Prime session / get context
```
# MCP (preferred)
session  action=prime  include_hints=true  include_tasks=true

# CLI
go run ./cmd/server session prime 2>/dev/null
```

### View backlog (Todo tasks)
```
# CLI
go run ./cmd/server task list 2>/dev/null                          # top 17 Todo tasks
go run ./cmd/server task list --status Todo 2>/dev/null            # same
go run ./cmd/server task list --priority high 2>/dev/null          # high-priority only
go run ./cmd/server task list --status "In Progress" 2>/dev/null   # active tasks

# MCP — use action=list (NOT action=sync, which does a full SQLite↔JSON sync)
task_workflow  action=list
task_workflow  action=list  status=Todo
task_workflow  action=list  priority=high
task_workflow  action=list  status="In Progress"
```

### Show task detail
```
go run ./cmd/server task show T-xxx 2>/dev/null
```

### Create a task
```
# CLI — task name as first positional arg; flags come after
go run ./cmd/server task create "Task name here" 2>/dev/null
go run ./cmd/server task create "Task name here" --priority high 2>/dev/null
go run ./cmd/server task create "Task name here" --priority medium --tags "tag1,tag2" 2>/dev/null

# MCP
task_workflow  action=create  name="Task name"  priority=high
```

### Update a task
```
go run ./cmd/server task update T-xxx --new-status "In Progress" 2>/dev/null
go run ./cmd/server task update T-xxx --new-status Done 2>/dev/null
go run ./cmd/server task update T-xxx --new-priority high 2>/dev/null
go run ./cmd/server task update T-xxx --name "Corrected name" 2>/dev/null
go run ./cmd/server task update T-xxx --description "More detail" 2>/dev/null

# MCP
task_workflow  action=update  task_id=T-xxx  new_status=Done
```

### Run AI on a task
```
go run ./cmd/server task run-with-ai T-xxx --backend ollama 2>/dev/null
go run ./cmd/server task summarize T-xxx 2>/dev/null
go run ./cmd/server task estimate "Task name" --local-ai-backend fm 2>/dev/null
```

> **Tip**: pipe through `2>/dev/null` in CLI commands to suppress log output in scripts.

## Task management (Todo2)

**Prefer MCP tools over direct file edits. Never edit `.todo2/state.todo2.json` or `.todo2/todo2.db` directly.**

```
# MCP tool (preferred)
task_workflow  action=list|create|update|delete|summarize|run_with_ai|sync|clarify

# action=list   — read-only task listing (use this, not sync, for viewing tasks)
# action=sync   — bidirectional SQLite↔JSON reconciliation (maintenance only)
```

Task statuses: `Todo` → `In Progress` → `Review` → `Done`

Local AI task subcommands: `task estimate`, `task summarize`, `task run-with-ai`; each supports `--local-ai-backend` or `--backend` (fm|mlx|ollama). `task create` and `task update` accept `--local-ai-backend` to set preferred backend.

## Build

```bash
make b              # build (short alias)
make build-apple-fm # with Apple Foundation Models (CGO, darwin/arm64)
make test           # all tests
make test-go        # Go tests only
make lint           # lint
make fmt            # format (NEVER run gofmt directly)
make tidy           # go mod tidy
make scorecard-fix  # auto-fix all fixable scorecard issues (tidy + fmt + lint-fix)
make sanity-check   # verify tool/prompt/resource counts
```

Binary: `bin/exarp-go` (project) — NOT `/Users/davidl/go/bin/exarp-go` (stale system install).
Use `go run ./cmd/server ...` for CLI ops during development.

**NEVER run go build, go test, go fmt, gofmt, or golangci-lint directly — always use make targets.**

## Key patterns

- **Error handling**: always `fmt.Errorf("context: %w", err)` — never ignore errors
- **Task store**: use `getTaskStore(ctx)` — never load JSON/DB directly in tool handlers
- **Preferred backend**: stored in `task.Metadata["preferred_backend"]` (fm|mlx|ollama); read with `GetPreferredBackend(task.Metadata)`
- **New task_workflow actions**: add to switch in `task_workflow_native.go`, handler in `task_workflow_actions.go` or `task_workflow_common.go`, enum in `registry.go`
- **Count sync**: when adding tools/prompts/resources, update comment + test assertions + expected lists
- **Middleware chain** (factory/server.go): recovery → cache → logging → hooks. Add new middleware via `gosdk.WithMiddleware()`
- **Singleflight**: scorecard uses `scorecardFlight.Do()` to dedup concurrent computations; tag cache uses `tagCacheFlight`
- **ResourcesAsTools**: `TrackResource()` in `resources/handlers.go` feeds `read_resource`/`list_resources` tools

## LLM backends

| Backend | Tool | When |
|---------|------|------|
| Apple FM | `apple_foundation_models` | darwin/arm64/cgo, on-device |
| Ollama | `ollama` | local server (`ollama serve`) |
| MLX | `mlx` | Apple Silicon, bridge-only |
| Auto | `text_generate provider=auto` | model router selects best |

Check `stdio://models` resource for `backends.fm_available` before calling FM tools.

## Reports and health

```
report   action=overview|scorecard|briefing
health   action=docs|git|cicd
session  action=prime|handoff
```

## Skills available

`use-exarp-tools`, `task-workflow`, `report-scorecard`, `session-handoff`, `task-cleanup`, `lint-docs`, `tractatus-decompose`, `thinking-workflow`

## Configuration examples

Example configs for Cursor (`.cursor/mcp.json`), OpenCode (`opencode.json`), Claude Code (`.claude/settings.json`), and optional project files: **`docs/examples/`** — see `docs/examples/README.md` for the index. Use these when suggesting MCP, Claude Code, or OpenCode setup to users or other agents.

## Go conventions

- Table-driven tests (`t.Run`)
- Interfaces for abstraction (`TextGenerator`, `MCPServer`)
- Context as first param always
- `internal/` packages for all implementation; `cmd/` for entry points
- SQLite primary storage; JSON fallback only

## Scorecard

```bash
make scorecard      # fast mode (skips build/test/vulncheck)
make scorecard-full # full mode (all checks)
make scorecard-fix  # auto-fix tidy + fmt + lint issues
```

Via MCP: `report` with `action=scorecard`, `skip_scorecard_cache=true` after fixes (5-min cache).

## Pre-commit hook

Runs `make build` + health check (no vulnerability scan). Run `make pre-release` before release for build + govulncheck + security scan.

## Make shortcuts

`make b` (build), `make tidy`, `make fmt`, `make test`, `make lint`, `make p` (push), `make pl` (pull), `make st` (status). Shell alias `r` = cd to repo root.
