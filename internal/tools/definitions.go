package tools

import (
	"strings"

	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// floatPtr returns a pointer to a float64 value.
func floatPtr(v float64) *float64 { return &v }

// boolPtr returns a pointer to a bool value.
func boolPtr(v bool) *bool { return &v }

// readOutputSchema describes the structured result every read-only tool
// returns: the FormattedResponse shape produced by FormatList / FormatItem /
// FormatAction (see internal/tools/formatter.go). Declaring it makes read
// tools Code Mode eligible — clients can drive them against a typed output
// shape instead of parsing free-form text.
//
// `data` is intentionally untyped (an object or array): the underlying
// ProductPlan payloads vary per endpoint and pass through verbatim, so the
// schema asserts the wrapper, not the per-endpoint body.
func readOutputSchema() *mcp.OutputSchema {
	return &mcp.OutputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"summary": {Type: "string", Description: "Human-readable summary of the result (e.g. \"Found 3 roadmaps\")"},
			"data":    {Type: "object", Description: "The ProductPlan payload (object or array) returned by the API, passed through verbatim"},
		},
		Required: []string{"summary", "data"},
	}
}

// BuildAllTools returns all ProductPlan tool definitions for MCP.
func BuildAllTools() []mcp.Tool {
	var tools []mcp.Tool

	// Roadmaps
	tools = append(tools, roadmapTools()...)

	// Bars
	tools = append(tools, barTools()...)

	// Objectives
	tools = append(tools, objectiveTools()...)

	// Ideas
	tools = append(tools, ideaTools()...)

	// Launches
	tools = append(tools, launchTools()...)

	// Utility
	tools = append(tools, utilityTools()...)

	for i := range tools {
		annotateTool(&tools[i])
	}

	return tools
}

// readOnlyToolName reports whether a tool name denotes a read-only tool.
func readOnlyToolName(name string) bool {
	return strings.HasPrefix(name, "get_") ||
		strings.HasPrefix(name, "list_") ||
		strings.HasPrefix(name, "check_") ||
		name == "health_check"
}

// annotateTool fills default annotations based on the tool name prefix,
// leaving tools with explicit annotations untouched.
//
// Read-only (get_*, list_*, check_*, health_check):
//
//	ReadOnlyHint=true. IdempotentHint is NOT set: the hint carries
//	meaning only for tools that modify state, so claiming idempotence
//	on a trivially repeatable read says nothing and misleads a client
//	reasoning about retry safety.
//
// manage_*:
//
//	DestructiveHint=true. Each manage_* tool dispatches across
//	action=create/update/delete and supports cascade-delete with
//	documented blast radius (e.g., manage_launch removes its sections
//	and tasks, manage_objective cascades to all key results). MCP
//	clients use DestructiveHint to gate user-confirmation prompts on
//	irreversible operations.
//
//	IdempotentHint is deliberately NOT set on manage_* tools.
//	Idempotency varies per action: action=create twice produces two
//	records; action=delete twice 404s on the second call. A blanket
//	"idempotent" annotation misleads retry-aware clients into
//	duplicate writes. Leaving the hint unset is the safe default.
func annotateTool(tool *mcp.Tool) {
	if tool.Annotations != nil {
		return
	}
	switch {
	case readOnlyToolName(tool.Name):
		tool.Annotations = &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		}
		// Read tools return the uniform FormattedResponse wrapper, so
		// they declare an OutputSchema and become Code Mode eligible.
		// Only set it when unset, mirroring the Annotations guard above.
		if tool.OutputSchema == nil {
			tool.OutputSchema = readOutputSchema()
		}
	case strings.HasPrefix(tool.Name, "manage_"):
		tool.Annotations = &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(true),
		}
	}
}

// utilityTools returns utility tool definitions.
func utilityTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "check_status",
			Description: `Check ProductPlan API status and authentication.

USE WHEN: "Is ProductPlan connected?", "Check API"
For MCP server internals (cache stats, rate limits), use health_check instead.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "health_check",
			Description: `Check MCP server health and cache stats.

USE WHEN: "Server status", "Rate limits", "Diagnose issues"
For API connectivity only, use check_status instead.
FAILS WHEN: deep=true and API is unreachable. Basic health (deep=false) always succeeds if server is running.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"deep": {Type: "boolean", Description: "Also verify API connectivity (~500ms)"},
				},
			},
		},
		{
			Name: "list_users",
			Description: `List all users in account.

USE WHEN: "Who has access?", "Team members"
Returns array of users with ID, name, email, and role. Use user IDs from this tool when assigning launch tasks via manage_launch_task.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "list_teams",
			Description: `List all teams in account.

USE WHEN: "What teams exist?", "Team structure"
For individual user details, use list_users instead.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
	}
}
