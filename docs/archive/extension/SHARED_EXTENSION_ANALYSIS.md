# Shared GUI Extension Analysis

**Date:** 2026-01-12  
**Question:** Should exarp-go and devwisdom-go share a single Cursor extension?  
**Status:** ✅ Analysis Complete

---

## Executive Summary

**Recommendation: ⚠️ CONDITIONAL YES - With Caveats**

A shared extension makes sense **IF** structured as a **multi-server extension framework** with pluggable server adapters. However, separate extensions may be simpler initially.

**Key Finding:** Both projects are designed to work together (exarp-go uses devwisdom-go for briefings), but have different complexity levels and use cases.

---

## Current Relationship

### Integration Points

1. **exarp-go → devwisdom-go**
   - exarp-go's `report` tool calls devwisdom-go for briefings
   - Complementary "crew" pattern (Executor + Advisor)
   - Both configured in same `.cursor/mcp.json`

2. **Shared Architecture Patterns**
   - Both use framework-agnostic design
   - Both have CLI/MCP dual mode
   - Both use similar Go patterns
   - Both follow similar MCP protocol implementation

3. **Different Purposes**
   - **exarp-go**: Project management, tasks, health, workflows (24 tools)
   - **devwisdom-go**: Wisdom quotes, advisors, guidance (5 tools)

---

## Shared Extension Architecture

### Option 1: Multi-Server Extension Framework ⭐ **RECOMMENDED**

**Structure:**
```
mcp-extension-framework/
├── package.json
├── src/
│   ├── core/
│   │   ├── mcpClient.ts        # Generic MCP client
│   │   ├── statusBar.ts         # Status bar manager
│   │   ├── webviewManager.ts    # Webview pooling
│   │   └── notificationManager.ts # Notifications
│   ├── adapters/
│   │   ├── exarpAdapter.ts      # exarp-go specific UI
│   │   └── devwisdomAdapter.ts  # devwisdom-go specific UI
│   ├── extension.ts             # Main entry point
│   └── config.ts                # Multi-server configuration
```

**Benefits:**
- ✅ Code reuse for common infrastructure
- ✅ Unified user experience
- ✅ Shared MCP client, status bar, webview patterns
- ✅ Easy to add more MCP servers later
- ✅ Single extension to install

**Drawbacks:**
- ⚠️ More complex architecture
- ⚠️ Need adapter pattern for server-specific UI
- ⚠️ Larger codebase to maintain

**Implementation:**
```typescript
// Core MCP client (reusable)
class MCPClientManager {
  private clients: Map<string, Client> = new Map();
  
  async getClient(serverName: string, config: ServerConfig): Promise<Client> {
    // Reusable client logic
  }
}

// Server-specific adapters
class ExarpAdapter {
  createHealthDashboard(): WebviewPanel { }
  createTaskPanel(): WebviewPanel { }
  // exarp-go specific UI
}

class DevWisdomAdapter {
  createBriefingView(): WebviewPanel { }
  createSourcesBrowser(): WebviewPanel { }
  // devwisdom-go specific UI
}
```

### Option 2: Separate Extensions

**Structure:**
```
exarp-go-extension/
└── (standalone extension)

devwisdom-go-extension/
└── (standalone extension)
```

**Benefits:**
- ✅ Simpler architecture
- ✅ Independent development
- ✅ No coupling between projects
- ✅ Easier to maintain separately
- ✅ Can be installed independently

**Drawbacks:**
- ❌ Code duplication (MCP client, status bar, etc.)
- ❌ Two extensions to install
- ❌ Less unified experience

---

## Code Reuse Analysis

### Shared Components (High Reuse Potential)

| Component | Reuse % | Complexity | Notes |
|-----------|---------|------------|-------|
| **MCP Client** | 95% | Medium | Generic stdio client, server-specific config |
| **Status Bar Manager** | 90% | Low | Generic status bar, server-specific data |
| **Webview Manager** | 85% | Medium | Generic pooling, server-specific content |
| **Notification System** | 80% | Low | Generic notifications, server-specific triggers |
| **Command Registry** | 70% | Medium | Generic command handling, server-specific commands |
| **Settings UI** | 60% | Low | Generic settings, server-specific options |

### Server-Specific Components (Low Reuse)

| Component | Reuse % | Complexity | Notes |
|-----------|---------|------------|-------|
| **exarp-go UI** | 0% | High | Task dashboard, dependency graphs, health metrics |
| **devwisdom-go UI** | 0% | Medium | Briefing view, sources browser, consultation history |
| **Tool Handlers** | 10% | High | Different tools, different parameters |
| **Data Models** | 20% | Medium | Different data structures (tasks vs wisdom) |

**Estimated Code Reuse:** ~60-70% of infrastructure code

---

## Complexity Comparison

### Shared Extension Complexity

**Infrastructure Code:**
- MCP client: ~500 lines (shared)
- Status bar: ~200 lines (shared)
- Webview manager: ~300 lines (shared)
- Notification system: ~200 lines (shared)
- **Total shared: ~1,200 lines**

**Server-Specific Code:**
- exarp-go adapter: ~1,500 lines (24 tools, complex UI)
- devwisdom-go adapter: ~800 lines (5 tools, simpler UI)
- **Total server-specific: ~2,300 lines**

**Total: ~3,500 lines**

### Separate Extensions Complexity

**exarp-go Extension:**
- Infrastructure: ~1,200 lines
- exarp-go specific: ~1,500 lines
- **Total: ~2,700 lines**

**devwisdom-go Extension:**
- Infrastructure: ~1,200 lines
- devwisdom-go specific: ~800 lines
- **Total: ~2,000 lines**

**Combined: ~4,700 lines** (duplication overhead)

**Savings with Shared:** ~1,200 lines (25% reduction)

---

## Use Case Analysis

### Unified Experience Benefits

**User Perspective:**
- ✅ Single extension to install
- ✅ Unified status bar (both servers visible)
- ✅ Single command palette namespace
- ✅ Integrated sidebar panels
- ✅ Cross-server workflows (exarp-go → devwisdom-go)

**Example Unified Workflow:**
1. User clicks status bar → Shows combined health + wisdom
2. User runs "Daily Check-in" → Uses both servers
3. User views sidebar → Combined dashboard (tasks + wisdom)

### Separation Benefits

**User Perspective:**
- ✅ Install only what you need
- ✅ Independent updates
- ✅ Clearer separation of concerns
- ✅ Smaller extension sizes

---

## Implementation Strategy

### Phase 1: Shared Core Library (Recommended)

**Approach:** Create shared TypeScript library, use in both extensions

**Structure:**
```
mcp-extension-core/          # Shared library (npm package)
├── src/
│   ├── mcpClient.ts
│   ├── statusBar.ts
│   ├── webviewManager.ts
│   └── notificationManager.ts
└── package.json

exarp-go-extension/          # Uses mcp-extension-core
├── package.json (depends on mcp-extension-core)
└── src/ (exarp-go specific)

devwisdom-go-extension/      # Uses mcp-extension-core
├── package.json (depends on mcp-extension-core)
└── src/ (devwisdom-go specific)
```

**Benefits:**
- ✅ Code reuse without coupling
- ✅ Independent extensions
- ✅ Shared library can be published to npm
- ✅ Other projects can use shared library

**Drawbacks:**
- ⚠️ Need to publish/maintain npm package
- ⚠️ Version management between library and extensions

### Phase 2: Multi-Server Extension (Future)

**If Phase 1 succeeds, consider:**
- Merge into single extension
- Use adapter pattern
- Unified experience

---

## Recommendation Matrix

| Scenario | Recommendation | Rationale |
|----------|---------------|-----------|
| **MVP/Initial Development** | Separate Extensions | Simpler, faster to implement |
| **Long-term (if both succeed)** | Shared Core Library | Code reuse, maintainability |
| **If high user overlap** | Multi-Server Extension | Unified experience |
| **If different user bases** | Separate Extensions | Independent evolution |

---

## Decision Framework

### Choose Shared Extension If:
- ✅ Users typically use both servers together
- ✅ You want unified user experience
- ✅ You're willing to maintain adapter pattern
- ✅ You want single extension to install

### Choose Separate Extensions If:
- ✅ Users may only need one server
- ✅ You want simpler architecture
- ✅ You want independent development cycles
- ✅ You want smaller, focused extensions

### Choose Shared Core Library If:
- ✅ You want code reuse without coupling
- ✅ You want to publish reusable components
- ✅ You want flexibility (can merge later)
- ✅ You want other projects to benefit

---

## Recommended Approach

### **Option A: Shared Core Library** ⭐ **BEST BALANCE**

**Implementation:**
1. Create `mcp-extension-core` npm package
2. Extract shared infrastructure (MCP client, status bar, etc.)
3. Build `exarp-go-extension` using core
4. Build `devwisdom-go-extension` using core
5. Publish core to npm (optional, can be local)

**Benefits:**
- Code reuse (60-70% shared)
- Independent extensions (no coupling)
- Can merge later if desired
- Reusable for other MCP servers

**Timeline:**
- Core library: 8-10 hours
- exarp-go extension: 32-40 hours (reduced from 40-50)
- devwisdom-go extension: 16-20 hours (reduced from 24-34)
- **Total: 56-70 hours** (vs 64-84 hours separate)

### **Option B: Multi-Server Extension** (If High User Overlap)

**Implementation:**
1. Build unified extension framework
2. Create server adapters
3. Single extension with multi-server support

**Benefits:**
- Unified experience
- Single installation
- Cross-server workflows

**Timeline:**
- Framework: 12-16 hours
- Adapters: 20-24 hours
- **Total: 32-40 hours** (most efficient if both needed)

---

## Comparison Table

| Aspect | Separate | Shared Core | Multi-Server |
|--------|----------|-------------|--------------|
| **Code Reuse** | 0% | 60-70% | 60-70% |
| **Complexity** | Low | Medium | High |
| **Coupling** | None | Low | Medium |
| **User Experience** | Separate | Separate | Unified |
| **Maintenance** | Independent | Shared core | Combined |
| **Installation** | 2 extensions | 2 extensions | 1 extension |
| **Development Time** | 64-84h | 56-70h | 32-40h |
| **Flexibility** | High | High | Medium |

---

## Conclusion

**Best Approach: Shared Core Library**

1. **Start with shared core library** - Extract common infrastructure
2. **Build separate extensions** - Use core library, maintain independence
3. **Evaluate later** - If high user overlap, consider merging to multi-server extension

**Rationale:**
- Balances code reuse with flexibility
- Allows independent development
- Can evolve to unified extension if needed
- Reusable for other MCP servers

**Next Steps:**
1. Create `mcp-extension-core` project structure
2. Extract shared components (MCP client, status bar, etc.)
3. Build exarp-go extension using core
4. Build devwisdom-go extension using core
5. Publish core to npm (optional)

---

**Status**: ✅ Analysis Complete  
**Recommendation**: Shared Core Library (Option A)  
**Last Updated**: 2026-01-12
