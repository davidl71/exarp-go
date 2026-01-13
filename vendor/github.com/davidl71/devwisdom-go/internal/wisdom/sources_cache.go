package wisdom

import (
	"os"
	"sync"
	"time"
)

// CacheEntry represents a cached source configuration
type CacheEntry struct {
	Config      *SourceConfig
	LoadedAt    time.Time
	FileModTime time.Time
	FilePath    string
	TTL         time.Duration
}

// SourceCache manages caching of source configurations
type SourceCache struct {
	mu         sync.RWMutex
	entries    map[string]*CacheEntry
	defaultTTL time.Duration
	maxAge     time.Duration
	enabled    bool
}

// NewSourceCache creates a new source cache
func NewSourceCache() *SourceCache {
	return &SourceCache{
		entries:    make(map[string]*CacheEntry),
		defaultTTL: 5 * time.Minute, // Default cache TTL
		maxAge:     1 * time.Hour,   // Maximum cache age
		enabled:    true,
	}
}

// WithTTL sets the default cache TTL
func (sc *SourceCache) WithTTL(ttl time.Duration) *SourceCache {
	sc.defaultTTL = ttl
	return sc
}

// WithMaxAge sets the maximum cache age
func (sc *SourceCache) WithMaxAge(maxAge time.Duration) *SourceCache {
	sc.maxAge = maxAge
	return sc
}

// Enable enables or disables caching
func (sc *SourceCache) Enable(enabled bool) *SourceCache {
	sc.enabled = enabled
	return sc
}

// Get retrieves a cached entry if valid
func (sc *SourceCache) Get(key string) (*SourceConfig, bool) {
	if !sc.enabled {
		return nil, false
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	entry, exists := sc.entries[key]
	if !exists {
		return nil, false
	}

	// Check if entry is still valid
	if !sc.isValid(entry) {
		// Entry expired, remove it
		sc.mu.RUnlock()
		sc.mu.Lock()
		delete(sc.entries, key)
		sc.mu.Unlock()
		sc.mu.RLock()
		return nil, false
	}

	// Check if file has been modified
	if entry.FilePath != "" {
		if modified, err := sc.isFileModified(entry.FilePath, entry.FileModTime); err == nil && modified {
			// File was modified, invalidate cache
			sc.mu.RUnlock()
			sc.mu.Lock()
			delete(sc.entries, key)
			sc.mu.Unlock()
			sc.mu.RLock()
			return nil, false
		}
	}

	return entry.Config, true
}

// Set stores a configuration in the cache
func (sc *SourceCache) Set(key string, config *SourceConfig, filePath string) {
	if !sc.enabled {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	var modTime time.Time
	if filePath != "" {
		if info, err := os.Stat(filePath); err == nil {
			modTime = info.ModTime()
		}
	}

	sc.entries[key] = &CacheEntry{
		Config:      config,
		LoadedAt:    time.Now(),
		FileModTime: modTime,
		FilePath:    filePath,
		TTL:         sc.defaultTTL,
	}
}

// Invalidate removes an entry from the cache
func (sc *SourceCache) Invalidate(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.entries, key)
}

// InvalidateAll clears the entire cache
func (sc *SourceCache) InvalidateAll() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.entries = make(map[string]*CacheEntry)
}

// ClearExpired removes expired entries from the cache
func (sc *SourceCache) ClearExpired() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	count := 0
	for key, entry := range sc.entries {
		if !sc.isValid(entry) {
			delete(sc.entries, key)
			count++
		}
	}

	return count
}

// Size returns the number of cached entries
func (sc *SourceCache) Size() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return len(sc.entries)
}

// isValid checks if a cache entry is still valid
func (sc *SourceCache) isValid(entry *CacheEntry) bool {
	age := time.Since(entry.LoadedAt)
	return age < entry.TTL && age < sc.maxAge
}

// isFileModified checks if a file has been modified since the cached time
func (sc *SourceCache) isFileModified(filePath string, cachedModTime time.Time) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	return info.ModTime().After(cachedModTime), nil
}

// StartCleanup starts a background goroutine to periodically clean expired entries
func (sc *SourceCache) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			sc.ClearExpired()
		}
	}()
}
