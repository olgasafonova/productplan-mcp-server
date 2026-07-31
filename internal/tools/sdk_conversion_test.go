package tools

import (
	"encoding/json"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// The go-sdk migration (bead claude-code-config-4fc4.14) is only safe if
// converting a tool definition does not change what goes on the wire. The unit
// tests next to mcp.BuildTool prove that for hand-written inputs; this proves it
// for all 47 real definitions, which is the set that actually ships.
//
// Schemas and identity must survive byte-identically. Annotations deliberately
// must NOT: go-sdk v1.7.0 dropped omitempty on readOnlyHint and idempotentHint,
// so those two keys start appearing on every annotated tool. That difference is
// asserted separately below rather than being allowed to hide inside a general
// "nothing changed" check.
func TestBuildToolPreservesEveryDefinition(t *testing.T) {
	tools := BuildAllTools()
	if len(tools) == 0 {
		t.Fatal("expected tool definitions")
	}

	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			built := mcp.BuildTool(tool)

			if built.Name != tool.Name {
				t.Errorf("name changed: want %q, got %q", tool.Name, built.Name)
			}
			if built.Description != tool.Description {
				t.Errorf("description changed for %q", tool.Name)
			}

			assertMarshalsIdentically(t, "inputSchema", tool.InputSchema, built.InputSchema)

			if tool.OutputSchema == nil {
				if built.OutputSchema != nil {
					t.Errorf("output schema appeared from nowhere for %q", tool.Name)
				}
			} else {
				assertMarshalsIdentically(t, "outputSchema", tool.OutputSchema, built.OutputSchema)
			}

			if (tool.Annotations == nil) != (built.Annotations == nil) {
				t.Errorf("annotation presence changed for %q: local nil=%v, built nil=%v",
					tool.Name, tool.Annotations == nil, built.Annotations == nil)
			}
			if tool.Annotations != nil && built.Annotations != nil {
				if built.Annotations.ReadOnlyHint != tool.Annotations.ReadOnlyHint {
					t.Errorf("readOnlyHint changed for %q", tool.Name)
				}
				if built.Annotations.IdempotentHint != tool.Annotations.IdempotentHint {
					t.Errorf("idempotentHint changed for %q", tool.Name)
				}
				if built.Annotations.Title != tool.Annotations.Title {
					t.Errorf("annotation title changed for %q", tool.Name)
				}
			}
		})
	}
}

func assertMarshalsIdentically(t *testing.T, field string, local, built any) {
	t.Helper()
	want, err := json.Marshal(local)
	if err != nil {
		t.Fatalf("marshal local %s: %v", field, err)
	}
	got, err := json.Marshal(built)
	if err != nil {
		t.Fatalf("marshal built %s: %v", field, err)
	}
	if string(got) != string(want) {
		t.Errorf("%s changed on conversion:\n want %s\n  got %s", field, want, got)
	}
}

// Pins the one intended wire change, so it is a recorded decision rather than a
// surprise the first time a client reads tools/list after the migration.
func TestConversionAddsBareBoolHintsToAnnotatedTools(t *testing.T) {
	var annotated, gainedKeys int

	for _, tool := range BuildAllTools() {
		if tool.Annotations == nil {
			continue
		}
		annotated++

		before, err := json.Marshal(tool.Annotations)
		if err != nil {
			t.Fatalf("marshal local annotations for %q: %v", tool.Name, err)
		}
		after, err := json.Marshal(mcp.BuildTool(tool).Annotations)
		if err != nil {
			t.Fatalf("marshal built annotations for %q: %v", tool.Name, err)
		}

		var localKeys, builtKeys map[string]any
		if err := json.Unmarshal(before, &localKeys); err != nil {
			t.Fatalf("unmarshal local annotations for %q: %v", tool.Name, err)
		}
		if err := json.Unmarshal(after, &builtKeys); err != nil {
			t.Fatalf("unmarshal built annotations for %q: %v", tool.Name, err)
		}

		for _, key := range []string{"readOnlyHint", "idempotentHint"} {
			if _, ok := builtKeys[key]; !ok {
				t.Errorf("%q: expected %s to be present after conversion", tool.Name, key)
			}
		}
		if len(builtKeys) > len(localKeys) {
			gainedKeys++
		}
	}

	if annotated == 0 {
		t.Fatal("expected at least one annotated tool")
	}
	t.Logf("%d annotated tools; %d gained a previously-omitted false hint", annotated, gainedKeys)
}
