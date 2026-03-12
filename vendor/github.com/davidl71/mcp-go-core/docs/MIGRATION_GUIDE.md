# 🔄 Migration Guide: exarp-go/devwisdom-go to mcp-go-core

This guide will help you migrate from the internal MCP packages in `exarp-go` or `devwisdom-go` to the shared `mcp-go-core` library.

**Last Updated:** 2026-01-12  
**mcp-go-core Version:** v0.2.0

---

## Overview

The `mcp-go-core` library extracts common MCP functionality from `exarp-go` and `devwisdom-go` into a shared, reusable package. This migration eliminates code duplication and provides a single source of truth for MCP infrastructure.

---

## Migration Steps

### Step 1: Add Dependency

Add `mcp-go-core` to your `go.mod`:

```bash
go get github.com/davidl71/mcp-go-core@latest
```

Or pin to a specific version:

```bash
go get github.com/davidl71/mcp-go-core@v0.2.0
```

### Step 2: Update Imports

#### For exarp-go

**Before:**
```go
import (
    "github.com/davidl71/exarp-go/internal/framework"
    "github.com/davidl71/exarp-go/internal/security"
    "github.com/davidl71/exarp-go/internal/factory"
)
```

**After:**
```go
import (
    "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
    "github.com/davidl71/mcp-go-core/pkg/mcp/security"
    "github.com/davidl71/mcp-go-core/pkg/mcp/factory"
)
```

#### For devwisdom-go

**Before:**
```go
import (
    "github.com/davidl71/devwisdom-go/internal/mcp"
    "github.com/davidl71/devwisdom-go/internal/logging"
)
```

**After:**
```go
import (
    "github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
    "github.com/davidl71/mcp-go-core/pkg/mcp/logging"
)
```

### Step 3: Update Package References

#### Framework Interface

**Before:**
```go
server, err := factory.NewServer(framework.FrameworkGoSDK, "my-server", "1.0.0")
```

**After:**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/config"
import "github.com/davidl71/mcp-go-core/pkg/mcp/factory"

server, err := factory.NewServer(config.FrameworkGoSDK, "my-server", "1.0.0")
```

#### Security Utilities

**Before:**
```go
import "github.com/davidl71/exarp-go/internal/security"

projectRoot, err := security.GetProjectRoot(".")
validatedPath, err := security.ValidatePath(path, projectRoot)
```

**After:**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/security"

projectRoot, err := security.GetProjectRoot(".")
validatedPath, err := security.ValidatePath(path, projectRoot)
```

**Note:** Function signatures are identical, so no code changes needed beyond imports.

#### Configuration

**Before:**
```go
import "github.com/davidl71/exarp-go/internal/config"

cfg, err := config.Load()
server, err := factory.NewServerFromConfig(cfg)
```

**After:**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/config"
import "github.com/davidl71/mcp-go-core/pkg/mcp/factory"

cfg, err := config.LoadBaseConfig()
// Your custom config can embed BaseConfig:
type MyConfig struct {
    config.BaseConfig
    // Your additional fields
}

server, err := factory.NewServerFromConfig(cfg)
```

#### Logging

**Before:**
```go
import "github.com/davidl71/devwisdom-go/internal/logging"

logger := logging.NewLogger()
logger.Info("message", "Hello, world!")
```

**After:**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/logging"

logger := logging.NewLogger()
logger.Infof("Hello, world!")
// Or with context:
logger.Infof("Tool call: %s", toolName)
```

**Note:** Logging API is slightly different - uses `Infof`, `Debugf`, `Errorf` instead of `Info`, `Debug`, `Error`.

#### Protocol Types

**Before:**
```go
import "github.com/davidl71/devwisdom-go/internal/mcp"

request := &mcp.JSONRPCRequest{
    JSONRPC: "2.0",
    Method:  "tools/call",
    Params:  params,
}
```

**After:**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/protocol"

request := &protocol.JSONRPCRequest{
    JSONRPC: "2.0",
    Method:  "tools/call",
    Params:  params,
}
```

### Step 4: Update Adapter Usage

**Before:**
```go
adapter := gosdk.NewGoSDKAdapter("my-server", "1.0.0")
```

**After:**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/logging"
import "github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"

logger := logging.NewLogger()
adapter := gosdk.NewGoSDKAdapter("my-server", "1.0.0",
    gosdk.WithLogger(logger),
    // Optional: gosdk.WithMiddleware(myMiddleware),
)
```

### Step 5: Remove Internal Packages

After verifying everything works:

1. Remove old internal packages:
   - `internal/framework` → use `pkg/mcp/framework`
   - `internal/security` → use `pkg/mcp/security`
   - `internal/factory` → use `pkg/mcp/factory`
   - `internal/config` → use `pkg/mcp/config`
   - `internal/logging` → use `pkg/mcp/logging`
   - `internal/mcp` → use `pkg/mcp/protocol`

2. Update any remaining references

3. Run tests to verify

---

## Breaking Changes

### Minimal Breaking Changes

Most APIs are backward compatible. The main changes are:

1. **Import paths** - All imports need to change
2. **Config API** - `config.Load()` → `config.LoadBaseConfig()`
3. **Logging API** - Method names changed (e.g., `Info()` → `Infof()`)

### Migration Strategy

**Option 1: Gradual Migration (Recommended)**
- Keep internal packages as wrappers initially
- Update one package at a time
- Test after each change
- Remove internal packages after full migration

**Option 2: Complete Migration**
- Update all imports at once
- Update all code to use new APIs
- Remove internal packages
- Test everything

---

## Common Issues and Solutions

### Issue 1: Import Conflicts

**Problem:** Both old and new imports exist during migration.

**Solution:** Remove old internal packages completely after migration.

### Issue 2: Config Structure Differences

**Problem:** Your config embeds the old `Config` struct.

**Solution:**
```go
// Before
type MyConfig struct {
    config.Config
    MyField string
}

// After
type MyConfig struct {
    config.BaseConfig
    MyField string
}
```

### Issue 3: Logging API Changes

**Problem:** Old logging API uses different method signatures.

**Solution:**
```go
// Before
logger.Info("context", "message: %s", value)

// After
logger.Infof("message: %s", value)
// Context is now included in the logger instance, not per call
```

### Issue 4: Missing Functionality

**Problem:** Some internal package functionality doesn't exist in mcp-go-core.

**Solution:**
- Check if it's been moved to a different package
- Check if it's in a newer version
- Consider if it's project-specific (keep in internal)
- File an issue if it should be in mcp-go-core

---

## Verification Checklist

After migration, verify:

- [ ] All imports updated
- [ ] All tests pass
- [ ] Server starts without errors
- [ ] Tools can be registered
- [ ] Tools can be called
- [ ] Prompts work
- [ ] Resources work
- [ ] Logging works
- [ ] Security utilities work
- [ ] No build errors
- [ ] No runtime errors

---

## Rollback Plan

If issues occur:

1. **Revert go.mod:**
   ```bash
   git checkout HEAD -- go.mod go.sum
   go mod tidy
   ```

2. **Revert import changes:**
   ```bash
   git checkout HEAD -- path/to/file.go
   ```

3. **Keep internal packages:**
   - Don't delete internal packages until migration is verified

---

## Getting Help

If you encounter issues during migration:

1. Check the [mcp-go-core README](../README.md)
2. Review the [API documentation](https://pkg.go.dev/github.com/davidl71/mcp-go-core)
3. Check [existing issues](https://github.com/davidl71/mcp-go-core/issues)
4. Create a new issue with:
   - Your Go version
   - mcp-go-core version
   - Error messages
   - Minimal reproduction case

---

## Post-Migration

After successful migration:

1. **Update documentation** - Update any docs referencing internal packages
2. **Remove internal packages** - Clean up old code
3. **Update CI/CD** - Ensure tests pass with new dependencies
4. **Update dependencies** - Keep mcp-go-core updated

---

## Example: Complete Migration

Here's a complete example of migrating a server setup:

**Before:**
```go
package main

import (
    "context"
    "log"
    
    "github.com/davidl71/exarp-go/internal/config"
    "github.com/davidl71/exarp-go/internal/factory"
    "github.com/davidl71/exarp-go/internal/framework"
    "github.com/davidl71/exarp-go/internal/security"
)

func main() {
    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create server
    server, err := factory.NewServerFromConfig(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get project root
    root, err := security.GetProjectRoot(".")
    if err != nil {
        log.Fatal(err)
    }
    
    // Run server
    transport := &framework.StdioTransport{}
    if err := server.Run(context.Background(), transport); err != nil {
        log.Fatal(err)
    }
}
```

**After:**
```go
package main

import (
    "context"
    "log"
    
    "github.com/davidl71/mcp-go-core/pkg/mcp/config"
    "github.com/davidl71/mcp-go-core/pkg/mcp/factory"
    "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
    "github.com/davidl71/mcp-go-core/pkg/mcp/logging"
    "github.com/davidl71/mcp-go-core/pkg/mcp/security"
)

func main() {
    // Load config
    baseCfg, err := config.LoadBaseConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create custom config (if needed)
    cfg := &MyConfig{
        BaseConfig: *baseCfg,
        // Add your custom fields
    }
    
    // Create server
    server, err := factory.NewServerFromConfig(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get project root
    root, err := security.GetProjectRoot(".")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create logger
    logger := logging.NewLogger()
    logger.Infof("Server starting in project: %s", root)
    
    // Run server
    transport := &framework.StdioTransport{}
    if err := server.Run(context.Background(), transport); err != nil {
        log.Fatal(err)
    }
}

type MyConfig struct {
    config.BaseConfig
    // Your custom fields here
}
```

---

*Happy migrating! 🚀*
