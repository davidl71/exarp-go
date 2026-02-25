// Package cache provides thread-safe in-memory and file caching utilities for MCP servers.
package cache

import (
	"sync"
	"time"
)

// ttlEntry holds a value and its expiration time.
type ttlEntry struct {
	value     []byte
	expiresAt time.Time
}

// TTLCache is a thread-safe key-value cache with per-entry TTL.
// Keys are strings; values are byte slices. Expired entries are removed on Get.
type TTLCache struct {
	mu    sync.RWMutex
	store map[string]*ttlEntry
}

// NewTTLCache returns a new TTL cache.
func NewTTLCache() *TTLCache {
	return &TTLCache{store: make(map[string]*ttlEntry)}
}

// Get returns the value for key if present and not expired.
// If the entry is expired, it is removed and Get returns (nil, false).
func (c *TTLCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.store[key]
	if !ok || e == nil {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		delete(c.store, key)
		return nil, false
	}
	// Return a copy so caller cannot mutate the stored slice
	out := make([]byte, len(e.value))
	copy(out, e.value)
	return out, true
}

// Set stores value for key with the given TTL.
func (c *TTLCache) Set(key string, value []byte, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.store == nil {
		c.store = make(map[string]*ttlEntry)
	}
	valCopy := make([]byte, len(value))
	copy(valCopy, value)
	c.store[key] = &ttlEntry{value: valCopy, expiresAt: time.Now().Add(ttl)}
}

// Invalidate removes key from the cache.
func (c *TTLCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

// Clear removes all entries.
func (c *TTLCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]*ttlEntry)
}
