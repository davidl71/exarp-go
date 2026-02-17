# exarp-go — Claude Code Guide

## Project

Go-based MCP server. 24 tools, 15 prompts, 6 resources. Primary language: Go. SQLite-backed task system (Todo2). Apple Foundation Models + Ollama + MLX for local AI.

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
```

Task statuses: `Todo` → `In Progress` → `Review` → `Done`

## Build

```bash
make build          # standard build (no CGO)
make build-apple-fm # with Apple Foundation Models (CGO, darwin/arm64)
make test           # all tests
make test-go        # Go tests only
make lint           # lint
make sanity-check   # verify tool/prompt/resource counts
```

Binary: `bin/exarp-go` (project) — NOT `/Users/davidl/go/bin/exarp-go` (stale system install).
Use `go run ./cmd/server ...` for CLI ops during development.

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

`use-exarp-tools`, `task-workflow`, `report-scorecard`, `session-handoff`, `task-cleanup`, `lint-docs`, `tractatus-decompose`

## Go conventions

- Table-driven tests (`t.Run`)
- Interfaces for abstraction (`TextGenerator`, `MCPServer`)
- Context as first param always
- `internal/` packages for all implementation; `cmd/` for entry points
- SQLite primary storage; JSON fallback only

## Pre-commit hook

Runs `make build` + health check + security scan automatically. Fix build errors before committing.
