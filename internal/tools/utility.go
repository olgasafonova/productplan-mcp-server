package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func checkStatusHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.CheckStatus(ctx)
	})
}

func healthCheckHandler(checker HealthChecker) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		deep := h.Bool("deep")
		report := checker.Check(ctx, deep)
		return json.Marshal(report)
	})
}
