# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Performance optimizations documentation

## [0.3.0] - 2026-01-12

### Added
- **Performance Optimizations**
  - Fast path for empty middleware chains (skip wrapping when no middleware)
  - Pre-allocated slice capacity (avoid reallocations)
  - Nil slice returns (better memory usage)
  - Optimized context validation (non-blocking check)
  - Performance tracking utilities (`pkg/mcp/framework/adapters/gosdk/performance.go`)
- **Migration Guide**
  - Comprehensive migration guide from exarp-go/devwisdom-go to mcp-go-core
  - Step-by-step instructions with before/after examples
  - Common issues and solutions
  - Verification checklist
- **Usage Examples**
  - Basic server example (tools, prompts, resources, CLI/MCP dual mode)
  - Advanced server example (logging, middleware, adapter options)
  - Example README with usage instructions
- **Documentation**
  - Performance documentation (`docs/PERFORMANCE.md`)
  - Migration guide (`docs/MIGRATION_GUIDE.md`)
  - Example documentation (`examples/README.md`)

### Changed
- **Middleware Performance**
  - Skip middleware wrapping when no middleware registered (fast path)
  - Optimized middleware chain operations
- **Memory Management**
  - Return nil slices instead of empty slices for better memory usage
  - Pre-allocate slice capacity to avoid reallocations

### Fixed
- Context validation now uses non-blocking select pattern

## [0.2.0] - 2026-01-12

### Added
- **SSE Transport Support**
  - Server-Sent Events (SSE) transport implementation
  - HTTP server integration
  - Client connection management
  - Graceful shutdown handling
- **Platform Detection**
  - OS and architecture detection utilities
  - Platform-specific path handling
  - Compatibility checking
- **CLI Utilities**
  - TTY detection (`pkg/mcp/cli`)
  - Command-line argument parsing
  - CLI/MCP dual mode support

### Changed
- Transport interface expanded to support SSE
- Framework adapters updated for SSE transport

## [0.1.0] - 2026-01-12

### Added
- Initial project structure
- **Core Framework**
  - MCPServer interface (`pkg/mcp/framework`)
  - Framework factory pattern (`pkg/mcp/factory`)
  - Go SDK adapter (`pkg/mcp/framework/adapters/gosdk`)
- **Security**
  - Project root detection (`pkg/mcp/security/GetProjectRoot()`)
  - Path validation utilities (`pkg/mcp/security/ValidatePath()`)
  - Access control (`pkg/mcp/security/access.go`)
  - Rate limiting (`pkg/mcp/security/ratelimit.go`)
- **Configuration**
  - Base configuration (`pkg/mcp/config`)
  - Configuration builder pattern
- **Logging**
  - Structured logging (`pkg/mcp/logging`)
  - Log levels and context support
- **Types**
  - Common MCP types (`pkg/mcp/types`)
  - Protocol types (`pkg/mcp/protocol`)
- **Middleware**
  - Middleware chain support for tools, prompts, resources
  - Adapter options pattern
- **Development Tools**
  - Basic Makefile with common targets
  - GitHub Actions CI workflow
  - Project documentation (README, verification guide)

### Changed
- N/A (initial release)

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- Path validation prevents directory traversal attacks
- Project root detection ensures operations stay within project boundaries
