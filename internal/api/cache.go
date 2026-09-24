package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	// DefaultCacheTTL is how long an idempotent GET is served from memory.
	// ProductPlan data changes at human speed; 60s absorbs the bursts an
	// agent produces (get_roadmap_complete, then get_roadmap_bars, then
	// get_roadmap_lanes for the same roadmap) without serving anything
	// meaningfully stale.
	DefaultCacheTTL = 60 * time.Second

	// CacheTTLEnv overrides DefaultCacheTTL. It accepts a Go duration
	// ("90s", "2m") or an integer number of seconds; 0 disables the cache.
	CacheTTLEnv = "PRODUCTPLAN_CACHE_TTL"

	// maxCacheEntries bounds memory. When a store would exceed it, expired
	// entries are swept first and, if that is not enough, the cache is
	// cleared. Crude, but correct and O(1) amortised for this workload.
	maxCacheEntries = 1024
)

// CacheTTLFromEnv reads CacheTTLEnv. Unset or empty yields DefaultCacheTTL.
// A malformed or negative value is an error, so a typo fails loudly at
// startup instead of silently running with a different cache policy.
func CacheTTLFromEnv() (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(CacheTTLEnv))
	if raw == "" {
		return DefaultCacheTTL, nil
	}
	if secs, err := strconv.Atoi(raw); err == nil {
		if secs < 0 {
			return 0, fmt.Errorf("%s must not be negative, got %q", CacheTTLEnv, raw)
		}
		return time.Duration(secs) * time.Second, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration like 60s or a number of seconds (0 disables), got %q", CacheTTLEnv, raw)
	}
	if d < 0 {
		return 0, fmt.Errorf("%s must not be negative, got %q", CacheTTLEnv, raw)
	}
	return d, nil
}

// CacheStats is the read-cache counter snapshot reported by health_check.
type CacheStats struct {
	Enabled       bool    `json:"enabled"`
	TTLSeconds    float64 `json:"ttl_seconds"`
	Entries       int     `json:"entries"`
	Hits          uint64  `json:"hits"`
	Misses        uint64  `json:"misses"`
	Coalesced     uint64  `json:"coalesced"`
	Invalidations uint64  `json:"invalidations"`
}

type cacheEntry struct {
	data    []byte
	expires time.Time
}

// readCache is an in-process TTL cache for idempotent GETs, keyed by
// path+query, with singleflight so concurrent identical requests share one
// upstream call.
//
// Invalidation rule: any non-GET request clears the whole cache. ProductPlan
// writes have cross-resource effects (a bar write changes the roadmap's bar
// list, its lane's contents, child counts on its parent, connection counts
// on other bars), so prefix-scoped invalidation would have to model those
// relationships and would be wrong the first time it missed one. Clearing
// everything is always correct and costs one cold fetch per key afterwards.
//
// Writes also bump a generation counter. A GET that was already in flight
// when the write happened carries the old generation, so its (possibly
// pre-write) result is returned to its own callers but never stored, and a
// GET issued after the write never joins the pre-write flight because the
// generation is part of the singleflight key.
type readCache struct {
	ttl time.Duration
	now func() time.Time

	mu      sync.Mutex
	entries map[string]cacheEntry
	gen     uint64

	group singleflight.Group

	hits, misses, fetches, invalidations atomic.Uint64
}

// newReadCache returns nil (caching disabled) for a non-positive TTL; every
// method is nil-safe.
func newReadCache(ttl time.Duration) *readCache {
	if ttl <= 0 {
		return nil
	}
	return &readCache{ttl: ttl, now: time.Now, entries: make(map[string]cacheEntry)}
}

// get returns the cached body for key or calls fetch once for all
// concurrent callers of the same key. The shared fetch runs detached from
// any single caller's cancellation (bounded by the HTTP client timeout), so
// one caller giving up does not fail the others; each caller still returns
// promptly when its own context is done.
func (c *readCache) get(ctx context.Context, key string, fetch func(context.Context) (json.RawMessage, error)) (json.RawMessage, error) {
	if c == nil {
		return fetch(ctx)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.mu.Lock()
	if e, ok := c.entries[key]; ok && c.now().Before(e.expires) {
		c.mu.Unlock()
		c.hits.Add(1)
		return cloneBytes(e.data), nil
	}
	gen := c.gen
	c.mu.Unlock()
	c.misses.Add(1)

	flightKey := strconv.FormatUint(gen, 10) + "|" + key
	ch := c.group.DoChan(flightKey, func() (any, error) {
		c.fetches.Add(1)
		data, err := fetch(context.WithoutCancel(ctx))
		if err != nil {
			return nil, err // errors are never cached
		}
		c.store(key, gen, data)
		return []byte(data), nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		if r.Err != nil {
			return nil, r.Err
		}
		data, ok := r.Val.([]byte)
		if !ok {
			return nil, fmt.Errorf("read cache: unexpected value type %T", r.Val)
		}
		return cloneBytes(data), nil
	}
}

// store records data under key unless a write has happened since the fetch
// began (generation moved on).
func (c *readCache) store(key string, gen uint64, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.gen != gen {
		return
	}
	now := c.now()
	if len(c.entries) >= maxCacheEntries {
		for k, e := range c.entries {
			if !now.Before(e.expires) {
				delete(c.entries, k)
			}
		}
		if len(c.entries) >= maxCacheEntries {
			c.entries = make(map[string]cacheEntry)
		}
	}
	c.entries[key] = cacheEntry{data: cloneBytes(data), expires: now.Add(c.ttl)}
}

// invalidate drops every entry and moves the generation on.
func (c *readCache) invalidate() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.gen++
	c.entries = make(map[string]cacheEntry)
	c.mu.Unlock()
	c.invalidations.Add(1)
}

// stats returns a counter snapshot.
func (c *readCache) stats() CacheStats {
	if c == nil {
		return CacheStats{}
	}
	c.mu.Lock()
	n := len(c.entries)
	c.mu.Unlock()
	misses, fetches := c.misses.Load(), c.fetches.Load()
	coalesced := uint64(0)
	if misses > fetches {
		coalesced = misses - fetches
	}
	return CacheStats{
		Enabled:       true,
		TTLSeconds:    c.ttl.Seconds(),
		Entries:       n,
		Hits:          c.hits.Load(),
		Misses:        misses,
		Coalesced:     coalesced,
		Invalidations: c.invalidations.Load(),
	}
}

func cloneBytes(b []byte) []byte {
	return append([]byte(nil), b...)
}
