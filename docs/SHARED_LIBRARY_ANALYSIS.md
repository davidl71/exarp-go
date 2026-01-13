# Shared Library Analysis: exarp-go & devwisdom-go

**Date:** 2026-01-12  
**Purpose:** Identify code that can be extracted to a shared library between exarp-go and devwisdom-go

---

## Executive Summary

Both projects are Go-based MCP servers with significant overlap in infrastructure code. A shared library would reduce duplication, improve maintainability, and ensure consistency across projects.

**Recommendation:** Create a shared library package `github.com/davidl71/mcp-go-core` containing common MCP infrastructure.

### Library Name: `mcp-go-core`

**Full module path:** `github.com/davidl71/mcp-go-core`

**Rationale:**
- ✅ Matches naming pattern (`-go` suffix like `exarp-go`, `devwisdom-go`)
- ✅ "core" clearly indicates foundational/essential code
- ✅ Clearly MCP-related
- ✅ Professional and concise
- ✅ Easy to import: `github.com/davidl71/mcp-go-core/pkg/mcp/...`

---

## 1. MCP Framework Abstraction Layer

### Current State

**exarp-go:**
- ✅ Framework-agnostic `MCPServer` interface (`internal/framework/server.go`)
- ✅ Factory pattern for server creation
- ✅ Adapter pattern for framework-specific implementations
- ✅ Transport abstraction

**devwisdom-go:**
- ✅ Custom JSON-RPC 2.0 implementation (`internal/mcp/server.go`)
- ✅ SDK adapter for official Go SDK (`internal/mcp/sdk_adapter.go`)
- ⚠️ No framework abstraction layer

### Shared Opportunity

**Extract to shared library (`github.com/davidl71/mcp-go-core`):**
```go
// pkg/mcp/framework/server.go
type MCPServer interface {
    RegisterTool(name, description string, schema ToolSchema, handler ToolHandler) error
    RegisterPrompt(name, description string, handler PromptHandler) error
    RegisterResource(uri, name, description, mimeType string, handler ResourceHandler) error
    Run(ctx context.Context, transport Transport) error
    GetName() string
    CallTool(ctx context.Context, name string, args json.RawMessage) ([]TextContent, error)
    ListTools() []ToolInfo
}

type ToolHandler func(ctx context.Context, args json.RawMessage) ([]TextContent, error)
type PromptHandler func(ctx context.Context, args map[string]interface{}) (string, error)
type ResourceHandler func(ctx context.Context, uri string) ([]byte, string, error)

type Transport interface {
    // Transport-specific methods
}
```

**Benefits:**
- Single source of truth for MCP interface
- Easy framework switching (SDK, custom, etc.)
- Consistent API across projects
- Better testability with mock implementations

**Files to Extract:**
- `exarp-go/internal/framework/server.go` → `mcp-go-core/pkg/mcp/framework/server.go`
- `exarp-go/internal/framework/factory.go` → `mcp-go-core/pkg/mcp/framework/factory.go`
- `exarp-go/internal/framework/adapters/` → `mcp-go-core/pkg/mcp/framework/adapters/`

---

## 2. CLI/TTY Detection

### Current State

**exarp-go:**
- ✅ `IsTTY()` function in `internal/cli/cli.go`
- ✅ Full CLI implementation with tool execution
- ✅ Interactive mode support

**devwisdom-go:**
- ✅ CLI implementation in `internal/cli/app.go`
- ⚠️ No TTY detection (mentioned in TODO but not implemented)

### Shared Opportunity

**Extract to shared library (`github.com/davidl71/mcp-go-core`):**
```go
// pkg/mcp/cli/tty.go
package cli

import "golang.org/x/term"
import "os"

// IsTTY checks if stdin is a terminal
func IsTTY() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

**Benefits:**
- Single implementation for TTY detection
- Consistent CLI behavior
- Easy to extend with additional terminal utilities

**Files to Extract:**
- `exarp-go/internal/cli/cli.go` (IsTTY function) → `mcp-go-core/pkg/mcp/cli/tty.go`
- Consider extracting CLI base utilities (flag parsing, command routing)

---

## 3. Configuration Management

### Current State

**exarp-go:**
- ✅ Environment-based config (`internal/config/config.go`)
- ✅ Framework selection via config
- ✅ Sensible defaults

**devwisdom-go:**
- ✅ File-based config (`internal/config/config.go`)
- ✅ Environment variable overrides
- ✅ Different purpose (wisdom-specific settings)

### Shared Opportunity

**Extract to shared library (`github.com/davidl71/mcp-go-core`):**
```go
// pkg/mcp/config/base.go
type BaseConfig struct {
    Name      string `yaml:"name" env:"MCP_SERVER_NAME"`
    Version   string `yaml:"version" env:"MCP_VERSION"`
    Framework string `yaml:"framework" env:"MCP_FRAMEWORK"`
}

// LoadBaseConfig loads base configuration from environment
func LoadBaseConfig() (*BaseConfig, error) {
    // Implementation
}
```

**Benefits:**
- Common config structure for all MCP servers
- Consistent environment variable naming
- Easy to extend with project-specific config

**Note:** Each project can still have its own config struct that embeds `BaseConfig`.

---

## 4. Security & Path Validation

### Current State

**exarp-go:**
- ✅ Comprehensive path validation (`internal/security/path.go`)
- ✅ Project root detection
- ✅ Directory traversal prevention
- ✅ Access control utilities

**devwisdom-go:**
- ❌ No security utilities

### Shared Opportunity

**Extract to shared library (`github.com/davidl71/mcp-go-core`):**
```go
// pkg/mcp/security/path.go
package security

// ValidatePath ensures a path is within the project root
func ValidatePath(path, projectRoot string) (string, error)

// ValidatePathExists ensures a path is valid AND exists
func ValidatePathExists(path, projectRoot string) (string, error)

// GetProjectRoot attempts to find the project root
func GetProjectRoot(startPath string) (string, error)
```

**Benefits:**
- Security best practices shared across projects
- Consistent path validation
- Prevents security vulnerabilities
- Easy to audit and improve

**Files to Extract:**
- `exarp-go/internal/security/path.go` → `mcp-go-core/pkg/mcp/security/path.go`
- `exarp-go/internal/security/access.go` → `mcp-go-core/pkg/mcp/security/access.go`
- `exarp-go/internal/security/ratelimit.go` → `mcp-go-core/pkg/mcp/security/ratelimit.go`

---

## 5. Logging Infrastructure

### Current State

**exarp-go:**
- ⚠️ Basic logging (standard library)
- ⚠️ No structured logging

**devwisdom-go:**
- ✅ Structured logging (`internal/logging/logger.go`)
- ✅ Request tracing
- ✅ Performance logging
- ✅ Log levels (DEBUG, INFO, WARN, ERROR)

### Shared Opportunity

**Extract to shared library (`github.com/davidl71/mcp-go-core`):**
```go
// pkg/mcp/logging/logger.go
package logging

type Logger struct {
    level         LogLevel
    output        io.Writer
    slowThreshold time.Duration
}

// NewLogger creates a new logger instance
func NewLogger() *Logger

// LogRequest logs the start of a request
func (l *Logger) LogRequest(requestID string, method string)

// LogRequestComplete logs the completion of a request
func (l *Logger) LogRequestComplete(requestID string, method string, duration time.Duration)
```

**Benefits:**
- Consistent logging across projects
- Better debugging and monitoring
- Performance tracking
- Request tracing

**Files to Extract:**
- `devwisdom-go/internal/logging/logger.go` → `mcp-go-core/pkg/mcp/logging/logger.go`
- `devwisdom-go/internal/logging/consultation_log.go` → Keep in devwisdom-go (domain-specific)

---

## 6. JSON-RPC 2.0 Protocol Types

### Current State

**exarp-go:**
- ✅ Uses official Go SDK (types in vendor)
- ⚠️ Framework abstraction hides protocol details

**devwisdom-go:**
- ✅ Custom JSON-RPC 2.0 types (`internal/mcp/protocol.go`)
- ✅ Error code constants
- ✅ Request/Response structures

### Shared Opportunity

**Extract to shared library (`github.com/davidl71/mcp-go-core`):**
```go
// pkg/mcp/protocol/types.go
package protocol

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      interface{}     `json:"id,omitempty"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
    JSONRPC string        `json:"jsonrpc"`
    ID      interface{}   `json:"id"`
    Result  interface{}   `json:"result,omitempty"`
    Error   *JSONRPCError `json:"error,omitempty"`
}

// JSON-RPC 2.0 error codes
const (
    ErrCodeParseError     = -32700
    ErrCodeInvalidRequest = -32600
    ErrCodeMethodNotFound = -32601
    ErrCodeInvalidParams  = -32602
    ErrCodeInternalError  = -32603
)
```

**Benefits:**
- Consistent protocol types
- Shared error handling
- Easy to extend with protocol utilities

**Files to Extract:**
- `devwisdom-go/internal/mcp/protocol.go` → `mcp-go-core/pkg/mcp/protocol/types.go`

---

## 7. Platform Detection

### Current State

**exarp-go:**
- ✅ Platform detection (`internal/platform/detection.go`)
- ✅ Apple Foundation Models support detection
- ✅ macOS version checking

**devwisdom-go:**
- ❌ No platform detection

### Shared Opportunity

**Extract to shared library (`github.com/davidl71/mcp-go-core`):**
```go
// pkg/mcp/platform/detection.go
package platform

// CheckAppleFoundationModelsSupport checks if platform supports Apple Foundation Models
func CheckAppleFoundationModelsSupport() AppleFoundationModelsSupport

// GetOSVersion gets the OS version string
func GetOSVersion() (string, error)
```

**Benefits:**
- Reusable platform utilities
- Consistent platform detection
- Easy to extend for other platforms

**Files to Extract:**
- `exarp-go/internal/platform/detection.go` → `mcp-go-core/pkg/mcp/platform/detection.go`

---

## 8. Common Types & Utilities

### Current State

**exarp-go:**
- ✅ `TextContent` type in framework
- ✅ `ToolSchema` type
- ✅ `ToolInfo` type

**devwisdom-go:**
- ✅ Similar types but different structure
- ⚠️ Some duplication

### Shared Opportunity

**Extract to shared library (`github.com/davidl71/mcp-go-core`):**
```go
// pkg/mcp/types/common.go
package types

// TextContent represents MCP text content
type TextContent struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

// ToolSchema represents tool input schema
type ToolSchema struct {
    Type       string                 `json:"type"`
    Properties map[string]interface{} `json:"properties"`
    Required   []string               `json:"required,omitempty"`
}

// ToolInfo represents tool metadata
type ToolInfo struct {
    Name        string
    Description string
    Schema      ToolSchema
}
```

**Benefits:**
- Consistent type definitions
- Single source of truth
- Better type safety

---

## Implementation Strategy

### Phase 1: Create Shared Library Repository

1. **Create new repository:**
   ```bash
   github.com/davidl71/mcp-go-core
   ```

2. **Initial structure:**
   ```
   mcp-go-core/
   ├── pkg/
   │   └── mcp/
   │       ├── framework/     # Framework abstraction
   │       ├── cli/           # CLI utilities
   │       ├── config/        # Base configuration
   │       ├── security/      # Security utilities
   │       ├── logging/       # Structured logging
   │       ├── protocol/      # JSON-RPC types
   │       ├── platform/      # Platform detection
   │       └── types/         # Common types
   ├── go.mod                 # module github.com/davidl71/mcp-go-core
   ├── README.md
   └── CHANGELOG.md
   ```

### Phase 2: Extract Common Code

**Priority Order:**
1. **High Priority:**
   - Framework abstraction (`framework/`)
   - Security utilities (`security/`)
   - Common types (`types/`)

2. **Medium Priority:**
   - Logging infrastructure (`logging/`)
   - CLI utilities (`cli/`)
   - Protocol types (`protocol/`)

3. **Low Priority:**
   - Platform detection (`platform/`)
   - Base configuration (`config/`)

### Phase 3: Update Projects

1. **Add dependency:**
   ```go
   // go.mod
   require github.com/davidl71/mcp-go-core v0.1.0
   ```

2. **Update imports:**
   ```go
   // Before
   import "github.com/davidl71/exarp-go/internal/framework"
   
   // After
   import "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
   ```

3. **Maintain backward compatibility:**
   - Keep internal packages as wrappers initially
   - Gradually migrate to shared library
   - Remove internal packages after migration

---

## Benefits Summary

### Code Reuse
- ✅ Eliminate duplication between projects
- ✅ Single source of truth for common functionality
- ✅ Easier maintenance and bug fixes

### Consistency
- ✅ Consistent APIs across projects
- ✅ Shared best practices
- ✅ Unified security model

### Development Speed
- ✅ Faster development of new MCP servers
- ✅ Reusable components
- ✅ Better testing infrastructure

### Quality
- ✅ Shared code gets more testing
- ✅ Easier to audit security
- ✅ Better documentation

---

## Risks & Mitigation

### Risk 1: Over-Engineering
**Mitigation:** Start small, extract only proven, stable code. Don't extract experimental features.

### Risk 2: Breaking Changes
**Mitigation:** Use semantic versioning, maintain changelog, provide migration guides.

### Risk 3: Dependency Management
**Mitigation:** Use Go modules, pin versions, test compatibility regularly.

### Risk 4: Project-Specific Needs
**Mitigation:** Keep project-specific code in projects. Only extract truly common code.

---

## Next Steps

1. **Create shared library repository**
2. **Extract high-priority components** (framework, security, types)
3. **Update exarp-go to use shared library**
4. **Update devwisdom-go to use shared library**
5. **Extract medium-priority components** (logging, CLI, protocol)
6. **Document migration process**
7. **Create examples and best practices**

---

## 9. Build System Patterns

### Current State

**exarp-go:**
- ✅ Comprehensive Makefile with 50+ targets
- ✅ Tool detection and configuration (`make config`)
- ✅ Development scripts (`dev.sh`, `dev-go.sh`)
- ✅ Cross-platform support
- ✅ Version information from git tags
- ✅ Color-coded output
- ✅ Conditional targets based on available tools

**devwisdom-go:**
- ✅ Clean Makefile with essential targets
- ✅ Watchdog script for crash monitoring
- ✅ Cross-compilation support (Windows, Linux, macOS)
- ✅ Release packaging (zip/tar.gz)
- ✅ Version information from git tags
- ✅ Benchmarking targets

### Shared Opportunity

**Extract to shared library:**
```makefile
# mcp-go-core/Makefile.common
# Common Makefile targets for MCP Go projects

# Version information (shared pattern)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Colors (shared pattern)
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
BLUE := \033[0;34m
NC := \033[0m

# Common targets
.PHONY: help build test clean fmt lint

help: ## Show help message
	@echo "$(BLUE)Available targets:$(NC)"
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "} {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "$(BLUE)Building...$(NC)"
	@go build -o $(BINARY_PATH) ./cmd/server

test: ## Run tests
	@go test ./... -v

clean: ## Clean build artifacts
	@rm -f $(BINARY_PATH)
	@go clean

fmt: ## Format code
	@go fmt ./...

lint: ## Lint code
	@golangci-lint run ./...
```

**Benefits:**
- Consistent build targets across projects
- Shared version detection logic
- Common color schemes
- Reusable Makefile patterns
- Easy to extend with project-specific targets

**Files to Extract:**
- Common Makefile patterns → `mcp-go-core/Makefile.common`
- Version detection logic → Shared Makefile variables
- Color definitions → Shared Makefile variables

---

## 10. CI/CD Workflow Patterns

### Current State

**exarp-go:**
- ✅ GitHub Actions workflow (`.github/workflows/go.yml`)
- ✅ Multiple jobs: test, lint, vet, build, security, module-check, scorecard
- ✅ Codecov integration
- ✅ Go module verification
- ✅ Security scanning with govulncheck

**devwisdom-go:**
- ✅ GitHub Actions workflow (`.github/workflows/ci.yml`)
- ✅ Matrix testing (multiple Go versions)
- ✅ Coverage upload as artifacts
- ✅ Release workflow (`.github/workflows/release.yml`)
- ✅ Cross-platform builds in CI

### Shared Opportunity

**Extract to shared library:**
```yaml
# mcp-go-core/.github/workflows/common-go.yml
# Reusable GitHub Actions workflow for MCP Go projects

name: Go CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    name: Go Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      - run: go mod download
      - run: go test ./... -v -coverprofile=coverage.out
      - uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  lint:
    name: Go Linting
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

  build:
    name: Go Build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go build -o bin/$(PROJECT_NAME) ./cmd/server
```

**Benefits:**
- Consistent CI/CD across projects
- Shared workflow templates
- Easy to maintain and update
- Best practices enforcement

**Files to Extract:**
- Common workflow steps → `mcp-go-core/.github/workflows/common-go.yml`
- Release workflow template → `mcp-go-core/.github/workflows/release-template.yml`

---

## 11. Development Scripts

### Current State

**exarp-go:**
- ✅ `dev.sh` - Development automation with file watching
- ✅ `dev-go.sh` - Go-specific development mode
- ✅ File watcher detection (fswatch, inotifywait, polling fallback)
- ✅ Auto-reload, auto-test, coverage generation
- ✅ Colored logging with timestamps

**devwisdom-go:**
- ✅ `watchdog.sh` - Crash monitoring and file watching
- ✅ PID file management
- ✅ Restart limits and delays
- ✅ File watching with reload/restart options
- ✅ Colored logging

### Shared Opportunity

**Extract to shared library:**
```bash
# mcp-go-core/scripts/dev-common.sh
# Common development script utilities

# Shared configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"

# Shared colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Shared logging function
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%H:%M:%S')
    
    case "$level" in
        INFO) echo -e "${GREEN}[$timestamp] [INFO]${NC} $message" ;;
        WARN) echo -e "${YELLOW}[$timestamp] [WARN]${NC} $message" ;;
        ERROR) echo -e "${RED}[$timestamp] [ERROR]${NC} $message" >&2 ;;
    esac
}

# Shared file watcher detection
detect_file_watcher() {
    if command -v fswatch >/dev/null 2>&1; then
        echo "fswatch"
    elif command -v inotifywait >/dev/null 2>&1; then
        echo "inotifywait"
    else
        echo "polling"
    fi
}
```

**Benefits:**
- Consistent development experience
- Shared utilities for file watching
- Common logging patterns
- Easy to extend

**Files to Extract:**
- Common script utilities → `mcp-go-core/scripts/dev-common.sh`
- File watcher detection → Shared function
- Logging utilities → Shared functions

---

## 12. Documentation Patterns

### Current State

**exarp-go:**
- ✅ Extensive `docs/` directory (200+ markdown files)
- ✅ Structured documentation (README, PRD, migration docs, etc.)
- ✅ Code examples in documentation
- ✅ Architecture diagrams and design docs
- ✅ Migration guides and lessons learned

**devwisdom-go:**
- ✅ Focused `docs/` directory (17 markdown files)
- ✅ Phase-based documentation
- ✅ Quick start guides
- ✅ API documentation
- ✅ Performance documentation

### Shared Opportunity

**Extract to shared library:**
```markdown
# mcp-go-core/docs/TEMPLATES/

## Template Files:
- README_TEMPLATE.md - Standard README structure
- CONTRIBUTING_TEMPLATE.md - Contribution guidelines
- ARCHITECTURE_TEMPLATE.md - Architecture documentation template
- API_DOCS_TEMPLATE.md - API documentation template
- MIGRATION_GUIDE_TEMPLATE.md - Migration guide template
```

**Benefits:**
- Consistent documentation structure
- Shared templates for common docs
- Best practices documentation
- Easy onboarding

**Documentation to Extract:**
- README template → `mcp-go-core/docs/TEMPLATES/README_TEMPLATE.md`
- Contributing guidelines template → `mcp-go-core/docs/TEMPLATES/CONTRIBUTING_TEMPLATE.md`
- Architecture documentation template → `mcp-go-core/docs/TEMPLATES/ARCHITECTURE_TEMPLATE.md`

---

## Updated Implementation Strategy

### Phase 1: Create Shared Library Repository

1. **Create new repository:**
   ```bash
   github.com/davidl71/mcp-go-core
   ```

2. **Initial structure:**
   ```
   mcp-go-core/
   ├── pkg/
   │   └── mcp/
   │       ├── framework/     # Framework abstraction
   │       ├── cli/           # CLI utilities
   │       ├── config/        # Base configuration
   │       ├── security/      # Security utilities
   │       ├── logging/       # Structured logging
   │       ├── protocol/      # JSON-RPC types
   │       ├── platform/      # Platform detection
   │       └── types/         # Common types
   ├── scripts/
   │   └── dev-common.sh      # Shared development utilities
   ├── .github/
   │   └── workflows/
   │       └── common-go.yml  # Reusable CI/CD workflow
   ├── docs/
   │   └── TEMPLATES/         # Documentation templates
   ├── Makefile.common        # Shared Makefile patterns
   ├── go.mod
   ├── README.md
   └── CHANGELOG.md
   ```

### Phase 2: Extract Common Code

**Priority Order:**
1. **High Priority:**
   - Framework abstraction (`framework/`)
   - Security utilities (`security/`)
   - Common types (`types/`)

2. **Medium Priority:**
   - Logging infrastructure (`logging/`)
   - CLI utilities (`cli/`)
   - Protocol types (`protocol/`)
   - Build system patterns (`Makefile.common`)
   - CI/CD workflows (`.github/workflows/`)

3. **Low Priority:**
   - Platform detection (`platform/`)
   - Base configuration (`config/`)
   - Development scripts (`scripts/`)
   - Documentation templates (`docs/TEMPLATES/`)

### Phase 3: Update Projects

1. **Add dependency:**
   ```go
   // go.mod
   require github.com/davidl71/mcp-go-core v0.1.0
   ```

2. **Include shared Makefile:**
   ```makefile
   # In project Makefile
   include $(shell go list -m -f '{{.Dir}}' github.com/davidl71/mcp-go-core)/Makefile.common
   ```

3. **Use shared CI/CD workflow:**
   ```yaml
   # .github/workflows/ci.yml
   # Import shared workflow
   ```

---

## Conclusion

There is significant opportunity to extract shared code between exarp-go and devwisdom-go. The shared library would improve code quality, reduce duplication, and make it easier to build new MCP servers in the future.

**Recommended Approach:** Start with a small, focused shared library containing the framework abstraction and security utilities, then gradually expand based on actual needs.
