---
description: Mark a task as Done in Todo2
argument-hint: "<task-id>"
---

Mark the task $ARGUMENTS as Done:

```bash
cd /Users/davidl/Projects/exarp-go && go run ./cmd/server -tool task_workflow -args "{\"action\":\"update\",\"task_ids\":\"$ARGUMENTS\",\"new_status\":\"Done\"}" 2>/dev/null
```

Then confirm the status change by showing the task:

```bash
cd /Users/davidl/Projects/exarp-go && go run ./cmd/server task show $ARGUMENTS 2>/dev/null
```
