package cache

import (
	"os"
	"sync"
	"time"
)

type cachedFile struct {
	content  []byte
	mtime    time.Time
	cachedAt time.Time
	ttl      time.Duration
}

// FileCache provides thread-safe file caching with mtime-based invalidation.
type FileCache struct {
	cache sync.Map
}

// NewFileCache creates a new file cache instance.
func NewFileCache() *FileCache {
	return &FileCache{}
}

// ReadFile reads a file, using cache if available and valid.
func (fc *FileCache) ReadFile(path string) ([]byte, bool, error) {
	if cached, ok := fc.cache.Load(path); ok {
		cf := cached.(*cachedFile)
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
		fc.cache.Store(path, &cachedFile{content: content, cachedAt: time.Now()})
		return content, false, nil
	}

	fc.cache.Store(path, &cachedFile{
		content:  content,
		mtime:    info.ModTime(),
		cachedAt: time.Now(),
	})

	return content, false, nil
}

var (
	globalFileCache     *FileCache
	globalFileCacheOnce sync.Once
)

// GetGlobalFileCache returns the global file cache instance.
func GetGlobalFileCache() *FileCache {
	globalFileCacheOnce.Do(func() {
		globalFileCache = NewFileCache()
	})
	return globalFileCache
}
