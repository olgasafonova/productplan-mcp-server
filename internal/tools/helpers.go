package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// Validatable is implemented by arg structs that have required field checks.
type Validatable interface {
	Validate() error
}

// typedHandler eliminates the ParseArgs + Validate boilerplate.
// It parses raw args into T, validates, then delegates to fn.
func typedHandler[T Validatable](fn func(ctx context.Context, a T) (json.RawMessage, error)) mcp.Handler {
	return typed[T]{fn: func(ctx context.Context, a T, _ map[string]any) (json.RawMessage, error) {
		return fn(ctx, a)
	}}
}

// typed is the handler typedHandler and queryHandler build. fn also gets
// the raw args, for arguments (list filters) that live only in the schema.
type typed[T Validatable] struct {
	fn func(ctx context.Context, a T, args map[string]any) (json.RawMessage, error)
}

// Handle parses args into T, validates, and runs fn.
func (h typed[T]) Handle(ctx context.Context, args map[string]any) (json.RawMessage, error) {
	a, err := ParseArgs[T](args)
	if err != nil {
		return nil, err
	}
	if err = a.Validate(); err != nil {
		return nil, err
	}
	return h.fn(ctx, a, args)
}

// argNames lists the JSON argument names T reads, including those of
// embedded structs. Tests use it to prove each tool's InputSchema declares
// every argument its handler reads, since unknown keys are rejected at
// dispatch (mcp.Tool.CheckArgumentKeys).
func (typed[T]) argNames() []string {
	return jsonFieldNames(reflect.TypeFor[T]())
}

// jsonFieldNames returns the JSON names of t's exported fields, flattening
// anonymous embedded structs the way encoding/json does.
func jsonFieldNames(t reflect.Type) []string {
	var names []string
	for f := range t.Fields() {
		tag, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		switch {
		case tag == "-" || !f.IsExported():
			continue
		case f.Anonymous && tag == "" && f.Type.Kind() == reflect.Struct:
			names = append(names, jsonFieldNames(f.Type)...)
		case tag == "":
			names = append(names, f.Name)
		default:
			names = append(names, tag)
		}
	}
	return names
}

// firstError runs checks in order and returns the first error, or nil.
func firstError(checks ...func() error) error {
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
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
	resource ItemType
	create   func(ctx context.Context, payload map[string]any) (json.RawMessage, error)
	update   func(ctx context.Context, id string, payload map[string]any) (json.RawMessage, error)
	delete   func(ctx context.Context, id string) (json.RawMessage, error)
}

// run dispatches the requested action to the matching client call and
// formats the result. Unknown actions fall through with empty data.
func (o topLevelOps) run(ctx context.Context, req manageRequest) (json.RawMessage, error) {
	data, err := o.dispatch(ctx, req)
	return formatManaged(data, err, o.resource, req)
}

// dispatch performs the client call for the requested action.
func (o topLevelOps) dispatch(ctx context.Context, req manageRequest) (json.RawMessage, error) {
	switch {
	case req.action == "create":
		return o.create(ctx, req.createPayload)
	case req.action == "update":
		return o.update(ctx, req.id, req.updatePayload)
	case req.action == "delete" && o.delete != nil:
		return o.delete(ctx, req.id)
	}
	return nil, nil
}

// formatManaged turns a manage-style client result into the tool response.
func formatManaged(data json.RawMessage, err error, resource ItemType, req manageRequest) (json.RawMessage, error) {
	if err != nil {
		return nil, err
	}
	return FormatAction(data, req.action, resource, req.id)
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
	resource ItemType
	create   func(ctx context.Context, parentID string, payload map[string]any) (json.RawMessage, error)
	update   func(ctx context.Context, parentID, id string, payload map[string]any) (json.RawMessage, error)
	delete   func(ctx context.Context, parentID, id string) (json.RawMessage, error)
}

// run dispatches the requested action to the matching client call and
// formats the result. Unknown actions fall through with empty data.
func (o parentScopedOps) run(ctx context.Context, req manageRequest) (json.RawMessage, error) {
	data, err := o.dispatch(ctx, req)
	return formatManaged(data, err, o.resource, req)
}

// dispatch performs the client call for the requested action.
func (o parentScopedOps) dispatch(ctx context.Context, req manageRequest) (json.RawMessage, error) {
	switch req.action {
	case "create":
		return o.create(ctx, req.parentID, req.createPayload)
	case "update":
		return o.update(ctx, req.parentID, req.id, req.updatePayload)
	case "delete":
		return o.delete(ctx, req.parentID, req.id)
	}
	return nil, nil
}

// numericID sends a numeric ID as a JSON integer (the documented type) and
// leaves anything else as the string the caller gave.
func numericID(id string) any {
	if n, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64); err == nil {
		return n
	}
	return id
}

// firstString returns the first non-empty string.
func firstString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// firstBool returns the first non-nil bool pointer.
func firstBool(values ...*bool) *bool {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

// derefString returns the pointed-to string, or "" for nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
