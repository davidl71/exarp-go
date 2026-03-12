# MCP Client Wrapper Design

**Design Date:** 2025-01-27  
**Purpose:** Wrapper functions around existing MCP client libraries for `mcp-go-core`

## Design Goals

1. **Reuse Existing Libraries:** Wrap `github.com/metoro-io/mcp-golang` (well-documented, actively maintained)
2. **Type Integration:** Convert between external library types and `mcp-go-core` types
3. **Testing Utilities:** Provide utilities for testing `mcp-go-core` servers
4. **Consistent Patterns:** Follow `mcp-go-core` design patterns and conventions
5. **Optional Dependency:** Make the wrapper optional (only import if needed)

## Package Structure

```
pkg/mcp/client/
├── client.go          # Main wrapper client
├── convert.go         # Type conversion utilities
├── testutil.go        # Testing utilities
└── client_test.go     # Tests
```

## API Design

### High-Level Client Wrapper

```go
package client

import (
    "context"
    "github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
    "github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// Client wraps an external MCP client library
type Client struct {
    // Underlying client (github.com/metoro-io/mcp-golang)
    underlying interface{} // *mcp.Client from mcp-golang
    
    // Client info
    clientInfo protocol.ClientInfo
}

// NewClient creates a new client wrapper
func NewClient(serverCommand string, clientInfo protocol.ClientInfo) (*Client, error)

// Initialize initializes the client session
func (c *Client) Initialize(ctx context.Context) (*protocol.InitializeResult, error)

// ListTools lists all available tools (using mcp-go-core types)
func (c *Client) ListTools(ctx context.Context) ([]types.ToolInfo, error)

// CallTool calls a tool with arguments
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) ([]types.TextContent, error)

// ListResources lists all available resources
func (c *Client) ListResources(ctx context.Context) ([]protocol.Resource, error)

// ReadResource reads a resource by URI
func (c *Client) ReadResource(ctx context.Context, uri string) ([]byte, string, error)

// Close closes the client connection
func (c *Client) Close() error
```

### Type Conversion Utilities

```go
package client

// ConvertExternalToolToToolInfo converts external library tool to mcp-go-core ToolInfo
func ConvertExternalToolToToolInfo(externalTool interface{}) (types.ToolInfo, error)

// ConvertTextContent converts external text content to mcp-go-core TextContent
func ConvertTextContent(externalContent interface{}) (types.TextContent, error)
```

### Testing Utilities

```go
package client

// TestServerConfig configures a test client for a server
type TestServerConfig struct {
    ServerCommand string
    ServerArgs    []string
    ClientInfo    protocol.ClientInfo
}

// NewTestClient creates a client for testing mcp-go-core servers
func NewTestClient(config TestServerConfig) (*Client, error)

// TestToolExecution tests a tool execution end-to-end
func TestToolExecution(ctx context.Context, client *Client, toolName string, args map[string]interface{}) error

// AssertToolExists asserts that a tool exists and matches expected schema
func AssertToolExists(ctx context.Context, client *Client, toolName string, expectedSchema types.ToolSchema) error
```

## Implementation Strategy

### 1. Optional Dependency Pattern

Use build tags to make the wrapper optional:

```go
// +build mcp_client

package client

import (
    mcp "github.com/metoro-io/mcp-golang"
    "github.com/metoro-io/mcp-golang/transport/stdio"
)
```

This way, users who don't need the client wrapper don't need to install the dependency.

### 2. Type Conversion Layer

Create conversion functions that map between external library types and mcp-go-core types:

```go
// External library types (from mcp-golang)
// - Tool (different structure)
// - TextContent (similar but may have different field names)

// mcp-go-core types
// - types.ToolInfo
// - types.TextContent
// - protocol.Resource
```

### 3. Wrapper Implementation

The wrapper stores the underlying client and provides methods that:
- Call the underlying client
- Convert types to mcp-go-core types
- Handle errors consistently

### 4. Testing Utilities

Provide utilities specifically for testing mcp-go-core servers:
- Easy client creation for testing
- Assertion helpers
- Integration test utilities

## Usage Examples

### Basic Usage

```go
import (
    "context"
    "github.com/davidl71/mcp-go-core/pkg/mcp/client"
    "github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
)

func main() {
    clientInfo := protocol.ClientInfo{
        Name:    "test-client",
        Version: "1.0.0",
    }
    
    c, err := client.NewClient("/path/to/server", clientInfo)
    if err != nil {
        log.Fatal(err)
    }
    defer c.Close()
    
    ctx := context.Background()
    
    // Initialize
    _, err = c.Initialize(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // List tools (returns mcp-go-core types.ToolInfo)
    tools, err := c.ListTools(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, tool := range tools {
        fmt.Printf("Tool: %s\n", tool.Name)
    }
    
    // Call tool
    args := map[string]interface{}{
        "message": "Hello, World!",
    }
    result, err := c.CallTool(ctx, "echo", args)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, content := range result {
        fmt.Println(content.Text)
    }
}
```

### Testing Usage

```go
import (
    "context"
    "testing"
    "github.com/davidl71/mcp-go-core/pkg/mcp/client"
    "github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
)

func TestMyServer(t *testing.T) {
    config := client.TestServerConfig{
        ServerCommand: "./bin/my-server",
        ClientInfo: protocol.ClientInfo{
            Name:    "test-client",
            Version: "1.0.0",
        },
    }
    
    c, err := client.NewTestClient(config)
    if err != nil {
        t.Fatal(err)
    }
    defer c.Close()
    
    ctx := context.Background()
    
    // Initialize
    _, err = c.Initialize(ctx)
    if err != nil {
        t.Fatal(err)
    }
    
    // Test tool execution
    args := map[string]interface{}{
        "operation": "add",
        "a": 10,
        "b": 5,
    }
    
    result, err := c.CallTool(ctx, "calculate", args)
    if err != nil {
        t.Fatal(err)
    }
    
    if len(result) == 0 {
        t.Error("Expected result, got empty")
    }
    
    // Assert tool exists
    expectedSchema := types.ToolSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "operation": map[string]interface{}{"type": "string"},
            "a": map[string]interface{}{"type": "number"},
            "b": map[string]interface{}{"type": "number"},
        },
        Required: []string{"operation", "a", "b"},
    }
    
    err = client.AssertToolExists(ctx, c, "calculate", expectedSchema)
    if err != nil {
        t.Error(err)
    }
}
```

## Benefits

1. **No Duplication:** Reuses existing, well-tested client library
2. **Type Consistency:** Uses mcp-go-core types throughout
3. **Testing Support:** Provides utilities for testing mcp-go-core servers
4. **Optional:** Only install dependency if wrapper is needed
5. **Maintainable:** Changes to underlying library don't require full reimplementation

## Considerations

1. **Dependency Management:** External library becomes a dependency (can be optional with build tags)
2. **Type Conversion:** Need to handle type conversions carefully
3. **API Stability:** Wrapper API should be stable even if underlying library changes
4. **Error Handling:** Need to convert errors from external library to mcp-go-core error patterns

## Alternatives Considered

1. **Full Implementation:** Rejected - too much duplication, existing libraries work well
2. **Direct Usage:** Rejected - doesn't provide type integration with mcp-go-core
3. **Wrapper (Chosen):** Best balance of reuse and integration

## Implementation Plan

1. **Phase 1: Basic Wrapper**
   - Create client wrapper structure
   - Implement Initialize, ListTools, CallTool
   - Type conversion utilities

2. **Phase 2: Full API**
   - Add ListResources, ReadResource
   - Add ListPrompts, GetPrompt
   - Complete type conversions

3. **Phase 3: Testing Utilities**
   - Test client creation
   - Assertion helpers
   - Integration test utilities

4. **Phase 4: Documentation & Examples**
   - Usage documentation
   - Example integrations
   - Testing examples
