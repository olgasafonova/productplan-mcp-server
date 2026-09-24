package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

type bulkData struct {
	DryRun    bool             `json:"dry_run"`
	Succeeded int              `json:"succeeded"`
	Failed    int              `json:"failed"`
	Results   []bulkItemResult `json:"results"`
}

func callBulk(t *testing.T, h mcp.Handler, args map[string]any) (string, bulkData, error) {
	t.Helper()
	out, err := h.Handle(context.Background(), args)
	if err != nil {
		return "", bulkData{}, err
	}
	var fr FormattedResponse
	if err := json.Unmarshal(out, &fr); err != nil {
		t.Fatalf("bad response: %v", err)
	}
	var d bulkData
	if err := json.Unmarshal(fr.Data, &d); err != nil {
		t.Fatalf("bad data: %v", err)
	}
	return fr.Summary, d, nil
}

func seededFake(ids ...string) *fakeAPI {
	f := newFakeAPI()
	for _, id := range ids {
		f.addBar(id, map[string]any{"parked": false})
	}
	return f
}

func updateItems(ids ...string) []any {
	items := make([]any, len(ids))
	for i, id := range ids {
		items[i] = map[string]any{"bar_id": id}
	}
	return items
}

func TestBulkUpdateDryRunSendsNothing(t *testing.T) {
	f := seededFake("1", "2", "3")
	client := f.start(t)
	summary, d, err := callBulk(t, bulkUpdateBarsHandler(client), map[string]any{
		"items": updateItems("1", "2", "3"), "set": map[string]any{"legend": "committed"}, "dry_run": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.writes()) != 0 {
		t.Fatalf("dry run wrote: %+v", f.writes())
	}
	if !strings.Contains(summary, "Dry run: 3 of 3 bars valid; nothing was sent") {
		t.Errorf("unexpected summary: %s", summary)
	}
	for _, r := range d.Results {
		if r.Payload["legend"] != "Committed" || len(r.Payload) != 1 {
			t.Errorf("item %d payload should be exactly {legend:Committed}, got %v", r.Index, r.Payload)
		}
	}
	if n := f.countGets("/roadmaps/100"); n != 1 {
		t.Errorf("expected one roadmap fetch for one distinct roadmap, got %d", n)
	}
}

func TestBulkUpdateRoadmapIDSkipsBarLookups(t *testing.T) {
	f := seededFake("1", "2")
	client := f.start(t)
	if _, _, err := callBulk(t, bulkUpdateBarsHandler(client), map[string]any{
		"items": updateItems("1", "2"), "set": map[string]any{"lane": "Mobile"}, "roadmap_id": "100",
	}); err != nil {
		t.Fatal(err)
	}
	if f.countGets("/bars/1")+f.countGets("/bars/2") != 0 {
		t.Error("roadmap_id was given, so bars should not be looked up")
	}
	if len(f.writes()) != 2 {
		t.Errorf("expected 2 PATCHes, got %d", len(f.writes()))
	}
}

func TestBulkUpdateValidationFailureSendsNothing(t *testing.T) {
	f := seededFake("1", "2", "3")
	client := f.start(t)
	items := updateItems("1", "2", "3")
	items[1] = map[string]any{"bar_id": "2", "legend": "Purple"}
	items[2] = map[string]any{"bar_id": "3", "legend_id": "4"}
	_, _, err := callBulk(t, bulkUpdateBarsHandler(client), map[string]any{
		"items": items, "set": map[string]any{"legend": "Committed"},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	for _, want := range []string{"validation failed for 2 of 3 items; nothing was written", "items[1] (bar 2)", `legend "Purple"`, "items[2] (bar 3)", "CLEAR the bar's color"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing %q: %v", want, err)
		}
	}
	if len(f.writes()) != 0 {
		t.Fatalf("validation failure wrote: %+v", f.writes())
	}
}

func TestBulkUpdatePartialFailure(t *testing.T) {
	f := seededFake("1", "2", "3")
	f.failPatch["2"] = 422
	client := f.start(t)
	summary, d, err := callBulk(t, bulkUpdateBarsHandler(client), map[string]any{
		"items": updateItems("1", "2", "3"), "set": map[string]any{"percent_done": 50},
	})
	if err != nil {
		t.Fatalf("partial failure should be a normal result, got error %v", err)
	}
	if !strings.HasPrefix(summary, "Updated 2 of 3 bars; 1 failed") {
		t.Errorf("unexpected summary: %s", summary)
	}
	if d.Results[1].OK || d.Results[1].BarID != "2" || !strings.Contains(d.Results[1].Error, "422") {
		t.Errorf("item 1 should report the 422: %+v", d.Results[1])
	}
	if !d.Results[0].OK || !d.Results[2].OK {
		t.Errorf("items 0 and 2 should succeed: %+v", d.Results)
	}
}

func TestBulkUpdateAllFailedIsError(t *testing.T) {
	f := seededFake("1", "2")
	f.failPatch["1"], f.failPatch["2"] = 500, 500
	client := f.start(t)
	_, _, err := callBulk(t, bulkUpdateBarsHandler(client), map[string]any{
		"items": updateItems("1", "2"), "set": map[string]any{"notes": "x"},
	})
	if err == nil || !strings.Contains(err.Error(), "Updated 0 of 2 bars; 2 failed") || !strings.Contains(err.Error(), "Nothing was changed") {
		t.Errorf("expected an error result when nothing succeeded, got %v", err)
	}
}

func TestBulkUpdateConcurrencyBound(t *testing.T) {
	ids := make([]string, 20)
	for i := range ids {
		ids[i] = fmt.Sprint(i + 1)
	}
	f := seededFake(ids...)
	f.delay = 20 * time.Millisecond
	client := f.start(t)
	if _, _, err := callBulk(t, bulkUpdateBarsHandler(client), map[string]any{
		"items": updateItems(ids...), "set": map[string]any{"notes": "x"},
	}); err != nil {
		t.Fatal(err)
	}
	peak := f.maxInFlight.Load()
	if peak > bulkConcurrency {
		t.Errorf("peak in-flight writes %d exceeds bound %d", peak, bulkConcurrency)
	}
	if peak < 2 {
		t.Errorf("peak in-flight writes %d: writes did not run concurrently", peak)
	}
}

func TestBulkArgsValidation(t *testing.T) {
	tooMany := make([]any, maxBulkItems+1)
	for i := range tooMany {
		tooMany[i] = map[string]any{"bar_id": fmt.Sprint(i + 1)}
	}
	f := newFakeAPI()
	client := f.start(t)
	tests := []struct {
		name    string
		h       mcp.Handler
		args    map[string]any
		wantErr string
	}{
		{"too many", bulkUpdateBarsHandler(client), map[string]any{"items": tooMany}, "limit is 100"},
		{"empty", bulkUpdateBarsHandler(client), map[string]any{"items": []any{}}, "at least one"},
		{"duplicate", bulkUpdateBarsHandler(client), map[string]any{"items": updateItems("1", "1")}, "repeats bar_id 1"},
		{"bad id", bulkUpdateBarsHandler(client), map[string]any{"items": updateItems("../x")}, "invalid"},
		{"create without lane", bulkCreateBarsHandler(client), map[string]any{"roadmap_id": "100", "items": []any{map[string]any{"name": "A"}}}, "items[0]: required parameter missing: lane"},
		{"delete without confirm", bulkDeleteBarsHandler(client), map[string]any{"bar_ids": []any{"1"}}, "confirm:true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.h.Handle(context.Background(), tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("want error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
	if len(f.requests) != 0 {
		t.Errorf("argument errors must not reach the API: %+v", f.requests)
	}
}

func TestBulkCreate(t *testing.T) {
	f := newFakeAPI()
	client := f.start(t)
	summary, d, err := callBulk(t, bulkCreateBarsHandler(client), map[string]any{
		"roadmap_id": "100",
		"set":        map[string]any{"lane": "Backend", "legend": "Exploring"},
		"items": []any{
			map[string]any{"name": "A", "starts_on": "2026-01-01", "ends_on": "2026-01-31"},
			map[string]any{"name": "B", "lane": "Mobile"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary != "Created 2 of 2 bars" {
		t.Errorf("unexpected summary: %s", summary)
	}
	for _, r := range d.Results {
		if !r.OK || r.BarID == "" {
			t.Errorf("expected created ID for each item: %+v", r)
		}
	}
	w := f.writes()
	if len(w) != 2 {
		t.Fatalf("expected 2 POSTs, got %d", len(w))
	}
	byName := map[string]map[string]any{}
	for _, r := range w {
		byName[r.Body["name"].(string)] = r.Body
	}
	if byName["A"]["lane"] != "Backend" || byName["A"]["parked"] != false || byName["A"]["legend"] != "Exploring" {
		t.Errorf("A should take set lane, set legend, and parked:false: %v", byName["A"])
	}
	if _, ok := byName["B"]["parked"]; ok || byName["B"]["lane"] != "Mobile" {
		t.Errorf("B should override lane and stay parked by default: %v", byName["B"])
	}
	if n := f.countGets("/roadmaps/100"); n != 1 {
		t.Errorf("expected one roadmap fetch, got %d", n)
	}
}

func TestBulkDeleteVerifiesByReadBack(t *testing.T) {
	f := seededFake("1", "2", "3")
	f.keepOnDel["2"] = true
	client := f.start(t)
	summary, d, err := callBulk(t, bulkDeleteBarsHandler(client), map[string]any{
		"bar_ids": []any{"1", "2", "3", "404"}, "confirm": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(summary, "Deleted 2 of 4 bars; 2 failed") {
		t.Errorf("unexpected summary: %s", summary)
	}
	if d.Results[1].OK || !strings.Contains(d.Results[1].Error, "still exists on read-back") {
		t.Errorf("bar 2 survived its delete and must be reported: %+v", d.Results[1])
	}
	if d.Results[3].OK || !strings.Contains(d.Results[3].Error, "not found") {
		t.Errorf("bar 404 should be reported not found: %+v", d.Results[3])
	}
	if f.countGets("/bars/1") != 1 {
		t.Error("expected a read-back GET after deleting bar 1")
	}
}

func TestBulkDeleteDryRunSendsNothing(t *testing.T) {
	f := seededFake("1", "2")
	client := f.start(t)
	summary, d, err := callBulk(t, bulkDeleteBarsHandler(client), map[string]any{
		"bar_ids": []any{"1", "2"}, "dry_run": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.writes()) != 0 {
		t.Fatalf("dry run deleted: %+v", f.writes())
	}
	if !strings.HasPrefix(summary, "Dry run: 2 of 2 bars valid") || d.Results[0].Name != "Bar 1" {
		t.Errorf("dry run should list bar names: %s %+v", summary, d.Results)
	}
}

func TestBulkToolAnnotations(t *testing.T) {
	want := map[string]bool{"bulk_update_bars": true, "bulk_create_bars": false, "bulk_delete_bars": true}
	found := 0
	for _, tool := range BuildAllTools() {
		destructive, ok := want[tool.Name]
		if !ok {
			continue
		}
		found++
		a := tool.Annotations
		if a == nil || a.DestructiveHint == nil || *a.DestructiveHint != destructive {
			t.Errorf("%s: DestructiveHint should be %t, got %+v", tool.Name, destructive, a)
			continue
		}
		if a.OpenWorldHint == nil || !*a.OpenWorldHint {
			t.Errorf("%s: OpenWorldHint should be true", tool.Name)
		}
		if a.ReadOnlyHint || a.IdempotentHint {
			t.Errorf("%s: must not claim ReadOnlyHint or IdempotentHint (Article VIII)", tool.Name)
		}
		if tool.OutputSchema != nil {
			t.Errorf("%s: write tools declare no output schema (Article X)", tool.Name)
		}
	}
	if found != len(want) {
		t.Errorf("expected %d bulk tools, found %d", len(want), found)
	}
}

func TestBulkToolsRegistered(t *testing.T) {
	for _, name := range []string{"bulk_update_bars", "bulk_create_bars", "bulk_delete_bars"} {
		if _, ok := handlerConstructors[name]; !ok {
			t.Errorf("%s missing from handlerConstructors", name)
		}
	}
}
