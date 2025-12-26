package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

func TestNewServer(t *testing.T) {
	registry := NewRegistry()
	server := NewServer("test", "1.0.0", registry)

	if server.name != "test" {
		t.Errorf("expected name 'test', got %q", server.name)
	}
	if server.version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", server.version)
	}
}

func TestServerWithOptions(t *testing.T) {
	registry := NewRegistry()
	var buf bytes.Buffer
	logger := logging.NewWithWriter(&buf, logging.LevelDebug)

	server := NewServer("test", "1.0.0", registry,
		WithInstructions("Test instructions"),
		WithLogger(logger),
	)

	if server.instructions != "Test instructions" {
		t.Errorf("expected instructions, got %q", server.instructions)
	}
}

func TestServerInitialize(t *testing.T) {
	registry := NewRegistry()
	server := NewServer("test-server", "2.0.0", registry,
		WithInstructions("Use these tools wisely"),
	)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	resp := server.ProcessRequest(context.Background(), req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(InitializeResult)
	if !ok {
		t.Fatalf("expected InitializeResult, got %T", resp.Result)
	}

	if result.ProtocolVersion != ProtocolVersion {
		t.Errorf("expected protocol version %q, got %q", ProtocolVersion, result.ProtocolVersion)
	}
	if result.ServerInfo.Name != "test-server" {
		t.Errorf("expected server name 'test-server', got %q", result.ServerInfo.Name)
	}
	if result.ServerInfo.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %q", result.ServerInfo.Version)
	}
	if result.Instructions != "Use these tools wisely" {
		t.Errorf("expected instructions, got %q", result.Instructions)
	}
}

func TestServerNotificationsInitialized(t *testing.T) {
	registry := NewRegistry()
	server := NewServer("test", "1.0.0", registry)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      nil,
		Method:  "notifications/initialized",
	}

	resp := server.ProcessRequest(context.Background(), req)

	// Should return empty response for notifications
	if resp.JSONRPC != "" {
		t.Error("expected empty JSONRPC for notification")
	}
}

func TestServerToolsList(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "tool1", Description: "First tool"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return nil, nil
		},
	)
	registry.RegisterFunc(
		Tool{Name: "tool2", Description: "Second tool"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return nil, nil
		},
	)

	server := NewServer("test", "1.0.0", registry)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	resp := server.ProcessRequest(context.Background(), req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(ToolsListResult)
	if !ok {
		t.Fatalf("expected ToolsListResult, got %T", resp.Result)
	}

	if len(result.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result.Tools))
	}
}

func TestServerToolsCall(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "echo", Description: "Echo tool"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			msg := args["message"].(string)
			return json.RawMessage(`{"echo": "` + msg + `"}`), nil
		},
	)

	server := NewServer("test", "1.0.0", registry)

	params, _ := json.Marshal(ToolCallParams{
		Name:      "echo",
		Arguments: map[string]any{"message": "hello world"},
	})

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  params,
	}

	resp := server.ProcessRequest(context.Background(), req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(ToolResult)
	if !ok {
		t.Fatalf("expected ToolResult, got %T", resp.Result)
	}

	if result.IsError {
		t.Error("expected IsError to be false")
	}

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}

	if !strings.Contains(result.Content[0].Text, "hello world") {
		t.Errorf("expected echo result, got %q", result.Content[0].Text)
	}
}

func TestServerToolsCallError(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "fail", Description: "Failing tool"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return nil, context.DeadlineExceeded
		},
	)

	server := NewServer("test", "1.0.0", registry)

	params, _ := json.Marshal(ToolCallParams{
		Name:      "fail",
		Arguments: map[string]any{},
	})

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  params,
	}

	resp := server.ProcessRequest(context.Background(), req)

	if resp.Error != nil {
		t.Fatalf("unexpected RPC error: %v", resp.Error)
	}

	result, ok := resp.Result.(ToolResult)
	if !ok {
		t.Fatalf("expected ToolResult, got %T", resp.Result)
	}

	if !result.IsError {
		t.Error("expected IsError to be true")
	}

	if !strings.Contains(result.Content[0].Text, "Error:") {
		t.Errorf("expected error message, got %q", result.Content[0].Text)
	}
}

func TestServerToolsCallInvalidParams(t *testing.T) {
	registry := NewRegistry()
	server := NewServer("test", "1.0.0", registry)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  json.RawMessage(`{invalid json`),
	}

	resp := server.ProcessRequest(context.Background(), req)

	if resp.Error == nil {
		t.Fatal("expected error for invalid params")
	}
	if resp.Error.Code != ErrInvalidParams {
		t.Errorf("expected code %d, got %d", ErrInvalidParams, resp.Error.Code)
	}
}

func TestServerMethodNotFound(t *testing.T) {
	registry := NewRegistry()
	server := NewServer("test", "1.0.0", registry)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
	}

	resp := server.ProcessRequest(context.Background(), req)

	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != ErrMethodNotFound {
		t.Errorf("expected code %d, got %d", ErrMethodNotFound, resp.Error.Code)
	}
	if !strings.Contains(resp.Error.Message, "unknown/method") {
		t.Errorf("expected method name in error, got %q", resp.Error.Message)
	}
}

func TestServerRun(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "ping", Description: "Ping tool"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{"pong": true}`), nil
		},
	)

	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"ping","arguments":{}}}
`)
	var output bytes.Buffer

	server := NewServer("test", "1.0.0", registry,
		WithIO(input, &output),
		WithLogger(logging.Nop()),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := server.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 responses, got %d: %v", len(lines), lines)
	}

	// Check first response (tools/list)
	var resp1 JSONRPCResponse
	json.Unmarshal([]byte(lines[0]), &resp1)
	if resp1.Error != nil {
		t.Errorf("unexpected error in first response: %v", resp1.Error)
	}

	// Check second response (tools/call)
	var resp2 JSONRPCResponse
	json.Unmarshal([]byte(lines[1]), &resp2)
	if resp2.Error != nil {
		t.Errorf("unexpected error in second response: %v", resp2.Error)
	}
}

func TestServerRunEmptyLines(t *testing.T) {
	registry := NewRegistry()
	input := strings.NewReader(`

{"jsonrpc":"2.0","id":1,"method":"initialize"}

`)
	var output bytes.Buffer

	server := NewServer("test", "1.0.0", registry,
		WithIO(input, &output),
		WithLogger(logging.Nop()),
	)

	err := server.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 response, got %d", len(lines))
	}
}

func TestServerRunInvalidJSON(t *testing.T) {
	registry := NewRegistry()
	input := strings.NewReader(`{not valid json}
{"jsonrpc":"2.0","id":1,"method":"initialize"}
`)
	var output bytes.Buffer

	server := NewServer("test", "1.0.0", registry,
		WithIO(input, &output),
		WithLogger(logging.Nop()),
	)

	err := server.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should skip invalid JSON and process the next valid request
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 response, got %d", len(lines))
	}
}

func TestServerWithIOOption(t *testing.T) {
	registry := NewRegistry()
	var input bytes.Buffer
	var output bytes.Buffer

	server := NewServer("test", "1.0.0", registry,
		WithIO(&input, &output),
	)

	if server.reader != &input {
		t.Error("expected custom reader to be set")
	}
	if server.writer != &output {
		t.Error("expected custom writer to be set")
	}
}

func BenchmarkServerProcessRequest(b *testing.B) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "bench"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{}`), nil
		},
	)

	server := NewServer("test", "1.0.0", registry)

	params, _ := json.Marshal(ToolCallParams{Name: "bench", Arguments: map[string]any{}})
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  params,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.ProcessRequest(ctx, req)
	}
}
