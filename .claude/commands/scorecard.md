---
description: Run exarp-go project scorecard (fast mode)
---

Run the project scorecard via MCP:

```bash
cd /Users/davidl/Projects/exarp-go && (printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"claude","version":"1.0"}}}\n'; sleep 0.3; printf '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"report","arguments":{"action":"scorecard","fast_mode":true}}}\n'; sleep 15) | ./bin/exarp-go 2>/dev/null | python3 -c "import sys,json; [print(json.dumps(json.loads(l),indent=2)) for l in sys.stdin if l.strip() and '\"id\": 2' in l]"
```

Summarize the scorecard results: overall score, per-dimension scores, any failing checks, and recommended next actions.
