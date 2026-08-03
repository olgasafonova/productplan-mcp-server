package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

// SDKServer serves the tool registry over the official MCP SDK.
//
// Slice 2 of the go-sdk migration (bead claude-code-config-4fc4.14). This
// replaces the hand-rolled JSON-RPC loop in server.go: the SDK now owns version
// negotiation, server/discover, the initialize handshake, cache hints and every
// future revision, which is the whole point of the migration. What stays local
// is tool authoring (the Tool type plus BuildTool) and tool dispatch (Registry).
//
// The old Server is left in place by this slice and removed in the next one,
// together with the hand-rolled protocol types and the tests written against
// them. Keeping both compiling for one slice keeps this diff to the transport.
type SDKServer struct {
	registry     *Registry
	logger       logging.Logger
	instructions string
	server       *mcp.Server
}

// SDKServerOption configures an SDKServer. Mirrors the option set the
// hand-rolled Server exposes, minus WithIO: the SDK owns the transport, and
// tests drive the server through an in-memory transport instead.
type SDKServerOption func(*SDKServer)

// WithSDKLogger sets the server logger.
func WithSDKLogger(logger logging.Logger) SDKServerOption {
	return func(s *SDKServer) {
		s.logger = logger
	}
}

// WithSDKInstructions sets the server instructions returned at initialize.
func WithSDKInstructions(instructions string) SDKServerOption {
	return func(s *SDKServer) {
		s.instructions = instructions
	}
}

// NewSDKServer builds an SDK-backed server over the given registry and
// registers every tool in it.
func NewSDKServer(name, version string, registry *Registry, opts ...SDKServerOption) *SDKServer {
	s := &SDKServer{
		registry: registry,
		logger:   logging.Nop(),
	}
	for _, opt := range opts {
		opt(s)
	}

	s.server = mcp.NewServer(
		&mcp.Implementation{Name: name, Version: version},
		&mcp.ServerOptions{Instructions: s.instructions},
	)

	for _, tool := range registry.Tools() {
		s.server.AddTool(BuildTool(tool), s.handlerFor(tool.Name))
	}

	return s
}

// Run serves over stdio until the client disconnects or ctx is cancelled.
func (s *SDKServer) Run(ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "ProductPlan MCP Server running on stdio\n")
	return s.run(ctx, &mcp.StdioTransport{})
}

// run serves over an arbitrary transport. This is the seam the old server
// exposed as WithIO: tests drive the server over an in-memory transport so they
// exercise a real session rather than calling handlers directly. It stays
// unexported because stdio is the only transport this binary ships.
func (s *SDKServer) run(ctx context.Context, transport mcp.Transport) error {
	s.logger.Info("MCP server starting",
		logging.F("tools", s.registry.Count()),
	)
	return s.server.Run(ctx, transport)
}

// handlerFor adapts one registry tool to the SDK's handler signature.
//
// Dispatch deliberately goes through Registry.Call rather than reaching for the
// Handler directly. Call is where the panic recovery lives (handler.go), so
// routing through it keeps HG-1 intact by construction: a panicking tool becomes
// a structured error result rather than a silent fake success. Re-implementing
// recovery in this adapter would mean two places to keep correct.
//
// Tool failures are reported as an error RESULT with IsError set, never as a
// returned Go error. Returning an error here would surface as a JSON-RPC
// protocol error, which tells the caller the request was malformed rather than
// that the tool ran and failed. This mirrors what handleToolCall does today.
func (s *SDKServer) handlerFor(name string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, err := decodeArguments(req)
		if err != nil {
			return errorResult(err), nil
		}

		s.logger.Debug("calling tool", logging.Tool(name))

		result, err := s.registry.Call(ctx, name, args)
		if err != nil {
			s.logger.Debug("tool call failed",
				logging.Tool(name),
				logging.Error(err),
			)
			return errorResult(err), nil
		}

		return toolResult(s.registry, name, result), nil
	}
}

// decodeArguments turns the request's raw arguments into the map the local
// Handler contract expects. Absent arguments are an empty map, not nil, so
// handlers can index without a nil check.
func decodeArguments(req *mcp.CallToolRequest) (map[string]any, error) {
	args := map[string]any{}
	if req == nil || req.Params == nil || len(req.Params.Arguments) == 0 {
		return args, nil
	}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}
	return args, nil
}

// toolResult shapes a successful handler payload, preserving the existing rule
// that structuredContent is emitted only for tools that declare an OutputSchema
// and only when the payload is valid JSON.
func toolResult(registry *Registry, name string, payload json.RawMessage) *mcp.CallToolResult {
	res := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(payload)}},
	}
	if registry.HasOutputSchema(name) && json.Valid(payload) {
		res.StructuredContent = payload
	}
	return res
}

// errorResult keeps the "Error: " prefix the hand-rolled server used, so a
// client sees the same text before and after the migration.
func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Error: " + err.Error()}},
		IsError: true,
	}
}
