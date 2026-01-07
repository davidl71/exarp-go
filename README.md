# MCP Stdio Tools

Stdio-based MCP server for tools that don't work with FastMCP's static analysis.

## Purpose

This server hosts tools, prompts, and resources that are broken in FastMCP due to static analysis issues. By using stdio mode, these items work correctly without FastMCP's problematic static analysis.

## Contents

### Tools (24 total)

**Originally migrated (12 tools):**
- `analyze_alignment` - Todo2 alignment analysis
- `generate_config` - Cursor config generation
- `health` - Project health checks
- `memory` - Memory management
- `memory_maint` - Memory maintenance
- `report` - Project reporting
- `security` - Security scanning
- `setup_hooks` - Git hooks setup
- `task_analysis` - Task analysis
- `task_discovery` - Task discovery
- `task_workflow` - Task workflow
- `testing` - Testing tools

**Newly migrated (12 tools - return dict instead of JSON strings):**
- `infer_session_mode` - Session mode inference
- `add_external_tool_hints` - Add Context7 hints to documentation
- `automation` - Unified automation tool (daily/nightly/sprint/discover)
- `tool_catalog` - Tool catalog and help
- `workflow_mode` - Workflow mode management
- `check_attribution` - Attribution compliance check
- `lint` - Linting tool
- `estimation` - Task duration estimation
- `ollama` - Ollama integration
- `mlx` - MLX integration
- `git_tools` - Git-inspired task management
- `session` - Session management (prime/handoff/prompts/assignee)

### Prompts (15 total)

**Core prompts (8):**
- `align` - Task alignment analysis
- `discover` - Task discovery
- `config` - Config generation
- `scan` - Security scanning
- `scorecard` - Project scorecard
- `overview` - Project overview
- `dashboard` - Project dashboard
- `remember` - Memory system

**Workflow prompts (7):**
- `daily_checkin` - Daily check-in workflow
- `sprint_start` - Sprint start workflow
- `sprint_end` - Sprint end workflow
- `pre_sprint` - Pre-sprint cleanup workflow
- `post_impl` - Post-implementation review workflow
- `sync` - Synchronize tasks between TODO and Todo2
- `dups` - Find and consolidate duplicate tasks

### Resources (6 total)
- `stdio://scorecard` - Project scorecard
- `stdio://memories` - All memories
- `stdio://memories/category/{category}` - Memories by category
- `stdio://memories/task/{task_id}` - Memories for task
- `stdio://memories/recent` - Recent memories
- `stdio://memories/session/{date}` - Session memories

## Installation

```bash
cd /Users/davidl/Projects/mcp-stdio-tools
uv sync
```

## Running

```bash
./bin/exarp-go
```

Or use the wrapper script:

```bash
./run_server.sh
```

## MCP Configuration

Add to `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "mcp-stdio-tools": {
      "command": "{{PROJECT_ROOT}}/../mcp-stdio-tools/bin/exarp-go",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Stdio-based MCP tools (broken in FastMCP) - 24 tools, 8 prompts, 6 resources migrated from exarp_pma"
    }
  }
}
```

## Development & Testing

### 🚀 Automated Development (Recommended)

**Streamlined workflow - no manual prompting needed:**

```bash
# ⭐ START HERE: Full development mode (auto-reload + auto-test + coverage)
make dev-full

# Just run this once, then edit files - everything happens automatically!
# - Server auto-reloads on file changes
# - Tests auto-run on file changes
# - Coverage reports auto-generate
```

**Quick commands:**
```bash
make test         # Run all tests (Go + Python)
make test-watch   # Test watch mode
make quick-test   # Quick test
```

**Go Development (Hot Reload):**
```bash
make go-dev       # Start Go dev mode with hot reload
make go-dev-test  # Go dev mode with auto-test
make go-build     # Build Go binary
make go-run       # Run Go binary
make go-test      # Run Go tests
```

**📚 Documentation:**
- **Quick Reference**: [docs/WORKFLOW_USAGE.md](docs/WORKFLOW_USAGE.md) - Start here!
- **Full Guide**: [docs/DEV_TEST_AUTOMATION.md](docs/DEV_TEST_AUTOMATION.md) - Complete documentation
- **Summary**: [docs/STREAMLINED_WORKFLOW_SUMMARY.md](docs/STREAMLINED_WORKFLOW_SUMMARY.md) - Overview

### Manual Testing

```bash
# Test server imports
make test-tools

# Test all imports
make test-all

# Run full test suite
make test

# Generate coverage report
make test-html
```

### Code Quality

```bash
make fmt          # Format code
make lint         # Lint code
make lint-fix     # Lint and auto-fix
```

## Migration Status

✅ **Complete** - All broken tools, prompts, and resources migrated from `exarp_pma`

See `MIGRATION_FINAL_SUMMARY.md` for complete details.
