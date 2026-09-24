package tools

import (
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// roadmapTools returns roadmap-related tool definitions.
func roadmapTools() []mcp.Tool {
	tools := roadmapReadTools()
	return append(tools, roadmapManageTools()...)
}

// roadmapReadTools returns read-only roadmap tool definitions.
func roadmapReadTools() []mcp.Tool {
	tools := roadmapMetaTools()
	return append(tools, roadmapDetailTools()...)
}

// roadmapMetaTools returns read tools for the roadmap itself (no sub-resources).
func roadmapMetaTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "list_roadmaps",
			Description: `List all roadmaps. START HERE to get roadmap IDs.

USE WHEN: "Show my roadmaps", "What roadmaps do I have?"
Returns array of roadmaps with ID, name, and creation date.
FAILS WHEN: API token invalid or expired (check PRODUCTPLAN_API_TOKEN env var).`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_roadmap",
			Description: `Get roadmap settings and metadata.

USE WHEN: "Tell me about roadmap X", "Roadmap settings", "What custom fields does this roadmap have?"
For all data in one call (bars, lanes, milestones), use get_roadmap_complete.
Returns roadmap name, date range, sharing settings, and metadata, including the vocabulary bar writes must use: legends (names), lanes (names), custom_text_fields [{label}], and custom_dropdown_fields [{label, allowed_values}]. manage_bar and the bulk_*_bars tools validate against exactly these.
FAILS WHEN: roadmap_id not found (get valid IDs from list_roadmaps first).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID from list_roadmaps"},
				},
				Required: []string{"roadmap_id"},
			},
		},
	}
}

// roadmapDetailTools returns read tools for roadmap sub-resources (bars, lanes, milestones, etc).
func roadmapDetailTools() []mcp.Tool {
	tools := roadmapComponentTools()
	return append(tools, roadmapAggregateTools()...)
}

// roadmapComponentTools returns read tools for individual roadmap sub-collections.
func roadmapComponentTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "get_roadmap_bars",
			Description: `Get all bars (features/items) on a roadmap.

USE WHEN: "What's on the roadmap?", "Show planned features", "What's in Q2?"
Returns array of bars with ID, name, dates, lane, legend, percent_done, and description.
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
Returns array of lanes with ID, name, and color.
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
Returns array of milestones with ID, title, and date.
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
			Description: `Get the legend names (bar colors) for a roadmap.

USE WHEN: "What colors are available?", "Show the legend", "Which legend can I color bars with?"
Returns an array of legend names. The ProductPlan API exposes legends by name only: there is no legend ID or hex color to return. Pass a name as ` + "`legend`" + ` on manage_bar, bulk_update_bars, or bulk_create_bars to color a bar.
For custom field definitions (labels, dropdown allowed_values), use get_roadmap instead.
FAILS WHEN: roadmap_id not found (use list_roadmaps).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
	}
}

// roadmapAggregateTools returns aggregate/cross-cutting roadmap read tools
// (complete dump, roadmap-level comments).
func roadmapAggregateTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "get_roadmap_complete",
			Description: `Get complete roadmap in one call. Details, bars, lanes, milestones combined.

USE WHEN: "Full roadmap overview", "Summarize roadmap X"
For settings/metadata only, use get_roadmap.
Returns combined roadmap details, bars, lanes, and milestones in one response.
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
Returns array of comments with author, body, and timestamp.
FAILS WHEN: roadmap_id not found (use list_roadmaps).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
	}
}

// roadmapManageTools returns roadmap mutation tool definitions (manage_lane, manage_milestone).
func roadmapManageTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "manage_lane",
			Description: `Create, update, or delete a lane on a roadmap.

USE WHEN: "Add Backend lane", "Rename Mobile lane", "Delete lane"
Actions: create (name), update (lane_id), delete (lane_id)
Returns the created/updated lane object, or confirmation on delete.
FAILS WHEN: create without name, update/delete without lane_id (get IDs from get_roadmap_lanes). WARNING: delete removes the lane and unassigns all bars in it.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":     {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"roadmap_id": {Type: "string", Description: "Roadmap ID"},
					"lane_id":    {Type: "string", Description: "Lane ID (for update/delete)"},
					"name":       {Type: "string", Description: "Lane name"},
					"color":      {Type: "string", Description: "Hex color (#FF5733)", Pattern: `^#[0-9A-Fa-f]{6}$`, Examples: []any{"#FF5733", "#4CAF50"}},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
		{
			Name: "manage_milestone",
			Description: `Create, update, or delete a milestone on a roadmap.

USE WHEN: "Add launch milestone", "Move demo date", "Delete milestone"
Actions: create (title+date), update (milestone_id), delete (milestone_id)
Returns the created/updated milestone object, or confirmation on delete.
FAILS WHEN: create without title or date, update/delete without milestone_id (get IDs from get_roadmap_milestones), date not in YYYY-MM-DD format.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":       {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"roadmap_id":   {Type: "string", Description: "Roadmap ID"},
					"milestone_id": {Type: "string", Description: "Milestone ID (for update/delete)"},
					"title":        {Type: "string", Description: "Milestone title"},
					"date":         {Type: "string", Description: "YYYY-MM-DD format", Pattern: `^\d{4}-\d{2}-\d{2}$`, Examples: []any{"2025-06-01"}},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
	}
}

// barTools returns bar-related tool definitions.
func barTools() []mcp.Tool {
	tools := barReadTools()
	return append(tools, barManageTools()...)
}

// barReadTools returns read-only bar tool definitions.
func barReadTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "get_bar",
			Description: `Get bar details including description, links, custom fields.

USE WHEN: "Tell me about this feature", "Bar details"
Returns bar name, dates, description, links, custom fields, percent_done, and lane info.
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
	}
}

// barManageTools returns bar mutation tool definitions.
func barManageTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "manage_bar",
			Description: manageBarDescription,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: manageBarProperties(),
				Required:   []string{"action"},
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
