# Go SDK Migration - Quick Start Guide

**Quick reference for starting the migration to Go SDK**

---

## Prerequisites

```bash
# Install Go (if not already installed)
brew install go  # macOS
# or download from https://go.dev/dl/

# Verify installation
go version  # Should be 1.21+

# Install Go SDK
go get github.com/modelcontextprotocol/go-sdk
```

---

## Project Setup

```bash
# Create new Go project directory
mkdir -p /Users/davidl/Projects/exarp-go
cd /Users/davidl/Projects/exarp-go

# Initialize Go module
go mod init github.com/davidl/exarp-go

# Add dependencies
go get github.com/modelcontextprotocol/go-sdk
```

---

## Project Structure

```
exarp-go/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── bridge/
│   │   └── python.go
│   ├── tools/
│   │   ├── registry.go
│   │   └── handlers.go
│   ├── prompts/
│   │   └── registry.go
│   └── resources/
│       └── handlers.go
├── bridge/
│   └── execute_tool.py
├── go.mod
├── go.sum
└── README.md
```

---

## Minimal Server Example

```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "exarp-go",
        Version: "1.0.0",
    }, nil)
    
    // Register a simple tool
    server.RegisterTool("test_tool", func(ctx context.Context, args json.RawMessage) ([]mcp.TextContent, error) {
        return []mcp.TextContent{
            {Type: "text", Text: `{"success": true, "message": "Hello from Go!"}`},
        }, nil
    })
    
    // Run with STDIO transport
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatal(err)
    }
}
```

---

## Build & Test

```bash
# Build
go build -o bin/exarp-go cmd/server/main.go

# Test locally
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./bin/exarp-go

# Update Cursor config
# Change command to: "{{PROJECT_ROOT}}/../exarp-go/bin/exarp-go"
```

---

## Python Bridge Quick Setup

```go
// internal/bridge/python.go
package bridge

import (
    "context"
    "encoding/json"
    "os/exec"
)

func ExecutePythonTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
    argsJSON, _ := json.Marshal(args)
    cmd := exec.CommandContext(ctx, "python3", "bridge/execute_tool.py", toolName, string(argsJSON))
    output, err := cmd.Output()
    return string(output), err
}
```

---

## Migration Order

1. **Week 1:** Foundation + Python bridge
2. **Week 2:** Simple tools (6 tools)
3. **Week 3:** Medium tools (8 tools)
4. **Week 4:** Advanced tools (8 tools)
5. **Week 5:** MLX tools (2 tools)
6. **Week 6:** Testing & docs

---

## Key Commands

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Lint
golangci-lint run

# Build for production
go build -ldflags="-s -w" -o bin/exarp-go cmd/server/main.go
```

---

## Troubleshooting

**Issue:** Go SDK not found
```bash
go get -u github.com/modelcontextprotocol/go-sdk
```

**Issue:** Python bridge fails
- Check Python path: `which python3`
- Verify PROJECT_ROOT environment variable
- Check Python dependencies installed

**Issue:** Cursor not connecting
- Verify binary path in `mcp.json`
- Check binary is executable: `chmod +x bin/exarp-go`
- Restart Cursor after config changes

---

## Next Steps

See [GO_SDK_MIGRATION_PLAN.md](./GO_SDK_MIGRATION_PLAN.md) for detailed migration plan.

