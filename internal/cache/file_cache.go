// file_cache.go — Thread-safe file cache with mtime-based invalidation.
//
// Package cache provides file-based and TTL in-memory caching for scorecard, reports, and other data.
package cache

import (
	"os"
	"sync"
	"time"
)

// cachedFile represents a cached file with its content and metadata.
type cachedFile struct {
	content  []byte
	mtime    time.Time
	cachedAt time.Time
	ttl      time.Duration
}

// FileCache provides thread-safe file caching with mtime-based invalidation.
type FileCache struct {
	cache sync.Map // map[string]*cachedFile
	mu    sync.RWMutex
}

// NewFileCache creates a new file cache instance.
func NewFileCache() *FileCache {
	return &FileCache{}
}

// ReadFile reads a file, using cache if available and valid
// Returns the file content and a boolean indicating if it was a cache hit.
func (fc *FileCache) ReadFile(path string) ([]byte, bool, error) {
	// Check cache first
	if cached, ok := fc.cache.Load(path); ok {
		cf := cached.(*cachedFile)

		// Check if cache entry has expired (TTL)
		if cf.ttl > 0 && time.Since(cf.cachedAt) > cf.ttl {
			// TTL expired, remove from cache
			fc.cache.Delete(path)
		} else {
			// Check file modification time
			info, err := os.Stat(path)
			if err != nil {
				// File doesn't exist or error - remove from cache
				fc.cache.Delete(path)
				return nil, false, err
			}

			// Compare mtime - if file hasn't changed, return cached content
			if info.ModTime().Equal(cf.mtime) || info.ModTime().Before(cf.mtime) {
				return cf.content, true, nil
			}

			// File has been modified, remove from cache
			fc.cache.Delete(path)
		}
	}

	// Cache miss or invalid - read file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}

	// Get file info for mtime
	info, err := os.Stat(path)
	if err != nil {
		// If stat fails after read, still cache the content
		// (file might have been deleted, but we have the content)
		fc.cache.Store(path, &cachedFile{
			content:  content,
			cachedAt: time.Now(),
		})

		return content, false, nil
	}

	// Store in cache
	fc.cache.Store(path, &cachedFile{
		content:  content,
		mtime:    info.ModTime(),
		cachedAt: time.Now(),
		ttl:      0, // No TTL by default (mtime-based invalidation)
	})

	return content, false, nil
}

// ReadFileWithTTL reads a file with a TTL (Time-To-Live) for cache expiration
// The cache entry will expire after the specified TTL, regardless of mtime.
func (fc *FileCache) ReadFileWithTTL(path string, ttl time.Duration) ([]byte, bool, error) {
	// Check cache first
	if cached, ok := fc.cache.Load(path); ok {
		cf := cached.(*cachedFile)

		// Check TTL expiration
		if time.Since(cf.cachedAt) > ttl {
			// TTL expired, remove from cache
			fc.cache.Delete(path)
		} else {
			// Check file modification time
			info, err := os.Stat(path)
			if err != nil {
				// File doesn't exist or error - remove from cache
				fc.cache.Delete(path)
				return nil, false, err
			}

			// Compare mtime - if file hasn't changed, return cached content
			if info.ModTime().Equal(cf.mtime) || info.ModTime().Before(cf.mtime) {
				return cf.content, true, nil
			}

			// File has been modified, remove from cache
			fc.cache.Delete(path)
		}
	}

	// Cache miss or invalid - read file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}

	// Get file info for mtime
	info, err := os.Stat(path)
	if err != nil {
		// If stat fails after read, still cache the content with TTL
		fc.cache.Store(path, &cachedFile{
			content:  content,
			cachedAt: time.Now(),
			ttl:      ttl,
		})

		return content, false, nil
	}

	// Store in cache with TTL
	fc.cache.Store(path, &cachedFile{
		content:  content,
		mtime:    info.ModTime(),
		cachedAt: time.Now(),
		ttl:      ttl,
	})

	return content, false, nil
}

// Invalidate removes a file from the cache.
func (fc *FileCache) Invalidate(path string) {
	fc.cache.Delete(path)
}

// Clear removes all entries from the cache.
func (fc *FileCache) Clear() {
	fc.cache.Range(func(key, value interface{}) bool {
		fc.cache.Delete(key)
		return true
	})
}

// GetStats returns cache statistics.
func (fc *FileCache) GetStats() map[string]interface{} {
	var count int

	var totalSize int64

	fc.cache.Range(func(key, value interface{}) bool {
		count++
		cf := value.(*cachedFile)
		totalSize += int64(len(cf.content))

		return true
	})

	return map[string]interface{}{
		"entries":          count,
		"total_size_bytes": totalSize,
	}
}

// Global file cache instance (optional - can be used as singleton).
var globalFileCache *FileCache
var globalFileCacheOnce sync.Once

// GetGlobalFileCache returns the global file cache instance (singleton).
func GetGlobalFileCache() *FileCache {
	globalFileCacheOnce.Do(func() {
		globalFileCache = NewFileCache()
	})

	return globalFileCache
}
