package tools

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// argNamer is implemented by handlers built with typedHandler/queryHandler.
type argNamer interface{ argNames() []string }

// Article VI: dispatch rejects argument keys a tool's schema does not
// declare, so every argument a handler's Args struct reads must be declared,
// or it could never be passed. This includes deprecated compatibility
// arguments that exist only to return an explanatory error.
func TestSchemasDeclareEveryArgTheHandlerReads(t *testing.T) {
	cfg := Config{HealthChecker: &mockHealthChecker{}}
	for _, tool := range BuildAllTools() {
		h, ok := createHandler(tool.Name, cfg).(argNamer)
		if !ok {
			continue // HandlerFunc tools read no arguments
		}
		for _, arg := range h.argNames() {
			if _, declared := tool.InputSchema.Properties[arg]; !declared {
				t.Errorf("%s: handler reads %q but the InputSchema does not declare it", tool.Name, arg)
			}
		}
	}
}

// Every registered tool, dispatched through the real registry: an unknown
// key is rejected by name before the handler runs (no HTTP request is made),
// and a call using only declared keys is never rejected as unknown.
func TestDispatchRejectsUnknownArgumentsForEveryTool(t *testing.T) {
	requests := 0
	server := testServer(t, func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	defer server.Close()
	registry := mcp.NewRegistry()
	RegisterAll(registry, Config{Client: testClient(t, server), HealthChecker: &mockHealthChecker{}})

	for _, tool := range registry.Tools() {
		t.Run(tool.Name, func(t *testing.T) {
			_, err := registry.Call(context.Background(), tool.Name, map[string]any{"definitely_not_an_arg": "x"})
			if err == nil || !strings.Contains(err.Error(), `unknown argument for `+tool.Name+`: "definitely_not_an_arg"`) {
				t.Fatalf("unknown key not rejected by name: %v", err)
			}
			if len(tool.InputSchema.Properties) > 0 && !strings.Contains(err.Error(), "Valid arguments: ") {
				t.Errorf("rejection does not list the valid arguments: %v", err)
			}

			known := map[string]any{}
			for name := range tool.InputSchema.Properties {
				known[name] = nil
			}
			if err := tool.CheckArgumentKeys(known); err != nil {
				t.Errorf("declared keys rejected: %v", err)
			}
		})
	}
	if requests != 0 {
		t.Errorf("%d HTTP requests made; unknown-key rejection must happen before the handler", requests)
	}
}

// The deprecated manage_bar inputs stay declared so callers reach their
// explanatory errors instead of a generic unknown-argument rejection.
func TestManageBarKeepsDeprecatedArgsDeclared(t *testing.T) {
	for _, tool := range BuildAllTools() {
		if tool.Name != "manage_bar" {
			continue
		}
		var missing []string
		for _, arg := range []string{"legend_id", "effort", "parent_id", "container"} {
			if _, ok := tool.InputSchema.Properties[arg]; !ok {
				missing = append(missing, arg)
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			t.Errorf("manage_bar no longer declares %v", missing)
		}
		return
	}
	t.Fatal("manage_bar not defined")
}
