package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

// slowRoadmapServer answers the four endpoints get_roadmap_complete needs,
// sleeping delay per request, and counts requests per path.
type slowRoadmapServer struct {
	*httptest.Server
	mu   sync.Mutex
	hits map[string]int
}

func newSlowRoadmapServer(tb testing.TB, delay time.Duration) *slowRoadmapServer {
	tb.Helper()
	s := &slowRoadmapServer{hits: map[string]int{}}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.hits[r.URL.Path]++
		s.mu.Unlock()
		time.Sleep(delay)
		switch {
		case strings.HasSuffix(r.URL.Path, "/bars"):
			_, _ = w.Write([]byte(`{"results":[{"id":1,"name":"A","lane":"Core"}],"paging":{"page_count":1}}`))
		case strings.HasSuffix(r.URL.Path, "/lanes"):
			_, _ = w.Write([]byte(`{"results":[{"id":9,"name":"Core"}],"paging":{"page_count":1}}`))
		case strings.HasSuffix(r.URL.Path, "/milestones"):
			_, _ = w.Write([]byte(`{"results":[{"id":3,"title":"GA","date":"2026-10-01"}],"paging":{"page_count":1}}`))
		default:
			_, _ = w.Write([]byte(`{"id":5,"name":"Roadmap"}`))
		}
	}))
	tb.Cleanup(s.Close)
	return s
}

func (s *slowRoadmapServer) total() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, v := range s.hits {
		n += v
	}
	return n
}

func completeClient(tb testing.TB, url string, ttl time.Duration) *api.Client {
	tb.Helper()
	c, err := api.New(api.Config{Token: "t", BaseURL: url, CacheTTL: ttl})
	if err != nil {
		tb.Fatal(err)
	}
	return c
}

// sequentialComplete is the naive baseline: the same five upstream calls
// get_roadmap_complete needs (roadmap, bars, lanes-for-bars, lanes,
// milestones), one after another.
func sequentialComplete(ctx context.Context, c *api.Client, id string) error {
	if _, err := c.GetRoadmap(ctx, id); err != nil {
		return err
	}
	if _, err := c.GetList(ctx, "/roadmaps/"+id+"/bars", api.Query{}); err != nil {
		return err
	}
	if _, err := c.GetList(ctx, "/roadmaps/"+id+"/lanes", api.Query{}); err != nil {
		return err
	}
	if _, err := c.GetRoadmapLanes(ctx, id); err != nil {
		return err
	}
	_, err := c.GetRoadmapMilestones(ctx, id)
	return err
}

func TestGetRoadmapComplete_ParallelAndDeduplicated(t *testing.T) {
	const delay = 100 * time.Millisecond
	srv := newSlowRoadmapServer(t, delay)
	h := getRoadmapCompleteHandler(completeClient(t, srv.URL, time.Minute))

	start := time.Now()
	out, err := h.Handle(context.Background(), map[string]any{"roadmap_id": "5"})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	// Sequentially this is 5 x 100ms; in parallel it is one round trip.
	if elapsed >= 2*delay {
		t.Errorf("get_roadmap_complete took %v; want about one %v round trip", elapsed, delay)
	}
	// bars enrichment and the lanes section both need /lanes; singleflight
	// collapses them to one upstream call.
	srv.mu.Lock()
	lanesHits := srv.hits["/roadmaps/5/lanes"]
	srv.mu.Unlock()
	if lanesHits != 1 || srv.total() != 4 {
		t.Errorf("upstream calls = %d (lanes %d), want 4 (lanes 1)", srv.total(), lanesHits)
	}

	var resp FormattedResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Summary != "Roadmap 5 retrieved" {
		t.Errorf("summary = %q", resp.Summary)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"roadmap", "bars", "lanes", "milestones"} {
		if len(data[k]) == 0 {
			t.Errorf("section %s missing", k)
		}
	}
	if !strings.Contains(string(data["bars"]), `"lane_id":9`) {
		t.Errorf("bars not enriched with lane_id: %s", data["bars"])
	}

	// A second call inside the TTL is served from memory.
	if _, err := h.Handle(context.Background(), map[string]any{"roadmap_id": "5"}); err != nil {
		t.Fatal(err)
	}
	if srv.total() != 4 {
		t.Errorf("second call went upstream: total = %d", srv.total())
	}
}

// BenchmarkGetRoadmapComplete compares the handler with a sequential
// baseline against a server that sleeps 20ms per request. The cache is off
// in both so every iteration pays the network cost.
//
//	go test ./internal/tools -run '^$' -bench GetRoadmapComplete -benchtime 20x
func BenchmarkGetRoadmapComplete(b *testing.B) {
	const delay = 20 * time.Millisecond
	ctx := context.Background()

	b.Run("sequential_baseline", func(b *testing.B) {
		srv := newSlowRoadmapServer(b, delay)
		c := completeClient(b, srv.URL, 0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := sequentialComplete(ctx, c, "5"); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("parallel_handler", func(b *testing.B) {
		srv := newSlowRoadmapServer(b, delay)
		h := getRoadmapCompleteHandler(completeClient(b, srv.URL, 0))
		args := map[string]any{"roadmap_id": "5"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := h.Handle(ctx, args); err != nil {
				b.Fatal(err)
			}
		}
	})
}
