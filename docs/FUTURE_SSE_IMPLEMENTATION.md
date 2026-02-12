# Future SSE (Server-Sent Events) Implementation

**Status:** Deferred for Future Implementation  
**Created:** 2026-02-02  
**Reason:** Current focus is on STDIO transport; SSE/HTTP support deferred until needed

## Nice to have (stdio-irrelevant)

The following work is **not required for STDIO MCP** (e.g. Cursor). It is documented here as **nice to have** for future HTTP/SSE or web deployment. These items are **not tracked in Todo2**; they are deferred until non-stdio transport is needed.

| Ref / former task | Description |
|-------------------|-------------|
| **T-145** | Use `url` and `transport` in `mcp.json` — Cursor uses `command` + `args` only; url/transport apply to HTTP clients. |
| **T-143** | Set up HTTP server with SSE endpoint — HTTP server + SSE endpoint for MCP over the web. |
| **T-1768253992311** | Complete SSE Transport Implementation — Full SSE transport (connection handling, JSON-RPC over SSE, adapter updates). |
| **T-231** | Implement load balancing — Requires HTTP/SSE first. |
| **T-147** | Support multiple concurrent connections — Only relevant when server is exposed over HTTP/SSE. |
| **T-276** | Add per-tool limits — STDIO has a single client per process; per-tool limits matter when multiple HTTP/SSE clients connect. Deleted from Todo2; see `docs/RATE_LIMITING_TASKS.md`. |
| **T-290** | Test per-tool limits — Same as T-276; tests deferred feature. Deleted from Todo2 2026-02-11. |

Optional / lower priority for stdio-only: **T-144** (Configure OAuth), **T-146** (Handle authentication tokens), **T-278** (Design access control model) — these matter when the transport has client identity (e.g. HTTP with tokens). T-144 and T-146 are documented in `docs/FUTURE_AUTHENTICATION_IMPLEMENTATION.md` and removed from Todo2 (2026-02-11).

## Overview

This document captures the planned implementation of Server-Sent Events (SSE) transport for HTTP-based MCP server deployment. Currently, exarp-go uses STDIO transport exclusively (as required by Cursor IDE). SSE support would enable web-based and HTTP deployment scenarios.

## Deferred Tasks

### T-1768253992311: Complete SSE Transport Implementation

**Priority:** Medium  
**Status:** Deferred  
**Tags:** #bug, #feature, #mcp, #testing

**Description:**
Complete the Server-Sent Events (SSE) transport implementation for HTTP-based MCP servers.

**Current State:**
- SSETransport exists as a placeholder in the codebase
- Transport interface is defined but SSE-specific implementation is incomplete

**Required Work:**
1. **HTTP Server Setup**
   - Implement HTTP server initialization
   - Configure SSE endpoint routes
   - Set up connection handling

2. **SSE Connection Handling**
   - Implement SSE connection establishment
   - Handle client connection lifecycle
   - Manage connection state

3. **Message Reading/Writing**
   - Implement message serialization for SSE format
   - Handle bidirectional communication over SSE
   - Support JSON-RPC 2.0 protocol over SSE

4. **Connection Management**
   - Connection pooling
   - Reconnection handling
   - Timeout management
   - Graceful shutdown

5. **Error Handling**
   - Connection errors
   - Protocol errors
   - Timeout handling
   - Error recovery

6. **Testing**
   - Unit tests for SSE transport
   - Integration tests with HTTP clients
   - Load testing for concurrent connections
   - Error scenario testing

7. **Adapter Updates**
   - Update framework adapters to support SSE transport
   - Ensure transport selection works correctly
   - Test with multiple transport types

**Dependencies:**
- T-1768318471622 (Transport Interface Implementation) - **COMPLETED**

**Impact:**
- Enables HTTP-based MCP server deployment
- Allows web browser clients to connect
- Supports cloud/containerized deployments
- Enables load balancing and scaling

---

### T-143: Set up HTTP server with SSE endpoint

**Priority:** Medium  
**Status:** Deferred  
**Tags:** #comparison, #discovered, #markdown, #mcp-cpp

**Description:**
Set up HTTP server with SSE endpoint for MCP protocol communication.

**Required Work:**
1. **HTTP Server Configuration**
   - Choose HTTP server framework (net/http, gin, echo, etc.)
   - Configure server ports and endpoints
   - Set up middleware (logging, CORS, etc.)

2. **SSE Endpoint Implementation**
   - `/sse` or `/mcp` endpoint for SSE connections
   - Proper SSE headers and content-type
   - Keep-alive mechanism

3. **MCP Protocol Integration**
   - Map JSON-RPC 2.0 messages to SSE events
   - Handle request/response correlation
   - Support notifications and streaming

4. **Security**
   - Authentication/authorization
   - Rate limiting
   - CORS configuration
   - TLS/HTTPS support

**Discovered From:**
- `docs/MCP_FRAMEWORKS_COMPARISON.md`
- Analysis of MCP C++ implementation patterns

---

### T-231: Implement load balancing

**Priority:** Medium  
**Status:** Deferred  
**Tags:** #discovered, #feature, #planned

**Description:**
Implement load balancing for distributed MCP server deployment.

**Why SSE/HTTP Dependent:**
Load balancing requires HTTP endpoints to distribute traffic across multiple server instances. With STDIO transport (current implementation), each process is independent and there's no network-level load balancing possible.

**Required Work:**
1. **HTTP Transport Prerequisites**
   - SSE/HTTP transport must be implemented first
   - Server must expose HTTP endpoints for load balancer health checks
   - Connection affinity may be needed for stateful sessions

2. **Load Balancer Integration**
   - Health check endpoints (`/health`, `/ready`)
   - Graceful connection draining
   - Session stickiness (if needed)
   - Metrics for load balancer monitoring

3. **Multi-Instance Support**
   - Shared state management (if any)
   - Request routing
   - Failover handling

**Discovered From:**
- `docs/MULTI_AGENT_PLAN.md`

**Prerequisites:**
- T-1768253992311: Complete SSE Transport Implementation
- T-143: Set up HTTP server with SSE endpoint

---

### T-147: Support multiple concurrent connections

**Priority:** Medium  
**Status:** Deferred  
**Tags:** #comparison, #concurrency, #discovered, #markdown, #mcp-cpp

**Description:**
Support multiple concurrent connections to the MCP server.

**Why SSE/HTTP Dependent:**
With STDIO transport (Cursor IDE), there is a single client connection per process. Multiple concurrent connections only apply when the server is exposed over HTTP/SSE and multiple clients (browsers, agents, or other tools) connect to the same server instance.

**Required Work (when SSE is implemented):**
- Connection lifecycle per client
- Per-connection request/response correlation
- Connection limits and backpressure
- Graceful handling of disconnects

**Discovered From:**
- `docs/MCP_FRAMEWORKS_COMPARISON.md`

**Prerequisites:**
- T-1768253992311: Complete SSE Transport Implementation
- T-143: Set up HTTP server with SSE endpoint

---

### Other SSE/HTTP-related tasks

These tasks are part of the SSE/HTTP scope (deferred until SSE transport is implemented):

| Task ID | Name | Why SSE/HTTP |
|---------|------|---------------|
| **T-144** | Configure OAuth if needed | OAuth/auth flows apply when the server is exposed over HTTP/SSE and clients authenticate. |
| **T-276** | Add per-tool limits | Per-tool rate/limits matter when multiple clients connect over HTTP/SSE; STDIO has a single client per process. |

---

## Technical Design (Preliminary)

### Transport Interface

The Transport interface is already defined in `pkg/mcp/framework/transport.go`:

```go
type Transport interface {
    Start() error
    Stop() error
    Type() string
}
```

### SSE Transport Implementation

```go
type SSETransport struct {
    server   *http.Server
    port     int
    endpoint string
    clients  map[string]*SSEClient
    mu       sync.RWMutex
}

type SSEClient struct {
    id       string
    writer   http.ResponseWriter
    flusher  http.Flusher
    done     chan struct{}
}

func (t *SSETransport) Start() error {
    // Initialize HTTP server
    // Set up SSE endpoint handler
    // Start listening
}

func (t *SSETransport) Stop() error {
    // Gracefully close all connections
    // Shutdown HTTP server
}

func (t *SSETransport) Type() string {
    return "sse"
}
```

### HTTP Server Setup

```go
func setupSSEServer(port int) *http.Server {
    mux := http.NewServeMux()
    
    // SSE endpoint
    mux.HandleFunc("/mcp", handleSSEConnection)
    
    // Health check
    mux.HandleFunc("/health", handleHealth)
    
    return &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
}

func handleSSEConnection(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // Create client connection
    // Handle messages
    // Manage lifecycle
}
```

### Message Format

SSE events for MCP JSON-RPC 2.0:

```
event: message
data: {"jsonrpc":"2.0","method":"tools/list","id":1}

event: message
data: {"jsonrpc":"2.0","result":{"tools":[...]},"id":1}
```

## Implementation Phases

### Phase 1: Basic SSE Transport (2-3 days)
- [ ] Implement SSETransport struct and methods
- [ ] Basic HTTP server setup
- [ ] SSE connection handling
- [ ] Simple message send/receive

### Phase 2: Protocol Integration (2-3 days)
- [ ] JSON-RPC 2.0 over SSE
- [ ] Request/response correlation
- [ ] Error handling
- [ ] Connection management

### Phase 3: Production Features (3-5 days)
- [ ] Authentication/authorization
- [ ] Rate limiting
- [ ] Connection pooling
- [ ] Graceful shutdown
- [ ] Metrics and monitoring

### Phase 4: Testing & Documentation (2-3 days)
- [ ] Unit tests
- [ ] Integration tests
- [ ] Load testing
- [ ] Documentation
- [ ] Example implementations

**Total Estimated Effort:** 9-14 days

## Current Workaround

**For now, exarp-go uses STDIO transport exclusively:**
- ✅ STDIO transport is fully implemented and tested
- ✅ Works with Cursor IDE (primary use case)
- ✅ Simpler deployment model
- ✅ No HTTP server overhead
- ✅ No security concerns with HTTP endpoints

**When SSE becomes needed:**
- Web-based MCP clients
- Browser extensions
- Cloud deployments
- Load-balanced scenarios
- Multi-client support

## References

### Code Locations
- Transport interface: `pkg/mcp/framework/transport.go`
- STDIO transport: `pkg/mcp/framework/transport.go` (StdioTransport)
- Adapter integration: `pkg/mcp/framework/adapters/gosdk/adapter.go`

### Related Documentation
- `docs/MCP_FRAMEWORKS_COMPARISON.md` - Framework comparison including SSE support
- MCP Specification: https://spec.modelcontextprotocol.io/specification/basic/transports/

### Similar Implementations
- MCP TypeScript SDK: SSE transport reference implementation
- MCP Python SDK: HTTP/SSE server examples

## Decision Log

**2026-02-02:** Deferred SSE implementation
- **Reason:** Current focus on STDIO transport for Cursor IDE integration
- **Impact:** No immediate impact; STDIO covers current use cases
- **Future Trigger:** When web-based or HTTP deployment is needed
- **Effort:** Estimated 9-14 days when prioritized

## Notes

- SSE is one-way (server → client), but can be combined with HTTP POST for bidirectional
- Alternative: WebSockets (more complex but truly bidirectional)
- Consider SSE for notifications, WebSockets for full duplex if needed
- HTTP/2 improves SSE performance with multiplexing

---

**Status:** This document serves as a reference for future SSE implementation. Tasks T-1768253992311, T-143, T-231, and T-147 have been removed from the active backlog and documented here for future reference.
