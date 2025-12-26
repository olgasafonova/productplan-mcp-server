package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// mockHealthChecker implements HealthChecker for testing.
type mockHealthChecker struct {
	report map[string]any
}

func (m *mockHealthChecker) Check(ctx context.Context, deep bool) any {
	if m.report != nil {
		return m.report
	}
	return map[string]any{
		"status": "healthy",
		"deep":   deep,
	}
}

func testServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func testClient(t *testing.T, server *httptest.Server) *api.Client {
	t.Helper()
	client, err := api.New(api.Config{
		Token:   "test-token",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

func TestRegisterAll(t *testing.T) {
	server := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})
	defer server.Close()

	client := testClient(t, server)
	registry := mcp.NewRegistry()
	checker := &mockHealthChecker{}

	RegisterAll(registry, Config{
		Client:        client,
		HealthChecker: checker,
	})

	// Should have all tools registered
	if registry.Count() < 30 {
		t.Errorf("expected at least 30 tools registered, got %d", registry.Count())
	}

	// Verify specific handlers exist
	handlers := []string{
		"list_roadmaps",
		"get_roadmap",
		"manage_bar",
		"list_objectives",
		"health_check",
	}

	for _, name := range handlers {
		if _, ok := registry.Handler(name); !ok {
			t.Errorf("expected handler %q to be registered", name)
		}
	}
}

func TestCreateHandlerReturnsValidHandlers(t *testing.T) {
	server := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"id": "123"})
	})
	defer server.Close()

	client := testClient(t, server)
	checker := &mockHealthChecker{}
	cfg := Config{Client: client, HealthChecker: checker}

	tests := []struct {
		name string
		args map[string]any
	}{
		{"list_roadmaps", nil},
		{"get_roadmap", map[string]any{"roadmap_id": "123"}},
		{"get_roadmap_bars", map[string]any{"roadmap_id": "123"}},
		{"get_bar", map[string]any{"bar_id": "456"}},
		{"list_objectives", nil},
		{"health_check", map[string]any{"deep": true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := createHandler(tt.name, cfg)
			if handler == nil {
				t.Fatalf("expected handler for %q", tt.name)
			}

			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("handler %q returned error: %v", tt.name, err)
			}
		})
	}
}

func TestCreateHandlerUnknownTool(t *testing.T) {
	cfg := Config{}
	handler := createHandler("unknown_tool", cfg)

	// Should return a no-op handler
	result, err := handler.Handle(context.Background(), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for unknown tool")
	}
}

func TestHealthCheckHandler(t *testing.T) {
	checker := &mockHealthChecker{
		report: map[string]any{
			"status":     "healthy",
			"version":    "1.0.0",
			"rate_limit": "ok",
		},
	}

	handler := healthCheckHandler(checker)

	result, err := handler.Handle(context.Background(), map[string]any{"deep": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report map[string]any
	if err := json.Unmarshal(result, &report); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if report["status"] != "healthy" {
		t.Errorf("expected status healthy, got %v", report["status"])
	}
}

func BenchmarkRegisterAll(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	}))
	defer server.Close()

	client, _ := api.New(api.Config{
		Token:   "bench-token",
		BaseURL: server.URL,
	})
	checker := &mockHealthChecker{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := mcp.NewRegistry()
		RegisterAll(registry, Config{
			Client:        client,
			HealthChecker: checker,
		})
	}
}

func BenchmarkCreateHandler(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"id": "123"})
	}))
	defer server.Close()

	client, _ := api.New(api.Config{
		Token:   "bench-token",
		BaseURL: server.URL,
	})
	checker := &mockHealthChecker{}
	cfg := Config{Client: client, HealthChecker: checker}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createHandler("list_roadmaps", cfg)
	}
}

func BenchmarkHandlerCall(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "Roadmap 1"},
			{"id": 2, "name": "Roadmap 2"},
		})
	}))
	defer server.Close()

	client, _ := api.New(api.Config{
		Token:   "bench-token",
		BaseURL: server.URL,
	})
	checker := &mockHealthChecker{}
	cfg := Config{Client: client, HealthChecker: checker}

	handler := createHandler("list_roadmaps", cfg)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx, nil)
	}
}
