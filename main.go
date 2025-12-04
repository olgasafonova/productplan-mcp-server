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
	version = "1.0.0"
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

	return respBody, nil
}

// API Methods
func (c *APIClient) ListRoadmaps() (json.RawMessage, error) {
	return c.request("GET", "/roadmaps", nil)
}

func (c *APIClient) GetRoadmap(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id, nil)
}

func (c *APIClient) GetRoadmapBars(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id+"/bars", nil)
}

func (c *APIClient) GetRoadmapLanes(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id+"/lanes", nil)
}

func (c *APIClient) GetRoadmapMilestones(id string) (json.RawMessage, error) {
	return c.request("GET", "/roadmaps/"+id+"/milestones", nil)
}

func (c *APIClient) GetBar(id string) (json.RawMessage, error) {
	return c.request("GET", "/bars/"+id, nil)
}

func (c *APIClient) CreateBar(data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/bars", data)
}

func (c *APIClient) UpdateBar(id string, data map[string]interface{}) (json.RawMessage, error) {
	return c.request("PATCH", "/bars/"+id, data)
}

func (c *APIClient) ListIdeas() (json.RawMessage, error) {
	return c.request("GET", "/discovery/ideas", nil)
}

func (c *APIClient) GetIdea(id string) (json.RawMessage, error) {
	return c.request("GET", "/discovery/ideas/"+id, nil)
}

func (c *APIClient) CreateIdea(data map[string]interface{}) (json.RawMessage, error) {
	return c.request("POST", "/discovery/ideas", data)
}

func (c *APIClient) ListOpportunities() (json.RawMessage, error) {
	return c.request("GET", "/discovery/opportunities", nil)
}

func (c *APIClient) ListObjectives() (json.RawMessage, error) {
	return c.request("GET", "/strategy/objectives", nil)
}

func (c *APIClient) GetObjective(id string) (json.RawMessage, error) {
	return c.request("GET", "/strategy/objectives/"+id, nil)
}

func (c *APIClient) ListKeyResults(objectiveID string) (json.RawMessage, error) {
	return c.request("GET", "/strategy/objectives/"+objectiveID+"/key-results", nil)
}

func (c *APIClient) ListLaunches() (json.RawMessage, error) {
	return c.request("GET", "/launches", nil)
}

func (c *APIClient) GetLaunch(id string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+id, nil)
}

func (c *APIClient) ListLaunchTasks(id string) (json.RawMessage, error) {
	return c.request("GET", "/launches/"+id+"/tasks", nil)
}

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
		{Name: "list_roadmaps", Description: "List all roadmaps in your ProductPlan account", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_roadmap", Description: "Get details of a specific roadmap", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Roadmap ID"}}, Required: []string{"id"}}},
		{Name: "get_roadmap_bars", Description: "Get all bars (items) from a roadmap", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}}, Required: []string{"roadmap_id"}}},
		{Name: "get_roadmap_lanes", Description: "Get all lanes from a roadmap", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}}, Required: []string{"roadmap_id"}}},
		{Name: "get_roadmap_milestones", Description: "Get all milestones from a roadmap", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}}, Required: []string{"roadmap_id"}}},
		{Name: "get_bar", Description: "Get details of a specific bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Bar ID"}}, Required: []string{"id"}}},
		{Name: "create_bar", Description: "Create a new bar on a roadmap", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"roadmap_id": {Type: "string", Description: "Roadmap ID"}, "lane_id": {Type: "string", Description: "Lane ID"}, "name": {Type: "string", Description: "Bar name"}, "start_date": {Type: "string", Description: "Start date (YYYY-MM-DD)"}, "end_date": {Type: "string", Description: "End date (YYYY-MM-DD)"}, "description": {Type: "string", Description: "Bar description"}}, Required: []string{"roadmap_id", "lane_id", "name"}}},
		{Name: "update_bar", Description: "Update an existing bar", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Bar ID"}, "name": {Type: "string", Description: "Bar name"}, "start_date": {Type: "string", Description: "Start date"}, "end_date": {Type: "string", Description: "End date"}, "description": {Type: "string", Description: "Description"}}, Required: []string{"id"}}},
		{Name: "list_ideas", Description: "List all ideas in Discovery", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_idea", Description: "Get details of a specific idea", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Idea ID"}}, Required: []string{"id"}}},
		{Name: "create_idea", Description: "Create a new idea", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"title": {Type: "string", Description: "Idea title"}, "description": {Type: "string", Description: "Idea description"}}, Required: []string{"title"}}},
		{Name: "list_opportunities", Description: "List all opportunities in Discovery", InputSchema: InputSchema{Type: "object"}},
		{Name: "list_objectives", Description: "List all strategic objectives", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_objective", Description: "Get details of a specific objective", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Objective ID"}}, Required: []string{"id"}}},
		{Name: "list_key_results", Description: "List key results for an objective", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"objective_id": {Type: "string", Description: "Objective ID"}}, Required: []string{"objective_id"}}},
		{Name: "list_launches", Description: "List all launches", InputSchema: InputSchema{Type: "object"}},
		{Name: "get_launch", Description: "Get details of a specific launch", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"id": {Type: "string", Description: "Launch ID"}}, Required: []string{"id"}}},
		{Name: "list_launch_tasks", Description: "List tasks for a launch", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{"launch_id": {Type: "string", Description: "Launch ID"}}, Required: []string{"launch_id"}}},
		{Name: "list_users", Description: "List all users in the account", InputSchema: InputSchema{Type: "object"}},
		{Name: "list_teams", Description: "List all teams in the account", InputSchema: InputSchema{Type: "object"}},
		{Name: "check_status", Description: "Check ProductPlan API status", InputSchema: InputSchema{Type: "object"}},
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
	case "list_roadmaps":
		return s.client.ListRoadmaps()
	case "get_roadmap":
		return s.client.GetRoadmap(getString("id"))
	case "get_roadmap_bars":
		return s.client.GetRoadmapBars(getString("roadmap_id"))
	case "get_roadmap_lanes":
		return s.client.GetRoadmapLanes(getString("roadmap_id"))
	case "get_roadmap_milestones":
		return s.client.GetRoadmapMilestones(getString("roadmap_id"))
	case "get_bar":
		return s.client.GetBar(getString("id"))
	case "create_bar":
		return s.client.CreateBar(args)
	case "update_bar":
		id := getString("id")
		delete(args, "id")
		return s.client.UpdateBar(id, args)
	case "list_ideas":
		return s.client.ListIdeas()
	case "get_idea":
		return s.client.GetIdea(getString("id"))
	case "create_idea":
		return s.client.CreateIdea(args)
	case "list_opportunities":
		return s.client.ListOpportunities()
	case "list_objectives":
		return s.client.ListObjectives()
	case "get_objective":
		return s.client.GetObjective(getString("id"))
	case "list_key_results":
		return s.client.ListKeyResults(getString("objective_id"))
	case "list_launches":
		return s.client.ListLaunches()
	case "get_launch":
		return s.client.GetLaunch(getString("id"))
	case "list_launch_tasks":
		return s.client.ListLaunchTasks(getString("launch_id"))
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
		// No response needed for notifications
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
	fmt.Fprintln(os.Stderr, "ProductPlan MCP Server running on stdio")
	scanner := bufio.NewScanner(os.Stdin)
	// Increase buffer size for large messages
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
			continue // Skip notifications
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

	case "launches":
		if len(subArgs) == 0 {
			result, err = client.ListLaunches()
		} else {
			result, err = client.GetLaunch(subArgs[0])
		}

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
  objectives [id]                      List objectives or get details
  key-results <objective_id>           List key results for an objective
  ideas [id]                           List ideas or get details
  launches [id]                        List launches or get details
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
