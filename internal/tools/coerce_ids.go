package tools

import (
	"math"
	"strconv"
	"strings"
)

// coerceIDs returns a copy of args in which every whole number under an
// ID key (`*_id`, or an element of `*_ids`) is replaced by its decimal
// string, at any depth (bulk items nest their own bar_id).
//
// Every ID argument is declared and decoded as a string, but ProductPlan
// returns IDs as JSON numbers, so an agent copying `"id": 36935840` from a
// read into a write sends a number. Before this, json.Unmarshal failed with
// "cannot unmarshal number into ... of type string" (found in a live
// round-trip on roadmap 641608, 24-09-2026). Non-integral numbers are left
// alone so they still fail loudly.
func coerceIDs(args map[string]any) map[string]any {
	out := make(map[string]any, len(args))
	for k, v := range args {
		out[k] = coerceValue(k, v)
	}
	return out
}

func coerceValue(key string, v any) any {
	switch x := v.(type) {
	case map[string]any:
		return coerceIDs(x)
	case []any:
		items := make([]any, len(x))
		for i, item := range x {
			items[i] = coerceValue(key, item)
		}
		return items
	case float64:
		if isIDKey(key) && x == math.Trunc(x) && !math.IsInf(x, 0) {
			return strconv.FormatFloat(x, 'f', 0, 64)
		}
	}
	return v
}

// isIDKey reports whether key names an ID or a list of IDs.
func isIDKey(key string) bool {
	return strings.HasSuffix(key, "_id") || strings.HasSuffix(key, "_ids")
}
