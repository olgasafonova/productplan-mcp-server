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
		Tool{Name: "echo", Description: "Echo tool"},
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

func TestArgHelper(t *testing.T) {
	args := map[string]any{
		"name":    "test",
		"count":   float64(42),
		"enabled": true,
		"empty":   "",
	}
	h := NewArgHelper(args)

	t.Run("String", func(t *testing.T) {
		if s := h.String("name"); s != "test" {
			t.Errorf("expected 'test', got %q", s)
		}
		if s := h.String("missing"); s != "" {
			t.Errorf("expected empty string, got %q", s)
		}
	})

	t.Run("Int", func(t *testing.T) {
		if i := h.Int("count"); i != 42 {
			t.Errorf("expected 42, got %d", i)
		}
		if i := h.Int("missing"); i != 0 {
			t.Errorf("expected 0, got %d", i)
		}
	})

	t.Run("Bool", func(t *testing.T) {
		if b := h.Bool("enabled"); !b {
			t.Error("expected true")
		}
		if b := h.Bool("missing"); b {
			t.Error("expected false")
		}
	})

	t.Run("Has", func(t *testing.T) {
		if !h.Has("name") {
			t.Error("expected Has('name') to be true")
		}
		if h.Has("empty") {
			t.Error("expected Has('empty') to be false for empty string")
		}
		if h.Has("missing") {
			t.Error("expected Has('missing') to be false")
		}
		if !h.Has("enabled") {
			t.Error("expected Has('enabled') to be true for bool")
		}
	})

	t.Run("BuildData", func(t *testing.T) {
		data := h.BuildData("name", "empty", "missing")
		if data["name"] != "test" {
			t.Errorf("expected 'test', got %v", data["name"])
		}
		if _, ok := data["empty"]; ok {
			t.Error("expected empty string to be excluded")
		}
		if _, ok := data["missing"]; ok {
			t.Error("expected missing key to be excluded")
		}
	})

	t.Run("RequiredString", func(t *testing.T) {
		s, err := h.RequiredString("name")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if s != "test" {
			t.Errorf("expected 'test', got %q", s)
		}

		_, err = h.RequiredString("missing")
		if err == nil {
			t.Error("expected error for missing required string")
		}

		_, err = h.RequiredString("empty")
		if err == nil {
			t.Error("expected error for empty required string")
		}
	})
}

func TestArgHelperIntTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int
	}{
		{"int", int(10), 10},
		{"float64", float64(20), 20},
		{"int64", int64(30), 30},
		{"string", "40", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewArgHelper(map[string]any{"value": tt.value})
			if got := h.Int("value"); got != tt.want {
				t.Errorf("expected %d, got %d", tt.want, got)
			}
		})
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

func BenchmarkArgHelper(b *testing.B) {
	args := map[string]any{
		"id":      "123",
		"count":   float64(42),
		"enabled": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := NewArgHelper(args)
		h.String("id")
		h.Int("count")
		h.Bool("enabled")
	}
}
