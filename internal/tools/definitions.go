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
			Description: `List all roadmaps. START HERE to get roadmap IDs.

Use when: "Show my roadmaps", "What roadmaps do I have?"
Returns: Array with IDs, names, metadata`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_roadmap",
			Description: `Get roadmap settings and metadata.

Use when: "Tell me about roadmap X", "Roadmap settings"
Returns: View settings, permissions, configuration
Requires: roadmap_id`,
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

Use when: "What's on the roadmap?", "Show planned features", "What's in Q2?"
Returns: Bars with names, dates, lanes, status
Requires: roadmap_id`,
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

Use when: "What lanes are on the roadmap?", "Show categories"
Returns: Lanes with IDs, names, colors
Requires: roadmap_id`,
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

Use when: "What are the key dates?", "Show milestones"
Returns: Milestones with names and dates
Requires: roadmap_id`,
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

Use when: "What colors are available?", "Show the legend"
Returns: Legends with IDs, colors, labels
Note: Use legend_id when creating/updating bars`,
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
			Description: `Get complete roadmap: details, bars, lanes, milestones in one call.

Use when: "Full roadmap overview", "Summarize roadmap X"
Returns: Combined details, bars, lanes, milestones
Performance: ~3x faster than sequential calls`,
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

Use when: "Show roadmap comments", "Roadmap discussion"
Returns: Comments with authors, dates, content
Requires: roadmap_id`,
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

Use when: "Add Backend lane", "Rename Mobile lane", "Delete lane"
Actions: create (name), update (lane_id), delete (lane_id)`,
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

Use when: "Add launch milestone", "Move demo date", "Delete milestone"
Actions: create (name+date), update (milestone_id), delete (milestone_id)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":       {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"roadmap_id":   {Type: "string", Description: "Roadmap ID"},
					"milestone_id": {Type: "string", Description: "Milestone ID (for update/delete)"},
					"name":         {Type: "string", Description: "Milestone name"},
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

Use when: "Tell me about this feature", "Bar details"
Returns: Description, dates, lane, links, custom fields
Requires: bar_id`,
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

Use when: "Show sub-tasks", "Child items", "Break down this feature"
Returns: Child bars with details
Requires: bar_id`,
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

Use when: "Show comments", "What's the feedback?"
Returns: Comments with author, date, text
Requires: bar_id`,
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

Use when: "What depends on this?", "Show dependencies"
Returns: Connected bars with relationship types
Requires: bar_id`,
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

Use when: "What's linked?", "Show Jira tickets"
Returns: Links with URLs and names
Requires: bar_id`,
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

Use when: "Add feature", "Update dates", "Delete item", "Change color"
Actions: create (roadmap_id+lane_id+name), update (bar_id), delete (bar_id)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":                 {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"bar_id":                 {Type: "string", Description: "Bar ID (for update/delete)"},
					"roadmap_id":             {Type: "string", Description: "Roadmap ID (for create)"},
					"lane_id":                {Type: "string", Description: "Lane ID (for create; update to move)"},
					"name":                   {Type: "string", Description: "Bar name"},
					"start_date":             {Type: "string", Description: "YYYY-MM-DD"},
					"end_date":               {Type: "string", Description: "YYYY-MM-DD"},
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
			Name: "manage_bar_comment",
			Description: `Add a comment to a bar.

Use when: "Add comment", "Leave feedback"`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_id": {Type: "string", Description: "Bar ID"},
					"body":   {Type: "string", Description: "Comment text"},
				},
				Required: []string{"bar_id", "body"},
			},
		},
		{
			Name: "manage_bar_connection",
			Description: `Create or delete dependency between bars.

Use when: "Link features", "Add dependency", "Remove dependency"
Actions: create (target_bar_id), delete (connection_id)`,
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
			Description: `Create, update, or delete external link on a bar.

Use when: "Link Jira ticket", "Add design doc", "Update URL"
Actions: create (url), update (link_id), delete (link_id)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":  {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"bar_id":  {Type: "string", Description: "Bar ID"},
					"link_id": {Type: "string", Description: "Link ID (for update/delete)"},
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

Use when: "Show OKRs", "What are our objectives?"
Returns: Objectives with IDs, names, time frames, progress`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_objective",
			Description: `Get objective details with key results.

Use when: "Tell me about objective X", "OKR progress"
Returns: Objective with key results, scores, status
Requires: objective_id`,
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

Use when: "What are the KRs?", "Show metrics"
Returns: Key results with targets, current values, progress
Requires: objective_id`,
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

Use when: "Tell me about this KR", "KR progress"
Returns: Target, current value, progress
Requires: objective_id + key_result_id`,
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

Use when: "Add Q1 objective", "Update objective", "Delete OKR"
Actions: create (name), update (objective_id), delete (objective_id)`,
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

Use when: "Add KR", "Update progress", "Delete KR"
Actions: create (name+target), update (key_result_id), delete (key_result_id)`,
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

Use when: "Show customer feedback", "What ideas do we have?"
Returns: Ideas with IDs, titles, votes, status`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_idea",
			Description: `Get idea details including description and metadata.

Use when: "Tell me about this idea", "Full request details"
Returns: Description, votes, customers, tags, status
Requires: idea_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "Idea ID from list_ideas"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name: "get_idea_customers",
			Description: `Get customers who requested an idea.

Use when: "Who requested this?", "Which customers want this?"
Returns: Customers with names, emails, vote counts
Requires: idea_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "Idea ID"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name: "get_idea_tags",
			Description: `Get tags on an idea.

Use when: "What tags does this have?", "How is this categorized?"
Returns: Tags with IDs and names
Requires: idea_id`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"idea_id": {Type: "string", Description: "Idea ID"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name: "list_opportunities",
			Description: `List all opportunities. START HERE for discovery.

Use when: "Show opportunities", "Discovery pipeline"
Returns: Opportunities with problem statements, status`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_opportunity",
			Description: `Get opportunity details with linked ideas.

Use when: "Tell me about this opportunity"
Returns: Description, linked ideas, workflow status
Requires: opportunity_id`,
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

Use when: "Show feedback forms", "What forms exist?"
Returns: Forms with IDs, names, configuration`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_idea_form",
			Description: `Get idea form details with fields.

Use when: "Show form fields", "What does this form collect?"
Returns: Fields, types, validation rules
Requires: form_id`,
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

Use when: "Who are our customers?", "All feedback sources"
Returns: Customers with IDs, names, emails, idea counts`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "list_all_tags",
			Description: `List all tags used across ideas.

Use when: "What tags exist?", "Show categories"
Returns: Tags with IDs and names`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "manage_idea",
			Description: `Create or update an idea. Note: delete not available via API.

Use when: "Add idea", "Update idea status"
Actions: create (title), update (idea_id)`,
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
			Name: "manage_idea_customer",
			Description: `Add or remove customer from an idea.

Use when: "Add customer to idea", "Remove customer"
Actions: add (name), remove (customer_id)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "add or remove", Enum: []string{"add", "remove"}},
					"idea_id":     {Type: "string", Description: "Idea ID"},
					"customer_id": {Type: "string", Description: "Customer ID (for remove)"},
					"name":        {Type: "string", Description: "Customer name (for add)"},
					"email":       {Type: "string", Description: "Customer email (for add)"},
				},
				Required: []string{"action", "idea_id"},
			},
		},
		{
			Name: "manage_idea_tag",
			Description: `Add or remove tag from an idea.

Use when: "Tag as mobile", "Remove enterprise tag"
Actions: add (name), remove (tag_id)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":  {Type: "string", Description: "add or remove", Enum: []string{"add", "remove"}},
					"idea_id": {Type: "string", Description: "Idea ID"},
					"tag_id":  {Type: "string", Description: "Tag ID (for remove)"},
					"name":    {Type: "string", Description: "Tag name (for add)"},
				},
				Required: []string{"action", "idea_id"},
			},
		},
		{
			Name: "manage_opportunity",
			Description: `Create, update, or delete an opportunity.

Use when: "Create opportunity", "Update problem", "Delete"
Actions: create (problem_statement), update (opportunity_id), delete (opportunity_id)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":            {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"opportunity_id":    {Type: "string", Description: "Opportunity ID (for update/delete)"},
					"problem_statement": {Type: "string", Description: "Problem statement (title)"},
					"description":       {Type: "string", Description: "Description"},
					"workflow_status":   {Type: "string", Description: "draft, in_discovery, validated"},
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

Use when: "Show launches", "Release schedule"
Returns: Launches with IDs, names, dates, status`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_launch",
			Description: `Get launch details with checklist.

Use when: "Tell me about this launch", "Launch readiness"
Returns: Dates, description, checklist, progress
Requires: launch_id`,
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

Use when: "Create launch", "Update date", "Delete launch"
Actions: create (name+date), update (launch_id), delete (launch_id)`,
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

Use when: "Show sections", "Checklist categories"
Returns: Sections with IDs, names, task counts
Requires: launch_id`,
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
			Description: `Get a specific checklist section.

Use when: "Section details"
Returns: Section ID, name, metadata
Requires: launch_id + section_id`,
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

Use when: "Add Marketing section", "Rename section", "Delete section"
Actions: create (name), update (section_id), delete (section_id)`,
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

Use when: "Show tasks", "What needs to be done?"
Returns: Tasks with IDs, names, assignees, due dates
Requires: launch_id`,
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
			Description: `Get a specific launch task.

Use when: "Task details", "Task status"
Returns: Name, description, assignee, due date, completed
Requires: launch_id + task_id`,
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

Use when: "Add task", "Mark complete", "Assign task", "Delete task"
Actions: create (name+section_id), update (task_id), delete (task_id)`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"launch_id":   {Type: "string", Description: "Launch ID"},
					"task_id":     {Type: "string", Description: "Task ID (for update/delete)"},
					"section_id":  {Type: "string", Description: "Section ID (for create)"},
					"name":        {Type: "string", Description: "Task name"},
					"description": {Type: "string", Description: "Task description"},
					"due_date":    {Type: "string", Description: "YYYY-MM-DD"},
					"assignee_id": {Type: "string", Description: "User ID to assign"},
					"completed":   {Type: "boolean", Description: "Is completed"},
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

Use when: "Is ProductPlan connected?", "Check API"
Returns: API status, auth state, account info`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "health_check",
			Description: `Check MCP server health and cache stats.

Use when: "Server status", "Rate limits", "Diagnose issues"
Returns: Uptime, rate limits, cache stats, API health (if deep)`,
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

Use when: "Who has access?", "Team members"
Returns: Users with IDs, names, emails, roles`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "list_teams",
			Description: `List all teams in account.

Use when: "What teams exist?", "Team structure"
Returns: Teams with IDs, names, member counts`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
	}
}
