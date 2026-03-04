#!/usr/bin/env python3
# Parse JSON-RPC (NDJSON) from stdin; print the response object whose "id" matches TARGET_ID.
# Usage: ... | ./bin/exarp-go 2>/dev/null | python3 scripts/parse_mcp_response.py [TARGET_ID]
# Default TARGET_ID is 2 (typical second request, e.g. tools/call).

import json
import sys

def main():
    target_id = int(sys.argv[1]) if len(sys.argv) > 1 else 2
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
            if obj.get("id") == target_id:
                print(json.dumps(obj, indent=2))
                return 0
        except json.JSONDecodeError:
            continue
    return 1

if __name__ == "__main__":
    sys.exit(main())
