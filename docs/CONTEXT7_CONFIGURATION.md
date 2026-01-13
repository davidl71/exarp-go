# Context7 MCP Configuration Instructions

**Date:** 2026-01-13  
**Purpose:** Configure Context7 MCP server in Cursor IDE for researching Go MCP client libraries

---

## Current Configuration

Your `.cursor/mcp.json` currently has:
- `advisor` - DevWisdom Go MCP Server
- `exarp-go` - Exarp Go MCP Server

---

## How to Add Context7

### Option 1: Manual Edit (Recommended)

1. Open `.cursor/mcp.json` in Cursor
2. Add the `context7` server configuration:

```json
{
  "mcpServers": {
    "advisor": {
      "command": "/home/dlowes/projects/devwisdom-go/run-advisor.sh",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Crew Role: Advisor - DevWisdom Go MCP Server - Wisdom quotes, trusted advisors, inspirational guidance"
    },
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"],
      "description": "Crew Role: Researcher - Context7 MCP Server - Documentation lookup and research"
    },
    "exarp-go": {
      "command": "/home/dlowes/projects/exarp-go/run_server.sh",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Crew Role: Executor - Go-based MCP server - 24 tools, 15 prompts, 6 resources"
    }
  }
}
```

3. Save the file
4. **Restart Cursor IDE** for changes to take effect

### Option 2: Using Terminal (if you prefer)

```bash
cd /home/dlowes/projects/exarp-go

# Backup current config
cp .cursor/mcp.json .cursor/mcp.json.backup-$(date +%Y%m%d)

# Use jq to add context7 (if jq is installed)
# Or manually edit as shown above
```

---

## Verification

After restarting Cursor:

1. Check Cursor's MCP status:
   - Settings → Features → Model Context Protocol
   - Verify `context7` appears in the list

2. Test context7 with a prompt:
   ```
   How does github.com/modelcontextprotocol/go-sdk support MCP clients? use context7
   ```

---

## Using Context7 for Research

Once configured, you can use context7 to research:

### Go MCP Client Libraries

1. **Official SDK Client Support:**
   - "How does github.com/modelcontextprotocol/go-sdk support MCP client functionality? use context7"
   - "Does Model Context Protocol Go SDK have client capabilities? use context7"

2. **Alternative Go Libraries:**
   - "What Go libraries exist for MCP client functionality? use context7"
   - "Golang Model Context Protocol client implementation examples use context7"

3. **Implementation Patterns:**
   - "How to implement MCP client in Go using JSON-RPC 2.0? use context7"
   - "MCP client pattern Go subprocess stdio use context7"

---

## Context7 Tools

Context7 provides these tools (via MCP):
- `resolve-library-id` - Resolve library name to Context7 ID
- `query-docs` - Query library documentation
- `get-library-info` - Get comprehensive library information

---

## Next Steps

1. ✅ Add context7 to `.cursor/mcp.json` (see above)
2. ✅ Restart Cursor IDE
3. ⏳ Use context7 to research Go MCP client libraries
4. ⏳ Document findings in `docs/MCP_CLIENT_RESEARCH.md`
5. ⏳ Implement solution based on research

---

## Troubleshooting

### Context7 Not Appearing in Cursor

1. Verify JSON syntax is valid:
   ```bash
   python3 -m json.tool .cursor/mcp.json
   ```

2. Check for Node.js/npx:
   ```bash
   which npx
   npx --version
   ```

3. Test context7 manually:
   ```bash
   npx -y @upstash/context7-mcp
   ```

4. Restart Cursor IDE completely (quit and reopen)

### Context7 Not Responding

1. Check Cursor's MCP logs:
   - Settings → Features → Model Context Protocol → Logs
   - Look for context7 errors

2. Verify npx can run the package:
   ```bash
   npx -y @upstash/context7-mcp --help
   ```

---

## References

- **Context7 Documentation:** https://context7.com/docs
- **MCP Protocol:** https://modelcontextprotocol.io
- **Research Plan:** `docs/MCP_CLIENT_RESEARCH.md`
