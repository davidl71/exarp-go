---
description: List current Todo2 tasks (default: Todo + In Progress, high priority first)
argument-hint: "[--status <status>] [--priority <priority>]"
---

Call the exarp-go `task_workflow` MCP tool to list tasks:

- If $ARGUMENTS is empty or not provided: call with `action=list`, `status=Todo`
- If $ARGUMENTS contains a status (e.g. "In Progress", "Done"): call with `action=list`, `status=<that value>`
- If $ARGUMENTS contains a priority: call with `action=list`, `priority=<that value>`

Display results as a readable table with ID, Priority, Status, and Content (truncated to 60 chars).
