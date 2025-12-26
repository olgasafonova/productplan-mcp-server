package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Handler handles a single tool call.
type Handler interface {
	Handle(ctx context.Context, args map[string]any) (json.RawMessage, error)
}

// HandlerFunc is an adapter to allow functions as Handlers.
type HandlerFunc func(ctx context.Context, args map[string]any) (json.RawMessage, error)

// Handle implements Handler.
func (f HandlerFunc) Handle(ctx context.Context, args map[string]any) (json.RawMessage, error) {
	return f(ctx, args)
}

// Registry manages tool definitions and handlers.
type Registry struct {
	mu       sync.RWMutex
	tools    []Tool
	handlers map[string]Handler
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]Handler),
	}
}

// Register adds a tool with its handler.
func (r *Registry) Register(tool Tool, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = append(r.tools, tool)
	r.handlers[tool.Name] = handler
}

// RegisterFunc adds a tool with a function handler.
func (r *Registry) RegisterFunc(tool Tool, fn func(ctx context.Context, args map[string]any) (json.RawMessage, error)) {
	r.Register(tool, HandlerFunc(fn))
}

// Tools returns all registered tools.
func (r *Registry) Tools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Tool, len(r.tools))
	copy(result, r.tools)
	return result
}

// Handler returns the handler for a tool by name.
func (r *Registry) Handler(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[name]
	return h, ok
}

// Call executes a tool by name with the given arguments.
func (r *Registry) Call(ctx context.Context, name string, args map[string]any) (json.RawMessage, error) {
	handler, ok := r.Handler(name)
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return handler.Handle(ctx, args)
}

// Count returns the number of registered tools.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// ArgHelper provides helper methods for extracting typed values from arguments.
type ArgHelper struct {
	args map[string]any
}

// NewArgHelper creates a new argument helper.
func NewArgHelper(args map[string]any) *ArgHelper {
	return &ArgHelper{args: args}
}

// String returns the string value for a key, or empty string if not found.
func (h *ArgHelper) String(key string) string {
	if v, ok := h.args[key].(string); ok {
		return v
	}
	return ""
}

// Int returns the int value for a key, or 0 if not found.
func (h *ArgHelper) Int(key string) int {
	switch v := h.args[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	}
	return 0
}

// Bool returns the bool value for a key, or false if not found.
func (h *ArgHelper) Bool(key string) bool {
	if v, ok := h.args[key].(bool); ok {
		return v
	}
	return false
}

// Has returns true if the key exists and has a non-empty value.
func (h *ArgHelper) Has(key string) bool {
	v, ok := h.args[key]
	if !ok {
		return false
	}
	if s, ok := v.(string); ok {
		return s != ""
	}
	return true
}

// BuildData creates a map from key-value pairs, excluding empty strings.
func (h *ArgHelper) BuildData(keys ...string) map[string]any {
	data := make(map[string]any)
	for _, key := range keys {
		if v := h.String(key); v != "" {
			data[key] = v
		}
	}
	return data
}

// RequiredString returns the string value for a key, or an error if not found.
func (h *ArgHelper) RequiredString(key string) (string, error) {
	if v := h.String(key); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("required parameter missing: %s", key)
}
