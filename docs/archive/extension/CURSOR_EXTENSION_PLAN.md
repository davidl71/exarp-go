# Cursor Extension for exarp-go

> **⚠️ FUTURE GOAL - NOT CURRENTLY IMPLEMENTED**  
> This document outlines a potential future enhancement. The cursor extension is not currently being developed. The MCP server works standalone and provides all core functionality through Cursor's built-in MCP integration.

## Overview

This document outlines the architecture and implementation plan for a Cursor/VS Code extension to complement the exarp-go MCP server. The extension would provide visual UI, better discoverability, and enhanced user experience for accessing project management tools, task workflows, health monitoring, and automation capabilities.

**Status**: Future enhancement - not in active development  
**Priority**: Very Low  
**Estimated Total Effort**: 40-50 hours

## Current State

### MCP Server (Implemented)
- ✅ Core MCP server with **24 tools**
- ✅ **18 prompts** (8 core + 7 workflow + 2 context + 1 task management)
- ✅ **6 resources** (scorecard, memories by category/task/session/recent)
- ✅ Accessible via Cursor's MCP integration
- ✅ Native Go implementation with Python bridge for complex tools

### Limitations
- No visual UI or persistent display
- Requires AI chat interaction to access
- No status bar indicators
- No command palette shortcuts
- No sidebar panels
- No task visualization
- No dependency graph display
- No health metrics dashboard

## Extension Value Proposition

### User Benefits
1. **Visual Feedback**: See project health, task status, and metrics at a glance
2. **Task Visualization**: View task dependencies, critical path, and hierarchy
3. **Discoverability**: Easy access via command palette and status bar
4. **Persistent Awareness**: Project health visible without chat interaction
5. **Better Integration**: Native Cursor UI components
6. **Notifications**: Proactive alerts for health issues, stale tasks, lock expiration
7. **Quick Actions**: One-click access to common workflows

### Technical Benefits
1. **TypeScript/Node.js**: Leverages existing ecosystem
2. **VS Code API**: Rich UI components available
3. **MCP Integration**: Can call MCP tools directly via stdio
4. **Extension Host**: Isolated from main process
5. **Reusable Architecture**: Similar to devwisdom-go extension plan

## Architecture

### Extension Structure

```
exarp-go-extension/
├── package.json              # Extension manifest
├── tsconfig.json             # TypeScript config
├── .vscodeignore             # Files to exclude from package
├── src/
│   ├── extension.ts         # Main entry point
│   ├── mcpClient.ts         # MCP server communication
│   ├── statusBar.ts         # Status bar integration
│   ├── commands.ts          # Command palette handlers
│   ├── sidebar/
│   │   ├── healthView.ts   # Project health dashboard
│   │   ├── taskView.ts     # Task management panel
│   │   ├── memoryView.ts   # Memory browser
│   │   └── toolCatalogView.ts # Tool catalog browser
│   ├── webviews/
│   │   ├── scorecardWebview.ts # Scorecard visualization
│   │   ├── taskGraphWebview.ts # Dependency graph visualization
│   │   └── reportWebview.ts    # Report viewer
│   └── notifications.ts     # Notification system
├── media/
│   └── icons/               # Extension icons
├── README.md                 # Extension documentation
└── CHANGELOG.md              # Version history
```

### MCP Communication

The extension will communicate with the exarp-go MCP server using:

1. **Direct MCP Protocol**: Use MCP client library to call tools
2. **Stdio Transport**: Connect to `bin/exarp-go` binary via stdio
3. **Resource Access**: Read `stdio://` resources for data
4. **JSON-RPC 2.0**: Standard MCP protocol

### Key Components

#### 1. Status Bar Integration
- Display current project health score (color-coded)
- Show active task count (Todo, In Progress, Review)
- Display stale lock count (if any)
- Click to open project scorecard
- Right-click for quick actions menu

#### 2. Command Palette
**Project Management:**
- `Exarp: Get Project Scorecard`
- `Exarp: Run Health Check`
- `Exarp: Generate Project Report`
- `Exarp: View Project Overview`

**Task Management:**
- `Exarp: Show Task Dashboard`
- `Exarp: View Active Tasks`
- `Exarp: Analyze Task Dependencies`
- `Exarp: Find Critical Path`
- `Exarp: Discover Tasks`
- `Exarp: Sync Tasks`

**Workflow:**
- `Exarp: Daily Check-in`
- `Exarp: Sprint Start`
- `Exarp: Sprint End`
- `Exarp: Pre-Sprint Cleanup`

**Tools:**
- `Exarp: Run Linter`
- `Exarp: Security Scan`
- `Exarp: Run Tests`
- `Exarp: Check Stale Locks`
- `Exarp: Memory Browser`
- `Exarp: Tool Catalog`

#### 3. Sidebar Panel
**Project Health Dashboard:**
- Overall health score with breakdown
- Key metrics (testing, documentation, security, etc.)
- Trend indicators (improving/declining)
- Quick action buttons

**Task Management Panel:**
- Task list with filtering (status, priority, tag)
- Task status breakdown (Todo, In Progress, Review, Done)
- Dependency graph visualization
- Critical path highlight
- Quick task actions (create, update, assign)

**Memory Browser:**
- Browse memories by category
- Filter by task ID
- View recent memories
- Search memories
- Session-based memory view

**Tool Catalog:**
- Browse all 24 tools
- Tool descriptions and parameters
- Quick access to tool documentation
- Tool usage examples

#### 4. Webviews
**Scorecard Visualization:**
- Interactive project scorecard
- Metric breakdowns
- Historical trends (if available)
- Export options

**Task Dependency Graph:**
- Visual dependency graph
- Critical path highlighting
- Task status indicators
- Interactive exploration

**Report Viewer:**
- Formatted project reports
- Export to markdown/PDF
- Print-friendly view

#### 5. Notifications
- Project health alerts (low scores)
- Stale task notifications
- Lock expiration warnings
- Daily standup reminders
- Sprint milestone alerts
- Security vulnerability alerts

## Implementation Phases

### Phase 1: Foundation (MVP)
**Priority**: Very Low  
**Estimated Time**: 12-16 hours

**Goals:**
- Basic extension structure
- MCP client integration
- Status bar with health score
- Single command: "Get Project Scorecard"
- Basic error handling

**Deliverables:**
- Extension package.json with activation events
- TypeScript setup and build configuration
- MCP stdio client implementation
- Status bar component with health score
- One working command with output display
- Basic logging and error handling

**Key Files:**
- `package.json` - Extension manifest
- `tsconfig.json` - TypeScript configuration
- `src/extension.ts` - Main entry point
- `src/mcpClient.ts` - MCP communication
- `src/statusBar.ts` - Status bar integration
- `src/commands.ts` - Command handlers

### Phase 2: Command Palette
**Priority**: Very Low  
**Estimated Time**: 10-14 hours

**Goals:**
- All command palette commands
- Command handlers for key tools
- Output formatting (webview or output channel)
- Error handling and user feedback

**Deliverables:**
- 15+ command handlers
- Output views (webview for complex data, output channel for simple)
- Progress indicators for long-running operations
- Error messages and recovery suggestions

**Commands to Implement:**
1. Project Management (4 commands)
2. Task Management (6 commands)
3. Workflow (4 commands)
4. Tools (5+ commands)

### Phase 3: Sidebar Panels
**Priority**: Very Low  
**Estimated Time**: 12-16 hours

**Goals:**
- Webview-based sidebar panels
- Project health dashboard
- Task management panel
- Memory browser
- Tool catalog

**Deliverables:**
- Sidebar panel implementation
- Four view components (health, tasks, memory, tools)
- Data refresh mechanism (polling + manual)
- Interactive UI elements

**Views:**
1. **Health Dashboard**: Score, metrics, trends
2. **Task Panel**: List, filters, status breakdown
3. **Memory Browser**: Category/task/session views
4. **Tool Catalog**: Tool list with descriptions

### Phase 4: Advanced Visualizations
**Priority**: Very Low  
**Estimated Time**: 8-10 hours

**Goals:**
- Task dependency graph visualization
- Critical path display
- Scorecard interactive view
- Report viewer

**Deliverables:**
- Graph visualization library integration
- Task graph webview
- Scorecard webview
- Report viewer webview
- Export functionality

### Phase 5: Notifications & Polish
**Priority**: Very Low  
**Estimated Time**: 6-8 hours

**Goals:**
- Notification system
- Settings/configuration UI
- Icon assets
- Documentation
- Marketplace preparation

**Deliverables:**
- Notification handlers
- Settings UI (VS Code settings)
- Extension README
- Marketplace assets (icons, screenshots)
- CHANGELOG

## Technical Decisions

### MCP Client Library
**Options:**
1. **@modelcontextprotocol/sdk** (Official MCP SDK)
   - Pros: Official, well-maintained, TypeScript, stdio support
   - Cons: May be overkill for simple use case
2. **Custom stdio client**
   - Pros: Lightweight, full control, minimal dependencies
   - Cons: More implementation work, protocol handling

**Recommendation**: Start with official SDK (`@modelcontextprotocol/sdk`), fallback to custom if needed

**Implementation:**
```typescript
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

const transport = new StdioClientTransport({
  command: "/path/to/bin/exarp-go",
  args: [],
  env: { PROJECT_ROOT: workspaceRoot }
});

const client = new Client({
  name: "exarp-go-extension",
  version: "0.1.0"
}, {
  capabilities: {}
});

await client.connect(transport);
```

### UI Framework
**Options:**
1. **VS Code Webview API** (Native)
   - Pros: Integrated, no external deps, lightweight
   - Cons: Limited styling options, manual DOM manipulation
2. **React + Webview**
   - Pros: Rich UI, component reuse, better DX
   - Cons: Larger bundle, more complexity, build setup

**Recommendation**: Start with native Webview API for MVP, consider React for complex visualizations (task graph, scorecard)

### Data Refresh Strategy
**Options:**
- **Polling**: Check for updates every N minutes
- **Event-driven**: Listen to file changes (`.todo2/state.todo2.json`, `.todo2/todo2.db`)
- **Manual**: User-triggered refresh
- **Hybrid**: Combination of above

**Recommendation**: Hybrid approach
- **Polling**: Health scores, task counts (every 5 minutes)
- **Event-driven**: File watchers for `.todo2/` directory changes
- **Manual**: Refresh button in sidebar panels
- **On-demand**: Refresh when panel becomes visible

### Graph Visualization
**Options:**
1. **D3.js** - Powerful, flexible, complex
2. **vis.js** - Simpler, good for graphs
3. **Cytoscape.js** - Graph-specific, good performance
4. **Custom SVG** - Full control, lightweight

**Recommendation**: Start with **vis.js** or **Cytoscape.js** for task dependency graphs (simpler API, good performance)

## Dependencies

### Required
```json
{
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.0.0",
    "@types/vscode": "^1.85.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.3.0",
    "vscode-test": "^1.6.0"
  }
}
```

### Optional (for advanced features)
```json
{
  "dependencies": {
    "vis-network": "^9.1.0",  // For task graph visualization
    "react": "^18.0.0",       // If using React
    "react-dom": "^18.0.0"
  }
}
```

## Configuration

### Extension Settings

```json
{
  "exarp.showStatusBar": {
    "type": "boolean",
    "default": true,
    "description": "Show project health in status bar"
  },
  "exarp.statusBarHealthScore": {
    "type": "boolean",
    "default": true,
    "description": "Show health score in status bar"
  },
  "exarp.statusBarTaskCount": {
    "type": "boolean",
    "default": true,
    "description": "Show active task count in status bar"
  },
  "exarp.refreshInterval": {
    "type": "number",
    "default": 300,
    "description": "Auto-refresh interval in seconds"
  },
  "exarp.healthScoreThreshold": {
    "type": "number",
    "default": 70,
    "description": "Health score threshold for alerts"
  },
  "exarp.notifications.enabled": {
    "type": "boolean",
    "default": true,
    "description": "Enable notifications"
  },
  "exarp.notifications.staleTasks": {
    "type": "boolean",
    "default": true,
    "description": "Notify about stale tasks"
  },
  "exarp.notifications.lockExpiration": {
    "type": "boolean",
    "default": true,
    "description": "Notify about lock expiration"
  },
  "exarp.mcpServerPath": {
    "type": "string",
    "default": "./bin/exarp-go",
    "description": "Path to exarp-go binary (relative to workspace root)"
  }
}
```

## User Stories

### US-EXT-1: Status Bar Health Indicator
*As a* developer,  
*I want* to see my project health score in the status bar,  
*So that* I'm aware of project status at a glance.

### US-EXT-2: Task Dashboard Command
*As a* developer,  
*I want* to run a command to see my task dashboard,  
*So that* I can quickly view active tasks and their status.

### US-EXT-3: Dependency Graph Visualization
*As a* developer,  
*I want* to visualize task dependencies in a graph,  
*So that* I can understand task relationships and critical path.

### US-EXT-4: Health Score Notification
*As a* developer,  
*I want* to receive notifications when health score drops below threshold,  
*So that* I'm alerted to project issues.

### US-EXT-5: Quick Task Actions
*As a* developer,  
*I want* to create and update tasks from the sidebar,  
*So that* I can manage tasks without using chat.

### US-EXT-6: Memory Browser
*As a* developer,  
*I want* to browse AI memories by category or task,  
*So that* I can review past insights and decisions.

### US-EXT-7: Tool Catalog
*As a* developer,  
*I want* to browse available tools and their descriptions,  
*So that* I can discover what tools are available.

### US-EXT-8: Workflow Shortcuts
*As a* developer,  
*I want* quick access to daily check-in and sprint workflows,  
*So that* I can run common workflows efficiently.

## Cursor Extension Requirements Research

### Key Findings

1. **VS Code API Compatibility**
   - Cursor is based on VS Code
   - Extensions use VS Code Extension API
   - Standard VS Code extension development applies
   - Reference: [VS Code Extension API](https://code.visualstudio.com/api)

2. **MCP Integration**
   - Extensions can call MCP tools directly
   - Use MCP SDK for stdio communication
   - Resources accessible via MCP protocol
   - Reference: [MCP Specification](https://modelcontextprotocol.io/)

3. **Extension API (Enterprise)**
   - Cursor provides Extension API for programmatic MCP server registration
   - Useful for enterprise environments
   - Not required for standard extensions
   - Reference: [Cursor MCP Extension API](https://docs.cursor.com/en/context/mcp-extension-api)

4. **Activation Events**
   - Use `onStartupFinished` for background extensions
   - Use `onCommand` for command-based activation
   - Use `onView` for sidebar panel activation
   - Minimize activation overhead

5. **Performance Considerations**
   - Lazy load heavy components
   - Use webview pooling for multiple panels
   - Cache MCP responses when appropriate
   - Minimize stdio communication overhead

### Cursor-Specific Notes

- **MCP Server Path**: Use workspace-relative paths when possible
- **Environment Variables**: `PROJECT_ROOT` is automatically set by Cursor
- **Extension Host**: Runs in isolated process
- **Error Handling**: Log to Output Channel for debugging
- **Testing**: Use `vscode-test` for extension testing

## Success Criteria

### MVP Success
- ✅ Extension installs and activates
- ✅ Status bar shows health score
- ✅ One command works end-to-end (Get Project Scorecard)
- ✅ MCP communication functional
- ✅ Basic error handling

### Phase 2 Success
- ✅ All command palette commands implemented
- ✅ Output formatting works (webview + output channel)
- ✅ Error handling comprehensive
- ✅ User feedback clear

### Phase 3 Success
- ✅ All sidebar panels functional
- ✅ Data refresh working (polling + manual)
- ✅ Interactive UI elements working
- ✅ Performance acceptable

### Full Success
- ✅ All features implemented
- ✅ Visualizations working (graph, scorecard)
- ✅ Notifications functional
- ✅ Settings UI complete
- ✅ Documentation complete
- ✅ User feedback positive
- ✅ Extension published to marketplace (optional)

## Risks & Mitigations

### Risk: MCP Communication Complexity
**Impact**: High  
**Probability**: Medium  
**Mitigation**: 
- Start with official MCP SDK
- Implement robust error handling
- Add retry logic for failed connections
- Provide clear error messages

### Risk: Performance Overhead
**Impact**: Medium  
**Probability**: Medium  
**Mitigation**: 
- Lazy loading of components
- Efficient polling intervals
- Cache responses when appropriate
- Minimize stdio communication

### Risk: Maintenance Burden
**Impact**: High  
**Probability**: Medium  
**Mitigation**: 
- Keep scope minimal initially
- Focus on core features
- Document architecture clearly
- Consider community contributions

### Risk: User Adoption
**Impact**: Low  
**Probability**: Low  
**Mitigation**: 
- Gather feedback early
- Iterate based on usage
- Provide clear documentation
- Make extension optional (MCP server works standalone)

### Risk: Graph Visualization Complexity
**Impact**: Medium  
**Probability**: Medium  
**Mitigation**: 
- Use proven libraries (vis.js, Cytoscape.js)
- Start with simple visualization
- Add complexity incrementally
- Provide fallback to text view

## Related Documentation

### Existing
- `docs/CURSOR_MCP_SETUP.md` - MCP server setup guide
- `docs/GUI_EXTENSION_STATUS.md` - Extension status investigation
- `.cursor/rules/mcp-configuration.mdc` - MCP configuration guidelines

### Reference
- `devwisdom-go/docs/CURSOR_EXTENSION.md` - Similar extension plan
- [VS Code Extension API](https://code.visualstudio.com/api)
- [MCP Specification](https://modelcontextprotocol.io/)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)
- [Cursor Extension Guidelines](https://docs.cursor.com) (if available)

## Implementation Checklist

### Phase 1: Foundation
- [ ] Create extension project structure
- [ ] Set up TypeScript configuration
- [ ] Implement MCP client with stdio transport
- [ ] Create status bar component
- [ ] Implement "Get Project Scorecard" command
- [ ] Add basic error handling
- [ ] Test extension activation

### Phase 2: Commands
- [ ] Implement all command handlers
- [ ] Create output views (webview + output channel)
- [ ] Add progress indicators
- [ ] Implement error messages
- [ ] Test all commands

### Phase 3: Sidebar
- [ ] Create sidebar panel structure
- [ ] Implement health dashboard view
- [ ] Implement task management panel
- [ ] Implement memory browser
- [ ] Implement tool catalog
- [ ] Add data refresh mechanism
- [ ] Test all panels

### Phase 4: Visualizations
- [ ] Integrate graph visualization library
- [ ] Implement task dependency graph
- [ ] Create scorecard webview
- [ ] Create report viewer
- [ ] Add export functionality
- [ ] Test visualizations

### Phase 5: Polish
- [ ] Implement notification system
- [ ] Create settings UI
- [ ] Design and add icons
- [ ] Write extension README
- [ ] Create marketplace assets
- [ ] Prepare CHANGELOG

## Notes

- This extension is **optional** - MCP server works standalone
- Priority is **very low** - focus on core MCP functionality first
- Can be implemented incrementally
- Consider community contributions if interest exists
- Similar architecture to devwisdom-go extension (can share patterns)
- More complex than devwisdom-go (24 tools vs 5 tools)

## Comparison with devwisdom-go Extension

| Aspect | devwisdom-go | exarp-go |
|--------|--------------|----------|
| **Tools** | 5 tools | 24 tools |
| **Complexity** | Simple | Complex |
| **MVP Time** | 8-12 hours | 12-16 hours |
| **Total Time** | 24-34 hours | 40-50 hours |
| **Key Features** | Wisdom quotes, briefings | Tasks, health, workflows |
| **Visualizations** | Simple (quotes, sources) | Complex (graphs, dashboards) |

---

**Status**: Documented for future implementation  
**Priority**: Very Low  
**Last Updated**: 2026-01-12
