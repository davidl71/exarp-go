# File I/O Caching Implementation

## Overview

File caching has been implemented to reduce repeated file I/O operations for frequently accessed files like `go.mod`, `pyproject.toml`, configuration files, and handoff files.

## Implementation

### Core Cache (`internal/cache/file_cache.go`)

The file cache provides:
- **Thread-safe caching** using `sync.Map`
- **mtime-based invalidation** - automatically invalidates when files are modified
- **TTL support** - optional time-based expiration
- **Global singleton** - `GetGlobalFileCache()` for easy access

### Key Features

1. **Automatic Invalidation**: Cache entries are automatically invalidated when file modification time changes
2. **Thread-Safe**: Uses `sync.Map` for concurrent access
3. **Memory Efficient**: Only caches file content, not metadata
4. **Simple API**: Drop-in replacement for `os.ReadFile`

### Usage

```go
import "github.com/davidl71/exarp-go/internal/cache"

// Get global cache instance
fileCache := cache.GetGlobalFileCache()

// Read file (returns content, cache hit bool, error)
content, hit, err := fileCache.ReadFile(path)

// Read file with TTL (expires after specified duration)
content, hit, err := fileCache.ReadFileWithTTL(path, 5*time.Minute)

// Invalidate specific file
fileCache.Invalidate(path)

// Clear all cache
fileCache.Clear()

// Get cache statistics
stats := fileCache.GetStats()
```

## Files Updated

The following files have been updated to use file caching:

1. **`internal/tools/report.go`** - `go.mod` reads
2. **`internal/tools/attribution_check.go`** - `go.mod` and `pyproject.toml` reads
3. **`internal/tools/health_check.go`** - `go.mod` and `pyproject.toml` reads
4. **`internal/tools/session.go`** - Handoff file reads (5 instances) and `cursor-agent.json`
5. **`internal/tools/hooks_setup.go`** - Config file reads
6. **`internal/tools/todo2_utils.go`** - `.todo2/state.todo2.json` reads
7. **`internal/tools/workflow_mode.go`** - State file reads

## Performance Impact

### Expected Improvements

- **50%+ reduction** in file I/O operations for cached files
- **Faster tool execution** for tools that read the same files multiple times
- **Reduced disk I/O** pressure, especially for frequently accessed config files

### Cache Hit Scenarios

Files that benefit most from caching:
- `go.mod` - Read by multiple tools (report, health, attribution)
- `pyproject.toml` - Read by health and attribution tools
- Handoff files - Read multiple times in session operations
- Config files - Read during tool initialization
- `.todo2/state.todo2.json` - Read frequently during task operations

## Testing

Comprehensive tests are included in `internal/cache/file_cache_test.go`:
- ✅ Basic read/cache hit behavior
- ✅ Invalidation on file modification
- ✅ TTL expiration
- ✅ Manual invalidation
- ✅ Cache clearing
- ✅ Statistics collection
- ✅ Global singleton behavior

Run tests:
```bash
go test ./internal/cache/... -v
```

## Future Enhancements

Potential improvements:
1. **LRU eviction** - Limit cache size with least-recently-used eviction
2. **Cache warming** - Pre-load frequently accessed files on startup
3. **Metrics integration** - Add cache hit/miss metrics to monitoring
4. **Configurable TTL** - Per-file or per-pattern TTL configuration
5. **Cache size limits** - Prevent unbounded memory growth

## Migration Guide

To migrate additional files to use caching:

1. Import the cache package:
```go
import "github.com/davidl71/exarp-go/internal/cache"
```

2. Replace `os.ReadFile`:
```go
// Before
data, err := os.ReadFile(path)

// After
fileCache := cache.GetGlobalFileCache()
data, _, err := fileCache.ReadFile(path)
```

3. Handle the cache hit boolean if needed (usually can be ignored):
```go
data, hit, err := fileCache.ReadFile(path)
// hit indicates if data came from cache (true) or disk (false)
```

## Notes

- The cache is **in-memory only** - entries are lost on restart
- Cache entries are **automatically invalidated** when files are modified
- The global cache is a **singleton** - all tools share the same cache instance
- **Thread-safe** - safe for concurrent access from multiple goroutines
