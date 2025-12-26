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

func run() int {
	args := os.Args[1:]

	// Handle help (before token validation)
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		printHelp()
		return 0
	}

	// Validate API token
	apiToken := os.Getenv("PRODUCTPLAN_API_TOKEN")
	if apiToken == "" {
		fmt.Fprintln(os.Stderr, "Error: PRODUCTPLAN_API_TOKEN environment variable is required")
		return 1
	}

	// Create logger (writes to stderr, stdout reserved for MCP protocol)
	logger := logging.New(logging.LevelInfo)

	// Create API client
	client, err := api.New(api.Config{
		Token:  apiToken,
		Logger: logger,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create API client: %v\n", err)
		return 1
	}

	// MCP server mode (default or explicit)
	if len(args) == 0 || args[0] == "serve" || args[0] == "mcp" {
		return runMCPServer(client, logger)
	}

	// CLI mode
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
