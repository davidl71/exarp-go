# MCP Client Wrapper

**Status:** ⚠️ Skeleton Implementation - Requires External Library Integration

This package provides wrapper functions around existing MCP client libraries (e.g., `github.com/metoro-io/mcp-golang`) to provide:

1. **Type Integration:** Converts between external library types and `mcp-go-core` types
2. **Testing Utilities:** Provides utilities for testing `mcp-go-core` servers
3. **Consistent API:** Provides a consistent API using `mcp-go-core` types

## Status

✅ **Package structure created** - Basic skeleton implementation  
⚠️ **External library integration pending** - Requires `github.com/metoro-io/mcp-golang`  
📋 **API design complete** - See `docs/CLIENT_WRAPPER_DESIGN.md` for design details

## Current Implementation

The package structure is in place with:
- `client.go` - Main client wrapper (skeleton implementation)
- `convert.go` - Type conversion utilities
- `testutil.go` - Testing utilities

**Note:** The actual integration with the external client library (`github.com/metoro-io/mcp-golang`) is not yet implemented. The current code provides the API structure and returns "not yet implemented" errors.

## Next Steps

1. Add `github.com/metoro-io/mcp-golang` as a dependency (optional/build tags)
2. Implement actual client initialization in `client.go`
3. Implement type conversions in `convert.go`
4. Add integration tests
5. Document usage examples

## Design Goals

- Wrap existing client libraries (no duplication)
- Use `mcp-go-core` types throughout
- Provide testing utilities
- Keep dependency optional

## Usage (Planned)

```go
import (
    "context"
    "github.com/davidl71/mcp-go-core/pkg/mcp/client"
    "github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
)

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
_, err = c.Initialize(ctx)
// ...
```

## Dependencies

This package will require:
- `github.com/metoro-io/mcp-golang` (or similar client library)

The dependency can be made optional using build tags if needed.
