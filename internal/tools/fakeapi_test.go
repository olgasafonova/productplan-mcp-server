package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

// recordedRequest is one request the fake ProductPlan API received.
type recordedRequest struct {
	Method string
	Path   string
	Body   map[string]any
}

// fakeAPI models the bar write contract probed live on 24-09-2026:
// POST /bars answers 201 {"location":"/api/v2/bars/<id>"}, PATCH answers
// 204, roadmaps embed legends/lanes as name strings.
type fakeAPI struct {
	mu        sync.Mutex
	requests  []recordedRequest
	roadmaps  map[string]map[string]any
	bars      map[string]map[string]any
	nextID    int
	failPatch map[string]int  // bar ID -> status to answer PATCH with
	keepOnDel map[string]bool // DELETE answers 204 but the bar survives
	delay     time.Duration

	inFlight    atomic.Int32
	maxInFlight atomic.Int32
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{
		roadmaps: map[string]map[string]any{
			"100": {
				"id":                 100,
				"name":               "Integrations",
				"legends":            []any{"Committed", "Exploring", "Blocked"},
				"lanes":              []any{"Backend", "Mobile"},
				"custom_text_fields": []any{map[string]any{"label": "Owner"}},
				"custom_dropdown_fields": []any{map[string]any{
					"label": "Objectives", "allowed_values": []any{"Grow", "Retain"},
				}},
			},
		},
		bars:      map[string]map[string]any{},
		nextID:    5000,
		failPatch: map[string]int{},
		keepOnDel: map[string]bool{},
	}
}

// addBar seeds a bar on roadmap 100.
func (f *fakeAPI) addBar(id string, fields map[string]any) {
	bar := map[string]any{"id": id, "name": "Bar " + id, "roadmap": map[string]any{"id": 100}}
	for k, v := range fields {
		bar[k] = v
	}
	f.bars[id] = bar
}

func (f *fakeAPI) start(t *testing.T) *api.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(server.Close)
	client, err := api.New(api.Config{Token: "test-token", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

func (f *fakeAPI) serve(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body) // empty bodies are expected on GET/DELETE
	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{Method: r.Method, Path: r.URL.Path, Body: body})
	f.mu.Unlock()

	if r.Method != http.MethodGet {
		n := f.inFlight.Add(1)
		defer f.inFlight.Add(-1)
		for {
			m := f.maxInFlight.Load()
			if n <= m || f.maxInFlight.CompareAndSwap(m, n) {
				break
			}
		}
		if f.delay > 0 {
			time.Sleep(f.delay)
		}
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	switch {
	case r.Method == http.MethodGet && len(parts) == 2 && parts[0] == "roadmaps":
		writeFakeObject(w, f.roadmaps[parts[1]])
	case r.Method == http.MethodGet && len(parts) == 2 && parts[0] == "bars":
		writeFakeObject(w, f.bars[parts[1]])
	case r.Method == http.MethodPost && r.URL.Path == "/bars":
		f.nextID++
		id := fmt.Sprint(f.nextID)
		f.bars[id] = map[string]any{"id": id, "name": body["name"]}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"location": "/api/v2/bars/" + id})
	case r.Method == http.MethodPatch && len(parts) == 2:
		if status := f.failPatch[parts[1]]; status != 0 {
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "rejected"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodDelete && len(parts) == 2:
		if _, ok := f.bars[parts[1]]; !ok {
			http.NotFound(w, r)
			return
		}
		if !f.keepOnDel[parts[1]] {
			delete(f.bars, parts[1])
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.NotFound(w, r)
	}
}

func writeFakeObject(w http.ResponseWriter, obj map[string]any) {
	if obj == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
		return
	}
	_ = json.NewEncoder(w).Encode(obj)
}

// writes returns the non-GET requests received.
func (f *fakeAPI) writes() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []recordedRequest
	for _, r := range f.requests {
		if r.Method != http.MethodGet {
			out = append(out, r)
		}
	}
	return out
}

// countGets returns how many GETs hit path.
func (f *fakeAPI) countGets(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, r := range f.requests {
		if r.Method == http.MethodGet && r.Path == path {
			n++
		}
	}
	return n
}
