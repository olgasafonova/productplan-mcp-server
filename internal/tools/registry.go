// Package tools provides ProductPlan tool handlers and registration for the MCP server.
package tools

import (
	"context"
	"encoding/json"

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

// createHandler returns the handler for a specific tool.
func createHandler(name string, cfg Config) mcp.Handler {
	switch name {
	// Roadmap handlers
	case "list_roadmaps":
		return listRoadmapsHandler(cfg.Client)
	case "get_roadmap":
		return getRoadmapHandler(cfg.Client)
	case "get_roadmap_bars":
		return getRoadmapBarsHandler(cfg.Client)
	case "get_roadmap_lanes":
		return getRoadmapLanesHandler(cfg.Client)
	case "get_roadmap_milestones":
		return getRoadmapMilestonesHandler(cfg.Client)
	case "manage_lane":
		return manageLaneHandler(cfg.Client)
	case "manage_milestone":
		return manageMilestoneHandler(cfg.Client)

	// Bar handlers
	case "get_bar":
		return getBarHandler(cfg.Client)
	case "get_bar_children":
		return getBarChildrenHandler(cfg.Client)
	case "get_bar_comments":
		return getBarCommentsHandler(cfg.Client)
	case "get_bar_connections":
		return getBarConnectionsHandler(cfg.Client)
	case "get_bar_links":
		return getBarLinksHandler(cfg.Client)
	case "manage_bar":
		return manageBarHandler(cfg.Client)
	case "manage_bar_comment":
		return manageBarCommentHandler(cfg.Client)
	case "manage_bar_connection":
		return manageBarConnectionHandler(cfg.Client)
	case "manage_bar_link":
		return manageBarLinkHandler(cfg.Client)

	// Objective handlers
	case "list_objectives":
		return listObjectivesHandler(cfg.Client)
	case "get_objective":
		return getObjectiveHandler(cfg.Client)
	case "list_key_results":
		return listKeyResultsHandler(cfg.Client)
	case "manage_objective":
		return manageObjectiveHandler(cfg.Client)
	case "manage_key_result":
		return manageKeyResultHandler(cfg.Client)

	// Idea handlers
	case "list_ideas":
		return listIdeasHandler(cfg.Client)
	case "get_idea":
		return getIdeaHandler(cfg.Client)
	case "get_idea_customers":
		return getIdeaCustomersHandler(cfg.Client)
	case "get_idea_tags":
		return getIdeaTagsHandler(cfg.Client)
	case "list_opportunities":
		return listOpportunitiesHandler(cfg.Client)
	case "get_opportunity":
		return getOpportunityHandler(cfg.Client)
	case "list_idea_forms":
		return listIdeaFormsHandler(cfg.Client)
	case "get_idea_form":
		return getIdeaFormHandler(cfg.Client)
	case "manage_idea":
		return manageIdeaHandler(cfg.Client)
	case "manage_idea_customer":
		return manageIdeaCustomerHandler(cfg.Client)
	case "manage_idea_tag":
		return manageIdeaTagHandler(cfg.Client)
	case "manage_opportunity":
		return manageOpportunityHandler(cfg.Client)

	// Launch handlers
	case "list_launches":
		return listLaunchesHandler(cfg.Client)
	case "get_launch":
		return getLaunchHandler(cfg.Client)

	// Utility handlers
	case "check_status":
		return checkStatusHandler(cfg.Client)
	case "health_check":
		return healthCheckHandler(cfg.HealthChecker)

	default:
		return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
			return nil, nil
		})
	}
}
