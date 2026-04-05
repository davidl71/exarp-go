# Tool Language Compatibility Matrix

This document classifies the current `exarp-go` MCP tools by how well they work across different client project languages.

The goal is practical compatibility, not implementation language. A tool counts as language-compatible when it works correctly from a client repo regardless of whether that repo is Go, Python, JavaScript/TypeScript, Rust, or another ecosystem.

## Categories

- **Language-agnostic**: works primarily on tasks, docs, git state, MCP resources, planning, or generic project metadata.
- **Multi-language**: works across multiple code ecosystems with language-aware detection or tool selection.
- **Partially compatible**: usable from non-Go repos, but some actions, defaults, or outputs are still Go-centric.
- **Environment-specific**: compatibility depends mostly on installed runtimes, local AI backends, or external services rather than repo language.

## Matrix

| Tool | Compatibility | Notes |
|------|---------------|-------|
| `infer_task_progress` | Language-agnostic | Uses file scanning and task inference across common source extensions. |
| *(removed)* `mlx` | — | The `mlx` MCP tool was removed; use `ollama` or `text_generate`. |
| `testing` | Partially compatible | `run`, `coverage`, and `validate` are now explicitly documented as Go-project flows; broader framework-aware runners are still not implemented. |
| `security` | Multi-language | Detects and scans Go, Python, Rust, and Node/TypeScript ecosystems. |
| `setup_hooks` | Language-agnostic | Repo/workflow setup, not code-language specific. |
| `estimation` | Language-agnostic | Task estimation and planning support. |
| `ollama` | Environment-specific | Depends on local Ollama availability. |
| *(removed)* `llamacpp` | — | Direct GGUF/llama.cpp in exarp-go was removed; use Ollama or `text_generate`. |
| `prompt_tracking` | Language-agnostic | Prompt logging and analysis. |
| `fm_plan_and_execute` | Language-agnostic | LLM planning/execution workflow. |
| `generate_config` | Language-agnostic | Config/rules generation; output may target editor/tooling rather than a language. |
| `infer_session_mode` | Language-agnostic | Session/workflow inference. |
| `task_discovery` | Language-agnostic | Finds TODOs/comments/docs work across repo content. |
| `health` | Language-agnostic | Server/git/docs/CI/tool health checks. |
| `context` | Language-agnostic | Context management and summarization. |
| `task_execute` | Language-agnostic | Task-driven workflow automation. |
| `add_external_tool_hints` | Language-agnostic | Hint enrichment based on source and docs. |
| `tool_catalog` | Language-agnostic | Tool help and discovery. |
| `workflow_mode` | Language-agnostic | Workflow/focus management. |
| `context_budget` | Language-agnostic | Token budgeting and summarization planning. |
| `report` | Partially compatible | Generally useful across repos; `scorecard` still has Go-centric behavior, but plan output wording is now language-neutral. |
| `read_resource` | Language-agnostic | Reads MCP resources. |
| `text_generate` | Environment-specific | Backend-dependent text generation, not repo-language dependent. |
| `recommend` | Language-agnostic | Recommendations for models/workflows/advisors. |
| `task_workflow` | Language-agnostic | Core task CRUD, sync, summaries, comments, approvals. |
| `task_analysis` | Language-agnostic | Backlog analysis, dependency inference, tagging, execution planning. |
| `analyze_alignment` | Language-agnostic | Backlog/PRD alignment. |
| `session` | Language-agnostic | Session priming and handoff. |
| `memory_maint` | Language-agnostic | Memory cleanup and lifecycle management. |
| `cursor_cloud_agent` | Environment-specific | Depends on Cursor Cloud Agents API access. |
| `automation` | Language-agnostic | Scheduled workflows and repo automation. |
| `memory` | Language-agnostic | Persistent AI memory. |
| `research_aggregator` | Language-agnostic | Aggregates multiple analysis tools. |
| `git_tools` | Language-agnostic | Git history, branches, diffs, and task-linking. |
| `lint` | Multi-language | Supports multiple ecosystems, but current defaults still need cleanup for non-Go repos. |
| `check_attribution` | Language-agnostic | Attribution/license analysis. |
| `list_resources` | Language-agnostic | Resource enumeration. |

## Known compatibility gaps

### 1. `lint` default was Go-biased in some paths

Status: fixed in the current working tree.

The schema advertises `linter=auto`, and the handler now defaults protobuf and runtime behavior to `auto` as well. Omitted `linter` values now flow through language auto-detection instead of a Go-only default.

Relevant code:
- `internal/tools/registry_infra.go`
- `internal/tools/registry_ai.go`

### 2. `testing` still exposes a generic surface but remains Go-specific for execution flows

Current behavior:
- `run` uses `go test`
- `coverage` is Go-only
- `validate` uses Go test validation
- default `test_path` is `./...`
- tool metadata now explicitly documents that `run|coverage|validate` are Go-project flows

This is functionally Go-specific even though the tool name and action names are still generic.

Relevant code:
- `internal/tools/testing.go`

### 3. `report` is broadly useful, but some behavior remains Go-centric

Current state:
- non-Go repos can still use reporting flows
- `scorecard` has non-Go handling, but parts of the narrative still assume Go as a primary language
- plan/report text now uses a neutral codebase summary instead of `Codebase: X files (Go: Y)`

Relevant code:
- `internal/tools/handlers.go`
- `internal/tools/report_plan_generate.go`

### 4. `text_generate` metadata had provider drift

Status: fixed in the current working tree.

Current state:
- runtime handler supports `gateway` (and other registered `text_generate` providers)
- registry hints, provider enum, and tool-catalog metadata stay aligned with `internal/tools/registry_ai.go`

Relevant code:
- `internal/tools/registry_ai.go`
- `internal/tools/tool_catalog.go`

## Recommended priorities

1. Decide whether `testing` should remain explicitly Go-only or gain framework-aware runners.
2. Continue reducing Go-centric narrative in `report` scorecard and adjacent docs.
3. Keep the compatibility matrix aligned as tool defaults and metadata evolve.

## Status of recent fixes

The following compatibility issues have been **resolved** as of commit `44b8649` (2026-03-08):

- ✅ **lint** - Fixed to use `auto` detection by default (no longer Go-biased)
- ✅ **text_generate** - Provider metadata unified (`fm`, `ollama`, `localai`, `gateway`, `auto`, …)
- ✅ **report plan** - Output made language-neutral (removed "Go: X files" bias)
- ✅ **testing validate** - Now explicitly guards against non-Go repos with clear error messages

## Remaining work

See active tasks in `.todo2/` or `exarp-go task list` for:
- Decision on `testing` tool strategy (Go-only vs multi-language)
- Scorecard Go-centric narrative audit
- Ongoing compatibility matrix maintenance
