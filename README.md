# MCP Stdio Tools

Stdio-based MCP server for tools that don't work with FastMCP's static analysis.

## Purpose

This server hosts tools, prompts, and resources that are broken in FastMCP due to static analysis issues. By using stdio mode, these items work correctly without FastMCP's problematic static analysis.

## Contents

### Tools (29 total)

**Originally migrated (23 tools):**
- `analyze_alignment` - Todo2 alignment analysis
- `generate_config` - Cursor config generation
- `health` - Project health checks, including explicit SQLite maintenance
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

### Explicit Database Maintenance

Use `health action=database` for SQLite maintenance. These operations are explicit and do not run as side effects of normal task CRUD.

```bash
./bin/exarp-go -tool health -args '{"action":"database","operation":"status"}'
./bin/exarp-go -tool health -args '{"action":"database","operation":"checkpoint","checkpoint_mode":"TRUNCATE"}'
./bin/exarp-go -tool health -args '{"action":"database","operation":"vacuum"}'
./bin/exarp-go -tool health -args '{"action":"database","operation":"analyze"}'
```

`status` reports fields such as `page_count`, `freelist_count`, `estimated_db_bytes`, and `estimated_free_bytes` so you can decide whether a manual `vacuum` is worthwhile.

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

**Task resources (7):**
- `stdio://tasks` - All tasks
- `stdio://tasks/{task_id}` - Task by ID
- `stdio://tasks/status/{status}` - Tasks by status
- `stdio://tasks/priority/{priority}` - Tasks by priority
- `stdio://tasks/tag/{tag}` - Tasks by tag
- `stdio://tasks/summary` - Task summary
- `stdio://suggested-tasks` - Dependency-ready tasks (for Cursor hints)

## Configuration

**Per-Project Configuration (protobuf mandatory):**
- Configuration file: `.exarp/config.pb` (required for file-based config; protobuf binary). No file = in-memory defaults.
- Generate default config: `exarp-go config init` (creates `.exarp/config.pb`)
- Validate config: `exarp-go config validate`
- View config: `exarp-go config show [yaml|json]`
- Export config: `exarp-go config export [yaml|json|protobuf]`
- Convert formats: `exarp-go config convert yaml protobuf` or `exarp-go config convert protobuf yaml`
- **Editing as YAML:** Run `exarp-go config export yaml` to emit YAML, edit, then `exarp-go config convert yaml protobuf` to save.
- See `docs/CONFIGURATION_REFERENCE.md` for full parameter reference and project-type examples.

**Configuration Categories:**
- **Timeouts**: Task locks, tool execution, HTTP clients, database retries
- **Thresholds**: Similarity, coverage, confidence, limits
- **Tasks**: Default status/priority/tags, status workflow, cleanup settings
- **Database**: Connection settings, retry configuration
- **Security**: Rate limiting, path validation, file size limits
- **Logging**: Output format, verbosity, file logging
- **Tools**: Tool-specific settings and overrides
- **Workflow**: Mode settings, automation parameters
- **Memory**: Memory management settings
- **Project**: Project-specific settings
  - Includes `project.task_discovery_ignore_paths` for repo-level discovery excludes

For detailed configuration options, see:
- **`docs/CONFIGURATION_REFERENCE.md`** — Full parameter reference and project-type examples
- `docs/CONFIGURATION_IMPLEMENTATION_PLAN.md`
- `docs/CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md`
- `docs/AUTOMATION_CONFIGURATION_ANALYSIS.md`

**Cursor rules (AI guidance):** If you use Cursor, see `docs/CURSOR_RULES.md` for project rules, including **code and planning tag hints** so generated plans and code stay aligned with Todo2 tags. For **how to use Cursor skills** (task-workflow, locking, git_tools, conflict detection), see [docs/CURSOR_SKILLS_GUIDE.md](docs/CURSOR_SKILLS_GUIDE.md).

**Codex quickstart:** If you use Codex or another terminal coding agent, start with [docs/CODEX.md](docs/CODEX.md) for the shortest repo-specific guide: what to read first, what to ignore, preferred `make` targets, and the fast verification command `make codex-smoke`.

**Model-assisted workflow:** For local LLM integration (CodeLlama/MLX/Ollama) for task breakdown, execution, and prompt optimization, see [docs/MODEL_ASSISTED_WORKFLOW.md](docs/MODEL_ASSISTED_WORKFLOW.md). The docs index is in [docs/README.md](docs/README.md).

**AI/LLM stack:** For the full backend stack (Apple FM, Ollama, MLX, LocalAI) and discovery, see [docs/GO_AI_ECOSYSTEM.md](docs/GO_AI_ECOSYSTEM.md) and [docs/LLM_NATIVE_ABSTRACTION_PATTERNS.md](docs/LLM_NATIVE_ABSTRACTION_PATTERNS.md).

**OpenCode / OAC:** To use exarp-go with [OpenCode](https://opencode.ai/) or [OpenAgentsControl (OAC)](https://github.com/darrenhinde/OpenAgentsControl), add exarp-go as an MCP server in your OpenCode config. See [docs/OPENCODE_INTEGRATION.md](docs/OPENCODE_INTEGRATION.md) for MCP, CLI, and HTTP API; [docs/OPENAGENTSCONTROL_EXARP_GO_COMBO_PLAN.md](docs/OPENAGENTSCONTROL_EXARP_GO_COMBO_PLAN.md) for OAC + exarp-go workflow. Example config: [docs/opencode-exarp-go.example.json](docs/opencode-exarp-go.example.json).

**Cursor Plugin:** exarp-go can be installed as a [Cursor Marketplace plugin](https://cursor.com/marketplace) for one-click MCP setup. See [docs/CURSOR_PLUGIN_PLAN.md](docs/CURSOR_PLUGIN_PLAN.md) for the plugin manifest and config.

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
- **Ninja** (optional) - Cross-platform build; useful when Make is unavailable (e.g. Windows)
  - Install: [ninja-build.org](https://ninja-build.org/) (e.g. `choco install ninja` on Windows, `brew install ninja` on macOS)
  - Use: `ninja` or `ninja build` (build), `ninja test`, `ninja fmt`, `ninja lint`, `ninja clean` — see `build.ninja`
- **golangci-lint** - Advanced Go linting (optional; v2 required for .golangci.yml)
  - Install: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0`  
  - Or binary: `curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.9.0`
  - Used by: `make lint`
- **govulncheck** - Go vulnerability scanner (optional)
  - Install: `go install golang.org/x/vuln/cmd/govulncheck@latest`
  - Used by: `make pre-release` (before release); not run on pre-commit/pre-push. See [docs/VULNERABILITY_CHECK_POLICY.md](docs/VULNERABILITY_CHECK_POLICY.md).

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
- **llama.cpp** - Required for `llamacpp` tool (direct GGUF inference)
  - Build submodule: `make build-libbinding` then `make build-llamacpp`; or run `./scripts/build-llamacpp.sh`. See [docs/llamacpp-build-requirements.md](docs/llamacpp-build-requirements.md).

## Installation

```bash
cd {{PROJECT_ROOT}}

# Install Python dependencies (if using uv)
uv sync

# Or install Python dev dependencies (optional)
uv sync --dev

# Build Go binary
make build
# or (with Ninja, e.g. on Windows)
ninja
# or
go build -o bin/exarp-go ./cmd/server
```

## Running

```bash
./bin/exarp-go          # Start MCP stdio server (default)
```

### CLI Subcommands

```bash
# Task management
exarp-go task list [--status Todo] [--priority high] [--limit 10]
exarp-go task create "Task name" --priority high [--local-ai-backend fm]
exarp-go task update T-xxx --new-status Done
exarp-go task show T-xxx
exarp-go task status T-xxx
exarp-go task estimate "Task name" [--local-ai-backend ollama]
exarp-go task summarize T-xxx [--local-ai-backend fm]
exarp-go task run-with-ai T-xxx [--backend ollama] [--instruction "..."]

# Configuration
exarp-go config init          # Create .exarp/config.pb with defaults
exarp-go config show [yaml|json]
exarp-go config export [yaml|json|protobuf]
exarp-go config convert yaml protobuf

# Interactive TUIs
exarp-go tui                  # Bubbletea terminal UI
exarp-go tui3270 [--port 3270] # IBM 3270 mainframe TUI (TN3270)

# Direct tool invocation
exarp-go -tool <name> -args '{"action":"..."}'
exarp-go -test <name>         # Run tool self-test
exarp-go -list                # List all registered tools
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

Add to `.cursor/mcp.json`. See **[docs/examples/](docs/examples/README.md)** for full examples (binary, wrapper script, per-project).

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

**Nix (development shell):**
```bash
nix develop
# or: nix develop -c $SHELL
# Provides: go, gcc, make, git, protoc, golangci-lint. Uses nixpkgs 24.11 (Go 1.22).
# Enable flakes if needed: nix --extra-experimental-features 'nix-command flakes' develop
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

### Benchmarking

Compare **Ollama (HTTP)** vs **llamacpp (local)** generate performance (latency, throughput):

```bash
make bench        # Run all Go benchmarks (Ollama benchmark runs if ollama serve is up)
```

- **Ollama:** Requires `ollama serve` and a model (e.g. `ollama run llama3.2`). Skipped if server is unreachable.
- **LlamaCpp:** Requires build with `-tags llamacpp,cgo` and a loaded model; skipped otherwise.

Quick one-off report (no full benchmark):

```bash
go test -run TestOllamaVsLlamaCppPerformanceReport -v ./internal/tools/
```

Benchmarks live in `internal/tools/ollama_llamacpp_bench_test.go`.

### Git hooks and release

- **Pre-commit** runs build + health docs (no vulnerability scan). **Pre-push** runs alignment only.
- **Before release:** run `make pre-release` for build + govulncheck + security scan.
- See [docs/VULNERABILITY_CHECK_POLICY.md](docs/VULNERABILITY_CHECK_POLICY.md) for the full policy.

## Automation (daily / sprint)

**For daily and sprint automation, use the Go automation tool:**

```bash
exarp-go -tool automation -args '{"action":"daily"}'
# or: exarp-go -tool automation -args '{"action":"sprint"}'
```

**Python `automate_daily` removed.** Use `exarp-go -tool automation` (Go) only. See `docs/PYTHON_SAFE_REMOVAL_AND_MIGRATION_PLAN.md`.

**Manual checks (not in Go daily):** Duplicate test names — run if needed: `uv run python -m project_management_automation.scripts.automate_check_duplicate_test_names`.

## Migration Status

✅ **Complete** - All broken tools, prompts, and resources migrated from `exarp_pma`

See `MIGRATION_FINAL_SUMMARY.md` for complete details.
