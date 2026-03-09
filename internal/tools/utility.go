package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func checkStatusHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.CheckStatus(ctx)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "status", "api")
	})
}

func healthCheckHandler(checker HealthChecker) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[HealthCheckArgs](args)
		if err != nil {
			return nil, err
		}
		// No validation needed for optional boolean
		report := checker.Check(ctx, a.Deep)
		data, err := json.Marshal(report)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "health check", "server")
	})
}

func listUsersHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.ListUsers(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "user")
	})
}

func listTeamsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.ListTeams(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "team")
	})
}
