#!/bin/bash
# Wrapper script to run exarp-go MCP server
# Handles path resolution for Go binary

set -e

# Find project root (exarp-go directory); respect PROJECT_ROOT if set by caller (e.g. AI agent)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${PROJECT_ROOT:-$SCRIPT_DIR}"

# Change to project root
cd "$PROJECT_ROOT"

# Run Go binary
exec "$PROJECT_ROOT/bin/exarp-go"

