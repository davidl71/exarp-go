package sefaria

import (
	"sync"
	"time"
)

// CacheEntry represents a cached API response
type CacheEntry struct {
	Response  *TextResponse
	Timestamp time.Time
	TTL       time.Duration
}

// Cache provides thread-safe caching for Sefaria API responses
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration // Default TTL: 24 hours
}

// NewCache creates a new Sefaria API response cache
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     24 * time.Hour, // 24-hour TTL for Hebrew texts
	}
}

// Get retrieves a cached response if it exists and is not expired
func (c *Cache) Get(key string) (*TextResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if entry is expired
	if time.Since(entry.Timestamp) > entry.TTL {
		// Entry expired, but don't delete here (cleanup happens elsewhere)
		return nil, false
	}

	return entry.Response, true
}

// Set stores a response in the cache
func (c *Cache) Set(key string, response *TextResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Response:  response,
		Timestamp: time.Now(),
		TTL:       c.ttl,
	}
}

// Invalidate removes a specific cache entry
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// InvalidateAll clears all cache entries
func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Cleanup removes expired entries from the cache
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(c.entries, key)
		}
	}
}

// SetTTL sets the default TTL for cache entries
func (c *Cache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = ttl
}
