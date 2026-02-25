package cache

import (
	"sync"

	mcpcache "github.com/davidl71/mcp-go-core/pkg/mcp/cache"
)

// FileCache provides thread-safe file caching with mtime-based invalidation (re-exported from core).
type FileCache = mcpcache.FileCache

// NewFileCache creates a new file cache instance.
var NewFileCache = mcpcache.NewFileCache

var (
	globalFileCache     *FileCache
	globalFileCacheOnce sync.Once
)

// GetGlobalFileCache returns the global file cache instance (singleton).
func GetGlobalFileCache() *FileCache {
	globalFileCacheOnce.Do(func() {
		globalFileCache = NewFileCache()
	})
	return globalFileCache
}
