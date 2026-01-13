package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

func setupTestServer(t *testing.T, response any) (*httptest.Server, *api.Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	client, err := api.New(api.Config{
		Token:   "test-token",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return server, client
}

func TestRoadmapHandlers(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "123", "name": "Test"})
	defer server.Close()

	tests := []struct {
		name    string
		handler func(*api.Client) Handler
		args    map[string]any
	}{
		{
			"listRoadmapsHandler",
			func(c *api.Client) Handler { return listRoadmapsHandler(c) },
			nil,
		},
		{
			"getRoadmapHandler",
			func(c *api.Client) Handler { return getRoadmapHandler(c) },
			map[string]any{"roadmap_id": "123"},
		},
		{
			"getRoadmapBarsHandler",
			func(c *api.Client) Handler { return getRoadmapBarsHandler(c) },
			map[string]any{"roadmap_id": "123"},
		},
		{
			"getRoadmapLanesHandler",
			func(c *api.Client) Handler { return getRoadmapLanesHandler(c) },
			map[string]any{"roadmap_id": "123"},
		},
		{
			"getRoadmapMilestonesHandler",
			func(c *api.Client) Handler { return getRoadmapMilestonesHandler(c) },
			map[string]any{"roadmap_id": "123"},
		},
		{
			"getRoadmapLegendsHandler",
			func(c *api.Client) Handler { return getRoadmapLegendsHandler(c) },
			map[string]any{"roadmap_id": "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.handler(client)
			result, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestRoadmapHandlersMissingRequired(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{})
	defer server.Close()

	tests := []struct {
		name    string
		handler func(*api.Client) Handler
	}{
		{"getRoadmapHandler", func(c *api.Client) Handler { return getRoadmapHandler(c) }},
		{"getRoadmapBarsHandler", func(c *api.Client) Handler { return getRoadmapBarsHandler(c) }},
		{"getRoadmapLanesHandler", func(c *api.Client) Handler { return getRoadmapLanesHandler(c) }},
		{"getRoadmapMilestonesHandler", func(c *api.Client) Handler { return getRoadmapMilestonesHandler(c) }},
		{"getRoadmapLegendsHandler", func(c *api.Client) Handler { return getRoadmapLegendsHandler(c) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.handler(client)
			_, err := handler.Handle(context.Background(), nil)
			if err == nil {
				t.Error("expected error for missing required param")
			}
		})
	}
}

func TestManageLaneHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "lane-1"})
	defer server.Close()

	handler := manageLaneHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{
			"create",
			map[string]any{"action": "create", "roadmap_id": "123", "name": "New Lane", "color": "#FF0000"},
		},
		{
			"update",
			map[string]any{"action": "update", "roadmap_id": "123", "lane_id": "456", "name": "Updated"},
		},
		{
			"delete",
			map[string]any{"action": "delete", "roadmap_id": "123", "lane_id": "456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestManageLaneHandlerMissingRequired(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{})
	defer server.Close()

	handler := manageLaneHandler(client)

	// Missing action
	_, err := handler.Handle(context.Background(), map[string]any{"roadmap_id": "123"})
	if err == nil {
		t.Error("expected error for missing action")
	}

	// Missing roadmap_id
	_, err = handler.Handle(context.Background(), map[string]any{"action": "create"})
	if err == nil {
		t.Error("expected error for missing roadmap_id")
	}
}

func TestManageMilestoneHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "milestone-1"})
	defer server.Close()

	handler := manageMilestoneHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{
			"create",
			map[string]any{"action": "create", "roadmap_id": "123", "name": "Launch", "date": "2024-06-01"},
		},
		{
			"update",
			map[string]any{"action": "update", "roadmap_id": "123", "milestone_id": "456", "name": "Updated"},
		},
		{
			"delete",
			map[string]any{"action": "delete", "roadmap_id": "123", "milestone_id": "456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBarHandlers(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "bar-1"})
	defer server.Close()

	tests := []struct {
		name    string
		handler func(*api.Client) Handler
		args    map[string]any
	}{
		{"getBarHandler", func(c *api.Client) Handler { return getBarHandler(c) }, map[string]any{"bar_id": "123"}},
		{"getBarChildrenHandler", func(c *api.Client) Handler { return getBarChildrenHandler(c) }, map[string]any{"bar_id": "123"}},
		{"getBarCommentsHandler", func(c *api.Client) Handler { return getBarCommentsHandler(c) }, map[string]any{"bar_id": "123"}},
		{"getBarConnectionsHandler", func(c *api.Client) Handler { return getBarConnectionsHandler(c) }, map[string]any{"bar_id": "123"}},
		{"getBarLinksHandler", func(c *api.Client) Handler { return getBarLinksHandler(c) }, map[string]any{"bar_id": "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.handler(client)
			result, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestManageBarHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "bar-1"})
	defer server.Close()

	handler := manageBarHandler(client)

	percentDone := 50
	container := true
	parked := false
	effort := 5

	tests := []struct {
		name string
		args map[string]any
	}{
		{
			"create",
			map[string]any{
				"action": "create", "roadmap_id": "123", "lane_id": "456",
				"name": "New Bar", "start_date": "2024-01-01", "end_date": "2024-03-01",
				"description": "Test bar",
			},
		},
		{
			"create_with_legend",
			map[string]any{
				"action": "create", "roadmap_id": "123", "lane_id": "456",
				"name": "Bar with Color", "legend_id": "legend-1",
			},
		},
		{
			"create_with_all_fields",
			map[string]any{
				"action": "create", "roadmap_id": "123", "lane_id": "456",
				"name": "Full Bar", "legend_id": "legend-1",
				"percent_done": percentDone, "container": container, "parked": parked,
				"parent_id": "parent-123", "strategic_value": "High priority",
				"notes": "Important notes", "effort": effort,
			},
		},
		{
			"update",
			map[string]any{"action": "update", "bar_id": "789", "name": "Updated Bar"},
		},
		{
			"update_legend",
			map[string]any{"action": "update", "bar_id": "789", "legend_id": "legend-2"},
		},
		{
			"update_percent_done",
			map[string]any{"action": "update", "bar_id": "789", "percent_done": percentDone},
		},
		{
			"delete",
			map[string]any{"action": "delete", "bar_id": "789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestManageBarCommentHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "comment-1"})
	defer server.Close()

	handler := manageBarCommentHandler(client)

	_, err := handler.Handle(context.Background(), map[string]any{
		"bar_id": "123",
		"body":   "This is a test comment",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Missing required params
	_, err = handler.Handle(context.Background(), map[string]any{"bar_id": "123"})
	if err == nil {
		t.Error("expected error for missing body")
	}

	_, err = handler.Handle(context.Background(), map[string]any{"body": "test"})
	if err == nil {
		t.Error("expected error for missing bar_id")
	}
}

func TestManageBarConnectionHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "conn-1"})
	defer server.Close()

	handler := manageBarConnectionHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"create", map[string]any{"action": "create", "bar_id": "123", "target_bar_id": "456"}},
		{"delete", map[string]any{"action": "delete", "bar_id": "123", "connection_id": "789"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestManageBarLinkHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "link-1"})
	defer server.Close()

	handler := manageBarLinkHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"create", map[string]any{"action": "create", "bar_id": "123", "url": "https://example.com", "name": "Example"}},
		{"update", map[string]any{"action": "update", "bar_id": "123", "link_id": "456", "url": "https://updated.com"}},
		{"delete", map[string]any{"action": "delete", "bar_id": "123", "link_id": "456"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestObjectiveHandlers(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "obj-1"})
	defer server.Close()

	tests := []struct {
		name    string
		handler func(*api.Client) Handler
		args    map[string]any
	}{
		{"listObjectivesHandler", func(c *api.Client) Handler { return listObjectivesHandler(c) }, nil},
		{"getObjectiveHandler", func(c *api.Client) Handler { return getObjectiveHandler(c) }, map[string]any{"objective_id": "123"}},
		{"listKeyResultsHandler", func(c *api.Client) Handler { return listKeyResultsHandler(c) }, map[string]any{"objective_id": "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.handler(client)
			result, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestManageObjectiveHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "obj-1"})
	defer server.Close()

	handler := manageObjectiveHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"create", map[string]any{"action": "create", "name": "New Objective", "description": "Test", "time_frame": "Q1 2024"}},
		{"update", map[string]any{"action": "update", "objective_id": "123", "name": "Updated"}},
		{"delete", map[string]any{"action": "delete", "objective_id": "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestManageKeyResultHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "kr-1"})
	defer server.Close()

	handler := manageKeyResultHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"create", map[string]any{"action": "create", "objective_id": "123", "name": "New KR", "target_value": "100", "current_value": "0"}},
		{"update", map[string]any{"action": "update", "objective_id": "123", "key_result_id": "456", "current_value": "50"}},
		{"delete", map[string]any{"action": "delete", "objective_id": "123", "key_result_id": "456"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIdeaHandlers(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "idea-1"})
	defer server.Close()

	tests := []struct {
		name    string
		handler func(*api.Client) Handler
		args    map[string]any
	}{
		{"listIdeasHandler", func(c *api.Client) Handler { return listIdeasHandler(c) }, nil},
		{"getIdeaHandler", func(c *api.Client) Handler { return getIdeaHandler(c) }, map[string]any{"idea_id": "123"}},
		{"getIdeaCustomersHandler", func(c *api.Client) Handler { return getIdeaCustomersHandler(c) }, map[string]any{"idea_id": "123"}},
		{"getIdeaTagsHandler", func(c *api.Client) Handler { return getIdeaTagsHandler(c) }, map[string]any{"idea_id": "123"}},
		{"listOpportunitiesHandler", func(c *api.Client) Handler { return listOpportunitiesHandler(c) }, nil},
		{"getOpportunityHandler", func(c *api.Client) Handler { return getOpportunityHandler(c) }, map[string]any{"opportunity_id": "123"}},
		{"listIdeaFormsHandler", func(c *api.Client) Handler { return listIdeaFormsHandler(c) }, nil},
		{"getIdeaFormHandler", func(c *api.Client) Handler { return getIdeaFormHandler(c) }, map[string]any{"form_id": "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.handler(client)
			result, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestManageIdeaHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "idea-1"})
	defer server.Close()

	handler := manageIdeaHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"create", map[string]any{"action": "create", "title": "New Idea", "description": "Test", "status": "new"}},
		{"update", map[string]any{"action": "update", "idea_id": "123", "title": "Updated", "description": "Updated desc", "status": "active"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestManageIdeaCustomerHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "customer-1"})
	defer server.Close()

	handler := manageIdeaCustomerHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"add", map[string]any{"action": "add", "idea_id": "123", "customer_id": "456"}},
		{"remove", map[string]any{"action": "remove", "idea_id": "123", "customer_id": "456"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestManageIdeaTagHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "tag-1"})
	defer server.Close()

	handler := manageIdeaTagHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"add", map[string]any{"action": "add", "idea_id": "123", "tag_id": "456"}},
		{"remove", map[string]any{"action": "remove", "idea_id": "123", "tag_id": "456"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestManageOpportunityHandler(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "opp-1"})
	defer server.Close()

	handler := manageOpportunityHandler(client)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"create", map[string]any{"action": "create", "problem_statement": "Test Problem", "description": "Desc", "workflow_status": "draft"}},
		{"update", map[string]any{"action": "update", "opportunity_id": "123", "problem_statement": "Updated"}},
		{"delete", map[string]any{"action": "delete", "opportunity_id": "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLaunchHandlers(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"id": "launch-1"})
	defer server.Close()

	tests := []struct {
		name    string
		handler func(*api.Client) Handler
		args    map[string]any
	}{
		{"listLaunchesHandler", func(c *api.Client) Handler { return listLaunchesHandler(c) }, nil},
		{"getLaunchHandler", func(c *api.Client) Handler { return getLaunchHandler(c) }, map[string]any{"launch_id": "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.handler(client)
			result, err := handler.Handle(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestUtilityHandlers(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{"status": "ok"})
	defer server.Close()

	handler := checkStatusHandler(client)
	result, err := handler.Handle(context.Background(), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestHandlerMissingRequiredParams(t *testing.T) {
	server, client := setupTestServer(t, map[string]any{})
	defer server.Close()

	tests := []struct {
		name    string
		handler Handler
		errPart string
	}{
		{"getBarHandler", getBarHandler(client), "bar_id"},
		{"getObjectiveHandler", getObjectiveHandler(client), "objective_id"},
		{"getIdeaHandler", getIdeaHandler(client), "idea_id"},
		{"getLaunchHandler", getLaunchHandler(client), "launch_id"},
		{"getOpportunityHandler", getOpportunityHandler(client), "opportunity_id"},
		{"getIdeaFormHandler", getIdeaFormHandler(client), "form_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.handler.Handle(context.Background(), nil)
			if err == nil {
				t.Error("expected error for missing required param")
			}
			if !strings.Contains(err.Error(), tt.errPart) {
				t.Errorf("expected error to mention %q, got %v", tt.errPart, err)
			}
		})
	}
}

// Handler is an alias for mcp.Handler for internal test usage
type Handler = interface {
	Handle(ctx context.Context, args map[string]any) (json.RawMessage, error)
}
