# Client Wrapper API Verification Results

**Date:** 2025-01-27  
**Package:** `pkg/mcp/client`  
**Status:** ❌ API Verification Failed - Fundamental Limitation Discovered

## Executive Summary

API verification revealed that **`mcp-golang` does not support stdio-based clients**, which is a fundamental limitation for our primary use case (calling stdio-based MCP servers like `devwisdom-go`). This makes the wrapper approach impractical without significant additional work.

## Test Results

### ✅ Build with `no_mcp_client` tag: SUCCESS
- `client_stub.go` works correctly
- Build tag structure is correct
- Stub implementations function as expected

### ❌ Build with `mcp-golang`: FAILED
- API mismatches discovered
- Implementation structure is correct
- API assumptions don't match actual library

## API Issues Discovered

### 1. stdio Client Transport ❌ Critical Issue

**Problem**: `stdio.NewStdioClientTransport()` does not exist

**Finding**:
- `stdio` package only provides `NewStdioServerTransport()` (server-side)
- `NewClient` requires `transport.Transport` interface
- No client-side stdio transport implementation available

**Impact**: 
- **Cannot use mcp-golang for stdio-based clients**
- This is a fundamental limitation for our use case (calling stdio-based MCP servers)

**Root Cause**:
`mcp-golang` is designed primarily for server implementations. The stdio transport package (`github.com/metoro-io/mcp-golang/transport/stdio`) only provides server transports. There is no client-side stdio transport available.

**Options**:
1. **Implement custom stdio client transport** (implements `transport.Transport` interface)
   - Requires implementing full `Transport` interface
   - Significant additional work
   - May not be worth it if alternatives exist

2. **Use different library** that supports stdio clients
   - Need to research alternatives
   - May not exist or may have other limitations

3. **Implement direct JSON-RPC client** for stdio (bypass mcp-golang)
   - More work but complete control
   - Bypasses mcp-golang entirely
   - Direct implementation of JSON-RPC 2.0

4. **Document limitation** and note that mcp-golang is not suitable for stdio clients
   - Keep wrapper structure for potential future use (HTTP clients)
   - Consider alternative approaches

### 2. ResourceResponse Structure ❌ Fixable (but blocked)

**Problem**: 
- `resource.Content` undefined (should be `resource.Contents`)
- `resource.MimeType` undefined (not a direct field)

**Finding**:
- `ResourceResponse` has `Contents []*EmbeddedResource` field
- `EmbeddedResource` contains the actual content and mimeType
- Need to access `Contents[0]` to get the resource data

**Actual Structure**:
```go
type ResourceResponse struct {
    Contents []*EmbeddedResource `json:"contents"`
}

type EmbeddedResource struct {
    EmbeddedResourceType embeddedResourceType
    TextResourceContents *TextResourceContents
    BlobResourceContents *BlobResourceContents
}
```

**Fix Required**: 
- Update `ReadResource` to use `Contents` array
- Extract content and mimeType from `EmbeddedResource`
- Handle both `TextResourceContents` and `BlobResourceContents`

**Note**: This fix is blocked by the stdio transport issue (cannot test without working transport)

### 3. Resource Field Pointers ⚠️ Fixable (but blocked)

**Problem**: 
- `resource.Description` is `*string`, not `string`
- `resource.MimeType` is `*string`, not `string`

**Fix Required**: 
- Dereference pointers with nil checks
- Use `*resource.Description` instead of `resource.Description`
- Handle nil pointers appropriately

**Example Fix**:
```go
description := ""
if resource.Description != nil {
    description = *resource.Description
}
```

**Note**: This fix is blocked by the stdio transport issue (cannot test without working transport)

## Recommendation

Given the stdio client transport limitation:

### Option A: Document Limitation and Consider Alternatives ⭐ Recommended

**Approach**: Document that mcp-golang is not suitable for stdio clients

**Pros**:
- Clear documentation of limitation
- Avoids wasted effort on incompatible approach
- Allows focus on alternative solutions

**Cons**:
- Abandons wrapper approach for stdio clients
- Need alternative solution for MCP client functionality

**Action Items**:
1. Document limitation clearly
2. Consider alternative approaches:
   - Direct JSON-RPC 2.0 client implementation
   - Python bridge for client functionality (current approach)
   - Research other Go MCP client libraries
3. Keep wrapper structure for potential future use (HTTP clients)

### Option B: Implement Custom stdio Client Transport

**Approach**: Implement `transport.Transport` interface for stdio clients

**Pros**:
- Enables use of mcp-golang for stdio clients
- Maintains wrapper approach

**Cons**:
- Requires implementing full `Transport` interface:
  - `Start(ctx context.Context) error`
  - `Send(ctx context.Context, message *BaseJsonRpcMessage) error`
  - `Close() error`
  - `SetCloseHandler(handler func())`
  - `SetErrorHandler(handler func(error))`
  - `SetMessageHandler(handler func(ctx context.Context, message *BaseJsonRpcMessage))`
- Significant additional work
- Need to handle stdio communication patterns
- May not be worth it if alternatives exist

**Estimated Effort**: High (2-3 days minimum)

### Option C: Use Direct JSON-RPC Implementation

**Approach**: Bypass mcp-golang entirely, implement JSON-RPC 2.0 client directly

**Pros**:
- Complete control over implementation
- No external dependency issues
- Can optimize for stdio communication
- Full understanding of codebase

**Cons**:
- More work to implement
- Need to handle JSON-RPC 2.0 protocol
- Need to handle MCP protocol specifics

**Estimated Effort**: Medium-High (1-2 days for basic implementation)

## Conclusion

The test successfully identified that **mcp-golang is not suitable for stdio-based clients** as it lacks stdio client transport support. This is a fundamental limitation that makes the wrapper approach impractical for our primary use case (calling stdio-based MCP servers like devwisdom-go).

**Recommended Action**: Document the limitation clearly and consider alternative approaches for MCP client functionality. The wrapper structure can be kept for potential future use with HTTP clients, but stdio client functionality requires a different approach.

## Next Steps

1. ✅ Document limitation (this document)
2. Consider alternative approaches:
   - Direct JSON-RPC 2.0 client implementation
   - Continue using Python bridge for client functionality
   - Research other Go MCP client libraries
3. Update `CLIENT_WRAPPER_LIMITATIONS.md` with this finding
4. Update integration plans to reflect this limitation

## Related Documents

- [Client Wrapper Limitations](../pkg/mcp/client/CLIENT_WRAPPER_LIMITATIONS.md)
- [Client Wrapper Usage](../pkg/mcp/client/CLIENT_WRAPPER_USAGE.md)
- [Integration Notes](../pkg/mcp/client/INTEGRATION_NOTES.md)
