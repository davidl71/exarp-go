# exarp-go Architecture

**Tag hints:** `#docs` `#refactor`

> A Go-based MCP (Model Context Protocol) server providing 35 tools (36 with Apple FM), 36 prompts, and 24 resources
> for AI-assisted project management, code quality, and local LLM integration.

## Package Map

| Package | Responsibility | Key Files |
|---|---|---|
| `cmd/server` | Binary entry point: CLI dispatch, MCP stdio, HTTP API, ACP server | `main.go` |
| `internal/acp` | Agent Client Protocol adapter (Zed, JetBrains, OpenCode) | `server.go` |
| `internal/api` | HTTP REST API for PWA UI | `server.go` |
| `internal/archive` | Archived/deprecated code | — |
| `internal/cache` | File-based caching (scorecard, reports) with TTL | `file_cache.go` |
| `internal/cli` | CLI subcommands + Bubbletea TUI + IBM 3270 TUI | `cli.go`, `tui*.go`, `tui3270*.go`, `task.go` |
| `internal/config` | Protobuf-based project configuration (`.exarp/config.pb`) | `loader.go`, `schema.go`, `writer.go` |
| `internal/database` | SQLite storage: tasks, comments, activities, locks, migrations | `tasks_crud.go`, `tasks_list.go`, `locking.go`, `schema.go` |
| `internal/factory` | MCP server factory (creates framework instances from config) | `server.go` |
| `internal/framework` | MCP server interface abstraction (re-exports from mcp-go-core) | `server.go` |
| `internal/logging` | Structured logging (slog-based) | `logger.go` |
| `internal/models` | Shared types and constants (no DB dependency) | `todo2.go`, `constants.go`, `task_id.go` |
| `internal/platform` | Platform-specific utilities (darwin detection, etc.) | — |
| `internal/projectroot` | Project root detection (go.mod, .git, .exarp) | — |
| `internal/prompts` | MCP prompt definitions and registration | `registry.go` |
| `internal/queue` | Redis + Asynq job queue for wave execution | `producer.go`, `worker.go`, `config.go` |
| `internal/resources` | MCP resource handlers (`stdio://` URIs) | `handlers.go`, `tasks.go`, `tools.go` |
| `internal/security` | Rate limiting, path validation, vulnerability scanning | `ratelimit.go`, `scanner.go` |
| `internal/taskanalysis` | Graph-based task analysis (gonum) | — |
| `internal/tasksync` | Cross-project task synchronization | — |
| `internal/tools` | All MCP tool handlers (business logic) — largest package | `handlers.go`, `registry.go`, 50+ handler files |
| `internal/utils` | Small shared utilities | — |
| `internal/web` | Embedded PWA UI + SPA handler | — |

## Data Flow

```mermaid
flowchart TD
    subgraph Entry["Entry Points"]
        CLI["CLI Request<br/>(exarp-go task list)"]
        MCP["MCP JSON-RPC<br/>(stdio from Cursor/Claude)"]
        HTTP["HTTP API<br/>(exarp-go -serve :8080)"]
        ACP_IN["ACP Protocol<br/>(exarp-go -acp)"]
    end

    subgraph Dispatch["Dispatch Layer"]
        MAIN["cmd/server/main.go<br/>Mode detection"]
        CLI_DISPATCH["internal/cli/cli.go<br/>Subcommand routing"]
        FACTORY["internal/factory/server.go<br/>Framework instantiation"]
    end

    subgraph Registration["Registration"]
        TOOLS_REG["internal/tools/registry.go<br/>35 tools (4 groups)"]
        PROMPTS_REG["internal/prompts/registry.go<br/>36 prompts"]
        RESOURCES_REG["internal/resources/handlers.go<br/>24 resources"]
    end

    subgraph Handlers["Tool Handlers"]
        HANDLERS["internal/tools/handlers.go<br/>Top-level dispatch"]
        TOOL_IMPL["Per-tool handler files<br/>(report.go, session.go, etc.)"]
    end

    subgraph Storage["Storage & Config"]
        DB["internal/database<br/>SQLite (.todo2/todo2.db)"]
        CONFIG["internal/config<br/>Protobuf (.exarp/config.pb)"]
        CACHE["internal/cache<br/>File-based TTL cache"]
    end

    subgraph LLM["LLM Backends"]
        FM["Apple Foundation Models<br/>(darwin/arm64/cgo)"]
        OLLAMA["Ollama Server<br/>(local HTTP)"]
        MLX["MLX<br/>(Apple Silicon)"]
    end

    CLI --> MAIN
    MCP --> MAIN
    HTTP --> MAIN
    ACP_IN --> MAIN

    MAIN -->|"CLI flags detected"| CLI_DISPATCH
    MAIN -->|"MCP stdio mode"| FACTORY
    MAIN -->|"-serve flag"| FACTORY
    MAIN -->|"-acp flag"| FACTORY

    FACTORY --> TOOLS_REG
    FACTORY --> PROMPTS_REG
    FACTORY --> RESOURCES_REG

    CLI_DISPATCH -->|"-tool flag"| HANDLERS
    CLI_DISPATCH -->|"task subcommand"| DB
    TOOLS_REG --> HANDLERS
    HANDLERS --> TOOL_IMPL

    TOOL_IMPL --> DB
    TOOL_IMPL --> CONFIG
    TOOL_IMPL --> CACHE
    TOOL_IMPL --> FM
    TOOL_IMPL --> OLLAMA
    TOOL_IMPL --> MLX
```

## Entry Points

| Entry Point | File | Purpose |
|---|---|---|
| CLI dispatch or MCP stdio | `cmd/server/main.go` | Detects mode from flags/TTY, routes to CLI or MCP server |
| CLI subcommand routing | `internal/cli/cli.go` | Parses `-tool`, `task`, `tui`, `config` subcommands |
| MCP tool dispatch | `internal/tools/handlers.go` | Routes JSON-RPC tool calls to per-tool handler functions |
| Tool registration | `internal/tools/registry.go` | Registers 35 tools in 4 semantic groups with schemas |
| HTTP API + PWA | `internal/api/server.go` + `internal/web/` | REST API wrapping MCP tools; embedded SPA |
| ACP server | `internal/acp/server.go` | Agent Client Protocol for Zed/JetBrains/OpenCode |

## Key Abstractions

| Abstraction | Location | Purpose |
|---|---|---|
| `framework.MCPServer` | `internal/framework/server.go` | MCP server interface (RegisterTool, RegisterResource, Run) |
| `framework.ToolHandler` | `internal/framework/server.go` | `func(ctx, json.RawMessage) ([]TextContent, error)` |
| `models.Todo2Task` | `internal/models/todo2.go` | Canonical task struct used across all packages |
| `models.Status*` / `Priority*` | `internal/models/constants.go` | Named constants for statuses, priorities, comment types |
| `database.TaskStore` | `internal/database/tasks_crud.go` | CRUD operations: GetTask, CreateTask, UpdateTask, DeleteTask |
| `database.ClaimTaskForAgent` | `internal/database/tasks_lock.go` | Distributed lock acquisition for multi-agent safety |
| `config.FullConfig` | `internal/config/schema.go` | Protobuf-based project configuration |
| `TextGenerator` interface | `internal/tools/text_generate.go` | LLM provider contract (FM, Ollama, MLX, LocalAI) |
| `cache.ScorecardCache` | `internal/cache/file_cache.go` | TTL-based cache for expensive scorecard computation |

## Tool Handler Pattern

Every MCP tool follows a consistent pattern:

```
handlers.go (dispatch)  →  <tool>_native.go (entry)  →  <tool>_common.go (shared logic)
                                                     →  <tool>_provider.go (external service)
```

1. **`handlers.go`**: Top-level dispatch function per tool. Parses protobuf/JSON args, applies defaults, routes to native handler.
2. **`*_native.go`**: Platform-specific entry point (action switch). Files ending `_nocgo.go` provide stubs for non-CGO builds.
3. **`*_common.go`**: Shared business logic that works across native/bridge implementations.
4. **`*_provider.go`**: External service clients (Ollama HTTP, MLX bridge, etc.).

## Adding a New Tool

1. **Create handler file** in `internal/tools/`:
   - Name it `<tool_name>.go` (or `<tool_name>_native.go` if platform-specific)
   - Add file-level orientation comment
   - Implement handler function: `func handle<ToolName>(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error)`

2. **Register in the appropriate registry file**:
   - Core tools (task_workflow, session, report, health): `registry_core.go`
   - AI/LLM tools (memory, estimation, ollama, mlx, text_generate, etc.): `registry_ai.go`
   - Infra tools (automation, git_tools, testing, lint, security, hooks): `registry_infra.go`
   - Misc tools (alignment, attribution, tool_catalog, workflow_mode, etc.): `registry_misc.go`
   - Provide tool name, description (with `[HINT: ...]`), JSON schema, and handler reference

3. **Add protobuf support** (optional):
   - Define request/response in `proto/*.proto`
   - Add `Parse<ToolName>Request()` and `<ToolName>RequestToParams()` in `protobuf_helpers.go`

4. **Update counts and tests**:
   - Update tool count in `registry.go` batch comment
   - Add to expected tool list in `internal/tools/registry_test.go`
   - Run `make sanity-check` to verify counts

5. **Update code-map**:
   - Add entry to `.cursor/rules/code-map.mdc` tool table

## Storage Architecture

```
.todo2/
├── todo2.db              # Primary: SQLite database (schema v8)
└── state.todo2.json      # Fallback: Legacy JSON (auto-migration available)

.exarp/
└── config.pb             # Project config (protobuf binary)
```

- **Database-first**: All task operations go through `internal/database` (SQLite with WAL mode)
- **JSON fallback**: `LoadTodo2Tasks()` / `SaveTodo2Tasks()` auto-detect and fall back to JSON if DB unavailable
- **Migrations**: `internal/database/migrations/*.sql` — applied automatically on Init
- **Locking**: `database.ClaimTaskForAgent()` provides lease-based distributed locks for multi-agent safety

## LLM Integration

The project supports multiple local LLM backends through a unified abstraction:

| Backend | Tool | Build Constraint | Provider |
|---|---|---|---|
| Apple Foundation Models | `apple_foundation_models` | `darwin && arm64 && cgo` | `DefaultFMProvider()` |
| Ollama | `ollama` | None (HTTP client) | `DefaultOllama()` |
| MLX | `mlx` | None (bridge) | `handleMlxNative()` |
| Auto-router | `text_generate` | None | `model_router.go` |

The `text_generate` tool with `provider=auto` uses `model_router.go` to select the best available backend.

## Large Files Reference

Files over 600 lines that are candidates for future splitting if they become merge-conflict hotspots:

| File | Lines | Notes |
|---|---|---|
| `internal/tools/task_workflow_native.go` | ~730 | Clarify sub-action is ~350 lines — candidate for `task_workflow_clarify.go` |
| `internal/tools/todo2_utils.go` | ~655 | Mixed: I/O, overview writing, helpers, status utils |
| `internal/tools/automation_scheduled.go` | ~635 | Three independent handlers (daily/nightly/sprint) |
| `internal/tools/task_workflow_maintenance.go` | ~664 | Sync + sanity + cleanup |
| `internal/config/protobuf.go` | ~1,071 | 15+ proto conversion pairs; could split by config section |
