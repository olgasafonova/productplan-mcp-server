package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

func TestBuildQuery_RoadmapBarsMapping(t *testing.T) {
	q, err := buildQuery("get_roadmap_bars", map[string]any{
		"roadmap_id":    "1",
		"name_contains": " Ship ",
		"starts_after":  "2026-01-01",
		"starts_before": "2026-03-31",
		"ends_after":    "2026-02-01",
		"ends_before":   "2026-12-31",
		"is_container":  false,
		"sort":          "starts_on DESC",
		"lane":          "ignored here: client-side",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"name_i_cont":        "Ship",
		"starts_on_gteq":     "2026-01-01",
		"starts_on_lteq":     "2026-03-31",
		"ends_on_gteq":       "2026-02-01",
		"ends_on_lteq":       "2026-12-31",
		"is_container_false": "1",
	}
	if len(q.Predicates) != len(want) {
		t.Errorf("predicates = %v", q.Predicates)
	}
	for k, v := range want {
		if q.Predicates[k] != v {
			t.Errorf("q[%s] = %q, want %q", k, q.Predicates[k], v)
		}
	}
	if q.Sort != "starts_on desc" {
		t.Errorf("sort = %q", q.Sort)
	}
}

func TestBuildQuery_BoolTrueAndStringBool(t *testing.T) {
	for _, v := range []any{true, "true", "TRUE"} {
		q, err := buildQuery("get_roadmap_bars", map[string]any{"is_container": v})
		if err != nil || q.Predicates["is_container_true"] != "1" {
			t.Errorf("is_container=%v -> %v, %v", v, q.Predicates, err)
		}
	}
	if _, err := buildQuery("get_roadmap_bars", map[string]any{"is_container": "yes"}); err == nil {
		t.Error("is_container=yes should be rejected")
	}
}

func TestBuildQuery_Rejections(t *testing.T) {
	cases := []struct {
		tool string
		args map[string]any
		want string
	}{
		{"get_roadmap_bars", map[string]any{"starts_after": "01-02-2026"}, "starts_after must be a calendar date in YYYY-MM-DD format"},
		{"get_roadmap_bars", map[string]any{"ends_before": "2026-02-30"}, "ends_before must be a calendar date"},
		{"get_roadmap_bars", map[string]any{"sort": "lane asc"}, `sort field "lane" is not sortable for get_roadmap_bars; allowed: created_at, ends_on, id, is_container, name, starts_on, updated_at`},
		{"get_roadmap_bars", map[string]any{"sort": "name sideways"}, "sort direction must be asc or desc"},
		{"get_roadmap_bars", map[string]any{"sort": "name asc extra"}, "sort must be"},
		{"get_roadmap_bars", map[string]any{"name_contains": 42.0}, "name_contains must be a string"},
		{"list_launches", map[string]any{"launch_after": "next week"}, "launch_after must be a calendar date"},
		{"list_roadmaps", map[string]any{"sort": "owner"}, "allowed: created_at, id, is_version, name, updated_at"},
	}
	for _, tc := range cases {
		_, err := buildQuery(tc.tool, tc.args)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s %v: err = %v, want containing %q", tc.tool, tc.args, err, tc.want)
		}
	}
}

func TestBuildQuery_EmptyAndUnknownToolDoNotFilter(t *testing.T) {
	q, err := buildQuery("get_roadmap_bars", map[string]any{"name_contains": "", "sort": ""})
	if err != nil || !q.IsZero() {
		t.Errorf("empty args should not filter: %+v, %v", q, err)
	}
	q, err = buildQuery("get_bar", map[string]any{"sort": "anything"})
	if err != nil || !q.IsZero() {
		t.Errorf("tools without a spec take no filters: %+v, %v", q, err)
	}
}

// The schema is generated from the same table buildQuery reads, so every
// advertised filter argument is one buildQuery accepts, and the sort
// Pattern accepts exactly the allowlist.
func TestFilterPropertiesMatchTable(t *testing.T) {
	tools := BuildAllTools()
	byName := map[string]map[string]bool{}
	for _, tl := range tools {
		props := map[string]bool{}
		for k := range tl.InputSchema.Properties {
			props[k] = true
		}
		byName[tl.Name] = props
	}
	for tool, spec := range listFilters {
		props, ok := byName[tool]
		if !ok {
			t.Errorf("filter spec for unknown tool %s", tool)
			continue
		}
		for _, fa := range spec.args {
			if !props[fa.arg] {
				t.Errorf("%s: filter arg %s missing from schema", tool, fa.arg)
			}
		}
		if !props["sort"] {
			t.Errorf("%s: sort missing from schema", tool)
		}
		re := regexp.MustCompile(sortPattern(spec.sortFields))
		for _, f := range spec.sortFields {
			if !re.MatchString(f) || !re.MatchString(f+" desc") {
				t.Errorf("%s: sort pattern rejects allowed field %s", tool, f)
			}
		}
		if re.MatchString("bogus asc") {
			t.Errorf("%s: sort pattern accepts a field outside the allowlist", tool)
		}
	}
}

// End to end through the handler: server-side filters reach the wire as
// q[...] params, client-side ones narrow the result, and the summary says
// "matched the filters" when nothing survives.
func TestGetRoadmapBarsHandler_Filters(t *testing.T) {
	var mu sync.Mutex
	var barsQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/lanes") {
			_, _ = w.Write([]byte(`{"results":[{"id":7,"name":"Backend"},{"id":8,"name":"Mobile"}],"paging":{"page_count":1}}`))
			return
		}
		mu.Lock()
		barsQuery = r.URL.Query()
		mu.Unlock()
		_, _ = w.Write([]byte(`{"results":[
			{"id":1,"name":"Ship A","lane":"Mobile","legend":"Committed","tags":["urgent"],"starts_on":"2026-04-02"},
			{"id":2,"name":"Ship B","lane":"Backend","legend":"Committed","tags":["urgent"],"starts_on":"2026-05-01"},
			{"id":3,"name":"Ship C","lane":"Mobile","legend":"Exploring","tags":[],"starts_on":"2026-06-01"}
		],"paging":{"page_count":1}}`))
	}))
	defer srv.Close()
	client, err := api.New(api.Config{Token: "t", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	h := getRoadmapBarsHandler(client)

	out, err := h.Handle(context.Background(), map[string]any{
		"roadmap_id": "5", "name_contains": "Ship", "starts_after": "2026-04-01", "sort": "starts_on",
		"lane": "mobile", "tag": "URGENT",
	})
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	gotQ := barsQuery
	mu.Unlock()
	if gotQ.Get("q[name_i_cont]") != "Ship" || gotQ.Get("q[starts_on_gteq]") != "2026-04-01" || gotQ.Get("q[s]") != "starts_on asc" {
		t.Errorf("server-side filters not on the wire: %v", gotQ)
	}
	if gotQ.Has("q[lane]") || gotQ.Has("lane") {
		t.Errorf("client-side filters must not be sent upstream: %v", gotQ)
	}
	var resp FormattedResponse
	if err = json.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Summary != "Found 1 bar" {
		t.Errorf("summary = %q", resp.Summary)
	}
	var data struct {
		Bars []map[string]any `json:"bars"`
	}
	if err = json.Unmarshal(resp.Data, &data); err != nil || len(data.Bars) != 1 || data.Bars[0]["id"] != float64(1) || data.Bars[0]["lane_id"] != float64(8) {
		t.Errorf("data = %s", resp.Data)
	}

	out, err = h.Handle(context.Background(), map[string]any{"roadmap_id": "5", "legend": "Parked"})
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal(out, &resp)
	if resp.Summary != "No bars matched the filters" {
		t.Errorf("summary = %q", resp.Summary)
	}

	if _, err := h.Handle(context.Background(), map[string]any{"roadmap_id": "5", "sort": "lane"}); err == nil {
		t.Error("bad sort must fail before any request")
	}
}

func TestListHandlersSendFilters(t *testing.T) {
	var mu sync.Mutex
	var got url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got = r.URL.Query()
		mu.Unlock()
		_, _ = w.Write([]byte(`{"results":[],"paging":{"page_count":1}}`))
	}))
	defer srv.Close()
	client, err := api.New(api.Config{Token: "t", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		h       Handler
		args    map[string]any
		wantKey string
		wantVal string
	}{
		{"list_roadmaps", listRoadmapsHandler(client), map[string]any{"name_contains": "Q3"}, "q[name_i_cont]", "Q3"},
		{"list_ideas", listIdeasHandler(client), map[string]any{"channel": "email"}, "q[channel_eq]", "email"},
		{"list_opportunities", listOpportunitiesHandler(client), map[string]any{"workflow_status": "validated"}, "q[workflow_status_eq]", "validated"},
		{"list_launches", listLaunchesHandler(client), map[string]any{"launch_before": "2026-12-31", "sort": "launch_date desc"}, "q[launch_date_lteq]", "2026-12-31"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.h.Handle(context.Background(), tc.args)
			if err != nil {
				t.Fatal(err)
			}
			mu.Lock()
			q := got
			mu.Unlock()
			if q.Get(tc.wantKey) != tc.wantVal {
				t.Errorf("%s = %q, want %q (query %v)", tc.wantKey, q.Get(tc.wantKey), tc.wantVal, q)
			}
			var resp FormattedResponse
			if err := json.Unmarshal(out, &resp); err != nil || !strings.Contains(resp.Summary, "matched the filters") {
				t.Errorf("empty filtered result summary = %q", resp.Summary)
			}
		})
	}
}
