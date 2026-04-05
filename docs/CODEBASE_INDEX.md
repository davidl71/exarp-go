# exarp-go Codebase Index

An MCP (Model Context Protocol) server for AI-augmented task management with dual TUI interfaces.

**See also:** [ARCHITECTURE.md](./ARCHITECTURE.md) (package map and data flow), [MODULARIZATION_PACKAGE_MAP.md](./MODULARIZATION_PACKAGE_MAP.md) (exarp-go vs `mcp-go-core` vs optional MCP splits).

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      cmd/server/main.go                     │
│                    (MCP Server Entry Point)                   │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│ internal/cli  │    │ internal/tools│    │internal/prompts│
│   (CLI/TUI)   │    │(MCP Tools)   │    │ (MCP Prompts) │
└───────────────┘    └───────────────┘    └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              ▼
                ┌─────────────────────────┐
                │   internal/framework   │
                │   (MCP Protocol Types) │
                └─────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│  internal/    │    │   internal/   │    │   internal/   │
│  database/    │    │   resources/  │    │    queue/     │
└───────────────┘    └───────────────┘    └───────────────┘
```

---

## Command Entry Points (`cmd/`)

| File | Purpose |
|------|---------|
| `cmd/server/main.go` | Main MCP server - registers tools, prompts, resources; starts transport (Stdio, HTTP) |
| `cmd/sanity-check/main.go` | Validates MCP registration counts (tools, prompts, resources) |
| `cmd/scorecard/main.go` | Standalone scorecard reporter |

---

## CLI & TUI (`internal/cli/`)

### Core CLI Dispatch
| File | Functions |
|------|-----------|
| `cli.go` | `RunCLI()`, subcommand dispatch (task, config, tui, tui3270, session, cursor, lock) |
| `task.go` | Task CRUD subcommand handlers |
| `config.go` | Config get/set/show subcommand handlers |
| `cursor.go` | Cursor CLI agent integration |
| `lock.go` | Lock management commands |
| `queue.go` | Queue/job management commands |
| `mode.go` | Session mode inference |
| `logging_adapter.go` | Logger wrapper for CLI output |

### Modern Bubble Tea TUI (`tui*.go`)
| File | Purpose |
|------|---------|
| `tui.go` | Model struct, `RunTUI()`, bubbletea.Program |
| `tui_views.go` | `View()` method - renders all views |
| `tui_update.go` | `Update()` method - message handling |
| `tui_modes.go` | Mode constants (Tasks, Config, Scorecard, etc.) |
| `tui_tasks.go` | Task list rendering (narrow/medium/wide layouts) |
| `tui_detail.go` | Task detail view |
| `tui_commands.go` | Background command factories (`loadTasks`, `updateTaskStatus`) |
| `tui_keybindings.go` | Key action constants |
| `tui_helpers.go` | Helpers: status cycling, transient message clearing |
| `tui_styles.go` | Lipgloss styles, terminal color detection |
| `tui_sorting.go` | Sort order constants and helpers |
| `tui_messages.go` | Custom message types (taskLoadedMsg, tickMsg, etc.) |
| `tui_palette.go` | Command palette |
| `tui_analysis.go` | Task analysis view |
| `tui_scorecard.go` | Scorecard view |
| `tui_handoffs.go` | Session handoffs view |
| `tui_waves.go` | Wave planning view |
| `tui_jobs.go` | Background jobs view |

### TUI Update Handlers
| File | Purpose |
|------|---------|
| `tui_update_actions.go` | Task actions (create, status change, bulk update) |
| `tui_update_filters.go` | Search/filter handling |
| `tui_update_handlers.go` | Global keys, mode transitions |
| `tui_update_navigation.go` | Cursor movement, scrolling |
| `tui_transitions.go` | View transition animations |

### 3270 Classic TUI (`tui3270*.go`)
| File | Purpose |
|------|---------|
| `tui3270.go` | Main 3270 TUI runner using go3270 |
| `tui3270_helpers.go` | Shared helpers, status filters |
| `tui3270_menu.go` | Menu handling |
| `tui3270_screen_tasklist.go` | Task list screen |
| `tui3270_screen_taskdetail.go` | Task detail screen |
| `tui3270_screen_config.go` | Config screen |
| `tui3270_screen_editor.go` | Text editor screen |
| `tui3270_screen_sprintboard.go` | Sprint board view |
| `tui3270_screen_scorecard.go` | Scorecard view |
| `tui3270_screen_handoff.go` | Handoff view |
| `tui3270_screen_health.go` | Health check view |
| `tui3270_screen_gitdashboard.go` | Git dashboard |

---

## MCP Tools (`internal/tools/`)

### Tool Registration
| File | Functions |
|------|-----------|
| `registry.go` | `RegisterAllTools()` — calls the four registries below in order |
| `registry_core.go` | `task_workflow`, `task_discovery`, `task_analysis`, `session`, `report`, `health`, `infer_task_progress` |
| `registry_ai.go` | `memory`, `memory_maint`, `estimation`, `ollama`, `text_generate`, `context`, `prompt_tracking`, `recommend`, `cursor_cloud_agent`, `fm_plan_and_execute`, `task_execute`, `research_aggregator` |
| `registry_infra.go` | `automation`, `git_tools`, `testing`, `lint`, `security`, `generate_config`, `setup_hooks` |
| `registry_misc.go` | `analyze_alignment`, `check_attribution`, `add_external_tool_hints`, `tool_catalog`, `workflow_mode`, `infer_session_mode`, `ask_client`, plus `read_resource` / `list_resources` via `RegisterResourcesAsTools` |

Canonical tool-name list and schema smoke test: `internal/tools/registry_test.go` (`TestRegisterAllTools`).

### Task Management
| File | Purpose |
|------|---------|
| `task_workflow_*.go` | Task CRUD, status updates, bulk operations |
| `task_workflow_actions.go` | Task action handlers |
| `task_workflow_ai_run.go` | AI-driven task execution |
| `task_workflow_create_ai.go` | AI-assisted task creation |
| `task_workflow_maintenance.go` | Cleanup, sync, batch operations |
| `task_discovery_*.go` | Discover tasks from code (TODOs, FIXME, docs) |
| `task_analysis_*.go` | Task analysis, dependencies, tags |
| `task_execute.go` | Execute tasks with agents |
| `task_store.go` | Task persistence layer |
| `task_analyzer.go` | Task complexity analysis |

### Session Management
| File | Purpose |
|------|---------|
| `session.go` | Session lifecycle (prime, handoff) |
| `session_handoff.go` | Session save/restore |
| `session_mode_inference.go` | Infer session mode (agent/ask/manual) |
| `session_helpers.go` | Session utilities |
| `session_assignee.go` | Assignee management |

### Reporting & Scorecard
| File | Purpose |
|------|---------|
| `report.go` | Report generation (overview, scorecard, briefing) |
| `report_plan.go` | Planning reports |
| `report_format.go` | Output formatters |
| `report_insights.go` | Report/scorecard AI insight helpers (FM chain via `DefaultReportInsight`) |
| `scorecard_*.go` | Scorecard checks (Go, multi-language) |

### Git Operations
| File | Purpose |
|------|---------|
| `git_tools.go` | Git operations (commits, branches, diff) |
| `git_tools_actions.go` | Git tool handlers |

### Health & Monitoring
| File | Purpose |
|------|---------|
| `health_check.go` | Health checks (server, git, docs, CICD, tools) |

### Memory & Persistence
| File | Purpose |
|------|---------|
| `memory.go` | AI memory storage |
| `memory_maint.go` | Memory maintenance (prune, consolidate) |

### LLM/AI Backends
| File | Purpose |
|------|---------|
| `llm_backends.go` | LLM provider discovery / backend status |
| `apple_foundation_helpers.go` (+ tests) | Apple FM helpers for `text_generate` / report insight / FM chain (no separate `apple_foundation_models` tool in `registry_ai.go`) |
| `ollama*.go` | Ollama local LLM (`ollama` tool) |
| `localai_provider.go` | LocalAI (`text_generate` provider) |
| `gateway_provider.go` | OpenAI-compatible gateway (`text_generate`) |
| `fm_*.go`, `fm_chain.go`, `insight_provider.go` | FM chain, Ollama bridge, report insight routing |
| `text_generate.go`, `model_router.go` | Unified generate + auto provider selection |

### Estimation & Planning
| File | Purpose |
|------|---------|
| `estimation_*.go` | Task duration estimation |
| `plan_*.go` | Planning tools |
| `recommend.go` | Recommendations (model, workflow, advisor) |

### Other Tools
| File | Purpose |
|------|---------|
| `automation_*.go` | Automation workflows |
| `linting*.go` | Linting tools |
| `security.go` | Security scanning |
| `statistics.go` | Statistics collection |
| `tool_catalog.go` | Tool listing |
| `workflow_mode.go` | Workflow mode management |

---

## MCP Framework (`internal/framework/`)

| File | Types |
|------|-------|
| `server.go` | `MCPServer` interface, `ToolHandler`, `PromptHandler`, `ResourceHandler`, `CreateMessageParams/Result` |
| `request.go` | Request types |
| `response.go` | Response types |

---

## MCP Resources (`internal/resources/`)

| File | Resources |
|------|-----------|
| `tasks.go` | `tasks://` URIs |
| `session.go` | `session://` URIs |
| `scorecard.go` | `scorecard://` URIs |
| `memories.go` | `memory://` URIs |
| `prompts.go` | `prompts://` URIs |
| `tools.go` | `tools://` URIs |
| `models.go` | `models://` URIs |
| `server.go` | `server://` status URIs |
| `cursor_skills.go` | `cursor://` skill URIs |

---

## MCP Prompts (`internal/prompts/`)

| File | Purpose |
|------|---------|
| `registry.go` | `RegisterAllPrompts()` |
| `templates.go` | Prompt templates |

---

## Database (`internal/database/`)

### Core
| File | Purpose |
|------|---------|
| `store.go` | Database store interface |
| `tasks.go` | Task model |
| `tasks_crud.go` | Task CRUD operations |
| `tasks_list.go` | Task listing/filtering |
| `tasks_lock.go` | Task locking |
| `comments.go` | Task comments |

### Drivers
| File | Dialect |
|------|---------|
| `driver.go` | Driver interface |
| `driver_sqlite.go` | SQLite |
| `driver_mysql.go` | MySQL |
| `driver_postgres.go` | PostgreSQL |
| `driver_rqlite.go` | Rqlite |

### Utilities
| File | Purpose |
|------|---------|
| `schema.go` | Table schemas, constants |
| `migrations.go` | Schema migrations |
| `retry.go` | Retry logic |
| `tag_cache.go` | Tag caching |

---

## Configuration (`internal/config/`)

| File | Purpose |
|------|---------|
| `config.go` | Main Config struct |
| `defaults.go` | Default values |
| `loader.go` | Config loading |
| `schema.go` | Config schema |
| `validation.go` | Config validation |
| `protobuf*.go` | Protobuf serialization |

---

## Async Queue (`internal/queue/`)

| File | Purpose |
|------|---------|
| `worker.go` | Asynq worker |
| `producer.go` | Job producer |
| `dispatcher.go` | Wave dispatcher |
| `config.go` | Queue config |

---

## Security (`internal/security/`)

| File | Purpose |
|------|---------|
| `access.go` | Access control |
| `ratelimit.go` | Rate limiting |
| `path.go` | Path validation |
| `semaphore.go` | Concurrency limiting |

---

## Other Packages

| Package | Purpose |
|---------|---------|
| `internal/logging` | Structured logging |
| `internal/models` | Data models (Todo2Task, etc.) |
| `internal/platform` | Platform detection (OS, arch, Apple Silicon) |
| `internal/projectroot` | Project root detection |
| `internal/cache` | File/TTL caching |
| `internal/archive` | Archive retention |
| `internal/api` | HTTP API server |
| `internal/web` | Embedded web dashboard |
| `internal/utils` | Utilities (file locking, process) |
| `internal/taskanalysis` | Task analysis algorithms |
| `internal/taskworkflowspec` | Task workflow spec types |
| `internal/factory` | MCP factory |

---

## Testing

- `tests/` - Integration tests
- `tests/fixtures/` - Test fixtures (mock server)
- `*_test.go` files throughout - Unit tests

---

## Key Patterns

1. **Tool Registration**: Tools registered in `registry*.go`, implement `ToolHandler` interface
2. **Database**: SQL-based with dialect drivers; uses gorp/squirrel
3. **TUI**: Bubble Tea v2 for modern TUI; go3270 for classic 3270
4. **MCP Protocol**: Uses mcp-go-core for transport/stub generation
5. **LLM abstraction**: `text_generate` (`fm`, `ollama`, `insight`, `localai`, `gateway`, `auto`) and the `ollama` tool; Apple FM via `provider=fm` and helpers — see § LLM/AI Backends above.
