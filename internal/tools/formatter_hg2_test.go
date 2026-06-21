package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestFormatList_CapsLongList covers the HG-2 default cap on the count-only
// FormatList path (used by comments and several other list tools): a list
// longer than defaultListCap is clipped, with the true total reported.
func TestFormatList_CapsLongList(t *testing.T) {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < 60; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, `{"id": %d}`, i)
	}
	sb.WriteByte(']')

	result, err := FormatList(json.RawMessage(sb.String()), "comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp FormattedResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !strings.Contains(resp.Summary, "Showing first 50 of 60 comments") {
		t.Errorf("summary = %q, want a truncation notice", resp.Summary)
	}

	var items []any
	if err := json.Unmarshal(resp.Data, &items); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if len(items) != defaultListCap {
		t.Errorf("data has %d items, want %d", len(items), defaultListCap)
	}
}
