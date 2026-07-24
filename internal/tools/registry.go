// Package tools provides ProductPlan tool handlers and registration for the MCP server.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// HealthChecker defines the interface for health checking.
type HealthChecker interface {
	Check(ctx context.Context, deep bool) any
}

// Config holds dependencies for tool handlers.
type Config struct {
	Client        *api.Client
	HealthChecker HealthChecker
}

// RegisterAll registers all ProductPlan tools with the MCP registry.
func RegisterAll(registry *mcp.Registry, cfg Config) {
	// Register tool definitions
	for _, tool := range BuildAllTools() {
		handler := createHandler(tool.Name, cfg)
		registry.Register(tool, handler)
	}
}

// handlerConstructors maps each client-backed tool name to the function
// that builds its handler. Tools with other dependencies (health_check)
// are wired directly in createHandler.
var handlerConstructors = map[string]func(*api.Client) mcp.Handler{
	// Roadmap handlers
	"list_roadmaps":          listRoadmapsHandler,
	"get_roadmap":            getRoadmapHandler,
	"get_roadmap_bars":       getRoadmapBarsHandler,
	"get_roadmap_lanes":      getRoadmapLanesHandler,
	"get_roadmap_milestones": getRoadmapMilestonesHandler,
	"get_roadmap_legends":    getRoadmapLegendsHandler,
	"get_roadmap_comments":   getRoadmapCommentsHandler,
	"get_roadmap_complete":   getRoadmapCompleteHandler,
	"manage_lane":            manageLaneHandler,
	"manage_milestone":       manageMilestoneHandler,

	// Bar handlers
	"get_bar":               getBarHandler,
	"get_bar_children":      getBarChildrenHandler,
	"get_bar_comments":      getBarCommentsHandler,
	"get_bar_connections":   getBarConnectionsHandler,
	"get_bar_links":         getBarLinksHandler,
	"manage_bar":            manageBarHandler,
	"manage_bar_connection": manageBarConnectionHandler,
	"manage_bar_link":       manageBarLinkHandler,

	// Objective handlers
	"list_objectives":   listObjectivesHandler,
	"get_objective":     getObjectiveHandler,
	"list_key_results":  listKeyResultsHandler,
	"get_key_result":    getKeyResultHandler,
	"manage_objective":  manageObjectiveHandler,
	"manage_key_result": manageKeyResultHandler,

	// Idea handlers
	"list_ideas":         listIdeasHandler,
	"get_idea":           getIdeaHandler,
	"list_opportunities": listOpportunitiesHandler,
	"get_opportunity":    getOpportunityHandler,
	"list_idea_forms":    listIdeaFormsHandler,
	"get_idea_form":      getIdeaFormHandler,
	"list_all_customers": listAllCustomersHandler,
	"list_all_tags":      listAllTagsHandler,
	"manage_idea":        manageIdeaHandler,
	"manage_opportunity": manageOpportunityHandler,

	// Launch handlers
	"list_launches":         listLaunchesHandler,
	"get_launch":            getLaunchHandler,
	"manage_launch":         manageLaunchHandler,
	"get_launch_sections":   getLaunchSectionsHandler,
	"get_launch_section":    getLaunchSectionHandler,
	"manage_launch_section": manageLaunchSectionHandler,
	"get_launch_tasks":      getLaunchTasksHandler,
	"get_launch_task":       getLaunchTaskHandler,
	"manage_launch_task":    manageLaunchTaskHandler,

	// Utility handlers
	"check_status": checkStatusHandler,
	"list_users":   listUsersHandler,
	"list_teams":   listTeamsHandler,
}

// createHandler returns the handler for a specific tool.
func createHandler(name string, cfg Config) mcp.Handler {
	if name == "health_check" {
		return healthCheckHandler(cfg.HealthChecker)
	}
	if construct, ok := handlerConstructors[name]; ok {
		return construct(cfg.Client)
	}
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return nil, fmt.Errorf("unknown tool: %s", name)
	})
}
