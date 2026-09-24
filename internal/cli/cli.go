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
// and fetches a single item otherwise. The command-line argument becomes the
// endpoint's typed ID; the endpoint validates it.
func listOrGet[ID ~string](list func(context.Context) (json.RawMessage, error), get func(context.Context, ID) (json.RawMessage, error)) command {
	return command{run: func(ctx context.Context, args []string) (json.RawMessage, error) {
		if len(args) == 0 {
			return list(ctx)
		}
		return get(ctx, ID(args[0]))
	}}
}

// withID builds a command that requires an ID as its first argument.
func withID[ID ~string](usage string, get func(context.Context, ID) (json.RawMessage, error)) command {
	return command{usage: usage, run: func(ctx context.Context, args []string) (json.RawMessage, error) {
		return get(ctx, ID(args[0]))
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

  PRODUCTPLAN_CACHE_TTL                Read cache lifetime, e.g. 60s or 2m (default 60s, 0 disables)

MCP server (run with no command): 50 tools
  - 35 READ tools, with filters on list tools (name, dates, lane, legend, tag, sort)
  - 12 action-based manage_* WRITE tools
  - 3 bulk bar tools: bulk_update_bars, bulk_create_bars, bulk_delete_bars
  - Bars are colored by legend name (see get_roadmap_legends)

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
