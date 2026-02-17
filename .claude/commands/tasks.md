---
description: List current Todo2 tasks (default: Todo + In Progress, high priority first)
argument-hint: "[--status <status>] [--priority <priority>]"
---

Run the following to list tasks, then display them in a clear table:

```bash
cd /Users/davidl/Projects/exarp-go && go run ./cmd/server task list --status "$ARGUMENTS" 2>/dev/null || go run ./cmd/server task list 2>/dev/null
```

If $ARGUMENTS is empty, show all Todo and In Progress tasks:

```bash
cd /Users/davidl/Projects/exarp-go && go run ./cmd/server task list --status "Todo" 2>/dev/null
```

Display results as a readable table with ID, Priority, Status, and Content (truncated to 60 chars).
