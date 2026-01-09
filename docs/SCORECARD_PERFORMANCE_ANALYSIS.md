# Scorecard Performance Analysis

**Issue:** Project scorecard generation takes ~15 seconds  
**Date:** 2026-01-09

## Performance Measurement

```bash
time go run cmd/scorecard/main.go
# Result: 12.72s user 6.87s system 130% cpu 15.066 total
```

## Root Cause Analysis

The scorecard generation performs **multiple expensive operations sequentially**:

### 1. File System Operations (Fast - ~1-2 seconds)
- `countGoFiles()` - Walks entire project, reads every `.go` file
- `countGoTestFiles()` - Walks entire project again, reads every `_test.go` file  
- `countPythonFiles()` - Walks entire project again, reads every `.py` file
- **Total:** ~3 file system walks, reading all source files

### 2. External Command Executions (SLOW - ~12-13 seconds)

These run **sequentially** and are the main bottleneck:

| Operation | Timeout | Actual Impact | Notes |
|-----------|---------|---------------|-------|
| `go mod tidy` | 30s | ~2-3s | Actually modifies go.mod/go.sum |
| `go build ./...` | 60s | ~3-5s | **Actually builds entire project** |
| `go vet ./...` | 30s | ~1-2s | Static analysis |
| `gofmt -l .` | 30s | ~0.5s | Format checking |
| `golangci-lint run` | 60s | ~2-4s | Full linting (if configured) |
| `go test ./...` | 120s | ~5-8s | **Runs all tests with coverage** |
| `govulncheck ./...` | 60s | ~1-2s | Vulnerability scanning |

**Total Sequential Time:** ~15-25 seconds (depending on project size)

## Key Problems

### 1. **Sequential Execution**
All operations run one after another, not in parallel:
```go
health.GoModTidyPasses = checkGoModTidy(ctx, projectRoot)      // Wait
health.GoBuildPasses = checkGoBuild(ctx, projectRoot)          // Wait
health.GoVetPasses = checkGoVet(ctx, projectRoot)              // Wait
health.GoTestPasses, health.GoTestCoverage = checkGoTest(...)   // Wait (longest!)
```

### 2. **Expensive Operations**
- `go test ./...` - Runs **all tests** with coverage (120s timeout)
- `go build ./...` - **Builds entire project** (60s timeout)
- `go mod tidy` - **Modifies go.mod/go.sum** (30s timeout)

### 3. **No Caching**
Results are not cached, so every scorecard generation repeats all operations.

### 4. **No Early Exit**
Even if one check fails, all subsequent checks still run.

## Optimization Strategies

### Strategy 1: Parallel Execution (High Impact)
Run independent checks in parallel using goroutines:

```go
var wg sync.WaitGroup
results := make(chan healthCheckResult, 10)

// Run checks in parallel
wg.Add(1)
go func() {
    defer wg.Done()
    results <- healthCheckResult{"build", checkGoBuild(ctx, projectRoot)}
}()

wg.Add(1)
go func() {
    defer wg.Done()
    results <- healthCheckResult{"vet", checkGoVet(ctx, projectRoot)}
}()

// ... etc
```

**Expected Improvement:** 15s → ~5-8s (60% faster)

### Strategy 2: Skip Expensive Operations (Medium Impact)
Make expensive checks optional or cached:

- **Skip `go mod tidy`** - Only check if go.mod/go.sum exist
- **Skip `go build`** - Only check if binary exists or use cached result
- **Skip `go test`** - Only check if tests exist, don't run them
- **Skip `golangci-lint`** - Only check if configured, don't run

**Expected Improvement:** 15s → ~3-5s (70% faster)

### Strategy 3: Result Caching (High Impact)
Cache results with TTL (e.g., 5 minutes):

```go
type cachedResult struct {
    Result    *GoScorecardResult
    Timestamp time.Time
}

var cache = make(map[string]*cachedResult)
var cacheMutex sync.RWMutex

func GetCachedScorecard(projectRoot string) (*GoScorecardResult, error) {
    cacheMutex.RLock()
    cached, exists := cache[projectRoot]
    cacheMutex.RUnlock()
    
    if exists && time.Since(cached.Timestamp) < 5*time.Minute {
        return cached.Result, nil
    }
    
    // Generate fresh
    result, err := GenerateGoScorecard(ctx, projectRoot)
    // ... cache it
    return result, err
}
```

**Expected Improvement:** 15s → ~0.1s (99% faster for cached requests)

### Strategy 4: Incremental Checks (Medium Impact)
Only run checks that have changed since last run:

- Track file modification times
- Skip checks if no relevant files changed
- Use file hashes for dependency tracking

**Expected Improvement:** 15s → ~2-5s (varies by change frequency)

### Strategy 5: Fast Mode (High Impact)
Add a "fast" mode that skips expensive operations:

```go
type ScorecardOptions struct {
    FastMode bool  // Skip expensive checks
    SkipTests bool // Skip test execution
    SkipBuild bool // Skip build check
}

func GenerateGoScorecardWithOptions(ctx context.Context, projectRoot string, opts ScorecardOptions) (*GoScorecardResult, error) {
    // ...
    if !opts.FastMode {
        health.GoTestPasses, health.GoTestCoverage = checkGoTest(ctx, projectRoot)
    }
    // ...
}
```

**Expected Improvement:** 15s → ~2-3s (80% faster)

## Recommended Solution

**Combination Approach:**

1. **Immediate:** Add "fast mode" option (Strategy 5)
   - Skip `go test`, `go build`, `go mod tidy`
   - Keep file counting and quick checks
   - **Result:** 15s → ~2-3s

2. **Short-term:** Implement result caching (Strategy 3)
   - Cache results for 5 minutes
   - Invalidate on file changes
   - **Result:** 2-3s → ~0.1s (cached)

3. **Long-term:** Parallel execution (Strategy 1)
   - Run independent checks concurrently
   - **Result:** 2-3s → ~1-2s (uncached)

## Implementation Priority

1. **High Priority:** Fast mode + caching
2. **Medium Priority:** Parallel execution
3. **Low Priority:** Incremental checks

## Code Changes Required

### 1. Add Options Struct
```go
type ScorecardOptions struct {
    FastMode  bool
    UseCache  bool
    CacheTTL  time.Duration
}
```

### 2. Modify Function Signature
```go
func GenerateGoScorecard(ctx context.Context, projectRoot string, opts *ScorecardOptions) (*GoScorecardResult, error)
```

### 3. Add Caching Layer
```go
var scorecardCache = sync.Map{} // map[string]*cachedScorecard
```

### 4. Skip Expensive Operations in Fast Mode
```go
if !opts.FastMode {
    health.GoTestPasses, health.GoTestCoverage = checkGoTest(ctx, projectRoot)
    health.GoBuildPasses = checkGoBuild(ctx, projectRoot)
    health.GoModTidyPasses = checkGoModTidy(ctx, projectRoot)
}
```

## Expected Performance After Optimization

| Scenario | Current | After Optimization |
|----------|---------|-------------------|
| First run (uncached) | 15s | 2-3s (fast mode) |
| Cached run | 15s | 0.1s |
| Parallel + fast | 15s | 1-2s |

**Overall Improvement:** 90-99% faster for typical usage

