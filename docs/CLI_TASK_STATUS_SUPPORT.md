# CLI Task Status Support

**Date:** 2026-03-24  
**Status:** Current Implementation

## Current Support

### ✅ Dedicated `task` Commands

The CLI already supports direct task operations without going through raw JSON:

```bash
exarp-go task list --status "In Progress"
exarp-go task show T-123
exarp-go task update --ids T-123 --new-status Done
exarp-go task create "Task name" --description "Task description"
```

### ✅ Advanced Execution Operations via `task_workflow`

The generic tool interface is still the advanced surface for execution-cockpit actions:

```bash
exarp-go -tool task_workflow -args '<json>'
```

### Task Workflow Actions

**1. List Tasks by Status**
```bash
exarp-go task list --status "In Progress"
exarp-go task list --status "Todo"
exarp-go task list --status "Done"
```

**2. Update Task Status (Batch)**
```bash
exarp-go task update --ids T-1,T-2,T-3 --new-status Done
exarp-go task update --ids T-123 --new-status "In Progress"
```

**3. Create New Task**
```bash
exarp-go task create "Task name" --description "Task description" --priority high
```

**4. Get Task Details**
```bash
exarp-go task show T-123
```

**5. Start an Execution Run**
```bash
exarp-go -tool task_workflow -args '{"action":"start_run","task_id":"T-123","summary":"Implement handlers"}'
```

**6. Record Verification and Partial Progress**
```bash
exarp-go -tool task_workflow -args '{"action":"verify","task_id":"T-123","run_id":"R-...","kind":"compile","result":"passed","command":"go build ./internal/tools"}'
exarp-go -tool task_workflow -args '{"action":"add_progress","task_id":"T-123","run_id":"R-...","summary":"Wired handlers","remaining_work":"Update docs"}'
```
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
