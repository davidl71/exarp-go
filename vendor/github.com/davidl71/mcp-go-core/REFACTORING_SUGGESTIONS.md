# 🔧 Refactoring Suggestions for mcp-go-core

**Generated:** 2026-01-12  
**Status:** Analysis Complete

---

## Executive Summary

The `mcp-go-core` library is well-structured but has several opportunities for improvement. This document outlines prioritized refactoring suggestions to enhance code quality, maintainability, and testability.

**Priority Levels:**
- 🔴 **High**: Critical for maintainability or correctness
- 🟡 **Medium**: Improves code quality and developer experience
- 🟢 **Low**: Nice-to-have improvements

---

## 🔴 High Priority Refactorings

### 1. Extract Validation Helpers

**Location:** `pkg/mcp/framework/adapters/gosdk/adapter.go`

**Problem:** Repeated validation code across `RegisterTool`, `RegisterPrompt`, and `RegisterResource` methods.

**Current Code:**
```go
// Repeated in RegisterTool, RegisterPrompt, RegisterResource
if name == "" {
    return fmt.Errorf("name cannot be empty")
}
if description == "" {
    return fmt.Errorf("description cannot be empty")
}
if handler == nil {
    return fmt.Errorf("handler cannot be nil")
}
```

**Suggested Refactoring:**
```go
// pkg/mcp/framework/adapters/gosdk/validation.go
package gosdk

import "fmt"

// ValidateRegistration validates common registration parameters
func ValidateRegistration(name, description string, handler interface{}) error {
    if name == "" {
        return fmt.Errorf("name cannot be empty")
    }
    if description == "" {
        return fmt.Errorf("description cannot be empty")
    }
    if handler == nil {
        return fmt.Errorf("handler cannot be nil")
    }
    return nil
}

// ValidateResourceRegistration validates resource-specific parameters
func ValidateResourceRegistration(uri, name, description string, handler interface{}) error {
    if err := ValidateRegistration(name, description, handler); err != nil {
        return err
    }
    if uri == "" {
        return fmt.Errorf("resource URI cannot be empty")
    }
    return nil
}
```

**Benefits:**
- ✅ Reduces code duplication
- ✅ Consistent error messages
- ✅ Easier to maintain validation logic
- ✅ Can add validation rules in one place

---

### 2. Implement Transport Interface Properly

**Location:** `pkg/mcp/framework/server.go`

**Problem:** `Transport` interface is empty and not properly used. The adapter ignores the transport parameter.

**Current Code:**
```go
// Transport abstracts transport mechanism
type Transport interface {
    // Transport-specific methods
    // Each framework will implement this differently
}
```

**Suggested Refactoring:**
```go
// pkg/mcp/framework/transport.go
package framework

import "context"

// Transport abstracts transport mechanism for MCP servers
type Transport interface {
    // Start initializes the transport
    Start(ctx context.Context) error
    
    // Stop shuts down the transport
    Stop(ctx context.Context) error
    
    // Type returns the transport type (stdio, sse, http, etc.)
    Type() string
}

// StdioTransport represents standard I/O transport
type StdioTransport struct{}

func (t *StdioTransport) Start(ctx context.Context) error { return nil }
func (t *StdioTransport) Stop(ctx context.Context) error  { return nil }
func (t *StdioTransport) Type() string                    { return "stdio" }

// SSETransport represents Server-Sent Events transport
type SSETransport struct{}

func (t *SSETransport) Start(ctx context.Context) error { return nil }
func (t *SSETransport) Stop(ctx context.Context) error  { return nil }
func (t *SSETransport) Type() string                     { return "sse" }
```

**Update Adapter:**
```go
// In adapter.go
func (a *GoSDKAdapter) Run(ctx context.Context, transport framework.Transport) error {
    if transport == nil {
        transport = &framework.StdioTransport{} // Default
    }
    
    // Use transport type to determine implementation
    switch transport.Type() {
    case "stdio":
        stdioTransport := &mcp.StdioTransport{}
        return a.server.Run(ctx, stdioTransport)
    default:
        return fmt.Errorf("unsupported transport type: %s", transport.Type())
    }
}
```

**Benefits:**
- ✅ Proper abstraction of transport mechanism
- ✅ Enables future transport types (SSE, HTTP, etc.)
- ✅ Better testability with mock transports
- ✅ Clearer API contract

---

### 3. Extract Type Conversion Helpers

**Location:** `pkg/mcp/framework/adapters/gosdk/adapter.go`

**Problem:** Type conversion logic is embedded in handler functions, making them harder to read and test.

**Current Code:**
```go
// In RegisterTool - embedded conversion logic
contents := make([]mcp.Content, len(result))
for i, content := range result {
    contents[i] = &mcp.TextContent{
        Text: content.Text,
    }
}
```

**Suggested Refactoring:**
```go
// pkg/mcp/framework/adapters/gosdk/converters.go
package gosdk

import (
    "github.com/davidl71/mcp-go-core/pkg/mcp/types"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TextContentToMCP converts framework TextContent to MCP Content
func TextContentToMCP(contents []types.TextContent) []mcp.Content {
    mcpContents := make([]mcp.Content, len(contents))
    for i, content := range contents {
        mcpContents[i] = &mcp.TextContent{
            Text: content.Text,
        }
    }
    return mcpContents
}

// ToolSchemaToMCP converts framework ToolSchema to MCP input schema
func ToolSchemaToMCP(schema types.ToolSchema) map[string]interface{} {
    inputSchema := map[string]interface{}{
        "type":       schema.Type,
        "properties": schema.Properties,
    }
    if len(schema.Required) > 0 {
        inputSchema["required"] = schema.Required
    }
    return inputSchema
}
```

**Benefits:**
- ✅ Cleaner handler functions
- ✅ Reusable conversion logic
- ✅ Easier to test conversions independently
- ✅ Can add validation during conversion

---

### 4. Add Tests for Factory and Config Packages

**Location:** `pkg/mcp/factory/` and `pkg/mcp/config/`

**Problem:** Zero test coverage for factory and config packages.

**Suggested Tests:**

```go
// pkg/mcp/factory/server_test.go
package factory

import (
    "testing"
    "github.com/davidl71/mcp-go-core/pkg/mcp/config"
)

func TestNewServer(t *testing.T) {
    tests := []struct {
        name          string
        frameworkType config.FrameworkType
        wantErr       bool
    }{
        {
            name:          "valid go-sdk",
            frameworkType: config.FrameworkGoSDK,
            wantErr:       false,
        },
        {
            name:          "invalid framework",
            frameworkType: config.FrameworkType("invalid"),
            wantErr:       true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server, err := NewServer(tt.frameworkType, "test-server", "1.0.0")
            if (err != nil) != tt.wantErr {
                t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && server == nil {
                t.Error("NewServer() returned nil server")
            }
        })
    }
}

func TestNewServerFromConfig(t *testing.T) {
    cfg := &config.BaseConfig{
        Framework: config.FrameworkGoSDK,
        Name:      "test-server",
        Version:   "1.0.0",
    }
    
    server, err := NewServerFromConfig(cfg)
    if err != nil {
        t.Fatalf("NewServerFromConfig() error = %v", err)
    }
    if server == nil {
        t.Fatal("NewServerFromConfig() returned nil server")
    }
    if server.GetName() != "test-server" {
        t.Errorf("server.GetName() = %q, want %q", server.GetName(), "test-server")
    }
}
```

```go
// pkg/mcp/config/base_test.go
package config

import (
    "os"
    "testing"
)

func TestLoadBaseConfig(t *testing.T) {
    // Save original env
    originalFramework := os.Getenv("MCP_FRAMEWORK")
    originalName := os.Getenv("MCP_SERVER_NAME")
    originalVersion := os.Getenv("MCP_VERSION")
    defer func() {
        os.Setenv("MCP_FRAMEWORK", originalFramework)
        os.Setenv("MCP_SERVER_NAME", originalName)
        os.Setenv("MCP_VERSION", originalVersion)
    }()
    
    // Test defaults
    os.Unsetenv("MCP_FRAMEWORK")
    os.Unsetenv("MCP_SERVER_NAME")
    os.Unsetenv("MCP_VERSION")
    
    cfg, err := LoadBaseConfig()
    if err != nil {
        t.Fatalf("LoadBaseConfig() error = %v", err)
    }
    if cfg.Framework != FrameworkGoSDK {
        t.Errorf("LoadBaseConfig().Framework = %v, want %v", cfg.Framework, FrameworkGoSDK)
    }
    if cfg.Name != "mcp-server" {
        t.Errorf("LoadBaseConfig().Name = %q, want %q", cfg.Name, "mcp-server")
    }
    
    // Test environment override
    os.Setenv("MCP_FRAMEWORK", "go-sdk")
    os.Setenv("MCP_SERVER_NAME", "custom-server")
    os.Setenv("MCP_VERSION", "2.0.0")
    
    cfg, err = LoadBaseConfig()
    if err != nil {
        t.Fatalf("LoadBaseConfig() error = %v", err)
    }
    if cfg.Name != "custom-server" {
        t.Errorf("LoadBaseConfig().Name = %q, want %q", cfg.Name, "custom-server")
    }
    if cfg.Version != "2.0.0" {
        t.Errorf("LoadBaseConfig().Version = %q, want %q", cfg.Version, "2.0.0")
    }
}
```

**Benefits:**
- ✅ Increases test coverage
- ✅ Validates configuration loading
- ✅ Catches regressions early

---

## 🟡 Medium Priority Refactorings

### 5. Extract Context Validation Helper

**Location:** `pkg/mcp/framework/adapters/gosdk/adapter.go`

**Problem:** Context cancellation checks are repeated in every handler.

**Suggested Refactoring:**
```go
// pkg/mcp/framework/adapters/gosdk/context.go
package gosdk

import (
    "context"
    "fmt"
)

// ValidateContext checks if context is valid and not cancelled
func ValidateContext(ctx context.Context) error {
    if ctx == nil {
        return fmt.Errorf("context cannot be nil")
    }
    if err := ctx.Err(); err != nil {
        return fmt.Errorf("context cancelled: %w", err)
    }
    return nil
}
```

**Usage:**
```go
// In handler functions
if err := ValidateContext(ctx); err != nil {
    return nil, err
}
```

**Benefits:**
- ✅ Consistent context validation
- ✅ Reduces boilerplate
- ✅ Can add context timeout checks

---

### 6. Create Error Types for Better Error Handling

**Location:** `pkg/mcp/framework/`

**Problem:** Using `fmt.Errorf` everywhere makes it hard to handle specific error types.

**Suggested Refactoring:**
```go
// pkg/mcp/framework/errors.go
package framework

import "fmt"

// ErrInvalidTool represents an invalid tool error
type ErrInvalidTool struct {
    ToolName string
    Reason   string
}

func (e *ErrInvalidTool) Error() string {
    return fmt.Sprintf("invalid tool %q: %s", e.ToolName, e.Reason)
}

// ErrToolNotFound represents a tool not found error
type ErrToolNotFound struct {
    ToolName string
}

func (e *ErrToolNotFound) Error() string {
    return fmt.Sprintf("tool %q not found", e.ToolName)
}

// IsToolNotFound checks if error is ErrToolNotFound
func IsToolNotFound(err error) bool {
    _, ok := err.(*ErrToolNotFound)
    return ok
}
```

**Benefits:**
- ✅ Type-safe error handling
- ✅ Can check error types with `errors.Is()`
- ✅ Better error messages
- ✅ Enables error wrapping

---

### 7. Add Builder Pattern for Server Configuration

**Location:** `pkg/mcp/config/`

**Problem:** Configuration is loaded from environment or defaults, but there's no programmatic way to build configs.

**Suggested Refactoring:**
```go
// pkg/mcp/config/builder.go
package config

// ConfigBuilder builds BaseConfig with fluent API
type ConfigBuilder struct {
    config *BaseConfig
}

// NewConfigBuilder creates a new config builder
func NewConfigBuilder() *ConfigBuilder {
    return &ConfigBuilder{
        config: &BaseConfig{
            Framework: FrameworkGoSDK,
            Name:      "mcp-server",
            Version:   "1.0.0",
        },
    }
}

// WithFramework sets the framework type
func (b *ConfigBuilder) WithFramework(framework FrameworkType) *ConfigBuilder {
    b.config.Framework = framework
    return b
}

// WithName sets the server name
func (b *ConfigBuilder) WithName(name string) *ConfigBuilder {
    b.config.Name = name
    return b
}

// WithVersion sets the server version
func (b *ConfigBuilder) WithVersion(version string) *ConfigBuilder {
    b.config.Version = version
    return b
}

// Build returns the built configuration
func (b *ConfigBuilder) Build() *BaseConfig {
    return b.config
}
```

**Usage:**
```go
cfg := config.NewConfigBuilder().
    WithName("my-server").
    WithVersion("2.0.0").
    Build()
```

**Benefits:**
- ✅ Fluent API for configuration
- ✅ Type-safe configuration building
- ✅ Can add validation in Build()
- ✅ Better for testing

---

### 8. Extract Request Validation Helpers

**Location:** `pkg/mcp/framework/adapters/gosdk/adapter.go`

**Problem:** Request validation is repeated in each handler.

**Suggested Refactoring:**
```go
// pkg/mcp/framework/adapters/gosdk/validation.go
package gosdk

import (
    "fmt"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ValidateCallToolRequest validates a CallToolRequest
func ValidateCallToolRequest(req *mcp.CallToolRequest) error {
    if req == nil {
        return fmt.Errorf("call tool request cannot be nil")
    }
    if req.Params == nil {
        return fmt.Errorf("call tool request params cannot be nil")
    }
    return nil
}

// ValidateGetPromptRequest validates a GetPromptRequest
func ValidateGetPromptRequest(req *mcp.GetPromptRequest) error {
    if req == nil {
        return fmt.Errorf("get prompt request cannot be nil")
    }
    if req.Params == nil {
        return fmt.Errorf("get prompt request params cannot be nil")
    }
    return nil
}

// ValidateReadResourceRequest validates a ReadResourceRequest
func ValidateReadResourceRequest(req *mcp.ReadResourceRequest) error {
    if req == nil {
        return fmt.Errorf("read resource request cannot be nil")
    }
    if req.Params == nil {
        return fmt.Errorf("read resource request params cannot be nil")
    }
    if req.Params.URI == "" {
        return fmt.Errorf("resource URI cannot be empty")
    }
    return nil
}
```

**Benefits:**
- ✅ Consistent validation
- ✅ Reusable validation logic
- ✅ Easier to test
- ✅ Can add more validation rules

---

## 🟢 Low Priority Refactorings

### 9. Remove Empty Directories

**Location:** `pkg/mcp/cli/` and `pkg/mcp/platform/`

**Problem:** Empty directories exist but aren't used.

**Action:** Either:
1. Remove empty directories, or
2. Add placeholder files with package documentation

**Suggested:**
```go
// pkg/mcp/cli/cli.go
// Package cli provides CLI utilities for MCP servers.
// This package will contain TTY detection, command-line parsing, and CLI-specific helpers.
package cli

// TODO: Implement CLI utilities
```

---

### 10. Add Package-Level Documentation

**Location:** All packages

**Problem:** Some packages lack package-level documentation.

**Suggested:** Add godoc comments to all packages:

```go
// Package framework provides framework-agnostic abstractions for MCP servers.
//
// The framework package defines interfaces and adapters that allow MCP servers
// to work with different underlying frameworks (go-sdk, mcp-go, etc.) without
// changing application code.
//
// Example:
//
//     server, err := factory.NewServer(config.FrameworkGoSDK, "my-server", "1.0.0")
//     if err != nil {
//         log.Fatal(err)
//     }
//     server.RegisterTool("my_tool", "Description", schema, handler)
//     server.Run(ctx, transport)
package framework
```

---

### 11. Consider Adding Options Pattern

**Location:** `pkg/mcp/framework/adapters/gosdk/`

**Problem:** `NewGoSDKAdapter` only takes name and version. Future options would require changing the signature.

**Suggested Refactoring:**
```go
// pkg/mcp/framework/adapters/gosdk/options.go
package gosdk

// AdapterOption configures a GoSDKAdapter
type AdapterOption func(*GoSDKAdapter)

// WithLogger sets a custom logger
func WithLogger(logger interface{}) AdapterOption {
    return func(a *GoSDKAdapter) {
        // Set logger if needed
    }
}

// WithMiddleware adds middleware
func WithMiddleware(middleware interface{}) AdapterOption {
    return func(a *GoSDKAdapter) {
        // Add middleware if needed
    }
}

// Update NewGoSDKAdapter
func NewGoSDKAdapter(name, version string, opts ...AdapterOption) *GoSDKAdapter {
    adapter := &GoSDKAdapter{
        server: mcp.NewServer(&mcp.Implementation{
            Name:    name,
            Version: version,
        }, nil),
        name:         name,
        toolHandlers: make(map[string]framework.ToolHandler),
        toolInfo:     make(map[string]types.ToolInfo),
    }
    
    for _, opt := range opts {
        opt(adapter)
    }
    
    return adapter
}
```

**Benefits:**
- ✅ Extensible without breaking changes
- ✅ Clean API for optional parameters
- ✅ Can add features incrementally

---

## Implementation Priority

### Phase 1 (Immediate)
1. ✅ Extract validation helpers (#1)
2. ✅ Add tests for factory and config (#4)
3. ✅ Extract type conversion helpers (#3)

### Phase 2 (This Sprint)
4. ✅ Implement Transport interface properly (#2)
5. ✅ Extract context validation helper (#5)
6. ✅ Create error types (#6)

### Phase 3 (Next Sprint)
7. ✅ Add builder pattern for config (#7)
8. ✅ Extract request validation helpers (#8)
9. ✅ Add package documentation (#10)

### Phase 4 (Future)
10. ✅ Remove empty directories or add placeholders (#9)
11. ✅ Consider options pattern (#11)

---

## Testing Strategy

For each refactoring:
1. ✅ Write tests first (TDD approach)
2. ✅ Ensure existing tests still pass
3. ✅ Add new tests for extracted functions
4. ✅ Verify test coverage increases
5. ✅ Update documentation

---

## Breaking Changes

**None of these refactorings introduce breaking changes** - they are all internal improvements that maintain the existing API.

---

## Metrics to Track

- **Test Coverage**: Target 80%+ overall
- **Code Duplication**: Reduce by extracting helpers
- **Cyclomatic Complexity**: Reduce by extracting functions
- **Package Dependencies**: Keep minimal

---

*Generated by analyzing mcp-go-core codebase*
