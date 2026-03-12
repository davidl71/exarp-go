# MCP Client Research

**Research Date:** 2025-01-27  
**Project:** mcp-go-core  
**Focus:** Model Context Protocol (MCP) Client Implementation Research

## Executive Summary

The `mcp-go-core` library currently focuses on **MCP server infrastructure**. This research document explores what an MCP client implementation would look like, the protocol requirements, and potential implementation patterns.

## Current State

### What mcp-go-core Provides (Server-Side)

✅ **Server Infrastructure:**
- Framework abstraction (`pkg/mcp/framework/server.go`)
- Transport implementations (stdio, SSE)
- Protocol types (`pkg/mcp/protocol/protocol.go`)
- Security utilities
- CLI utilities
- Configuration management

❌ **Missing (Client-Side):**
- Client connection management
- Client request/response handling
- Client transport implementations
- Client session management
- Client tool/resource discovery

## MCP Protocol Overview

### Protocol Specification

**Transport:** JSON-RPC 2.0 over stdio or HTTP/SSE  
**Protocol Version:** 2024-11-05 (current)

### Key Protocol Types (from `protocol/protocol.go`)

```go
// Client → Server types
type InitializeParams struct {
    ProtocolVersion string             `json:"protocolVersion"`
    Capabilities    ClientCapabilities `json:"capabilities"`
    ClientInfo      ClientInfo         `json:"clientInfo"`
}

type ClientCapabilities struct {
    Experimental map[string]interface{} `json:"experimental,omitempty"`
}

type ClientInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

// Server → Client types
type InitializeResult struct {
    ProtocolVersion string             `json:"protocolVersion"`
    Capabilities    ServerCapabilities `json:"capabilities"`
    ServerInfo      ServerInfo         `json:"serverInfo"`
}
```

### Message Flow

1. **Initialize Handshake:**
   - Client sends `initialize` request with client info
   - Server responds with server capabilities and info
   - Client sends `initialized` notification (optional)

2. **Tool Discovery:**
   - Client requests `tools/list` to discover available tools
   - Server responds with tool definitions

3. **Tool Execution:**
   - Client sends `tools/call` with tool name and arguments
   - Server executes tool and returns results

4. **Resource Discovery:**
   - Client requests `resources/list` to discover available resources
   - Client requests `resources/read` to read specific resources

5. **Prompt Templates:**
   - Client requests `prompts/list` to discover prompts
   - Client requests `prompts/get` to get prompt templates

## MCP Client Requirements

### Core Functionality

1. **Connection Management:**
   - Establish connection to MCP server (stdio or HTTP/SSE)
   - Handle connection lifecycle
   - Manage reconnection logic

2. **Protocol Communication:**
   - Send JSON-RPC 2.0 requests
   - Handle JSON-RPC 2.0 responses
   - Manage request/response correlation (IDs)
   - Handle errors and notifications

3. **Session Management:**
   - Initialize session with server
   - Track session state
   - Handle session termination

4. **Capability Discovery:**
   - Discover server capabilities
   - Discover available tools
   - Discover available resources
   - Discover available prompts

5. **Tool Execution:**
   - Execute tools with arguments
   - Handle tool responses
   - Support streaming responses (if applicable)

6. **Resource Access:**
   - List available resources
   - Read resource contents
   - Handle resource metadata

## Potential Client Implementation

### High-Level Interface

```go
package client

// MCPClient represents an MCP client connection
type MCPClient interface {
    // Initialize establishes a connection and initializes the session
    Initialize(ctx context.Context, params InitializeParams) (*InitializeResult, error)
    
    // ListTools returns all available tools
    ListTools(ctx context.Context) ([]Tool, error)
    
    // CallTool executes a tool
    CallTool(ctx context.Context, name string, arguments map[string]interface{}) ([]TextContent, error)
    
    // ListResources returns all available resources
    ListResources(ctx context.Context) ([]Resource, error)
    
    // ReadResource reads a resource by URI
    ReadResource(ctx context.Context, uri string) ([]byte, string, error)
    
    // ListPrompts returns all available prompts
    ListPrompts(ctx context.Context) ([]Prompt, error)
    
    // GetPrompt gets a prompt template
    GetPrompt(ctx context.Context, name string, args map[string]interface{}) (string, error)
    
    // Close closes the connection
    Close() error
}
```

### Transport Abstraction

```go
// ClientTransport abstracts transport for MCP clients
type ClientTransport interface {
    // Connect establishes a connection
    Connect(ctx context.Context) error
    
    // SendRequest sends a JSON-RPC request and returns response
    SendRequest(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error)
    
    // Close closes the transport
    Close() error
}
```

### Stdio Transport Implementation

```go
type StdioTransport struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.Reader
    stderr io.Reader
    mu     sync.Mutex
    reqID  int64
    pending map[interface{}]chan *protocol.JSONRPCResponse
}

func (t *StdioTransport) Connect(ctx context.Context) error {
    // Start server process
    // Set up stdin/stdout pipes
    // Start response reader goroutine
}

func (t *StdioTransport) SendRequest(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
    // Assign request ID
    // Create response channel
    // Send request via stdin
    // Wait for response or timeout
}
```

### HTTP/SSE Transport Implementation

```go
type HTTPTransport struct {
    baseURL string
    client  *http.Client
    conn    *sse.Conn
}

func (t *HTTPTransport) Connect(ctx context.Context) error {
    // Establish SSE connection
    // Start event stream reader
}

func (t *HTTPTransport) SendRequest(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
    // Send HTTP POST request
    // Parse response
}
```

## Existing MCP Clients

### Application-Level Clients (Hosts)

1. **Cursor IDE** - MCP client built into Cursor
   - Connects to MCP servers via stdio
   - Configuration via `.cursor/mcp.json`
   - Used by exarp-go and devwisdom-go

2. **Claude Desktop** - Anthropic's MCP client
   - Standard MCP client implementation
   - Connects to servers via stdio

3. **Model Context Client** - Web-based MCP client
   - Supports multiple transports
   - Web interface for MCP interaction

### Go Client Libraries

Several Go client libraries have been developed by the community:

#### 1. github.com/metoro-io/mcp-golang

**Status:** Active, well-documented  
**Documentation:** https://mcpgolang.com/client  
**Features:**
- High-level client API
- Stdio and HTTP transport support
- Type-safe tool/prompt arguments
- Context support for timeouts/cancellation
- Pagination support for `ListTools()` and `ListPrompts()`
- Comprehensive error handling

**API Example:**
```go
import (
    mcp "github.com/metoro-io/mcp-golang"
    "github.com/metoro-io/mcp-golang/transport/stdio"
)

// Create transport and client
transport := stdio.NewStdioClientTransport()
client := mcp.NewClient(transport)

// Initialize
response, err := client.Initialize(context.Background())
if err != nil {
    log.Fatalf("Failed to initialize: %v", err)
}

// List tools
tools, err := client.ListTools(context.Background(), nil)
if err != nil {
    log.Fatalf("Failed to list tools: %v", err)
}

// Call tool with type-safe arguments
type CalculateArgs struct {
    Operation string `json:"operation"`
    A         int    `json:"a"`
    B         int    `json:"b"`
}

args := CalculateArgs{Operation: "add", A: 10, B: 5}
response, err := client.CallTool(context.Background(), "calculate", args)
```

**Transport Options:**
- Stdio: `stdio.NewStdioClientTransport()` (supports bidirectional communication)
- HTTP: `http.NewHTTPClientTransport("/mcp")` (stateless, no notifications)

**Key Methods:**
- `Initialize(ctx)` - Initialize client session
- `ListTools(ctx, cursor)` - List available tools (with pagination)
- `CallTool(ctx, name, args)` - Execute a tool
- `ListPrompts(ctx, cursor)` - List available prompts (with pagination)
- `GetPrompt(ctx, name, args)` - Get prompt template
- `ListResources(ctx, cursor)` - List available resources
- `ReadResource(ctx, uri)` - Read resource content

#### 2. github.com/mark3labs/mcp-go

**Status:** Active  
**Documentation:** https://pkg.go.dev/github.com/mark3labs/mcp-go/client  
**Features:**
- Client package with multiple transport options
- STDIO, StreamableHTTP, and SSE client implementations
- Programmatic interface for MCP operations

**Resources:**
- Package documentation: https://pkg.go.dev/github.com/mark3labs/mcp-go/client
- Repository: https://github.com/mark3labs/mcp-go
- Client guide: https://mcp-go.dev/clients/

#### 3. Other Implementations

- **FastMCP** (Python-focused but provides architectural insights)
- Community implementations mentioned in articles and discussions

### Comparison: Existing Libraries vs mcp-go-core

**What existing libraries provide:**
- ✅ Complete client implementations
- ✅ Transport abstractions (stdio, HTTP)
- ✅ High-level APIs
- ✅ Type-safe argument handling
- ✅ Context support

**What mcp-go-core could contribute:**
- ✅ Reusable protocol types (already exists)
- ✅ Consistent patterns with server implementation
- ✅ Integration with existing server infrastructure
- ✅ Shared transport implementations (if client/server share code)

**Key Insight:**
Since multiple Go MCP client libraries already exist, `mcp-go-core` should consider:
1. **Integration approach:** Provide client utilities that work with existing libraries
2. **Protocol alignment:** Ensure protocol types are compatible with existing clients
3. **Testing utilities:** Provide client utilities for testing MCP servers
4. **Documentation:** Link to existing client libraries in documentation

### Client-Server Communication Pattern

**From Cursor/Claude Desktop perspective:**
- Client launches server process (stdio)
- Client sends JSON-RPC requests via stdin
- Client reads JSON-RPC responses from stdout
- Client handles errors from stderr

**Example from devwisdom-go:**
```json
// Client sends:
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "cursor-ide",
      "version": "1.0.0"
    }
  }
}

// Server responds:
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {},
      "resources": {}
    },
    "serverInfo": {
      "name": "devwisdom-go",
      "version": "0.1.0"
    }
  }
}
```

## Implementation Considerations

### Design Patterns

1. **Builder Pattern** for client construction:
   ```go
   client := NewClient().
       WithTransport(StdioTransport{Command: "my-server"}).
       WithClientInfo(ClientInfo{Name: "my-client", Version: "1.0.0"}).
       Build()
   ```

2. **Request/Response Correlation:**
   - Use request IDs to match responses
   - Handle concurrent requests
   - Timeout handling

3. **Error Handling:**
   - JSON-RPC error codes
   - Transport errors
   - Protocol errors

4. **Connection Management:**
   - Connection pooling (for HTTP)
   - Reconnection logic
   - Graceful shutdown

### Dependencies

**Would need:**
- JSON-RPC 2.0 implementation (already in `pkg/mcp/protocol`)
- Transport implementations (stdio, HTTP/SSE)
- Context support for cancellation
- Error handling utilities

**Could reuse from mcp-go-core:**
- Protocol types (`pkg/mcp/protocol`)
- Logging utilities (`pkg/mcp/logging`)
- Error types (if extended)

### Testing Strategy

1. **Unit Tests:**
   - Request/response handling
   - Protocol encoding/decoding
   - Error handling

2. **Integration Tests:**
   - Test against real MCP servers (exarp-go, devwisdom-go)
   - Test stdio transport
   - Test HTTP/SSE transport (if implemented)

3. **Example Clients:**
   - CLI client for testing
   - Library client for embedding

## Use Cases for MCP Client

### 1. Testing MCP Servers

A Go client could be used to:
- Test MCP server implementations
- Validate protocol compliance
- Automated testing of tools/resources/prompts

### 2. CLI Tools

A client library could enable:
- Command-line tools that connect to MCP servers
- Scripting with MCP servers
- Automation workflows

### 3. Embedded Clients

A client library could enable:
- Go applications that consume MCP servers
- Microservices that use MCP for communication
- Multi-server orchestration

### 4. Development Tools

A client library could enable:
- MCP server debugging tools
- Protocol analyzers
- Server capability inspectors

## Recommendations

Given that multiple Go MCP client libraries already exist, `mcp-go-core` should take a **complementary approach** rather than reimplementing client functionality.

### Recommended Approach: Integration & Utilities

**Option 1: Protocol Compatibility (Recommended)**
- Ensure `mcp-go-core` protocol types are compatible with existing clients
- Provide shared types that clients can import
- Focus on server infrastructure (current focus)
- Document compatibility with existing client libraries

**Option 2: Testing Utilities**
- Create lightweight client utilities for **testing MCP servers**
- Provide test helpers that work with existing client libraries
- Create integration test utilities
- Focus on developer tooling rather than full client implementation

**Option 3: Documentation & Examples**
- Document how to use existing client libraries with `mcp-go-core` servers
- Provide example integrations
- Link to recommended client libraries
- Create comparison guides

### Short Term Actions

1. **Research & Documentation:**
   - ✅ Document protocol requirements (this document)
   - ✅ Research existing Go client libraries
   - Document compatibility with existing clients
   - Add links to client libraries in README

2. **Protocol Alignment:**
   - Verify protocol types are compatible with existing clients
   - Test `mcp-go-core` servers with existing client libraries
   - Document any compatibility considerations

3. **Examples & Integration:**
   - Create example showing `mcp-golang` client connecting to `mcp-go-core` server
   - Create example showing `mcp-go` client connecting to `mcp-go-core` server
   - Document integration patterns

### Medium Term (If Client Utilities Needed)

1. **Testing Utilities:**
   - Create lightweight client utilities for testing
   - Provide test helpers for server integration tests
   - Create mock client implementations for testing

2. **Development Tools:**
   - Create CLI tools for testing MCP servers
   - Create debugging utilities
   - Create protocol analyzers

### Long Term (Only if Unique Value Proposition)

1. **Client Implementation (Only if needed):**
   - Only implement if existing libraries don't meet specific needs
   - Focus on unique features (e.g., advanced testing, debugging)
   - Integrate with `mcp-go-core` server infrastructure

2. **Ecosystem Integration:**
   - Collaborate with existing client library maintainers
   - Ensure protocol types are shared/standardized
   - Contribute improvements to existing libraries if needed

## Questions to Answer

1. **Scope:**
   - Should client be part of mcp-go-core or separate library?
   - What's the primary use case (testing, CLI, embedding)?

2. **Transport Priority:**
   - Start with stdio (most common)?
   - Include HTTP/SSE from start?
   - Support multiple transports simultaneously?

3. **API Design:**
   - High-level API (like interface above)?
   - Low-level JSON-RPC API?
   - Both?

4. **Dependencies:**
   - Minimal dependencies?
   - Use existing JSON-RPC libraries?
   - Reuse mcp-go-core protocol types?

## References

### Codebase References

- `pkg/mcp/protocol/protocol.go` - Protocol types and structures
- `pkg/mcp/framework/transport.go` - Server transport implementations
- `pkg/mcp/framework/server.go` - Server interface (reference for client design)

### External References

- [Model Context Protocol Specification](https://modelcontextprotocol.io/) (official spec)
- [Anthropic MCP Documentation](https://docs.anthropic.com/en/docs/build-with-claude/mcp)
- Cursor IDE MCP integration (practical example)
- Claude Desktop MCP integration (reference implementation)

### Go MCP Client Libraries

- **[mcp-golang](https://mcpgolang.com/client)** - Comprehensive Go MCP client library (`github.com/metoro-io/mcp-golang`)
  - Documentation: https://mcpgolang.com/client
  - Supports stdio and HTTP transports
  - Type-safe APIs, context support, pagination

- **[mcp-go](https://github.com/mark3labs/mcp-go)** - Go MCP client package (`github.com/mark3labs/mcp-go`)
  - Package docs: https://pkg.go.dev/github.com/mark3labs/mcp-go/client
  - Client guide: https://mcp-go.dev/clients/
  - Multiple transport options (STDIO, StreamableHTTP, SSE)

- **[Creating an MCP Server Using Go](https://dev.to/eminetto/creating-an-mcp-server-using-go-3foe)** - Article discussing MCP in Go ecosystem

### Related Projects

- `exarp-go` - MCP server implementation (test target)
- `devwisdom-go` - MCP server implementation (test target)
- `mcp-go-core` - Current project (server infrastructure)

## Conclusion

**Key Findings:**

1. **Existing Go Client Libraries:**
   - ✅ Multiple Go MCP client libraries already exist (mcp-golang, mcp-go)
   - ✅ Libraries provide comprehensive client functionality
   - ✅ Well-documented and actively maintained

2. **mcp-go-core Status:**
   - ✅ Protocol types exist in `pkg/mcp/protocol`
   - ✅ Server implementations exist (exarp-go, devwisdom-go) for testing
   - ✅ Server infrastructure is complete and well-designed
   - ❌ Client infrastructure not needed (existing libraries available)

3. **Recommendation:**
   - **DO NOT** reimplement client functionality (already exists)
   - **DO** ensure protocol type compatibility with existing clients
   - **DO** document integration with existing client libraries
   - **DO** consider lightweight testing utilities if needed
   - **DO** create examples showing integration patterns

**Recommended Next Steps:**

1. ✅ Research existing Go MCP client libraries (completed)
2. ✅ Design client wrapper approach (see `docs/CLIENT_WRAPPER_DESIGN.md`)
3. **Implement client wrapper** (optional, recommended)
4. Test `mcp-go-core` servers with existing client libraries
5. Document compatibility and integration patterns
6. Create example integrations with `mcp-golang` and `mcp-go`
7. Add links to client libraries in README and documentation

**Strategic Approach:**

Rather than building a full client implementation, `mcp-go-core` should:
- **Server infrastructure** (current strength) ✅
- **Protocol compatibility** with existing clients ✅
- **Client wrapper utilities** (NEW - recommended)
  - Wrap existing client libraries (e.g., `mcp-golang`)
  - Convert types to `mcp-go-core` types
  - Provide testing utilities for `mcp-go-core` servers
  - Keep dependency optional
- **Integration documentation** and examples
- **Testing utilities** via wrapper package

**Wrapper Approach Benefits:**
- ✅ No duplication - reuses existing, well-tested libraries
- ✅ Type integration - uses `mcp-go-core` types throughout
- ✅ Testing support - utilities for testing `mcp-go-core` servers
- ✅ Optional dependency - only needed if wrapper is used
- ✅ Maintainable - changes to underlying library don't require full reimplementation

See `docs/CLIENT_WRAPPER_DESIGN.md` for detailed design.
