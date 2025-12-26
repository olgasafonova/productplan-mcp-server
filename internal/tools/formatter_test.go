package tools

import (
	"encoding/json"
	"testing"
)

func TestFormatList_Empty(t *testing.T) {
	data := json.RawMessage(`[]`)
	result, err := FormatList(data, "roadmap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp FormattedResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Summary != "No roadmaps found" {
		t.Errorf("expected 'No roadmaps found', got %q", resp.Summary)
	}
}

func TestFormatList_Single(t *testing.T) {
	data := json.RawMessage(`[{"id": "1", "name": "Test"}]`)
	result, err := FormatList(data, "bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp FormattedResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Summary != "Found 1 bar" {
		t.Errorf("expected 'Found 1 bar', got %q", resp.Summary)
	}
}

func TestFormatList_Multiple(t *testing.T) {
	data := json.RawMessage(`[{"id": "1"}, {"id": "2"}, {"id": "3"}]`)
	result, err := FormatList(data, "idea")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp FormattedResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Summary != "Found 3 ideas" {
		t.Errorf("expected 'Found 3 ideas', got %q", resp.Summary)
	}
}

func TestFormatList_NotArray(t *testing.T) {
	data := json.RawMessage(`{"id": "1", "name": "Test"}`)
	result, err := FormatList(data, "roadmap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return original data unchanged
	if string(result) != string(data) {
		t.Errorf("expected original data, got %s", result)
	}
}

func TestFormatItem(t *testing.T) {
	data := json.RawMessage(`{"id": "abc123", "name": "My Roadmap"}`)
	result, err := FormatItem(data, "roadmap", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp FormattedResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Summary != "Roadmap abc123 retrieved successfully" {
		t.Errorf("unexpected summary: %q", resp.Summary)
	}
}

func TestFormatAction_Create(t *testing.T) {
	data := json.RawMessage(`{"id": "new123", "name": "New Bar"}`)
	result, err := FormatAction(data, "create", "bar", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp FormattedResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Summary != "Bar created successfully" {
		t.Errorf("unexpected summary: %q", resp.Summary)
	}
}

func TestFormatAction_Update(t *testing.T) {
	data := json.RawMessage(`{"id": "abc123", "name": "Updated"}`)
	result, err := FormatAction(data, "update", "lane", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp FormattedResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Summary != "Lane abc123 updated successfully" {
		t.Errorf("unexpected summary: %q", resp.Summary)
	}
}

func TestFormatAction_Delete(t *testing.T) {
	data := json.RawMessage(`{}`)
	result, err := FormatAction(data, "delete", "milestone", "xyz789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp FormattedResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Summary != "Milestone xyz789 deleted successfully" {
		t.Errorf("unexpected summary: %q", resp.Summary)
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		word     string
		count    int
		expected string
	}{
		{"roadmap", 0, "roadmaps"},
		{"roadmap", 1, "roadmap"},
		{"roadmap", 2, "roadmaps"},
		{"bar", 5, "bars"},
		{"idea", 1, "idea"},
	}

	for _, tc := range tests {
		result := pluralize(tc.word, tc.count)
		if result != tc.expected {
			t.Errorf("pluralize(%q, %d) = %q, expected %q", tc.word, tc.count, result, tc.expected)
		}
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"roadmap", "Roadmap"},
		{"bar", "Bar"},
		{"idea", "Idea"},
		{"", ""},
	}

	for _, tc := range tests {
		result := capitalize(tc.input)
		if result != tc.expected {
			t.Errorf("capitalize(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
