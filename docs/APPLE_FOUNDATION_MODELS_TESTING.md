# Apple Foundation Models - Testing Guide

**Date**: 2026-01-07  
**Status**: Implementation Complete  
**Project**: exarp-go

---

## Testing Strategy

### Unit Tests (No Swift Bridge Required)

Unit tests for helper functions can run on all platforms:

```bash
make test-apple-fm-unit
```

**Tests:**
- `TestGetTemperature` - Parameter extraction for temperature
- `TestGetMaxTokens` - Parameter extraction for max_tokens

**Status**: ✅ All passing

### Integration Tests (Requires Swift Bridge)

Integration tests require the Swift bridge to be built:

```bash
# Build Swift bridge first
make build-swift-bridge

# Run integration tests
make test-apple-fm-integration
```

**Tests:**
- Handler argument parsing
- Platform detection integration
- Action routing (generate, summarize, classify)
- Error handling
- MCP tool registration
- Actual Foundation Models API calls

---

## Build Tags

The Apple Foundation Models code uses conditional compilation:

```go
//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo
```

This ensures:
- ✅ Code compiles on all platforms
- ✅ Apple FM features only enabled on supported platforms
- ✅ No build failures on Linux/Windows/Intel Macs

---

## Makefile Targets

### Build Targets

- `build-apple-fm` - Build with Apple Foundation Models support (CGO_ENABLED=1)
- `build-swift-bridge` - Build Swift bridge in go-foundationmodels package

### Test Targets

- `test-apple-fm` - Run all Apple Foundation Models tests
- `test-apple-fm-unit` - Run unit tests (no Swift bridge required)
- `test-apple-fm-integration` - Run integration tests (requires Swift bridge)

---

## Iterative Workflow

### Step 1: Unit Tests (No Swift Bridge)

```bash
# Run unit tests - works on all platforms
make test-apple-fm-unit
```

**Expected**: All tests pass ✅

### Step 2: Build Verification

```bash
# Build without CGO (should work on all platforms)
CGO_ENABLED=0 go build ./cmd/server

# Build with CGO (requires Swift bridge on Apple platforms)
make build-apple-fm
```

### Step 3: Build Swift Bridge (One-Time Setup)

```bash
# Build Swift bridge
make build-swift-bridge
```

**Requirements:**
- Xcode installed
- macOS 26+ (Tahoe)
- Apple Intelligence enabled

### Step 4: Integration Tests

```bash
# Run integration tests
make test-apple-fm-integration
```

---

## Test Files

### Unit Tests
- `internal/tools/apple_foundation_helpers_test.go` - Helper function tests
- `internal/tools/apple_foundation_test.go` - Handler tests (requires build tags)

### Integration Tests
- `tests/integration/mcp/test_apple_foundation_models.py` - MCP integration tests

### Platform Tests
- `internal/platform/detection_test.go` - Platform detection tests

---

## Current Test Status

✅ **Unit Tests**: All passing
- `TestGetTemperature` - ✅ PASS
- `TestGetMaxTokens` - ✅ PASS

✅ **Platform Detection**: All passing
- `TestCheckAppleFoundationModelsSupport` - ✅ PASS
- `TestIsMacOS26OrLater` - ✅ PASS

⚠️ **Integration Tests**: Require Swift bridge
- Handler tests need Swift bridge to be built
- MCP integration tests need server with Swift bridge

---

## Troubleshooting

### Build Failures

**Error**: `clang: error: no such file or directory: 'libFMShim.a'`

**Solution**: Build Swift bridge first:
```bash
make build-swift-bridge
```

### Test Failures

**Unit tests fail with linking errors**:
- Use `make test-apple-fm-unit` which tests only helper functions
- Handler tests require Swift bridge

**Integration tests fail**:
- Ensure Swift bridge is built
- Verify Apple Intelligence is enabled
- Check platform support: `go test ./internal/platform -v`

---

## Next Steps

1. ✅ Unit tests complete and passing
2. ⏳ Build Swift bridge when ready for integration tests
3. ⏳ Run integration tests with actual Foundation Models API
4. ⏳ Test on unsupported platforms to verify graceful fallback

