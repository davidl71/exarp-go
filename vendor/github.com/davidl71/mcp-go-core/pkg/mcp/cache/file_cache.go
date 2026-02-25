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

// ReadFile reads a file, using cache if available and valid.
// Returns the file content and a boolean indicating if it was a cache hit.
func (fc *FileCache) ReadFile(path string) ([]byte, bool, error) {
	// Check cache first
	if cached, ok := fc.cache.Load(path); ok {
		cf := cached.(*cachedFile)

		// Check if cache entry has expired (TTL)
		if cf.ttl > 0 && time.Since(cf.cachedAt) > cf.ttl {
			fc.cache.Delete(path)
		} else {
			info, err := os.Stat(path)
			if err != nil {
				fc.cache.Delete(path)
				return nil, false, err
			}
			if info.ModTime().Equal(cf.mtime) || info.ModTime().Before(cf.mtime) {
				return cf.content, true, nil
			}
			fc.cache.Delete(path)
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}

	info, err := os.Stat(path)
	if err != nil {
		fc.cache.Store(path, &cachedFile{
			content:  content,
			cachedAt: time.Now(),
		})
		return content, false, nil
	}

	fc.cache.Store(path, &cachedFile{
		content:  content,
		mtime:    info.ModTime(),
		cachedAt: time.Now(),
		ttl:      0,
	})

	return content, false, nil
}

// ReadFileWithTTL reads a file with a TTL for cache expiration.
func (fc *FileCache) ReadFileWithTTL(path string, ttl time.Duration) ([]byte, bool, error) {
	if cached, ok := fc.cache.Load(path); ok {
		cf := cached.(*cachedFile)

		if time.Since(cf.cachedAt) > ttl {
			fc.cache.Delete(path)
		} else {
			info, err := os.Stat(path)
			if err != nil {
				fc.cache.Delete(path)
				return nil, false, err
			}
			if info.ModTime().Equal(cf.mtime) || info.ModTime().Before(cf.mtime) {
				return cf.content, true, nil
			}
			fc.cache.Delete(path)
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}

	info, err := os.Stat(path)
	if err != nil {
		fc.cache.Store(path, &cachedFile{
			content:  content,
			cachedAt: time.Now(),
			ttl:      ttl,
		})
		return content, false, nil
	}

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
