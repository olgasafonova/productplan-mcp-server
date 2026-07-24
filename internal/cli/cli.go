// Package cli provides command-line interface functionality.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

// Config holds CLI configuration.
type Config struct {
	Version string
	Output  io.Writer
	Error   io.Writer
}

// CLI handles command-line operations.
type CLI struct {
	client  *api.Client
	cfg     Config
	output  io.Writer
	errOut  io.Writer
	version string
}

// New creates a new CLI instance.
func New(client *api.Client, cfg Config) *CLI {
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}
	errOut := cfg.Error
	if errOut == nil {
		errOut = os.Stderr
	}
	return &CLI{
		client:  client,
		cfg:     cfg,
		output:  output,
		errOut:  errOut,
		version: cfg.Version,
	}
}

// apiCall fetches data from the ProductPlan API for one CLI invocation.
type apiCall func(ctx context.Context, args []string) (json.RawMessage, error)

// command describes one CLI subcommand: how to run it and, when a required
// first argument is missing, the usage line to print instead.
type command struct {
	usage string // non-empty when the first argument is required
	run   apiCall
}

// listOrGet builds a command that lists the collection when no ID is given
// and fetches a single item otherwise.
func listOrGet(list func(context.Context) (json.RawMessage, error), get func(context.Context, string) (json.RawMessage, error)) command {
	return command{run: func(ctx context.Context, args []string) (json.RawMessage, error) {
		if len(args) == 0 {
			return list(ctx)
		}
		return get(ctx, args[0])
	}}
}

// withID builds a command that requires an ID as its first argument.
func withID(usage string, get func(context.Context, string) (json.RawMessage, error)) command {
	return command{usage: usage, run: func(ctx context.Context, args []string) (json.RawMessage, error) {
		return get(ctx, args[0])
	}}
}

// commands maps each subcommand name to its implementation.
func (c *CLI) commands() map[string]command {
	cl := c.client
	return map[string]command{
		"roadmaps":      listOrGet(cl.ListRoadmaps, cl.GetRoadmap),
		"bars":          withID("Usage: productplan bars <roadmap_id>", cl.GetRoadmapBars),
		"lanes":         withID("Usage: productplan lanes <roadmap_id>", cl.GetRoadmapLanes),
		"milestones":    withID("Usage: productplan milestones <roadmap_id>", cl.GetRoadmapMilestones),
		"objectives":    listOrGet(cl.ListObjectives, cl.GetObjective),
		"key-results":   withID("Usage: productplan key-results <objective_id>", cl.ListKeyResults),
		"ideas":         listOrGet(cl.ListIdeas, cl.GetIdea),
		"launches":      listOrGet(cl.ListLaunches, cl.GetLaunch),
		"opportunities": listOrGet(cl.ListOpportunities, cl.GetOpportunity),
		"status": {run: func(ctx context.Context, _ []string) (json.RawMessage, error) {
			return cl.CheckStatus(ctx)
		}},
	}
}

// Run executes the CLI with given arguments.
// Returns exit code (0 for success, 1 for error).
func (c *CLI) Run(args []string) int {
	if len(args) < 1 {
		c.PrintUsage()
		return 1
	}

	cmd, ok := c.commands()[args[0]]
	if !ok {
		c.PrintUsage()
		return 1
	}

	subArgs := args[1:]
	if cmd.usage != "" && len(subArgs) == 0 {
		_, _ = fmt.Fprintln(c.errOut, cmd.usage)
		return 1
	}

	result, err := cmd.run(context.Background(), subArgs)
	if err != nil {
		_, _ = fmt.Fprintf(c.errOut, "Error: %v\n", err)
		return 1
	}

	c.printJSON(result)
	return 0
}

// PrintUsage outputs usage information.
func (c *CLI) PrintUsage() {
	_, _ = fmt.Fprintf(c.output, `ProductPlan CLI & MCP Server v%s

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
  opportunities [id]                   List opportunities or get details
  launches [id]                        List launches or get details
  status                               Check API status

Environment:
  PRODUCTPLAN_API_TOKEN                Your ProductPlan API token (required)

Design (v4.2):
  - 24 granular READ tools (no params needed for lists)
  - 12 consolidated WRITE tools (action-based)
  - Bar relationships: children, comments, connections, links
  - Discovery module: ideas CRUD, customers, tags, opportunities, idea forms
  - Enriched responses (bars include lane names)
  - Clear tool descriptions for AI decision-making

`, c.version)
}

func (c *CLI) printJSON(data json.RawMessage) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		// If indentation fails, print raw
		_, _ = fmt.Fprintln(c.output, string(data))
		return
	}
	_, _ = fmt.Fprintln(c.output, pretty.String())
}
