package tools

import (
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// objectiveTools returns OKR-related tool definitions.
func objectiveTools() []mcp.Tool {
	tools := objectiveReadTools()
	return append(tools, objectiveManageTools()...)
}

// objectiveReadTools returns read-only OKR tool definitions.
func objectiveReadTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "list_objectives",
			Description: `List all OKR objectives. START HERE for OKRs.

USE WHEN: "Show OKRs", "What are our objectives?"
Returns array of objectives with ID, name, time_frame, and key result count.
FAILS WHEN: API token invalid. Returns empty list if no objectives exist.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_objective",
			Description: `Get objective details with key results.

USE WHEN: "Tell me about objective X", "OKR progress"
FAILS WHEN: objective_id not found (get valid IDs from list_objectives).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"objective_id": {Type: "string", Description: "Objective ID from list_objectives"},
				},
				Required: []string{"objective_id"},
			},
		},
		{
			Name: "list_key_results",
			Description: `List key results for an objective.

USE WHEN: "What are the KRs?", "Show metrics"
Returns array of key results with ID, name, target_value, and current_value.
FAILS WHEN: objective_id not found (use list_objectives).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"objective_id": {Type: "string", Description: "Objective ID"},
				},
				Required: []string{"objective_id"},
			},
		},
		{
			Name: "get_key_result",
			Description: `Get key result details.

USE WHEN: "Tell me about this KR", "KR progress"
FAILS WHEN: objective_id or key_result_id not found (use list_key_results to get valid KR IDs).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"objective_id":  {Type: "string", Description: "Parent objective ID"},
					"key_result_id": {Type: "string", Description: "Key result ID"},
				},
				Required: []string{"objective_id", "key_result_id"},
			},
		},
	}
}

// objectiveManageTools returns OKR mutation tool definitions.
func objectiveManageTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "manage_objective",
			Description: `Create, update, or delete an objective.

USE WHEN: "Add Q1 objective", "Update objective", "Delete OKR"
Actions: create (name), update (objective_id), delete (objective_id)
Returns the created/updated objective, or confirmation on delete.
FAILS WHEN: create without name, update/delete without objective_id. WARNING: delete also removes all key results under this objective.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":       {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"objective_id": {Type: "string", Description: "Objective ID (for update/delete)"},
					"name":         {Type: "string", Description: "Objective name"},
					"description":  {Type: "string", Description: "Description"},
					"time_frame":   {Type: "string", Description: "Q1 2024, H1 2024, 2024", Examples: []any{"Q1 2025", "H2 2025", "2025"}},
				},
				Required: []string{"action"},
			},
		},
		{
			Name: "manage_key_result",
			Description: `Create, update, or delete a key result.

USE WHEN: "Add KR", "Update progress", "Delete KR"
Actions: create (name+target), update (key_result_id), delete (key_result_id)
Returns the created/updated key result, or confirmation on delete.
FAILS WHEN: create without name or target_value, update/delete without key_result_id (use list_key_results).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":        {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"objective_id":  {Type: "string", Description: "Parent objective ID"},
					"key_result_id": {Type: "string", Description: "Key result ID (for update/delete)"},
					"name":          {Type: "string", Description: "Key result name"},
					"target_value":  {Type: "string", Description: "Target (100, 50%, $1M)"},
					"current_value": {Type: "string", Description: "Current progress"},
				},
				Required: []string{"action", "objective_id"},
			},
		},
	}
}
