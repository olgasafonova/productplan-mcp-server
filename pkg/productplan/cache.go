package productplan

import (
	"sync"
	"time"
)

// CacheEntry represents a cached item with TTL.
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired returns true if the entry has expired.
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache provides simple in-memory caching with TTL and LRU eviction.
type Cache struct {
	entries    map[string]*CacheEntry
	order      []string // LRU order (oldest first)
	maxEntries int
	mu         sync.RWMutex
}

// CacheConfig configures the cache.
type CacheConfig struct {
	MaxEntries int
}

// DefaultCacheConfig returns sensible defaults.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxEntries: 100,
	}
}

// NewCache creates a new cache with the given config.
func NewCache(config CacheConfig) *Cache {
	return &Cache{
		entries:    make(map[string]*CacheEntry),
		order:      make([]string, 0, config.MaxEntries),
		maxEntries: config.MaxEntries,
	}
}

// Get retrieves a value from the cache.
// Returns the value and true if found and not expired.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		c.Delete(key)
		return nil, false
	}

	// Move to end of LRU order (most recently used)
	c.mu.Lock()
	c.moveToEnd(key)
	c.mu.Unlock()

	return entry.Data, true
}

// Set stores a value in the cache with the given TTL.
func (c *Cache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if _, exists := c.entries[key]; exists {
		// Update existing entry
		c.entries[key] = &CacheEntry{
			Data:      data,
			ExpiresAt: time.Now().Add(ttl),
		}
		c.moveToEnd(key)
		return
	}

	// Evict oldest entries if at capacity
	for len(c.entries) >= c.maxEntries && len(c.order) > 0 {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.entries, oldest)
	}

	// Add new entry
	c.entries[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
	c.order = append(c.order, key)
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	c.removeFromOrder(key)
}

// InvalidatePrefix removes all keys with the given prefix.
func (c *Cache) InvalidatePrefix(prefix string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	keysToRemove := make([]string, 0)

	for key := range c.entries {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keysToRemove = append(keysToRemove, key)
		}
	}

	for _, key := range keysToRemove {
		delete(c.entries, key)
		c.removeFromOrder(key)
		count++
	}

	return count
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.order = make([]string, 0, c.maxEntries)
}

// Size returns the number of entries in the cache.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Stats returns cache statistics.
type CacheStats struct {
	Entries int
	MaxSize int
}

// Stats returns current cache statistics.
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Entries: len(c.entries),
		MaxSize: c.maxEntries,
	}
}

// Helper: move key to end of order slice (most recently used)
func (c *Cache) moveToEnd(key string) {
	c.removeFromOrder(key)
	c.order = append(c.order, key)
}

// Helper: remove key from order slice
func (c *Cache) removeFromOrder(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}

// CacheKey generates a cache key from operation and arguments.
func CacheKey(operation string, args ...string) string {
	key := operation
	for _, arg := range args {
		key += ":" + arg
	}
	return key
}
