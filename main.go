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
	version = "4.0.0"
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
// Smart Response Formatting - Enriched data to reduce follow-up calls
// ============================================================================

// FormatRoadmapList returns roadmaps with useful counts
func (c *APIClient) FormatRoadmapList(data json.RawMessage) json.RawMessage {
	var roadmaps []map[string]interface{}
	if err := json.Unmarshal(data, &roadmaps); err != nil {
		return data
	}

	results := make([]map[string]interface{}, 0, len(roadmaps))
	for _, rm := range roadmaps {
		results = append(results, map[string]interface{}{
			"id":         rm["id"],
			"name":       rm["name"],
			"updated_at": rm["updated_at"],
		})
	}

	output, _ := json.Marshal(map[string]interface{}{
		"count":    len(results),
		"roadmaps": results,
		"hint":     "Use get_roadmap_bars with a roadmap id to see its items",
	})
	return output
}

// FormatBarsWithContext enriches bars with lane names
func (c *APIClient) FormatBarsWithContext(bars json.RawMessage, lanes json.RawMessage) json.RawMessage {
	var barList []map[string]interface{}
	var laneList []map[string]interface{}

	json.Unmarshal(bars, &barList)
	json.Unmarshal(lanes, &laneList)

	// Build lane lookup
	laneLookup := make(map[float64]string)
	for _, lane := range laneList {
		if id, ok := lane["id"].(float64); ok {
			if name, ok := lane["name"].(string); ok {
				laneLookup[id] = name
			}
		}
	}

	results := make([]map[string]interface{}, 0, len(barList))
	for _, bar := range barList {
		laneID, _ := bar["lane_id"].(float64)
		laneName := laneLookup[laneID]
		if laneName == "" {
			laneName = "Unknown"
		}

		results = append(results, map[string]interface{}{
			"id":         bar["id"],
			"name":       bar["name"],
			"start_date": bar["start_date"],
			"end_date":   bar["end_date"],
			"lane_id":    bar["lane_id"],
			"lane_name":  laneName,
		})
	}

	output, _ := json.Marshal(map[string]interface{}{
		"count": len(results),
		"bars":  results,
	})
	return output
}

// FormatLanes formats lane list
func FormatLanes(data json.RawMessage) json.RawMessage {
	var lanes []map[string]interface{}
	if err := json.Unmarshal(data, &lanes); err != nil {
		return data
	}

	results := make([]map[string]interface{}, 0, len(lanes))
	for _, lane := range lanes {
		results = append(results, map[string]interface{}{
			"id":    lane["id"],
			"name":  lane["name"],
			"color": lane["color"],
		})
	}

	output, _ := json.Marshal(map[string]interface{}{
		"count": len(results),
		"lanes": results,
	})
	return output
}

// FormatMilestones formats milestone list
func FormatMilestones(data json.RawMessage) json.RawMessage {
	var milestones []map[string]interface{}
	if err := json.Unmarshal(data, &milestones); err != nil {
		return data
	}

	results := make([]map[string]interface{}, 0, len(milestones))
	for _, m := range milestones {
		results = append(results, map[string]interface{}{
			"id":   m["id"],
			"name": m["name"],
			"date": m["date"],
		})
	}

	output, _ := json.Marshal(map[string]interface{}{
		"count":      len(results),
		"milestones": results,
	})
	return output
}

// FormatObjectives formats objective list
func FormatObjectives(data json.RawMessage) json.RawMessage {
	var objectives []map[string]interface{}
	if err := json.Unmarshal(data, &objectives); err != nil {
		return data
	}

	results := make([]map[string]interface{}, 0, len(objectives))
	for _, obj := range objectives {
		results = append(results, map[string]interface{}{
			"id":         obj["id"],
			"name":       obj["name"],
			"status":     obj["status"],
			"time_frame": obj["time_frame"],
		})
	}

	output, _ := json.Marshal(map[string]interface{}{
		"count":      len(results),
		"objectives": results,
		"hint":       "Use get_objective with an id for full details including key results",
	})
	return output
}

// FormatIdeas formats idea list
func FormatIdeas(data json.RawMessage) json.RawMessage {
	var wrapper struct {
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		// Try as array
		var ideas []map[string]interface{}
		if err := json.Unmarshal(data, &ideas); err != nil {
			return data
		}
		wrapper.Results = ideas
	}

	results := make([]map[string]interface{}, 0, len(wrapper.Results))
	for _, idea := range wrapper.Results {
		results = append(results, map[string]interface{}{
			"id":         idea["id"],
			"title":      idea["title"],
			"status":     idea["status"],
			"created_at": idea["created_at"],
		})
	}

	output, _ := json.Marshal(map[string]interface{}{
		"count": len(results),
		"ideas": results,
	})
	return output
}

// FormatLaunches formats launch list
func FormatLaunches(data json.RawMessage) json.RawMessage {
	var launches []map[string]interface{}
	if err := json.Unmarshal(data, &launches); err != nil {
		return data
	}

	results := make([]map[string]interface{}, 0, len(launches))
	for _, launch := range launches {
		results = append(results, map[string]interface{}{
			"id":     launch["id"],
			"name":   launch["name"],
			"date":   launch["date"],
			"status": launch["status"],
		})
	}

	output, _ := json.Marshal(map[string]interface{}{
		"count":    len(results),
		"launches": results,
	})
	return output
}

// ============================================================================
// API Methods
// ============================================================================

// Roadmaps
func (c *APIClient) ListRoadmaps() (json.RawMessage, error) {
	data, err := c.request("GET", "/roadmaps", nil)
	if err != nil {
		return nil, err
	}
	return c.FormatRoadmapList(data), nil
}

func (c *APIClient) GetRoadmap(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id, nil)
}

func (c *APIClient) GetRoadmapBars(id string) (json.RawMessage, error) {
	bars, err := c.request("GET", "/roadmaps/"+id+"/bars", nil)
	if err != nil {
		return nil, err
	}
	lanes, _ := c.request("GET", "/roadmaps/"+id+"/lanes", nil)
	return c.FormatBarsWithContext(bars, lanes), nil
}

func (c *APIClient) GetRoadmapLanes(id string) (json.RawMessage, error) {
	data, err := c.request("GET", "/roadmaps/"+id+"/lanes", nil)
	if err != nil {
		return nil, err
	}
	return FormatLanes(data), nil
}

func (c *APIClient) GetRoadmapMilestones(id string) (json.RawMessage, error) {
	data, err := c.request("GET", "/roadmaps/"+id+"/milestones", nil)
	if err != nil {
		return nil, err
	}
	return FormatMilestones(data), nil
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

// Lanes
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
func (c *APIClient) CreateMilestone(roadmapID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/roadmaps/"+roadmapID+"/milestones", data)
}

func (c *APIClient) UpdateMilestone(roadmapID, milestoneID string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/roadmaps/"+roadmapID+"/milestones/"+milestoneID, data)
}

func (c *APIClient) DeleteMilestone(roadmapID, milestoneID string) (json.RawMessage, error) {
	return c.request("DELETE", "/roadmaps/"+roadmapID+"/milestones/"+milestoneID, nil)
}

// Objectives
func (c *APIClient) ListObjectives() (json.RawMessage, error) {
	data, err := c.request("GET", "/strategy/objectives", nil)
	if err != nil {
		return nil, err
	}
	return FormatObjectives(data), nil
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
	return c.request("GET", "/strategy/objectives/"+objectiveID+"/key_results", nil)
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

// Ideas
func (c *APIClient) ListIdeas() (json.RawMessage, error) {
	data, err := c.request("GET", "/discovery/ideas", nil)
	if err != nil {
		return nil, err
	}
	return FormatIdeas(data), nil
}

func (c *APIClient) GetIdea(id string) (json.RawMessage, error) {
	return c.request("GET", "/discovery/ideas/"+id, nil)
}

// Launches
func (c *APIClient) ListLaunches() (json.RawMessage, error) {
	data, err := c.request("GET", "/launches", nil)
	if err != nil {
		return nil, err
	}
	return FormatLaunches(data), nil
}

func (c *APIClient) GetLaunch(id string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+id, nil)
}

// Admin
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
// MCP Server Implementation - Optimized for AI understanding
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

// getTools returns well-designed tools optimized for AI decision-making
func (s *MCPServer) getTools() []Tool {
	return []Tool{
		// =====================
		// READ TOOLS - Granular, no-hassle queries
		// =====================
		{
			Name:        "list_roadmaps",
			Description: "List all roadmaps. Call this FIRST to get roadmap IDs before querying bars or lanes. No parameters needed.",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "get_roadmap",
			Description: "Get full details of a specific roadmap including settings and metadata.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID (get from list_roadmaps)"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name:        "get_roadmap_bars",
			Description: "Get all bars (items) in a roadmap. Returns bars with their lane names for context. Use this to see what's planned on a roadmap.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name:        "get_roadmap_lanes",
			Description: "Get all lanes (swim lanes/rows) in a roadmap. Lanes organize bars into categories.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name:        "get_roadmap_milestones",
			Description: "Get all milestones (key dates) in a roadmap.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"roadmap_id": {Type: "string", Description: "The roadmap ID"},
				},
				Required: []string{"roadmap_id"},
			},
		},
		{
			Name:        "get_bar",
			Description: "Get full details of a specific bar including description, links, and custom fields.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"bar_id": {Type: "string", Description: "The bar ID"},
				},
				Required: []string{"bar_id"},
			},
		},
		{
			Name:        "list_objectives",
			Description: "List all OKR objectives. Call this to see strategic goals. No parameters needed.",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "get_objective",
			Description: "Get full details of an objective including its key results.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"objective_id": {Type: "string", Description: "The objective ID"},
				},
				Required: []string{"objective_id"},
			},
		},
		{
			Name:        "list_key_results",
			Description: "List key results for a specific objective.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"objective_id": {Type: "string", Description: "The objective ID"},
				},
				Required: []string{"objective_id"},
			},
		},
		{
			Name:        "list_ideas",
			Description: "List all ideas in the discovery/feedback pipeline. No parameters needed.",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "get_idea",
			Description: "Get full details of a specific idea.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"idea_id": {Type: "string", Description: "The idea ID"},
				},
				Required: []string{"idea_id"},
			},
		},
		{
			Name:        "list_launches",
			Description: "List all product launches. No parameters needed.",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "get_launch",
			Description: "Get full details of a specific launch including checklist.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"launch_id": {Type: "string", Description: "The launch ID"},
				},
				Required: []string{"launch_id"},
			},
		},
		{
			Name:        "check_status",
			Description: "Check ProductPlan API status and authentication. Use to verify connection.",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},

		// =====================
		// WRITE TOOLS - Consolidated by entity type
		// =====================
		{
			Name:        "manage_bar",
			Description: "Create, update, or delete a bar on a roadmap.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":      {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"bar_id":      {Type: "string", Description: "Bar ID (required for update/delete)"},
					"roadmap_id":  {Type: "string", Description: "Roadmap ID (required for create)"},
					"lane_id":     {Type: "string", Description: "Lane ID (required for create)"},
					"name":        {Type: "string", Description: "Bar name"},
					"start_date":  {Type: "string", Description: "Start date YYYY-MM-DD"},
					"end_date":    {Type: "string", Description: "End date YYYY-MM-DD"},
					"description": {Type: "string", Description: "Description text"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name:        "manage_lane",
			Description: "Create, update, or delete a lane on a roadmap.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":     {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"roadmap_id": {Type: "string", Description: "Roadmap ID (required for all actions)"},
					"lane_id":    {Type: "string", Description: "Lane ID (required for update/delete)"},
					"name":       {Type: "string", Description: "Lane name"},
					"color":      {Type: "string", Description: "Lane color hex code"},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
		{
			Name:        "manage_milestone",
			Description: "Create, update, or delete a milestone on a roadmap.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":       {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"roadmap_id":   {Type: "string", Description: "Roadmap ID (required for all actions)"},
					"milestone_id": {Type: "string", Description: "Milestone ID (required for update/delete)"},
					"name":         {Type: "string", Description: "Milestone name"},
					"date":         {Type: "string", Description: "Date YYYY-MM-DD"},
				},
				Required: []string{"action", "roadmap_id"},
			},
		},
		{
			Name:        "manage_objective",
			Description: "Create, update, or delete an OKR objective.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":       {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"objective_id": {Type: "string", Description: "Objective ID (required for update/delete)"},
					"name":         {Type: "string", Description: "Objective name"},
					"description":  {Type: "string", Description: "Description"},
					"time_frame":   {Type: "string", Description: "Time frame (e.g., Q1 2024)"},
				},
				Required: []string{"action"},
			},
		},
		{
			Name:        "manage_key_result",
			Description: "Create, update, or delete a key result for an objective.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"action":        {Type: "string", Description: "Action to perform", Enum: []string{"create", "update", "delete"}},
					"objective_id":  {Type: "string", Description: "Parent objective ID (required for all actions)"},
					"key_result_id": {Type: "string", Description: "Key result ID (required for update/delete)"},
					"name":          {Type: "string", Description: "Key result name"},
					"target_value":  {Type: "string", Description: "Target value"},
					"current_value": {Type: "string", Description: "Current value"},
				},
				Required: []string{"action", "objective_id"},
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

	switch name {
	// READ TOOLS
	case "list_roadmaps":
		return s.client.ListRoadmaps()

	case "get_roadmap":
		return s.client.GetRoadmap(getString("roadmap_id"))

	case "get_roadmap_bars":
		return s.client.GetRoadmapBars(getString("roadmap_id"))

	case "get_roadmap_lanes":
		return s.client.GetRoadmapLanes(getString("roadmap_id"))

	case "get_roadmap_milestones":
		return s.client.GetRoadmapMilestones(getString("roadmap_id"))

	case "get_bar":
		return s.client.GetBar(getString("bar_id"))

	case "list_objectives":
		return s.client.ListObjectives()

	case "get_objective":
		return s.client.GetObjective(getString("objective_id"))

	case "list_key_results":
		return s.client.ListKeyResults(getString("objective_id"))

	case "list_ideas":
		return s.client.ListIdeas()

	case "get_idea":
		return s.client.GetIdea(getString("idea_id"))

	case "list_launches":
		return s.client.ListLaunches()

	case "get_launch":
		return s.client.GetLaunch(getString("launch_id"))

	case "check_status":
		return s.client.CheckStatus()

	// WRITE TOOLS
	case "manage_bar":
		action := getString("action")
		switch action {
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
			return s.client.UpdateBar(getString("bar_id"), data)
		case "delete":
			return s.client.DeleteBar(getString("bar_id"))
		}

	case "manage_lane":
		action := getString("action")
		roadmapID := getString("roadmap_id")
		switch action {
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

	case "manage_milestone":
		action := getString("action")
		roadmapID := getString("roadmap_id")
		switch action {
		case "create":
			data := map[string]interface{}{
				"name": getString("name"),
				"date": getString("date"),
			}
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

	case "manage_objective":
		action := getString("action")
		switch action {
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
			return s.client.UpdateObjective(getString("objective_id"), data)
		case "delete":
			return s.client.DeleteObjective(getString("objective_id"))
		}

	case "manage_key_result":
		action := getString("action")
		objectiveID := getString("objective_id")
		switch action {
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
	}

	return nil, fmt.Errorf("unknown tool: %s", name)
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
		result, err = client.GetRoadmapLanes(subArgs[0])

	case "milestones":
		if len(subArgs) == 0 {
			fmt.Println("Usage: productplan milestones <roadmap_id>")
			os.Exit(1)
		}
		result, err = client.GetRoadmapMilestones(subArgs[0])

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

	case "launches":
		if len(subArgs) == 0 {
			result, err = client.ListLaunches()
		} else {
			result, err = client.GetLaunch(subArgs[0])
		}

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
	fmt.Printf(`ProductPlan CLI & MCP Server v%s

Usage:
  productplan <command> [arguments]
  productplan serve                    Start MCP server (for AI assistants)

Commands:
  roadmaps [id]                        List roadmaps or get details
  bars <roadmap_id>                    List bars in a roadmap (with lane names)
  lanes <roadmap_id>                   List lanes in a roadmap
  milestones <roadmap_id>              List milestones in a roadmap
  objectives [id]                      List objectives or get details
  key-results <objective_id>           List key results for an objective
  ideas [id]                           List ideas or get details
  launches [id]                        List launches or get details
  status                               Check API status

Environment:
  PRODUCTPLAN_API_TOKEN                Your ProductPlan API token (required)

Design (v4.0):
  - 14 granular READ tools (no params needed for lists)
  - 5 consolidated WRITE tools (action-based)
  - Enriched responses (bars include lane names)
  - Clear tool descriptions for AI decision-making

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
