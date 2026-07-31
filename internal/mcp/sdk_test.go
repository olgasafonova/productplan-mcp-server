package mcp

import (
	"encoding/json"
	"testing"
)

// Local to the test: internal/tools has its own floatPtr/boolPtr, but this
// package cannot import it without a cycle.
func floatPtr(v float64) *float64 { return &v }
func boolPtr(v bool) *bool        { return &v }

func TestBuildToolCopiesIdentity(t *testing.T) {
	built := BuildTool(Tool{
		Name:        "list_roadmaps",
		Description: "List roadmaps",
		InputSchema: InputSchema{Type: "object"},
	})

	if built.Name != "list_roadmaps" {
		t.Errorf("expected name 'list_roadmaps', got %q", built.Name)
	}
	if built.Description != "List roadmaps" {
		t.Errorf("expected description 'List roadmaps', got %q", built.Description)
	}
	if built.InputSchema == nil {
		t.Error("expected InputSchema to be set")
	}
}

// The whole point of the conversion is that the bytes on the wire do not move.
// Marshalling the local schema and the converted tool's schema and comparing
// them is the only assertion that actually proves that; comparing Go values
// would pass even if the SDK re-encoded the schema differently.
func TestBuildToolSchemasMarshalIdentically(t *testing.T) {
	in := InputSchema{
		Type: "object",
		Properties: map[string]Property{
			"id":    {Type: "string", Description: "Roadmap ID"},
			"limit": {Type: "integer", Description: "Max results", Minimum: floatPtr(1)},
		},
		Required: []string{"id"},
	}
	out := &OutputSchema{
		Type:       "object",
		Properties: map[string]Property{"summary": {Type: "string", Description: "Summary"}},
		Required:   []string{"summary"},
	}

	built := BuildTool(Tool{Name: "get_roadmap", InputSchema: in, OutputSchema: out})

	wantIn, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal local input schema: %v", err)
	}
	gotIn, err := json.Marshal(built.InputSchema)
	if err != nil {
		t.Fatalf("marshal built input schema: %v", err)
	}
	if string(gotIn) != string(wantIn) {
		t.Errorf("input schema changed on conversion:\n want %s\n  got %s", wantIn, gotIn)
	}

	wantOut, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal local output schema: %v", err)
	}
	gotOut, err := json.Marshal(built.OutputSchema)
	if err != nil {
		t.Fatalf("marshal built output schema: %v", err)
	}
	if string(gotOut) != string(wantOut) {
		t.Errorf("output schema changed on conversion:\n want %s\n  got %s", wantOut, gotOut)
	}
}

func TestBuildToolOmitsOutputSchemaWhenUnset(t *testing.T) {
	built := BuildTool(Tool{Name: "manage_bar", InputSchema: InputSchema{Type: "object"}})
	if built.OutputSchema != nil {
		t.Errorf("expected nil OutputSchema for a tool that declares none, got %v", built.OutputSchema)
	}
}

func TestBuildToolConvertsAnnotations(t *testing.T) {
	built := BuildTool(Tool{
		Name: "get_roadmap",
		Annotations: &ToolAnnotations{
			Title:          "Get Roadmap",
			ReadOnlyHint:   true,
			IdempotentHint: true,
			OpenWorldHint:  boolPtr(false),
		},
	})

	if built.Annotations == nil {
		t.Fatal("expected annotations to be converted")
	}
	if built.Annotations.Title != "Get Roadmap" {
		t.Errorf("expected title 'Get Roadmap', got %q", built.Annotations.Title)
	}
	if !built.Annotations.ReadOnlyHint || !built.Annotations.IdempotentHint {
		t.Error("expected readOnly and idempotent hints to survive conversion")
	}
	if built.Annotations.OpenWorldHint == nil || *built.Annotations.OpenWorldHint {
		t.Error("expected openWorldHint to survive as an explicit false")
	}
	if built.Annotations.DestructiveHint != nil {
		t.Error("expected an undeclared destructiveHint to stay nil rather than becoming false")
	}
	// The SDK Tool's own Title stays empty on purpose: the display name lives on
	// the annotations, which is the key this server already emits.
	if built.Title != "" {
		t.Errorf("expected Tool.Title to stay empty, got %q", built.Title)
	}
}

func TestBuildToolNilAnnotationsStayNil(t *testing.T) {
	built := BuildTool(Tool{Name: "check_status", InputSchema: InputSchema{Type: "object"}})
	if built.Annotations != nil {
		t.Errorf("expected nil annotations for a tool that declares none, got %+v", built.Annotations)
	}
}

// go-sdk v1.7.0 dropped omitempty on both bare-bool hints, so they now ship on
// every tool that declares any annotations at all. Pinning that here means the
// change is asserted rather than discovered later against a live server.
func TestBuildToolAlwaysEmitsBareBoolHints(t *testing.T) {
	built := BuildTool(Tool{
		Name:        "manage_bar",
		Annotations: &ToolAnnotations{Title: "Manage Bar"},
	})

	encoded, err := json.Marshal(built.Annotations)
	if err != nil {
		t.Fatalf("marshal annotations: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal annotations: %v", err)
	}
	for _, key := range []string{"readOnlyHint", "idempotentHint"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("expected %q to be present even when false, got %s", key, encoded)
		}
	}
	for _, key := range []string{"destructiveHint", "openWorldHint"} {
		if _, ok := decoded[key]; ok {
			t.Errorf("expected %q to be omitted when unset, got %s", key, encoded)
		}
	}
}
