package tools

import "testing"

// directlyWired names the tools createHandler builds without
// handlerConstructors, because they need something other than the client.
var directlyWired = map[string]bool{"health_check": true}

// Constitution Article I: every advertised tool has a handler, and every
// handler constructor belongs to an advertised tool. A tool missing from
// handlerConstructors would register but answer "unknown tool" at call time;
// an orphan constructor is dead code that looks wired.
func TestEveryToolHasAHandlerAndNoOrphans(t *testing.T) {
	advertised := map[string]bool{}
	for _, tool := range BuildAllTools() {
		if advertised[tool.Name] {
			t.Errorf("tool %s is defined twice", tool.Name)
		}
		advertised[tool.Name] = true
		_, constructed := handlerConstructors[tool.Name]
		switch {
		case constructed && directlyWired[tool.Name]:
			t.Errorf("tool %s is both in handlerConstructors and wired directly", tool.Name)
		case !constructed && !directlyWired[tool.Name]:
			t.Errorf("tool %s has no entry in handlerConstructors", tool.Name)
		}
	}
	for name := range handlerConstructors {
		if !advertised[name] {
			t.Errorf("handlerConstructors[%q] has no tool in BuildAllTools", name)
		}
	}
	for name := range directlyWired {
		if !advertised[name] {
			t.Errorf("directly wired tool %q has no tool in BuildAllTools", name)
		}
	}
}
