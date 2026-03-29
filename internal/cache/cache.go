// Package cache re-exports generic caches from mcp-go-core and keeps exarp-only singletons.
package cache

import (
	"sync"

	mcpcache "github.com/davidl71/mcp-go-core/pkg/mcp/cache"
)

// Re-export mcp-go-core caches (see docs/MODULARIZATION_PACKAGE_MAP.md).
type (
	FileCache = mcpcache.FileCache
	TTLCache  = mcpcache.TTLCache
)

var (
	NewFileCache       = mcpcache.NewFileCache
	GetGlobalFileCache = mcpcache.GetGlobalFileCache
	NewTTLCache        = mcpcache.NewTTLCache
)

var (
	scorecardCache     *TTLCache
	scorecardCacheOnce sync.Once
)

// GetScorecardCache returns a singleton TTL cache for scorecard results.
func GetScorecardCache() *TTLCache {
	scorecardCacheOnce.Do(func() {
		scorecardCache = NewTTLCache()
	})
	return scorecardCache
}
