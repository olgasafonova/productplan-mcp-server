package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

// TestIntegrationFullSession tests a complete MCP session from initialization to tool calls.
func TestIntegrationFullSession(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{
			Name:        "list_roadmaps",
			Description: "List all roadmaps",
			InputSchema: InputSchema{Type: "object"},
		},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{"count": 2, "roadmaps": [{"id": 1}, {"id": 2}]}`), nil
		},
	)
	registry.RegisterFunc(
		Tool{
			Name:        "get_roadmap",
			Description: "Get roadmap details",
			InputSchema: InputSchema{
				Type:     "object",
				Required: []string{"roadmap_id"},
				Properties: map[string]Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
				},
			},
		},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			id := args["roadmap_id"].(string)
			return json.RawMessage(`{"id": "` + id + `", "name": "Test Roadmap"}`), nil
		},
	)

	// Simulate a complete MCP session
	session := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_roadmaps","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_roadmap","arguments":{"roadmap_id":"123"}}}`,
	}

	input := strings.NewReader(strings.Join(session, "\n"))
	var output bytes.Buffer

	server := NewServer("ProductPlan", "1.0.0", registry,
		WithIO(input, &output),
		WithLogger(logging.Nop()),
		WithInstructions("Test instructions"),
	)

	err := server.Run(context.Background())
	if err != nil {
		t.Fatalf("server.Run failed: %v", err)
	}

	// Parse responses
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")

	// Expected: 4 responses (initialize, tools/list, 2x tools/call)
	// notifications/initialized doesn't produce a response
	if len(lines) != 4 {
		t.Fatalf("expected 4 responses, got %d: %v", len(lines), lines)
	}

	// Verify initialize response
	var initResp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[0]), &initResp); err != nil {
		t.Fatalf("failed to parse initialize response: %v", err)
	}
	if initResp.ID != float64(1) {
		t.Errorf("expected id 1, got %v", initResp.ID)
	}
	if initResp.Error != nil {
		t.Errorf("initialize error: %v", initResp.Error)
	}

	// Verify tools/list response
	var listResp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[1]), &listResp); err != nil {
		t.Fatalf("failed to parse tools/list response: %v", err)
	}
	if listResp.ID != float64(2) {
		t.Errorf("expected id 2, got %v", listResp.ID)
	}

	// Check tools were returned
	resultMap, ok := listResp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected result map, got %T", listResp.Result)
	}
	tools, ok := resultMap["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got %T", resultMap["tools"])
	}
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}

	// Verify tool call responses
	var callResp1, callResp2 JSONRPCResponse
	json.Unmarshal([]byte(lines[2]), &callResp1)
	json.Unmarshal([]byte(lines[3]), &callResp2)

	if callResp1.Error != nil {
		t.Errorf("list_roadmaps call error: %v", callResp1.Error)
	}
	if callResp2.Error != nil {
		t.Errorf("get_roadmap call error: %v", callResp2.Error)
	}
}

// TestIntegrationConcurrentToolCalls simulates multiple tool calls in sequence.
func TestIntegrationConcurrentToolCalls(t *testing.T) {
	callCount := 0
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "counter"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			callCount++
			return json.RawMessage(`{"call": ` + fmt.Sprintf("%d", callCount) + `}`), nil
		},
	)

	var calls []string
	for i := 1; i <= 10; i++ {
		calls = append(calls, fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":{"name":"counter","arguments":{}}}`, i))
	}

	input := strings.NewReader(strings.Join(calls, "\n") + "\n")
	var output bytes.Buffer

	server := NewServer("test", "1.0.0", registry,
		WithIO(input, &output),
		WithLogger(logging.Nop()),
	)

	server.Run(context.Background())

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 10 {
		t.Errorf("expected 10 responses, got %d", len(lines))
	}

	if callCount != 10 {
		t.Errorf("expected 10 tool calls, got %d", callCount)
	}
}

// TestIntegrationErrorRecovery tests that the server continues after errors.
func TestIntegrationErrorRecovery(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "good"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{"status": "ok"}`), nil
		},
	)

	// Mix of valid, invalid, and unknown method calls
	session := []string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"good","arguments":{}}}`,
		`{invalid json`,
		`{"jsonrpc":"2.0","id":2,"method":"unknown/method"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"nonexistent","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"good","arguments":{}}}`,
	}

	input := strings.NewReader(strings.Join(session, "\n"))
	var output bytes.Buffer

	server := NewServer("test", "1.0.0", registry,
		WithIO(input, &output),
		WithLogger(logging.Nop()),
	)

	err := server.Run(context.Background())
	if err != nil {
		t.Fatalf("server should not return error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	// Should have 4 responses: 2 good calls, 1 method not found, 1 tool not found
	// Invalid JSON is skipped with no response
	if len(lines) != 4 {
		t.Errorf("expected 4 responses, got %d: %v", len(lines), lines)
	}

	// First should succeed
	var resp1 JSONRPCResponse
	json.Unmarshal([]byte(lines[0]), &resp1)
	if resp1.Error != nil {
		t.Errorf("first call should succeed: %v", resp1.Error)
	}

	// Second should be method not found
	var resp2 JSONRPCResponse
	json.Unmarshal([]byte(lines[1]), &resp2)
	if resp2.Error == nil || resp2.Error.Code != ErrMethodNotFound {
		t.Errorf("second should be method not found: %v", resp2.Error)
	}

	// Last should succeed
	var resp4 JSONRPCResponse
	json.Unmarshal([]byte(lines[3]), &resp4)
	if resp4.Error != nil {
		t.Errorf("last call should succeed: %v", resp4.Error)
	}
}

// TestIntegrationToolWithComplexArgs tests tools with complex argument structures.
func TestIntegrationToolWithComplexArgs(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{
			Name: "create_item",
			InputSchema: InputSchema{
				Type:     "object",
				Required: []string{"name"},
				Properties: map[string]Property{
					"name":        {Type: "string", Description: "Item name"},
					"description": {Type: "string", Description: "Item description"},
					"tags":        {Type: "array", Description: "Item tags"},
					"metadata":    {Type: "object", Description: "Item metadata"},
				},
			},
		},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			name := args["name"].(string)
			return json.RawMessage(`{"created": "` + name + `"}`), nil
		},
	)

	params := `{
		"name": "create_item",
		"arguments": {
			"name": "Test Item",
			"description": "A test item with complex args",
			"tags": ["tag1", "tag2"],
			"metadata": {"key": "value", "nested": {"deep": true}}
		}
	}`

	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":` + params + `}`)
	var output bytes.Buffer

	server := NewServer("test", "1.0.0", registry,
		WithIO(input, &output),
		WithLogger(logging.Nop()),
	)

	server.Run(context.Background())

	var resp JSONRPCResponse
	json.Unmarshal(output.Bytes(), &resp)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

// TestIntegrationContextCancellation tests server shutdown via context.
func TestIntegrationContextCancellation(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "test"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{}`), nil
		},
	)

	// Create a reader that will block
	input := &blockingReader{done: make(chan struct{})}
	var output bytes.Buffer

	server := NewServer("test", "1.0.0", registry,
		WithIO(input, &output),
		WithLogger(logging.Nop()),
	)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Run(ctx)
	}()

	// Cancel context
	cancel()

	// Signal the blocking reader to complete
	close(input.done)

	err := <-errCh
	// The error will be wrapped as a scanner error
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got %v", err)
	}
}

// blockingReader blocks until done channel is closed.
type blockingReader struct {
	done chan struct{}
}

func (r *blockingReader) Read(p []byte) (int, error) {
	<-r.done
	return 0, context.Canceled
}

// BenchmarkIntegrationFullSession benchmarks a complete session.
func BenchmarkIntegrationFullSession(b *testing.B) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "list"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{"items": []}`), nil
		},
	)
	registry.RegisterFunc(
		Tool{Name: "get"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{"item": {}}`), nil
		},
	)

	session := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get","arguments":{}}}`,
	}, "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := strings.NewReader(session)
		var output bytes.Buffer

		server := NewServer("bench", "1.0.0", registry,
			WithIO(input, &output),
			WithLogger(logging.Nop()),
		)

		server.Run(context.Background())
	}
}
