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
			Name: "list_roadmaps",
			Description: `List all roadmaps in ProductPlan. START HERE to get roadmap IDs before accessing bars, lanes, or milestones.

Use when: "Show my roadmaps", "What roadmaps do I have?", "List all product roadmaps"
Returns: Array of roadmaps with IDs, names, and basic metadata
No parameters needed.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_roadmap",
			Description: `Get detailed settings and metadata for a specific roadmap.

Use when: "Tell me about roadmap X", "What are the settings for this roadmap?"
Returns: Full roadmap details including view settings, permissions, and configuration
Requires: roadmap_id (get from list_roadmaps first)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID (get from list_roadmaps)"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_bars",
			Description: `Get all bars (planned items/features) on a roadmap with their lane context.

Use when: "What's on the roadmap?", "Show me planned features", "What's in Q2?"
Returns: Array of bars with names, dates, lanes, and status
Requires: roadmap_id (get from list_roadmaps first)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_lanes",
			Description: `Get all lanes (swim lanes/categories) on a roadmap. Lanes organize bars into rows like "Mobile", "Web", "API".

Use when: "What lanes are on the roadmap?", "Show me the categories"
Returns: Array of lanes with IDs, names, and colors
Requires: roadmap_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_milestones",
			Description: `Get all milestones (key dates/deadlines) on a roadmap like launches, demos, or releases.

Use when: "What are the key dates?", "Show milestones", "When are releases planned?"
Returns: Array of milestones with names and dates
Requires: roadmap_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_complete",
			Description: `Get complete roadmap data: details, bars, lanes, and milestones in a single fast call.

Use when: "Give me everything about this roadmap", "Full roadmap overview", "Summarize roadmap X"
Returns: Combined object with roadmap details, all bars, lanes, and milestones
Performance: Fetches all data in parallel (~3x faster than sequential calls)
Requires: roadmap_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID (get from list_roadmaps)"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "manage_lane",
			Description: `Create, update, or delete a lane (swim lane/category) on a roadmap.

Use when: "Add a new lane for Backend", "Rename the Mobile lane", "Delete unused lane"
Actions: create (needs name), update (needs lane_id + fields), delete (needs lane_id)
Requires: action + roadmap_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":     {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"roadmap_id": {Type: "string", Description: "Roadmap ID (required for all actions)"},
					"lane_id":    {Type: "string", Description: "Lane ID (required for update/delete)"},
					"name":       {Type: "string", Description: "Lane name"},
					"color":      {Type: "string", Description: "Lane color hex code (e.g., #FF5733)"},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
		{
			Name: "manage_milestone",
			Description: `Create, update, or delete a milestone (key date) on a roadmap.

Use when: "Add a launch milestone for March 1st", "Move the demo date", "Delete milestone"
Actions: create (needs name + date), update (needs milestone_id + fields), delete (needs milestone_id)
Requires: action + roadmap_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":       {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"roadmap_id":   {Type: "string", Description: "Roadmap ID (required for all actions)"},
					"milestone_id": {Type: "string", Description: "Milestone ID (required for update/delete)"},
					"name":         {Type: "string", Description: "Milestone name"},
					"date":         {Type: "string", Description: "Date in YYYY-MM-DD format"},
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
			Name: "get_bar",
			Description: `Get full details of a specific bar (feature/item) including description, links, and custom fields.

Use when: "Tell me about this feature", "What are the details of bar X?"
Returns: Complete bar data with description, dates, lane, links, and custom fields
Requires: bar_id (get from get_roadmap_bars)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The bar ID (get from get_roadmap_bars)"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "get_bar_children",
			Description: `Get child bars (sub-items/tasks) nested under a parent bar.

Use when: "What are the sub-tasks?", "Show child items", "Break down this feature"
Returns: Array of child bars with their details
Requires: bar_id of the parent bar`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The parent bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "get_bar_comments",
			Description: `Get all comments and discussion on a bar.

Use when: "What's the feedback on this?", "Show comments", "What did the team say?"
Returns: Array of comments with author, date, and text
Requires: bar_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "get_bar_connections",
			Description: `Get dependencies and connections between bars (what blocks what).

Use when: "What depends on this?", "Show dependencies", "What's blocking this feature?"
Returns: Array of connected bars with relationship types
Requires: bar_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "get_bar_links",
			Description: `Get external links attached to a bar (Jira tickets, docs, designs, PRDs).

Use when: "What's linked to this?", "Show Jira tickets", "Where's the design doc?"
Returns: Array of links with URLs and display names
Requires: bar_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "The bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "manage_bar",
			Description: `Create, update, or delete a bar (feature/item) on a roadmap.

Use when: "Add a new feature", "Update the dates", "Delete this item", "Move to different lane"
Actions: create (needs roadmap_id + lane_id + name), update (needs bar_id + fields), delete (needs bar_id)
Requires: action`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"bar_id":      {Type: "string", Description: "Bar ID (required for update/delete)"},
					"roadmap_id":  {Type: "string", Description: "Roadmap ID (required for create)"},
					"lane_id":     {Type: "string", Description: "Lane ID (required for create, optional for update to move bar)"},
					"name":        {Type: "string", Description: "Bar name/title"},
					"start_date":  {Type: "string", Description: "Start date in YYYY-MM-DD format"},
					"end_date":    {Type: "string", Description: "End date in YYYY-MM-DD format"},
					"description": {Type: "string", Description: "Description text (supports markdown)"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name: "manage_bar_comment",
			Description: `Add a comment to a bar for discussion and feedback.

Use when: "Add a comment", "Leave feedback", "Note that..."
Requires: bar_id + body`,
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
			Name: "manage_bar_connection",
			Description: `Create or delete a dependency/connection between two bars.

Use when: "Link these features", "This depends on that", "Remove dependency"
Actions: create (needs bar_id + target_bar_id), delete (needs bar_id + connection_id)
Requires: action + bar_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":        {Type: "string", Description: "Action to perform", Enum: []string{"create", "delete"}},
					"bar_id":        {Type: "string", Description: "Source bar ID"},
					"target_bar_id": {Type: "string", Description: "Target bar ID to connect to (for create)"},
					"connection_id": {Type: "string", Description: "Connection ID (required for delete, get from get_bar_connections)"},
				},
				Required: []string{"action", "bar_id"},
			},
		},
		{
			Name: "manage_bar_link",
			Description: `Create, update, or delete an external link on a bar (Jira, docs, designs).

Use when: "Link this Jira ticket", "Add a design doc link", "Update the URL"
Actions: create (needs url), update (needs link_id + fields), delete (needs link_id)
Requires: action + bar_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":  {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"bar_id":  {Type: "string", Description: "The bar ID"},
					"link_id": {Type: "string", Description: "Link ID (required for update/delete, get from get_bar_links)"},
					"url":     {Type: "string", Description: "The URL to link to"},
					"name":    {Type: "string", Description: "Display name for the link (optional, defaults to URL)"},
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
			Name: "list_objectives",
			Description: `List all OKR objectives (strategic goals). START HERE for OKRs.

Use when: "Show our OKRs", "What are our objectives?", "List strategic goals"
Returns: Array of objectives with IDs, names, time frames, and progress
No parameters needed.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_objective",
			Description: `Get full details of an objective including its key results and progress.

Use when: "Tell me about objective X", "What are the key results?", "Show OKR progress"
Returns: Complete objective with all key results, scores, and status
Requires: objective_id (get from list_objectives)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"objective_id": {Type: "string", Description: "The objective ID (get from list_objectives)"},
				},
				Required: []string{"objective_id"},
			},
		},
		{
			Name: "list_key_results",
			Description: `List key results (measurable outcomes) for a specific objective.

Use when: "What are the KRs?", "Show metrics for this objective"
Returns: Array of key results with targets, current values, and progress
Requires: objective_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"objective_id": {Type: "string", Description: "The objective ID"},
				},
				Required: []string{"objective_id"},
			},
		},
		{
			Name: "manage_objective",
			Description: `Create, update, or delete an OKR objective.

Use when: "Add a new objective for Q1", "Update the objective name", "Delete this OKR"
Actions: create (needs name), update (needs objective_id + fields), delete (needs objective_id)
Requires: action`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":       {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"objective_id": {Type: "string", Description: "Objective ID (required for update/delete)"},
					"name":         {Type: "string", Description: "Objective name (the O in OKR)"},
					"description":  {Type: "string", Description: "Description of the objective"},
					"time_frame":   {Type: "string", Description: "Time frame like Q1 2024, H1 2024, or 2024"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name: "manage_key_result",
			Description: `Create, update, or delete a key result (KR) for an objective.

Use when: "Add a key result", "Update progress to 75%", "Delete this KR"
Actions: create (needs name + target_value), update (needs key_result_id + fields), delete (needs key_result_id)
Requires: action + objective_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":        {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"objective_id":  {Type: "string", Description: "Parent objective ID (required for all actions)"},
					"key_result_id": {Type: "string", Description: "Key result ID (required for update/delete)"},
					"name":          {Type: "string", Description: "Key result name (the KR in OKR)"},
					"target_value":  {Type: "string", Description: "Target value (e.g., 100, 50%, $1M)"},
					"current_value": {Type: "string", Description: "Current value/progress"},
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
			Name: "list_ideas",
			Description: `List all ideas in the discovery/feedback pipeline. START HERE for ideas.

Use when: "Show customer feedback", "What ideas do we have?", "List feature requests"
Returns: Array of ideas with IDs, titles, votes, status, and submission info
No parameters needed.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_idea",
			Description: `Get full details of a specific idea including description and metadata.

Use when: "Tell me about this idea", "What's the full request?"
Returns: Complete idea with description, votes, customers, tags, and status
Requires: idea_id (get from list_ideas)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "The idea ID (get from list_ideas)"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name: "get_idea_customers",
			Description: `Get customers who requested or are impacted by an idea.

Use when: "Who requested this?", "Which customers want this feature?"
Returns: Array of customers with names, emails, and vote counts
Requires: idea_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "The idea ID"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name: "get_idea_tags",
			Description: `Get tags/labels attached to an idea for categorization.

Use when: "What tags does this have?", "How is this categorized?"
Returns: Array of tags with IDs and names
Requires: idea_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "The idea ID"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name: "list_opportunities",
			Description: `List all opportunities (validated ideas worth pursuing). START HERE for discovery.

Use when: "Show opportunities", "What problems are we exploring?", "Discovery pipeline"
Returns: Array of opportunities with problem statements, status, and scores
No parameters needed.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_opportunity",
			Description: `Get full details of a specific opportunity including linked ideas.

Use when: "Tell me about this opportunity", "What's the problem statement?"
Returns: Complete opportunity with description, linked ideas, and workflow status
Requires: opportunity_id (get from list_opportunities)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"opportunity_id": {Type: "string", Description: "The opportunity ID (get from list_opportunities)"},
				},
				Required: []string{"opportunity_id"},
			},
		},
		{
			Name: "list_idea_forms",
			Description: `List all idea submission forms for collecting feedback.

Use when: "Show feedback forms", "What forms do we have?"
Returns: Array of forms with IDs, names, and configuration
No parameters needed.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_idea_form",
			Description: `Get full details of an idea form including its fields and settings.

Use when: "Show form fields", "What does this form collect?"
Returns: Complete form with all fields, types, and validation rules
Requires: form_id (get from list_idea_forms)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"form_id": {Type: "string", Description: "The idea form ID (get from list_idea_forms)"},
				},
				Required: []string{"form_id"},
			},
		},
		{
			Name: "manage_idea",
			Description: `Create or update an idea in the discovery pipeline.

Use when: "Add a new idea", "Update idea status", "Change idea title"
Actions: create (needs title), update (needs idea_id + fields)
Note: Ideas cannot be deleted via API
Requires: action`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"create", "update"}},
					"idea_id":     {Type: "string", Description: "Idea ID (required for update)"},
					"title":       {Type: "string", Description: "Idea title"},
					"description": {Type: "string", Description: "Idea description (supports markdown)"},
					"status":      {Type: "string", Description: "Idea status (e.g., new, under_review, planned)"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name: "manage_idea_customer",
			Description: `Add or remove a customer from an idea (who requested/wants it).

Use when: "Add customer to this idea", "Remove customer from idea"
Actions: add (needs name), remove (needs customer_id)
Requires: action + idea_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"add", "remove"}},
					"idea_id":     {Type: "string", Description: "The idea ID"},
					"customer_id": {Type: "string", Description: "Customer ID (required for remove, get from get_idea_customers)"},
					"name":        {Type: "string", Description: "Customer name (for add)"},
					"email":       {Type: "string", Description: "Customer email (optional, for add)"},
				},
				Required: []string{"action", "idea_id"},
			},
		},
		{
			Name: "manage_idea_tag",
			Description: `Add or remove a tag/label from an idea for categorization.

Use when: "Tag this as mobile", "Remove the enterprise tag"
Actions: add (needs name), remove (needs tag_id)
Requires: action + idea_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":  {Type: "string", Description: "Action to perform", Enum: []string{"add", "remove"}},
					"idea_id": {Type: "string", Description: "The idea ID"},
					"tag_id":  {Type: "string", Description: "Tag ID (required for remove, get from get_idea_tags)"},
					"name":    {Type: "string", Description: "Tag name (for add, creates if doesn't exist)"},
				},
				Required: []string{"action", "idea_id"},
			},
		},
		{
			Name: "manage_opportunity",
			Description: `Create, update, or delete an opportunity in the discovery pipeline.

Use when: "Create an opportunity", "Update problem statement", "Delete opportunity"
Actions: create (needs problem_statement), update (needs opportunity_id + fields), delete (needs opportunity_id)
Requires: action`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":            {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"opportunity_id":    {Type: "string", Description: "Opportunity ID (required for update/delete)"},
					"problem_statement": {Type: "string", Description: "The opportunity problem statement (main title)"},
					"description":       {Type: "string", Description: "Detailed description of the opportunity"},
					"workflow_status":   {Type: "string", Description: "Status: draft, in_discovery, validated, etc."},
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
			Name: "list_launches",
			Description: `List all product launches with their status and dates. START HERE for launches.

Use when: "Show upcoming launches", "What launches do we have?", "Release schedule"
Returns: Array of launches with IDs, names, dates, and status
No parameters needed.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_launch",
			Description: `Get full details of a specific launch including checklist items.

Use when: "Tell me about this launch", "What's the launch checklist?", "Launch readiness"
Returns: Complete launch with dates, description, checklist items, and progress
Requires: launch_id (get from list_launches)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "The launch ID (get from list_launches)"},
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
			Name: "check_status",
			Description: `Check ProductPlan API status and verify authentication is working.

Use when: "Is ProductPlan connected?", "Check API status", "Verify my token works"
Returns: API status, authentication state, and account info
No parameters needed.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "health_check",
			Description: `Check MCP server health, rate limits, and cache statistics.

Use when: "Server status", "Am I rate limited?", "Cache stats", "Diagnose issues"
Returns: Server uptime, rate limit status, cache hit rates, and API health (if deep=true)
Optional: deep=true for full API connectivity check (slower)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"deep": {Type: "boolean", Description: "If true, also verifies ProductPlan API connectivity (adds ~500ms)"},
				},
			},
		},
	}
}
