#!/usr/bin/env bash
# Validate Cursor plugin structure: .cursor-plugin/plugin.json and mcp.json.
# See docs/HUMAN_TASK_DEPENDENCIES.md, docs/CURSOR_PLUGIN_PLAN.md

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$ROOT"
ERRORS=0

check() {
  if [ "$1" = 0 ]; then
    printf "${GREEN}✓${NC} %s\n" "$2"
  else
    printf "${RED}✗${NC} %s\n" "$2"
    ((ERRORS++)) || true
  fi
}

# .cursor-plugin/plugin.json
MANIFEST=".cursor-plugin/plugin.json"
if [ ! -f "$MANIFEST" ]; then
  check 1 "Manifest exists: $MANIFEST"
  exit 1
fi
check 0 "Manifest exists: $MANIFEST"

# Required manifest fields
for field in name displayName description version license; do
  if python3 -c "
import json
with open('$MANIFEST') as f:
    d = json.load(f)
v = d.get('$field')
if v is None or (isinstance(v, str) and not v.strip()):
    exit(1)
" 2>/dev/null; then
    check 0 "Manifest has '$field'"
  else
    check 1 "Manifest has '$field'"
  fi
done

# name must be lowercase kebab-case
if python3 -c "
import json, re
with open('$MANIFEST') as f:
    d = json.load(f)
n = d.get('name', '')
if not re.match(r'^[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?$', n):
    exit(1)
" 2>/dev/null; then
  check 0 "Manifest name is lowercase kebab-case"
else
  check 1 "Manifest name must be lowercase kebab-case"
fi

# mcp.json
if [ ! -f "mcp.json" ]; then
  check 1 "mcp.json exists"
  exit 1
fi
check 0 "mcp.json exists"

if ! python3 -c "
import json
with open('mcp.json') as f:
    d = json.load(f)
if 'mcpServers' not in d or 'exarp-go' not in d.get('mcpServers', {}):
    exit(1)
" 2>/dev/null; then
  check 1 "mcp.json has exarp-go server entry"
else
  check 0 "mcp.json has exarp-go server entry"
fi

if [ "$ERRORS" -gt 0 ]; then
  exit 1
fi
echo ""
printf "${GREEN}Plugin validation passed. Run 'make validate-plugin' for full check.${NC}\n"
