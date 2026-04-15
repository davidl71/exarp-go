---
description: Prime the session with exarp-go context (tasks, hints, handoffs)
---

Call the exarp-go `session` MCP tool with `action=prime`, `include_hints=true`, `include_tasks=true`.

Since MCP tools from other servers aren't directly callable in this context, run:

```bash
cd /Users/davidl/Projects/exarp-go && (printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"claude","version":"1.0"}}}\n'; sleep 0.3; printf '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"session","arguments":{"action":"prime","include_hints":true,"include_tasks":true}}}\n'; sleep 6) | ./bin/exarp-go 2>/dev/null | python3 -c "import sys,json; [print(json.dumps(json.loads(l),indent=2)) for l in sys.stdin if l.strip() and '\"id\": 2' in l]"
```

Parse and summarize the result: show current focus task, suggested next tasks, task counts, and any hints or handoffs.
