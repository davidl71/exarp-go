# Cursor Workspace Setup

**Date:** 2026-01-12  
**Purpose:** Add extension projects to Cursor workspace

---

## Current Workspace

**Active Projects:**
- ✅ `exarp-go` - MCP server (24 tools, 15 prompts, 6 resources)
- ✅ `devwisdom-go` - Wisdom MCP server
- ✅ `mcp-go-core` - (already in workspace)

**Extension Projects (Need to Add):**
- ⏳ `mcp-extension-core` - Shared extension library
- ⏳ `exarp-go-extension` - exarp-go GUI extension
- ⏳ `devwisdom-go-extension` - devwisdom-go GUI extension

---

## Adding Projects to Cursor Workspace

### Method 1: Via Cursor UI

1. **File → Add Folder to Workspace...**
2. Add these folders:
   - `/home/dlowes/projects/mcp-extension-core`
   - `/home/dlowes/projects/exarp-go-extension`
   - `/home/dlowes/projects/devwisdom-go-extension`
3. Save workspace: **File → Save Workspace As...**

### Method 2: Via Workspace File

Create or edit `.code-workspace` file:

```json
{
  "folders": [
    {
      "path": "/home/dlowes/projects/exarp-go",
      "name": "exarp-go"
    },
    {
      "path": "/home/dlowes/projects/devwisdom-go",
      "name": "devwisdom-go"
    },
    {
      "path": "/home/dlowes/projects/mcp-go-core",
      "name": "mcp-go-core"
    },
    {
      "path": "/home/dlowes/projects/mcp-extension-core",
      "name": "mcp-extension-core"
    },
    {
      "path": "/home/dlowes/projects/exarp-go-extension",
      "name": "exarp-go-extension"
    },
    {
      "path": "/home/dlowes/projects/devwisdom-go-extension",
      "name": "devwisdom-go-extension"
    }
  ],
  "settings": {}
}
```

---

## Project Verification

### ✅ exarp-go (Correct Project)

**Purpose:** MCP server for project management tools
- **24 tools** - Task management, health, reporting, etc.
- **15 prompts** - Workflow prompts and analysis
- **6 resources** - Scorecard, memories, etc.
- **Status:** ✅ Active MCP server

**Scorecard Analysis:** ✅ Correct project
- Go project with 94 files
- 24 MCP tools registered
- Project management focus

### Extension Projects

**mcp-extension-core:**
- Shared TypeScript library
- Reusable infrastructure for extensions
- MCP client, status bar, webview managers

**exarp-go-extension:**
- GUI extension for exarp-go
- Status bar, commands, webviews
- Uses mcp-extension-core

**devwisdom-go-extension:**
- GUI extension for devwisdom-go
- Wisdom quotes, advisors, briefings
- Uses mcp-extension-core

---

## Quick Add Commands

```bash
# Verify projects exist
ls -d /home/dlowes/projects/{mcp-extension-core,exarp-go-extension,devwisdom-go-extension}

# Open in Cursor (if cursor command available)
cursor /home/dlowes/projects/mcp-extension-core
cursor /home/dlowes/projects/exarp-go-extension
cursor /home/dlowes/projects/devwisdom-go-extension
```

---

**Status:** Ready to add to workspace  
**Last Updated:** 2026-01-12
