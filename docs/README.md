# Documentation Index

**Last Updated:** 2026-01-07

---

## Current Documentation

### Preferred Tool Surface
- Primary entry points: `task_workflow`, `task_analysis`, `task_discovery`, `report`, `health`, `session`, `automation`, `testing`, `lint`, `security`, `git_tools`, `memory`, `memory_maint`, `recommend`, `text_generate`, `workflow_mode`, `tool_catalog`, `generate_config`, `setup_hooks`
- Specialist tools: backend- or domain-specific helpers such as `ollama`, `mlx`, `cursor_cloud_agent`, `fm_plan_and_execute`, `analyze_alignment`, `check_attribution`
- Compatibility aliases: `task_execute` -> `task_workflow`, `infer_session_mode` -> `session`, `scan_dependency_security` -> `security`, `context_budget` -> `context`
- See `TOOL_CONSOLIDATION_ANALYSIS.md` for the current consolidation map and migration guidance

### Architecture & Design
- `ARCHITECTURE.md` - High-level package map, data flow, link to modularization map
- `MODULARIZATION_PACKAGE_MAP.md` - exarp-go vs `mcp-go-core` vs optional MCP server splits (`internal/tools` file clusters)
- `CODEBASE_INDEX.md` - File-oriented index (CLI, tools, database, queue)
- `FRAMEWORK_AGNOSTIC_DESIGN.md` - Framework-agnostic architecture pattern
- `DEVWISDOM_GO_LESSONS.md` - Go development best practices and lessons learned
- `BRIDGE_ANALYSIS.md` - Python bridge architecture and implementation
- `BRIDGE_ANALYSIS_TABLE.md` - Bridge reference table

### Cursor & AI
- `CODEX.md` - Compact Codex/agent quickstart: what to read first, what to ignore, and the preferred verification command
- `CURSOR_RULES.md` - Cursor rules index and **code/planning tag hints** for Todo2 alignment
- `OPENCODE_INTEGRATION.md` - Use exarp-go with OpenCode (MCP, CLI, HTTP API)
- `GO_AI_ECOSYSTEM.md` - AI/LLM backend stack (FM, Ollama, MLX, LocalAI)
- `LLM_NATIVE_ABSTRACTION_PATTERNS.md` - LLM abstraction patterns and discovery
- `research/LLM_ROUTER_AND_ROUTELLM_RESEARCH.md` - radlab llm-router (gateway) and RouteLLM (ML cost routing) research

### Active Workflows
- TASK_LANES_AND_FILE_OWNERSHIP_PLAN.md - Planning proposal for ownership-aware lanes, file-collision analysis, and safer parallel execution
- `HANDOFF_VIA_GIT.md` - Hand off so remote gets exarp task list via git (export handoff + task snapshot to tracked docs)
- `DEV_TEST_AUTOMATION.md` - Development and testing automation
- `WORKFLOW_USAGE.md` - Workflow usage guide
- `WORKFLOW_MODE_TOOL_GROUPS.md` - Tool groups enable/disable functionality
- `WORKFLOW_MODE_TOOL_GROUPS_TEST_RESULTS.md` - Tool groups test results
- `BACKLOG_PLAN_2026_03_24.md` - Current backlog order after removing llamacpp and cleaning execution-cockpit task state
- `EXARP_EXECUTION_COCKPIT_GAPS.md` - Real-world execution-state gaps and recommended exarp-go modifications
- `STREAMLINED_WORKFLOW_SUMMARY.md` - Current workflow summary

### Current Features
- `TASK_TOOL_ENRICHMENT_DESIGN.md` - Task tool enrichment (recommended_tools, tag-based enrichment, session prime / task show)
- `TASK_LANES_AND_FILE_OWNERSHIP_PLAN.md` - Task lanes and file ownership for collision-aware parallelization (Phase 1 complete)
- `SCORECARD_GO_MODIFICATIONS.md` - Scorecard implementation details
- `SCORECARD_GO_IMPLEMENTATION.md` - Scorecard feature documentation
- `MARKDOWN_LINTING_RESEARCH.md` - Markdown linting research

### Index and discoverability
- `DOCS_AND_CODE_INDEX.md` - Purpose of docs/code index; whether an index helps Cursor and other agents; recommendations
- `CTAGS_USAGE.md` - Using universal-ctags in CI, Make, TUI, and other non-Cursor features; `make tags` target

### Reference Documentation
- `PROTOBUF_USAGE.md` - Protobuf usage and build tooling (make proto, make proto-buf)
- `PROTOBUF_IMPLEMENTATION_STATUS.md` - Protobuf implementation status
- `GO_SDK_MIGRATION_QUICK_START.md` - Quick start guide (may be outdated)
- `MCP_FRAMEWORKS_COMPARISON.md` - MCP framework comparison
- `MCP_GO_FRAMEWORK_COMPARISON.md` - Go framework comparison details
- `MIGRATION_TASKS_SUMMARY.md` - Migration task reference (if still relevant)

### Analysis & Planning
- `MULTI_AGENT_PLAN.md` - Multi-agent execution plan
- `MODEL_ASSISTED_WORKFLOW.md` - Model-assisted workflow design (local LLMs, task breakdown, execution, Phase 6 testing/docs)
- `MLX_ARCHITECTURE_ANALYSIS.md` - MLX integration analysis

### Cleanup & Maintenance
- `PYTHON_CODE_AUDIT_REPORT.md` - Python code audit results

---

## Archived Documentation

Historical documentation has been moved to `docs/archive/` and is excluded from:
- ✅ Markdown linting
- ✅ Automated tests
- ✅ CI/CD checks

See `docs/archive/ARCHIVE_RETENTION_POLICY.md` for retention policy and deletion schedule.

---

## Documentation Standards

- **Format:** Markdown (.md)
- **Linting:** gomarklint (native Go linter)
- **Exclusions:** Archive directory, `.cursor/`, build artifacts
- **Style:** First heading should be level 2 (`##`)

---

## Contributing

When adding new documentation:
1. Follow markdown linting standards
2. Use level 2 headings for main sections
3. Update this index if adding new major sections
4. Archive outdated docs rather than deleting

---

**Total Active Docs:** 22 files  
**Total Archived Docs:** 49 files
