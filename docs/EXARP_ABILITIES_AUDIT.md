# exarp-go Abilities Audit

**Last Updated**: 2026-03-08  
**Version**: v0.3.5  
**Total Tools**: 39

This document provides a complete audit of exarp-go's MCP tools, organized by category and capability.

---

## Executive Summary

exarp-go is a **comprehensive project automation and AI-powered task management system** with 39 MCP tools spanning:

- **Task Management**: Full lifecycle management, analysis, discovery, and AI-assisted execution
- **AI/LLM Integration**: Multi-backend support (Apple FM, Ollama, MLX, llamacpp) with unified interface
- **Project Health**: Health checks, testing, linting, security scanning
- **Developer Experience**: Session management, context budgeting, workflow modes
- **Documentation**: Automated generation, PRD alignment, report generation
- **Git Integration**: Task branches, commit tracking, merge workflows

**Key Differentiators**:
- Native Go implementation (no Python bridge)
- PROJECT_ROOT-aware tooling (workspace context)
- Multi-language support (Go, Python, Rust, Node, Shell, Ansible, Markdown)
- Local-first AI with Apple Silicon optimization

---

## Tool Categories

### 1. Task Management (7 tools)

#### Core Task Operations
- **task_workflow** - Full CRUD + workflow management
  - Actions: `list`, `create`, `update`, `sync`, `delete`, `approve`, `clarify`, `cleanup`, `summarize`, `run`
  - Supports bulk operations, status transitions, filtering
  - Database locking for multi-agent safety

- **task_analysis** - Task intelligence and structure analysis
  - Actions: `duplicates`, `tags`, `discover_tags`, `dependencies`, `execution_plan`, `complexity`, `suggest_splits`
  - Identifies duplicate tasks, analyzes dependency graphs
  - Recommends task splits for large items

- **task_execute** - AI-assisted task execution
  - Loads task from Todo2, generates execution plan via LLM
  - Supports multiple AI backends (FM, Ollama, MLX)
  - Provides cost/time estimates

#### Task Discovery
- **task_discovery** - Find tasks in code/docs
  - Actions: `comments`, `markdown`, `orphans`, `git_json`, `planning_links`, `all`
  - Scans TODO comments, markdown checklists, orphaned tasks
  - Optional auto-creation of discovered tasks
  - Respects deprecation rules (no strikethrough/removed items)

- **infer_task_progress** - Auto-detect completed tasks
  - Analyzes codebase against task descriptions
  - Infers completion status from code changes
  - Reduces manual task updates

#### Task Planning & Alignment
- **analyze_alignment** - Check task/PRD alignment
  - Actions: `todo2`, `prd`
  - Validates backlog against requirements
  - Identifies gaps and misalignments

- **estimation** - Task duration estimation
  - Actions: `estimate`, `analyze`, `stats`, `estimate_batch`
  - Machine learning-based time prediction
  - Learns from historical task data

---

### 2. AI & LLM (8 tools)

#### Backend Implementations
- **apple_foundation_models** - Apple FM API integration
  - Actions: `generate`, `respond`, `status`, `models`
  - Optimized for Apple Silicon
  - Lowest latency on macOS

- **ollama** - Ollama backend (recommended for local inference)
  - Actions: `status`, `models`, `generate`, `pull`, `hardware`, `docs`, `quality`, `summary`
  - Full model lifecycle management
  - Hardware capability detection
  - Quality assessment and documentation generation

- **mlx** - MLX backend for Apple Silicon
  - Actions: `status`, `hardware`, `models`, `generate`
  - Apple Silicon GPU acceleration
  - Alternative to Ollama for local models

- **llamacpp** - llama.cpp backend (stub implementation)
  - Actions: `status`, `models`, `generate`, `load`, `unload`
  - Note: Currently dormant (use Ollama instead)
  - See `docs/LLAMACPP_FUTURE.md`

#### Unified Interfaces
- **text_generate** - Universal LLM interface
  - Providers: `fm`, `ollama`, `insight`, `mlx`, `localai`, `gateway`, `llamacpp`, `auto`
  - Auto-selects best available backend
  - Consistent API across providers

- **fm_plan_and_execute** - Plan-and-execute workflow
  - Breaks complex tasks into subtasks (planner)
  - Executes each subtask (executor)
  - Multi-model support

#### Cursor Cloud Integration
- **cursor_cloud_agent** - Cursor Cloud Agents API
  - Actions: `launch`, `status`, `list`, `follow_up`, `delete`
  - Beta feature for cloud-based agent orchestration

#### Model Selection
- **recommend** - Get recommendations
  - Actions: `model`, `workflow`, `advisor`
  - Recommends best model/workflow for task
  - Context-aware suggestions

---

### 3. Project Health & Quality (6 tools)

#### Health Monitoring
- **health** - Multi-facet health checks
  - Actions: `server`, `git`, `docs`, `dod`, `cicd`, `tools`, `ctags`
  - Checks repo status, documentation health, CI/CD state
  - Validates tool availability

#### Testing
- **testing** - Go test execution and analysis
  - Actions: `run`, `coverage`, `suggest`, `validate`
  - **Language Support**: Go only (requires `go.mod`)
  - Runs Go tests with optional coverage
  - Coverage analysis and gap identification
  - Test structure validation
  - **Note**: For non-Go projects, use `automation` tool with test commands

#### Linting
- **lint** - Multi-language linting
  - Actions: `run`, `analyze`
  - Supports: Go (golangci-lint), Markdown (markdownlint/gomarklint), Shell (shellcheck), YAML, Ansible
  - Link checking in markdown docs
  - See `docs/LINT_TARGETS.md`

#### Security
- **security** - Security scanning
  - Actions: `scan`, `alerts`, `report`
  - Vulnerability detection for Go/Python/Rust/Node
  - Dependency security audit

- **scan_dependency_security** - Alias for security scan
  - Same as `security` action `scan`

- **check_attribution** - License compliance
  - Verifies third-party attribution
  - Checks license requirements

---

### 4. Reporting & Documentation (5 tools)

#### Project Reports
- **report** - Comprehensive reporting
  - Actions: `overview`, `scorecard`, `briefing`, `prd`, `plan`
  - **overview**: Full project snapshot
  - **scorecard**: Quality/status rollup
  - **briefing**: Short standup/handoff summary
  - **prd**: Product requirements document
  - **plan**: Implementation plan

#### Config Generation
- **generate_config** - Generate project configs
  - Actions: `rules`, `ignore`, `simplify`
  - Creates `.cursor/rules`, `.gitignore`, etc.
  - Simplifies complex configs

#### Research & Analysis
- **research_aggregator** - Multi-tool analysis combiner
  - Runs multiple analysis tools
  - Combines outputs into unified report
  - Use for comprehensive project audits

#### Context Management
- **context** - Context management
  - Actions: `summarize`, `budget`, `batch`
  - Token usage estimation
  - Batch processing strategies

- **context_budget** - Token budget estimator
  - Estimates token usage for operations
  - Suggests context reduction strategies

---

### 5. Session & Workflow (5 tools)

#### Session Management
- **session** - Session lifecycle
  - Actions: `prime`, `handoff`, `prompts`, `assignee`
  - **prime**: Start session with context (suggested_next tasks)
  - **handoff**: Create handoff note for session continuity
  - Returns dependency-ordered task suggestions

- **infer_session_mode** - Auto-detect session mode
  - Infers mode: AGENT/ASK/MANUAL
  - Confidence scoring
  - Adapts behavior to user intent

#### Workflow Modes
- **workflow_mode** - Workflow management
  - Actions: `focus`, `suggest`, `stats`
  - Manages focus modes (deep work, review, etc.)
  - Tracks workflow statistics

#### Prompt Management
- **prompt_tracking** - Prompt usage analytics
  - Actions: `log`, `analyze`
  - Tracks prompt patterns
  - Analyzes effectiveness

---

### 6. Git & Version Control (1 tool)

- **git_tools** - Git operations
  - Actions: `commits`, `local_commits`, `branches`, `tasks`, `diff`, `graph`, `merge`, `set_branch`
  - **Task branches**: Automatic branch creation per task
  - **commits**: Show task-related commit history
  - **graph**: Visualize task/commit relationships
  - **merge**: Merge task branch changes
  - Enables task versioning and history

---

### 7. Automation & Maintenance (4 tools)

#### Scheduled Automation
- **automation** - Scheduled workflows
  - Actions: `daily`, `nightly`, `sprint`, `discover`
  - Automated maintenance tasks
  - Sprint planning automation

#### Memory Management
- **memory** - AI memory persistence
  - Actions: `save`, `recall`, `search`
  - Stores AI discoveries and learnings
  - Retrieves context across sessions

- **memory_maint** - Memory lifecycle
  - Actions: `health`, `gc`, `prune`, `consolidate`, `dream`
  - Garbage collection for old memories
  - Memory consolidation and optimization

#### Setup & Onboarding
- **setup_hooks** - Install automation
  - Actions: `git`, `patterns`
  - Installs git hooks
  - Sets up automation patterns

- **add_external_tool_hints** - Tool hint injection
  - Scans source files
  - Adds tool-usage hints for discovery
  - Improves onboarding

---

### 8. Discovery & Catalog (3 tools)

- **list_resources** - MCP resource catalog
  - Lists all registered MCP resources
  - Returns URIs, names, descriptions

- **read_resource** - Read MCP resource
  - Reads resource by URI
  - Use `list_resources` to discover available resources

- **tool_catalog** - Tool documentation
  - Action: `help`
  - Gets detailed help for specific tool
  - Use `stdio://tools` resource for full catalog

---

## Key Resources (stdio://)

exarp-go exposes rich information via MCP resources:

| Resource URI | Content |
|--------------|---------|
| `stdio://config` | Current configuration values (JSON) |
| `stdio://config/schema` | Configuration schema with fields/types |
| `stdio://tools` | Full tool catalog with categories |
| `stdio://tools/{category}` | Tools filtered by category |
| `stdio://prompts` | All prompt names and descriptions |
| `stdio://prompts/mode/{mode}` | Prompts for specific mode |
| `stdio://prompts/persona/{persona}` | Prompts for persona |
| `stdio://prompts/category/{category}` | Prompts by category |
| `stdio://models` | Model catalog + backend availability |
| `stdio://tasks` | Full task list |
| `stdio://tasks/status/{status}` | Tasks by status |
| `stdio://suggested-tasks` | Dependency-ready tasks |
| `stdio://cursor/skills` | Available Cursor skills |

**Pro tip**: Always use MCP resources instead of running `exarp-go --help` or spawning processes.

---

## Config CLI

The CLI provides full config management with validation:

```bash
# View
exarp-go config show              # Show all config
exarp-go config get <key>        # Get specific value
exarp-go config diff             # Compare vs defaults
exarp-go config history          # Change history

# Modify (with validation)
exarp-go config set <key>=<value>   # Set (validated)
exarp-go config reset <key>          # Reset to default
exarp-go config template dev         # Apply template

# Validate
exarp-go config validate          # Validate config
exarp-go config reload           # Reload and validate
```

**Validation**: Values are validated before setting (durations, floats 0-1, status, priority, log levels, booleans).

**Templates**: `dev`, `prod`, `minimal`

### 1. Starting a Session
```json
{
  "tool": "session",
  "args": {
    "action": "prime",
    "include_tasks": true,
    "include_hints": true
  }
}
```
Returns: Project context + `suggested_next` (tasks in dependency order)

### 2. Finding Work
```bash
# CLI
exarp-go task list --status Todo --priority high

# MCP resource
stdio://suggested-tasks
```

### 3. Executing a Task
```json
{
  "tool": "task_execute",
  "args": {
    "task_id": "T-123",
    "backend": "ollama",
    "model": "llama3.2"
  }
}
```

### 4. Project Status
```json
{
  "tool": "report",
  "args": {
    "action": "scorecard"
  }
}
```

### 5. Running Tests
```json
{
  "tool": "testing",
  "args": {
    "action": "run",
    "language": "go"
  }
}
```

### 6. Security Scan
```json
{
  "tool": "security",
  "args": {
    "action": "scan"
  }
}
```

---

## Language Support Matrix

| Language | Testing | Linting | Security | Task Discovery |
|----------|---------|---------|----------|----------------|
| **Go** | ✅ (`testing` tool) | ✅ (golangci-lint) | ✅ (govulncheck) | ✅ (TODO comments) |
| **Python** | ⚠️ (via `automation`) | ⚠️ (partial) | ✅ (pip-audit) | ✅ (TODO comments) |
| **Rust** | ⚠️ (via `automation`) | ✅ (clippy) | ✅ (cargo-audit) | ✅ (TODO comments) |
| **Node.js** | ⚠️ (via `automation`) | ⚠️ (partial) | ✅ (npm-audit) | ✅ (TODO comments) |
| **Shell** | ⚠️ (via `automation`) | ✅ (shellcheck) | N/A | ✅ (TODO comments) |
| **Ansible** | ⚠️ (via `automation`) | ✅ (ansible-lint) | N/A | ✅ (TODO comments) |
| **Markdown** | N/A | ✅ (markdownlint, gomarklint) | N/A | ✅ (checklists) |
| **YAML** | N/A | ✅ (yamllint) | N/A | N/A |

**Note**: The `testing` tool is Go-only. For other languages, use the `automation` tool to schedule your test commands (pytest, npm test, cargo test, etc.).

---

## Best Practices

### 1. Use PROJECT_ROOT Context
- exarp-go uses `PROJECT_ROOT` from MCP config
- Don't pass project root in tool args
- All file paths are relative to PROJECT_ROOT

### 2. Prefer High-Level Workflows
- Use `session` for starting work
- Use `report` for status updates
- Use CLI shortcuts: `exarp-go task list` over raw JSON

### 3. Check Backend Availability
```json
{
  "tool": "list_resources",
  "args": {}
}
```
Read `stdio://models` to check which LLM backends are available

### 4. Use Task Discovery for Onboarding
```json
{
  "tool": "task_discovery",
  "args": {
    "action": "all",
    "create_tasks": true
  }
}
```

### 5. Leverage Resources for Speed
- `stdio://tasks` is faster than running task list
- `stdio://tools` avoids process spawns
- Resources are cached and updated automatically

---

## Tool Performance Notes

### Fast Operations (< 100ms)
- Resource reads (`stdio://`)
- `health` checks (most actions)
- `task_workflow` list/show
- `git_tools` status queries

### Medium Operations (100ms - 1s)
- `lint` runs (depends on file count)
- `testing` runs (depends on test suite)
- `task_analysis` (depends on task count)

### Slow Operations (> 1s)
- LLM inference (`text_generate`, `task_execute`)
- `security` scans (depends on dependencies)
- `task_discovery` with `all` action
- `report` with `scorecard` (comprehensive analysis)

### Optimization Tips
- Use filters to reduce task list size
- Run security scans asynchronously
- Cache LLM responses when possible
- Use `limit` parameter in task queries

---

## Integration Points

### Cursor Integration
- MCP config: `.cursor/mcp.json`
- Skills: `.cursor/skills/`
- Rules: `.cursor/rules/`
- Prompts: Accessible via `stdio://prompts`

### OpenCode Integration  
- Config: `opencode.json`
- Same MCP protocol
- Optimized for command-line usage
- See `docs/INDEX.md` for OpenCode documentation

### CI/CD Integration
- Health checks: `exarp-go -tool health -args '{"action":"cicd"}'`
- Test runs: `exarp-go -tool testing -args '{"action":"run"}'`
- Security scans: `make security-scan`
- Lint: `make lint`

### Git Hooks
- Pre-commit: Runs linters
- Commit-msg: Validates format
- Post-commit: Updates task status
- Install: `exarp-go -tool setup_hooks -args '{"action":"git"}'`

---

## Debugging & Troubleshooting

### Tool Not Found
```bash
# Check tool exists
exarp-go -list | grep tool_name

# Get tool help
exarp-go -tool tool_catalog -args '{"action":"help","tool_name":"task_workflow"}'
```

### Backend Not Available
```bash
# Check model backends
exarp-go -tool list_resources -args '{}' | jq '.data.backends'

# Check specific backend
exarp-go -tool ollama -args '{"action":"status"}'
```

### Slow Performance
- Enable debug logging: `EXARP_DEBUG=1`
- Check slow operation warnings in logs
- Use filters and limits to reduce data
- Consider async/batch operations

### Task Database Issues
- Never edit `.todo2/todo2.db` directly
- Use `task_workflow` or `exarp-go task` CLI
- Check for lock files: `.todo2/.git-sync.lock`
- Verify database integrity: `exarp-go task list`

---

## Future Capabilities

See `docs/LLAMACPP_FUTURE.md` for:
- llamacpp full integration (if needed)
- Alternative backend options
- Performance benchmarks

See `docs/IMPLEMENTATION_PLAN.md` for:
- Track 2: Compatibility improvements
- Track 3: TUI enhancements
- Track 5: Documentation expansion

---

## Summary: When to Use exarp-go

Use exarp-go when you need:

✅ **Task Management**  
- Full lifecycle management (create, update, track)
- AI-assisted task execution
- Task discovery from code/docs

✅ **Project Automation**  
- Health monitoring
- Automated testing and linting
- Security scanning

✅ **AI Integration**  
- Multi-backend LLM support
- Local inference (Ollama, MLX, Apple FM)
- Model recommendations

✅ **Developer Experience**  
- Session management and handoffs
- Context-aware suggestions
- Workflow mode management

✅ **Reporting & Analysis**  
- Project scorecards
- Status briefings
- PRD generation

✅ **Multi-Language Support**  
- Go, Python, Rust, Node, Shell, Ansible, Markdown
- Language-neutral tooling

---

## Related Documentation

- `docs/TASK_TOOLS_GUIDE.md` - In-depth task tool guide
- `docs/INDEX.md` - Full documentation index
- `docs/LINT_TARGETS.md` - Complete linting reference
- `docs/IMPLEMENTATION_PLAN.md` - Roadmap and priorities
- `docs/LLAMACPP_FUTURE.md` - llamacpp integration notes
- `.cursor/skills/` - Cursor skill documentation
- `AGENTS.md` - Agent usage rules

For the most up-to-date tool list:
```bash
exarp-go -list
```

For tool-specific help:
```bash
exarp-go -tool tool_catalog -args '{"action":"help","tool_name":"<tool>"}'
```
