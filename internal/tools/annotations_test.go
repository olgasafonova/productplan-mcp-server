package tools

import (
	"strings"
	"testing"
)

// TestManageToolsHaveDestructiveHint is the HG-3 regression test. Every
// manage_* tool dispatches across action=create/update/delete and supports
// cascade deletes. MCP clients use DestructiveHint to gate user-confirmation
// prompts on irreversible operations; without the annotation, hosts that
// distinguish destructive calls cannot challenge them.
func TestManageToolsHaveDestructiveHint(t *testing.T) {
	tools := BuildAllTools()

	manageCount := 0
	for _, tool := range tools {
		if !strings.HasPrefix(tool.Name, "manage_") {
			continue
		}
		manageCount++
		if tool.Annotations == nil {
			t.Errorf("manage_* tool %q has no Annotations", tool.Name)
			continue
		}
		if tool.Annotations.DestructiveHint == nil {
			t.Errorf("manage_* tool %q has nil DestructiveHint (HG-3 regression)", tool.Name)
			continue
		}
		if !*tool.Annotations.DestructiveHint {
			t.Errorf("manage_* tool %q has DestructiveHint=false; expected true (HG-3 regression)", tool.Name)
		}
	}

	// Sanity check: must find all 12 manage_* tools enumerated by the scan.
	if manageCount < 12 {
		t.Errorf("expected at least 12 manage_* tools, got %d (did the auto-annotate loop change?)", manageCount)
	}
}

// TestManageToolsDoNotClaimIdempotency asserts the previous "idempotency lie"
// is gone. manage_* tools support action=create (NOT idempotent: two calls =
// two records) and action=delete (NOT idempotent: second call 404s). A blanket
// IdempotentHint misleads retry-aware clients into duplicate writes.
func TestManageToolsDoNotClaimIdempotency(t *testing.T) {
	tools := BuildAllTools()

	for _, tool := range tools {
		if !strings.HasPrefix(tool.Name, "manage_") {
			continue
		}
		if tool.Annotations == nil {
			continue
		}
		if tool.Annotations.IdempotentHint {
			t.Errorf("manage_* tool %q has IdempotentHint=true; idempotency varies per action and should not be claimed at the tool level", tool.Name)
		}
	}
}

// TestReadOnlyToolsAnnotation guards the success path for read-only tools.
// They must keep ReadOnlyHint=true and IdempotentHint=true and must NOT carry
// DestructiveHint.
func TestReadOnlyToolsAnnotation(t *testing.T) {
	tools := BuildAllTools()

	roCount := 0
	for _, tool := range tools {
		isReadOnly := strings.HasPrefix(tool.Name, "get_") ||
			strings.HasPrefix(tool.Name, "list_") ||
			strings.HasPrefix(tool.Name, "check_") ||
			tool.Name == "health_check"
		if !isReadOnly {
			continue
		}
		roCount++
		if tool.Annotations == nil {
			t.Errorf("read-only tool %q has no Annotations", tool.Name)
			continue
		}
		if !tool.Annotations.ReadOnlyHint {
			t.Errorf("read-only tool %q missing ReadOnlyHint=true", tool.Name)
		}
		if !tool.Annotations.IdempotentHint {
			t.Errorf("read-only tool %q missing IdempotentHint=true", tool.Name)
		}
		if tool.Annotations.DestructiveHint != nil && *tool.Annotations.DestructiveHint {
			t.Errorf("read-only tool %q has DestructiveHint=true; should not be set", tool.Name)
		}
	}

	if roCount == 0 {
		t.Error("found no read-only tools; expected several get_*/list_* tools")
	}
}

// isReadOnlyName mirrors the read-only classification in BuildAllTools.
func isReadOnlyName(name string) bool {
	return strings.HasPrefix(name, "get_") ||
		strings.HasPrefix(name, "list_") ||
		strings.HasPrefix(name, "check_") ||
		name == "health_check"
}

// TestReadOnlyToolsHaveOutputSchema guards Code Mode eligibility: every
// read-only tool must declare an OutputSchema describing the FormattedResponse
// wrapper, so clients can drive it against a typed output shape.
func TestReadOnlyToolsHaveOutputSchema(t *testing.T) {
	tools := BuildAllTools()

	roCount := 0
	for _, tool := range tools {
		if !isReadOnlyName(tool.Name) {
			continue
		}
		roCount++
		if tool.OutputSchema == nil {
			t.Errorf("read-only tool %q missing OutputSchema (Code Mode ineligible)", tool.Name)
			continue
		}
		if tool.OutputSchema.Type != "object" {
			t.Errorf("read-only tool %q OutputSchema.Type = %q, want \"object\"", tool.Name, tool.OutputSchema.Type)
		}
		if _, ok := tool.OutputSchema.Properties["summary"]; !ok {
			t.Errorf("read-only tool %q OutputSchema missing \"summary\" property", tool.Name)
		}
		if _, ok := tool.OutputSchema.Properties["data"]; !ok {
			t.Errorf("read-only tool %q OutputSchema missing \"data\" property", tool.Name)
		}
	}

	if roCount != 35 {
		t.Errorf("expected 35 read-only tools, found %d", roCount)
	}
}

// TestManageToolsHaveNoOutputSchema asserts write tools do not declare an
// OutputSchema. Their results vary per action (create/update/delete) and are
// not part of this Code Mode pass; leaving them unset keeps the contract honest.
func TestManageToolsHaveNoOutputSchema(t *testing.T) {
	tools := BuildAllTools()

	for _, tool := range tools {
		if !strings.HasPrefix(tool.Name, "manage_") {
			continue
		}
		if tool.OutputSchema != nil {
			t.Errorf("manage tool %q unexpectedly has an OutputSchema", tool.Name)
		}
	}
}
