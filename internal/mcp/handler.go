package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
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
	mu      sync.RWMutex
	tools   []Tool
	entries map[string]registration
}

// registration is one tool's definition and handler, looked up by name.
type registration struct {
	tool    Tool
	handler Handler
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		entries: make(map[string]registration),
	}
}

// Register adds a tool with its handler.
func (r *Registry) Register(tool Tool, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = append(r.tools, tool)
	r.entries[tool.Name] = registration{tool: tool, handler: handler}
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

// lookup returns the registration for a tool by name.
func (r *Registry) lookup(name string) (registration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[name]
	return e, ok
}

// Handler returns the handler for a tool by name.
func (r *Registry) Handler(name string) (Handler, bool) {
	e, ok := r.lookup(name)
	return e.handler, ok
}

// HasOutputSchema reports whether the named tool declares an OutputSchema.
// The server uses this to decide whether to emit structuredContent.
func (r *Registry) HasOutputSchema(name string) bool {
	e, ok := r.lookup(name)
	return ok && e.tool.OutputSchema != nil
}

// Call executes a tool by name with the given arguments.
func (r *Registry) Call(ctx context.Context, name string, args map[string]any) (result json.RawMessage, err error) {
	e, ok := r.lookup(name)
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("Panic recovered in tool handler",
				"tool", name,
				"panic", rec,
				"stack", string(debug.Stack()))
			result = nil
			err = fmt.Errorf("internal error in %s", name)
		}
	}()

	return e.handler.Handle(ctx, args)
}

// Count returns the number of registered tools.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}
