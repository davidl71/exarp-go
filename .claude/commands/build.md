---
description: Build exarp-go and run sanity check
---

Build the project and verify counts:

```bash
cd /Users/davidl/Projects/exarp-go && make build 2>&1 && make sanity-check 2>&1
```

Report: build success/failure, any errors, and sanity check results (tool/prompt/resource counts).
