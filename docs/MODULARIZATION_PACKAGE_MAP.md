# Modularization package map (exarp-go ↔ mcp-go-core ↔ optional MCPs)

**Tag hints:** `#refactor` `#mcp` `#planning`

Living map for **Go modularization**: what stays in **exarp-go**, what should move to **`mcp-go-core`**, and what is a candidate for a **separate MCP server**. Regenerate dependency hints after large refactors (e.g. `rg 'internal/database' internal/tools/*.go`).

## Legend

| Symbol | Meaning |
|--------|---------|
| **Stay** | Remains in `github.com/davidl71/exarp-go`; product / Todo2 / SQLite / exarp workflows. |
| **Core** | Candidate for `github.com/davidl71/mcp-go-core` — reusable across Go MCP servers (no Todo2 schema). |
| **MCP‑LLM** | Optional second server: local model backends, CGO variance, large tool surface. |
| **MCP‑Dev** | Optional second server: linters, tests, security scanners, heavy external binaries. |
| **MCP‑Cursor** | Optional second server: Cursor-only config generation and Cloud Agents API. |

**Heuristic:** If a file imports `internal/database`, treat it as **Stay** unless you are explicitly extracting a data layer behind an interface (not started).

---

## 1. `internal/` packages (outside `tools`)

| Package | Role | Target |
|---------|------|--------|
| `cmd/server` | Entry: CLI, MCP stdio, HTTP, ACP | **Stay** |
| `internal/acp` | ACP adapter | **Stay** |
| `internal/api` | HTTP API for PWA | **Stay** |
| `internal/archive` | Deprecated | **Stay** / delete over time |
| `internal/cache` | TTL + file cache (scorecard, etc.) | **Core** (already planned in extraction notes) |
| `internal/cli` | CLI + TUIs | **Stay** (product UX) |
| `internal/config` | `.exarp` protobuf config | **Stay** (exarp project model) |
| `internal/database` | SQLite Todo2, migrations, locks | **Stay** |
| `internal/factory` | MCP factory + middleware | **Stay** wiring; **Core** for generic middleware patterns |
| `internal/framework` | Re-exports mcp-go-core `framework` + types | **Core** (already) |
| `internal/logging` | slog wrapper | **Core** if duplicated elsewhere; else **Stay** |
| `internal/models` | `Todo2Task`, status/priority constants | **Stay** |
| `internal/platform` | OS helpers | **Core** candidate if generic |
| `internal/projectroot` | `.exarp` / `.todo2` discovery | **Split**: generic walker → **Core**; marker list → **Stay** or injectable |
| `internal/prompts` | MCP prompt catalog | **Stay** (exarp-specific prompts) |
| `internal/proto` | Generated + request types | **Stay** (tool contracts tied to exarp) |
| `internal/queue` | Redis + Asynq waves | **Stay** (or separate **worker binary**, not necessarily MCP) |
| `internal/resources` | `stdio://` resource handlers | **Stay** URIs; **Core** could host generic resource registration helpers |
| `internal/security` | Rate limit, vuln scan, path re-export | Path helpers → **Core** (partial); scanner policy → **Stay** / **MCP‑Dev** |
| `internal/taskanalysis` | Small classifier | **Stay** (task semantics) |
| `internal/tasksync` | Cross-project sync | **Stay** |
| `internal/taskworkflowspec` | JSON schema for `task_workflow` tool | **Stay** |
| `internal/tools` | MCP tool handlers | See §2 |
| `internal/utils` | Misc | **Core** if copy-pasted to other repos; else **Stay** |

---

## 2. `internal/tools` — clusters and files

Tools are registered from `registry.go` → `registry_core.go`, `registry_ai.go`, `registry_infra.go`, `registry_misc.go`.

### 2.1 Core registry (`registerCoreTools`)

| Concern | Primary files | Target |
|---------|---------------|--------|
| **task_workflow** | `task_workflow_native.go`, `task_workflow_actions.go`, `task_workflow_common.go`, `task_workflow_crud.go`, `task_workflow_maintenance*.go`, `task_workflow_execution.go`, `task_workflow_agent.go`, `task_workflow_create_ai.go`, `task_workflow_ai_run.go`, `task_workflow_followup.go`, `task_store.go`, `todo2_db_adapter.go`, `todo2_utils.go`, `todo2_json.go`, `plan_sync.go`, `task_tool_rules.go` | **Stay** |
| **task_discovery** | `task_discovery_native.go`, `task_discovery_native_scanners.go`, `task_discovery_common.go`, `task_discovery_native_nocgo.go` | **Stay** |
| **task_analysis** | `task_analysis_shared.go`, `task_analysis_graph.go`, `task_analysis_algorithms.go`, `task_analysis_deps.go`, `task_analysis_deps_analysis.go`, `task_analysis_suggest_deps.go`, `task_analysis_ownership.go`, `task_analysis_tags.go`, `task_analysis_tags_discover.go`, `task_analysis_tags_llm.go`, `task_analyzer.go`, `graph_helpers.go`, `plan_waves.go`, `planning_links_helpers.go`, `statistics.go`, `normalization.go` | **Stay** (gonum + `models.Todo2Task`) |
| **session** | `session.go`, `session_handoff.go`, `session_helpers.go`, `session_helpers_handoff.go`, `session_assignee.go`, `session_ledger.go`, `session_lazy_context.go`, `session_mode_inference.go` | **Stay** |
| **report** | `report.go`, `report_format.go`, `report_data.go`, `report_plan.go`, `report_plan_generate.go`, `report_plan_overrides.go`, `report_mlx.go`, `scorecard_*.go`, `wisdom.go` | **Stay** |
| **health** | `health_check.go` | **Stay** (DB + project checks); thin “run command” runners could move to **MCP‑Dev** |
| **infer_task_progress** | `infer_task_progress.go`, `infer_task_progress_evidence.go` | **Stay** |

### 2.2 AI registry (`registerAITools`)

**LLM surface (as of this doc):**

- There is **no** `llamacpp` MCP tool in the tree (GGUF path removed; see `docs/BACKLOG_PLAN_2026_03_24.md`). Older docs (`docs/llamacpp-build-requirements.md`, `README.md`) may still mention build steps — treat as historical unless the tool is reintroduced.
- **Apple Foundation Models** are used via **`text_generate`** (`provider=fm`), `fm_plan_and_execute`, scorecard/report insight providers, and `apple_foundation_helpers.go` — not via a separate `apple_foundation_models` tool in `registry_ai.go`. **`cmd/sanity-check`** and **`ExpectedToolCountBase`** in `health_check.go` expect **36** tools; update them when registration changes.

| Concern | Primary files | Target |
|---------|---------------|--------|
| **memory / memory_maint** | `memory.go`, `memory_maint.go`, `memory_maint_utils.go`, `process_memory.go` | **Stay** (exarp memory store / task coupling) |
| **estimation** | `estimation_native.go`, `estimation_native_nocgo.go`, `estimation_shared.go`, `estimation_shared_v2.go`, `estimation_historical.go`, `estimation_analytics.go` | **Stay** (historical uses DB); pure math helpers → **Core** if extracted |
| **ollama / mlx / FM / routers** | `ollama_native.go`, `ollama_native_handlers.go`, `ollama_provider.go`, `mlx_*.go`, `fm_*.go`, `apple_foundation_helpers.go`, `model_router.go`, `text_generate.go`, `llm_backends.go`, `llm_response.go`, `localai_provider.go`, `gateway_provider.go`, `insight_provider.go`, `fm_chain.go`, `fm_plan_execute.go` | **MCP‑LLM** (optional split) |
| **context / prompt_tracking** | `context.go`, `context_shared.go`, `context_native.go`, `context_native_nocgo.go`, `prompt_tracking.go` | **MCP‑LLM** or **Core** if made backend-agnostic |
| **recommend** | `recommend.go` | **Stay** or **MCP‑LLM** if only used with local models |
| **cursor_cloud_agent** | `cursor_cloud_agent.go`, `agent_runner*.go` | **MCP‑Cursor** |
| **fm_plan_and_execute / task_execute** | `fm_plan_execute.go`, `task_execute.go`, `execution_apply.go` | **Stay** (task store); planner-only → **MCP‑LLM** |
| **research_aggregator** | `research_aggregator.go` | **Stay** (orchestrates exarp tools) |

### 2.3 Infra registry (`registerInfraTools`)

| Concern | Primary files | Target |
|---------|---------------|--------|
| **automation** | `automation_native.go`, `automation_schedule.go`, `automation_scheduled.go`, `automation_discover.go` | **Stay** |
| **git_tools** | `git_tools.go`, `git_tools_actions.go` | **Stay** (task-linked actions); pure git wrapper → **Core** or **MCP‑Dev** |
| **testing** | `testing.go` | **MCP‑Dev** |
| **lint** | `linting.go`, `linting_*.go` | **MCP‑Dev** |
| **security** | `security.go` | **MCP‑Dev** |
| **generate_config** | `config_generator.go`, `config_generator_rules.go`, `config_generator_ignore_simplify.go` | **MCP‑Cursor** |
| **setup_hooks** | `hooks_setup.go` | **Stay** (exarp hook recipes) or **MCP‑Dev** |

### 2.4 Misc registry (`registerMiscTools`)

| Concern | Primary files | Target |
|---------|---------------|--------|
| **analyze_alignment** | `alignment_analysis.go` | **Stay** |
| **check_attribution** | `attribution_check.go` | **Stay** |
| **add_external_tool_hints** | `external_tool_hints.go` | **Stay** or **MCP‑Cursor** |
| **tool_catalog** | `tool_catalog.go` | **Core** (generic help-by-tool-name pattern) |
| **workflow_mode / infer_session_mode** | `workflow_mode.go`, `session_mode_inference.go` | **Stay** |
| **ask_client** | `sampling_tool.go` | **Core** (sampling bridge); policy in **Stay** |
| **resources_as_tools** | `resources_as_tools.go`, `resource_notifications.go` | **Core** pattern; URI handlers stay in `internal/resources` |

### 2.5 Cross-cutting handler plumbing

| File(s) | Target |
|---------|--------|
| `registry.go`, `registry_*.go` | **Stay** |
| `handlers.go`, `handlers_ai.go`, `handlers_wrap.go`, `handlers_wrap_test.go` | **Stay** dispatch; **Core** for `WrapHandler` / generic adapters |
| `protobuf_helpers.go`, `protobuf_helpers_report.go`, `protobuf_helpers_tools.go` | **Core** (generic PB↔map; already partially there) |
| `params_helpers.go`, `helpers.go`, `path_validation.go`, `response_compact.go` | **Core** |
| `server_status.go`, `gotohuman.go` | **Stay** (product integrations) |
| `conflict_detection.go` | **Stay** |

---

## 3. Optional MCP servers (summary)

| Server | Tools / scope | Why split |
|--------|----------------|-----------|
| **MCP‑LLM** | `ollama`, `mlx`, `text_generate` (incl. `provider=fm`), `fm_plan_and_execute`, FM helper stack (`fm_*.go`, `apple_foundation_helpers.go`); optional `estimation` generate path | CGO, binary size, model daemons, crashes isolated from task server. No separate `llamacpp` tool today. |
| **MCP‑Dev** | `lint`, `testing`, `security` | Many external binaries; overlaps dedicated security MCPs |
| **MCP‑Cursor** | `generate_config`, `cursor_cloud_agent`, maybe `external_tool_hints` | Irrelevant for non-Cursor clients |

Default **exarp-go** MCP keeps: **`task_workflow`**, **`session`**, **`report`**, **`task_analysis`**, **`task_discovery`**, **`health`** (slim), **`infer_task_progress`**, **`automation`**, **`git_tools`** (if you want single-round-trip linking).

---

## 4. mcp-go-core — already in use

exarp-go already depends on `mcp-go-core` for:

- `pkg/mcp/framework`, `pkg/mcp/types`, `pkg/mcp/framework/adapters/gosdk`
- `pkg/mcp/cli`, `pkg/mcp/logging`, `pkg/mcp/security` (path validation)

Extend **Core** with: shared caches, file locks, compact JSON responses, tool error types, and reusable resource-as-tool registration — aligned with `docs/plans/mcp-go-core-extraction.plan.md` (historical task list in `docs/task_discovery_report.json`).

---

## 5. Maintenance

- **Re-scan:** `rg 'internal/database' internal/tools/*.go` → must remain **Stay** unless introducing a repository interface.
- **Counts:** After registry changes, run `make sanity-check` and update `docs/ARCHITECTURE.md` tool/prompt/resource counts if documented there.
- **Related:** `docs/ARCHITECTURE.md` (high-level package table), `.cursor/rules/code-map.mdc` (tool → file map).
