# Human-in-the-Loop MCP Comparison

## Summary

**Found:** The `devwisdom-go` project has a **`gotohuman`** MCP server configured for human-in-the-loop functionality.

## devwisdom-go Configuration

### MCP Server: `gotohuman`

**Location:** `/Users/davidl/Projects/devwisdom-go/.cursor/mcp.json`

```json
{
  "mcpServers": {
    "gotohuman": {
      "command": "uvx",
      "args": [
        "mcpower-proxy==0.0.87",
        "--wrapped-config",
        "{\"command\": \"npx\", \"args\": [\"-y\", \"@gotohuman/mcp-server\"]}",
        "--name",
        "gotohuman"
      ],
      "description": "Human-in-the-loop platform - Allow AI agents and automations to send requests for approval to your gotoHuman inbox"
    }
  }
}
```

### What is gotoHuman?

**gotoHuman** is a human-in-the-loop platform that:
- Allows AI agents and automations to send requests for approval
- Provides an inbox for human review and approval
- Enables human-in-the-middle workflows for AI operations
- Uses MCP (Model Context Protocol) for integration

### Implementation Details

- **Command:** `uvx` (UV package runner)
- **Package:** `mcpower-proxy==0.0.87` (MCP proxy wrapper)
- **Wrapped Server:** `@gotohuman/mcp-server` (npm package)
- **Purpose:** Human approval workflow for AI operations

## Current Project (mcp-stdio-tools) Configuration

### Current MCP Servers

```json
{
  "mcpServers": {
    "advisor": {
      "command": "/Users/davidl/Projects/devwisdom-go/devwisdom",
      "description": "Crew Role: Advisor - DevWisdom Go MCP Server"
    },
    "researcher": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"],
      "description": "Crew Role: Researcher - Context7 MCP Server"
    },
    "analyst": {
      "command": "npx",
      "args": ["-y", "tractatus_thinking"],
      "description": "Crew Role: Analyst - Tractatus Thinking MCP Server"
    },
    "exarp-go": {
      "command": "/Users/davidl/Projects/mcp-stdio-tools/bin/exarp-go",
      "description": "Crew Role: Executor - Go-based MCP server"
    }
  }
}
```

### Missing: Human-in-the-Loop Server

**Status:** ❌ `gotohuman` server is **NOT** configured in the current project.

## Comparison

| Feature | devwisdom-go | mcp-stdio-tools |
|---------|--------------|-----------------|
| **Human-in-the-Loop MCP** | ✅ `gotohuman` configured | ❌ Not configured |
| **Approval Workflow** | ✅ Via gotoHuman inbox | ⚠️ Only in Todo2 workflow rules |
| **MCP Integration** | ✅ Native MCP server | ❌ Missing |
| **Workflow-Level HITL** | ✅ Todo2 rules + MCP server | ✅ Todo2 rules only |

## Human-in-the-Loop Patterns

### 1. Workflow-Level (Both Projects)

Both projects have human-in-the-loop patterns in **Todo2 workflow rules**:

```
[IN PROGRESS]
    ↓ status: Review + result
[REVIEW - AWAITING HUMAN FEEDBACK] ⚠️ HUMAN APPROVAL REQUIRED ⚠️
    ↓ human approval → status: Done
```

**Location:** `.cursor/rules/todo2.mdc`

### 2. MCP-Level (devwisdom-go Only)

**devwisdom-go** has an additional **MCP server** for human-in-the-loop:

- **Server:** `gotohuman`
- **Purpose:** Allow AI agents to request human approval via MCP tools
- **Integration:** Native MCP protocol integration
- **Workflow:** AI → gotoHuman inbox → Human approval → AI continues

## Adding gotoHuman to mcp-stdio-tools

To add human-in-the-loop MCP functionality to the current project:

### Option 1: Add gotoHuman MCP Server

Update `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "exarp-go": { ... },
    "gotohuman": {
      "command": "uvx",
      "args": [
        "mcpower-proxy==0.0.87",
        "--wrapped-config",
        "{\"command\": \"npx\", \"args\": [\"-y\", \"@gotohuman/mcp-server\"]}",
        "--name",
        "gotohuman"
      ],
      "description": "Human-in-the-loop platform - Allow AI agents and automations to send requests for approval to your gotoHuman inbox"
    }
  }
}
```

### Option 2: Implement Native Human-in-the-Loop Tools

Add tools to `exarp-go` server:

- `request_approval` - Request human approval for an action
- `check_approval_status` - Check status of approval request
- `get_pending_approvals` - List pending approvals

### Option 3: Integrate with Todo2 Workflow

Enhance Todo2 workflow tools to use gotoHuman:

- When task moves to "Review" → Send approval request to gotoHuman
- When human approves → Mark task as "Done"
- When human requests changes → Move task back to "In Progress"

## Recommendations

1. **Add gotoHuman MCP Server** - Quick integration for human-in-the-loop
2. **Enhance Todo2 Integration** - Connect Todo2 workflow with gotoHuman
3. **Native Tools** - Consider implementing approval tools in exarp-go

## References

- **gotoHuman:** `@gotohuman/mcp-server` (npm package)
- **MCP Proxy:** `mcpower-proxy==0.0.87` (UV package)
- **devwisdom-go Config:** `/Users/davidl/Projects/devwisdom-go/.cursor/mcp.json`

---

**Last Updated:** 2026-01-07
**Status:** ✅ Analysis complete - gotoHuman found in devwisdom-go

