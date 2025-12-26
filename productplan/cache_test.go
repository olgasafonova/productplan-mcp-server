package productplan

import (
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	cache := NewCache(DefaultCacheConfig())

	// Set a value
	cache.Set("key1", "value1", time.Hour)

	// Get the value
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Get non-existent key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent key")
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(DefaultCacheConfig())

	// Set a value with very short TTL
	cache.Set("key1", "value1", time.Millisecond)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Should not find expired value
	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(DefaultCacheConfig())

	cache.Set("key1", "value1", time.Hour)
	cache.Delete("key1")

	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be deleted")
	}
}

func TestCache_InvalidatePrefix(t *testing.T) {
	cache := NewCache(DefaultCacheConfig())

	cache.Set("roadmap:1:bars", "bars1", time.Hour)
	cache.Set("roadmap:1:lanes", "lanes1", time.Hour)
	cache.Set("roadmap:2:bars", "bars2", time.Hour)
	cache.Set("objectives", "obj", time.Hour)

	// Invalidate roadmap:1 prefix
	count := cache.InvalidatePrefix("roadmap:1")

	if count != 2 {
		t.Errorf("Expected 2 invalidated, got %d", count)
	}

	// roadmap:1 keys should be gone
	if _, ok := cache.Get("roadmap:1:bars"); ok {
		t.Error("Expected roadmap:1:bars to be invalidated")
	}
	if _, ok := cache.Get("roadmap:1:lanes"); ok {
		t.Error("Expected roadmap:1:lanes to be invalidated")
	}

	// Other keys should remain
	if _, ok := cache.Get("roadmap:2:bars"); !ok {
		t.Error("Expected roadmap:2:bars to remain")
	}
	if _, ok := cache.Get("objectives"); !ok {
		t.Error("Expected objectives to remain")
	}
}

func TestCache_LRUEviction(t *testing.T) {
	config := CacheConfig{MaxEntries: 3}
	cache := NewCache(config)

	// Fill cache
	cache.Set("key1", "v1", time.Hour)
	cache.Set("key2", "v2", time.Hour)
	cache.Set("key3", "v3", time.Hour)

	// Access key1 to make it recently used
	cache.Get("key1")

	// Add key4, should evict key2 (oldest not recently used)
	cache.Set("key4", "v4", time.Hour)

	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}

	// key2 should be evicted
	if _, ok := cache.Get("key2"); ok {
		t.Error("Expected key2 to be evicted")
	}

	// key1, key3, key4 should remain
	if _, ok := cache.Get("key1"); !ok {
		t.Error("Expected key1 to remain")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(DefaultCacheConfig())

	cache.Set("key1", "v1", time.Hour)
	cache.Set("key2", "v2", time.Hour)

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected empty cache, got size %d", cache.Size())
	}
}

func TestCache_Stats(t *testing.T) {
	config := CacheConfig{MaxEntries: 50}
	cache := NewCache(config)

	cache.Set("key1", "v1", time.Hour)
	cache.Set("key2", "v2", time.Hour)

	stats := cache.Stats()

	if stats.Entries != 2 {
		t.Errorf("Expected 2 entries, got %d", stats.Entries)
	}
	if stats.MaxSize != 50 {
		t.Errorf("Expected max size 50, got %d", stats.MaxSize)
	}
}

func TestCacheKey(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		args      []string
		want      string
	}{
		{"no args", "list_roadmaps", nil, "list_roadmaps"},
		{"one arg", "get_roadmap", []string{"123"}, "get_roadmap:123"},
		{"two args", "get_bar_comments", []string{"123", "456"}, "get_bar_comments:123:456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CacheKey(tt.operation, tt.args...)
			if got != tt.want {
				t.Errorf("CacheKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
