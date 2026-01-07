# Migration Plan: Python MCP Server → Go SDK

**Date:** 2025-01-01  
**Status:** 📋 Planning  
**Target:** Official Go SDK (`github.com/modelcontextprotocol/go-sdk`)

---

## Executive Summary

This document outlines the migration plan from the current Python-based MCP stdio server to the official Go SDK. The migration will improve performance, reduce dependencies, and provide a single-binary deployment while maintaining full compatibility with Cursor IDE.

**Key Benefits:**
- ✅ Single binary deployment (no Python dependencies)
- ✅ Faster startup and execution
- ✅ Lower memory footprint
- ✅ Official SDK support and updates
- ✅ Better performance for concurrent requests
- ✅ Easier distribution and deployment

**Challenges:**
- ⚠️ 24 tools currently implemented in Python
- ⚠️ MLX integration (Apple Silicon specific)
- ⚠️ Complex tool logic needs porting or bridging
- ⚠️ Resource handlers in Python

---

## Current State Analysis

### Current Implementation

**Server Details:**
- **Language:** Python 3.10+
- **Framework:** `mcp` Python SDK
- **Transport:** STDIO
- **Server Name:** `exarp-go`
- **Location:** `/Users/davidl/Projects/mcp-stdio-tools`

**Components:**
- **Tools:** 24 tools (all call Python functions from `project_management_automation`)
- **Prompts:** 8 prompts (imported from Python modules)
- **Resources:** 6 resources (handled by Python functions)

**Dependencies:**
- `mcp>=1.0.0` (Python MCP SDK)
- `mlx>=0.20.0` (Apple Silicon ML acceleration)
- `mlx-lm>=0.20.0` (MLX language models)
- Python project: `project_management_automation` (external dependency)

**Tool Categories:**

1. **Project Management (12 tools):**
   - `analyze_alignment`, `generate_config`, `health`, `memory`, `memory_maint`
   - `report`, `security`, `setup_hooks`, `task_analysis`, `task_discovery`
   - `task_workflow`, `testing`

2. **Advanced Tools (12 tools):**
   - `infer_session_mode`, `add_external_tool_hints`, `automation`
   - `tool_catalog`, `workflow_mode`, `check_attribution`, `lint`
   - `estimation`, `ollama`, `mlx`, `git_tools`, `session`

**Integration Points:**
- All tools import from `project_management_automation.tools.consolidated`
- Prompts import from `project_management_automation.prompts`
- Resources import from `project_management_automation.resources.memories`

---

## Target State: Go SDK

### Go SDK Overview

**Repository:** `github.com/modelcontextprotocol/go-sdk`  
**Official Status:** ✅ Maintained by Model Context Protocol organization  
**Latest Activity:** 4 days ago (as of December 2025)  
**Protocol Support:** MCP 2025-06-18+  
**Transport:** STDIO, HTTP, SSE

**Key Features:**
- Native Go implementation
- Type-safe APIs
- High performance
- Official specification compliance
- Comprehensive testing utilities
- Production-ready logging

### Go SDK Structure

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "exarp-go",
        Version: "1.0.0",
    }, nil)
    
    // Register tools, prompts, resources
    
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatal(err)
    }
}
```

---

## Migration Strategy

### Approach: Hybrid Bridge Architecture

Given the complexity of porting 24 tools from Python to Go, we'll use a **hybrid approach**:

1. **Go SDK Server** - Main MCP server in Go
2. **Python Bridge** - Execute Python tools via subprocess/exec
3. **Gradual Porting** - Migrate tools to Go incrementally

**Benefits:**
- ✅ Fast migration (reuse existing Python code)
- ✅ Incremental porting (migrate tools one by one)
- ✅ No breaking changes during migration
- ✅ Can test Go server with Python tools immediately

**Architecture:**

```
┌─────────────────┐
│   Cursor IDE    │
└────────┬────────┘
         │ STDIO
         ▼
┌─────────────────┐
│  Go SDK Server  │  ← New Go implementation
│  (MCP Protocol) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Python Bridge  │  ← Execute Python tools
│  (Subprocess)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Python Tools    │  ← Existing Python code
│ (project_...)   │
└─────────────────┘
```

### Migration Phases

#### Phase 1: Foundation (Week 1)
- [ ] Set up Go project structure
- [ ] Install and configure Go SDK
- [ ] Create basic STDIO server skeleton
- [ ] Implement Python bridge mechanism
- [ ] Test basic tool execution via bridge

#### Phase 2: Tool Migration - Batch 1 (Week 2)
- [ ] Migrate simple tools (no MLX dependency)
- [ ] Implement tool registration system
- [ ] Test tool execution
- [ ] Verify Cursor integration

**Batch 1 Tools (Simple):**
- `analyze_alignment`
- `generate_config`
- `health`
- `setup_hooks`
- `check_attribution`
- `add_external_tool_hints`

#### Phase 3: Tool Migration - Batch 2 (Week 3)
- [ ] Migrate medium-complexity tools
- [ ] Implement prompt system
- [ ] Test prompts in Cursor

**Batch 2 Tools (Medium):**
- `memory`
- `memory_maint`
- `report`
- `security`
- `task_analysis`
- `task_discovery`
- `task_workflow`
- `testing`

#### Phase 4: Tool Migration - Batch 3 (Week 4)
- [ ] Migrate advanced tools
- [ ] Implement resource handlers
- [ ] Test resources in Cursor

**Batch 3 Tools (Advanced):**
- `automation`
- `tool_catalog`
- `workflow_mode`
- `lint`
- `estimation`
- `git_tools`
- `session`
- `infer_session_mode`

#### Phase 5: MLX Integration (Week 5)
- [ ] Handle MLX tools (`ollama`, `mlx`)
- [ ] Implement MLX bridge or alternative
- [ ] Test MLX tools

**MLX Tools:**
- `ollama` - May work via HTTP API
- `mlx` - Requires Apple Silicon, may need Python bridge

#### Phase 6: Testing & Optimization (Week 6)
- [ ] Comprehensive testing
- [ ] Performance optimization
- [ ] Documentation
- [ ] Deployment preparation

---

## Detailed Migration Plan

### Phase 1: Foundation Setup

#### 1.1 Go Project Structure

```
exarp-go/
├── cmd/
│   └── server/
│       └── main.go          # Main server entry point
├── internal/
│   ├── bridge/
│   │   └── python.go       # Python execution bridge
│   ├── tools/
│   │   ├── registry.go      # Tool registration
│   │   └── handlers.go      # Tool handlers
│   ├── prompts/
│   │   └── registry.go     # Prompt registration
│   └── resources/
│       └── handlers.go     # Resource handlers
├── pkg/
│   └── config/
│       └── config.go       # Configuration management
├── go.mod
├── go.sum
└── README.md
```

#### 1.2 Go Module Setup

```bash
# Initialize Go module
cd /Users/davidl/Projects/exarp-go
go mod init github.com/davidl/exarp-go

# Add Go SDK dependency
go get github.com/modelcontextprotocol/go-sdk

# Add other dependencies
go get github.com/spf13/cobra  # CLI (optional)
go get github.com/spf13/viper  # Config (optional)
```

#### 1.3 Basic Server Skeleton

```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    "os"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "exarp-go",
        Version: "1.0.0",
    }, nil)

    // Register handlers
    registerTools(server)
    registerPrompts(server)
    registerResources(server)

    // Run with STDIO transport
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatal(err)
    }
}
```

#### 1.4 Python Bridge Implementation

```go
// internal/bridge/python.go
package bridge

import (
    "context"
    "encoding/json"
    "os/exec"
    "path/filepath"
)

type PythonBridge struct {
    pythonPath    string
    scriptPath    string
    projectRoot   string
}

func NewPythonBridge(projectRoot string) *PythonBridge {
    return &PythonBridge{
        pythonPath:  "python3",
        scriptPath:  filepath.Join(projectRoot, "bridge", "execute_tool.py"),
        projectRoot: projectRoot,
    }
}

func (b *PythonBridge) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
    // Serialize arguments
    argsJSON, err := json.Marshal(args)
    if err != nil {
        return "", err
    }

    // Execute Python tool
    cmd := exec.CommandContext(ctx, b.pythonPath, b.scriptPath, toolName, string(argsJSON))
    cmd.Env = append(os.Environ(), "PROJECT_ROOT="+b.projectRoot)
    
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    return string(output), nil
}
```

**Python Bridge Script:**

```python
# bridge/execute_tool.py
#!/usr/bin/env python3
"""Bridge script to execute Python tools from Go."""

import sys
import json
import os
from pathlib import Path

# Add project to path
PROJECT_ROOT = Path(os.getenv("PROJECT_ROOT", "."))
sys.path.insert(0, str(PROJECT_ROOT / "project-management-automation"))

def main():
    if len(sys.argv) < 3:
        print(json.dumps({"error": "Missing arguments"}))
        sys.exit(1)
    
    tool_name = sys.argv[1]
    args_json = sys.argv[2]
    args = json.loads(args_json)
    
    # Import and execute tool
    from project_management_automation.tools.consolidated import *
    
    tool_func = globals().get(tool_name)
    if not tool_func:
        print(json.dumps({"error": f"Tool {tool_name} not found"}))
        sys.exit(1)
    
    try:
        result = tool_func(**args)
        if isinstance(result, str):
            print(result)
        else:
            print(json.dumps(result, indent=2))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)

if __name__ == "__main__":
    main()
```

---

### Phase 2-4: Tool Migration

#### Tool Registration Pattern

```go
// internal/tools/registry.go
package tools

import (
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/davidl/exarp-go/internal/bridge"
)

type ToolRegistry struct {
    bridge *bridge.PythonBridge
    tools  []mcp.Tool
}

func NewToolRegistry(bridge *bridge.PythonBridge) *ToolRegistry {
    return &ToolRegistry{
        bridge: bridge,
        tools:  []mcp.Tool{},
    }
}

func (r *ToolRegistry) RegisterTool(name string, description string, schema mcp.ToolSchema) {
    r.tools = append(r.tools, mcp.Tool{
        Name:        name,
        Description: description,
        InputSchema: schema,
    })
}

func (r *ToolRegistry) GetTools() []mcp.Tool {
    return r.tools
}
```

#### Tool Handler Pattern

```go
// internal/tools/handlers.go
package tools

import (
    "context"
    "encoding/json"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/davidl/exarp-go/internal/bridge"
)

func HandleTool(ctx context.Context, name string, args json.RawMessage, bridge *bridge.PythonBridge) ([]mcp.TextContent, error) {
    // Parse arguments
    var argsMap map[string]interface{}
    if err := json.Unmarshal(args, &argsMap); err != nil {
        return nil, err
    }
    
    // Execute via Python bridge
    result, err := bridge.ExecuteTool(ctx, name, argsMap)
    if err != nil {
        return nil, err
    }
    
    return []mcp.TextContent{
        {Type: "text", Text: result},
    }, nil
}
```

#### Example: Simple Tool Migration

**Before (Python):**
```python
@server.call_tool()
async def call_tool(name: str, arguments: dict[str, Any]) -> list[TextContent]:
    if name == "analyze_alignment":
        result = _analyze_alignment(
            action=arguments.get("action", "todo2"),
            create_followup_tasks=arguments.get("create_followup_tasks", True),
            output_path=arguments.get("output_path"),
        )
        return [TextContent(type="text", text=result)]
```

**After (Go with Bridge):**
```go
func registerAnalyzeAlignment(server *mcp.Server, bridge *bridge.PythonBridge) {
    server.RegisterTool("analyze_alignment", func(ctx context.Context, args json.RawMessage) ([]mcp.TextContent, error) {
        return HandleTool(ctx, "analyze_alignment", args, bridge)
    })
}
```

---

### Phase 5: MLX Integration

#### Option 1: Python Bridge (Recommended Initially)

Keep MLX tools in Python and execute via bridge:

```go
// Special handling for MLX tools
func HandleMLXTools(ctx context.Context, name string, args json.RawMessage, bridge *bridge.PythonBridge) ([]mcp.TextContent, error) {
    // MLX requires Python runtime, use bridge
    return HandleTool(ctx, name, args, bridge)
}
```

#### Option 2: HTTP API (For Ollama)

If Ollama has HTTP API, call directly from Go:

```go
func HandleOllama(ctx context.Context, args json.RawMessage) ([]mcp.TextContent, error) {
    // Call Ollama HTTP API directly
    // No Python bridge needed
}
```

#### Option 3: Native Go MLX (Future)

If Go MLX bindings become available, migrate natively.

---

### Phase 6: Prompts & Resources

#### Prompt Registration

```go
// internal/prompts/registry.go
package prompts

import "github.com/modelcontextprotocol/go-sdk/mcp"

func RegisterPrompts(server *mcp.Server) {
    prompts := map[string]string{
        "align":      loadPrompt("TASK_ALIGNMENT_ANALYSIS"),
        "discover":   loadPrompt("TASK_DISCOVERY"),
        "config":     loadPrompt("CONFIG_GENERATION"),
        // ... etc
    }
    
    for name, content := range prompts {
        server.RegisterPrompt(name, func(ctx context.Context, args map[string]interface{}) (string, error) {
            return content, nil
        })
    }
}

func loadPrompt(name string) string {
    // Load from Python module or Go file
    // Could use Python bridge or port to Go
}
```

#### Resource Handlers

```go
// internal/resources/handlers.go
package resources

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/davidl/exarp-go/internal/bridge"
)

func RegisterResources(server *mcp.Server, bridge *bridge.PythonBridge) {
    server.RegisterResource("stdio://scorecard", func(ctx context.Context, uri string) (string, error) {
        return bridge.ExecuteTool(ctx, "generate_project_scorecard", map[string]interface{}{
            "format": "json",
        })
    })
    
    // ... other resources
}
```

---

## Testing Strategy

### Unit Tests

```go
// internal/tools/handlers_test.go
package tools

import (
    "context"
    "testing"
)

func TestHandleTool(t *testing.T) {
    bridge := bridge.NewPythonBridge("/path/to/project")
    ctx := context.Background()
    
    args := json.RawMessage(`{"action": "todo2"}`)
    result, err := HandleTool(ctx, "analyze_alignment", args, bridge)
    
    if err != nil {
        t.Fatalf("Tool execution failed: %v", err)
    }
    
    if len(result) == 0 {
        t.Fatal("No result returned")
    }
}
```

### Integration Tests

```go
// Test full MCP server
func TestMCPServer(t *testing.T) {
    // Start server
    // Send MCP requests
    // Verify responses
}
```

### Cursor Integration Tests

1. Update `.cursor/mcp.json` to point to Go binary
2. Restart Cursor
3. Test each tool via Cursor interface
4. Verify prompts work
5. Verify resources accessible

---

## Deployment Plan

### Build Process

```bash
# Build single binary
go build -o bin/exarp-go cmd/server/main.go

# Cross-compile (if needed)
GOOS=darwin GOARCH=arm64 go build -o bin/exarp-go-darwin-arm64 cmd/server/main.go
```

### Update Cursor Configuration

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "{{PROJECT_ROOT}}/../exarp-go/bin/exarp-go",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      }
    }
  }
}
```

### Dependencies

**Go Binary:**
- Self-contained (no runtime dependencies)
- Requires Python 3.10+ for bridge (temporary)
- Requires `project-management-automation` Python package

**Future (After Full Migration):**
- Self-contained binary
- No Python dependencies

---

## Migration Checklist

### Phase 1: Foundation
- [ ] Create Go project structure
- [ ] Set up Go module and dependencies
- [ ] Implement basic STDIO server
- [ ] Create Python bridge mechanism
- [ ] Test bridge with one simple tool

### Phase 2: Batch 1 Tools
- [ ] Migrate `analyze_alignment`
- [ ] Migrate `generate_config`
- [ ] Migrate `health`
- [ ] Migrate `setup_hooks`
- [ ] Migrate `check_attribution`
- [ ] Migrate `add_external_tool_hints`
- [ ] Test all Batch 1 tools in Cursor

### Phase 3: Batch 2 Tools
- [ ] Migrate `memory`
- [ ] Migrate `memory_maint`
- [ ] Migrate `report`
- [ ] Migrate `security`
- [ ] Migrate `task_analysis`
- [ ] Migrate `task_discovery`
- [ ] Migrate `task_workflow`
- [ ] Migrate `testing`
- [ ] Implement prompt system
- [ ] Test prompts in Cursor

### Phase 4: Batch 3 Tools
- [ ] Migrate `automation`
- [ ] Migrate `tool_catalog`
- [ ] Migrate `workflow_mode`
- [ ] Migrate `lint`
- [ ] Migrate `estimation`
- [ ] Migrate `git_tools`
- [ ] Migrate `session`
- [ ] Migrate `infer_session_mode`
- [ ] Implement resource handlers
- [ ] Test resources in Cursor

### Phase 5: MLX Tools
- [ ] Handle `ollama` (via HTTP or bridge)
- [ ] Handle `mlx` (via bridge)
- [ ] Test MLX tools

### Phase 6: Finalization
- [ ] Comprehensive testing
- [ ] Performance optimization
- [ ] Documentation
- [ ] Update README
- [ ] Create migration guide
- [ ] Deploy to production

---

## Risk Assessment

### High Risk
- **MLX Integration:** Apple Silicon specific, may require Python bridge permanently
- **Complex Tool Logic:** Some tools have complex Python dependencies

### Medium Risk
- **Python Bridge Performance:** Subprocess overhead
- **Error Handling:** Bridge error propagation
- **Testing:** Need comprehensive test coverage

### Low Risk
- **Go SDK Maturity:** Official SDK, well-maintained
- **STDIO Transport:** Standard, well-supported
- **Cursor Integration:** Same protocol, should work seamlessly

---

## Success Criteria

### Functional Requirements
- ✅ All 24 tools work via Go server
- ✅ All 8 prompts accessible
- ✅ All 6 resources accessible
- ✅ Cursor integration works seamlessly
- ✅ No regression in functionality

### Performance Requirements
- ✅ Server startup < 100ms
- ✅ Tool execution overhead < 50ms (bridge overhead)
- ✅ Memory usage < 50MB baseline

### Quality Requirements
- ✅ Comprehensive test coverage (>80%)
- ✅ Error handling for all edge cases
- ✅ Logging and monitoring
- ✅ Documentation complete

---

## Timeline Estimate

**Total Duration:** 6 weeks

- **Week 1:** Foundation setup
- **Week 2:** Batch 1 tools (6 tools)
- **Week 3:** Batch 2 tools (8 tools)
- **Week 4:** Batch 3 tools (8 tools)
- **Week 5:** MLX integration (2 tools)
- **Week 6:** Testing, optimization, documentation

**Contingency:** +2 weeks for unexpected issues

---

## Next Steps

1. **Review and approve migration plan**
2. **Set up Go development environment**
3. **Create Go project structure**
4. **Begin Phase 1: Foundation setup**

---

## References

- [Go SDK Repository](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://modelcontextprotocol.io)
- [Cursor MCP Documentation](https://docs.cursor.com/context/model-context-protocol)
- [Go Documentation](https://go.dev/doc/)

---

**Status:** Ready for review and approval  
**Next Action:** Begin Phase 1 implementation

