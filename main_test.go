package main

import (
	"encoding/json"
	"testing"
)

// Test Format Functions

func TestFormatLanes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "formats lanes correctly",
			input:     `[{"id": "1", "name": "Now", "color": "#FF0000"}, {"id": "2", "name": "Next", "color": "#00FF00"}]`,
			wantCount: 2,
		},
		{
			name:      "empty array",
			input:     `[]`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatLanes(json.RawMessage(tt.input))
			var wrapper struct {
				Count int                      `json:"count"`
				Lanes []map[string]interface{} `json:"lanes"`
			}
			if err := json.Unmarshal(result, &wrapper); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if wrapper.Count != tt.wantCount {
				t.Errorf("expected count %d, got %d", tt.wantCount, wrapper.Count)
			}
		})
	}
}

func TestFormatMilestones(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "formats milestones correctly",
			input:     `[{"id": "1", "name": "Q1 Release", "date": "2024-03-31"}]`,
			wantCount: 1,
		},
		{
			name:      "empty array",
			input:     `[]`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMilestones(json.RawMessage(tt.input))
			var wrapper struct {
				Count int `json:"count"`
			}
			if err := json.Unmarshal(result, &wrapper); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if wrapper.Count != tt.wantCount {
				t.Errorf("expected count %d, got %d", tt.wantCount, wrapper.Count)
			}
		})
	}
}

func TestFormatObjectives(t *testing.T) {
	input := `[{"id": "1", "name": "Increase Revenue", "progress": 0.75, "key_results": []}]`
	result := FormatObjectives(json.RawMessage(input))

	var wrapper struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(result, &wrapper); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if wrapper.Count != 1 {
		t.Fatalf("expected count 1, got %d", wrapper.Count)
	}
}

func TestFormatIdeas(t *testing.T) {
	input := `[{"id": "1", "name": "Mobile App", "status": "under_review", "vote_score": 42}]`
	result := FormatIdeas(json.RawMessage(input))

	var wrapper struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(result, &wrapper); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if wrapper.Count != 1 {
		t.Fatalf("expected count 1, got %d", wrapper.Count)
	}
}

func TestFormatLaunches(t *testing.T) {
	input := `[{"id": "1", "name": "Q1 Launch", "launch_date": "2024-03-15", "status": "scheduled"}]`
	result := FormatLaunches(json.RawMessage(input))

	var wrapper struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(result, &wrapper); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if wrapper.Count != 1 {
		t.Fatalf("expected count 1, got %d", wrapper.Count)
	}
}

// Test MCP Server

func TestNewMCPServer(t *testing.T) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	if server == nil {
		t.Fatal("expected non-nil server")
	}
	if server.client != client {
		t.Error("server client mismatch")
	}
}

func TestMCPServerGetTools(t *testing.T) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	tools := server.getTools()
	if len(tools) == 0 {
		t.Fatal("expected at least one tool")
	}

	// Check that essential tools exist
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	essentialTools := []string{
		"list_roadmaps",
		"get_roadmap",
		"get_roadmap_bars",
		"get_bar",
		"list_objectives",
		"list_ideas",
		"list_launches",
	}

	for _, name := range essentialTools {
		if !toolNames[name] {
			t.Errorf("essential tool %s not found", name)
		}
	}
}

func TestMCPServerToolsHaveDescriptions(t *testing.T) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	tools := server.getTools()
	for _, tool := range tools {
		if tool.Description == "" {
			t.Errorf("tool %s has no description", tool.Name)
		}
		if tool.InputSchema.Type != "object" {
			t.Errorf("tool %s has invalid input schema type: %s", tool.Name, tool.InputSchema.Type)
		}
	}
}

// Test JSON-RPC Request Handling

func TestHandleRequestInitialize(t *testing.T) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	}

	resp := server.handleRequest(req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("unexpected protocol version: %v", result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]string)
	if !ok {
		t.Fatal("expected serverInfo map")
	}
	if serverInfo["name"] != "productplan-mcp-server" {
		t.Errorf("unexpected server name: %s", serverInfo["name"])
	}

	// Check instructions are present
	if _, ok := result["instructions"]; !ok {
		t.Error("expected instructions field")
	}
}

func TestHandleRequestToolsList(t *testing.T) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "tools/list",
	}

	resp := server.handleRequest(req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}

	tools, ok := result["tools"].([]Tool)
	if !ok {
		t.Fatal("expected tools array")
	}

	if len(tools) == 0 {
		t.Error("expected at least one tool")
	}
}

func TestHandleRequestNotificationsInitialized(t *testing.T) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	resp := server.handleRequest(req)

	// Should return empty response for notifications
	if resp.Result != nil {
		t.Errorf("expected nil result for notification, got: %v", resp.Result)
	}
}

func TestHandleRequestToolsCallInvalidParams(t *testing.T) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`3`),
		Method:  "tools/call",
		Params:  json.RawMessage(`invalid json`),
	}

	resp := server.handleRequest(req)

	if resp.Error == nil {
		t.Fatal("expected error for invalid params")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("expected error code -32602, got %d", resp.Error.Code)
	}
}

func TestHandleRequestUnknownMethod(t *testing.T) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`4`),
		Method:  "unknown/method",
	}

	resp := server.handleRequest(req)

	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", resp.Error.Code)
	}
}

// Test APIClient initialization

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient("test-token")

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.token != "test-token" {
		t.Errorf("expected token 'test-token', got '%s'", client.token)
	}
	if client.httpClient == nil {
		t.Error("expected non-nil http client")
	}
	if client.rateLimiter == nil {
		t.Error("expected non-nil rate limiter")
	}
	if client.cache == nil {
		t.Error("expected non-nil cache")
	}
}

// Test Tool structure

func TestToolInputSchemaProperties(t *testing.T) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	tools := server.getTools()

	// Find get_roadmap_bars tool and check it has required roadmap_id
	for _, tool := range tools {
		if tool.Name == "get_roadmap_bars" {
			if len(tool.InputSchema.Required) == 0 {
				t.Error("get_roadmap_bars should have required fields")
			}
			hasRoadmapID := false
			for _, req := range tool.InputSchema.Required {
				if req == "roadmap_id" {
					hasRoadmapID = true
					break
				}
			}
			if !hasRoadmapID {
				t.Error("get_roadmap_bars should require roadmap_id")
			}
			return
		}
	}
	t.Error("get_roadmap_bars tool not found")
}

// Test Format functions handle invalid JSON gracefully

func TestFormatLanesInvalidJSON(t *testing.T) {
	// Should not panic on invalid JSON
	result := FormatLanes(json.RawMessage(`invalid`))
	// Should return the original invalid data
	if string(result) != "invalid" {
		t.Errorf("expected original data returned, got: %s", string(result))
	}
}

func TestFormatObjectivesInvalidJSON(t *testing.T) {
	result := FormatObjectives(json.RawMessage(`not json`))
	if string(result) != "not json" {
		t.Errorf("expected original data returned, got: %s", string(result))
	}
}

func TestFormatIdeasInvalidJSON(t *testing.T) {
	result := FormatIdeas(json.RawMessage(`{broken`))
	if string(result) != "{broken" {
		t.Errorf("expected original data returned, got: %s", string(result))
	}
}

// Test version variable

func TestVersionVariable(t *testing.T) {
	// Version should be set (either "dev" or injected at build time)
	if version == "" {
		t.Error("version should not be empty")
	}
}

// Benchmark tests

func BenchmarkGetTools(b *testing.B) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.getTools()
	}
}

func BenchmarkHandleInitialize(b *testing.B) {
	client := NewAPIClient("test-token")
	server := NewMCPServer(client)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.handleRequest(req)
	}
}

func BenchmarkFormatLanes(b *testing.B) {
	input := json.RawMessage(`[{"id": "1", "name": "Now", "color": "#FF0000"}, {"id": "2", "name": "Next", "color": "#00FF00"}]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatLanes(input)
	}
}
