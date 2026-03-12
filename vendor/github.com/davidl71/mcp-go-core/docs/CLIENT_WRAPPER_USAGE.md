# MCP Client Wrapper Usage Guide

This guide demonstrates how to use the `mcp-go-core` client wrapper to interact with MCP servers.

## Overview

The client wrapper provides a convenient way to interact with MCP servers using `mcp-go-core` types. It wraps external client libraries (like `github.com/metoro-io/mcp-golang`) and converts types to match `mcp-go-core` conventions.

## Installation

### Prerequisites

1. **Add the external client library dependency:**
   ```bash
   go get github.com/metoro-io/mcp-golang
   ```

2. **Ensure mcp-go-core is in your project:**
   ```bash
   go get github.com/davidl71/mcp-go-core
   ```

### Building

To build with the client wrapper (default):
```bash
go build
```

To build without the client wrapper (to avoid the dependency):
```bash
go build -tags no_mcp_client
```

## Basic Usage

### Creating a Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/davidl71/mcp-go-core/pkg/mcp/client"
    "github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
)

func main() {
    // Define client information
    clientInfo := protocol.ClientInfo{
        Name:    "my-client",
        Version: "1.0.0",
    }

    // Create a client pointing to your MCP server
    c, err := client.NewClient("/path/to/your/server", clientInfo)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer c.Close()

    ctx := context.Background()

    // Initialize the client
    initResult, err := c.Initialize(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }

    fmt.Printf("Connected to server: %s v%s\n",
        initResult.ServerInfo.Name,
        initResult.ServerInfo.Version)
}
```

### Listing Tools

```go
// List all available tools
tools, err := c.ListTools(ctx)
if err != nil {
    log.Fatalf("Failed to list tools: %v", err)
}

fmt.Printf("Available tools (%d):\n", len(tools))
for _, tool := range tools {
    fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
    fmt.Printf("    Schema: %s\n", tool.Schema.Type)
}
```

### Calling a Tool

```go
// Call a tool with arguments
args := map[string]interface{}{
    "message": "Hello, MCP!",
}

result, err := c.CallTool(ctx, "echo", args)
if err != nil {
    log.Fatalf("Failed to call tool: %v", err)
}

// Process results (returns []types.TextContent)
for _, content := range result {
    fmt.Printf("Response: %s\n", content.Text)
}
```

### Working with Resources

```go
// List available resources
resources, err := c.ListResources(ctx)
if err != nil {
    log.Fatalf("Failed to list resources: %v", err)
}

fmt.Printf("Available resources (%d):\n", len(resources))
for _, resource := range resources {
    fmt.Printf("  - %s: %s\n", resource.URI, resource.Name)
}

// Read a specific resource
content, mimeType, err := c.ReadResource(ctx, "resource://example/data")
if err != nil {
    log.Fatalf("Failed to read resource: %v", err)
}

fmt.Printf("Resource MIME type: %s\n", mimeType)
fmt.Printf("Resource content:\n%s\n", string(content))
```

### Working with Prompts

```go
// List available prompts
prompts, err := c.ListPrompts(ctx)
if err != nil {
    log.Fatalf("Failed to list prompts: %v", err)
}

fmt.Printf("Available prompts (%d):\n", len(prompts))
for _, prompt := range prompts {
    fmt.Printf("  - %s: %s\n", prompt.Name, prompt.Description)
}

// Get a prompt template
promptArgs := map[string]interface{}{
    "name": "Alice",
}

promptText, err := c.GetPrompt(ctx, "greeting", promptArgs)
if err != nil {
    log.Fatalf("Failed to get prompt: %v", err)
}

fmt.Printf("Prompt result:\n%s\n", promptText)
```

## Advanced Usage

### Using Context for Timeouts

```go
import "time"

// Create a context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// All client operations respect the context
tools, err := c.ListTools(ctx)
if err != nil {
    if err == context.DeadlineExceeded {
        log.Fatal("Operation timed out")
    }
    log.Fatalf("Error: %v", err)
}
```

### Using Server Arguments

```go
// Create client with server arguments
serverArgs := []string{"--verbose", "--config", "/path/to/config"}
c, err := client.NewClientWithArgs(
    "/path/to/server",
    serverArgs,
    clientInfo,
)
```

### Type-Safe Tool Arguments

While the client accepts `map[string]interface{}` for flexibility, you can use structs:

```go
type CalculateArgs struct {
    Operation string `json:"operation"`
    A         int    `json:"a"`
    B         int    `json:"b"`
}

args := CalculateArgs{
    Operation: "add",
    A:         10,
    B:         20,
}

// Convert to map for the client
argsMap := map[string]interface{}{
    "operation": args.Operation,
    "a":         args.A,
    "b":         args.B,
}

result, err := c.CallTool(ctx, "calculate", argsMap)
```

## Testing Usage

### Using Test Utilities

The package provides utilities specifically for testing:

```go
package mypackage_test

import (
    "context"
    "testing"

    "github.com/davidl71/mcp-go-core/pkg/mcp/client"
    "github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
    "github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

func TestMyServer(t *testing.T) {
    // Create test client
    config := client.TestServerConfig{
        ServerCommand: "./bin/my-server",
        ClientInfo: protocol.ClientInfo{
            Name:    "test-client",
            Version: "1.0.0",
        },
    }

    c, err := client.NewTestClient(config)
    if err != nil {
        t.Fatalf("Failed to create test client: %v", err)
    }
    defer c.Close()

    ctx := context.Background()

    // Initialize
    _, err = c.Initialize(ctx)
    if err != nil {
        t.Fatalf("Failed to initialize: %v", err)
    }

    // Test tool execution
    args := map[string]interface{}{
        "message": "test",
    }
    result, err := client.TestToolExecution(ctx, c, "echo", args)
    if err != nil {
        t.Fatalf("Tool execution failed: %v", err)
    }

    if len(result) == 0 {
        t.Error("Expected result, got empty")
    }

    // Assert tool exists with schema validation
    expectedSchema := types.ToolSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "message": map[string]interface{}{
                "type": "string",
            },
        },
        Required: []string{"message"},
    }

    err = client.AssertToolExists(ctx, c, "echo", &expectedSchema)
    if err != nil {
        t.Errorf("Tool assertion failed: %v", err)
    }

    // Test server capabilities
    caps, err := client.TestServerCapabilities(ctx, c)
    if err != nil {
        t.Fatalf("Failed to test capabilities: %v", err)
    }

    t.Logf("Server capabilities: Tools=%v, Resources=%v, Prompts=%v",
        caps.ToolsAvailable,
        caps.ResourcesAvailable,
        caps.PromptsAvailable)
}
```

## Complete Example

Here's a complete example that demonstrates all features:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/davidl71/mcp-go-core/pkg/mcp/client"
    "github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
)

func main() {
    if len(os.Args) < 2 {
        log.Fatal("Usage: client-example <server-command>")
    }

    serverCommand := os.Args[1]

    // Create client
    clientInfo := protocol.ClientInfo{
        Name:    "example-client",
        Version: "1.0.0",
    }

    c, err := client.NewClient(serverCommand, clientInfo)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer c.Close()

    ctx := context.Background()

    // Initialize
    fmt.Println("Initializing client...")
    initResult, err := c.Initialize(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }
    fmt.Printf("Connected to: %s v%s\n\n",
        initResult.ServerInfo.Name,
        initResult.ServerInfo.Version)

    // List tools
    fmt.Println("Listing tools...")
    tools, err := c.ListTools(ctx)
    if err != nil {
        log.Fatalf("Failed to list tools: %v", err)
    }
    fmt.Printf("Found %d tools:\n", len(tools))
    for _, tool := range tools {
        fmt.Printf("  - %s\n", tool.Name)
    }
    fmt.Println()

    // Call a tool (if available)
    if len(tools) > 0 {
        toolName := tools[0].Name
        fmt.Printf("Calling tool: %s\n", toolName)

        args := map[string]interface{}{
            "message": "Hello from client!",
        }

        result, err := c.CallTool(ctx, toolName, args)
        if err != nil {
            log.Printf("Failed to call tool: %v", err)
        } else {
            fmt.Println("Tool result:")
            for _, content := range result {
                fmt.Printf("  %s\n", content.Text)
            }
        }
        fmt.Println()
    }

    // List resources
    fmt.Println("Listing resources...")
    resources, err := c.ListResources(ctx)
    if err != nil {
        log.Printf("Failed to list resources: %v", err)
    } else {
        fmt.Printf("Found %d resources:\n", len(resources))
        for _, resource := range resources {
            fmt.Printf("  - %s\n", resource.URI)
        }
    }
}
```

## Error Handling

All client methods return errors that should be checked:

```go
result, err := c.CallTool(ctx, "tool-name", args)
if err != nil {
    // Handle different error types
    if err == context.DeadlineExceeded {
        log.Println("Operation timed out")
    } else {
        log.Printf("Error: %v", err)
    }
    return
}
```

Common errors:
- `client must be initialized` - Call `Initialize()` first
- `tool name cannot be empty` - Invalid tool name
- `failed to initialize` - Server connection/initialization failed
- Context errors - Timeout or cancellation

## Best Practices

1. **Always initialize:** Call `Initialize()` before other operations
2. **Use context:** Pass context for timeouts and cancellation
3. **Handle errors:** Check all errors appropriately
4. **Close client:** Use `defer c.Close()` to ensure cleanup
5. **Type safety:** Use structs for complex tool arguments
6. **Testing:** Use test utilities for integration tests

## Limitations

- The client wrapper requires `github.com/metoro-io/mcp-golang` as a dependency
- Build with `-tags no_mcp_client` to exclude this package if not needed
- Type conversions use JSON marshaling (small performance overhead)
- Some external library features may not be fully exposed

## See Also

- [Client Wrapper Design](CLIENT_WRAPPER_DESIGN.md) - Design documentation
- [Client Wrapper Implementation](CLIENT_WRAPPER_IMPLEMENTATION.md) - Implementation status
- [MCP Client Research](MCP_CLIENT_RESEARCH.md) - Research on MCP clients
