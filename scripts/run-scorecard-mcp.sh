#!/usr/bin/env bash
# Run exarp-go project scorecard via JSON-RPC over stdio (no Cursor MCP).
# Use when MCP is not available (e.g. terminal, CI). Requires PROJECT_ROOT.
# Usage: ./scripts/run-scorecard-mcp.sh [project_root]
# Example: PROJECT_ROOT=/path/to/project ./scripts/run-scorecard-mcp.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROJECT_ROOT="${PROJECT_ROOT:-$(pwd)}"
if [[ $# -ge 1 ]]; then
  PROJECT_ROOT="$1"
fi
export PROJECT_ROOT

EXARP_GO_BIN="${REPO_ROOT}/bin/exarp-go"
if [[ ! -x "${EXARP_GO_BIN}" ]]; then
  echo "Build exarp-go first: make b" >&2
  exit 1
fi

# Initialize + tools/call report scorecard; parse response with id 2
(
  printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"cli","version":"1.0"}}}'
  sleep 0.3
  printf '%s\n' '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"report","arguments":{"action":"scorecard","fast_mode":true}}}'
  sleep 15
) | "${EXARP_GO_BIN}" 2>/dev/null | python3 "${SCRIPT_DIR}/parse_mcp_response.py" 2
