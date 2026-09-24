package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

// countingServer answers every GET with {"n": <request number>} after an
// optional delay, and counts requests per method.
type countingServer struct {
	*httptest.Server
	gets, writes atomic.Int32
}

func newCountingServer(t *testing.T, delay time.Duration) *countingServer {
	t.Helper()
	cs := &countingServer{}
	cs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			cs.writes.Add(1)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		n := cs.gets.Add(1)
		time.Sleep(delay)
		_, _ = w.Write([]byte(`{"n":` + strconv.Itoa(int(n)) + `}`))
	}))
	t.Cleanup(cs.Close)
	return cs
}

func cachedClient(t *testing.T, url string, ttl time.Duration) *Client {
	t.Helper()
	c, err := New(Config{Token: "t", BaseURL: url, Logger: logging.Nop(), CacheTTL: ttl})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func bodyN(t *testing.T, raw json.RawMessage) int {
	t.Helper()
	var v struct{ N int }
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("bad body %s: %v", raw, err)
	}
	return v.N
}

func TestCache_MissThenHit(t *testing.T) {
	srv := newCountingServer(t, 0)
	c := cachedClient(t, srv.URL, time.Minute)
	ctx := context.Background()

	a, err := c.get(ctx, "/roadmaps/1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.get(ctx, "/roadmaps/1")
	if err != nil {
		t.Fatal(err)
	}
	if bodyN(t, a) != 1 || bodyN(t, b) != 1 {
		t.Errorf("second GET should be served from cache: %s / %s", a, b)
	}
	if srv.gets.Load() != 1 {
		t.Errorf("upstream GETs = %d, want 1", srv.gets.Load())
	}
	st := c.CacheStats()
	if !st.Enabled || st.Hits != 1 || st.Misses != 1 || st.Entries != 1 {
		t.Errorf("stats = %+v, want 1 hit, 1 miss, 1 entry", st)
	}

	// The query string is part of the key.
	if _, err := c.get(ctx, "/roadmaps/1?q%5Bname_i_cont%5D=x"); err != nil {
		t.Fatal(err)
	}
	if srv.gets.Load() != 2 {
		t.Errorf("a different query must miss; upstream GETs = %d", srv.gets.Load())
	}
}

func TestCache_ReturnsIndependentCopies(t *testing.T) {
	srv := newCountingServer(t, 0)
	c := cachedClient(t, srv.URL, time.Minute)
	a, _ := c.get(context.Background(), "/x")
	a[0] = 'X' // a caller scribbling on its slice must not corrupt the cache
	b, _ := c.get(context.Background(), "/x")
	if b[0] != '{' {
		t.Errorf("cache entry was mutated through a returned slice: %s", b)
	}
}

func TestCache_Expiry(t *testing.T) {
	srv := newCountingServer(t, 0)
	c := cachedClient(t, srv.URL, time.Minute)
	var mu sync.Mutex
	now := time.Unix(1_000_000, 0)
	c.cache.now = func() time.Time { mu.Lock(); defer mu.Unlock(); return now }

	ctx := context.Background()
	if _, err := c.get(ctx, "/x"); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	now = now.Add(59 * time.Second)
	mu.Unlock()
	if _, err := c.get(ctx, "/x"); err != nil {
		t.Fatal(err)
	}
	if srv.gets.Load() != 1 {
		t.Fatalf("still within TTL; upstream GETs = %d", srv.gets.Load())
	}
	mu.Lock()
	now = now.Add(2 * time.Second)
	mu.Unlock()
	got, err := c.get(ctx, "/x")
	if err != nil {
		t.Fatal(err)
	}
	if bodyN(t, got) != 2 || srv.gets.Load() != 2 {
		t.Errorf("expired entry should refetch; body=%s upstream=%d", got, srv.gets.Load())
	}
}

func TestCache_InvalidatedByEveryWriteMethod(t *testing.T) {
	for _, write := range []struct {
		name string
		do   func(c *Client) error
	}{
		{"POST", func(c *Client) error {
			_, err := c.request(context.Background(), http.MethodPost, "/bars", map[string]any{"name": "n"})
			return err
		}},
		{"PATCH", func(c *Client) error {
			_, err := c.request(context.Background(), http.MethodPatch, "/bars/1", map[string]any{"name": "n"})
			return err
		}},
		{"DELETE", func(c *Client) error {
			_, err := c.request(context.Background(), http.MethodDelete, "/bars/1", nil)
			return err
		}},
	} {
		t.Run(write.name, func(t *testing.T) {
			srv := newCountingServer(t, 0)
			c := cachedClient(t, srv.URL, time.Minute)
			ctx := context.Background()
			if _, err := c.get(ctx, "/roadmaps/1/bars"); err != nil {
				t.Fatal(err)
			}
			if err := write.do(c); err != nil {
				t.Fatal(err)
			}
			got, err := c.get(ctx, "/roadmaps/1/bars")
			if err != nil {
				t.Fatal(err)
			}
			if bodyN(t, got) != 2 {
				t.Errorf("read after %s served stale cache: %s", write.name, got)
			}
			if st := c.CacheStats(); st.Invalidations != 1 {
				t.Errorf("invalidations = %d, want 1", st.Invalidations)
			}
		})
	}
}

func TestCache_FailedWriteStillInvalidates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"n":1}`))
	}))
	defer srv.Close()
	c := cachedClient(t, srv.URL, time.Minute)
	_, _ = c.get(context.Background(), "/x")
	if _, err := c.request(context.Background(), http.MethodDelete, "/bars/1", nil); err == nil {
		t.Fatal("expected write error")
	}
	if st := c.CacheStats(); st.Entries != 0 || st.Invalidations != 1 {
		t.Errorf("a failed write may still have landed upstream; stats = %+v", st)
	}
}

// A GET already in flight when a write happens must not store its
// (pre-write) result, or the next read would be stale for a full TTL.
func TestCache_InFlightReadDuringWriteIsNotStored(t *testing.T) {
	release := make(chan struct{})
	var gets atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		n := gets.Add(1)
		if n == 1 {
			<-release
		}
		_, _ = w.Write([]byte(`{"n":` + strconv.Itoa(int(n)) + `}`))
	}))
	defer srv.Close()
	c := cachedClient(t, srv.URL, time.Minute)
	ctx := context.Background()

	done := make(chan json.RawMessage)
	go func() {
		b, _ := c.get(ctx, "/x")
		done <- b
	}()
	for gets.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	if _, err := c.request(ctx, http.MethodPatch, "/bars/1", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	close(release)
	if first := <-done; bodyN(t, first) != 1 {
		t.Fatalf("in-flight caller should get its own response, got %s", first)
	}
	after, err := c.get(ctx, "/x")
	if err != nil {
		t.Fatal(err)
	}
	if bodyN(t, after) != 2 {
		t.Errorf("pre-write in-flight result was cached: %s", after)
	}
}

func TestCache_ConcurrentIdenticalRequestsShareOneFetch(t *testing.T) {
	srv := newCountingServer(t, 100*time.Millisecond)
	c := cachedClient(t, srv.URL, time.Minute)

	const callers = 20
	var wg sync.WaitGroup
	results := make([]json.RawMessage, callers)
	errs := make([]error, callers)
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = c.get(context.Background(), "/roadmaps/9/lanes")
		}()
	}
	wg.Wait()
	for i := range results {
		if errs[i] != nil {
			t.Fatalf("caller %d: %v", i, errs[i])
		}
		if bodyN(t, results[i]) != 1 {
			t.Errorf("caller %d got %s", i, results[i])
		}
	}
	if srv.gets.Load() != 1 {
		t.Errorf("upstream GETs = %d, want 1 (singleflight)", srv.gets.Load())
	}
	st := c.CacheStats()
	if st.Misses != callers || st.Coalesced != callers-1 {
		t.Errorf("stats = %+v, want %d misses, %d coalesced", st, callers, callers-1)
	}
}

// One waiter giving up must not fail the other waiters on the same flight.
func TestCache_CancelledWaiterDoesNotFailOthers(t *testing.T) {
	srv := newCountingServer(t, 100*time.Millisecond)
	c := cachedClient(t, srv.URL, time.Minute)

	short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	var wg sync.WaitGroup
	var shortErr, longErr error
	var longBody json.RawMessage
	wg.Add(2)
	go func() { defer wg.Done(); _, shortErr = c.get(short, "/x") }()
	go func() { defer wg.Done(); longBody, longErr = c.get(context.Background(), "/x") }()
	wg.Wait()

	if !errors.Is(shortErr, context.DeadlineExceeded) {
		t.Errorf("short caller err = %v, want deadline exceeded", shortErr)
	}
	if longErr != nil || bodyN(t, longBody) != 1 {
		t.Errorf("long caller got %s, %v", longBody, longErr)
	}
}

func TestCache_ErrorsAreNotCached(t *testing.T) {
	var gets atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gets.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"n":2}`))
	}))
	defer srv.Close()
	c := cachedClient(t, srv.URL, time.Minute)
	if _, err := c.get(context.Background(), "/x"); err == nil {
		t.Fatal("expected first call to fail")
	}
	got, err := c.get(context.Background(), "/x")
	if err != nil || bodyN(t, got) != 2 {
		t.Errorf("retry after an error should refetch, got %s, %v", got, err)
	}
}

func TestCache_DisabledByZeroTTL(t *testing.T) {
	srv := newCountingServer(t, 0)
	c := cachedClient(t, srv.URL, 0)
	_, _ = c.get(context.Background(), "/x")
	_, _ = c.get(context.Background(), "/x")
	if srv.gets.Load() != 2 {
		t.Errorf("TTL 0 must disable caching; upstream GETs = %d", srv.gets.Load())
	}
	if st := c.CacheStats(); st.Enabled {
		t.Errorf("stats = %+v, want disabled", st)
	}
}

func TestCache_StatusProbeBypassesCache(t *testing.T) {
	srv := newCountingServer(t, 0)
	c := cachedClient(t, srv.URL, time.Minute)
	_, _ = c.CheckStatus(context.Background())
	_, _ = c.CheckStatus(context.Background())
	if srv.gets.Load() != 2 {
		t.Errorf("check_status must always hit the API; upstream GETs = %d", srv.gets.Load())
	}
}

func TestCache_BoundedEntries(t *testing.T) {
	rc := newReadCache(time.Minute)
	for i := 0; i < maxCacheEntries+10; i++ {
		rc.store("/k"+strconv.Itoa(i), 0, []byte("{}"))
	}
	if n := rc.stats().Entries; n > maxCacheEntries {
		t.Errorf("entries = %d, want <= %d", n, maxCacheEntries)
	}
}

func TestCacheTTLFromEnv(t *testing.T) {
	cases := []struct {
		val     string
		want    time.Duration
		wantErr bool
	}{
		{"", DefaultCacheTTL, false},
		{"0", 0, false},
		{"30", 30 * time.Second, false},
		{"2m", 2 * time.Minute, false},
		{"-5", 0, true},
		{"-1s", 0, true},
		{"soon", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.val, func(t *testing.T) {
			t.Setenv(CacheTTLEnv, tc.val)
			got, err := CacheTTLFromEnv()
			if (err != nil) != tc.wantErr || got != tc.want {
				t.Errorf("CacheTTLFromEnv(%q) = %v, %v; want %v, err=%v", tc.val, got, err, tc.want, tc.wantErr)
			}
		})
	}
}
