#!/bin/bash
# Cursor MCP wrapper: delegates to start.sh (one script to rule them all).
# Preserves PROJECT_ROOT, EXARP_WATCH, EXARP_DEBUG_LOG.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export PROJECT_ROOT="${PROJECT_ROOT:-$SCRIPT_DIR}"
exec "$SCRIPT_DIR/start.sh" "$@"
