// Package mcp defines this server's tool-authoring types and its tool registry,
// and serves them over the official MCP SDK.
//
// The JSON-RPC framing, the initialize handshake, protocol-version negotiation
// and every wire type used to live here as a hand-rolled implementation pinned
// to protocol revision 2025-11-25. Slice 2 of bead claude-code-config-4fc4.14
// moved all of that to github.com/modelcontextprotocol/go-sdk, and slice 3
// deleted the hand-rolled half. What remains is deliberately only the part the
// SDK does not own:
//
//   - the Tool authoring format below, converted at registration by BuildTool
//     (sdk.go)
//   - the Registry and Handler contract (handler.go), which is where tool
//     dispatch and its panic recovery live
//
// Keeping the authoring format local is the pattern the rest of the portfolio
// uses; it is what lets internal/tools declare 47 tools without importing the
// SDK.
package mcp

// ToolAnnotations provides optional hints about a tool's behavior.
//
// The SDK's equivalent drops omitempty on ReadOnlyHint and IdempotentHint, so
// both ship on every annotated tool once converted. See BuildTool.
type ToolAnnotations struct {
	Title           string `json:"title,omitempty"`
	ReadOnlyHint    bool   `json:"readOnlyHint,omitempty"`
	IdempotentHint  bool   `json:"idempotentHint,omitempty"`
	DestructiveHint *bool  `json:"destructiveHint,omitempty"`
	OpenWorldHint   *bool  `json:"openWorldHint,omitempty"`
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
	// OutputSchema is the optional JSON Schema describing the tool's
	// structured result. When set, the server returns the handler result
	// both as serialized JSON text (for backwards compatibility) and as a
	// structuredContent object that conforms to this schema. This makes the
	// tool eligible for Code Mode, where clients drive tools programmatically
	// against the declared output shape (MCP spec 2025-06-18).
	OutputSchema *OutputSchema    `json:"outputSchema,omitempty"`
	Annotations  *ToolAnnotations `json:"annotations,omitempty"`
}

// InputSchema defines the JSON Schema for tool inputs.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// OutputSchema defines the JSON Schema for a tool's structured result.
// It mirrors InputSchema but is kept as a distinct type so output-only
// fields can diverge from input-only fields without coupling the two.
type OutputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property defines a single property in the input schema.
type Property struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Enum        []string  `json:"enum,omitempty"`
	Minimum     *float64  `json:"minimum,omitempty"`
	Maximum     *float64  `json:"maximum,omitempty"`
	Pattern     string    `json:"pattern,omitempty"`
	Items       *Property `json:"items,omitempty"`
	Examples    []any     `json:"examples,omitempty"`
}
