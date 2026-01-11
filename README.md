# MCP Stdio Tools

Stdio-based MCP server for tools that don't work with FastMCP's static analysis.

## Purpose

This server hosts tools, prompts, and resources that are broken in FastMCP due to static analysis issues. By using stdio mode, these items work correctly without FastMCP's problematic static analysis.

## Contents

### Tools (29 total)

**Originally migrated (23 tools):**
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

**Phase 3 migration (6 tools):**
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

### Prompts (18 total)

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

### Resources (21 total)
**Base resources (11):**
- `stdio://scorecard` - Project scorecard
- `stdio://memories` - All memories
- `stdio://memories/category/{category}` - Memories by category
- `stdio://memories/task/{task_id}` - Memories for task
- `stdio://memories/recent` - Recent memories
- `stdio://memories/session/{date}` - Session memories
- `stdio://prompts` - All prompts
- `stdio://prompts/mode/{mode}` - Prompts by mode
- `stdio://prompts/persona/{persona}` - Prompts by persona
- `stdio://prompts/category/{category}` - Prompts by category
- `stdio://session/mode` - Session mode
- `stdio://server/status` - Server status
- `stdio://models` - Available models
- `stdio://tools` - All tools
- `stdio://tools/{category}` - Tools by category

**Task resources (6):**
- `stdio://tasks` - All tasks
- `stdio://tasks/{task_id}` - Task by ID
- `stdio://tasks/status/{status}` - Tasks by status
- `stdio://tasks/priority/{priority}` - Tasks by priority
- `stdio://tasks/tag/{tag}` - Tasks by tag
- `stdio://tasks/summary` - Task summary

## Dependencies

### Required Dependencies

**System Requirements:**
- **Go 1.24.0+** - Required for building and running the server
  - Install from: https://go.dev/dl/
  - Verify: `go version`
- **Python 3.10+** - Required for Python bridge scripts
  - Verify: `python3 --version`
- **uv** - Python package manager (recommended) or pip
  - Install: `curl -LsSf https://astral.sh/uv/install.sh | sh`
  - Verify: `uv --version`

**Go Modules (automatically managed):**
- `github.com/modelcontextprotocol/go-sdk v1.2.0` - MCP SDK (required)
- `golang.org/x/term v0.38.0` - Terminal utilities (required)
- Indirect dependencies are automatically resolved by `go mod`

**Python Dependencies:**
- No runtime Python dependencies - bridge scripts use standard library only
- Python bridge scripts are self-contained

### Optional Dependencies

**Development Tools:**
- **make** - Build automation (recommended for development)
  - Most systems include make by default
- **golangci-lint** - Advanced Go linting (optional)
  - Install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
  - Used by: `make lint`
- **govulncheck** - Go vulnerability scanner (optional)
  - Install: `go install golang.org/x/vuln/cmd/govulncheck@latest`
  - Used by: Health check scripts

**File Watching Tools (for hot reload):**
- **fswatch** (macOS) - Recommended for file watching
  - Install: `brew install fswatch`
  - Used by: `dev-go.sh --watch`
- **inotifywait** (Linux) - Recommended for file watching
  - Install: `sudo apt-get install inotify-tools` (Debian/Ubuntu)
  - Used by: `dev-go.sh --watch`
- **Polling fallback** - Works without file watchers (slower)

**Python Development Dependencies (optional):**
Install with: `uv sync --dev` or `uv pip install -e ".[dev]"`
- `pytest>=7.0.0` - Python testing framework
- `black>=23.0.0` - Python code formatter
- `ruff>=0.1.0` - Python linter

**Feature-Specific Optional Dependencies:**
- **Ollama** - Required for `ollama` tool functionality
  - Install: https://ollama.ai/
- **MLX** - Required for `mlx` tool functionality
  - Install: https://ml-explore.github.io/mlx/

## Installation

```bash
cd /Users/davidl/Projects/exarp-go

# Install Python dependencies (if using uv)
uv sync

# Or install Python dev dependencies (optional)
uv sync --dev

# Build Go binary
make build
# or
go build -o bin/exarp-go ./cmd/server
```

## Running

```bash
./bin/exarp-go
```

Or use the wrapper script:

```bash
./run_server.sh
```

## Shell Completion

Shell completion is available for bash, zsh, and fish. See [completions/README.md](completions/README.md) for detailed installation instructions.

**Quick setup:**

```bash
# Bash
eval "$(exarp-go -completion bash)" >> ~/.bashrc

# Zsh
eval "$(exarp-go -completion zsh)" >> ~/.zshrc

# Fish
exarp-go -completion fish | source >> ~/.config/fish/config.fish
```

After installation, you'll get tab completion for:
- Tool names when using `-tool` or `-test` flags
- All command-line flags
- Shell types for `-completion` flag

## MCP Configuration

Add to `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "{{PROJECT_ROOT}}/../exarp-go/bin/exarp-go",
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
