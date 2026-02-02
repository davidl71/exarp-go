#!/usr/bin/env bash
###############################################################################
# Generate TUI demo assets using agent-tui
#
# Drives exarp-go TUI (Bubble Tea) via agent-tui: run TUI, send key sequence,
# capture screenshots. Outputs go to docs/demo/ (text or JSON).
#
# Prerequisites:
#   - agent-tui installed and on PATH (https://github.com/pproenca/agent-tui)
#   - exarp-go built: make build (binary at bin/exarp-go)
#
# Usage:
#   From repo root: ./scripts/demo-tui-agent-tui.sh
###############################################################################

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEMO_DIR="$PROJECT_ROOT/docs/demo"
BINARY_PATH="$PROJECT_ROOT/bin/exarp-go"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

if ! command -v agent-tui &>/dev/null; then
  echo -e "${RED}❌ agent-tui not found on PATH.${NC}"
  echo "Install: curl -fsSL https://raw.githubusercontent.com/pproenca/agent-tui/master/install.sh | bash"
  echo "Or: npm install -g agent-tui"
  exit 1
fi

if [[ ! -f "$BINARY_PATH" ]]; then
  echo -e "${RED}❌ exarp-go binary not found: $BINARY_PATH${NC}"
  echo "Run: make build"
  exit 1
fi

mkdir -p "$DEMO_DIR"
cd "$PROJECT_ROOT"

echo -e "${BLUE}=== agent-tui TUI demo ===${NC}"

# Start daemon if not running
if ! agent-tui daemon status &>/dev/null; then
  echo "Starting agent-tui daemon..."
  agent-tui daemon start
  sleep 2
fi

# Run exarp-go TUI under agent-tui in background (session stays alive)
echo "Starting exarp-go TUI under agent-tui..."
agent-tui run "$BINARY_PATH" tui &
RUN_PID=$!
trap 'agent-tui kill 2>/dev/null; wait "$RUN_PID" 2>/dev/null; exit 0' EXIT

# Give TUI time to load
sleep 3

# Wait for TUI to show (Todo / In Progress / status bar)
agent-tui wait "Todo" -t 5000 2>/dev/null || agent-tui wait "q" -t 2000 2>/dev/null || true

# Step 1: initial screen
agent-tui screenshot --no-color 2>/dev/null > "$DEMO_DIR/step1.txt" || agent-tui screenshot 2>/dev/null > "$DEMO_DIR/step1.txt"
echo "Captured step1.txt"

# Step 2: navigate down
agent-tui press ArrowDown ArrowDown
sleep 1
agent-tui screenshot --no-color 2>/dev/null > "$DEMO_DIR/step2.txt" || agent-tui screenshot 2>/dev/null > "$DEMO_DIR/step2.txt"
echo "Captured step2.txt"

# Step 3: optional Enter (task detail) if any task selected
agent-tui press Enter
sleep 1
agent-tui screenshot --no-color 2>/dev/null > "$DEMO_DIR/step3.txt" || agent-tui screenshot 2>/dev/null > "$DEMO_DIR/step3.txt"
echo "Captured step3.txt"

# End session (trap will also kill)
agent-tui kill 2>/dev/null || true
wait "$RUN_PID" 2>/dev/null || true
trap - EXIT

echo -e "${GREEN}Done. Demo outputs in $DEMO_DIR/${NC}"
ls -la "$DEMO_DIR"
