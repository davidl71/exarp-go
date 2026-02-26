# exarp-go — Claude Code Guide

## Project

Go-based MCP server. 35 tools (36 with Apple FM), 36 prompts, 24 resources. Primary language: Go. Also: shell scripts (scripts/, ansible/), Ansible (YAML playbooks/roles in ansible/). SQLite-backed task system (Todo2). Apple Foundation Models + Ollama + MLX for local AI.

## MCP servers available in this session

| Server | Role |
|--------|------|
| `exarp-go` | Executor — task mgmt, reports, health, session, testing |
| `tractatus_thinking` | Analyst — logical decomposition |
| `context7` | Researcher — library docs |
| `devwisdom` | Guidance — wisdom, advisors |

## Session start (always)

Call `session` tool with `action=prime`, `include_hints=true`, `include_tasks=true` at the start of every session to get current task context, handoffs, and mode hints.

## Task management (Todo2)

**Prefer MCP tools over direct file edits. Never edit `.todo2/state.todo2.json` or `.todo2/todo2.db` directly.**

```
# MCP tool (preferred)
task_workflow  action=create|update|delete|summarize|run_with_ai|sync|clarify

# CLI (for quick ops in terminal)
go run ./cmd/server task list --status Todo
go run ./cmd/server task create "Name" --priority high --local-ai-backend fm
go run ./cmd/server task update T-xxx --new-status Done
go run ./cmd/server task show T-xxx
go run ./cmd/server task estimate "Task name" --local-ai-backend ollama
go run ./cmd/server task summarize T-xxx [--local-ai-backend fm]
go run ./cmd/server task run-with-ai T-xxx [--backend ollama] [--instruction "..."]
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
- **New task_workflow actions**: add to switch in `task_workflow_native.go`, handler in `task_workflow_common.go`, enum in `registry.go`
- **Count sync**: when adding tools/prompts/resources, update comment + test assertions + expected lists

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
