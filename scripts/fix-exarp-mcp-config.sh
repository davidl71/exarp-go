#!/usr/bin/env bash
# Evaluate and optionally fix exarp-go MCP configuration for Cursor.
#
# After 'make install', run this to update ~/.cursor/mcp.json or a project's
# .cursor/mcp.json so the exarp-go entry uses the installed runner and PROJECT_ROOT.
#
# Usage:
#   scripts/fix-exarp-mcp-config.sh --cursor-global              # fix ~/.cursor/mcp.json
#   scripts/fix-exarp-mcp-config.sh --cursor-project=/path/to/repo
#   scripts/fix-exarp-mcp-config.sh --cursor-global --dry-run    # show what would change
#   scripts/fix-exarp-mcp-config.sh --eval-only                  # only report current config
#
# Requires: go (for go env GOPATH), python3 (for JSON merge).

set -euo pipefail

CURSOR_GLOBAL=
CURSOR_PROJECT=
DRY_RUN=
EVAL_ONLY=
USE_BINARY=

while [[ $# -gt 0 ]]; do
  case "$1" in
    --cursor-global)     CURSOR_GLOBAL=1; shift ;;
    --cursor-project=*)  CURSOR_PROJECT="${1#--cursor-project=}"; shift ;;
    --dry-run)           DRY_RUN=1; shift ;;
    --eval-only)         EVAL_ONLY=1; shift ;;
    --use-binary)        USE_BINARY=1; shift ;;
    -h|--help)
      echo "Usage: $0 [--cursor-global] [--cursor-project=DIR] [--dry-run] [--eval-only] [--use-binary]"
      echo "  --cursor-global     Update ~/.cursor/mcp.json"
      echo "  --cursor-project=DIR  Update DIR/.cursor/mcp.json"
      echo "  --dry-run           Print changes only, do not write"
      echo "  --eval-only         Only print current exarp-go config and recommendation"
      echo "  --use-binary        Point at exarp-go binary instead of run_exarp_go.sh"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$CURSOR_GLOBAL" ]] && [[ -z "$CURSOR_PROJECT" ]] && [[ -z "$EVAL_ONLY" ]]; then
  echo "Specify at least one of: --cursor-global, --cursor-project=DIR, --eval-only" >&2
  exit 1
fi

GOPATH_BIN=""
if command -v go >/dev/null 2>&1; then
  GOPATH_BIN="$(go env GOPATH 2>/dev/null)/bin"
  [[ "$GOPATH_BIN" == "/bin" ]] && GOPATH_BIN=""
fi

RUNNER_PATH="${GOPATH_BIN}/run_exarp_go.sh"
BINARY_PATH="${GOPATH_BIN}/exarp-go"

if [[ -n "$USE_BINARY" ]]; then
  RECOMMENDED_CMD="$BINARY_PATH"
else
  RECOMMENDED_CMD="$RUNNER_PATH"
fi

eval_cursor_config() {
  local file="$1"
  local label="$2"
  echo "--- $label: $file ---"
  if [[ ! -f "$file" ]]; then
    echo "  (file not found)"
    echo "  Recommended: create with exarp-go entry using: $RECOMMENDED_CMD"
    echo ""
    return
  fi
  python3 - "$file" "$RECOMMENDED_CMD" << 'PYEOF'
import json, sys
path, rec_cmd = sys.argv[1], sys.argv[2]
try:
    with open(path) as f:
        d = json.load(f)
    servers = d.get('mcpServers') or {}
    exarp = servers.get('exarp-go')
    if exarp is None:
        print("  No exarp-go entry.")
        print("  Recommended: add exarp-go with command=%s and env.PROJECT_ROOT={{PROJECT_ROOT}}" % rec_cmd)
    else:
        print("  Current exarp-go entry:")
        for line in json.dumps(exarp, indent=2).splitlines():
            print("   ", line)
        cur_cmd = exarp.get('command', '')
        if cur_cmd == rec_cmd:
            print("  Status: command matches recommended.")
        else:
            print("  Recommended command:", rec_cmd)
except Exception as e:
    print("  (error:", e, ")")
print()
PYEOF
}

write_cursor_config() {
  local file="$1"
  local label="$2"
  local dir
  dir="$(dirname "$file")"
  if [[ -n "$DRY_RUN" ]]; then
    echo "--- $label (dry-run): $file ---"
    if [[ -f "$file" ]]; then
      echo "  Current exarp-go entry would be replaced with recommended (command=$RECOMMENDED_CMD)."
    else
      echo "  Would create $file with exarp-go entry."
    fi
    echo ""
    return
  fi
  mkdir -p "$dir"
  python3 - "$file" "$RECOMMENDED_CMD" << 'PYEOF'
import json, sys
path, rec_cmd = sys.argv[1], sys.argv[2]
entry = {
    "command": rec_cmd,
    "args": [],
    "env": {"PROJECT_ROOT": "{{PROJECT_ROOT}}"}
}
try:
    with open(path) as f:
        data = json.load(f)
except FileNotFoundError:
    data = {}
except json.JSONDecodeError:
    data = {}

if 'mcpServers' not in data:
    data['mcpServers'] = {}
data['mcpServers']['exarp-go'] = entry

with open(path, 'w') as f:
    json.dump(data, f, indent=2)
    f.write('\n')
print('Updated:', path)
PYEOF
}

if [[ -n "$EVAL_ONLY" ]]; then
  echo "Installed runner: $RUNNER_PATH"
  [[ -x "$RUNNER_PATH" ]] && echo "  (exists)" || echo "  (missing - run 'make install' or 'make install-runner')"
  echo "Installed binary: $BINARY_PATH"
  [[ -x "$BINARY_PATH" ]] && echo "  (exists)" || echo "  (missing)"
  echo ""
  if [[ -n "$CURSOR_GLOBAL" ]]; then
    eval_cursor_config "${HOME}/.cursor/mcp.json" "Cursor global"
  fi
  if [[ -n "$CURSOR_PROJECT" ]]; then
    eval_cursor_config "${CURSOR_PROJECT}/.cursor/mcp.json" "Cursor project"
  fi
  if [[ -z "$CURSOR_GLOBAL" ]] && [[ -z "$CURSOR_PROJECT" ]]; then
    eval_cursor_config "${HOME}/.cursor/mcp.json" "Cursor global"
  fi
  exit 0
fi

if [[ -n "$CURSOR_GLOBAL" ]]; then
  write_cursor_config "${HOME}/.cursor/mcp.json" "Cursor global"
fi
if [[ -n "$CURSOR_PROJECT" ]]; then
  if [[ ! -d "$CURSOR_PROJECT" ]]; then
    echo "Error: project dir not found: $CURSOR_PROJECT" >&2
    exit 1
  fi
  write_cursor_config "${CURSOR_PROJECT}/.cursor/mcp.json" "Cursor project"
fi

echo "Done. Restart Cursor or reload MCP to pick up changes."
