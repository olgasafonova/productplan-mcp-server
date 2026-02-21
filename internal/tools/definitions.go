package tools

import (
	"strings"

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

	// Auto-annotate: read-only tools get ReadOnlyHint
	for i := range tools {
		if tools[i].Annotations != nil {
			continue
		}
		isReadOnly := strings.HasPrefix(tools[i].Name, "get_") ||
			strings.HasPrefix(tools[i].Name, "list_") ||
			strings.HasPrefix(tools[i].Name, "check_") ||
			tools[i].Name == "health_check"
		if isReadOnly {
			tools[i].Annotations = &mcp.ToolAnnotations{ReadOnlyHint: true}
		}
	}

	return tools
}

// roadmapTools returns roadmap-related tool definitions.
func roadmapTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "list_roadmaps",
			Description: `List all roadmaps. START HERE to get roadmap IDs.

USE WHEN: "Show my roadmaps", "What roadmaps do I have?"
FAILS WHEN: API token invalid or expired (check PRODUCTPLAN_API_TOKEN env var).`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_roadmap",
			Description: `Get roadmap settings and metadata.

USE WHEN: "Tell me about roadmap X", "Roadmap settings"
For all data in one call (bars, lanes, milestones), use get_roadmap_complete.
FAILS WHEN: roadmap_id not found (get valid IDs from list_roadmaps first).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID from list_roadmaps"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_bars",
			Description: `Get all bars (features/items) on a roadmap.

USE WHEN: "What's on the roadmap?", "Show planned features", "What's in Q2?"
FAILS WHEN: roadmap_id not found (use list_roadmaps). Returns empty list if roadmap has no bars.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_lanes",
			Description: `Get lanes (categories) on a roadmap. Lanes organize bars into rows.

USE WHEN: "What lanes are on the roadmap?", "Show categories"
FAILS WHEN: roadmap_id not found (use list_roadmaps).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_milestones",
			Description: `Get milestones (key dates) on a roadmap.

USE WHEN: "What are the key dates?", "Show milestones"
FAILS WHEN: roadmap_id not found (use list_roadmaps).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_legends",
			Description: `Get legend entries (bar colors) for a roadmap.

USE WHEN: "What colors are available?", "Show the legend"
Note: Use legend_id when creating/updating bars.
FAILS WHEN: roadmap_id not found (use list_roadmaps).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_complete",
			Description: `Get complete roadmap in one call (~3x faster than sequential). Details, bars, lanes, milestones combined.

USE WHEN: "Full roadmap overview", "Summarize roadmap X"
For settings/metadata only, use get_roadmap.
WHY: Makes 3 parallel API calls internally (~3x faster than calling get_roadmap + get_roadmap_bars + get_roadmap_lanes sequentially).
FAILS WHEN: roadmap_id not found (use list_roadmaps).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "get_roadmap_comments",
			Description: `Get roadmap-level comments (not bar comments).

USE WHEN: "Show roadmap comments", "Roadmap discussion"
For bar-level comments, use get_bar_comments instead.
FAILS WHEN: roadmap_id not found (use list_roadmaps).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name: "manage_lane",
			Description: `Create, update, or delete a lane on a roadmap.

USE WHEN: "Add Backend lane", "Rename Mobile lane", "Delete lane"
Actions: create (name), update (lane_id), delete (lane_id)
FAILS WHEN: create without name, update/delete without lane_id (get IDs from get_roadmap_lanes). WARNING: delete removes the lane and unassigns all bars in it.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":     {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
					"lane_id":    {Type: "string", Description: "Lane ID (for update/delete)"},
					"name":       {Type: "string", Description: "Lane name"},
					"color":      {Type: "string", Description: "Hex color (#FF5733)"},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
		{
			Name: "manage_milestone",
			Description: `Create, update, or delete a milestone on a roadmap.

USE WHEN: "Add launch milestone", "Move demo date", "Delete milestone"
Actions: create (title+date), update (milestone_id), delete (milestone_id)
FAILS WHEN: create without title or date, update/delete without milestone_id (get IDs from get_roadmap_milestones), date not in YYYY-MM-DD format.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":       {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"roadmap_id":   {Type: "string", Description: "Roadmap ID"},
					"milestone_id": {Type: "string", Description: "Milestone ID (for update/delete)"},
					"title":        {Type: "string", Description: "Milestone title"},
					"date":         {Type: "string", Description: "YYYY-MM-DD format"},
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
			Description: `Get bar details including description, links, custom fields.

USE WHEN: "Tell me about this feature", "Bar details"
FAILS WHEN: bar_id not found (get valid IDs from get_roadmap_bars).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "Bar ID from get_roadmap_bars"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "get_bar_children",
			Description: `Get child bars nested under a parent bar.

USE WHEN: "Show sub-tasks", "Child items", "Break down this feature"
FAILS WHEN: bar_id not found. Returns empty list if bar has no children (not all bars are containers).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "Parent bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "get_bar_comments",
			Description: `Get comments on a bar.

USE WHEN: "Show comments", "What's the feedback on this bar?"
For roadmap-level comments, use get_roadmap_comments instead.
FAILS WHEN: bar_id not found.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "Bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "get_bar_connections",
			Description: `Get bar dependencies (what blocks what).

USE WHEN: "What depends on this?", "Show dependencies"
FAILS WHEN: bar_id not found. Returns empty list if bar has no connections.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "Bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "get_bar_links",
			Description: `Get external links on a bar (Jira, docs, designs).

USE WHEN: "What's linked?", "Show Jira tickets"
FAILS WHEN: bar_id not found. Returns empty list if bar has no external links.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "Bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name: "manage_bar",
			Description: `Create, update, or delete a bar on a roadmap.

USE WHEN: "Add feature", "Update dates", "Delete item", "Change color"
Actions: create (roadmap_id+lane_id+name), update (bar_id), delete (bar_id)
FAILS WHEN: create without roadmap_id, lane_id, or name (all three required). Update/delete without bar_id. Use get_roadmap_legends for valid legend_id values. WARNING: delete is permanent and cannot be undone.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":                 {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"bar_id":                 {Type: "string", Description: "Bar ID (for update/delete)"},
					"roadmap_id":             {Type: "string", Description: "Roadmap ID (for create)"},
					"lane_id":                {Type: "string", Description: "Lane ID (for create; update to move)"},
					"name":                   {Type: "string", Description: "Bar name"},
					"starts_on":              {Type: "string", Description: "Start date YYYY-MM-DD"},
					"ends_on":                {Type: "string", Description: "End date YYYY-MM-DD"},
					"description":            {Type: "string", Description: "Description (markdown)"},
					"legend_id":              {Type: "string", Description: "Color from get_roadmap_legends"},
					"percent_done":           {Type: "integer", Description: "Progress 0-100"},
					"container":              {Type: "boolean", Description: "Is container for children"},
					"parked":                 {Type: "boolean", Description: "Not actively scheduled"},
					"parent_id":              {Type: "string", Description: "Parent bar ID for nesting"},
					"strategic_value":        {Type: "string", Description: "Strategic importance"},
					"notes":                  {Type: "string", Description: "Additional notes"},
					"effort":                 {Type: "integer", Description: "Effort estimate"},
					"tags":                   {Type: "array", Description: "Tag strings [\"mobile\",\"urgent\"]"},
					"custom_text_fields":     {Type: "array", Description: "[{name,value}] custom text fields"},
					"custom_dropdown_fields": {Type: "array", Description: "[{name,value}] custom dropdowns"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name: "manage_bar_connection",
			Description: `Create or delete dependency between bars.

USE WHEN: "Link features", "Add dependency", "Remove dependency"
Actions: create (target_bar_id), delete (connection_id)
FAILS WHEN: create without target_bar_id, delete without connection_id (get IDs from get_bar_connections).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":        {Type: "string", Description: "create or delete", Enum: []string{"create", "delete"}},
					"bar_id":        {Type: "string", Description: "Source bar ID"},
					"target_bar_id": {Type: "string", Description: "Target bar (for create)"},
					"connection_id": {Type: "string", Description: "Connection ID (for delete)"},
				},
				Required: []string{"action", "bar_id"},
			},
		},
		{
			Name: "manage_bar_link",
			Description: `Create or delete external link on a bar.

USE WHEN: "Link Jira ticket", "Add design doc", "Remove link"
Actions: create (url), delete (link_id)
FAILS WHEN: create without url, delete without link_id (get IDs from get_bar_links). Note: update not available via API; delete and re-create instead.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":  {Type: "string", Description: "create or delete", Enum: []string{"create", "delete"}},
					"bar_id":  {Type: "string", Description: "Bar ID"},
					"link_id": {Type: "string", Description: "Link ID (for delete)"},
					"url":     {Type: "string", Description: "URL to link"},
					"name":    {Type: "string", Description: "Display name"},
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
			Description: `List all OKR objectives. START HERE for OKRs.

USE WHEN: "Show OKRs", "What are our objectives?"
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
		{
			Name: "manage_objective",
			Description: `Create, update, or delete an objective.

USE WHEN: "Add Q1 objective", "Update objective", "Delete OKR"
Actions: create (name), update (objective_id), delete (objective_id)
FAILS WHEN: create without name, update/delete without objective_id. WARNING: delete also removes all key results under this objective.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":       {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"objective_id": {Type: "string", Description: "Objective ID (for update/delete)"},
					"name":         {Type: "string", Description: "Objective name"},
					"description":  {Type: "string", Description: "Description"},
					"time_frame":   {Type: "string", Description: "Q1 2024, H1 2024, 2024"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name: "manage_key_result",
			Description: `Create, update, or delete a key result.

USE WHEN: "Add KR", "Update progress", "Delete KR"
Actions: create (name+target), update (key_result_id), delete (key_result_id)
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

// ideaTools returns idea and discovery tool definitions.
func ideaTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "list_ideas",
			Description: `List all ideas in discovery pipeline. START HERE for ideas.

USE WHEN: "Show customer feedback", "What ideas do we have?"
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
		{
			Name: "list_opportunities",
			Description: `List all opportunities. START HERE for discovery.

USE WHEN: "Show opportunities", "Discovery pipeline"
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
			Description: `List all customers across ideas.

USE WHEN: "Who are our customers?", "All feedback sources"
For customers linked to a specific idea, use get_idea_customers instead.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "list_all_tags",
			Description: `List all tags used across ideas.

USE WHEN: "What tags exist?", "Show categories"
For tags on a specific idea, use get_idea_tags instead.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "manage_idea",
			Description: `Create or update an idea. Note: delete not available via API.

USE WHEN: "Add idea", "Update idea status"
Actions: create (title), update (idea_id)
FAILS WHEN: create without title, update without idea_id. Note: delete is not available via the ProductPlan API; archive ideas by updating status instead.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "create or update", Enum: []string{"create", "update"}},
					"idea_id":     {Type: "string", Description: "Idea ID (for update)"},
					"title":       {Type: "string", Description: "Idea title"},
					"description": {Type: "string", Description: "Description (markdown)"},
					"status":      {Type: "string", Description: "new, under_review, planned"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name: "manage_opportunity",
			Description: `Create or update an opportunity. Note: delete not available via API.

USE WHEN: "Create opportunity", "Update problem"
Actions: create (problem_statement), update (opportunity_id)
FAILS WHEN: create without problem_statement, update without opportunity_id (get IDs from list_opportunities). Note: delete is not available via the ProductPlan API; archive opportunities by updating workflow_status instead.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":            {Type: "string", Description: "create or update", Enum: []string{"create", "update"}},
					"opportunity_id":    {Type: "string", Description: "Opportunity ID (for update)"},
					"problem_statement": {Type: "string", Description: "Problem statement (title)"},
					"description":       {Type: "string", Description: "Description"},
					"workflow_status":   {Type: "string", Description: "draft, in_discovery, validated, invalidated, completed"},
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
			Description: `List all launches. START HERE for launches.

USE WHEN: "Show launches", "Release schedule"
FAILS WHEN: API token invalid. Returns empty list if no launches exist.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_launch",
			Description: `Get launch details with checklist.

USE WHEN: "Tell me about this launch", "Launch readiness"
FAILS WHEN: launch_id not found (get valid IDs from list_launches).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "Launch ID from list_launches"},
				},
				Required: []string{"launch_id"},
			},
		},
		{
			Name: "manage_launch",
			Description: `Create, update, or delete a launch.

USE WHEN: "Create launch", "Update date", "Delete launch"
Actions: create (name+date), update (launch_id), delete (launch_id)
FAILS WHEN: create without name or date, update/delete without launch_id, date not in YYYY-MM-DD format. WARNING: delete removes the launch and all its sections and tasks.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"launch_id":   {Type: "string", Description: "Launch ID (for update/delete)"},
					"name":        {Type: "string", Description: "Launch name"},
					"date":        {Type: "string", Description: "YYYY-MM-DD"},
					"description": {Type: "string", Description: "Description"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name: "get_launch_sections",
			Description: `Get checklist sections for a launch.

USE WHEN: "Show sections", "Checklist categories"
For one specific section, use get_launch_section.
FAILS WHEN: launch_id not found.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "Launch ID"},
				},
				Required: []string{"launch_id"},
			},
		},
		{
			Name: "get_launch_section",
			Description: `Get a specific checklist section by ID.

USE WHEN: "Section details"
For all sections, use get_launch_sections.
FAILS WHEN: launch_id or section_id not found.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id":  {Type: "string", Description: "Launch ID"},
					"section_id": {Type: "string", Description: "Section ID"},
				},
				Required: []string{"launch_id", "section_id"},
			},
		},
		{
			Name: "manage_launch_section",
			Description: `Create, update, or delete a checklist section.

USE WHEN: "Add Marketing section", "Rename section", "Delete section"
Actions: create (name), update (section_id), delete (section_id)
FAILS WHEN: create without name, update/delete without section_id (get IDs from get_launch_sections). WARNING: delete removes the section and all tasks in it.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":     {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"launch_id":  {Type: "string", Description: "Launch ID"},
					"section_id": {Type: "string", Description: "Section ID (for update/delete)"},
					"name":       {Type: "string", Description: "Section name"},
				},
				Required: []string{"action", "launch_id"},
			},
		},
		{
			Name: "get_launch_tasks",
			Description: `Get all tasks for a launch.

USE WHEN: "Show tasks", "What needs to be done?"
For one specific task, use get_launch_task.
FAILS WHEN: launch_id not found.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "Launch ID"},
				},
				Required: []string{"launch_id"},
			},
		},
		{
			Name: "get_launch_task",
			Description: `Get a specific launch task by ID.

USE WHEN: "Task details", "Task status"
For all tasks, use get_launch_tasks.
FAILS WHEN: launch_id or task_id not found.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "Launch ID"},
					"task_id":   {Type: "string", Description: "Task ID"},
				},
				Required: []string{"launch_id", "task_id"},
			},
		},
		{
			Name: "manage_launch_task",
			Description: `Create, update, or delete a launch task.

USE WHEN: "Add task", "Mark complete", "Assign task", "Delete task"
Actions: create (name+section_id), update (task_id), delete (task_id)
FAILS WHEN: create without name or section_id, update/delete without task_id (get IDs from get_launch_tasks). Use list_users to get valid assigned_user_id values.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":           {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"launch_id":        {Type: "string", Description: "Launch ID"},
					"task_id":          {Type: "string", Description: "Task ID (for update/delete)"},
					"section_id":       {Type: "string", Description: "Section ID (for create)"},
					"name":             {Type: "string", Description: "Task name"},
					"description":      {Type: "string", Description: "Task description"},
					"due_date":         {Type: "string", Description: "YYYY-MM-DD"},
					"assigned_user_id": {Type: "string", Description: "User ID to assign (get from list_users)"},
					"status":           {Type: "string", Description: "Task status", Enum: []string{"to_do", "in_progress", "completed", "blocked"}},
				},
				Required: []string{"action", "launch_id"},
			},
		},
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
Use user IDs from this tool when assigning launch tasks via manage_launch_task.`,
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
