#!/usr/bin/env bash
# Validate OpenCode config for exarp-go MCP integration.
# Checks config file exists, exarp-go entry present, command path, PROJECT_ROOT.
# See docs/HUMAN_TASK_DEPENDENCIES.md, docs/OPENCODE_INTEGRATION.md

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

CONFIG_GLOBAL="${HOME}/.config/opencode/opencode.json"
CONFIG_GLOBAL_JSONC="${HOME}/.config/opencode/opencode.jsonc"
CONFIG_PROJECT="opencode.json"
ERRORS=0

check() {
  if [ "$1" = 0 ]; then
    printf "${GREEN}✓${NC} %s\n" "$2"
  else
    printf "${RED}✗${NC} %s\n" "$2"
    ((ERRORS++)) || true
  fi
}

warn() {
  printf "${YELLOW}⚠${NC} %s\n" "$1"
}

# Resolve config path (project overrides global)
CONFIG=""
if [ -f "$CONFIG_PROJECT" ]; then
  CONFIG="$CONFIG_PROJECT"
elif [ -f "$CONFIG_GLOBAL" ]; then
  CONFIG="$CONFIG_GLOBAL"
elif [ -f "$CONFIG_GLOBAL_JSONC" ]; then
  CONFIG="$CONFIG_GLOBAL_JSONC"
fi

if [ -z "$CONFIG" ]; then
  printf "${RED}✗${NC} No OpenCode config found. Expected:\n"
  printf "  - %s\n" "$CONFIG_GLOBAL"
  printf "  - %s (project root)\n" "$CONFIG_PROJECT"
  exit 1
fi

check 0 "Config file exists: $CONFIG"

# Parse JSON (strip comments if jsonc)
RAW=$(cat "$CONFIG")
if [[ "$CONFIG" == *.jsonc ]]; then
  RAW=$(echo "$RAW" | sed 's|//.*||' | sed 's|/\*.*\*/||')
fi

if ! echo "$RAW" | python3 -c "import sys, json; json.load(sys.stdin)" 2>/dev/null; then
  check 1 "Config is valid JSON"
  exit 1
fi
check 0 "Config is valid JSON"

# Check exarp-go in mcp
HAS_MCP=$(echo "$RAW" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    mcp = d.get('mcp') or {}
    eg = mcp.get('exarp-go') or {}
    print('name' if eg else '')
    if eg:
        cmd = eg.get('command')
        env = eg.get('environment') or {}
        proot = env.get('PROJECT_ROOT')
        print('cmd:' + str(cmd))
        print('proot:' + str(proot))
except Exception as e:
    print('err:' + str(e))
" 2>/dev/null || echo "")

if [ -z "$HAS_MCP" ] || [ "$HAS_MCP" = "err:"* ]; then
  check 1 "exarp-go MCP entry present in mcp"
  if [[ "$HAS_MCP" == err:* ]]; then
    warn "Parse error: $HAS_MCP"
  fi
  exit 1
fi
check 0 "exarp-go MCP entry present in mcp"

# Extract command and PROJECT_ROOT
CMD=$(echo "$RAW" | python3 -c "
import sys, json
d = json.load(sys.stdin)
eg = (d.get('mcp') or {}).get('exarp-go') or {}
cmd = eg.get('command')
if isinstance(cmd, list):
    cmd = cmd[0] if cmd else ''
elif not isinstance(cmd, str):
    cmd = ''
print(cmd)
" 2>/dev/null)
PROOT=$(echo "$RAW" | python3 -c "
import sys, json
d = json.load(sys.stdin)
eg = (d.get('mcp') or {}).get('exarp-go') or {}
env = eg.get('environment') or {}
print(env.get('PROJECT_ROOT', ''))
" 2>/dev/null)

if [ -z "$CMD" ] || [ "$CMD" = "/absolute/path/to/exarp-go/bin/exarp-go" ] || [ "$CMD" = "/path/to/bin/exarp-go" ]; then
  warn "command is placeholder. Set to actual exarp-go binary path."
else
  # Resolve path (expand PROJECT_ROOT or ~ if used)
  RESOLVED="$CMD"
  if [[ "$RESOLVED" == *'{{PROJECT_ROOT}}'* ]] || [[ "$RESOLVED" == *'\${PROJECT_ROOT}'* ]]; then
    if [ -n "$PROOT" ] && [ -d "$PROOT" ]; then
      RESOLVED=$(echo "$RESOLVED" | sed "s|{{PROJECT_ROOT}}|$PROOT|g" | sed "s|\${PROJECT_ROOT}|$PROOT|g")
    else
      warn "PROJECT_ROOT not set or invalid; cannot resolve command path."
    fi
  fi
  if [ -n "$RESOLVED" ] && [ -f "$RESOLVED" ] && [ -x "$RESOLVED" ]; then
    check 0 "command path exists and is executable"
  elif [ "$CMD" = "exarp-go" ] || [[ "$CMD" == *"exarp-go" ]]; then
    if command -v exarp-go >/dev/null 2>&1; then
      check 0 "exarp-go on PATH"
    else
      warn "command is 'exarp-go' but not found on PATH"
    fi
  else
    warn "command path not verified: $CMD"
  fi
fi

if [ -n "$PROOT" ] && [ "$PROOT" != "/path/to/your/project" ]; then
  if [ -d "$PROOT" ]; then
    check 0 "PROJECT_ROOT exists: $PROOT"
  else
    check 1 "PROJECT_ROOT exists: $PROOT"
  fi
else
  warn "PROJECT_ROOT is placeholder. Set to your project root."
fi

if [ "$ERRORS" -gt 0 ]; then
  exit 1
fi
echo ""
printf "${GREEN}All checks passed. Run OpenCode and verify tools (task_workflow, report, session).${NC}\n"
