---
description: Create a session handoff note for exarp-go
argument-hint: "<summary of what was done>"
---

Create a session handoff using the exarp-go session tool:

```bash
cd /Users/davidl/Projects/exarp-go && (printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"claude","version":"1.0"}}}\n'; sleep 0.3; printf '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"session","arguments":{"action":"handoff","sub_action":"end","summary":"$ARGUMENTS","include_tasks":true,"include_git_status":true}}}\n'; sleep 6) | ./bin/exarp-go 2>/dev/null | python3 -c "import sys,json; [print(json.dumps(json.loads(l),indent=2)) for l in sys.stdin if l.strip() and '\"id\": 2' in l]"
```

If $ARGUMENTS is empty, ask the user: "What should I include in the handoff summary?"

Confirm the handoff was saved and show the key fields (summary, tasks, git status).
