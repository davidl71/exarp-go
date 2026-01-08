# Apple Foundation Models Integration Plan

**Date**: 2026-01-07  
**Status**: Planning  
**Project**: exarp-go

---

## Overview

This document outlines the integration plan for Apple Foundation Models into exarp-go, enabling on-device AI capabilities on supported Apple Silicon Macs.

## System Requirements

- **Operating System**: macOS 26 (Tahoe) or later
- **Hardware**: Apple Silicon Mac (M1, M2, M3, M4, or newer)
- **Software**: Apple Intelligence enabled in System Settings
- **Go Version**: Go 1.24+ (already satisfied - exarp-go uses Go 1.24.0)
- **Xcode**: 15.x or later (for Swift bridge compilation if using go-foundationmodels)

## Integration Options

### Option 1: go-foundationmodels (Recommended)
- **Package**: `github.com/blacktop/go-foundationmodels`
- **Approach**: Direct Go interface using CGO + Swift bridge
- **Pros**: 
  - Self-contained binary (no external dependencies)
  - Native Go API
  - Direct integration with Apple Foundation Models
- **Cons**: 
  - Requires Xcode for compilation
  - CGO dependency

### Option 2: libai (C Bridge)
- **Package**: C library bridge
- **Approach**: C bindings through CGO
- **Pros**: 
  - Language-agnostic
  - Supports MCP protocol
- **Cons**: 
  - External C library dependency
  - More complex integration

## Integration Architecture

### Platform Detection
- Detect macOS version (26+)
- Detect Apple Silicon architecture (arm64)
- Check Apple Intelligence availability
- Graceful fallback on unsupported platforms

### Tool Integration
- Add new tool: `apple_foundation_models` or integrate into existing tools
- Support for:
  - Text summarization
  - Classification
  - Short dialogues
  - On-device processing (privacy-focused)

### Configuration
- Add configuration options for Apple Foundation Models
- Enable/disable feature based on platform support
- Configurable model selection

## Implementation Phases

1. **Research & Evaluation** - Compare integration libraries, test compatibility
2. **Platform Detection** - Implement runtime detection for supported machines
3. **Integration Design** - Design architecture for seamless integration
4. **Implementation** - Implement Apple Foundation Models support
5. **Testing** - Test on supported and unsupported platforms
6. **Documentation** - Document usage and requirements

## Files to Create/Modify

- `internal/platform/detection.go` - Platform detection utilities
- `internal/tools/apple_foundation.go` - Apple Foundation Models tool handler
- `internal/config/config.go` - Add Apple Foundation Models configuration
- `go.mod` - Add dependency (go-foundationmodels or libai)
- `docs/APPLE_FOUNDATION_MODELS_USAGE.md` - Usage documentation

## Considerations

- **Privacy**: On-device processing ensures user data remains private
- **Performance**: Optimized for Apple Silicon
- **Fallback**: Must gracefully handle unsupported platforms
- **Optional**: Feature should be optional, not required for exarp-go to function

