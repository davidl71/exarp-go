# Framework-Agnostic MCP Server Design

**Goal:** Design code that can easily switch between different Go MCP frameworks (go-sdk, mcp-go, go-mcp, etc.)

**Difficulty:** 🟢 **Easy to Moderate** - With proper abstraction, framework switching becomes a configuration change.

---

## Design Approach: Interface-Based Abstraction

### Core Principle

**Abstract the MCP framework behind Go interfaces.** Your application code only knows about your interfaces, not the specific framework implementation.

```
┌─────────────────────────────────────┐
│     Application Code                │
│  (Tools, Prompts, Resources)        │
└──────────────┬──────────────────────┘
               │ Uses
               ▼
┌─────────────────────────────────────┐
│     Framework Abstraction Layer     │
│  (Interfaces + Adapters)            │
└──────────────┬──────────────────────┘
               │ Implements
               ▼
┌─────────────────────────────────────┐
│     Framework Implementations       │
│  go-sdk | mcp-go | go-mcp           │
└─────────────────────────────────────┘
```

---

## Core Interfaces

### 1. MCPServer Interface

```go
// internal/framework/server.go
package framework

import (
    "context"
    "encoding/json"
)

// MCPServer abstracts MCP server functionality
type MCPServer interface {
    // RegisterTool registers a tool handler
    RegisterTool(name, description string, schema ToolSchema, handler ToolHandler) error
    
    // RegisterPrompt registers a prompt template
    RegisterPrompt(name, description string, handler PromptHandler) error
    
    // RegisterResource registers a resource handler
    RegisterResource(uri, name, description, mimeType string, handler ResourceHandler) error
    
    // Run starts the server with the given transport
    Run(ctx context.Context, transport Transport) error
    
    // GetName returns the server name
    GetName() string
}

// ToolHandler handles tool execution
type ToolHandler func(ctx context.Context, args json.RawMessage) ([]TextContent, error)

// PromptHandler handles prompt requests
type PromptHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// ResourceHandler handles resource requests
type ResourceHandler func(ctx context.Context, uri string) (string, error)

// TextContent represents MCP text content
type TextContent struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

// ToolSchema represents tool input schema
type ToolSchema struct {
    Type       string                 `json:"type"`
    Properties map[string]interface{} `json:"properties"`
    Required   []string               `json:"required,omitempty"`
}

// Transport abstracts transport mechanism
type Transport interface {
    // Transport-specific methods
    // Each framework will implement this differently
}
```

---

## Framework Adapters

### 2. Go SDK Adapter

```go
// internal/framework/adapters/gosdk/adapter.go
package gosdk

import (
    "context"
    "encoding/json"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/davidl/exarp-go/internal/framework"
)

type GoSDKAdapter struct {
    server *mcp.Server
    name   string
}

func NewGoSDKAdapter(name, version string) *GoSDKAdapter {
    return &GoSDKAdapter{
        server: mcp.NewServer(&mcp.Implementation{
            Name:    name,
            Version: version,
        }, nil),
        name: name,
    }
}

func (a *GoSDKAdapter) RegisterTool(name, description string, schema framework.ToolSchema, handler framework.ToolHandler) error {
    a.server.RegisterTool(name, func(ctx context.Context, args json.RawMessage) ([]mcp.TextContent, error) {
        // Convert framework types to go-sdk types
        result, err := handler(ctx, args)
        if err != nil {
            return nil, err
        }
        
        // Convert to go-sdk TextContent
        mcpContents := make([]mcp.TextContent, len(result))
        for i, content := range result {
            mcpContents[i] = mcp.TextContent{
                Type: content.Type,
                Text: content.Text,
            }
        }
        return mcpContents, nil
    })
    return nil
}

func (a *GoSDKAdapter) RegisterPrompt(name, description string, handler framework.PromptHandler) error {
    a.server.RegisterPrompt(name, func(ctx context.Context, args map[string]interface{}) (string, error) {
        return handler(ctx, args)
    })
    return nil
}

func (a *GoSDKAdapter) RegisterResource(uri, name, description, mimeType string, handler framework.ResourceHandler) error {
    a.server.RegisterResource(uri, func(ctx context.Context, uri string) (string, error) {
        return handler(ctx, uri)
    })
    return nil
}

func (a *GoSDKAdapter) Run(ctx context.Context, transport framework.Transport) error {
    // Convert framework transport to go-sdk transport
    stdioTransport := &mcp.StdioTransport{}
    return a.server.Run(ctx, stdioTransport)
}

func (a *GoSDKAdapter) GetName() string {
    return a.name
}
```

### 3. MCP-Go Adapter (mark3labs)

```go
// internal/framework/adapters/mcpgo/adapter.go
package mcpgo

import (
    "context"
    "encoding/json"
    
    "github.com/mark3labs/mcp-go/server"
    "github.com/davidl/exarp-go/internal/framework"
)

type MCPGoAdapter struct {
    server *server.Server
    name   string
}

func NewMCPGoAdapter(name string) *MCPGoAdapter {
    return &MCPGoAdapter{
        server: server.New(name),
        name:   name,
    }
}

func (a *MCPGoAdapter) RegisterTool(name, description string, schema framework.ToolSchema, handler framework.ToolHandler) error {
    // Adapt to mcp-go's API
    a.server.HandleTool(name, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
        // Convert args to json.RawMessage
        argsJSON, _ := json.Marshal(args)
        
        // Call framework handler
        result, err := handler(ctx, json.RawMessage(argsJSON))
        if err != nil {
            return nil, err
        }
        
        // Convert result to mcp-go format
        if len(result) > 0 {
            return result[0].Text, nil
        }
        return "", nil
    })
    return nil
}

func (a *MCPGoAdapter) RegisterPrompt(name, description string, handler framework.PromptHandler) error {
    // Adapt to mcp-go's prompt API
    a.server.HandlePrompt(name, func(ctx context.Context, args map[string]interface{}) (string, error) {
        return handler(ctx, args)
    })
    return nil
}

func (a *MCPGoAdapter) RegisterResource(uri, name, description, mimeType string, handler framework.ResourceHandler) error {
    // Adapt to mcp-go's resource API
    a.server.HandleResource(uri, func(ctx context.Context, uri string) (string, error) {
        return handler(ctx, uri)
    })
    return nil
}

func (a *MCPGoAdapter) Run(ctx context.Context, transport framework.Transport) error {
    // Adapt to mcp-go's transport
    return a.server.Run(ctx)
}

func (a *MCPGoAdapter) GetName() string {
    return a.name
}
```

### 4. Go-MCP Adapter (ThinkInAIXYZ)

```go
// internal/framework/adapters/gomcp/adapter.go
package gomcp

import (
    "context"
    "encoding/json"
    
    "github.com/ThinkInAIXYZ/go-mcp/sdk"
    "github.com/davidl/exarp-go/internal/framework"
)

type GoMCPAdapter struct {
    server *sdk.Server
    name   string
}

func NewGoMCPAdapter(name string) *GoMCPAdapter {
    return &GoMCPAdapter{
        server: sdk.NewServer(name),
        name:   name,
    }
}

// Similar adapter pattern for go-mcp
// Implementation details depend on go-mcp's actual API
```

---

## Factory Pattern for Framework Selection

### 5. Framework Factory

```go
// internal/framework/factory.go
package framework

import (
    "github.com/davidl/exarp-go/internal/framework/adapters/gosdk"
    "github.com/davidl/exarp-go/internal/framework/adapters/mcpgo"
    "github.com/davidl/exarp-go/internal/framework/adapters/gomcp"
)

type FrameworkType string

const (
    FrameworkGoSDK  FrameworkType = "go-sdk"
    FrameworkMCPGo  FrameworkType = "mcp-go"
    FrameworkGoMCP  FrameworkType = "go-mcp"
)

// NewServer creates a new MCP server using the specified framework
func NewServer(frameworkType FrameworkType, name, version string) (MCPServer, error) {
    switch frameworkType {
    case FrameworkGoSDK:
        return gosdk.NewGoSDKAdapter(name, version), nil
    case FrameworkMCPGo:
        return mcpgo.NewMCPGoAdapter(name), nil
    case FrameworkGoMCP:
        return gomcp.NewGoMCPAdapter(name), nil
    default:
        return nil, fmt.Errorf("unknown framework: %s", frameworkType)
    }
}

// NewServerFromConfig creates server from configuration
func NewServerFromConfig(cfg *Config) (MCPServer, error) {
    return NewServer(cfg.Framework, cfg.Name, cfg.Version)
}
```

---

## Configuration-Based Framework Selection

### 6. Configuration

```go
// internal/config/config.go
package config

type Config struct {
    Framework FrameworkType `yaml:"framework" env:"MCP_FRAMEWORK" default:"go-sdk"`
    Name      string       `yaml:"name" env:"MCP_SERVER_NAME" default:"exarp-go"`
    Version   string       `yaml:"version" env:"MCP_VERSION" default:"1.0.0"`
}

// Load from file, environment, or defaults
func Load() (*Config, error) {
    // Implementation
}
```

**config.yaml:**
```yaml
framework: go-sdk  # or "mcp-go" or "go-mcp"
name: exarp-go
version: 1.0.0
```

**Environment Variable:**
```bash
export MCP_FRAMEWORK=mcp-go
```

---

## Application Code (Framework-Agnostic)

### 7. Tool Registration (Framework-Independent)

```go
// internal/tools/registry.go
package tools

import (
    "context"
    "encoding/json"
    
    "github.com/davidl/exarp-go/internal/framework"
)

// RegisterAllTools registers all tools with any framework
func RegisterAllTools(server framework.MCPServer) error {
    // Register analyze_alignment
    server.RegisterTool(
        "analyze_alignment",
        "[HINT: Alignment analysis. action=todo2|prd. Unified alignment analysis tool.]",
        framework.ToolSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "action": map[string]interface{}{
                    "type":    "string",
                    "enum":    []string{"todo2", "prd"},
                    "default": "todo2",
                },
                "create_followup_tasks": map[string]interface{}{
                    "type":    "boolean",
                    "default": true,
                },
            },
        },
        handleAnalyzeAlignment,
    )
    
    // Register other tools...
    return nil
}

// Tool handlers are framework-agnostic
func handleAnalyzeAlignment(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    // Parse arguments
    var params struct {
        Action              string `json:"action"`
        CreateFollowupTasks bool   `json:"create_followup_tasks"`
    }
    json.Unmarshal(args, &params)
    
    // Execute tool (via Python bridge or native Go)
    result := executePythonTool(ctx, "analyze_alignment", params)
    
    return []framework.TextContent{
        {Type: "text", Text: result},
    }, nil
}
```

### 8. Main Application

```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    
    "github.com/davidl/exarp-go/internal/config"
    "github.com/davidl/exarp-go/internal/framework"
    "github.com/davidl/exarp-go/internal/tools"
    "github.com/davidl/exarp-go/internal/prompts"
    "github.com/davidl/exarp-go/internal/resources"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create server using configured framework
    server, err := framework.NewServerFromConfig(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    // Register components (framework-agnostic)
    if err := tools.RegisterAllTools(server); err != nil {
        log.Fatal(err)
    }
    
    if err := prompts.RegisterAllPrompts(server); err != nil {
        log.Fatal(err)
    }
    
    if err := resources.RegisterAllResources(server); err != nil {
        log.Fatal(err)
    }
    
    // Run server (framework handles transport)
    transport := framework.NewStdioTransport() // Framework-specific
    if err := server.Run(context.Background(), transport); err != nil {
        log.Fatal(err)
    }
}
```

---

## Difficulty Assessment

### Implementation Effort

| Component | Difficulty | Effort | Notes |
|-----------|-----------|--------|-------|
| **Core Interfaces** | 🟢 Easy | 2-4 hours | Define once, used everywhere |
| **Go SDK Adapter** | 🟢 Easy | 2-3 hours | Official SDK, well-documented |
| **MCP-Go Adapter** | 🟡 Moderate | 4-6 hours | Need to study API differences |
| **Go-MCP Adapter** | 🟡 Moderate | 4-6 hours | Need to study API differences |
| **Factory Pattern** | 🟢 Easy | 1-2 hours | Simple switch statement |
| **Configuration** | 🟢 Easy | 2-3 hours | Standard Go config patterns |
| **Testing** | 🟡 Moderate | 4-8 hours | Test each adapter |

**Total Estimated Effort:** 19-32 hours (2.5-4 days)

### Complexity Factors

**Easy Parts:**
- ✅ Go interfaces are powerful for abstraction
- ✅ Adapter pattern is well-understood
- ✅ Configuration is straightforward
- ✅ Application code stays the same

**Moderate Challenges:**
- ⚠️ Each framework has different APIs (need to study)
- ⚠️ Type conversions between frameworks
- ⚠️ Transport differences (stdio vs HTTP)
- ⚠️ Error handling differences

**Potential Issues:**
- ⚠️ Framework API changes (version compatibility)
- ⚠️ Missing features in some frameworks
- ⚠️ Performance differences

---

## Benefits of This Design

### 1. Easy Framework Switching

**Before (Framework-Specific):**
```go
// Hard-coded to go-sdk
server := mcp.NewServer(...)
server.RegisterTool(...)
```

**After (Framework-Agnostic):**
```yaml
# config.yaml
framework: mcp-go  # Just change this!
```

### 2. Testing Multiple Frameworks

```go
func TestWithAllFrameworks(t *testing.T) {
    frameworks := []framework.FrameworkType{
        framework.FrameworkGoSDK,
        framework.FrameworkMCPGo,
        framework.FrameworkGoMCP,
    }
    
    for _, fw := range frameworks {
        t.Run(string(fw), func(t *testing.T) {
            server, _ := framework.NewServer(fw, "test", "1.0.0")
            // Test with this framework
        })
    }
}
```

### 3. Gradual Migration

- Start with go-sdk (official)
- Test with mcp-go (largest community)
- Compare performance
- Switch based on results

### 4. Vendor Lock-In Prevention

- Not tied to one framework
- Can switch if framework becomes unmaintained
- Can use best features from each

---

## Project Structure

```
exarp-go/
├── cmd/
│   └── server/
│       └── main.go              # Framework-agnostic main
├── internal/
│   ├── framework/               # Framework abstraction
│   │   ├── server.go           # Core interfaces
│   │   ├── factory.go          # Framework factory
│   │   └── adapters/
│   │       ├── gosdk/
│   │       │   └── adapter.go  # Go SDK implementation
│   │       ├── mcpgo/
│   │       │   └── adapter.go  # MCP-Go implementation
│   │       └── gomcp/
│   │           └── adapter.go  # Go-MCP implementation
│   ├── tools/
│   │   └── registry.go         # Framework-agnostic tools
│   ├── prompts/
│   │   └── registry.go         # Framework-agnostic prompts
│   ├── resources/
│   │   └── handlers.go         # Framework-agnostic resources
│   └── config/
│       └── config.go           # Configuration management
├── config.yaml                 # Framework selection
└── go.mod
```

---

## Migration Path

### Step 1: Start with One Framework

```go
// Start with go-sdk
server, _ := framework.NewServer(framework.FrameworkGoSDK, "exarp-go", "1.0.0")
```

### Step 2: Add Interface Abstraction

```go
// Extract to interface
var mcpServer framework.MCPServer = gosdk.NewGoSDKAdapter(...)
```

### Step 3: Add Factory

```go
// Use factory
server, _ := framework.NewServer(framework.FrameworkGoSDK, ...)
```

### Step 4: Add Configuration

```go
// Load from config
cfg := config.Load()
server, _ := framework.NewServerFromConfig(cfg)
```

### Step 5: Add More Adapters

```go
// Add mcp-go adapter
// Add go-mcp adapter
```

### Step 6: Test All Frameworks

```go
// Test with each framework
// Compare performance
// Choose best for production
```

---

## Example: Switching Frameworks

### Current (go-sdk):
```yaml
# config.yaml
framework: go-sdk
```

### Switch to mcp-go:
```yaml
# config.yaml
framework: mcp-go
```

**That's it!** No code changes needed.

### Or via environment:
```bash
export MCP_FRAMEWORK=mcp-go
./bin/exarp-go
```

---

## Testing Strategy

### 1. Interface Tests

```go
// Test framework interface compliance
func TestFrameworkInterface(t *testing.T) {
    frameworks := []framework.FrameworkType{...}
    for _, fw := range frameworks {
        server, _ := framework.NewServer(fw, "test", "1.0.0")
        // Test interface methods
    }
}
```

### 2. Integration Tests

```go
// Test actual MCP protocol
func TestMCPProtocol(t *testing.T) {
    // Test with each framework
    // Verify protocol compliance
}
```

### 3. Performance Tests

```go
// Benchmark each framework
func BenchmarkGoSDK(b *testing.B) { ... }
func BenchmarkMCPGo(b *testing.B) { ... }
func BenchmarkGoMCP(b *testing.B) { ... }
```

---

## Recommendations

### 1. Start Simple

- Begin with go-sdk (official)
- Add abstraction layer early
- Add other frameworks later

### 2. Keep It Simple

- Don't over-engineer
- Add adapters only when needed
- Focus on core functionality first

### 3. Test Thoroughly

- Test each adapter
- Verify protocol compliance
- Compare performance

### 4. Document Differences

- Document API differences
- Note framework-specific features
- Keep migration notes

---

## Conclusion

**Difficulty:** 🟢 **Easy to Moderate**

**Key Points:**
- ✅ Go interfaces make abstraction natural
- ✅ Adapter pattern is straightforward
- ✅ Configuration-based switching is easy
- ✅ Application code stays framework-agnostic
- ⚠️ Need to study each framework's API
- ⚠️ Some type conversions required

**Recommendation:**
- **Start with abstraction layer** (2-3 days)
- **Add go-sdk adapter first** (1 day)
- **Add other adapters as needed** (1-2 days each)
- **Total: 4-7 days** for full framework-agnostic design

**Benefits:**
- Easy framework switching
- No vendor lock-in
- Can test multiple frameworks
- Future-proof design

---

## Next Steps

1. **Design core interfaces** (framework/server.go)
2. **Implement go-sdk adapter** (start with official)
3. **Add factory pattern** (framework/factory.go)
4. **Add configuration** (config/config.go)
5. **Test with go-sdk** (verify abstraction works)
6. **Add mcp-go adapter** (when needed)
7. **Add go-mcp adapter** (when needed)

**Result:** Framework-agnostic code that can switch with a config change! 🎉

