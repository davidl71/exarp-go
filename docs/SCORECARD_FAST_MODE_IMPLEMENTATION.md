# Scorecard Fast Mode Implementation

**Date:** 2026-01-09  
**Status:** ✅ Implemented

## Performance Results

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| **Execution Time** | ~15 seconds | ~1.3 seconds | **91% faster** |
| **User CPU** | 12.72s | 1.87s | 85% reduction |
| **System CPU** | 6.87s | 0.97s | 86% reduction |

## Implementation Details

### Changes Made

1. **Added `ScorecardOptions` struct** (`internal/tools/scorecard_go.go`)
   ```go
   type ScorecardOptions struct {
       FastMode bool // Skip expensive operations
   }
   ```

2. **Updated `GenerateGoScorecard` function signature**
   - Now accepts `opts *ScorecardOptions` parameter
   - If `nil`, uses default (full checks)
   - If `FastMode: true`, skips expensive operations

3. **Modified `performGoHealthChecks` function**
   - Skips `go mod tidy` in fast mode
   - Skips `go build ./...` in fast mode
   - Skips `golangci-lint run` in fast mode
   - Skips `go test ./...` in fast mode (sets coverage to 0.0)
   - Skips `govulncheck` in fast mode
   - Still runs: `go vet`, `gofmt`, file system checks, security checks

4. **Updated handlers**
   - `handleReport` in `internal/tools/handlers.go` uses fast mode by default
   - `cmd/scorecard/main.go` uses fast mode by default

### Operations Skipped in Fast Mode

| Operation | Timeout | Skipped | Reason |
|-----------|---------|---------|--------|
| `go mod tidy` | 30s | ✅ | Modifies files, slow |
| `go build ./...` | 60s | ✅ | Builds entire project |
| `golangci-lint run` | 60s | ✅ | Full linting |
| `go test ./...` | 120s | ✅ | Runs all tests with coverage |
| `govulncheck ./...` | 60s | ✅ | Vulnerability scanning |

### Operations Still Run in Fast Mode

| Operation | Timeout | Status | Notes |
|-----------|---------|--------|-------|
| `go vet ./...` | 30s | ✅ | Quick static analysis |
| `gofmt -l .` | 30s | ✅ | Format checking |
| File counting | - | ✅ | Fast file system walk |
| Security checks | - | ✅ | File existence checks |
| Go version | 5s | ✅ | Quick command |

## Usage

### Default (Fast Mode)

```go
opts := &tools.ScorecardOptions{FastMode: true}
scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot, opts)
```

### Full Mode (All Checks)

```go
opts := &tools.ScorecardOptions{FastMode: false}
// or
scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot, nil)
```

## Impact

### Before Optimization
- **Total time:** ~15 seconds
- **Bottlenecks:**
  - `go test ./...` with coverage: ~5-8s
  - `go build ./...`: ~3-5s
  - `go mod tidy`: ~2-3s
  - `golangci-lint`: ~2-4s

### After Optimization
- **Total time:** ~1.3 seconds
- **Remaining operations:**
  - File counting: ~0.5s
  - `go vet`: ~1-2s
  - `gofmt`: ~0.5s
  - Security checks: ~0.1s

## Trade-offs

### What's Lost in Fast Mode
- ❌ Test execution status
- ❌ Test coverage percentage
- ❌ Build verification
- ❌ Lint results
- ❌ Vulnerability scan results
- ❌ Go mod tidy verification

### What's Retained
- ✅ File and line counts
- ✅ Go version
- ✅ `go vet` results
- ✅ `gofmt` compliance
- ✅ Security feature checks
- ✅ Overall project metrics

## Future Enhancements

1. **Result Caching** - Cache results for 5 minutes
2. **Parallel Execution** - Run independent checks concurrently
3. **Incremental Checks** - Only check changed files
4. **Configurable Options** - Allow users to choose which checks to skip

## Code Locations

- **Options struct:** `internal/tools/scorecard_go.go:60-63`
- **Health checks:** `internal/tools/scorecard_go.go:108-154`
- **Handler:** `internal/tools/handlers.go:224`
- **CLI:** `cmd/scorecard/main.go:21`

