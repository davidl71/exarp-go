---
description: Mark a task as Done in Todo2
argument-hint: "<task-id>"
---

Mark task $ARGUMENTS as Done using the exarp-go `task_workflow` MCP tool:

Call with `action=update`, `task_id=$ARGUMENTS`, `new_status=Done`.

Then confirm by calling `task_workflow` again with `action=list` and `task_id=$ARGUMENTS` (or note the updated status from the response).
