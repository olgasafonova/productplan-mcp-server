package mcp

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	err := NewError(ErrInvalidParams, "invalid params")
	if err.Code != ErrInvalidParams {
		t.Errorf("expected code %d, got %d", ErrInvalidParams, err.Code)
	}
	if err.Message != "invalid params" {
		t.Errorf("expected message 'invalid params', got %q", err.Message)
	}
}

func TestNewTextResult(t *testing.T) {
	result := NewTextResult("hello")
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	if result.Content[0].Type != "text" {
		t.Errorf("expected type 'text', got %q", result.Content[0].Type)
	}
	if result.Content[0].Text != "hello" {
		t.Errorf("expected text 'hello', got %q", result.Content[0].Text)
	}
	if result.IsError {
		t.Error("expected IsError to be false")
	}
}

func TestNewErrorResult(t *testing.T) {
	result := NewErrorResult(errors.New("something went wrong"))
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	if result.Content[0].Text != "Error: something went wrong" {
		t.Errorf("expected error text, got %q", result.Content[0].Text)
	}
	if !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestJSONRPCRequestUnmarshal(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{"name":"test"}}`
	var req JSONRPCRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc '2.0', got %q", req.JSONRPC)
	}
	if req.ID != float64(1) {
		t.Errorf("expected id 1, got %v", req.ID)
	}
	if req.Method != "tools/list" {
		t.Errorf("expected method 'tools/list', got %q", req.Method)
	}
}

func TestJSONRPCResponseMarshal(t *testing.T) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  map[string]string{"status": "ok"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(data, &parsed)

	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("expected jsonrpc '2.0', got %v", parsed["jsonrpc"])
	}
	if parsed["id"] != float64(1) {
		t.Errorf("expected id 1, got %v", parsed["id"])
	}
}

func TestJSONRPCResponseWithError(t *testing.T) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Error:   NewError(ErrMethodNotFound, "method not found"),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(data, &parsed)

	errorObj := parsed["error"].(map[string]any)
	if errorObj["code"] != float64(ErrMethodNotFound) {
		t.Errorf("expected code %d, got %v", ErrMethodNotFound, errorObj["code"])
	}
}

func TestToolMarshal(t *testing.T) {
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"id": {Type: "string", Description: "The ID"},
			},
			Required: []string{"id"},
		},
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(data, &parsed)

	if parsed["name"] != "test_tool" {
		t.Errorf("expected name 'test_tool', got %v", parsed["name"])
	}
	if parsed["description"] != "A test tool" {
		t.Errorf("expected description 'A test tool', got %v", parsed["description"])
	}
}

func TestToolCallParamsUnmarshal(t *testing.T) {
	input := `{"name":"get_roadmap","arguments":{"roadmap_id":"123"}}`
	var params ToolCallParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if params.Name != "get_roadmap" {
		t.Errorf("expected name 'get_roadmap', got %q", params.Name)
	}
	if params.Arguments["roadmap_id"] != "123" {
		t.Errorf("expected roadmap_id '123', got %v", params.Arguments["roadmap_id"])
	}
}

func TestInitializeResultMarshal(t *testing.T) {
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		ServerInfo:      ServerInfo{Name: "test", Version: "1.0.0"},
		Capabilities:    Capabilities{Tools: map[string]any{}},
		Instructions:    "Test instructions",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(data, &parsed)

	if parsed["protocolVersion"] != ProtocolVersion {
		t.Errorf("expected protocolVersion %q, got %v", ProtocolVersion, parsed["protocolVersion"])
	}
	if parsed["instructions"] != "Test instructions" {
		t.Errorf("expected instructions, got %v", parsed["instructions"])
	}
}

func TestToolsListResultMarshal(t *testing.T) {
	result := ToolsListResult{
		Tools: []Tool{
			{Name: "tool1", Description: "Tool 1"},
			{Name: "tool2", Description: "Tool 2"},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(data, &parsed)

	tools := parsed["tools"].([]any)
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
}

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		code int
		name string
	}{
		{ErrParseError, "ParseError"},
		{ErrInvalidRequest, "InvalidRequest"},
		{ErrMethodNotFound, "MethodNotFound"},
		{ErrInvalidParams, "InvalidParams"},
		{ErrInternalError, "InternalError"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code >= 0 {
				t.Errorf("error code %s should be negative", tt.name)
			}
		})
	}
}
