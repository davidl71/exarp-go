// Package cache provides file-based and TTL in-memory caching for scorecard, reports, and other data.
package cache

import (
	"sync"

	mcpcache "github.com/davidl71/mcp-go-core/pkg/mcp/cache"
)

// TTLCache is a thread-safe key-value cache with per-entry TTL (re-exported from core).
type TTLCache = mcpcache.TTLCache

// NewTTLCache returns a new TTL cache.
var NewTTLCache = mcpcache.NewTTLCache

var (
	scorecardCache     *TTLCache
	scorecardCacheOnce sync.Once
)

// GetScorecardCache returns a singleton TTL cache for scorecard results (e.g. 5-minute TTL).
func GetScorecardCache() *TTLCache {
	scorecardCacheOnce.Do(func() {
		scorecardCache = NewTTLCache()
	})
	return scorecardCache
}
