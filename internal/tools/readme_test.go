package tools

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// The README's "MCP tool reference" lists every tool by name and states
// the count. Both drifted before (47 listed while 50 shipped), so they are
// checked against BuildAllTools.
func TestReadmeToolReferenceMatchesBuildAllTools(t *testing.T) {
	raw, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(raw)
	start := strings.Index(readme, "<summary>MCP tool reference</summary>")
	if start < 0 {
		t.Fatal("README has no MCP tool reference section")
	}
	section := readme[start:]
	section = section[:strings.Index(section, "Example:")]

	listed := map[string]bool{}
	for _, line := range strings.Split(section, "\n") {
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		line = parenthetical.ReplaceAllString(line, "") // "(... `dry_run` supported)" names an argument
		for _, m := range toolName.FindAllStringSubmatch(line, -1) {
			listed[m[1]] = true
		}
	}

	tools := BuildAllTools()
	var missing []string
	built := map[string]bool{}
	for _, tool := range tools {
		built[tool.Name] = true
		if !listed[tool.Name] {
			missing = append(missing, tool.Name)
		}
	}
	var extra []string
	for name := range listed {
		if !built[name] {
			extra = append(extra, name)
		}
	}
	slices.Sort(missing)
	slices.Sort(extra)
	if len(missing) > 0 || len(extra) > 0 {
		t.Errorf("README tool list drifted: missing %v, not a tool %v", missing, extra)
	}
	reads := 0
	for _, tool := range tools {
		if tool.Annotations != nil && tool.Annotations.ReadOnlyHint {
			reads++
		}
	}
	want := fmt.Sprintf("%d tools available: %d READ tools and %d WRITE tools", len(tools), reads, len(tools)-reads)
	if !strings.Contains(section, want) {
		t.Errorf("README does not say %q", want)
	}
}

var (
	parenthetical = regexp.MustCompile(`\([^)]*\)`)
	toolName      = regexp.MustCompile("`([a-z_]+)`")
)
