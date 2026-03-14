---
description: Create a session handoff note for exarp-go
argument-hint: "<summary of what was done>"
---

If $ARGUMENTS is empty, ask the user: "What should I include in the handoff summary?"

Otherwise, call the exarp-go `session` MCP tool with:
- `action=handoff`
- `sub_action=end`
- `summary=$ARGUMENTS`
- `include_tasks=true`
- `include_git_status=true`

Confirm the handoff was saved and show the key fields (summary, tasks, git status).
