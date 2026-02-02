# Cursor Extension Requirements Research

**Date:** 2026-01-12  
**Status:** ✅ Research Complete

---

## Summary

Research conducted on Cursor extension development requirements, MCP integration patterns, and best practices for building GUI extensions that complement MCP servers.

## Key Findings

### 1. VS Code API Compatibility

**Finding:** Cursor is based on VS Code and uses the VS Code Extension API.

**Implications:**
- Standard VS Code extension development applies
- All VS Code Extension API features available
- Extension marketplace compatibility
- Standard tooling and workflows

**References:**
- [VS Code Extension API Documentation](https://code.visualstudio.com/api)
- [Extension API Overview](https://code.visualstudio.com/api/get-started/your-first-extension)

**Key APIs for exarp-go Extension:**
- `vscode.window.createStatusBarItem()` - Status bar integration
- `vscode.commands.registerCommand()` - Command palette
- `vscode.window.createWebviewPanel()` - Sidebar panels
- `vscode.window.createOutputChannel()` - Output display
- `vscode.workspace.createFileSystemWatcher()` - File change detection

### 2. MCP Integration

**Finding:** Extensions can call MCP tools directly using the MCP SDK.

**Implications:**
- Direct stdio communication with MCP servers
- Access to all tools, prompts, and resources
- No need for intermediate API layer
- Standard JSON-RPC 2.0 protocol

**MCP SDK Usage:**
```typescript
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

// Connect to exarp-go MCP server
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

// Call MCP tools
const result = await client.callTool({
  name: "health",
  arguments: {}
});
```

**References:**
- [MCP Specification](https://modelcontextprotocol.io/)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)
- [MCP SDK Documentation](https://modelcontextprotocol.io/docs/sdks/typescript)

### 3. Cursor Extension API (Enterprise)

**Finding:** Cursor provides an Extension API for programmatic MCP server registration.

**Use Case:** Enterprise environments and onboarding tools that need to dynamically configure MCP servers.

**Key Features:**
- Programmatic MCP server registration
- No direct `mcp.json` file modification required
- Useful for enterprise deployments
- Onboarding automation

**Note:** Not required for standard extensions. Standard extensions can work with `.cursor/mcp.json` configuration.

**Reference:**
- [Cursor MCP Extension API](https://docs.cursor.com/en/context/mcp-extension-api)

**For exarp-go Extension:**
- Standard extension approach (using `.cursor/mcp.json`) is sufficient
- Extension API only needed if dynamic server registration required
- Not necessary for initial implementation

### 4. Activation Events

**Finding:** Extensions should use appropriate activation events to minimize overhead.

**Best Practices:**
- `onStartupFinished` - For background extensions (health monitoring)
- `onCommand` - For command-based activation (lazy loading)
- `onView` - For sidebar panel activation (lazy loading)
- Avoid `*` activation (activates on every startup)

**For exarp-go Extension:**
```json
{
  "activationEvents": [
    "onStartupFinished",  // For status bar
    "onCommand:exarp.getScorecard",  // Lazy load on command
    "onView:exarpHealthDashboard"  // Lazy load on view
  ]
}
```

**Benefits:**
- Faster Cursor startup
- Lower memory usage
- Better user experience

### 5. Performance Considerations

**Finding:** Extensions should minimize overhead and optimize for performance.

**Key Considerations:**

1. **Lazy Loading**
   - Load components only when needed
   - Defer heavy operations
   - Use activation events effectively

2. **Webview Pooling**
   - Reuse webview panels when possible
   - Dispose unused webviews
   - Limit concurrent webviews

3. **Caching**
   - Cache MCP responses when appropriate
   - Cache health scores and task data
   - Invalidate cache on file changes

4. **Polling Efficiency**
   - Use reasonable intervals (5+ minutes)
   - Stop polling when panel not visible
   - Use file watchers instead of polling when possible

5. **Stdio Communication**
   - Batch multiple tool calls when possible
   - Handle connection errors gracefully
   - Implement retry logic

**For exarp-go Extension:**
- Poll health score every 5 minutes (configurable)
- Use file watchers for `.todo2/` directory
- Cache tool catalog (rarely changes)
- Lazy load graph visualization

### 6. Error Handling

**Finding:** Extensions should provide clear error messages and recovery options.

**Best Practices:**
- Log errors to Output Channel
- Show user-friendly error messages
- Provide recovery suggestions
- Handle MCP server unavailability gracefully

**For exarp-go Extension:**
```typescript
// Error handling pattern
try {
  const result = await client.callTool({ name: "health", arguments: {} });
} catch (error) {
  if (error.code === "ECONNREFUSED") {
    vscode.window.showErrorMessage(
      "Cannot connect to exarp-go MCP server. Is it running?",
      "Check Configuration"
    );
  } else {
    outputChannel.appendLine(`Error: ${error.message}`);
    vscode.window.showErrorMessage(`Exarp: ${error.message}`);
  }
}
```

### 7. Extension Testing

**Finding:** Use `vscode-test` for extension testing.

**Testing Approach:**
- Unit tests for core logic
- Integration tests for MCP communication
- E2E tests for UI components
- Mock MCP server for testing

**For exarp-go Extension:**
```typescript
import * as vscode from 'vscode';
import * as assert from 'assert';

suite('Extension Test Suite', () => {
  test('Should activate extension', async () => {
    const extension = vscode.extensions.getExtension('exarp-go.extension');
    assert.ok(extension);
    await extension.activate();
    assert.ok(extension.isActive);
  });
});
```

### 8. Extension Marketplace

**Finding:** Extensions can be published to VS Code Marketplace (Cursor compatible).

**Requirements:**
- Valid `package.json` manifest
- README with screenshots
- CHANGELOG.md
- License file
- Icon assets
- Marketplace metadata

**For exarp-go Extension:**
- Optional: Can be published if there's community interest
- Primary: Local development and use
- Consider: Open source on GitHub

## Technical Stack Recommendations

### Core Dependencies
```json
{
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.0.0",
    "@types/vscode": "^1.85.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.3.0",
    "vscode-test": "^1.6.0",
    "@typescript-eslint/eslint-plugin": "^6.0.0",
    "@typescript-eslint/parser": "^6.0.0",
    "eslint": "^8.0.0"
  }
}
```

### Optional Dependencies (for visualizations)
```json
{
  "dependencies": {
    "vis-network": "^9.1.0",  // Task dependency graphs
    "react": "^18.0.0",      // If using React
    "react-dom": "^18.0.0"
  }
}
```

## Architecture Patterns

### 1. MCP Client Singleton
```typescript
class MCPClientManager {
  private static instance: MCPClientManager;
  private client: Client | null = null;

  static getInstance(): MCPClientManager {
    if (!MCPClientManager.instance) {
      MCPClientManager.instance = new MCPClientManager();
    }
    return MCPClientManager.instance;
  }

  async getClient(): Promise<Client> {
    if (!this.client) {
      await this.connect();
    }
    return this.client!;
  }

  private async connect() {
    // Connect to exarp-go MCP server
  }
}
```

### 2. Status Bar Manager
```typescript
class StatusBarManager {
  private statusBarItem: vscode.StatusBarItem;

  constructor() {
    this.statusBarItem = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Right,
      100
    );
    this.statusBarItem.command = 'exarp.showScorecard';
  }

  async update() {
    const healthScore = await this.getHealthScore();
    this.statusBarItem.text = `$(heart) Exarp: ${healthScore}`;
    this.statusBarItem.show();
  }
}
```

### 3. Webview Panel Manager
```typescript
class WebviewPanelManager {
  private panels: Map<string, vscode.WebviewPanel> = new Map();

  getOrCreatePanel(viewType: string, title: string): vscode.WebviewPanel {
    let panel = this.panels.get(viewType);
    if (!panel) {
      panel = vscode.window.createWebviewPanel(
        viewType,
        title,
        vscode.ViewColumn.Two,
        { enableScripts: true }
      );
      panel.onDidDispose(() => this.panels.delete(viewType));
      this.panels.set(viewType, panel);
    }
    return panel;
  }
}
```

## Security Considerations

### 1. Input Validation
- Validate all user inputs
- Sanitize file paths
- Prevent path traversal attacks

### 2. MCP Communication
- Validate MCP server responses
- Handle malformed JSON gracefully
- Timeout long-running operations

### 3. Webview Security
- Use Content Security Policy (CSP)
- Validate webview messages
- Sanitize HTML content

## Development Workflow

### 1. Setup
```bash
# Create extension project
npm init -y
npm install --save-dev @types/vscode typescript
npm install @modelcontextprotocol/sdk

# Initialize TypeScript
tsc --init
```

### 2. Development
```bash
# Watch mode
npm run watch

# Test extension
# Press F5 in VS Code/Cursor to launch Extension Development Host
```

### 3. Packaging
```bash
# Install vsce (VS Code Extension manager)
npm install -g @vscode/vsce

# Package extension
vsce package
```

## Comparison with devwisdom-go Extension

| Aspect | devwisdom-go | exarp-go |
|--------|--------------|----------|
| **MCP Tools** | 5 tools | 24 tools |
| **Complexity** | Simple | Complex |
| **Visualizations** | Basic (quotes) | Advanced (graphs, dashboards) |
| **MCP SDK** | Same | Same |
| **Architecture** | Similar | Similar (more components) |

## Key Takeaways

1. ✅ **VS Code API Compatibility** - Standard extension development applies
2. ✅ **MCP SDK Available** - Direct stdio communication supported
3. ✅ **Performance Matters** - Use lazy loading and efficient polling
4. ✅ **Error Handling Critical** - Provide clear user feedback
5. ✅ **Testing Important** - Use vscode-test for integration tests
6. ✅ **Marketplace Optional** - Can be local development tool

## Next Steps

1. **Review Extension Plan** - See `docs/CURSOR_EXTENSION_PLAN.md`
2. **Start with MVP** - Implement Phase 1 (Foundation)
3. **Iterate Incrementally** - Add features phase by phase
4. **Gather Feedback** - Test with real usage scenarios
5. **Document Learnings** - Update this document with findings

## References

### Official Documentation
- [VS Code Extension API](https://code.visualstudio.com/api)
- [MCP Specification](https://modelcontextprotocol.io/)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)
- [Cursor MCP Extension API](https://docs.cursor.com/en/context/mcp-extension-api)

### Community Resources
- [VS Code Extension Samples](https://github.com/microsoft/vscode-extension-samples)
- [MCP SDK Examples](https://github.com/modelcontextprotocol/typescript-sdk/tree/main/examples)

---

**Status**: ✅ Research Complete  
**Last Updated**: 2026-01-12
