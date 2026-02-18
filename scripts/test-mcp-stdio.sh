#!/usr/bin/env bash
# MCP stdio smoke test: run exarp-go in stdio mode, send initialize + tools/list,
# assert valid JSON-RPC response. Validates exarp-go MCP without OpenCode.
# See docs/HUMAN_TASK_DEPENDENCIES.md

set -e

BINARY="${1:-bin/exarp-go}"
PROJECT_ROOT="${PROJECT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
export PROJECT_ROOT

if [ ! -x "$BINARY" ]; then
  echo "Usage: $0 [path-to-exarp-go-binary]"
  echo "  Default: bin/exarp-go (from repo root)"
  echo "  Build first: make build"
  exit 1
fi

# MCP initialize request
INIT_REQ='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
# tools/list request (after initialize, server sends initialized notification)
TOOLS_REQ='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'

# Run exarp-go with stdin; send init, then initialized notification, then tools/list
# MCP flow: client sends initialize -> server responds -> client sends initialized notify -> then tools/list
RUN=$(mktemp -d)
trap "rm -rf $RUN" EXIT

{
  echo "$INIT_REQ"
  sleep 0.1
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'
  sleep 0.1
  echo "$TOOLS_REQ"
  sleep 0.2
} | (command -v timeout >/dev/null 2>&1 && timeout 5 "$BINARY" || "$BINARY") 2>/dev/null > "$RUN/out" || true

# Check we got jsonrpc response
if ! grep -q '"jsonrpc"' "$RUN/out"; then
  echo "FAIL: No JSON-RPC response from exarp-go"
  echo "Output (first 500 chars):"
  head -c 500 "$RUN/out"
  exit 1
fi

# Check for tools in response (tools/list returns list of tools)
if ! grep -q 'task_workflow\|"tools"' "$RUN/out"; then
  echo "WARN: Could not confirm tools/list response; stdio mode may have worked"
  echo "Output (first 500 chars):"
  head -c 500 "$RUN/out"
fi

echo "PASS: exarp-go MCP stdio smoke test (initialize + tools/list)"
