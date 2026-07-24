package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// Validatable is implemented by arg structs that have required field checks.
type Validatable interface {
	Validate() error
}

// typedHandler eliminates the ParseArgs + Validate boilerplate.
// It parses raw args into T, validates, then delegates to fn.
func typedHandler[T Validatable](fn func(ctx context.Context, a T) (json.RawMessage, error)) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[T](args)
		if err != nil {
			return nil, err
		}
		if err = a.Validate(); err != nil {
			return nil, err
		}
		return fn(ctx, a)
	})
}

// buildPayload assembles an API payload: entries in always are included
// unconditionally, optional fields only when their value is non-empty.
func buildPayload(always map[string]any, optional ...fieldCheck) map[string]any {
	payload := make(map[string]any, len(always)+len(optional))
	for k, v := range always {
		payload[k] = v
	}
	for _, f := range optional {
		if f.value != "" {
			payload[f.name] = f.value
		}
	}
	return payload
}

// manageRequest carries the per-call inputs for a manage-style handler:
// the requested action, the IDs addressing the resource, and the payloads
// to send for create and update actions.
type manageRequest struct {
	action        string
	parentID      string
	id            string
	createPayload map[string]any
	updatePayload map[string]any
}

// topLevelOps bundles the client calls for a top-level resource
// (ideas, opportunities, objectives) managed through a single
// action-dispatch tool. delete is optional; resources without it treat
// "delete" like any other unsupported action.
type topLevelOps struct {
	resource string
	create   func(ctx context.Context, payload map[string]any) (json.RawMessage, error)
	update   func(ctx context.Context, id string, payload map[string]any) (json.RawMessage, error)
	delete   func(ctx context.Context, id string) (json.RawMessage, error)
}

// run dispatches the requested action to the matching client call and
// formats the result. Unknown actions fall through with empty data.
func (o topLevelOps) run(ctx context.Context, req manageRequest) (json.RawMessage, error) {
	var data json.RawMessage
	var err error
	switch {
	case req.action == "create":
		data, err = o.create(ctx, req.createPayload)
	case req.action == "update":
		data, err = o.update(ctx, req.id, req.updatePayload)
	case req.action == "delete" && o.delete != nil:
		data, err = o.delete(ctx, req.id)
	}
	if err != nil {
		return nil, err
	}
	return FormatAction(data, req.action, o.resource, req.id)
}

// manageOps is implemented by the ops bundles that dispatch a manageRequest
// (topLevelOps, parentScopedOps).
type manageOps interface {
	run(ctx context.Context, req manageRequest) (json.RawMessage, error)
}

// manageHandler wires a manage-style tool: args are parsed and validated by
// typedHandler, translated into a manageRequest by build, then dispatched
// through ops.
func manageHandler[T Validatable](ops manageOps, build func(a T) manageRequest) mcp.Handler {
	return typedHandler[T](func(ctx context.Context, a T) (json.RawMessage, error) {
		return ops.run(ctx, build(a))
	})
}

// parentScopedOps bundles the client calls for a resource nested under a
// parent (lanes and milestones under a roadmap) managed through a single
// action-dispatch tool.
type parentScopedOps struct {
	resource string
	create   func(ctx context.Context, parentID string, payload map[string]any) (json.RawMessage, error)
	update   func(ctx context.Context, parentID, id string, payload map[string]any) (json.RawMessage, error)
	delete   func(ctx context.Context, parentID, id string) (json.RawMessage, error)
}

// run dispatches the requested action to the matching client call and
// formats the result. Unknown actions fall through with empty data.
func (o parentScopedOps) run(ctx context.Context, req manageRequest) (json.RawMessage, error) {
	var data json.RawMessage
	var err error
	switch req.action {
	case "create":
		data, err = o.create(ctx, req.parentID, req.createPayload)
	case "update":
		data, err = o.update(ctx, req.parentID, req.id, req.updatePayload)
	case "delete":
		data, err = o.delete(ctx, req.parentID, req.id)
	}
	if err != nil {
		return nil, err
	}
	return FormatAction(data, req.action, o.resource, req.id)
}
