// Package mcp provides the MCP (Model Context Protocol) server implementation.
package mcp

import "encoding/json"

// ProtocolVersion is the MCP protocol version supported by this server.
const ProtocolVersion = "2025-11-25"

// JSONRPCRequest represents an incoming JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Standard JSON-RPC error codes.
const (
	ErrParseError     = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternalError  = -32603
)

// NewError creates a new RPC error.
func NewError(code int, message string) *RPCError {
	return &RPCError{Code: code, Message: message}
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema defines the JSON Schema for tool inputs.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property defines a single property in the input schema.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolContent represents content returned from a tool call.
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolResult represents the result of a tool call.
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// NewTextResult creates a successful text result.
func NewTextResult(text string) ToolResult {
	return ToolResult{
		Content: []ToolContent{{Type: "text", Text: text}},
	}
}

// NewErrorResult creates an error result.
func NewErrorResult(err error) ToolResult {
	return ToolResult{
		Content: []ToolContent{{Type: "text", Text: "Error: " + err.Error()}},
		IsError: true,
	}
}

// ToolCallParams represents the parameters for a tools/call request.
type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// InitializeResult represents the result of an initialize request.
type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	ServerInfo      ServerInfo     `json:"serverInfo"`
	Capabilities    Capabilities   `json:"capabilities"`
	Instructions    string         `json:"instructions,omitempty"`
}

// ServerInfo contains server identification information.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Capabilities describes what the server supports.
type Capabilities struct {
	Tools map[string]any `json:"tools,omitempty"`
}

// ToolsListResult represents the result of a tools/list request.
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}
