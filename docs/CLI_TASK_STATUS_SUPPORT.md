# CLI Task Status Support

**Date:** 2026-01-12  
**Status:** Current Implementation + Enhancement Proposal

## Current Support

### ✅ Task Operations via `task_workflow` Tool

The CLI supports task status operations through the generic tool execution interface using the `task_workflow` tool.

**Available via:**
```bash
exarp-go -tool task_workflow -args '<json>'
```

### Task Workflow Actions

**1. List Tasks by Status**
```bash
# List all "In Progress" tasks
exarp-go -tool task_workflow -args '{"action":"sync","sub_action":"list","status":"In Progress"}'

# List all "Todo" tasks
exarp-go -tool task_workflow -args '{"action":"sync","sub_action":"list","status":"Todo"}'

# List all "Done" tasks
exarp-go -tool task_workflow -args '{"action":"sync","sub_action":"list","status":"Done"}'
```

**2. Update Task Status (Batch)**
```bash
# Update multiple tasks from "Todo" to "Done"
exarp-go -tool task_workflow -args '{
  "action":"approve",
  "status":"Todo",
  "new_status":"Done",
  "task_ids":"[\"T-1\",\"T-2\",\"T-3\"]"
}'

# Update single task to "In Progress"
exarp-go -tool task_workflow -args '{
  "action":"approve",
  "status":"Todo",
  "new_status":"In Progress",
  "task_ids":"[\"T-123\"]"
}'
```

**3. Create New Task**
```bash
exarp-go -tool task_workflow -args '{
  "action":"create",
  "name":"Task name",
  "long_description":"Task description",
  "tags":["tag1","tag2"],
  "priority":"high"
}'
```

**4. Get Task Details**
```bash
# Get specific task by ID
exarp-go -tool task_workflow -args '{
  "action":"sync",
  "sub_action":"list",
  "task_id":"T-123"
}'
```

## Current Limitations

### ❌ No Convenience Commands

The CLI currently requires:
- Full JSON argument syntax
- Knowledge of tool names and action parameters
- Manual JSON escaping in shell

**Example (verbose):**
```bash
exarp-go -tool task_workflow -args '{"action":"sync","sub_action":"list","status":"In Progress"}'
```

**Could be simplified to:**
```bash
exarp-go task list --status "In Progress"
# or
exarp-go task status "In Progress"
```

## Enhancement Proposal

### Add Dedicated Task Commands

**Proposed CLI Structure:**
```bash
# Task listing
exarp-go task list                    # List all tasks
exarp-go task list --status "Todo"    # List by status
exarp-go task list --priority "high"  # List by priority
exarp-go task list --tag "migration"  # List by tag

# Task status operations
exarp-go task status <task-id>                    # Get task status
exarp-go task update <task-id> --status "Done"   # Update single task
exarp-go task update --status "Todo" --new-status "Done" --ids "T-1,T-2"  # Batch update

# Task creation
exarp-go task create "Task name" --description "Description" --priority "high"

# Task details
exarp-go task show <task-id>          # Show full task details
```

### Implementation Approach

**Option 1: Add Subcommands to CLI**
- Add `task` subcommand parser
- Map to `task_workflow` tool internally
- Provide friendly argument parsing

**Option 2: Extend Interactive Mode**
- Add task-specific commands in interactive mode
- `task list`, `task status <id>`, etc.

**Option 3: Shell Aliases/Wrapper Script**
- Create wrapper script with convenience functions
- Keep core CLI generic

## Recommended Approach

**Hybrid: Add Subcommands + Keep Tool Interface**

1. **Add `task` subcommand** for common operations
2. **Keep `-tool` interface** for advanced/scripted use
3. **Both use same underlying tool** (`task_workflow`)

**Benefits:**
- ✅ Convenience for common operations
- ✅ Power users can still use `-tool` directly
- ✅ Consistent with existing tool infrastructure
- ✅ No duplication of logic

## Implementation Plan

### Phase 1: Basic Task Commands

**Add to `internal/cli/cli.go`:**

```go
// New flag
var taskCmd = flag.String("task", "", "Task command (list|status|update|create|show)")

// In Run() function:
case *taskCmd != "":
    return handleTaskCommand(server, *taskCmd, flag.Args())
```

**New function:**
```go
func handleTaskCommand(server framework.MCPServer, cmd string, args []string) error {
    switch cmd {
    case "list":
        return handleTaskList(server, args)
    case "status":
        return handleTaskStatus(server, args)
    case "update":
        return handleTaskUpdate(server, args)
    case "create":
        return handleTaskCreate(server, args)
    case "show":
        return handleTaskShow(server, args)
    default:
        return fmt.Errorf("unknown task command: %s", cmd)
    }
}
```

### Phase 2: Argument Parsing

Use `flag` package for subcommand arguments:
```go
// task list --status "In Progress" --limit 10
// task update T-1 --status "Done"
// task create "Name" --description "Desc" --priority "high"
```

### Phase 3: Interactive Mode Enhancement

Add task commands to interactive mode:
```
exarp-go> task list
exarp-go> task status T-123
exarp-go> task update T-123 --status Done
```

## Examples

### Current (Verbose)
```bash
# List tasks
exarp-go -tool task_workflow -args '{"action":"sync","sub_action":"list","status":"In Progress"}'

# Update status
exarp-go -tool task_workflow -args '{"action":"approve","status":"Todo","new_status":"Done","task_ids":"[\"T-1\"]"}'
```

### Proposed (Convenient)
```bash
# List tasks
exarp-go task list --status "In Progress"

# Update status
exarp-go task update T-1 --status "Done"

# Batch update
exarp-go task update --status "Todo" --new-status "Done" --ids "T-1,T-2,T-3"
```

## Priority

**High Value Enhancement:**
- ✅ Common operation (task status management)
- ✅ Reduces verbosity significantly
- ✅ Improves developer experience
- ✅ Low implementation complexity (wrapper around existing tool)

## Status

**Current:** ✅ Convenience commands implemented  
**Fallback:** ✅ Direct `task_workflow` tool calls still supported for advanced operations

---

## Quick Reference

### Current Usage

```bash
# List all In Progress tasks
exarp-go -tool task_workflow -args '{"action":"sync","sub_action":"list","status":"In Progress"}'

# Update task T-1 to Done
exarp-go -tool task_workflow -args '{"action":"approve","status":"Todo","new_status":"Done","task_ids":"[\"T-1\"]"}'

# Create new task
exarp-go -tool task_workflow -args '{"action":"create","name":"Task name","long_description":"Description"}'

# Get task details
exarp-go -tool task_workflow -args '{"action":"sync","sub_action":"list","task_id":"T-123"}'
```

### All Task Workflow Actions

- `sync` (sub_action: `list`) - List/filter tasks
- `approve` - Batch update task status
- `create` - Create new task
- `clarity` - Improve task clarity
- `cleanup` - Clean up stale tasks
