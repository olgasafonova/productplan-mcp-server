package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestToolOutputSchemaOmittedWhenNil(t *testing.T) {
	data, err := json.Marshal(Tool{Name: "t", Description: "d", InputSchema: InputSchema{Type: "object"}})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(data), "outputSchema") {
		t.Errorf("expected outputSchema to be omitted when nil, got %s", data)
	}
}

func TestToolOutputSchemaMarshaled(t *testing.T) {
	tool := Tool{
		Name:         "t",
		Description:  "d",
		InputSchema:  InputSchema{Type: "object"},
		OutputSchema: &OutputSchema{Type: "object", Required: []string{"summary"}},
	}
	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(data), `"outputSchema"`) {
		t.Errorf("expected outputSchema in marshaled tool, got %s", data)
	}
}

func TestToolMarshal(t *testing.T) {
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"id": {Type: "string", Description: "The ID"},
			},
			Required: []string{"id"},
		},
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(data, &parsed)

	if parsed["name"] != "test_tool" {
		t.Errorf("expected name 'test_tool', got %v", parsed["name"])
	}
	if parsed["description"] != "A test tool" {
		t.Errorf("expected description 'A test tool', got %v", parsed["description"])
	}
}
