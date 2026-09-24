package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

// pagedServer serves `pages` pages of `perPage` records each at any path,
// honouring ?page=N, and counts requests. failPage (if > 0) returns 500.
func pagedServer(t *testing.T, pages, perPage, failPage int, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		page := 1
		if p := r.URL.Query().Get("page"); p != "" {
			page, _ = strconv.Atoi(p)
		}
		if page == failPage {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"errors":["boom"]}`))
			return
		}
		results := make([]map[string]any, 0, perPage)
		for i := 0; i < perPage; i++ {
			results = append(results, map[string]any{"id": (page-1)*perPage + i, "name": fmt.Sprintf("item %d", (page-1)*perPage+i)})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": results,
			"paging":  map[string]any{"record_count": pages * perPage, "page_count": pages, "current_page": page, "page_size": perPage},
		})
	}))
}

func TestGetList_FollowsPageCount(t *testing.T) {
	var hits atomic.Int32
	srv := pagedServer(t, 3, 4, 0, &hits)
	defer srv.Close()

	data, err := testClient(t, srv).GetList(context.Background(), "/roadmaps", Query{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	env, ok := parseEnvelope(data)
	if !ok {
		t.Fatalf("merged body is not an envelope: %s", data)
	}
	if len(env.Results) != 12 {
		t.Fatalf("merged %d results, want 12", len(env.Results))
	}
	// Order must be page order, not completion order.
	var last map[string]any
	if err := json.Unmarshal(env.Results[11], &last); err != nil || last["id"] != float64(11) {
		t.Errorf("last result = %v, want id 11", last)
	}
	if env.Paging.incomplete() {
		t.Error("a fully fetched merge must not be marked incomplete")
	}
	if hits.Load() != 3 {
		t.Errorf("server hit %d times, want 3", hits.Load())
	}
}

func TestGetList_SinglePagePassesThrough(t *testing.T) {
	var hits atomic.Int32
	srv := pagedServer(t, 1, 2, 0, &hits)
	defer srv.Close()

	data, err := testClient(t, srv).GetList(context.Background(), "/roadmaps", Query{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"current_page":1`) {
		t.Errorf("single-page body should be returned untouched, got %s", data)
	}
	if hits.Load() != 1 {
		t.Errorf("hits = %d, want 1", hits.Load())
	}
}

func TestGetList_BareArrayPassesThrough(t *testing.T) {
	srv := testServer(t, map[string]string{"/users": `[{"id":1}]`})
	defer srv.Close()
	data, err := testClient(t, srv).GetList(context.Background(), "/users", Query{})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `[{"id":1}]` {
		t.Errorf("got %s", data)
	}
}

func TestGetList_PageFailureFailsWholeCall(t *testing.T) {
	var hits atomic.Int32
	srv := pagedServer(t, 4, 2, 3, &hits)
	defer srv.Close()

	_, err := testClient(t, srv).GetList(context.Background(), "/roadmaps", Query{})
	if err == nil {
		t.Fatal("expected an error when a later page fails; a partial merge must not look complete")
	}
	if !strings.Contains(err.Error(), "page 3 of 4") {
		t.Errorf("error should name the failing page, got %v", err)
	}
}

func TestGetList_CapsAtMaxPagesAndSaysSo(t *testing.T) {
	var hits atomic.Int32
	srv := pagedServer(t, maxListPages+5, 1, 0, &hits)
	defer srv.Close()

	data, err := testClient(t, srv).GetList(context.Background(), "/roadmaps", Query{})
	if err != nil {
		t.Fatal(err)
	}
	env, _ := parseEnvelope(data)
	if len(env.Results) != maxListPages {
		t.Errorf("merged %d results, want %d", len(env.Results), maxListPages)
	}
	if !env.Paging.incomplete() || env.Paging.PagesFetched != maxListPages || env.Paging.PageCount != maxListPages+5 {
		t.Errorf("paging = %+v, want pages_fetched=%d < page_count=%d", env.Paging, maxListPages, maxListPages+5)
	}
	if int(hits.Load()) != maxListPages {
		t.Errorf("hits = %d, want %d", hits.Load(), maxListPages)
	}

	// The formatter surfaces the cap instead of presenting a partial total.
	out := FormatRoadmapList(data)
	var parsed map[string]any
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["incomplete"] != true || parsed["record_count"] != float64(maxListPages+5) {
		t.Errorf("formatted payload missing incomplete marker: %s", out)
	}
	if note, _ := parsed["note"].(string); !strings.Contains(note, "first 50 of 55 pages") {
		t.Errorf("note = %q", note)
	}
}

func TestListURL_EncodesQueryDeterministically(t *testing.T) {
	q := Query{
		Predicates: map[string]string{"starts_on_gteq": "2026-01-01", "name_i_cont": "Ship & Sail"},
		Sort:       "name asc",
	}
	got := listURL("/roadmaps/1/bars", q, 2)
	want := "/roadmaps/1/bars?page=2&page_size=500&q%5Bname_i_cont%5D=Ship+%26+Sail&q%5Bs%5D=name+asc&q%5Bstarts_on_gteq%5D=2026-01-01"
	if got != want {
		t.Errorf("listURL =\n %s\nwant\n %s", got, want)
	}
	if got != listURL("/roadmaps/1/bars", q, 2) {
		t.Error("listURL must be deterministic for cache keys")
	}
	if strings.Contains(listURL("/x", Query{}, 1), "page=") {
		t.Error("page 1 should not send page=")
	}
}

func TestGetList_SendsQueryToServer(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"results":[],"paging":{"record_count":0,"page_count":1,"current_page":1,"page_size":500}}`))
	}))
	defer srv.Close()

	_, err := testClient(t, srv).ListRoadmapsWhere(context.Background(), Query{Predicates: map[string]string{"name_i_cont": "Q3"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "q%5Bname_i_cont%5D=Q3") || !strings.Contains(gotQuery, "page_size=500") {
		t.Errorf("raw query = %q", gotQuery)
	}
}
