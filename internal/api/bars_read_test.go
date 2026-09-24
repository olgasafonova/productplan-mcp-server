package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetRoadmapBars_LanesFailureKeepsBarsAndWarns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/lanes") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"errors":["lanes down"]}`))
			return
		}
		_, _ = w.Write([]byte(`{"results":[{"id":1,"name":"A","lane":"Backend"}],"paging":{"page_count":1}}`))
	}))
	defer srv.Close()

	out, err := testClient(t, srv).GetRoadmapBars(context.Background(), "5")
	if err != nil {
		t.Fatalf("a lanes failure must not fail the bars call: %v", err)
	}
	var parsed struct {
		Count    int              `json:"count"`
		Warnings []string         `json:"warnings"`
		Bars     []map[string]any `json:"bars"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Count != 1 || parsed.Bars[0]["lane_name"] != "Backend" {
		t.Errorf("bars lost: %s", out)
	}
	if len(parsed.Warnings) != 1 || !strings.Contains(parsed.Warnings[0], "lane lookup failed") {
		t.Errorf("warnings = %v, want the lanes failure surfaced", parsed.Warnings)
	}
}

func TestGetRoadmapBars_BarsFailureFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/bars") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	if _, err := testClient(t, srv).GetRoadmapBars(context.Background(), "5"); err == nil {
		t.Fatal("expected error when bars fetch fails")
	}
}

// Bars and lanes are fetched in parallel: two 150ms requests should take
// about 150ms, not 300ms.
func TestGetRoadmapBars_FetchesBarsAndLanesConcurrently(t *testing.T) {
	const delay = 150 * time.Millisecond
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	start := time.Now()
	if _, err := testClient(t, srv).GetRoadmapBars(context.Background(), "5"); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed >= 2*delay {
		t.Errorf("took %v; bars and lanes appear to be fetched sequentially", elapsed)
	}
}

// TestFormatBars_LiveShape pins the documented GET /roadmaps/{id}/bars shape:
// a {results, paging} envelope, lane and legend as name strings, tags as a
// string array, starts_on/ends_on dates, and no lane_id.
func TestFormatBars_LiveShape(t *testing.T) {
	bars := `{"results": [
		{"id": 11, "name": "Sail", "starts_on": "2026-01-05", "ends_on": "2026-03-31", "lane": "Backend", "legend": "Committed", "tags": ["sea", "Q1"], "percent_done": 40, "is_container": false, "parked": false, "description": "long markdown"},
		{"id": 12, "name": "Dock", "starts_on": "2026-04-01", "ends_on": "2026-06-30", "lane": "Frontend", "legend": "Exploring", "tags": [], "percent_done": 0, "is_container": true, "parked": true}
	], "paging": {"record_count": 2, "page_count": 1, "current_page": 1, "page_size": 500}}`
	lanes := `{"results": [{"id": 7, "name": "Backend"}, {"id": 8, "name": "Frontend"}], "paging": {"page_count": 1}}`

	result := FormatBarsWithContext(json.RawMessage(bars), json.RawMessage(lanes))
	var parsed struct {
		Count int              `json:"count"`
		Bars  []map[string]any `json:"bars"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil || parsed.Count != 2 {
		t.Fatalf("projection failed on the live envelope: %s", result)
	}
	b := parsed.Bars[0]
	want := map[string]any{"starts_on": "2026-01-05", "ends_on": "2026-03-31", "lane_name": "Backend", "lane_id": float64(7), "legend": "Committed", "percent_done": float64(40)}
	for k, v := range want {
		if b[k] != v {
			t.Errorf("%s = %v, want %v", k, b[k], v)
		}
	}
	if tags, _ := b["tags"].([]any); len(tags) != 2 {
		t.Errorf("tags = %v", b["tags"])
	}
	if _, ok := b["description"]; ok {
		t.Error("description stays behind get_bar")
	}
}

func TestFormatBars_ClientSideFilters(t *testing.T) {
	bars := `[
		{"id": 1, "name": "A", "lane": "Backend", "legend": "Committed", "tags": ["mobile"]},
		{"id": 2, "name": "B", "lane": "backend", "legend": "Exploring", "tags": ["web"]},
		{"id": 3, "name": "C", "lane": "Frontend", "legend": "Committed", "tags": ["Mobile", "web"]}
	]`
	lanes := `[{"id": 7, "name": "Backend"}, {"id": 8, "name": "Frontend"}]`

	cases := []struct {
		name string
		f    BarFilter
		want []float64
	}{
		{"lane by name, case-insensitive", BarFilter{Lane: "BACKEND"}, []float64{1, 2}},
		{"lane by id", BarFilter{Lane: "8"}, []float64{3}},
		{"legend", BarFilter{Legend: "committed"}, []float64{1, 3}},
		{"tag", BarFilter{Tag: "mobile"}, []float64{1, 3}},
		{"combined", BarFilter{Legend: "Committed", Tag: "web"}, []float64{3}},
		{"no match", BarFilter{Tag: "none"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := barListing{lanes: json.RawMessage(lanes), filter: tc.f}.format(json.RawMessage(bars))
			var parsed struct {
				Count    int              `json:"count"`
				Scanned  int              `json:"scanned"`
				Filtered bool             `json:"filtered"`
				Bars     []map[string]any `json:"bars"`
			}
			if err := json.Unmarshal(out, &parsed); err != nil {
				t.Fatal(err)
			}
			if parsed.Count != len(tc.want) || parsed.Scanned != 3 || !parsed.Filtered {
				t.Fatalf("count=%d scanned=%d filtered=%v, want %d/3/true", parsed.Count, parsed.Scanned, parsed.Filtered, len(tc.want))
			}
			for i, id := range tc.want {
				if parsed.Bars[i]["id"] != id {
					t.Errorf("bar %d id = %v, want %v", i, parsed.Bars[i]["id"], id)
				}
			}
		})
	}
}

// Client-side filtering must happen before the 50-item cap, otherwise a
// match beyond position 50 would be silently lost.
func TestFormatBars_FilterBeforeCap(t *testing.T) {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < 120; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		tag := "other"
		if i >= 100 {
			tag = "late"
		}
		fmt.Fprintf(&sb, `{"id": %d, "name": "b%d", "lane": "L", "tags": [%q]}`, i, i, tag)
	}
	sb.WriteByte(']')
	out := barListing{filter: BarFilter{Tag: "late"}}.format(json.RawMessage(sb.String()))
	var parsed struct {
		Count int  `json:"count"`
		Trunc bool `json:"truncated"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Count != 20 || parsed.Trunc {
		t.Errorf("count=%d truncated=%v, want 20/false", parsed.Count, parsed.Trunc)
	}
}

// Two lanes with one name: lane_id must not be guessed.
func TestFormatBars_AmbiguousLaneNameNoID(t *testing.T) {
	out := FormatBarsWithContext(json.RawMessage(`[{"id":1,"lane":"Dup"}]`), json.RawMessage(`[{"id":1,"name":"Dup"},{"id":2,"name":"dup"}]`))
	if strings.Contains(string(out), `"lane_id":1`) || strings.Contains(string(out), `"lane_id":2`) {
		t.Errorf("ambiguous lane name must not resolve to an id: %s", out)
	}
}
