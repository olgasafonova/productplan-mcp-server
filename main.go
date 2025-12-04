package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	apiBase = "https://app.productplan.com/api/v2"
	version = "2.0.0"
)

var apiToken string

// ============================================================================
// ProductPlan API Client
// ============================================================================

type APIClient struct {
	token      string
	httpClient *http.Client
}

func NewAPIClient(token string) *APIClient {
	return &APIClient{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *APIClient) request(method, endpoint string, body interface{}) (json.RawMessage, error) {
	url := apiBase + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	// For DELETE operations that return no content
	if resp.StatusCode == 204 {
		return json.RawMessage(`{"success": true}`), nil
	}

	return respBody, nil
}

// ============================================================================
// Roadmaps
// ============================================================================

func (c *APIClient) ListRoadmaps() (json.RawMessage, error) {
	return c.request("GET", "/roadmaps", nil)
}

func (c *APIClient) GetRoadmap(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id, nil)
}

func (c *APIClient) GetRoadmapBars(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id+"/bars", nil)
}

func (c *APIClient) GetRoadmapComments(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id+"/comments", nil)
}

// ============================================================================
// Lanes
// ============================================================================

func (c *APIClient) ListLanes(roadmapID string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+roadmapID+"/lanes", nil)
}

func (c *APIClient) CreateLane(roadmapID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/roadmaps/"+roadmapID+"/lanes", data)
}

func (c *APIClient) UpdateLane(roadmapID, laneID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/roadmaps/"+roadmapID+"/lanes/"+laneID, data)
}

func (c *APIClient) DeleteLane(roadmapID, laneID string) (json.RawMessage, error) {
	return c.request("DELETE", "/roadmaps/"+roadmapID+"/lanes/"+laneID, nil)
}

// ============================================================================
// Milestones
// ============================================================================

func (c *APIClient) ListMilestones(roadmapID string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+roadmapID+"/milestones", nil)
}

func (c *APIClient) CreateMilestone(roadmapID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/roadmaps/"+roadmapID+"/milestones", data)
}

func (c *APIClient) UpdateMilestone(roadmapID, milestoneID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/roadmaps/"+roadmapID+"/milestones/"+milestoneID, data)
}

func (c *APIClient) DeleteMilestone(roadmapID, milestoneID string) (json.RawMessage, error) {
	return c.request("DELETE", "/roadmaps/"+roadmapID+"/milestones/"+milestoneID, nil)
}

// ============================================================================
// Bars
// ============================================================================

func (c *APIClient) GetBar(id string) (json.RawMessage, error) {
	return c.request("GET", "/bars/"+id, nil)
}

func (c *APIClient) CreateBar(data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/bars", data)
}

func (c *APIClient) UpdateBar(id string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/bars/"+id, data)
}

func (c *APIClient) DeleteBar(id string) (json.RawMessage, error) {
	return c.request("DELETE", "/bars/"+id, nil)
}

func (c *APIClient) GetBarChildBars(id string) (json.RawMessage, error) {
	return c.request("GET", "/bars/"+id+"/child-bars", nil)
}

func (c *APIClient) GetBarComments(id string) (json.RawMessage, error) {
	return c.request("GET", "/bars/"+id+"/comments", nil)
}

// Bar Connections
func (c *APIClient) ListBarConnections(barID string) (json.RawMessage, error) {
	return c.request("GET", "/bars/"+barID+"/connections", nil)
}

func (c *APIClient) CreateBarConnection(barID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/bars/"+barID+"/connections", data)
}

func (c *APIClient) DeleteBarConnection(barID, connectionID string) (json.RawMessage, error) {
	return c.request("DELETE", "/bars/"+barID+"/connections/"+connectionID, nil)
}

// Bar Links
func (c *APIClient) ListBarLinks(barID string) (json.RawMessage, error) {
	return c.request("GET", "/bars/"+barID+"/links", nil)
}

func (c *APIClient) CreateBarLink(barID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/bars/"+barID+"/links", data)
}

func (c *APIClient) DeleteBarLink(barID, linkID string) (json.RawMessage, error) {
	return c.request("DELETE", "/bars/"+barID+"/links/"+linkID, nil)
}

// ============================================================================
// Discovery - Ideas
// ============================================================================

func (c *APIClient) ListIdeas() (json.RawMessage, error) {
	return c.request("GET", "/discovery/ideas", nil)
}

func (c *APIClient) GetIdea(id string) (json.RawMessage, error) {
	return c.request("GET", "/discovery/ideas/"+id, nil)
}

func (c *APIClient) CreateIdea(data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/discovery/ideas", data)
}

func (c *APIClient) UpdateIdea(id string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/discovery/ideas/"+id, data)
}

func (c *APIClient) ListIdeaCustomers() (json.RawMessage, error) {
	return c.request("GET", "/discovery/ideas/customers", nil)
}

func (c *APIClient) ListIdeaTags() (json.RawMessage, error) {
	return c.request("GET", "/discovery/ideas/tags", nil)
}

// ============================================================================
// Discovery - Opportunities
// ============================================================================

func (c *APIClient) ListOpportunities() (json.RawMessage, error) {
	return c.request("GET", "/discovery/opportunities", nil)
}

func (c *APIClient) GetOpportunity(id string) (json.RawMessage, error) {
	return c.request("GET", "/discovery/opportunities/"+id, nil)
}

func (c *APIClient) CreateOpportunity(data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/discovery/opportunities", data)
}

func (c *APIClient) UpdateOpportunity(id string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/discovery/opportunities/"+id, data)
}

// ============================================================================
// Discovery - Idea Forms
// ============================================================================

func (c *APIClient) ListIdeaForms() (json.RawMessage, error) {
	return c.request("GET", "/discovery/idea-forms", nil)
}

func (c *APIClient) GetIdeaForm(id string) (json.RawMessage, error) {
	return c.request("GET", "/discovery/idea-forms/"+id, nil)
}

// ============================================================================
// Strategy - Objectives
// ============================================================================

func (c *APIClient) ListObjectives() (json.RawMessage, error) {
	return c.request("GET", "/strategy/objectives", nil)
}

func (c *APIClient) GetObjective(id string) (json.RawMessage, error) {
	return c.request("GET", "/strategy/objectives/"+id, nil)
}

func (c *APIClient) CreateObjective(data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/strategy/objectives", data)
}

func (c *APIClient) UpdateObjective(id string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/strategy/objectives/"+id, data)
}

func (c *APIClient) DeleteObjective(id string) (json.RawMessage, error) {
	return c.request("DELETE", "/strategy/objectives/"+id, nil)
}

// ============================================================================
// Strategy - Key Results
// ============================================================================

func (c *APIClient) ListKeyResults(objectiveID string) (json.RawMessage, error) {
	return c.request("GET", "/strategy/objectives/"+objectiveID+"/key-results", nil)
}

func (c *APIClient) GetKeyResult(objectiveID, keyResultID string) (json.RawMessage, error) {
	return c.request("GET", "/strategy/objectives/"+objectiveID+"/key-results/"+keyResultID, nil)
}

func (c *APIClient) CreateKeyResult(objectiveID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/strategy/objectives/"+objectiveID+"/key-results", data)
}

func (c *APIClient) UpdateKeyResult(objectiveID, keyResultID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/strategy/objectives/"+objectiveID+"/key-results/"+keyResultID, data)
}

func (c *APIClient) DeleteKeyResult(objectiveID, keyResultID string) (json.RawMessage, error) {
	return c.request("DELETE", "/strategy/objectives/"+objectiveID+"/key-results/"+keyResultID, nil)
}

// ============================================================================
// Launches
// ============================================================================

func (c *APIClient) ListLaunches() (json.RawMessage, error) {
	return c.request("GET", "/launches", nil)
}

func (c *APIClient) GetLaunch(id string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+id, nil)
}

func (c *APIClient) CreateLaunch(data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/launches", data)
}

func (c *APIClient) UpdateLaunch(id string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/launches/"+id, data)
}

func (c *APIClient) DeleteLaunch(id string) (json.RawMessage, error) {
	return c.request("DELETE", "/launches/"+id, nil)
}

// ============================================================================
// Launches - Checklist Sections
// ============================================================================

func (c *APIClient) ListChecklistSections(launchID string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+launchID+"/checklist-sections", nil)
}

func (c *APIClient) GetChecklistSection(launchID, sectionID string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+launchID+"/checklist-sections/"+sectionID, nil)
}

func (c *APIClient) CreateChecklistSection(launchID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/launches/"+launchID+"/checklist-sections", data)
}

func (c *APIClient) UpdateChecklistSection(launchID, sectionID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/launches/"+launchID+"/checklist-sections/"+sectionID, data)
}

func (c *APIClient) DeleteChecklistSection(launchID, sectionID string) (json.RawMessage, error) {
	return c.request("DELETE", "/launches/"+launchID+"/checklist-sections/"+sectionID, nil)
}

// ============================================================================
// Launches - Tasks
// ============================================================================

func (c *APIClient) ListLaunchTasks(launchID string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+launchID+"/tasks", nil)
}

func (c *APIClient) GetLaunchTask(launchID, taskID string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+launchID+"/tasks/"+taskID, nil)
}

func (c *APIClient) CreateLaunchTask(launchID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/launches/"+launchID+"/tasks", data)
}

func (c *APIClient) UpdateLaunchTask(launchID, taskID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/launches/"+launchID+"/tasks/"+taskID, data)
}

func (c *APIClient) DeleteLaunchTask(launchID, taskID string) (json.RawMessage, error) {
	return c.request("DELETE", "/launches/"+launchID+"/tasks/"+taskID, nil)
}

// ============================================================================
// Administration
// ============================================================================

func (c *APIClient) ListUsers() (json.RawMessage, error) {
	return c.request("GET", "/users", nil)
}

func (c *APIClient) ListTeams() (json.RawMessage, error) {
	return c.request("GET", "/teams", nil)
}

func (c *APIClient) CheckStatus() (json.RawMessage, error) {
	return c.request("GET", "/status", nil)
}

// ============================================================================
// MCP Server Implementation
// ============================================================================

type MCPServer struct {
	client *APIClient
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewMCPServer(client *APIClient) *MCPServer {
	return &MCPServer{client: client}
}

func (s *MCPServer) getTools() []Tool {
	return []Tool{
		// Roadmaps
		{Name: "list_roadmaps", Description: "List all roadmaps", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_roadmap", Description: "Get roadmap details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Roadmap ID"}}, Required: []string{"id"}}},
		{Name: "get_roadmap_bars", Description: "Get all bars from a roadmap", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}}, Required: []string{"roadmap_id"}}},
		{Name: "get_roadmap_comments", Description: "Get comments on a roadmap", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}}, Required: []string{"roadmap_id"}}},

		// Lanes
		{Name: "list_lanes", Description: "List all lanes in a roadmap", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}}, Required: []string{"roadmap_id"}}},
		{Name: "create_lane", Description: "Create a new lane", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}, "name": {Type: "string", Description: "Lane name"}, "color": {Type: "string", Description: "Lane color (hex)"}}, Required: []string{"roadmap_id", "name"}}},
		{Name: "update_lane", Description: "Update a lane", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}, "lane_id": {Type: "string", Description: "Lane ID"}, "name": {Type: "string", Description: "Lane name"}, "color": {Type: "string", Description: "Lane color"}}, Required: []string{"roadmap_id", "lane_id"}}},
		{Name: "delete_lane", Description: "Delete a lane", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}, "lane_id": {Type: "string", Description: "Lane ID"}}, Required: []string{"roadmap_id", "lane_id"}}},

		// Milestones
		{Name: "list_milestones", Description: "List all milestones in a roadmap", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}}, Required: []string{"roadmap_id"}}},
		{Name: "create_milestone", Description: "Create a new milestone", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}, "name": {Type: "string", Description: "Milestone name"}, "date": {Type: "string", Description: "Date (YYYY-MM-DD)"}}, Required: []string{"roadmap_id", "name", "date"}}},
		{Name: "update_milestone", Description: "Update a milestone", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}, "milestone_id": {Type: "string", Description: "Milestone ID"}, "name": {Type: "string", Description: "Name"}, "date": {Type: "string", Description: "Date"}}, Required: []string{"roadmap_id", "milestone_id"}}},
		{Name: "delete_milestone", Description: "Delete a milestone", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}, "milestone_id": {Type: "string", Description: "Milestone ID"}}, Required: []string{"roadmap_id", "milestone_id"}}},

		// Bars
		{Name: "get_bar", Description: "Get bar details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Bar ID"}}, Required: []string{"id"}}},
		{Name: "create_bar", Description: "Create a new bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}, "lane_id": {Type: "string", Description: "Lane ID"}, "name": {Type: "string", Description: "Bar name"}, "start_date": {Type: "string", Description: "Start date (YYYY-MM-DD)"}, "end_date": {Type: "string", Description: "End date (YYYY-MM-DD)"}, "description": {Type: "string", Description: "Description"}}, Required: []string{"roadmap_id", "lane_id", "name"}}},
		{Name: "update_bar", Description: "Update a bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Bar ID"}, "name": {Type: "string", Description: "Name"}, "start_date": {Type: "string", Description: "Start date"}, "end_date": {Type: "string", Description: "End date"}, "description": {Type: "string", Description: "Description"}}, Required: []string{"id"}}},
		{Name: "delete_bar", Description: "Delete a bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Bar ID"}}, Required: []string{"id"}}},
		{Name: "get_bar_child_bars", Description: "Get child bars of a bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"bar_id": {Type: "string", Description: "Bar ID"}}, Required: []string{"bar_id"}}},
		{Name: "get_bar_comments", Description: "Get comments on a bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"bar_id": {Type: "string", Description: "Bar ID"}}, Required: []string{"bar_id"}}},

		// Bar Connections
		{Name: "list_bar_connections", Description: "List connections for a bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"bar_id": {Type: "string", Description: "Bar ID"}}, Required: []string{"bar_id"}}},
		{Name: "create_bar_connection", Description: "Create a connection between bars", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"bar_id": {Type: "string", Description: "Source Bar ID"}, "target_bar_id": {Type: "string", Description: "Target Bar ID"}}, Required: []string{"bar_id", "target_bar_id"}}},
		{Name: "delete_bar_connection", Description: "Delete a bar connection", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"bar_id": {Type: "string", Description: "Bar ID"}, "connection_id": {Type: "string", Description: "Connection ID"}}, Required: []string{"bar_id", "connection_id"}}},

		// Bar Links
		{Name: "list_bar_links", Description: "List external links for a bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"bar_id": {Type: "string", Description: "Bar ID"}}, Required: []string{"bar_id"}}},
		{Name: "create_bar_link", Description: "Create an external link on a bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"bar_id": {Type: "string", Description: "Bar ID"}, "url": {Type: "string", Description: "URL"}, "name": {Type: "string", Description: "Link name"}}, Required: []string{"bar_id", "url"}}},
		{Name: "delete_bar_link", Description: "Delete a bar link", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"bar_id": {Type: "string", Description: "Bar ID"}, "link_id": {Type: "string", Description: "Link ID"}}, Required: []string{"bar_id", "link_id"}}},

		// Ideas
		{Name: "list_ideas", Description: "List all ideas", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_idea", Description: "Get idea details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Idea ID"}}, Required: []string{"id"}}},
		{Name: "create_idea", Description: "Create a new idea", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"title": {Type: "string", Description: "Title"}, "description": {Type: "string", Description: "Description"}}, Required: []string{"title"}}},
		{Name: "update_idea", Description: "Update an idea", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Idea ID"}, "title": {Type: "string", Description: "Title"}, "description": {Type: "string", Description: "Description"}}, Required: []string{"id"}}},
		{Name: "list_idea_customers", Description: "List idea customers", InputSchema: InputSchema{Type: "object"}},
		{Name: "list_idea_tags", Description: "List idea tags", InputSchema: InputSchema{Type: "object"}},

		// Opportunities
		{Name: "list_opportunities", Description: "List all opportunities", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_opportunity", Description: "Get opportunity details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Opportunity ID"}}, Required: []string{"id"}}},
		{Name: "create_opportunity", Description: "Create a new opportunity", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"name": {Type: "string", Description: "Name"}, "description": {Type: "string", Description: "Description"}}, Required: []string{"name"}}},
		{Name: "update_opportunity", Description: "Update an opportunity", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Opportunity ID"}, "name": {Type: "string", Description: "Name"}, "description": {Type: "string", Description: "Description"}}, Required: []string{"id"}}},

		// Idea Forms
		{Name: "list_idea_forms", Description: "List idea forms", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_idea_form", Description: "Get idea form details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Form ID"}}, Required: []string{"id"}}},

		// Objectives
		{Name: "list_objectives", Description: "List all objectives (OKRs)", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_objective", Description: "Get objective details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Objective ID"}}, Required: []string{"id"}}},
		{Name: "create_objective", Description: "Create a new objective", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"name": {Type: "string", Description: "Name"}, "description": {Type: "string", Description: "Description"}, "time_frame": {Type: "string", Description: "Time frame"}}, Required: []string{"name"}}},
		{Name: "update_objective", Description: "Update an objective", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Objective ID"}, "name": {Type: "string", Description: "Name"}, "description": {Type: "string", Description: "Description"}}, Required: []string{"id"}}},
		{Name: "delete_objective", Description: "Delete an objective", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Objective ID"}}, Required: []string{"id"}}},

		// Key Results
		{Name: "list_key_results", Description: "List key results for an objective", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"objective_id": {Type: "string", Description: "Objective ID"}}, Required: []string{"objective_id"}}},
		{Name: "get_key_result", Description: "Get key result details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"objective_id": {Type: "string", Description: "Objective ID"}, "key_result_id": {Type: "string", Description: "Key Result ID"}}, Required: []string{"objective_id", "key_result_id"}}},
		{Name: "create_key_result", Description: "Create a new key result", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"objective_id": {Type: "string", Description: "Objective ID"}, "name": {Type: "string", Description: "Name"}, "target_value": {Type: "string", Description: "Target value"}, "current_value": {Type: "string", Description: "Current value"}}, Required: []string{"objective_id", "name"}}},
		{Name: "update_key_result", Description: "Update a key result", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"objective_id": {Type: "string", Description: "Objective ID"}, "key_result_id": {Type: "string", Description: "Key Result ID"}, "name": {Type: "string", Description: "Name"}, "current_value": {Type: "string", Description: "Current value"}}, Required: []string{"objective_id", "key_result_id"}}},
		{Name: "delete_key_result", Description: "Delete a key result", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"objective_id": {Type: "string", Description: "Objective ID"}, "key_result_id": {Type: "string", Description: "Key Result ID"}}, Required: []string{"objective_id", "key_result_id"}}},

		// Launches
		{Name: "list_launches", Description: "List all launches", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_launch", Description: "Get launch details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Launch ID"}}, Required: []string{"id"}}},
		{Name: "create_launch", Description: "Create a new launch", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"name": {Type: "string", Description: "Name"}, "date": {Type: "string", Description: "Launch date (YYYY-MM-DD)"}}, Required: []string{"name"}}},
		{Name: "update_launch", Description: "Update a launch", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Launch ID"}, "name": {Type: "string", Description: "Name"}, "date": {Type: "string", Description: "Date"}}, Required: []string{"id"}}},
		{Name: "delete_launch", Description: "Delete a launch", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Launch ID"}}, Required: []string{"id"}}},

		// Checklist Sections
		{Name: "list_checklist_sections", Description: "List checklist sections for a launch", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}}, Required: []string{"launch_id"}}},
		{Name: "get_checklist_section", Description: "Get checklist section details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}, "section_id": {Type: "string", Description: "Section ID"}}, Required: []string{"launch_id", "section_id"}}},
		{Name: "create_checklist_section", Description: "Create a checklist section", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}, "name": {Type: "string", Description: "Section name"}}, Required: []string{"launch_id", "name"}}},
		{Name: "update_checklist_section", Description: "Update a checklist section", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}, "section_id": {Type: "string", Description: "Section ID"}, "name": {Type: "string", Description: "Name"}}, Required: []string{"launch_id", "section_id"}}},
		{Name: "delete_checklist_section", Description: "Delete a checklist section", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}, "section_id": {Type: "string", Description: "Section ID"}}, Required: []string{"launch_id", "section_id"}}},

		// Launch Tasks
		{Name: "list_launch_tasks", Description: "List tasks for a launch", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}}, Required: []string{"launch_id"}}},
		{Name: "get_launch_task", Description: "Get launch task details", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}, "task_id": {Type: "string", Description: "Task ID"}}, Required: []string{"launch_id", "task_id"}}},
		{Name: "create_launch_task", Description: "Create a launch task", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}, "name": {Type: "string", Description: "Task name"}, "section_id": {Type: "string", Description: "Checklist section ID"}}, Required: []string{"launch_id", "name"}}},
		{Name: "update_launch_task", Description: "Update a launch task", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}, "task_id": {Type: "string", Description: "Task ID"}, "name": {Type: "string", Description: "Name"}, "completed": {Type: "string", Description: "Completed (true/false)"}}, Required: []string{"launch_id", "task_id"}}},
		{Name: "delete_launch_task", Description: "Delete a launch task", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}, "task_id": {Type: "string", Description: "Task ID"}}, Required: []string{"launch_id", "task_id"}}},

		// Administration
		{Name: "list_users", Description: "List all users", InputSchema: InputSchema{Type: "object"}},
		{Name: "list_teams", Description: "List all teams", InputSchema: InputSchema{Type: "object"}},
		{Name: "check_status", Description: "Check API status", InputSchema: InputSchema{Type: "object"}},
	}
}

func (s *MCPServer) handleToolCall(name string, args map[string]interface{}) (json.RawMessage, error) {
	getString := func(key string) string {
		if v, ok := args[key].(string); ok {
			return v
		}
		return ""
	}

	// Helper to remove specific keys from args for update operations
	removeKeys := func(keys ...string) {
		for _, k := range keys {
			delete(args, k)
		}
	}

	switch name {
	// Roadmaps
	case "list_roadmaps":
		return s.client.ListRoadmaps()
	case "get_roadmap":
		return s.client.GetRoadmap(getString("id"))
	case "get_roadmap_bars":
		return s.client.GetRoadmapBars(getString("roadmap_id"))
	case "get_roadmap_comments":
		return s.client.GetRoadmapComments(getString("roadmap_id"))

	// Lanes
	case "list_lanes":
		return s.client.ListLanes(getString("roadmap_id"))
	case "create_lane":
		roadmapID := getString("roadmap_id")
		removeKeys("roadmap_id")
		return s.client.CreateLane(roadmapID, args)
	case "update_lane":
		roadmapID, laneID := getString("roadmap_id"), getString("lane_id")
		removeKeys("roadmap_id", "lane_id")
		return s.client.UpdateLane(roadmapID, laneID, args)
	case "delete_lane":
		return s.client.DeleteLane(getString("roadmap_id"), getString("lane_id"))

	// Milestones
	case "list_milestones":
		return s.client.ListMilestones(getString("roadmap_id"))
	case "create_milestone":
		roadmapID := getString("roadmap_id")
		removeKeys("roadmap_id")
		return s.client.CreateMilestone(roadmapID, args)
	case "update_milestone":
		roadmapID, milestoneID := getString("roadmap_id"), getString("milestone_id")
		removeKeys("roadmap_id", "milestone_id")
		return s.client.UpdateMilestone(roadmapID, milestoneID, args)
	case "delete_milestone":
		return s.client.DeleteMilestone(getString("roadmap_id"), getString("milestone_id"))

	// Bars
	case "get_bar":
		return s.client.GetBar(getString("id"))
	case "create_bar":
		return s.client.CreateBar(args)
	case "update_bar":
		id := getString("id")
		removeKeys("id")
		return s.client.UpdateBar(id, args)
	case "delete_bar":
		return s.client.DeleteBar(getString("id"))
	case "get_bar_child_bars":
		return s.client.GetBarChildBars(getString("bar_id"))
	case "get_bar_comments":
		return s.client.GetBarComments(getString("bar_id"))

	// Bar Connections
	case "list_bar_connections":
		return s.client.ListBarConnections(getString("bar_id"))
	case "create_bar_connection":
		barID := getString("bar_id")
		removeKeys("bar_id")
		return s.client.CreateBarConnection(barID, args)
	case "delete_bar_connection":
		return s.client.DeleteBarConnection(getString("bar_id"), getString("connection_id"))

	// Bar Links
	case "list_bar_links":
		return s.client.ListBarLinks(getString("bar_id"))
	case "create_bar_link":
		barID := getString("bar_id")
		removeKeys("bar_id")
		return s.client.CreateBarLink(barID, args)
	case "delete_bar_link":
		return s.client.DeleteBarLink(getString("bar_id"), getString("link_id"))

	// Ideas
	case "list_ideas":
		return s.client.ListIdeas()
	case "get_idea":
		return s.client.GetIdea(getString("id"))
	case "create_idea":
		return s.client.CreateIdea(args)
	case "update_idea":
		id := getString("id")
		removeKeys("id")
		return s.client.UpdateIdea(id, args)
	case "list_idea_customers":
		return s.client.ListIdeaCustomers()
	case "list_idea_tags":
		return s.client.ListIdeaTags()

	// Opportunities
	case "list_opportunities":
		return s.client.ListOpportunities()
	case "get_opportunity":
		return s.client.GetOpportunity(getString("id"))
	case "create_opportunity":
		return s.client.CreateOpportunity(args)
	case "update_opportunity":
		id := getString("id")
		removeKeys("id")
		return s.client.UpdateOpportunity(id, args)

	// Idea Forms
	case "list_idea_forms":
		return s.client.ListIdeaForms()
	case "get_idea_form":
		return s.client.GetIdeaForm(getString("id"))

	// Objectives
	case "list_objectives":
		return s.client.ListObjectives()
	case "get_objective":
		return s.client.GetObjective(getString("id"))
	case "create_objective":
		return s.client.CreateObjective(args)
	case "update_objective":
		id := getString("id")
		removeKeys("id")
		return s.client.UpdateObjective(id, args)
	case "delete_objective":
		return s.client.DeleteObjective(getString("id"))

	// Key Results
	case "list_key_results":
		return s.client.ListKeyResults(getString("objective_id"))
	case "get_key_result":
		return s.client.GetKeyResult(getString("objective_id"), getString("key_result_id"))
	case "create_key_result":
		objectiveID := getString("objective_id")
		removeKeys("objective_id")
		return s.client.CreateKeyResult(objectiveID, args)
	case "update_key_result":
		objectiveID, keyResultID := getString("objective_id"), getString("key_result_id")
		removeKeys("objective_id", "key_result_id")
		return s.client.UpdateKeyResult(objectiveID, keyResultID, args)
	case "delete_key_result":
		return s.client.DeleteKeyResult(getString("objective_id"), getString("key_result_id"))

	// Launches
	case "list_launches":
		return s.client.ListLaunches()
	case "get_launch":
		return s.client.GetLaunch(getString("id"))
	case "create_launch":
		return s.client.CreateLaunch(args)
	case "update_launch":
		id := getString("id")
		removeKeys("id")
		return s.client.UpdateLaunch(id, args)
	case "delete_launch":
		return s.client.DeleteLaunch(getString("id"))

	// Checklist Sections
	case "list_checklist_sections":
		return s.client.ListChecklistSections(getString("launch_id"))
	case "get_checklist_section":
		return s.client.GetChecklistSection(getString("launch_id"), getString("section_id"))
	case "create_checklist_section":
		launchID := getString("launch_id")
		removeKeys("launch_id")
		return s.client.CreateChecklistSection(launchID, args)
	case "update_checklist_section":
		launchID, sectionID := getString("launch_id"), getString("section_id")
		removeKeys("launch_id", "section_id")
		return s.client.UpdateChecklistSection(launchID, sectionID, args)
	case "delete_checklist_section":
		return s.client.DeleteChecklistSection(getString("launch_id"), getString("section_id"))

	// Launch Tasks
	case "list_launch_tasks":
		return s.client.ListLaunchTasks(getString("launch_id"))
	case "get_launch_task":
		return s.client.GetLaunchTask(getString("launch_id"), getString("task_id"))
	case "create_launch_task":
		launchID := getString("launch_id")
		removeKeys("launch_id")
		return s.client.CreateLaunchTask(launchID, args)
	case "update_launch_task":
		launchID, taskID := getString("launch_id"), getString("task_id")
		removeKeys("launch_id", "task_id")
		return s.client.UpdateLaunchTask(launchID, taskID, args)
	case "delete_launch_task":
		return s.client.DeleteLaunchTask(getString("launch_id"), getString("task_id"))

	// Administration
	case "list_users":
		return s.client.ListUsers()
	case "list_teams":
		return s.client.ListTeams()
	case "check_status":
		return s.client.CheckStatus()

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *MCPServer) handleRequest(req JSONRPCRequest) JSONRPCResponse {
	resp := JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo":      map[string]string{"name": "productplan-mcp-server", "version": version},
			"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
		}

	case "notifications/initialized":
		return JSONRPCResponse{}

	case "tools/list":
		resp.Result = map[string]interface{}{"tools": s.getTools()}

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &RPCError{Code: -32602, Message: err.Error()}
			return resp
		}

		result, err := s.handleToolCall(params.Name, params.Arguments)
		if err != nil {
			resp.Result = map[string]interface{}{
				"content": []ToolContent{{Type: "text", Text: "Error: " + err.Error()}},
				"isError": true,
			}
		} else {
			var pretty bytes.Buffer
			json.Indent(&pretty, result, "", "  ")
			resp.Result = map[string]interface{}{
				"content": []ToolContent{{Type: "text", Text: pretty.String()}},
			}
		}

	default:
		resp.Error = &RPCError{Code: -32601, Message: "Method not found: " + req.Method}
	}

	return resp
}

func (s *MCPServer) Run() {
	fmt.Fprintln(os.Stderr, "ProductPlan MCP Server v"+version+" running on stdio")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		resp := s.handleRequest(req)
		if resp.JSONRPC == "" {
			continue
		}

		respJSON, _ := json.Marshal(resp)
		fmt.Println(string(respJSON))
	}
}

// ============================================================================
// CLI Implementation
// ============================================================================

func printJSON(data json.RawMessage) {
	var pretty bytes.Buffer
	json.Indent(&pretty, data, "", "  ")
	fmt.Println(pretty.String())
}

func printTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Repeat("-", len(strings.Join(headers, "  "))))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

func runCLI(args []string) {
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewAPIClient(apiToken)
	cmd := args[0]
	subArgs := args[1:]

	var result json.RawMessage
	var err error

	switch cmd {
	case "roadmaps":
		if len(subArgs) == 0 {
			result, err = client.ListRoadmaps()
			if err == nil {
				printRoadmapsList(result)
				return
			}
		} else {
			result, err = client.GetRoadmap(subArgs[0])
		}

	case "bars":
		if len(subArgs) == 0 {
			fmt.Println("Usage: productplan bars <roadmap_id>")
			os.Exit(1)
		}
		result, err = client.GetRoadmapBars(subArgs[0])

	case "lanes":
		if len(subArgs) == 0 {
			fmt.Println("Usage: productplan lanes <roadmap_id>")
			os.Exit(1)
		}
		result, err = client.ListLanes(subArgs[0])

	case "milestones":
		if len(subArgs) == 0 {
			fmt.Println("Usage: productplan milestones <roadmap_id>")
			os.Exit(1)
		}
		result, err = client.ListMilestones(subArgs[0])

	case "objectives":
		if len(subArgs) == 0 {
			result, err = client.ListObjectives()
			if err == nil {
				printObjectivesList(result)
				return
			}
		} else {
			result, err = client.GetObjective(subArgs[0])
		}

	case "key-results":
		if len(subArgs) == 0 {
			fmt.Println("Usage: productplan key-results <objective_id>")
			os.Exit(1)
		}
		result, err = client.ListKeyResults(subArgs[0])

	case "ideas":
		if len(subArgs) == 0 {
			result, err = client.ListIdeas()
		} else {
			result, err = client.GetIdea(subArgs[0])
		}

	case "opportunities":
		if len(subArgs) == 0 {
			result, err = client.ListOpportunities()
		} else {
			result, err = client.GetOpportunity(subArgs[0])
		}

	case "launches":
		if len(subArgs) == 0 {
			result, err = client.ListLaunches()
		} else {
			result, err = client.GetLaunch(subArgs[0])
		}

	case "tasks":
		if len(subArgs) == 0 {
			fmt.Println("Usage: productplan tasks <launch_id>")
			os.Exit(1)
		}
		result, err = client.ListLaunchTasks(subArgs[0])

	case "users":
		result, err = client.ListUsers()

	case "teams":
		result, err = client.ListTeams()

	case "status":
		result, err = client.CheckStatus()
		if err == nil {
			printStatus(result)
			return
		}

	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(result)
}

func printRoadmapsList(data json.RawMessage) {
	var roadmaps []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &roadmaps); err != nil {
		printJSON(data)
		return
	}

	headers := []string{"ID", "NAME"}
	var rows [][]string
	for _, r := range roadmaps {
		rows = append(rows, []string{fmt.Sprintf("%d", r.ID), r.Name})
	}
	printTable(headers, rows)
}

func printObjectivesList(data json.RawMessage) {
	var objectives []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(data, &objectives); err != nil {
		printJSON(data)
		return
	}

	headers := []string{"ID", "STATUS", "NAME"}
	var rows [][]string
	for _, o := range objectives {
		status := o.Status
		if status == "" {
			status = "-"
		}
		rows = append(rows, []string{fmt.Sprintf("%d", o.ID), status, o.Name})
	}
	printTable(headers, rows)
}

func printStatus(data json.RawMessage) {
	var status struct {
		API      string `json:"api"`
		Database string `json:"database"`
		User     struct {
			Name string `json:"name"`
		} `json:"user"`
	}
	if err := json.Unmarshal(data, &status); err != nil {
		printJSON(data)
		return
	}

	fmt.Printf("API:      %s\n", status.API)
	fmt.Printf("Database: %s\n", status.Database)
	fmt.Printf("User:     %s\n", status.User.Name)
}

func printUsage() {
	fmt.Printf(`ProductPlan CLI & MCP Server v%s

Usage:
  productplan <command> [arguments]
  productplan serve                    Start MCP server (for AI assistants)

Commands:
  roadmaps [id]                        List roadmaps or get details
  bars <roadmap_id>                    List bars in a roadmap
  lanes <roadmap_id>                   List lanes in a roadmap
  milestones <roadmap_id>              List milestones in a roadmap
  objectives [id]                      List objectives or get details
  key-results <objective_id>           List key results for an objective
  ideas [id]                           List ideas or get details
  opportunities [id]                   List opportunities or get details
  launches [id]                        List launches or get details
  tasks <launch_id>                    List tasks for a launch
  users                                List users
  teams                                List teams
  status                               Check API status

Environment:
  PRODUCTPLAN_API_TOKEN                Your ProductPlan API token (required)

Examples:
  productplan status                   Check connection
  productplan roadmaps                 List all roadmaps
  productplan objectives               List all OKRs
  productplan bars 12345               List bars in roadmap 12345

MCP Server (for Claude Code, Cursor, etc.):
  productplan serve

`, version)
}

// ============================================================================
// Main
// ============================================================================

func main() {
	apiToken = os.Getenv("PRODUCTPLAN_API_TOKEN")
	if apiToken == "" {
		fmt.Fprintln(os.Stderr, "Error: PRODUCTPLAN_API_TOKEN environment variable is required")
		os.Exit(1)
	}

	args := os.Args[1:]

	// No args or "serve" = MCP server mode
	if len(args) == 0 || args[0] == "serve" || args[0] == "mcp" {
		client := NewAPIClient(apiToken)
		server := NewMCPServer(client)
		server.Run()
		return
	}

	// Help
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printUsage()
		return
	}

	// CLI mode
	runCLI(args)
}
