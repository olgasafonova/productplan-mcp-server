package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	apiBase        = "https://app.productplan.com/api/v2"
	version        = "3.0.0"
	defaultLimit   = 10
	maxLimit       = 100
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

	if resp.StatusCode == 204 {
		return json.RawMessage(`{"success": true}`), nil
	}

	return respBody, nil
}

// ============================================================================
// Response Summarization - Key token saver!
// ============================================================================

// SummarizeList extracts only essential fields from list responses
func SummarizeList(data json.RawMessage, itemType string) json.RawMessage {
	// Define which fields to keep for each type
	fieldsToKeep := map[string][]string{
		"roadmap":    {"id", "name", "updated_at"},
		"bar":        {"id", "name", "start_date", "end_date", "lane_id"},
		"lane":       {"id", "name", "color"},
		"milestone":  {"id", "name", "date"},
		"idea":       {"id", "title", "status", "created_at"},
		"objective":  {"id", "name", "status", "time_frame"},
		"key_result": {"id", "name", "current_value", "target_value"},
		"launch":     {"id", "name", "date", "status"},
		"task":       {"id", "name", "status"},
		"user":       {"id", "name", "email"},
		"team":       {"id", "name"},
		"default":    {"id", "name"},
	}

	fields, ok := fieldsToKeep[itemType]
	if !ok {
		fields = fieldsToKeep["default"]
	}

	// Try to parse as array first
	var items []map[string]interface{}
	if err := json.Unmarshal(data, &items); err != nil {
		// Try parsing as object with "results" key (ProductPlan API wrapper)
		var wrapper struct {
			Results []map[string]interface{} `json:"results"`
		}
		if err := json.Unmarshal(data, &wrapper); err != nil {
			return data // Can't parse, return as-is
		}
		items = wrapper.Results
	}

	summarized := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		summary := make(map[string]interface{})
		for _, field := range fields {
			if val, exists := item[field]; exists {
				summary[field] = val
			}
		}
		summarized = append(summarized, summary)
	}

	result, _ := json.Marshal(map[string]interface{}{
		"count": len(summarized),
		"items": summarized,
	})
	return result
}

// ============================================================================
// API Methods with Pagination
// ============================================================================

// Roadmaps
func (c *APIClient) ListRoadmaps(limit int) (json.RawMessage, error) {
	data, err := c.request("GET", "/roadmaps", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "roadmap"), nil
}

func (c *APIClient) GetRoadmap(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id, nil)
}

func (c *APIClient) GetRoadmapBars(id string, limit int) (json.RawMessage, error) {
	data, err := c.request("GET", "/roadmaps/"+id+"/bars", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "bar"), nil
}

func (c *APIClient) GetRoadmapComments(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id+"/comments", nil)
}

// Lanes
func (c *APIClient) ListLanes(roadmapID string) (json.RawMessage, error) {
	data, err := c.request("GET", "/roadmaps/"+roadmapID+"/lanes", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "lane"), nil
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

// Milestones
func (c *APIClient) ListMilestones(roadmapID string) (json.RawMessage, error) {
	data, err := c.request("GET", "/roadmaps/"+roadmapID+"/milestones", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "milestone"), nil
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

// Bars
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
	data, err := c.request("GET", "/bars/"+id+"/child_bars", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "bar"), nil
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

// Ideas
func (c *APIClient) ListIdeas() (json.RawMessage, error) {
	data, err := c.request("GET", "/discovery/ideas", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "idea"), nil
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

// Opportunities
func (c *APIClient) ListOpportunities() (json.RawMessage, error) {
	data, err := c.request("GET", "/discovery/opportunities", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "default"), nil
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

// Idea Forms
func (c *APIClient) ListIdeaForms() (json.RawMessage, error) {
	return c.request("GET", "/discovery/idea_forms", nil)
}

func (c *APIClient) GetIdeaForm(id string) (json.RawMessage, error) {
	return c.request("GET", "/discovery/idea_forms/"+id, nil)
}

// Objectives
func (c *APIClient) ListObjectives() (json.RawMessage, error) {
	data, err := c.request("GET", "/strategy/objectives", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "objective"), nil
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

// Key Results
func (c *APIClient) ListKeyResults(objectiveID string) (json.RawMessage, error) {
	data, err := c.request("GET", "/strategy/objectives/"+objectiveID+"/key_results", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "key_result"), nil
}

func (c *APIClient) GetKeyResult(objectiveID, keyResultID string) (json.RawMessage, error) {
	return c.request("GET", "/strategy/objectives/"+objectiveID+"/key_results/"+keyResultID, nil)
}

func (c *APIClient) CreateKeyResult(objectiveID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/strategy/objectives/"+objectiveID+"/key_results", data)
}

func (c *APIClient) UpdateKeyResult(objectiveID, keyResultID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/strategy/objectives/"+objectiveID+"/key_results/"+keyResultID, data)
}

func (c *APIClient) DeleteKeyResult(objectiveID, keyResultID string) (json.RawMessage, error) {
	return c.request("DELETE", "/strategy/objectives/"+objectiveID+"/key_results/"+keyResultID, nil)
}

// Launches
func (c *APIClient) ListLaunches() (json.RawMessage, error) {
	data, err := c.request("GET", "/launches", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "launch"), nil
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

// Checklist Sections
func (c *APIClient) ListChecklistSections(launchID string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+launchID+"/checklist_sections", nil)
}

func (c *APIClient) GetChecklistSection(launchID, sectionID string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+launchID+"/checklist_sections/"+sectionID, nil)
}

func (c *APIClient) CreateChecklistSection(launchID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/launches/"+launchID+"/checklist_sections", data)
}

func (c *APIClient) UpdateChecklistSection(launchID, sectionID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/launches/"+launchID+"/checklist_sections/"+sectionID, data)
}

func (c *APIClient) DeleteChecklistSection(launchID, sectionID string) (json.RawMessage, error) {
	return c.request("DELETE", "/launches/"+launchID+"/checklist_sections/"+sectionID, nil)
}

// Launch Tasks
func (c *APIClient) ListLaunchTasks(launchID string) (json.RawMessage, error) {
	data, err := c.request("GET", "/launches/"+launchID+"/tasks", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "task"), nil
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

// Administration
func (c *APIClient) ListUsers() (json.RawMessage, error) {
	data, err := c.request("GET", "/users", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "user"), nil
}

func (c *APIClient) ListTeams() (json.RawMessage, error) {
	data, err := c.request("GET", "/teams", nil)
	if err != nil {
		return nil, err
	}
	return SummarizeList(data, "team"), nil
}

func (c *APIClient) CheckStatus() (json.RawMessage, error) {
	return c.request("GET", "/status", nil)
}

// ============================================================================
// MCP Server Implementation - OPTIMIZED with 15 consolidated tools
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
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewMCPServer(client *APIClient) *MCPServer {
	return &MCPServer{client: client}
}

// getTools returns CONSOLIDATED tools - 15 instead of 58!
func (s *MCPServer) getTools() []Tool {
	return []Tool{
		// 1. Roadmaps - consolidated
		{
			Name:        "roadmaps",
			Description: "Manage roadmaps: list all, get details, get bars, or get comments",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":     {Type: "string", Description: "Action to perform", Enum: []string{"list", "get", "get_bars", "get_comments"}},
					"id":         {Type: "string", Description: "Roadmap ID (required for get, get_bars, get_comments)"},
				},
				Required: []string{"action"},
			},
		},
		// 2. Lanes - consolidated
		{
			Name:        "lanes",
			Description: "Manage lanes in a roadmap: list, create, update, or delete",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":     {Type: "string", Description: "Action to perform", Enum: []string{"list", "create", "update", "delete"}},
					"roadmap_id": {Type: "string", Description: "Roadmap ID (required for all actions)"},
					"lane_id":    {Type: "string", Description: "Lane ID (required for update, delete)"},
					"name":       {Type: "string", Description: "Lane name (for create, update)"},
					"color":      {Type: "string", Description: "Lane color hex (for create, update)"},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
		// 3. Milestones - consolidated
		{
			Name:        "milestones",
			Description: "Manage milestones in a roadmap: list, create, update, or delete",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":       {Type: "string", Description: "Action to perform", Enum: []string{"list", "create", "update", "delete"}},
					"roadmap_id":   {Type: "string", Description: "Roadmap ID (required for all actions)"},
					"milestone_id": {Type: "string", Description: "Milestone ID (required for update, delete)"},
					"name":         {Type: "string", Description: "Milestone name (for create, update)"},
					"date":         {Type: "string", Description: "Date YYYY-MM-DD (for create, update)"},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
		// 4. Bars - consolidated
		{
			Name:        "bars",
			Description: "Manage bars: get details, create, update, delete, get children, or get comments",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"get", "create", "update", "delete", "get_children", "get_comments"}},
					"id":          {Type: "string", Description: "Bar ID (required for get, update, delete, get_children, get_comments)"},
					"roadmap_id":  {Type: "string", Description: "Roadmap ID (required for create)"},
					"lane_id":     {Type: "string", Description: "Lane ID (required for create)"},
					"name":        {Type: "string", Description: "Bar name (for create, update)"},
					"start_date":  {Type: "string", Description: "Start date YYYY-MM-DD"},
					"end_date":    {Type: "string", Description: "End date YYYY-MM-DD"},
					"description": {Type: "string", Description: "Description"},
				},
				Required: []string{"action"},
			},
		},
		// 5. Bar connections - consolidated
		{
			Name:        "bar_connections",
			Description: "Manage bar connections: list, create, or delete dependencies between bars",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":            {Type: "string", Description: "Action to perform", Enum: []string{"list", "create", "delete"}},
					"bar_id":            {Type: "string", Description: "Source bar ID (required for all)"},
					"target_bar_id":     {Type: "string", Description: "Target bar ID (for create)"},
					"relationship_type": {Type: "string", Description: "Type: requires or required_by (for create)", Enum: []string{"requires", "required_by"}},
					"connection_id":     {Type: "string", Description: "Connection ID (for delete)"},
				},
				Required: []string{"action", "bar_id"},
			},
		},
		// 6. Bar links - consolidated
		{
			Name:        "bar_links",
			Description: "Manage external links on bars: list, create, or delete",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":  {Type: "string", Description: "Action to perform", Enum: []string{"list", "create", "delete"}},
					"bar_id":  {Type: "string", Description: "Bar ID (required for all)"},
					"url":     {Type: "string", Description: "URL (for create)"},
					"name":    {Type: "string", Description: "Link name (for create)"},
					"link_id": {Type: "string", Description: "Link ID (for delete)"},
				},
				Required: []string{"action", "bar_id"},
			},
		},
		// 7. Ideas - consolidated
		{
			Name:        "ideas",
			Description: "Manage ideas: list, get, create, update, or get metadata (customers/tags)",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"list", "get", "create", "update", "list_customers", "list_tags"}},
					"id":          {Type: "string", Description: "Idea ID (for get, update)"},
					"title":       {Type: "string", Description: "Title (for create, update)"},
					"description": {Type: "string", Description: "Description (for create, update)"},
				},
				Required: []string{"action"},
			},
		},
		// 8. Opportunities - consolidated
		{
			Name:        "opportunities",
			Description: "Manage opportunities: list, get, create, or update",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"list", "get", "create", "update"}},
					"id":          {Type: "string", Description: "Opportunity ID (for get, update)"},
					"name":        {Type: "string", Description: "Name (for create, update)"},
					"description": {Type: "string", Description: "Description (for create, update)"},
				},
				Required: []string{"action"},
			},
		},
		// 9. Idea forms - consolidated
		{
			Name:        "idea_forms",
			Description: "Manage idea forms: list or get details",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action": {Type: "string", Description: "Action to perform", Enum: []string{"list", "get"}},
					"id":     {Type: "string", Description: "Form ID (for get)"},
				},
				Required: []string{"action"},
			},
		},
		// 10. Objectives (OKRs) - consolidated
		{
			Name:        "objectives",
			Description: "Manage OKR objectives: list, get, create, update, or delete",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"list", "get", "create", "update", "delete"}},
					"id":          {Type: "string", Description: "Objective ID (for get, update, delete)"},
					"name":        {Type: "string", Description: "Name (for create, update)"},
					"description": {Type: "string", Description: "Description (for create, update)"},
					"time_frame":  {Type: "string", Description: "Time frame (for create)"},
				},
				Required: []string{"action"},
			},
		},
		// 11. Key Results - consolidated
		{
			Name:        "key_results",
			Description: "Manage key results for objectives: list, get, create, update, or delete",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":        {Type: "string", Description: "Action to perform", Enum: []string{"list", "get", "create", "update", "delete"}},
					"objective_id":  {Type: "string", Description: "Objective ID (required for all)"},
					"key_result_id": {Type: "string", Description: "Key Result ID (for get, update, delete)"},
					"name":          {Type: "string", Description: "Name (for create, update)"},
					"target_value":  {Type: "string", Description: "Target value (for create, update)"},
					"current_value": {Type: "string", Description: "Current value (for create, update)"},
				},
				Required: []string{"action", "objective_id"},
			},
		},
		// 12. Launches - consolidated
		{
			Name:        "launches",
			Description: "Manage launches: list, get, create, update, or delete",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action": {Type: "string", Description: "Action to perform", Enum: []string{"list", "get", "create", "update", "delete"}},
					"id":     {Type: "string", Description: "Launch ID (for get, update, delete)"},
					"name":   {Type: "string", Description: "Name (for create, update)"},
					"date":   {Type: "string", Description: "Date YYYY-MM-DD (for create, update)"},
				},
				Required: []string{"action"},
			},
		},
		// 13. Checklist sections - consolidated
		{
			Name:        "checklist_sections",
			Description: "Manage checklist sections for launches: list, get, create, update, or delete",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":     {Type: "string", Description: "Action to perform", Enum: []string{"list", "get", "create", "update", "delete"}},
					"launch_id":  {Type: "string", Description: "Launch ID (required for all)"},
					"section_id": {Type: "string", Description: "Section ID (for get, update, delete)"},
					"name":       {Type: "string", Description: "Section name (for create, update)"},
				},
				Required: []string{"action", "launch_id"},
			},
		},
		// 14. Launch tasks - consolidated
		{
			Name:        "launch_tasks",
			Description: "Manage tasks for launches: list, get, create, update, or delete",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":               {Type: "string", Description: "Action to perform", Enum: []string{"list", "get", "create", "update", "delete"}},
					"launch_id":            {Type: "string", Description: "Launch ID (required for all)"},
					"task_id":              {Type: "string", Description: "Task ID (for get, update, delete)"},
					"name":                 {Type: "string", Description: "Task name (for create, update)"},
					"checklist_section_id": {Type: "string", Description: "Section ID (for create)"},
					"status":               {Type: "string", Description: "Status (for update)", Enum: []string{"to_do", "in_progress", "completed"}},
				},
				Required: []string{"action", "launch_id"},
			},
		},
		// 15. Admin - consolidated
		{
			Name:        "admin",
			Description: "Administrative actions: list users, list teams, or check API status",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action": {Type: "string", Description: "Action to perform", Enum: []string{"list_users", "list_teams", "check_status"}},
				},
				Required: []string{"action"},
			},
		},
	}
}

func (s *MCPServer) handleToolCall(name string, args map[string]interface{}) (json.RawMessage, error) {
	getString := func(key string) string {
		if v, ok := args[key].(string); ok {
			return v
		}
		return ""
	}

	action := getString("action")

	switch name {
	// 1. Roadmaps
	case "roadmaps":
		switch action {
		case "list":
			return s.client.ListRoadmaps(defaultLimit)
		case "get":
			return s.client.GetRoadmap(getString("id"))
		case "get_bars":
			return s.client.GetRoadmapBars(getString("id"), defaultLimit)
		case "get_comments":
			return s.client.GetRoadmapComments(getString("id"))
		}

	// 2. Lanes
	case "lanes":
		roadmapID := getString("roadmap_id")
		switch action {
		case "list":
			return s.client.ListLanes(roadmapID)
		case "create":
			data := map[string]interface{}{"name": getString("name")}
			if c := getString("color"); c != "" {
				data["color"] = c
			}
			return s.client.CreateLane(roadmapID, data)
		case "update":
			data := make(map[string]interface{})
			if n := getString("name"); n != "" {
				data["name"] = n
			}
			if c := getString("color"); c != "" {
				data["color"] = c
			}
			return s.client.UpdateLane(roadmapID, getString("lane_id"), data)
		case "delete":
			return s.client.DeleteLane(roadmapID, getString("lane_id"))
		}

	// 3. Milestones
	case "milestones":
		roadmapID := getString("roadmap_id")
		switch action {
		case "list":
			return s.client.ListMilestones(roadmapID)
		case "create":
			data := map[string]interface{}{"name": getString("name"), "date": getString("date")}
			return s.client.CreateMilestone(roadmapID, data)
		case "update":
			data := make(map[string]interface{})
			if n := getString("name"); n != "" {
				data["name"] = n
			}
			if d := getString("date"); d != "" {
				data["date"] = d
			}
			return s.client.UpdateMilestone(roadmapID, getString("milestone_id"), data)
		case "delete":
			return s.client.DeleteMilestone(roadmapID, getString("milestone_id"))
		}

	// 4. Bars
	case "bars":
		switch action {
		case "get":
			return s.client.GetBar(getString("id"))
		case "create":
			data := map[string]interface{}{
				"roadmap_id": getString("roadmap_id"),
				"lane_id":    getString("lane_id"),
				"name":       getString("name"),
			}
			if sd := getString("start_date"); sd != "" {
				data["start_date"] = sd
			}
			if ed := getString("end_date"); ed != "" {
				data["end_date"] = ed
			}
			if desc := getString("description"); desc != "" {
				data["description"] = desc
			}
			return s.client.CreateBar(data)
		case "update":
			data := make(map[string]interface{})
			if n := getString("name"); n != "" {
				data["name"] = n
			}
			if sd := getString("start_date"); sd != "" {
				data["start_date"] = sd
			}
			if ed := getString("end_date"); ed != "" {
				data["end_date"] = ed
			}
			if desc := getString("description"); desc != "" {
				data["description"] = desc
			}
			return s.client.UpdateBar(getString("id"), data)
		case "delete":
			return s.client.DeleteBar(getString("id"))
		case "get_children":
			return s.client.GetBarChildBars(getString("id"))
		case "get_comments":
			return s.client.GetBarComments(getString("id"))
		}

	// 5. Bar connections
	case "bar_connections":
		barID := getString("bar_id")
		switch action {
		case "list":
			return s.client.ListBarConnections(barID)
		case "create":
			targetBarID := getString("target_bar_id")
			relationshipType := getString("relationship_type")
			targetInt, _ := strconv.Atoi(targetBarID)
			payload := map[string]interface{}{relationshipType: targetInt}
			return s.client.CreateBarConnection(barID, payload)
		case "delete":
			return s.client.DeleteBarConnection(barID, getString("connection_id"))
		}

	// 6. Bar links
	case "bar_links":
		barID := getString("bar_id")
		switch action {
		case "list":
			return s.client.ListBarLinks(barID)
		case "create":
			data := map[string]interface{}{"url": getString("url")}
			if n := getString("name"); n != "" {
				data["name"] = n
			}
			return s.client.CreateBarLink(barID, data)
		case "delete":
			return s.client.DeleteBarLink(barID, getString("link_id"))
		}

	// 7. Ideas
	case "ideas":
		switch action {
		case "list":
			return s.client.ListIdeas()
		case "get":
			return s.client.GetIdea(getString("id"))
		case "create":
			data := map[string]interface{}{"title": getString("title")}
			if desc := getString("description"); desc != "" {
				data["description"] = desc
			}
			return s.client.CreateIdea(data)
		case "update":
			data := make(map[string]interface{})
			if t := getString("title"); t != "" {
				data["title"] = t
			}
			if desc := getString("description"); desc != "" {
				data["description"] = desc
			}
			return s.client.UpdateIdea(getString("id"), data)
		case "list_customers":
			return s.client.ListIdeaCustomers()
		case "list_tags":
			return s.client.ListIdeaTags()
		}

	// 8. Opportunities
	case "opportunities":
		switch action {
		case "list":
			return s.client.ListOpportunities()
		case "get":
			return s.client.GetOpportunity(getString("id"))
		case "create":
			data := map[string]interface{}{"name": getString("name")}
			if desc := getString("description"); desc != "" {
				data["description"] = desc
			}
			return s.client.CreateOpportunity(data)
		case "update":
			data := make(map[string]interface{})
			if n := getString("name"); n != "" {
				data["name"] = n
			}
			if desc := getString("description"); desc != "" {
				data["description"] = desc
			}
			return s.client.UpdateOpportunity(getString("id"), data)
		}

	// 9. Idea forms
	case "idea_forms":
		switch action {
		case "list":
			return s.client.ListIdeaForms()
		case "get":
			return s.client.GetIdeaForm(getString("id"))
		}

	// 10. Objectives
	case "objectives":
		switch action {
		case "list":
			return s.client.ListObjectives()
		case "get":
			return s.client.GetObjective(getString("id"))
		case "create":
			data := map[string]interface{}{"name": getString("name")}
			if desc := getString("description"); desc != "" {
				data["description"] = desc
			}
			if tf := getString("time_frame"); tf != "" {
				data["time_frame"] = tf
			}
			return s.client.CreateObjective(data)
		case "update":
			data := make(map[string]interface{})
			if n := getString("name"); n != "" {
				data["name"] = n
			}
			if desc := getString("description"); desc != "" {
				data["description"] = desc
			}
			return s.client.UpdateObjective(getString("id"), data)
		case "delete":
			return s.client.DeleteObjective(getString("id"))
		}

	// 11. Key results
	case "key_results":
		objectiveID := getString("objective_id")
		switch action {
		case "list":
			return s.client.ListKeyResults(objectiveID)
		case "get":
			return s.client.GetKeyResult(objectiveID, getString("key_result_id"))
		case "create":
			data := map[string]interface{}{"name": getString("name")}
			if tv := getString("target_value"); tv != "" {
				data["target_value"] = tv
			}
			if cv := getString("current_value"); cv != "" {
				data["current_value"] = cv
			}
			return s.client.CreateKeyResult(objectiveID, data)
		case "update":
			data := make(map[string]interface{})
			if n := getString("name"); n != "" {
				data["name"] = n
			}
			if cv := getString("current_value"); cv != "" {
				data["current_value"] = cv
			}
			return s.client.UpdateKeyResult(objectiveID, getString("key_result_id"), data)
		case "delete":
			return s.client.DeleteKeyResult(objectiveID, getString("key_result_id"))
		}

	// 12. Launches
	case "launches":
		switch action {
		case "list":
			return s.client.ListLaunches()
		case "get":
			return s.client.GetLaunch(getString("id"))
		case "create":
			data := map[string]interface{}{"name": getString("name")}
			if d := getString("date"); d != "" {
				data["date"] = d
			}
			return s.client.CreateLaunch(data)
		case "update":
			data := make(map[string]interface{})
			if n := getString("name"); n != "" {
				data["name"] = n
			}
			if d := getString("date"); d != "" {
				data["date"] = d
			}
			return s.client.UpdateLaunch(getString("id"), data)
		case "delete":
			return s.client.DeleteLaunch(getString("id"))
		}

	// 13. Checklist sections
	case "checklist_sections":
		launchID := getString("launch_id")
		switch action {
		case "list":
			return s.client.ListChecklistSections(launchID)
		case "get":
			return s.client.GetChecklistSection(launchID, getString("section_id"))
		case "create":
			data := map[string]interface{}{"name": getString("name")}
			return s.client.CreateChecklistSection(launchID, data)
		case "update":
			data := map[string]interface{}{"name": getString("name")}
			return s.client.UpdateChecklistSection(launchID, getString("section_id"), data)
		case "delete":
			return s.client.DeleteChecklistSection(launchID, getString("section_id"))
		}

	// 14. Launch tasks
	case "launch_tasks":
		launchID := getString("launch_id")
		switch action {
		case "list":
			return s.client.ListLaunchTasks(launchID)
		case "get":
			return s.client.GetLaunchTask(launchID, getString("task_id"))
		case "create":
			data := map[string]interface{}{
				"name":                 getString("name"),
				"checklist_section_id": getString("checklist_section_id"),
			}
			return s.client.CreateLaunchTask(launchID, data)
		case "update":
			data := make(map[string]interface{})
			if n := getString("name"); n != "" {
				data["name"] = n
			}
			if st := getString("status"); st != "" {
				data["status"] = st
			}
			return s.client.UpdateLaunchTask(launchID, getString("task_id"), data)
		case "delete":
			return s.client.DeleteLaunchTask(launchID, getString("task_id"))
		}

	// 15. Admin
	case "admin":
		switch action {
		case "list_users":
			return s.client.ListUsers()
		case "list_teams":
			return s.client.ListTeams()
		case "check_status":
			return s.client.CheckStatus()
		}
	}

	return nil, fmt.Errorf("unknown tool or action: %s/%s", name, action)
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
			// OPTIMIZATION: Use compact JSON instead of pretty-printed
			resp.Result = map[string]interface{}{
				"content": []ToolContent{{Type: "text", Text: string(result)}},
			}
		}

	default:
		resp.Error = &RPCError{Code: -32601, Message: "Method not found: " + req.Method}
	}

	return resp
}

func (s *MCPServer) Run() {
	fmt.Fprintln(os.Stderr, "ProductPlan MCP Server v"+version+" (optimized) running on stdio")
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
			result, err = client.ListRoadmaps(defaultLimit)
		} else {
			result, err = client.GetRoadmap(subArgs[0])
		}

	case "bars":
		if len(subArgs) == 0 {
			fmt.Println("Usage: productplan bars <roadmap_id>")
			os.Exit(1)
		}
		result, err = client.GetRoadmapBars(subArgs[0], defaultLimit)

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

func printUsage() {
	fmt.Printf(`ProductPlan CLI & MCP Server v%s (Optimized)

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

Optimizations (v3.0):
  - Consolidated 58 tools into 15 tools (74%% reduction)
  - Response summarization for list operations
  - Compact JSON responses (no pretty-printing in MCP mode)

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

	if len(args) == 0 || args[0] == "serve" || args[0] == "mcp" {
		client := NewAPIClient(apiToken)
		server := NewMCPServer(client)
		server.Run()
		return
	}

	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printUsage()
		return
	}

	runCLI(args)
}
