# devwisdom-go MCP SDK Migration Plan

**Date:** 2026-01-09  
**Status:** 📋 Planning  
**Priority:** High  
**Estimated Effort:** 1-2 days

---

## Executive Summary

Migrate devwisdom-go from custom JSON-RPC 2.0 implementation to official `modelcontextprotocol/go-sdk` to:
- ✅ Unify MCP implementation approach with exarp-go
- ✅ Gain official SDK support and spec compliance
- ✅ Automatic updates with MCP specification changes
- ✅ Better long-term maintenance

---

## Current State Analysis

### devwisdom-go Implementation
- **Architecture:** Custom lightweight JSON-RPC 2.0
- **Files:**
  - `internal/mcp/protocol.go` (~149 lines) - JSON-RPC structures
  - `internal/mcp/server.go` (~885 lines) - Server implementation
- **Total:** ~1034 lines of custom code
- **Dependencies:** None (stdlib only)
- **Tools:** 4 tools (consult_advisor, get_wisdom, get_daily_briefing, get_consultation_log)
- **Resources:** 5 resources (tools, sources, advisors, advisor/{id}, consultations/{days})

### Target State
- **Architecture:** Official SDK + minimal adapter
- **SDK:** `github.com/modelcontextprotocol/go-sdk v1.2.0`
- **Adapter:** Simple adapter (similar to exarp-go but simpler)
- **Dependencies:** Official SDK
- **Benefits:** Official support, spec compliance, future-proof

---

## Migration Strategy

### Phase 1: Setup SDK Dependency ✅
**Duration:** 30 minutes

1. Add SDK to `go.mod`:
   ```bash
   cd /Users/davidl/Projects/devwisdom-go
   go get github.com/modelcontextprotocol/go-sdk@latest
   go mod tidy
   ```

2. Verify SDK version:
   ```bash
   go list -m github.com/modelcontextprotocol/go-sdk
   ```

**Acceptance Criteria:**
- ✅ SDK added to go.mod
- ✅ No dependency conflicts
- ✅ Build succeeds

---

### Phase 2: Create Minimal Adapter ✅
**Duration:** 2-3 hours

Create `internal/mcp/adapter.go` with minimal adapter pattern (simpler than exarp-go's abstraction):

```go
package mcp

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/davidl71/devwisdom-go/internal/wisdom"
    "github.com/davidl71/devwisdom-go/internal/logging"
)

type WisdomServer struct {
    server    *mcp.Server
    wisdom    *wisdom.Engine
    logger    *logging.ConsultationLogger
    appLogger *logging.Logger
}

func NewWisdomServer() *WisdomServer {
    // Initialize wisdom engine
    wisdomEngine := wisdom.NewEngine()
    
    // Initialize loggers
    consultationLogger, _ := logging.NewConsultationLogger(".devwisdom")
    appLogger := logging.NewLogger()
    
    // Create SDK server
    sdkServer := mcp.NewServer(&mcp.Implementation{
        Name:    "devwisdom",
        Version: "0.1.0",
    }, nil)
    
    return &WisdomServer{
        server:    sdkServer,
        wisdom:    wisdomEngine,
        logger:    consultationLogger,
        appLogger: appLogger,
    }
}

func (s *WisdomServer) Run(ctx context.Context) error {
    // Initialize wisdom engine
    if err := s.wisdom.Initialize(); err != nil {
        return fmt.Errorf("failed to initialize wisdom engine: %w", err)
    }
    
    // Register tools
    if err := s.registerTools(); err != nil {
        return fmt.Errorf("failed to register tools: %w", err)
    }
    
    // Register resources
    if err := s.registerResources(); err != nil {
        return fmt.Errorf("failed to register resources: %w", err)
    }
    
    // Run with stdio transport
    transport := &mcp.StdioTransport{}
    return s.server.Run(ctx, transport)
}
```

**Acceptance Criteria:**
- ✅ Adapter compiles
- ✅ Server can be created
- ✅ Run() method exists

---

### Phase 3: Migrate Tools ✅
**Duration:** 3-4 hours

Migrate each tool one by one:

#### 3.1: consult_advisor Tool
```go
func (s *WisdomServer) registerTools() error {
    // consult_advisor
    tool := &mcp.Tool{
        Name:        "consult_advisor",
        Description: "Consult a wisdom advisor based on metric, tool, or stage",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "metric": map[string]interface{}{
                    "type":        "string",
                    "description": "Metric name (e.g., 'security', 'testing')",
                },
                "tool": map[string]interface{}{
                    "type":        "string",
                    "description": "Tool name (e.g., 'project_scorecard')",
                },
                "stage": map[string]interface{}{
                    "type":        "string",
                    "description": "Stage name (e.g., 'daily_checkin')",
                },
                "score": map[string]interface{}{
                    "type":        "number",
                    "description": "Project health score (0-100)",
                },
                "context": map[string]interface{}{
                    "type":        "string",
                    "description": "Additional context for the consultation",
                },
            },
        },
    }
    
    handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // Extract parameters
        args := req.Params.Arguments
        
        // Call existing handleConsultAdvisor logic
        result, err := s.handleConsultAdvisor(args)
        if err != nil {
            return &mcp.CallToolResult{
                IsError: true,
                Content: []mcp.Content{
                    &mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
                },
            }, nil
        }
        
        // Convert result to SDK format
        return &mcp.CallToolResult{
            Content: []mcp.Content{
                &mcp.TextContent{Text: formatConsultation(result)},
            },
        }, nil
    }
    
    s.server.AddTool(tool, handler)
    return nil
}
```

#### 3.2: get_wisdom Tool
Similar pattern - migrate existing `handleGetWisdom` logic.

#### 3.3: get_daily_briefing Tool
Similar pattern - migrate existing `handleGetDailyBriefing` logic.

#### 3.4: get_consultation_log Tool
Similar pattern - migrate existing `handleGetConsultationLog` logic.

**Acceptance Criteria:**
- ✅ All 4 tools registered
- ✅ Tools compile
- ✅ Handler signatures match SDK API

---

### Phase 4: Migrate Resources ✅
**Duration:** 2-3 hours

Migrate resources using SDK resource API:

```go
func (s *WisdomServer) registerResources() error {
    // wisdom://tools
    resource := &mcp.Resource{
        URI:         "wisdom://tools",
        Name:        "Available Tools",
        Description: "List all available MCP tools with descriptions and parameters",
        MIMEType:    "application/json",
    }
    
    handler := func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
        // Use existing handleToolsResource logic
        tools := s.getToolsList()
        return &mcp.ReadResourceResult{
            Contents: []*mcp.ResourceContents{
                {
                    URI:      "wisdom://tools",
                    MIMEType: "application/json",
                    Text:     formatToolsJSON(tools),
                },
            },
        }, nil
    }
    
    s.server.AddResource(resource, handler)
    
    // Repeat for other resources...
    return nil
}
```

**Acceptance Criteria:**
- ✅ All 5 resources registered
- ✅ Resources compile
- ✅ Handler signatures match SDK API

---

### Phase 5: Update Main Entry Point ✅
**Duration:** 30 minutes

Update `cmd/server/main.go`:

```go
package main

import (
    "context"
    "log"
    "github.com/davidl71/devwisdom-go/internal/mcp"
)

func main() {
    server := mcp.NewWisdomServer()
    
    if err := server.Run(context.Background()); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

**Acceptance Criteria:**
- ✅ Main compiles
- ✅ Server starts successfully

---

### Phase 6: Testing ✅
**Duration:** 2-3 hours

1. **Unit Tests:**
   - Test each tool handler
   - Test each resource handler
   - Test error cases

2. **Integration Tests:**
   - Test full server startup
   - Test tool calls via MCP client
   - Test resource reads via MCP client

3. **Manual Testing:**
   - Test with Cursor IDE
   - Verify all tools work
   - Verify all resources work
   - Check logging output

**Acceptance Criteria:**
- ✅ All unit tests pass
- ✅ Integration tests pass
- ✅ Manual testing successful
- ✅ No regressions

---

### Phase 7: Cleanup ✅
**Duration:** 1 hour

1. Remove old custom implementation:
   - Delete `internal/mcp/protocol.go` (or keep for reference initially)
   - Delete `internal/mcp/server.go` (or keep for reference initially)

2. Update documentation:
   - Update README.md
   - Update MCP_IMPLEMENTATION_SUMMARY.md

3. Update version:
   - Bump version to 0.2.0 (breaking change)

**Acceptance Criteria:**
- ✅ Old code removed
- ✅ Documentation updated
- ✅ Version bumped

---

## Rollback Plan

If migration fails:

1. **Keep old implementation:** Don't delete until migration is verified
2. **Git branch:** Work in feature branch (`mcp-sdk-migration`)
3. **Quick rollback:** Revert to previous commit if needed
4. **Gradual migration:** Can migrate one tool at a time

---

## Risk Mitigation

### Risks:
1. **Breaking Changes:** SDK API might differ from custom implementation
2. **Performance:** SDK might have different performance characteristics
3. **Dependencies:** Adding SDK dependency increases binary size

### Mitigation:
1. **Thorough Testing:** Test each phase before proceeding
2. **Feature Branch:** Work in isolation
3. **Keep Old Code:** Don't delete until verified
4. **Performance Testing:** Benchmark if needed

---

## Success Criteria

- ✅ All 4 tools work correctly
- ✅ All 5 resources work correctly
- ✅ Server starts and runs successfully
- ✅ No regressions in functionality
- ✅ Tests pass
- ✅ Documentation updated
- ✅ Old code removed

---

## Timeline Estimate

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 1: Setup SDK | 30 min | 30 min |
| Phase 2: Create Adapter | 2-3 hours | 3-3.5 hours |
| Phase 3: Migrate Tools | 3-4 hours | 6-7.5 hours |
| Phase 4: Migrate Resources | 2-3 hours | 8-10.5 hours |
| Phase 5: Update Main | 30 min | 9-11 hours |
| Phase 6: Testing | 2-3 hours | 11-14 hours |
| Phase 7: Cleanup | 1 hour | 12-15 hours |

**Total:** 1.5-2 days (12-15 hours)

---

## Next Steps

1. ✅ Review this migration plan
2. ⏳ Create feature branch: `git checkout -b mcp-sdk-migration`
3. ⏳ Start Phase 1: Add SDK dependency
4. ⏳ Proceed through phases sequentially
5. ⏳ Test thoroughly at each phase
6. ⏳ Merge when complete

---

## References

- [Official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)
- [exarp-go SDK Adapter](../exarp-go/internal/framework/adapters/gosdk/adapter.go)
- [MCP Specification](https://modelcontextprotocol.io/specification)

