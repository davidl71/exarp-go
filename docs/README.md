# Documentation Index

**Last Updated:** 2026-01-07

---

## Current Documentation

### Architecture & Design
- `FRAMEWORK_AGNOSTIC_DESIGN.md` - Framework-agnostic architecture pattern
- `DEVWISDOM_GO_LESSONS.md` - Go development best practices and lessons learned
- `BRIDGE_ANALYSIS.md` - Python bridge architecture and implementation
- `BRIDGE_ANALYSIS_TABLE.md` - Bridge reference table

### Cursor & AI
- `CURSOR_RULES.md` - Cursor rules index and **code/planning tag hints** for Todo2 alignment
- `OPENCODE_INTEGRATION.md` - Use exarp-go with OpenCode (MCP, CLI, HTTP API)

### Active Workflows
- `DEV_TEST_AUTOMATION.md` - Development and testing automation
- `WORKFLOW_USAGE.md` - Workflow usage guide
- `WORKFLOW_MODE_TOOL_GROUPS.md` - Tool groups enable/disable functionality
- `WORKFLOW_MODE_TOOL_GROUPS_TEST_RESULTS.md` - Tool groups test results
- `STREAMLINED_WORKFLOW_SUMMARY.md` - Current workflow summary

### Current Features
- `TASK_TOOL_ENRICHMENT_DESIGN.md` - Task tool enrichment (recommended_tools, tag-based enrichment, session prime / task show)
- `SCORECARD_GO_MODIFICATIONS.md` - Scorecard implementation details
- `SCORECARD_GO_IMPLEMENTATION.md` - Scorecard feature documentation
- `MARKDOWN_LINTING_RESEARCH.md` - Markdown linting research
- `MARKDOWN_LINTING_TEST_RESULTS.md` - Markdown linting test results

### Reference Documentation
- `GO_SDK_MIGRATION_QUICK_START.md` - Quick start guide (may be outdated)
- `MCP_FRAMEWORKS_COMPARISON.md` - MCP framework comparison
- `MCP_GO_FRAMEWORK_COMPARISON.md` - Go framework comparison details
- `MIGRATION_TASKS_SUMMARY.md` - Migration task reference (if still relevant)

### Analysis & Planning
- `MULTI_AGENT_PLAN.md` - Multi-agent execution plan
- `MODEL_ASSISTED_WORKFLOW.md` - Model-assisted workflow design (local LLMs, task breakdown, execution, Phase 6 testing/docs)
- `MLX_ARCHITECTURE_ANALYSIS.md` - MLX integration analysis

### Cleanup & Maintenance
- `DOCUMENTATION_CLEANUP_ANALYSIS.md` - Documentation cleanup analysis
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

