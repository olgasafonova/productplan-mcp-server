package tools

import (
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

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

	return tools
}

// roadmapTools returns roadmap-related tool definitions.
func roadmapTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "list_roadmaps",
			Description: "List all roadmaps. Call this FIRST to get roadmap IDs before querying bars or lanes. No parameters needed.",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name:        "get_roadmap",
			Description: "Get full details of a specific roadmap including settings and metadata.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID (get from list_roadmaps)"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name:        "get_roadmap_bars",
			Description: "Get all bars (items) in a roadmap. Returns bars with their lane names for context. Use this to see what's planned on a roadmap.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name:        "get_roadmap_lanes",
			Description: "Get all lanes (swim lanes/rows) in a roadmap. Lanes organize bars into categories.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name:        "get_roadmap_milestones",
			Description: "Get all milestones (key dates) in a roadmap.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name:        "manage_lane",
			Description: "Create, update, or delete a lane on a roadmap.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":     {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"roadmap_id": {Type: "string", Description: "Roadmap ID (required for all actions)"},
					"lane_id":    {Type: "string", Description: "Lane ID (required for update/delete)"},
					"name":       {Type: "string", Description: "Lane name"},
					"color":      {Type: "string", Description: "Lane color hex code"},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
		{
			Name:        "manage_milestone",
			Description: "Create, update, or delete a milestone on a roadmap.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":       {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"roadmap_id":   {Type: "string", Description: "Roadmap ID (required for all actions)"},
					"milestone_id": {Type: "string", Description: "Milestone ID (required for update/delete)"},
					"name":         {Type: "string", Description: "Milestone name"},
					"date":         {Type: "string", Description: "Date YYYY-MM-DD"},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
	}
}

// barTools returns bar-related tool definitions.
func barTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "get_bar",
			Description: "Get full details of a specific bar including description, links, and custom fields.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name:        "get_bar_children",
			Description: "Get child bars (sub-items) of a specific bar. Returns nested items under a parent bar.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The parent bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name:        "get_bar_comments",
			Description: "Get all comments on a specific bar. Shows discussion and feedback on roadmap items.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name:        "get_bar_connections",
			Description: "Get connections (dependencies) between bars. Shows what bars are linked together.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name:        "get_bar_links",
			Description: "Get external links attached to a bar (URLs to Jira, docs, designs, etc).",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name:        "manage_bar",
			Description: "Create, update, or delete a bar on a roadmap.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"bar_id":      {Type: "string", Description: "Bar ID (required for update/delete)"},
					"roadmap_id":  {Type: "string", Description: "Roadmap ID (required for create)"},
					"lane_id":     {Type: "string", Description: "Lane ID (required for create)"},
					"name":        {Type: "string", Description: "Bar name"},
					"start_date":  {Type: "string", Description: "Start date YYYY-MM-DD"},
					"end_date":    {Type: "string", Description: "End date YYYY-MM-DD"},
					"description": {Type: "string", Description: "Description text"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name:        "manage_bar_comment",
			Description: "Add a comment to a bar.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The bar ID to comment on"},
					"body":   {Type: "string", Description: "Comment text content"},
				},
				Required: []string{"bar_id", "body"},
			},
		},
		{
			Name:        "manage_bar_connection",
			Description: "Create or delete a connection (dependency) between bars.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":        {Type: "string", Description: "Action to perform", Enum: []string{"create", "delete"}},
					"bar_id":        {Type: "string", Description: "Source bar ID"},
					"target_bar_id": {Type: "string", Description: "Target bar ID to connect to (for create)"},
					"connection_id": {Type: "string", Description: "Connection ID (required for delete)"},
				},
				Required: []string{"action", "bar_id"},
			},
		},
		{
			Name:        "manage_bar_link",
			Description: "Create, update, or delete an external link on a bar.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":  {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"bar_id":  {Type: "string", Description: "The bar ID"},
					"link_id": {Type: "string", Description: "Link ID (required for update/delete)"},
					"url":     {Type: "string", Description: "The URL to link to"},
					"name":    {Type: "string", Description: "Display name for the link"},
				},
				Required: []string{"action", "bar_id"},
			},
		},
	}
}

// objectiveTools returns OKR-related tool definitions.
func objectiveTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "list_objectives",
			Description: "List all OKR objectives. Call this to see strategic goals. No parameters needed.",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name:        "get_objective",
			Description: "Get full details of an objective including its key results.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"objective_id": {Type: "string", Description: "The objective ID"},
				},
				Required: []string{"objective_id"},
			},
		},
		{
			Name:        "list_key_results",
			Description: "List key results for a specific objective.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"objective_id": {Type: "string", Description: "The objective ID"},
				},
				Required: []string{"objective_id"},
			},
		},
		{
			Name:        "manage_objective",
			Description: "Create, update, or delete an OKR objective.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":       {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"objective_id": {Type: "string", Description: "Objective ID (required for update/delete)"},
					"name":         {Type: "string", Description: "Objective name"},
					"description":  {Type: "string", Description: "Description"},
					"time_frame":   {Type: "string", Description: "Time frame (e.g., Q1 2024)"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name:        "manage_key_result",
			Description: "Create, update, or delete a key result for an objective.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":        {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"objective_id":  {Type: "string", Description: "Parent objective ID (required for all actions)"},
					"key_result_id": {Type: "string", Description: "Key result ID (required for update/delete)"},
					"name":          {Type: "string", Description: "Key result name"},
					"target_value":  {Type: "string", Description: "Target value"},
					"current_value": {Type: "string", Description: "Current value"},
				},
				Required: []string{"action", "objective_id"},
			},
		},
	}
}

// ideaTools returns idea and discovery tool definitions.
func ideaTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "list_ideas",
			Description: "List all ideas in the discovery/feedback pipeline. No parameters needed.",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name:        "get_idea",
			Description: "Get full details of a specific idea.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "The idea ID"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name:        "get_idea_customers",
			Description: "Get customers associated with an idea. Shows who requested or is impacted by an idea.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "The idea ID"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name:        "get_idea_tags",
			Description: "Get tags attached to an idea. Tags help categorize and filter ideas.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "The idea ID"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name:        "list_opportunities",
			Description: "List all opportunities in the discovery pipeline. Opportunities are validated ideas worth pursuing.",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name:        "get_opportunity",
			Description: "Get full details of a specific opportunity.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"opportunity_id": {Type: "string", Description: "The opportunity ID"},
				},
				Required: []string{"opportunity_id"},
			},
		},
		{
			Name:        "list_idea_forms",
			Description: "List all idea submission forms. Forms collect feedback from users and stakeholders.",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name:        "get_idea_form",
			Description: "Get full details of an idea form including its fields.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"form_id": {Type: "string", Description: "The idea form ID"},
				},
				Required: []string{"form_id"},
			},
		},
		{
			Name:        "manage_idea",
			Description: "Create or update an idea in the discovery pipeline.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"create", "update"}},
					"idea_id":     {Type: "string", Description: "Idea ID (required for update)"},
					"title":       {Type: "string", Description: "Idea title"},
					"description": {Type: "string", Description: "Idea description"},
					"status":      {Type: "string", Description: "Idea status"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name:        "manage_idea_customer",
			Description: "Add or remove a customer from an idea.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"add", "remove"}},
					"idea_id":     {Type: "string", Description: "The idea ID"},
					"customer_id": {Type: "string", Description: "Customer ID (required for remove)"},
					"name":        {Type: "string", Description: "Customer name (for add)"},
					"email":       {Type: "string", Description: "Customer email (for add)"},
				},
				Required: []string{"action", "idea_id"},
			},
		},
		{
			Name:        "manage_idea_tag",
			Description: "Add or remove a tag from an idea.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":  {Type: "string", Description: "Action to perform", Enum: []string{"add", "remove"}},
					"idea_id": {Type: "string", Description: "The idea ID"},
					"tag_id":  {Type: "string", Description: "Tag ID (required for remove)"},
					"name":    {Type: "string", Description: "Tag name (for add)"},
				},
				Required: []string{"action", "idea_id"},
			},
		},
		{
			Name:        "manage_opportunity",
			Description: "Create, update, or delete an opportunity.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":            {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"opportunity_id":    {Type: "string", Description: "Opportunity ID (required for update/delete)"},
					"problem_statement": {Type: "string", Description: "The opportunity problem statement (name)"},
					"description":       {Type: "string", Description: "Opportunity description"},
					"workflow_status":   {Type: "string", Description: "Status: draft, in_discovery, etc."},
				},
				Required: []string{"action"},
			},
		},
	}
}

// launchTools returns launch-related tool definitions.
func launchTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "list_launches",
			Description: "List all product launches. No parameters needed.",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name:        "get_launch",
			Description: "Get full details of a specific launch including checklist.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "The launch ID"},
				},
				Required: []string{"launch_id"},
			},
		},
	}
}

// utilityTools returns utility tool definitions.
func utilityTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "check_status",
			Description: "Check ProductPlan API status and authentication. Use to verify connection.",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name:        "health_check",
			Description: "Check server health, rate limit status, and cache statistics. Use 'deep' mode to also verify ProductPlan API connectivity.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"deep": {Type: "boolean", Description: "If true, also checks ProductPlan API connectivity (slower but more thorough)"},
				},
			},
		},
	}
}
