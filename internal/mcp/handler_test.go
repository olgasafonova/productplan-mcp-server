package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestHandlerFunc(t *testing.T) {
	fn := HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return json.RawMessage(`{"result": "ok"}`), nil
	})

	result, err := fn.Handle(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `{"result": "ok"}` {
		t.Errorf("unexpected result: %s", string(result))
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: InputSchema{Type: "object"},
	}

	handler := HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return json.RawMessage(`{"status": "ok"}`), nil
	})

	r.Register(tool, handler)

	if r.Count() != 1 {
		t.Errorf("expected 1 tool, got %d", r.Count())
	}

	tools := r.Tools()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "test_tool" {
		t.Errorf("expected tool name 'test_tool', got %q", tools[0].Name)
	}

	h, ok := r.Handler("test_tool")
	if !ok {
		t.Fatal("expected to find handler")
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}

	_, ok = r.Handler("nonexistent")
	if ok {
		t.Error("expected not to find handler")
	}
}

func TestRegistryRegisterFunc(t *testing.T) {
	r := NewRegistry()

	tool := Tool{Name: "func_tool", Description: "Function tool"}
	r.RegisterFunc(tool, func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return json.RawMessage(`{"type": "func"}`), nil
	})

	if r.Count() != 1 {
		t.Errorf("expected 1 tool, got %d", r.Count())
	}
}

func TestRegistryCall(t *testing.T) {
	r := NewRegistry()

	r.RegisterFunc(
		Tool{Name: "echo", Description: "Echo tool", InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]Property{"message": {Type: "string"}},
		}},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			msg := args["message"].(string)
			return json.RawMessage(`{"echo": "` + msg + `"}`), nil
		},
	)

	result, err := r.Call(context.Background(), "echo", map[string]any{"message": "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]string
	json.Unmarshal(result, &parsed)
	if parsed["echo"] != "hello" {
		t.Errorf("expected echo 'hello', got %q", parsed["echo"])
	}
}

func TestRegistryCallUnknown(t *testing.T) {
	r := NewRegistry()

	_, err := r.Call(context.Background(), "unknown_tool", nil)
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestRegistryCallError(t *testing.T) {
	r := NewRegistry()

	r.RegisterFunc(
		Tool{Name: "failing", Description: "Failing tool"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return nil, errors.New("intentional failure")
		},
	)

	_, err := r.Call(context.Background(), "failing", nil)
	if err == nil {
		t.Error("expected error")
	}
	if err.Error() != "intentional failure" {
		t.Errorf("expected 'intentional failure', got %q", err.Error())
	}
}

func TestRegistryToolsCopy(t *testing.T) {
	r := NewRegistry()
	r.RegisterFunc(Tool{Name: "tool1"}, func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return nil, nil
	})

	tools := r.Tools()
	_ = append(tools, Tool{Name: "tool2"})

	if r.Count() != 1 {
		t.Error("modifying returned slice should not affect registry")
	}
}

func BenchmarkRegistryCall(b *testing.B) {
	r := NewRegistry()
	r.RegisterFunc(
		Tool{Name: "bench_tool"},
		func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{}`), nil
		},
	)

	ctx := context.Background()
	args := map[string]any{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Call(ctx, "bench_tool", args)
	}
}
