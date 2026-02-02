#!/bin/bash

###############################################################################
# MCP Server Sanity Check
# 
# Verifies that the expected number of tools, resources, and prompts are
# registered and visible through the MCP protocol.
#
# Usage:
#   ./scripts/sanity-check.sh
###############################################################################

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Expected counts (from codebase; tools = 30 base or 31 with Apple FM on darwin/arm64/cgo)
EXPECTED_TOOLS=30
EXPECTED_PROMPTS=35
EXPECTED_RESOURCES=23

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== MCP Server Sanity Check ===${NC}"
echo ""

# Check if binary exists
BINARY_PATH="$PROJECT_ROOT/bin/exarp-go"
if [[ ! -f "$BINARY_PATH" ]]; then
    echo -e "${RED}❌ Binary not found: $BINARY_PATH${NC}"
    echo "Run: go build -o $BINARY_PATH ./cmd/server"
    exit 1
fi

# Start server in background
echo -e "${BLUE}Starting MCP server...${NC}"
"$BINARY_PATH" > /tmp/exarp-go-stdout.log 2> /tmp/exarp-go-stderr.log &
SERVER_PID=$!

# Wait for server to be ready
sleep 1

# Cleanup function
cleanup() {
    echo ""
    echo -e "${BLUE}Cleaning up...${NC}"
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    rm -f /tmp/exarp-go-stdout.log /tmp/exarp-go-stderr.log
}

trap cleanup EXIT

# Function to send JSON-RPC request
send_request() {
    local method="$1"
    local params="${2:-{}}"
    
    local request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "$method",
  "params": $params
}
EOF
)
    
    echo "$request" | "$BINARY_PATH" 2>/dev/null | jq -r '.result' 2>/dev/null || echo ""
}

# Initialize MCP connection
echo -e "${BLUE}Initializing MCP connection...${NC}"
initialize_request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "sanity-check",
      "version": "1.0.0"
    }
  }
}
EOF
)

# Send initialize
echo "$initialize_request" > /tmp/mcp-init.json
response=$(cat /tmp/mcp-init.json | timeout 2 "$BINARY_PATH" 2>/dev/null || echo "")

if [[ -z "$response" ]]; then
    echo -e "${RED}❌ Failed to communicate with server${NC}"
    echo "Server logs:"
    cat /tmp/exarp-go-stderr.log
    exit 1
fi

# Send initialized notification
initialized_notification=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "notifications/initialized"
}
EOF
)

echo "$initialized_notification" > /tmp/mcp-notify.json

# Wait a bit for server to be fully ready
sleep 0.5

# Query tools list
echo -e "${BLUE}Querying tools...${NC}"
tools_request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}
EOF
)

tools_response=$(echo "$tools_request" | timeout 2 "$BINARY_PATH" 2>/dev/null || echo "")
actual_tools=$(echo "$tools_response" | jq -r '.result.tools | length' 2>/dev/null || echo "0")

# Query prompts list
echo -e "${BLUE}Querying prompts...${NC}"
prompts_request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "prompts/list",
  "params": {}
}
EOF
)

prompts_response=$(echo "$prompts_request" | timeout 2 "$BINARY_PATH" 2>/dev/null || echo "")
actual_prompts=$(echo "$prompts_response" | jq -r '.result.prompts | length' 2>/dev/null || echo "0")

# Query resources list
echo -e "${BLUE}Querying resources...${NC}"
resources_request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "resources/list",
  "params": {}
}
EOF
)

resources_response=$(echo "$resources_request" | timeout 2 "$BINARY_PATH" 2>/dev/null || echo "")
actual_resources=$(echo "$resources_response" | jq -r '.result.resources | length' 2>/dev/null || echo "0")

# Print results
echo ""
echo -e "${BLUE}=== Results ===${NC}"
echo ""

# Tools
if [[ "$actual_tools" == "$EXPECTED_TOOLS" ]]; then
    echo -e "${GREEN}✅ Tools: $actual_tools/$EXPECTED_TOOLS${NC}"
else
    echo -e "${RED}❌ Tools: $actual_tools/$EXPECTED_TOOLS (MISMATCH)${NC}"
fi

# Prompts
if [[ "$actual_prompts" == "$EXPECTED_PROMPTS" ]]; then
    echo -e "${GREEN}✅ Prompts: $actual_prompts/$EXPECTED_PROMPTS${NC}"
else
    echo -e "${RED}❌ Prompts: $actual_prompts/$EXPECTED_PROMPTS (MISMATCH)${NC}"
fi

# Resources
if [[ "$actual_resources" == "$EXPECTED_RESOURCES" ]]; then
    echo -e "${GREEN}✅ Resources: $actual_resources/$EXPECTED_RESOURCES${NC}"
else
    echo -e "${RED}❌ Resources: $actual_resources/$EXPECTED_RESOURCES (MISMATCH)${NC}"
fi

# Summary
echo ""
if [[ "$actual_tools" == "$EXPECTED_TOOLS" ]] && \
   [[ "$actual_prompts" == "$EXPECTED_PROMPTS" ]] && \
   [[ "$actual_resources" == "$EXPECTED_RESOURCES" ]]; then
    echo -e "${GREEN}✅ All counts match!${NC}"
    exit 0
else
    echo -e "${RED}❌ Count mismatches detected!${NC}"
    echo ""
    echo "Details:"
    if [[ "$actual_tools" != "$EXPECTED_TOOLS" ]]; then
        echo -e "${YELLOW}Tools: Expected $EXPECTED_TOOLS, got $actual_tools${NC}"
    fi
    if [[ "$actual_prompts" != "$EXPECTED_PROMPTS" ]]; then
        echo -e "${YELLOW}Prompts: Expected $EXPECTED_PROMPTS, got $actual_prompts${NC}"
    fi
    if [[ "$actual_resources" != "$EXPECTED_RESOURCES" ]]; then
        echo -e "${YELLOW}Resources: Expected $EXPECTED_RESOURCES, got $actual_resources${NC}"
    fi
    exit 1
fi

