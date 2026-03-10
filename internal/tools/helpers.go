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
