package tools

import (
	"reflect"
	"testing"
)

func TestCoerceIDs(t *testing.T) {
	in := map[string]any{
		"bar_id":           float64(36935840),
		"container_bar_id": float64(36935840),
		"bar_ids":          []any{float64(1), "2", float64(3)},
		"percent_done":     float64(30),
		"legend_id":        float64(1.5),
		"name":             "x",
		"items":            []any{map[string]any{"bar_id": float64(7), "starts_on": "2027-01-01"}},
		"set":              map[string]any{"container_bar_id": float64(9)},
	}
	want := map[string]any{
		"bar_id":           "36935840",
		"container_bar_id": "36935840",
		"bar_ids":          []any{"1", "2", "3"},
		"percent_done":     float64(30),
		"legend_id":        float64(1.5),
		"name":             "x",
		"items":            []any{map[string]any{"bar_id": "7", "starts_on": "2027-01-01"}},
		"set":              map[string]any{"container_bar_id": "9"},
	}
	if got := coerceIDs(in); !reflect.DeepEqual(got, want) {
		t.Fatalf("coerceIDs:\n got %#v\nwant %#v", got, want)
	}
	if _, ok := in["bar_id"].(float64); !ok {
		t.Error("coerceIDs mutated its input")
	}
}

// TestParseArgsAcceptsNumericIDs is the live-found regression: a numeric
// container_bar_id (as copied from a read) used to fail to decode.
func TestParseArgsAcceptsNumericIDs(t *testing.T) {
	a, err := ParseArgs[ManageBarArgs](map[string]any{
		"action": "create", "roadmap_id": float64(641608), "container_bar_id": float64(36935840),
	})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if a.RoadmapID != "641608" || a.ContainerBarID != "36935840" {
		t.Fatalf("got roadmap_id=%q container_bar_id=%q", a.RoadmapID, a.ContainerBarID)
	}
	b, err := ParseArgs[BulkDeleteBarsArgs](map[string]any{"bar_ids": []any{float64(5), float64(6)}, "confirm": true})
	if err != nil || !reflect.DeepEqual(b.BarIDs, []string{"5", "6"}) {
		t.Fatalf("bulk bar_ids: %v %v", b.BarIDs, err)
	}
}
