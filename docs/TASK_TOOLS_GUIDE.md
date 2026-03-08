# Task Tools Guide

**Last Updated**: 2026-03-08  
**Implementation**: Native Go (no Python bridge)

This guide covers the three main task management tools in exarp-go.

---

## Quick Reference

| Tool | Purpose | Common Actions |
|------|---------|----------------|
| `task_workflow` | Manage task lifecycle | `list`, `create`, `update`, `sync` |
| `task_analysis` | Analyze task structure | `duplicates`, `tags`, `dependencies` |
| `task_execute` | Execute tasks with AI | `run`, `estimate`, `summarize` |

---

## 1. task_workflow - Task Lifecycle Management

**Purpose**: CRUD operations and workflow management for Todo2 tasks

### Common Actions

#### List Tasks
```bash
exarp-go -tool task_workflow -args '{"action":"list","status":"Todo"}'
exarp-go task list --status Todo  # CLI shortcut
```

#### Create Task
```bash
exarp-go task create "Fix bug" --description "Details" --priority high
```

#### Update Task
```bash
exarp-go task update T-123 --new-status Done
```

#### Sync Tasks
```bash
exarp-go -tool task_workflow -args '{"action":"sync"}'
```

### All Actions
- `list` - List tasks with filters
- `create` - Create new task(s)
- `update` - Update task status/priority/fields
- `delete` - Delete task (wrong project, duplicate)
- `sync` - Sync SQLite ↔ JSON
- `add_comment` - Add comment to task
- `show` - Show full task details
- `estimate` - Estimate task duration
- `summarize` - Generate AI summary
- `run_with_ai` - Get implementation guidance

**Docs**: See `internal/cli/task.go` and `internal/tools/task_workflow*.go`

---

## 2. task_analysis - Task Structure Analysis

**Purpose**: Analyze task quality, find duplicates, analyze dependencies

### Common Actions

#### Find Duplicates
```bash
exarp-go -tool task_analysis -args '{"action":"duplicates","similarity_threshold":0.85}'
```

Finds similar tasks that might be duplicates. Uses content similarity matching.

#### Analyze Tags
```bash
exarp-go -tool task_analysis -args '{"action":"tags"}'
```

Shows tag distribution and suggests consolidation opportunities.

#### Check Dependencies
```bash
exarp-go -tool task_analysis -args '{"action":"dependencies","task_id":"T-123"}'
```

Maps dependency chains and finds circular dependencies.

#### Find Execution Plan
```bash
exarp-go -tool task_analysis -args '{"action":"execution_plan","epic_id":"T-100"}'
```

Creates an execution plan with waves for parallel execution.

### All Actions
- `duplicates` - Find duplicate/similar tasks
- `tags` - Analyze tag usage and patterns
- `dependencies` - Map dependency chains
- `hierarchy` - Analyze parent-child relationships
- `execution_plan` - Generate wave-based execution plan
- `analyze` - General analysis with prompt

**Docs**: See `internal/tools/task_analysis*.go`

---

## 3. task_execute - AI-Assisted Task Execution

**Purpose**: Get AI assistance for task implementation

### Common Actions

#### Run Task with AI
```bash
exarp-go task run-with-ai T-123 --backend ollama
exarp-go task run-with-ai T-123 --instruction "Use TypeScript"
```

Gets step-by-step implementation guidance from local LLM.

#### Estimate Task
```bash
exarp-go task estimate "Add user authentication" --priority high
```

Estimates time/complexity using local AI.

#### Summarize Task
```bash
exarp-go task summarize T-123
```

Generates concise summary of task using AI.

### All Actions
- `run` - Execute task with AI guidance
- `estimate` - Estimate duration and complexity
- `summarize` - Generate AI summary
- `refine_prompt` - Improve task description
- `apply` - Apply execution plan results

**Docs**: See `internal/tools/task_execute*.go`

---

## Workflow Examples

### Finding and Cleaning Duplicates
```bash
# Find duplicates
exarp-go -tool task_analysis -args '{"action":"duplicates"}'

# Review output, then delete duplicates manually
exarp-go task delete T-duplicate-id
```

### Creating and Executing a Task
```bash
# Create task
exarp-go task create "Implement feature X" \
  --description "Add new feature" \
  --priority high \
  --tags "backend,api"

# Get AI guidance
exarp-go task run-with-ai T-new-task-id

# Mark as done
exarp-go task update T-new-task-id --new-status Done
```

### Analyzing Project Health
```bash
# Check for duplicates
exarp-go -tool task_analysis -args '{"action":"duplicates"}'

# Analyze dependencies
exarp-go -tool task_analysis -args '{"action":"dependencies"}'

# Check tag consistency
exarp-go -tool task_analysis -args '{"action":"tags"}'
```

---

## Advanced Features

### Local AI Backend Selection

All AI-powered features support backend selection:
- `fm` - Apple Foundation Models (Mac only)
- `ollama` - Ollama (recommended, cross-platform)
- `mlx` - MLX (Mac Silicon only)

```bash
exarp-go task estimate "Complex task" --local-ai-backend ollama
```

### Batch Operations

Use `task_workflow` with `tasks` array for batch creation:
```json
{
  "action": "create",
  "tasks": [
    {"name": "Task 1", "priority": "high"},
    {"name": "Task 2", "priority": "medium"}
  ]
}
```

### Recommended Tools by Task

Tasks can specify `recommended_tools` for hints:
```bash
exarp-go task create "Deploy to prod" \
  --recommended-tools "security,health,git_tools"
```

---

## Implementation Notes

**All three tools are native Go** (as of 2026-03):
- No Python bridge required
- Fast execution
- Single binary deployment
- Clean error handling

**Database**: SQLite + JSON fallback
- Primary: `.todo2/todo2.db`
- Fallback: `.todo2/state.todo2.json`
- Auto-sync on operations

**MCP Integration**: All tools available via MCP protocol
- Use from Cursor, Claude Desktop, OpenCode
- Standard JSON-RPC interface
- Streaming support for long operations

---

## See Also

- `TASK_TOOLS_COMPARISON.md` - Detailed comparison (legacy Python bridge info)
- `CLI_MAKE_CI_USAGE.md` - CLI usage guide
- `internal/cli/task.go` - CLI implementation
- `internal/tools/task_*.go` - Tool implementations
