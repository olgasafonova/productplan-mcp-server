package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BuildTool converts a locally-declared Tool into the official SDK's tool type.
//
// This is the first step of the go-sdk migration (bead
// claude-code-config-4fc4.14): the local Tool stays the authoring format, and
// this function is the single point where it becomes protocol. That split is
// the pattern the rest of the portfolio already uses, where nordic-registry,
// mediawiki and miro each keep a local spec table and convert at registration
// time, and it is what keeps internal/tools' 47 definitions free of any SDK
// import.
//
// Note this package is itself named mcp and imports the SDK's mcp package
// without an alias. That is legal because a package never refers to itself by
// name, so throughout this file a bare Tool is the local type and mcp.Tool is
// the SDK's. public360's internal/mcp/sdk_server.go does the same.
//
// Schemas pass through untouched. The SDK declares InputSchema and
// OutputSchema as `any` and marshals whatever it is given, so the existing
// InputSchema and OutputSchema structs produce byte-identical JSON to what the
// hand-rolled server emits today. Converting them to jsonschema-go types would
// change the wire output for no gain.
func BuildTool(tool Tool) *mcp.Tool {
	built := &mcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: tool.InputSchema,
	}
	if tool.OutputSchema != nil {
		built.OutputSchema = tool.OutputSchema
	}
	built.Annotations = buildAnnotations(tool.Annotations)
	return built
}

// buildAnnotations converts the local annotation hints, or returns nil when a
// tool declares none.
//
// The title deliberately stays on the annotations rather than being promoted to
// the SDK Tool's own Title field. The SDK resolves display names in the order
// title, annotations.title, then name, so populating both would be redundant,
// and populating Title alone would move the value to a different JSON key than
// the one this server emits today.
//
// Callers should expect readOnlyHint and idempotentHint to start appearing on
// every converted tool, including those that set neither. Both are bare bools
// with no omitempty in go-sdk v1.7.0, whereas the local ToolAnnotations marks
// both omitempty. That is a real wire change and it is the protocol's, not a
// defect here; destructiveHint and openWorldHint are pointers in both types and
// keep omitting when unset.
func buildAnnotations(ann *ToolAnnotations) *mcp.ToolAnnotations {
	if ann == nil {
		return nil
	}
	return &mcp.ToolAnnotations{
		Title:           ann.Title,
		ReadOnlyHint:    ann.ReadOnlyHint,
		IdempotentHint:  ann.IdempotentHint,
		DestructiveHint: ann.DestructiveHint,
		OpenWorldHint:   ann.OpenWorldHint,
	}
}
