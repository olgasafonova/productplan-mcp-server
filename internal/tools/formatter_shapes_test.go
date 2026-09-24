package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func decodeFormatted(t *testing.T, raw json.RawMessage) FormattedResponse {
	t.Helper()
	var resp FormattedResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("not a FormattedResponse: %v (%s)", err, raw)
	}
	return resp
}

// The live API wraps collections in {"results", "paging"}; before this,
// FormatList passed such bodies through with no summary and no cap.
func TestFormatList_EnvelopeIsCappedAndSummarised(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`{"results":[`)
	for i := 0; i < 70; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, `{"id":%d}`, i)
	}
	sb.WriteString(`],"paging":{"record_count":70,"page_count":1,"current_page":1,"page_size":500}}`)

	resp := decodeFormatted(t, mustFormat(t, sb.String(), "comment"))
	if resp.Summary != "Showing first 50 of 70 comments (refine to narrow)" {
		t.Errorf("summary = %q", resp.Summary)
	}
	var items []any
	if err := json.Unmarshal(resp.Data, &items); err != nil || len(items) != defaultListCap {
		t.Errorf("data should be the capped results array, got %d items (%v)", len(items), err)
	}
}

func TestFormatList_EmptyEnvelopeSaysNoneFound(t *testing.T) {
	resp := decodeFormatted(t, mustFormat(t, `{"results":[],"paging":{"record_count":0,"page_count":1}}`, "link"))
	if resp.Summary != "No links found" {
		t.Errorf("summary = %q", resp.Summary)
	}
}

func TestFormatList_IncompleteEnvelopeSaysSo(t *testing.T) {
	resp := decodeFormatted(t, mustFormat(t, `{"results":[{"id":1}],"paging":{"record_count":30000,"page_count":60,"page_size":500,"pages_fetched":50}}`, "user"))
	if !strings.Contains(resp.Summary, "Only the first 50 of 60 pages were fetched (30000 records upstream)") {
		t.Errorf("summary = %q", resp.Summary)
	}
}

func TestFormatList_ProjectedPayload(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"found", `{"count":2,"total":2,"roadmaps":[{},{}]}`, "Found 2 roadmaps"},
		{"empty", `{"count":0,"total":0,"roadmaps":[]}`, "No roadmaps found"},
		{"empty filtered", `{"count":0,"total":0,"filtered":true,"roadmaps":[]}`, "No roadmaps matched the filters"},
		{"truncated", `{"count":50,"total":80,"truncated":true,"roadmaps":[]}`, "Showing first 50 of 80 roadmaps (refine to narrow)"},
		{"warning", `{"count":1,"total":1,"warnings":["lane lookup failed: boom"],"roadmaps":[{}]}`, "Found 1 roadmap. Warning: lane lookup failed: boom"},
		{"incomplete", `{"count":1,"total":1,"incomplete":true,"note":"Only the first 50 of 51 pages were fetched","roadmaps":[{}]}`, "Found 1 roadmap. Only the first 50 of 51 pages were fetched"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := mustFormat(t, tc.in, "roadmap")
			resp := decodeFormatted(t, raw)
			if resp.Summary != tc.want {
				t.Errorf("summary = %q, want %q", resp.Summary, tc.want)
			}
			if string(resp.Data) != tc.in {
				t.Errorf("projected data must pass through unchanged, got %s", resp.Data)
			}
		})
	}
}

func mustFormat(t *testing.T, in, itemType string) json.RawMessage {
	t.Helper()
	out, err := FormatList(json.RawMessage(in), itemType)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestFormatFilteredList_EmptySaysFiltersExcludedEverything(t *testing.T) {
	out, err := FormatFilteredList(json.RawMessage(`[]`), "launch", true)
	if err != nil {
		t.Fatal(err)
	}
	if resp := decodeFormatted(t, out); resp.Summary != "No launches matched the filters" {
		t.Errorf("summary = %q", resp.Summary)
	}
}

func TestPluralize_IrregularItemTypes(t *testing.T) {
	for word, want := range map[string]string{
		"opportunity": "opportunities",
		"launch":      "launches",
		"key result":  "key results",
		"day":         "days",
		"bar":         "bars",
	} {
		if got := pluralize(word, 2); got != want {
			t.Errorf("pluralize(%q) = %q, want %q", word, got, want)
		}
	}
}
