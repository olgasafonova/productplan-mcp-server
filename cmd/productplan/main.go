// ProductPlan MCP Server - Model Context Protocol server for ProductPlan API.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/cli"
	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
	"github.com/olgasafonova/productplan-mcp-server/internal/tools"
)

// Injected at build time via ldflags.
var version = "dev"

func main() {
	os.Exit(run())
}

// isHelpArg returns true when the first arg requests help.
func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

// isServerArg returns true when the first arg selects the MCP server mode.
// Empty args (mode default) also yield true; pass the first positional arg or "".
func isServerArg(arg string) bool {
	return arg == "" || arg == "serve" || arg == "mcp"
}

// requireToken reads PRODUCTPLAN_API_TOKEN and reports a friendly error to stderr if missing.
func requireToken() (string, bool) {
	token := os.Getenv("PRODUCTPLAN_API_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: PRODUCTPLAN_API_TOKEN environment variable is required")
		return "", false
	}
	return token, true
}

func run() int {
	args := os.Args[1:]

	if len(args) > 0 && isHelpArg(args[0]) {
		printHelp()
		return 0
	}

	apiToken, ok := requireToken()
	if !ok {
		return 1
	}

	logger := logging.New(logging.LevelInfo)
	client, err := api.New(api.Config{Token: apiToken, Logger: logger})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create API client: %v\n", err)
		return 1
	}

	first := ""
	if len(args) > 0 {
		first = args[0]
	}
	if isServerArg(first) {
		return runMCPServer(client, logger)
	}
	return runCLI(client, args)
}

func runMCPServer(client *api.Client, logger logging.Logger) int {
	// Create MCP registry and register tools
	registry := mcp.NewRegistry()
	tools.RegisterAll(registry, tools.Config{
		Client:        client,
		HealthChecker: newHealthChecker(client, version),
	})

	// Create and run MCP server
	server := mcp.NewServer("productplan", version, registry,
		mcp.WithLogger(logger),
	)

	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		logger.Error("MCP server error", logging.Error(err))
		return 1
	}
	return 0
}

func runCLI(client *api.Client, args []string) int {
	c := cli.New(client, cli.Config{
		Version: version,
	})
	return c.Run(args)
}

func printHelp() {
	fmt.Printf(`ProductPlan MCP Server v%s

Usage:
  productplan [serve|mcp]              Start MCP server (default)
  productplan <command> [args]         Run CLI command

For CLI commands, run: productplan help
`, version)
}

// healthChecker implements tools.HealthChecker.
type healthChecker struct {
	client  *api.Client
	version string
}

func newHealthChecker(client *api.Client, version string) *healthChecker {
	return &healthChecker{client: client, version: version}
}

func (h *healthChecker) Check(ctx context.Context, deep bool) any {
	report := map[string]any{
		"status":  "healthy",
		"version": h.version,
	}
	if deep {
		status, err := h.client.CheckStatus(ctx)
		if err != nil {
			report["api"] = map[string]any{"status": "error", "error": err.Error()}
		} else {
			report["api"] = status
		}
	}
	return report
}
