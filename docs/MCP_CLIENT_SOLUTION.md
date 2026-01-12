# MCP Client Dependencies - Solution Analysis

**Date:** 2026-01-13  
**Problem:** 3 actions need to call devwisdom-go MCP server (recommend/advisor, report/briefing, memory_maint/dream)  
**Question:** How do we fix MCP client dependencies? Can context7 help?

---

## Current Situation

### Tools Requiring MCP Client
1. **`recommend` / advisor** - Calls devwisdom-go `consult_advisor` tool
2. **`report` / briefing** - Calls devwisdom-go `get_daily_briefing` tool  
3. **`memory_maint` / dream** - Calls devwisdom-go `consult_advisor` tool

### Current Implementation
- **Python Bridge**: Uses Python's MCP client capabilities to call devwisdom-go
- **Reason**: Python has easier MCP client libraries
- **Problem**: Requires Python bridge for these 3 actions

---

## Solution Options

### Option 1: Implement Go MCP Client (Recommended)

**Approach:** Use Go SDK's client capabilities (if available) or implement MCP client in Go

**Status Check Needed:**
- Does `github.com/modelcontextprotocol/go-sdk` support client functionality?
- Are there Go MCP client libraries available?
- What's the pattern for calling other MCP servers from Go?

**Implementation Steps:**
1. Check Go SDK for client support
2. If not available, research Go MCP client libraries
3. Implement client wrapper for devwisdom-go
4. Replace Python bridge calls with native Go client calls

**Pros:**
- ✅ Fully native Go (no Python bridge)
- ✅ Better performance
- ✅ Single binary deployment
- ✅ Consistent with migration goals

**Cons:**
- ⚠️ May require additional library or custom implementation
- ⚠️ More complex than Python bridge
- ⚠️ Need to maintain MCP client code

---

### Option 2: Context7 Analysis

**Question:** Can context7 help?

**Analysis:**
- ❌ **Context7 is an MCP SERVER**, not a client library
- ✅ Context7 provides documentation lookup (not wisdom/advisory features)
- ✅ Context7 is already integrated via `add_external_tool_hints` tool
- ❌ Context7 cannot call devwisdom-go MCP server

**Conclusion:** Context7 cannot help with MCP client dependencies. It's a documentation server, not a client library.

---

### Option 3: Keep Python Bridge (Current)

**Approach:** Keep Python bridge for MCP client dependencies

**Pros:**
- ✅ Works now (no implementation needed)
- ✅ Python has mature MCP client libraries
- ✅ Low maintenance (already working)

**Cons:**
- ❌ Requires Python bridge (not fully native Go)
- ❌ Adds Python dependency
- ❌ Inconsistent with migration goals

---

## Recommended Solution

### Short-term: Keep Python Bridge
- ✅ Works well for MCP client dependencies
- ✅ Low priority (only 3 actions)
- ✅ No blocking issues

### Long-term: Implement Go MCP Client
**Steps:**
1. **Research Go MCP Client Libraries**
   - Check if `github.com/modelcontextprotocol/go-sdk` has client support
   - Research other Go MCP client libraries
   - Evaluate implementation complexity

2. **If Go SDK Supports Clients:**
   - Use official SDK client functionality
   - Implement client wrapper for devwisdom-go
   - Replace Python bridge calls

3. **If No Go Client Library:**
   - Implement basic MCP client using JSON-RPC 2.0 over stdio
   - Use subprocess execution for devwisdom-go binary
   - Implement tool call protocol

4. **Implementation Pattern:**
   ```go
   // Pseudo-code for Go MCP client
   type MCPClient struct {
       command string
       args    []string
   }

   func (c *MCPClient) CallTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
       // Start devwisdom-go process
       // Send initialize request
       // Send tools/call request
       // Parse response
       // Return result
   }
   ```

---

## Next Steps

1. **Research Phase:**
   - Check Go SDK documentation for client support
   - Search for Go MCP client libraries
   - Review devwisdom-go MCP protocol

2. **Prototype Phase:**
   - Implement basic MCP client in Go
   - Test with devwisdom-go server
   - Verify tool calls work

3. **Integration Phase:**
   - Replace Python bridge calls with Go client
   - Update handlers for recommend, report, memory_maint
   - Test all 3 actions

4. **Documentation Phase:**
   - Document MCP client implementation
   - Update migration status
   - Create usage guide

---

## Context7 Clarification

**Context7 Role:**
- ✅ Provides documentation lookup (researcher tool)
- ✅ Already integrated via `add_external_tool_hints`
- ❌ Cannot help with MCP client dependencies
- ❌ Not a client library for calling other MCP servers

**Context7 Use Cases:**
- Documentation retrieval (already working)
- Library API reference lookup
- Code examples and patterns
- **NOT** for calling other MCP servers

---

## Conclusion

**For MCP Client Dependencies:**
1. **Short-term**: Keep Python bridge (works well, low priority)
2. **Long-term**: Implement Go MCP client (fully native Go)
3. **Context7**: Cannot help (it's a server, not a client library)

**Priority:** Low (only 3 actions, current solution works)
**Complexity:** Medium (requires MCP client implementation)
**Value:** Medium (completes native Go migration)
