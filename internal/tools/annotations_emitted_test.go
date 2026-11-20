package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// unannotatedDefinitions returns the tool definitions as written, before
// BuildAllTools runs annotateTool and applyTitle over them. It must list the
// same groups as BuildAllTools; TestEveryToolIsClassified checks the counts.
func unannotatedDefinitions() []mcp.Tool {
	var tools []mcp.Tool
	tools = append(tools, roadmapTools()...)
	tools = append(tools, barTools()...)
	tools = append(tools, objectiveTools()...)
	tools = append(tools, ideaTools()...)
	tools = append(tools, launchTools()...)
	tools = append(tools, utilityTools()...)
	return tools
}

// TestEveryToolIsClassified fails when a tool would reach annotateTool's
// default case: it neither matches a known prefix nor declares annotations in
// its definition. The default keeps such a tool from shipping with nil
// annotations, but a real tool should be classified on purpose.
func TestEveryToolIsClassified(t *testing.T) {
	raw := unannotatedDefinitions()
	if len(raw) != len(BuildAllTools()) {
		t.Fatalf("unannotatedDefinitions has %d tools, BuildAllTools has %d; keep the group lists in sync",
			len(raw), len(BuildAllTools()))
	}
	for _, tool := range raw {
		if tool.Annotations != nil {
			continue
		}
		if readOnlyToolName(tool.Name) || strings.HasPrefix(tool.Name, "manage_") {
			continue
		}
		t.Errorf("tool %q matches no annotation prefix and declares no annotations; it falls through to the default case", tool.Name)
	}
}

// TestAnnotateToolDefaultCase pins the fall-through: an unrecognised tool gets
// explicit conservative annotations rather than nil.
func TestAnnotateToolDefaultCase(t *testing.T) {
	tool := mcp.Tool{Name: "frobnicate_widget"}
	annotateTool(&tool)
	if tool.Annotations == nil {
		t.Fatal("default case left Annotations nil")
	}
	a := tool.Annotations
	if a.ReadOnlyHint {
		t.Error("default case set ReadOnlyHint=true")
	}
	if a.DestructiveHint == nil || !*a.DestructiveHint {
		t.Error("default case should set DestructiveHint=true explicitly")
	}
	if a.OpenWorldHint == nil || !*a.OpenWorldHint {
		t.Error("default case should set OpenWorldHint=true explicitly")
	}
}

// TestEveryToolHasTitle checks every tool carries its explicit title from
// toolTitles, and that toolTitles has no entries for tools that do not exist.
func TestEveryToolHasTitle(t *testing.T) {
	tools := BuildAllTools()
	seen := make(map[string]bool, len(tools))
	for _, tool := range tools {
		seen[tool.Name] = true
		want, ok := toolTitles[tool.Name]
		if !ok || want == "" {
			t.Errorf("tool %q has no entry in toolTitles", tool.Name)
			continue
		}
		if tool.Annotations == nil || tool.Annotations.Title != want {
			got := ""
			if tool.Annotations != nil {
				got = tool.Annotations.Title
			}
			t.Errorf("tool %q title = %q, want %q", tool.Name, got, want)
		}
	}
	for name := range toolTitles {
		if !seen[name] {
			t.Errorf("toolTitles has an entry for unknown tool %q", name)
		}
	}
	if len(tools) != 50 {
		t.Errorf("expected 50 tools, got %d", len(tools))
	}
}

// TestEmittedAnnotations asserts what goes on the wire after conversion to the
// go-sdk type, per tool class, including false values and openWorldHint.
// openWorldHint and destructiveHint are pointers with omitempty, so the test
// checks the marshalled keys rather than only the Go fields.
func TestEmittedAnnotations(t *testing.T) {
	for _, tool := range BuildAllTools() {
		t.Run(tool.Name, func(t *testing.T) {
			built := mcp.BuildTool(tool)
			if built.Annotations == nil {
				t.Fatal("emitted annotations are nil")
			}
			raw, err := json.Marshal(built.Annotations)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var wire map[string]any
			if err := json.Unmarshal(raw, &wire); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if title, _ := wire["title"].(string); title == "" || title != toolTitles[tool.Name] {
				t.Errorf("title = %v, want %q", wire["title"], toolTitles[tool.Name])
			}
			if wire["openWorldHint"] != true {
				t.Errorf("openWorldHint = %v, want true (sent explicitly)", wire["openWorldHint"])
			}

			switch {
			case readOnlyToolName(tool.Name):
				if wire["readOnlyHint"] != true {
					t.Errorf("readOnlyHint = %v, want true", wire["readOnlyHint"])
				}
				if _, ok := wire["destructiveHint"]; ok {
					t.Errorf("read-only tool emits destructiveHint = %v", wire["destructiveHint"])
				}
			case tool.Name == "bulk_create_bars":
				if wire["readOnlyHint"] != false {
					t.Errorf("readOnlyHint = %v, want false", wire["readOnlyHint"])
				}
				if wire["destructiveHint"] != false {
					t.Errorf("destructiveHint = %v, want false (additive)", wire["destructiveHint"])
				}
			default:
				if wire["readOnlyHint"] != false {
					t.Errorf("readOnlyHint = %v, want false", wire["readOnlyHint"])
				}
				if wire["destructiveHint"] != true {
					t.Errorf("destructiveHint = %v, want true", wire["destructiveHint"])
				}
				if wire["idempotentHint"] != false {
					t.Errorf("idempotentHint = %v, want false", wire["idempotentHint"])
				}
			}
		})
	}
}
