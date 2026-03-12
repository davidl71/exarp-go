# Client Wrapper Testing Guide

## Test Structure

The client wrapper package includes several types of tests:

1. **Unit Tests** (`convert_test.go`) - Test conversion utilities (no external dependencies)
2. **Integration Tests** (`client_integration_test.go`) - Test with real MCP servers

## Running Tests

### Unit Tests (No External Dependencies)

These tests work without the `mcp-golang` library:

```bash
cd /home/dlowes/projects/mcp-go-core

# Run all conversion utility tests
go test ./pkg/mcp/client -run TestConvert -v

# Run specific test
go test ./pkg/mcp/client -run TestConvertExternalToolToToolInfo -v
```

### Integration Tests (Requires mcp-golang)

Integration tests require:
- `github.com/metoro-io/mcp-golang` installed
- An MCP server to test against

```bash
# Set the server to test against
export MCP_TEST_SERVER="/path/to/your/mcp/server"

# Run integration tests
go test -tags integration ./pkg/mcp/client -v

# Run specific integration test
go test -tags integration ./pkg/mcp/client -run TestClientInitialization -v
```

### Example: Testing with exarp-go

```bash
# Build exarp-go server first (in exarp-go directory)
cd /home/dlowes/projects/exarp-go
make build

# Set test server
export MCP_TEST_SERVER="/home/dlowes/projects/exarp-go/bin/exarp-go"

# Run integration tests
cd /home/dlowes/projects/mcp-go-core
go test -tags integration ./pkg/mcp/client -v
```

### Example: Testing with devwisdom-go

```bash
# Build devwisdom server first (in devwisdom-go directory)
cd /home/dlowes/projects/devwisdom-go
make build

# Set test server
export MCP_TEST_SERVER="/home/dlowes/projects/devwisdom-go/devwisdom"

# Run integration tests
cd /home/dlowes/projects/mcp-go-core
go test -tags integration ./pkg/mcp/client -v
```

## Test Coverage

### Unit Tests (`convert_test.go`)

- ✅ `TestConvertExternalToolToToolInfo` - Tool conversion with various scenarios
- ✅ `TestConvertExternalTextContent` - Text content conversion
- ✅ `TestConvertExternalTextContentSlice` - Slice conversion
- ✅ `TestConvertClientInfoToExternal` - Client info conversion

### Integration Tests (`client_integration_test.go`)

- ⚠️ `TestClientInitialization` - Basic client initialization
- ⚠️ `TestListTools` - Tool listing
- ⚠️ `TestCallTool` - Tool execution
- ⚠️ `TestListResources` - Resource listing
- ⚠️ `TestTestServerCapabilities` - Capability testing utility
- ⚠️ `TestAssertToolExists` - Tool assertion utility
- ⚠️ `TestToolExecution` - Tool execution utility

⚠️ = Requires external dependency and MCP server

## Test Environment Variables

- `MCP_TEST_SERVER` - Path to MCP server binary (required for integration tests)
- `MCP_TEST_SERVER_ARGS` - Optional arguments for the server (space-separated)

## Continuous Integration

For CI/CD, integration tests can be:

1. **Skipped by default** (use build tags)
2. **Run conditionally** (set environment variable)
3. **Run in separate job** (integration test job)

Example CI configuration:
```yaml
# Run unit tests (always)
- run: go test ./pkg/mcp/client -run TestConvert

# Run integration tests (if server available)
- run: |
    if [ -n "$MCP_TEST_SERVER" ]; then
      go test -tags integration ./pkg/mcp/client
    fi
```

## Troubleshooting

### "package github.com/metoro-io/mcp-golang: cannot find package"

**Solution:** Install the dependency:
```bash
go get github.com/metoro-io/mcp-golang
```

### "Skipping integration test: MCP_TEST_SERVER not set"

**Solution:** Set the environment variable:
```bash
export MCP_TEST_SERVER="/path/to/server"
```

### Tests fail with "connection refused" or similar

**Solution:** 
- Ensure the server binary exists and is executable
- Check the server path is correct
- Verify the server starts correctly when run directly

### Build errors in client_impl.go

**Solution:**
- Verify `mcp-golang` API matches the implementation
- Check that the external library types match expected structures
- Update `client_impl.go` to match actual API

## Next Steps

1. Install `mcp-golang`: `go get github.com/metoro-io/mcp-golang`
2. Verify API compatibility
3. Run unit tests: `go test ./pkg/mcp/client -run TestConvert`
4. Set up test server and run integration tests
5. Fix any API mismatches in `client_impl.go`
6. Add more test cases as needed
