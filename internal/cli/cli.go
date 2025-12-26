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

// Run executes the CLI with given arguments.
// Returns exit code (0 for success, 1 for error).
func (c *CLI) Run(args []string) int {
	if len(args) < 1 {
		c.PrintUsage()
		return 1
	}

	cmd := args[0]
	subArgs := args[1:]
	ctx := context.Background()

	var result json.RawMessage
	var err error

	switch cmd {
	case "roadmaps":
		if len(subArgs) == 0 {
			result, err = c.client.ListRoadmaps(ctx)
		} else {
			result, err = c.client.GetRoadmap(ctx, subArgs[0])
		}

	case "bars":
		if len(subArgs) == 0 {
			_, _ = fmt.Fprintln(c.errOut, "Usage: productplan bars <roadmap_id>")
			return 1
		}
		result, err = c.client.GetRoadmapBars(ctx, subArgs[0])

	case "lanes":
		if len(subArgs) == 0 {
			_, _ = fmt.Fprintln(c.errOut, "Usage: productplan lanes <roadmap_id>")
			return 1
		}
		result, err = c.client.GetRoadmapLanes(ctx, subArgs[0])

	case "milestones":
		if len(subArgs) == 0 {
			_, _ = fmt.Fprintln(c.errOut, "Usage: productplan milestones <roadmap_id>")
			return 1
		}
		result, err = c.client.GetRoadmapMilestones(ctx, subArgs[0])

	case "objectives":
		if len(subArgs) == 0 {
			result, err = c.client.ListObjectives(ctx)
		} else {
			result, err = c.client.GetObjective(ctx, subArgs[0])
		}

	case "key-results":
		if len(subArgs) == 0 {
			_, _ = fmt.Fprintln(c.errOut, "Usage: productplan key-results <objective_id>")
			return 1
		}
		result, err = c.client.ListKeyResults(ctx, subArgs[0])

	case "ideas":
		if len(subArgs) == 0 {
			result, err = c.client.ListIdeas(ctx)
		} else {
			result, err = c.client.GetIdea(ctx, subArgs[0])
		}

	case "launches":
		if len(subArgs) == 0 {
			result, err = c.client.ListLaunches(ctx)
		} else {
			result, err = c.client.GetLaunch(ctx, subArgs[0])
		}

	case "opportunities":
		if len(subArgs) == 0 {
			result, err = c.client.ListOpportunities(ctx)
		} else {
			result, err = c.client.GetOpportunity(ctx, subArgs[0])
		}

	case "status":
		result, err = c.client.CheckStatus(ctx)

	default:
		c.PrintUsage()
		return 1
	}

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
