package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

// Server is an MCP server that communicates over stdio.
type Server struct {
	name         string
	version      string
	instructions string
	registry     *Registry
	logger       logging.Logger
	reader       io.Reader
	writer       io.Writer
}

// ServerOption configures a Server.
type ServerOption func(*Server)

// WithInstructions sets the server instructions.
func WithInstructions(instructions string) ServerOption {
	return func(s *Server) {
		s.instructions = instructions
	}
}

// WithLogger sets the server logger.
func WithLogger(logger logging.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithIO sets custom reader/writer for testing.
func WithIO(reader io.Reader, writer io.Writer) ServerOption {
	return func(s *Server) {
		s.reader = reader
		s.writer = writer
	}
}

// NewServer creates a new MCP server.
func NewServer(name, version string, registry *Registry, opts ...ServerOption) *Server {
	s := &Server{
		name:     name,
		version:  version,
		registry: registry,
		logger:   logging.Nop(),
		reader:   os.Stdin,
		writer:   os.Stdout,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Run starts the server, reading requests from stdin and writing responses to stdout.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("MCP server starting",
		logging.F("name", s.name),
		logging.F("version", s.version),
	)
	fmt.Fprintf(os.Stderr, "%s MCP Server v%s running on stdio\n", s.name, s.version)

	scanner := bufio.NewScanner(s.reader)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.logger.Debug("failed to parse request",
				logging.Error(err),
			)
			continue
		}

		s.logger.Debug("received request",
			logging.F("method", req.Method),
			logging.F("id", req.ID),
		)

		resp := s.handleRequest(ctx, req)
		if resp.JSONRPC == "" {
			continue
		}

		respJSON, err := json.Marshal(resp)
		if err != nil {
			s.logger.Error("failed to marshal response",
				logging.Error(err),
			)
			continue
		}

		_, _ = fmt.Fprintln(s.writer, string(respJSON))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

func (s *Server) handleRequest(ctx context.Context, req JSONRPCRequest) JSONRPCResponse {
	resp := JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case "initialize":
		resp.Result = InitializeResult{
			ProtocolVersion: ProtocolVersion,
			ServerInfo:      ServerInfo{Name: s.name, Version: s.version},
			Capabilities:    Capabilities{Tools: map[string]any{}},
			Instructions:    s.instructions,
		}

	case "notifications/initialized":
		// No response needed for notifications
		return JSONRPCResponse{}

	case "tools/list":
		resp.Result = ToolsListResult{Tools: s.registry.Tools()}

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = NewError(ErrInvalidParams, err.Error())
			return resp
		}

		s.logger.Debug("calling tool",
			logging.Tool(params.Name),
		)

		result, err := s.registry.Call(ctx, params.Name, params.Arguments)
		if err != nil {
			s.logger.Debug("tool call failed",
				logging.Tool(params.Name),
				logging.Error(err),
			)
			resp.Result = NewErrorResult(err)
		} else {
			resp.Result = NewTextResult(string(result))
		}

	default:
		resp.Error = NewError(ErrMethodNotFound, "Method not found: "+req.Method)
	}

	return resp
}

// ProcessRequest handles a single request for testing.
func (s *Server) ProcessRequest(ctx context.Context, req JSONRPCRequest) JSONRPCResponse {
	return s.handleRequest(ctx, req)
}
