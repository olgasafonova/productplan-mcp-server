package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// listJSON builds a JSON array of n objects using the given per-item template
// (which receives the index).
func listJSON(n int, tmpl func(i int) string) string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(tmpl(i))
	}
	sb.WriteByte(']')
	return sb.String()
}

// TestFormatList_CapsAtDefault covers the HG-2 default list cap on the
// projecting formatList path (exercised via FormatLanes).
func TestFormatList_CapsAtDefault(t *testing.T) {
	input := listJSON(60, func(i int) string {
		return fmt.Sprintf(`{"id": %d, "name": "Lane %d", "color": "#FFF"}`, i, i)
	})

	result := FormatLanes(json.RawMessage(input))

	var parsed struct {
		Count     int              `json:"count"`
		Total     int              `json:"total"`
		Truncated bool             `json:"truncated"`
		Lanes     []map[string]any `json:"lanes"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if parsed.Count != defaultListCap {
		t.Errorf("count = %d, want %d", parsed.Count, defaultListCap)
	}
	if parsed.Total != 60 {
		t.Errorf("total = %d, want 60", parsed.Total)
	}
	if !parsed.Truncated {
		t.Error("expected truncated=true")
	}
	if len(parsed.Lanes) != defaultListCap {
		t.Errorf("returned %d lanes, want %d", len(parsed.Lanes), defaultListCap)
	}
}

// TestFormatList_UnderCapNotTruncated confirms a short list is not flagged.
func TestFormatList_UnderCapNotTruncated(t *testing.T) {
	result := FormatLanes(json.RawMessage(`[{"id":1,"name":"A","color":"#FFF"}]`))

	var parsed struct {
		Count     int  `json:"count"`
		Total     int  `json:"total"`
		Truncated bool `json:"truncated"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if parsed.Count != 1 || parsed.Total != 1 {
		t.Errorf("count/total = %d/%d, want 1/1", parsed.Count, parsed.Total)
	}
	if parsed.Truncated {
		t.Error("under-cap result should not be truncated")
	}
}

// TestFormatBarsWithContext_CapsAtDefault covers the HG-2 cap on the bar
// formatter, which has its own loop rather than going through formatList.
func TestFormatBarsWithContext_CapsAtDefault(t *testing.T) {
	bars := listJSON(60, func(i int) string {
		return fmt.Sprintf(`{"id": %d, "name": "Bar %d", "lane_id": 1}`, i, i)
	})
	lanes := `[{"id":1,"name":"Eng"}]`

	result := FormatBarsWithContext(json.RawMessage(bars), json.RawMessage(lanes))

	var parsed struct {
		Count     int  `json:"count"`
		Total     int  `json:"total"`
		Truncated bool `json:"truncated"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if parsed.Count != defaultListCap || parsed.Total != 60 || !parsed.Truncated {
		t.Errorf("got count=%d total=%d truncated=%v, want %d/60/true", parsed.Count, parsed.Total, parsed.Truncated, defaultListCap)
	}
}
