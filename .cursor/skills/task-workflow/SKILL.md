---
name: task-workflow
description: Todo2 task management via exarp-go. Prefer task_workflow MCP tool when exarp-go MCP is available; fall back to exarp-go task CLI in terminal/scripts or when MCP fails. Use when listing, updating, creating, showing, or deleting tasks; when the user mentions Todo2, task status, or .todo2; or when you must avoid editing .todo2/state.todo2.json directly.
---

# exarp-go Task Workflow (Todo2)

Apply this skill when the workspace uses exarp-go and you need to work with Todo2 tasks.

## Prefer MCP, Fallback to CLI

- **When exarp-go is available as MCP** (e.g. in Cursor): use the `task_workflow` tool first.
- **Fallback to CLI** when:
  - Running in a terminal or script (no MCP).
  - MCP call fails or exarp-go MCP is not configured.
  - User explicitly runs or requests CLI commands.

Never edit `.todo2/state.todo2.json` or `.todo2/todo2.db` directly; use the tool or CLI.

---

## MCP: task_workflow tool

When calling the exarp-go MCP server, use `task_workflow` with these patterns:

| Operation | MCP args |
|-----------|----------|
| List tasks | `action=sync`, `sub_action=list`, optional: `status`, `filter_tag`, `priority`, `limit`, `order=execution` |
| Show one task | `action=sync`, `sub_action=list`, `task_id=<id>` |
| Update status (batch) | `action=approve`, `status=<current>`, `new_status=<new>`, optional: `task_ids`, `filter_tag`, `clarification_none` |
| Create task | `action=create`, `name`, optional: `long_description`, `priority`, `tags` |
| Delete task (single) | `action=delete`, `task_id=<id>` |
| Delete tasks (batch) | `action=delete`, `task_ids=<id1>,<id2>,...` (comma-separated; one sync at end) |
| Clarity / cleanup | `action=clarity` or `action=cleanup`, optional: `task_id`, `stale_threshold_hours` |

Examples (MCP):

```json
{"action":"sync","sub_action":"list","status":"Todo","filter_tag":"migration","order":"execution"}
{"action":"approve","status":"Todo","new_status":"In Progress","task_ids":"T-1,T-2,T-3"}
{"action":"create","name":"New task","long_description":"Description","priority":"high"}
{"action":"delete","task_id":"T-1768325400284"}
{"action":"delete","task_ids":"T-321,T-401,T-402,T-403"}
```

---

## Fallback: exarp-go task CLI

When MCP is not available or you are in a terminal/script, use:

```bash
# List
exarp-go task list
exarp-go task list --status "In Progress"
exarp-go task list --status "Todo" --tag migration --order execution
exarp-go task list --priority high --limit 10

# Show one
exarp-go task show T-123
exarp-go task status T-123

# Update
exarp-go task update T-123 --new-status "Done"
exarp-go task update --status "Todo" --new-status "In Progress" --ids "T-1,T-2,T-3"

# Create
exarp-go task create "Task name" --description "Description" --priority high --tags "migration,bug"

# Delete (e.g. wrong project)
exarp-go task delete T-123
```

---

## Makefile (if present)

```bash
make task-list-todo
make task-list-in-progress
make task-update TASK_ID=T-123 NEW_STATUS=Done
```

---

## Decision flow

1. **In Cursor/IDE with exarp-go MCP?** → Use `task_workflow` tool.
2. **In terminal, script, or MCP failed?** → Use `exarp-go task <cmd>`.
3. **Advanced ops (clarity, cleanup, complex filters)?** → Prefer `task_workflow`; CLI has no direct equivalent for clarity/cleanup, so use the tool when MCP is available.
