# Client Wrapper Integration Notes

## Dependency Installation

To use the client wrapper, install the external dependency:

```bash
cd /home/dlowes/projects/mcp-go-core
go get github.com/metoro-io/mcp-golang
```

This will add the dependency to `go.mod`.

## Build Tags

The implementation uses build tags to make the dependency optional:

- **Normal build:** Includes client wrapper (requires `mcp-golang`)
- **Build without client:** `go build -tags no_mcp_client` (excludes client wrapper)

## API Verification Required

The `client_impl.go` file contains implementation patterns based on the documented API from `mcp-golang`. The actual API structure may differ. Verify:

1. **Client creation:**
   ```go
   transport := stdio.NewStdioClientTransport()
   client := mcp.NewClient(transport)
   ```

2. **Initialize method signature:**
   - Parameters expected by `client.Initialize()`
   - Return type structure

3. **Method signatures:**
   - `ListTools(ctx, cursor)` - cursor type and usage
   - `CallTool(ctx, name, args)` - args format
   - `ReadResource(ctx, uri)` - return structure
   - `ListPrompts(ctx, cursor)` - return structure
   - `GetPrompt(ctx, name, args)` - return structure

4. **Type structures:**
   - Tool structure fields
   - TextContent structure
   - Resource structure
   - Prompt structure

## Testing

### Unit Tests (No External Dependency)

```bash
# Test conversion utilities (works without mcp-golang)
go test ./pkg/mcp/client -run TestConvert -v
```

### Integration Tests (Requires mcp-golang)

```bash
# Full test suite (requires mcp-golang)
go test ./pkg/mcp/client -v

# Integration tests only
go test ./pkg/mcp/client -run TestIntegration -v
```

## Implementation Status

- ✅ Package structure complete
- ✅ Type conversion utilities implemented and tested
- ⚠️ Client implementation structure ready (needs API verification)
- ⚠️ Integration tests structure ready (needs API verification)

## Next Steps

1. Install dependency: `go get github.com/metoro-io/mcp-golang`
2. Verify API by checking actual library types and methods
3. Adjust `client_impl.go` based on actual API
4. Run unit tests: `go test ./pkg/mcp/client -run TestConvert`
5. Create and run integration tests
6. Test with real MCP servers (exarp-go, devwisdom-go)
