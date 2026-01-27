---
name: task-workflow
description: Todo2 and task management via exarp-go. Use when listing, updating, creating, or showing tasks; when the user mentions Todo2, task status, or .todo2; or when you must avoid editing .todo2/state.todo2.json directly.
---

# exarp-go Task Workflow (Todo2)

Apply this skill when the workspace uses exarp-go and you need to work with Todo2 tasks.

## Prefer Convenience Commands

Use `exarp-go task` subcommands for everyday operations:

```bash
# List tasks
exarp-go task list --status "In Progress"
exarp-go task list --status "Todo"
exarp-go task list --priority "high"

# Update task status
exarp-go task update T-123 --new-status "Done"
exarp-go task update --status "Todo" --new-status "Done" --ids "T-1,T-2"

# Create task
exarp-go task create "Task name" --description "Description" --priority "high"

# Show task details
exarp-go task show T-123
exarp-go task status T-123
```

## Fallback: Direct Tool Calls

Use the `task_workflow` MCP tool for advanced operations or when convenience commands don’t support what you need:

- **Clarity / cleanup:** `task_workflow` with `action=clarity` or `action=cleanup`, and `task_id` or `stale_threshold_hours` as needed.
- **Complex filters:** `task_workflow` with `action=sync`, `sub_action=list`, and filters like `status`, `filter_tag`.
- **Batch / scripting:** `task_workflow` with JSON args when automation or custom logic is required.

Example:

```json
{"action":"sync","sub_action":"list","status":"Review","filter_tag":"migration"}
```

## Rules

1. **Prefer** `exarp-go task` for list, update, create, show.
2. **Fallback** to `task_workflow` for clarity, cleanup, and complex filters.
3. **Never** edit `.todo2/state.todo2.json` or `.todo2/todo2.db` directly; always use the CLI or `task_workflow`.
4. **When in doubt**, use convenience commands first.

## Makefile (if present)

If the repo has a Makefile with task targets, you can use them too, e.g.:

```bash
make task-list-todo
make task-list-in-progress
make task-update TASK_ID=T-123 NEW_STATUS=Done
```
