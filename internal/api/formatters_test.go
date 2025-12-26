package api

import (
	"encoding/json"
	"testing"
)

func TestFormatRoadmapList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{
			name: "valid roadmaps",
			input: `[
				{"id": 1, "name": "Q1 Roadmap", "updated_at": "2024-01-15", "extra": "ignored"},
				{"id": 2, "name": "Q2 Roadmap", "updated_at": "2024-04-01"}
			]`,
			wantKeys: []string{"count", "roadmaps", "hint"},
		},
		{
			name:     "empty array",
			input:    `[]`,
			wantKeys: []string{"count", "roadmaps", "hint"},
		},
		{
			name:     "invalid JSON returns original",
			input:    `not json`,
			wantKeys: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRoadmapList(json.RawMessage(tt.input))

			if tt.wantKeys == nil {
				// Should return original on error
				if string(result) != tt.input {
					t.Errorf("expected original input on error")
				}
				return
			}

			var parsed map[string]any
			if err := json.Unmarshal(result, &parsed); err != nil {
				t.Fatalf("failed to parse result: %v", err)
			}

			for _, key := range tt.wantKeys {
				if _, ok := parsed[key]; !ok {
					t.Errorf("missing key %q in result", key)
				}
			}

			// Verify hint is present
			if hint, ok := parsed["hint"].(string); ok {
				if hint == "" {
					t.Error("expected non-empty hint")
				}
			}
		})
	}
}

func TestFormatRoadmapListContent(t *testing.T) {
	input := `[
		{"id": 123, "name": "Product Roadmap", "updated_at": "2024-12-26", "description": "ignored"}
	]`

	result := FormatRoadmapList(json.RawMessage(input))

	var parsed struct {
		Count    int              `json:"count"`
		Roadmaps []map[string]any `json:"roadmaps"`
		Hint     string           `json:"hint"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed.Count != 1 {
		t.Errorf("expected count 1, got %d", parsed.Count)
	}

	if len(parsed.Roadmaps) != 1 {
		t.Fatalf("expected 1 roadmap, got %d", len(parsed.Roadmaps))
	}

	rm := parsed.Roadmaps[0]
	if rm["id"] != float64(123) {
		t.Errorf("expected id 123, got %v", rm["id"])
	}
	if rm["name"] != "Product Roadmap" {
		t.Errorf("expected name 'Product Roadmap', got %v", rm["name"])
	}
	if _, ok := rm["description"]; ok {
		t.Error("description should be filtered out")
	}
}

func TestFormatBarsWithContext(t *testing.T) {
	bars := `[
		{"id": 1, "name": "Feature A", "start_date": "2024-01-01", "end_date": "2024-03-31", "lane_id": 100},
		{"id": 2, "name": "Feature B", "start_date": "2024-02-01", "end_date": "2024-04-30", "lane_id": 200}
	]`
	lanes := `[
		{"id": 100, "name": "Engineering"},
		{"id": 200, "name": "Design"}
	]`

	result := FormatBarsWithContext(json.RawMessage(bars), json.RawMessage(lanes))

	var parsed struct {
		Count int              `json:"count"`
		Bars  []map[string]any `json:"bars"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed.Count != 2 {
		t.Errorf("expected count 2, got %d", parsed.Count)
	}

	// Check lane names are enriched
	if parsed.Bars[0]["lane_name"] != "Engineering" {
		t.Errorf("expected lane_name 'Engineering', got %v", parsed.Bars[0]["lane_name"])
	}
	if parsed.Bars[1]["lane_name"] != "Design" {
		t.Errorf("expected lane_name 'Design', got %v", parsed.Bars[1]["lane_name"])
	}
}

func TestFormatBarsWithContextUnknownLane(t *testing.T) {
	bars := `[{"id": 1, "name": "Feature", "lane_id": 999}]`
	lanes := `[{"id": 100, "name": "Known Lane"}]`

	result := FormatBarsWithContext(json.RawMessage(bars), json.RawMessage(lanes))

	var parsed struct {
		Bars []map[string]any `json:"bars"`
	}
	json.Unmarshal(result, &parsed)

	if parsed.Bars[0]["lane_name"] != "Unknown" {
		t.Errorf("expected 'Unknown' for missing lane, got %v", parsed.Bars[0]["lane_name"])
	}
}

func TestFormatBarsWithContextInvalidJSON(t *testing.T) {
	// Invalid bars JSON should return original
	bars := `not valid json`
	lanes := `[{"id": 1}]`

	result := FormatBarsWithContext(json.RawMessage(bars), json.RawMessage(lanes))
	if string(result) != bars {
		t.Error("expected original bars on parse error")
	}

	// Invalid lanes JSON should return original bars
	bars = `[{"id": 1}]`
	lanes = `not valid json`

	result = FormatBarsWithContext(json.RawMessage(bars), json.RawMessage(lanes))
	if string(result) != bars {
		t.Error("expected original bars on lanes parse error")
	}
}

func TestFormatLanes(t *testing.T) {
	input := `[
		{"id": 1, "name": "Engineering", "color": "#FF0000", "order": 1},
		{"id": 2, "name": "Design", "color": "#00FF00", "order": 2}
	]`

	result := FormatLanes(json.RawMessage(input))

	var parsed struct {
		Count int              `json:"count"`
		Lanes []map[string]any `json:"lanes"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed.Count != 2 {
		t.Errorf("expected count 2, got %d", parsed.Count)
	}

	// Verify only expected fields are included
	lane := parsed.Lanes[0]
	if _, ok := lane["order"]; ok {
		t.Error("order should be filtered out")
	}
	if lane["color"] != "#FF0000" {
		t.Errorf("expected color '#FF0000', got %v", lane["color"])
	}
}

func TestFormatLanesInvalidJSON(t *testing.T) {
	input := `invalid`
	result := FormatLanes(json.RawMessage(input))
	if string(result) != input {
		t.Error("expected original input on parse error")
	}
}

func TestFormatMilestones(t *testing.T) {
	input := `[
		{"id": 1, "name": "Launch", "date": "2024-06-01", "description": "Product launch"},
		{"id": 2, "name": "Beta", "date": "2024-04-15"}
	]`

	result := FormatMilestones(json.RawMessage(input))

	var parsed struct {
		Count      int              `json:"count"`
		Milestones []map[string]any `json:"milestones"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed.Count != 2 {
		t.Errorf("expected count 2, got %d", parsed.Count)
	}

	// Verify description is filtered out
	if _, ok := parsed.Milestones[0]["description"]; ok {
		t.Error("description should be filtered out")
	}
}

func TestFormatMilestonesInvalidJSON(t *testing.T) {
	input := `invalid`
	result := FormatMilestones(json.RawMessage(input))
	if string(result) != input {
		t.Error("expected original input on parse error")
	}
}

func TestFormatObjectives(t *testing.T) {
	input := `[
		{"id": 1, "name": "Increase Revenue", "status": "on_track", "time_frame": "Q1 2024", "description": "ignored"},
		{"id": 2, "name": "Improve NPS", "status": "at_risk", "time_frame": "Q2 2024"}
	]`

	result := FormatObjectives(json.RawMessage(input))

	var parsed struct {
		Count      int              `json:"count"`
		Objectives []map[string]any `json:"objectives"`
		Hint       string           `json:"hint"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed.Count != 2 {
		t.Errorf("expected count 2, got %d", parsed.Count)
	}

	if parsed.Hint == "" {
		t.Error("expected non-empty hint")
	}

	// Verify fields
	obj := parsed.Objectives[0]
	if obj["status"] != "on_track" {
		t.Errorf("expected status 'on_track', got %v", obj["status"])
	}
	if obj["time_frame"] != "Q1 2024" {
		t.Errorf("expected time_frame 'Q1 2024', got %v", obj["time_frame"])
	}
}

func TestFormatObjectivesInvalidJSON(t *testing.T) {
	input := `invalid`
	result := FormatObjectives(json.RawMessage(input))
	if string(result) != input {
		t.Error("expected original input on parse error")
	}
}

func TestFormatIdeas(t *testing.T) {
	// Test with wrapper format
	input := `{
		"results": [
			{"id": 1, "name": "New Feature", "channel": "customer", "opportunities_count": 5},
			{"id": 2, "name": "Enhancement", "channel": "internal", "opportunities_count": 2}
		]
	}`

	result := FormatIdeas(json.RawMessage(input))

	var parsed struct {
		Count int              `json:"count"`
		Ideas []map[string]any `json:"ideas"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed.Count != 2 {
		t.Errorf("expected count 2, got %d", parsed.Count)
	}

	if parsed.Ideas[0]["channel"] != "customer" {
		t.Errorf("expected channel 'customer', got %v", parsed.Ideas[0]["channel"])
	}
}

func TestFormatIdeasArrayFormat(t *testing.T) {
	// Test with array format (no wrapper)
	input := `[
		{"id": 1, "name": "Idea 1", "channel": "feedback", "opportunities_count": 3}
	]`

	result := FormatIdeas(json.RawMessage(input))

	var parsed struct {
		Count int              `json:"count"`
		Ideas []map[string]any `json:"ideas"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed.Count != 1 {
		t.Errorf("expected count 1, got %d", parsed.Count)
	}
}

func TestFormatIdeasInvalidJSON(t *testing.T) {
	input := `invalid`
	result := FormatIdeas(json.RawMessage(input))
	if string(result) != input {
		t.Error("expected original input on parse error")
	}
}

func TestFormatOpportunities(t *testing.T) {
	// Test with wrapper format
	input := `{
		"results": [
			{"id": 1, "problem_statement": "Users need better search", "workflow_status": "researching", "ideas_count": 3},
			{"id": 2, "problem_statement": "Slow performance", "workflow_status": "validated", "ideas_count": 5}
		]
	}`

	result := FormatOpportunities(json.RawMessage(input))

	var parsed struct {
		Count         int              `json:"count"`
		Opportunities []map[string]any `json:"opportunities"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed.Count != 2 {
		t.Errorf("expected count 2, got %d", parsed.Count)
	}

	opp := parsed.Opportunities[0]
	if opp["workflow_status"] != "researching" {
		t.Errorf("expected workflow_status 'researching', got %v", opp["workflow_status"])
	}
}

func TestFormatOpportunitiesArrayFormat(t *testing.T) {
	input := `[{"id": 1, "problem_statement": "Problem", "workflow_status": "new", "ideas_count": 1}]`

	result := FormatOpportunities(json.RawMessage(input))

	var parsed struct {
		Count int `json:"count"`
	}
	json.Unmarshal(result, &parsed)

	if parsed.Count != 1 {
		t.Errorf("expected count 1, got %d", parsed.Count)
	}
}

func TestFormatOpportunitiesInvalidJSON(t *testing.T) {
	input := `invalid`
	result := FormatOpportunities(json.RawMessage(input))
	if string(result) != input {
		t.Error("expected original input on parse error")
	}
}

func TestFormatLaunches(t *testing.T) {
	input := `[
		{"id": 1, "name": "v1.0 Launch", "date": "2024-06-01", "status": "planned", "description": "ignored"},
		{"id": 2, "name": "v1.1 Launch", "date": "2024-09-01", "status": "in_progress"}
	]`

	result := FormatLaunches(json.RawMessage(input))

	var parsed struct {
		Count    int              `json:"count"`
		Launches []map[string]any `json:"launches"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed.Count != 2 {
		t.Errorf("expected count 2, got %d", parsed.Count)
	}

	launch := parsed.Launches[0]
	if launch["status"] != "planned" {
		t.Errorf("expected status 'planned', got %v", launch["status"])
	}
	if _, ok := launch["description"]; ok {
		t.Error("description should be filtered out")
	}
}

func TestFormatLaunchesInvalidJSON(t *testing.T) {
	input := `invalid`
	result := FormatLaunches(json.RawMessage(input))
	if string(result) != input {
		t.Error("expected original input on parse error")
	}
}

func BenchmarkFormatRoadmapList(b *testing.B) {
	input := json.RawMessage(`[
		{"id": 1, "name": "Roadmap 1", "updated_at": "2024-01-01"},
		{"id": 2, "name": "Roadmap 2", "updated_at": "2024-02-01"},
		{"id": 3, "name": "Roadmap 3", "updated_at": "2024-03-01"}
	]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatRoadmapList(input)
	}
}

func BenchmarkFormatBarsWithContext(b *testing.B) {
	bars := json.RawMessage(`[
		{"id": 1, "name": "Bar 1", "start_date": "2024-01-01", "end_date": "2024-03-31", "lane_id": 1},
		{"id": 2, "name": "Bar 2", "start_date": "2024-02-01", "end_date": "2024-04-30", "lane_id": 2},
		{"id": 3, "name": "Bar 3", "start_date": "2024-03-01", "end_date": "2024-05-31", "lane_id": 1}
	]`)
	lanes := json.RawMessage(`[
		{"id": 1, "name": "Lane A"},
		{"id": 2, "name": "Lane B"}
	]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatBarsWithContext(bars, lanes)
	}
}

func BenchmarkFormatLanes(b *testing.B) {
	input := json.RawMessage(`[
		{"id": 1, "name": "Engineering", "color": "#FF0000", "order": 1},
		{"id": 2, "name": "Design", "color": "#00FF00", "order": 2},
		{"id": 3, "name": "Product", "color": "#0000FF", "order": 3}
	]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatLanes(input)
	}
}

func BenchmarkFormatMilestones(b *testing.B) {
	input := json.RawMessage(`[
		{"id": 1, "name": "Launch", "date": "2024-06-01"},
		{"id": 2, "name": "Beta", "date": "2024-04-15"},
		{"id": 3, "name": "Alpha", "date": "2024-02-01"}
	]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatMilestones(input)
	}
}

func BenchmarkFormatObjectives(b *testing.B) {
	input := json.RawMessage(`[
		{"id": 1, "name": "Increase Revenue", "status": "on_track", "time_frame": "Q1 2024"},
		{"id": 2, "name": "Improve NPS", "status": "at_risk", "time_frame": "Q2 2024"},
		{"id": 3, "name": "Expand Market", "status": "on_track", "time_frame": "Q3 2024"}
	]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatObjectives(input)
	}
}

func BenchmarkFormatIdeas(b *testing.B) {
	input := json.RawMessage(`{
		"results": [
			{"id": 1, "name": "Feature A", "channel": "customer", "opportunities_count": 5},
			{"id": 2, "name": "Feature B", "channel": "internal", "opportunities_count": 2},
			{"id": 3, "name": "Feature C", "channel": "feedback", "opportunities_count": 8}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatIdeas(input)
	}
}

func BenchmarkFormatOpportunities(b *testing.B) {
	input := json.RawMessage(`{
		"results": [
			{"id": 1, "problem_statement": "Problem 1", "workflow_status": "researching", "ideas_count": 3},
			{"id": 2, "problem_statement": "Problem 2", "workflow_status": "validated", "ideas_count": 5},
			{"id": 3, "problem_statement": "Problem 3", "workflow_status": "new", "ideas_count": 1}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatOpportunities(input)
	}
}

func BenchmarkFormatLaunches(b *testing.B) {
	input := json.RawMessage(`[
		{"id": 1, "name": "v1.0 Launch", "date": "2024-06-01", "status": "planned"},
		{"id": 2, "name": "v1.1 Launch", "date": "2024-09-01", "status": "in_progress"},
		{"id": 3, "name": "v2.0 Launch", "date": "2024-12-01", "status": "planned"}
	]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatLaunches(input)
	}
}

// BenchmarkFormatBarsWithContextLarge tests performance with larger datasets.
func BenchmarkFormatBarsWithContextLarge(b *testing.B) {
	// Generate 100 bars across 10 lanes
	bars := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		bars[i] = map[string]any{
			"id":         i + 1,
			"name":       "Bar " + string(rune('A'+i%26)),
			"start_date": "2024-01-01",
			"end_date":   "2024-03-31",
			"lane_id":    (i % 10) + 1,
		}
	}
	barsJSON, _ := json.Marshal(bars)

	lanes := make([]map[string]any, 10)
	for i := 0; i < 10; i++ {
		lanes[i] = map[string]any{
			"id":   i + 1,
			"name": "Lane " + string(rune('A'+i)),
		}
	}
	lanesJSON, _ := json.Marshal(lanes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatBarsWithContext(json.RawMessage(barsJSON), json.RawMessage(lanesJSON))
	}
}
