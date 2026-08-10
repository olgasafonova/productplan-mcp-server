package tools

import (
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// ideaTools returns idea and discovery tool definitions.
func ideaTools() []mcp.Tool {
	tools := ideaReadTools()
	return append(tools, ideaManageTools()...)
}

// ideaReadTools returns read-only idea/discovery tool definitions.
func ideaReadTools() []mcp.Tool {
	tools := ideaCoreReadTools()
	return append(tools, opportunityReadTools()...)
}

// ideaCoreReadTools returns read tools for ideas themselves.
func ideaCoreReadTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "list_ideas",
			Description: `List all ideas in discovery pipeline. START HERE for ideas.

USE WHEN: "Show customer feedback", "What ideas do we have?"
Returns array of ideas with ID, title, status, and vote count.
FAILS WHEN: API token invalid. Returns empty list if no ideas exist.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_idea",
			Description: `Get idea details including description and metadata.

USE WHEN: "Tell me about this idea", "Full request details"
FAILS WHEN: idea_id not found (get valid IDs from list_ideas).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "Idea ID from list_ideas"},
				},
				Required: []string{"idea_id"},
			},
		},
	}
}

// opportunityReadTools returns read tools for opportunities and discovery metadata.
func opportunityReadTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "list_opportunities",
			Description: `List all opportunities. START HERE for discovery.

USE WHEN: "Show opportunities", "Discovery pipeline"
Returns array of opportunities with ID, problem_statement, workflow_status, and linked idea count.
FAILS WHEN: API token invalid. Returns empty list if no opportunities exist.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_opportunity",
			Description: `Get opportunity details with linked ideas.

USE WHEN: "Tell me about this opportunity"
FAILS WHEN: opportunity_id not found (get valid IDs from list_opportunities).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"opportunity_id": {Type: "string", Description: "Opportunity ID from list_opportunities"},
				},
				Required: []string{"opportunity_id"},
			},
		},
		{
			Name: "list_idea_forms",
			Description: `List idea submission forms.

USE WHEN: "Show feedback forms", "What forms exist?"
Returns array of forms with ID and name.
FAILS WHEN: API token invalid.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_idea_form",
			Description: `Get idea form details with fields.

USE WHEN: "Show form fields", "What does this form collect?"
FAILS WHEN: form_id not found (get valid IDs from list_idea_forms).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"form_id": {Type: "string", Description: "Form ID from list_idea_forms"},
				},
				Required: []string{"form_id"},
			},
		},
		{
			Name: "list_all_customers",
			Description: `List all customers across ideas. Returns customer names and linked idea counts.

USE WHEN: "Who are our customers?", "All feedback sources"
FAILS WHEN: API token invalid.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "list_all_tags",
			Description: `List all tags used across ideas. Returns tag names and usage counts across ideas.

USE WHEN: "What tags exist?", "Show categories"
FAILS WHEN: API token invalid.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
	}
}

// ideaManageTools returns idea/opportunity mutation tool definitions.
func ideaManageTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "manage_idea",
			Description: `Create or update an idea. Note: delete not available via API.

USE WHEN: "Add idea", "Update idea status"
Actions: create (title), update (idea_id)
Returns the created/updated idea object.
FAILS WHEN: create without title, update without idea_id. Note: delete is not available via the ProductPlan API; archive ideas by updating status instead.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "create or update", Enum: []string{"create", "update"}},
					"idea_id":     {Type: "string", Description: "Idea ID (for update)"},
					"title":       {Type: "string", Description: "Idea title"},
					"description": {Type: "string", Description: "Description (markdown)"},
					"status":      {Type: "string", Description: "Idea workflow status", Enum: []string{"new", "under_review", "planned"}},
				},
				Required: []string{"action"},
			},
		},
		{
			Name: "manage_opportunity",
			Description: `Create or update an opportunity. Note: delete not available via API.

USE WHEN: "Create opportunity", "Update problem"
Actions: create (problem_statement), update (opportunity_id)
Returns the created/updated opportunity object.
FAILS WHEN: create without problem_statement, update without opportunity_id (get IDs from list_opportunities). Note: delete is not available via the ProductPlan API; archive opportunities by updating workflow_status instead.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":            {Type: "string", Description: "create or update", Enum: []string{"create", "update"}},
					"opportunity_id":    {Type: "string", Description: "Opportunity ID (for update)"},
					"problem_statement": {Type: "string", Description: "Problem statement (title)"},
					"description":       {Type: "string", Description: "Description"},
					"workflow_status":   {Type: "string", Description: "Opportunity workflow status", Enum: []string{"draft", "in_discovery", "validated", "invalidated", "completed"}},
				},
				Required: []string{"action"},
			},
		},
	}
}
