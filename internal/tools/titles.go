package tools

import "github.com/olgasafonova/productplan-mcp-server/internal/mcp"

// toolTitles holds the human-readable display name for every tool, sent as
// annotations.title. Titles are written out per tool rather than derived from
// the name, so a client shows "Get Complete Roadmap" instead of a mechanical
// "Get Roadmap Complete". TestEveryToolHasTitle fails if a tool is added
// without an entry here, or if an entry names a tool that no longer exists.
var toolTitles = map[string]string{
	// Roadmaps
	"list_roadmaps":          "List Roadmaps",
	"get_roadmap":            "Get Roadmap",
	"get_roadmap_bars":       "Get Roadmap Bars",
	"get_roadmap_lanes":      "Get Roadmap Lanes",
	"get_roadmap_legends":    "Get Roadmap Legends",
	"get_roadmap_milestones": "Get Roadmap Milestones",
	"get_roadmap_comments":   "Get Roadmap Comments",
	"get_roadmap_complete":   "Get Complete Roadmap",
	"manage_lane":            "Manage Lane",
	"manage_milestone":       "Manage Milestone",

	// Bars
	"get_bar":               "Get Bar",
	"get_bar_children":      "Get Bar Children",
	"get_bar_comments":      "Get Bar Comments",
	"get_bar_connections":   "Get Bar Connections",
	"get_bar_links":         "Get Bar Links",
	"manage_bar":            "Manage Bar",
	"manage_bar_connection": "Manage Bar Connection",
	"manage_bar_link":       "Manage Bar Link",
	"bulk_create_bars":      "Bulk Create Bars",
	"bulk_update_bars":      "Bulk Update Bars",
	"bulk_delete_bars":      "Bulk Delete Bars",

	// Objectives
	"list_objectives":   "List Objectives",
	"get_objective":     "Get Objective",
	"list_key_results":  "List Key Results",
	"get_key_result":    "Get Key Result",
	"manage_objective":  "Manage Objective",
	"manage_key_result": "Manage Key Result",

	// Ideas
	"list_ideas":         "List Ideas",
	"get_idea":           "Get Idea",
	"list_opportunities": "List Opportunities",
	"get_opportunity":    "Get Opportunity",
	"list_idea_forms":    "List Idea Forms",
	"get_idea_form":      "Get Idea Form",
	"list_all_customers": "List All Customers",
	"list_all_tags":      "List All Tags",
	"manage_idea":        "Manage Idea",
	"manage_opportunity": "Manage Opportunity",

	// Launches
	"list_launches":         "List Launches",
	"get_launch":            "Get Launch",
	"get_launch_sections":   "Get Launch Sections",
	"get_launch_section":    "Get Launch Section",
	"get_launch_tasks":      "Get Launch Tasks",
	"get_launch_task":       "Get Launch Task",
	"manage_launch":         "Manage Launch",
	"manage_launch_section": "Manage Launch Section",
	"manage_launch_task":    "Manage Launch Task",

	// Utility
	"check_status": "Check API Status",
	"health_check": "Server Health Check",
	"list_users":   "List Users",
	"list_teams":   "List Teams",
}

// applyTitle sets annotations.title from toolTitles when the definition does
// not already carry one. It runs after annotateTool, so Annotations is non-nil.
func applyTitle(tool *mcp.Tool) {
	if tool.Annotations == nil || tool.Annotations.Title != "" {
		return
	}
	tool.Annotations.Title = toolTitles[tool.Name]
}
