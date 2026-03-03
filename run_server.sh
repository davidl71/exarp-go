#!/bin/bash
# Run exarp-go MCP server. Delegates to start.sh (build if needed, then run).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export PROJECT_ROOT="${PROJECT_ROOT:-$SCRIPT_DIR}"
exec "$SCRIPT_DIR/start.sh" "$@"
