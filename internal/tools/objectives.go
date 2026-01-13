package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func listObjectivesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListObjectives(ctx)
	})
}

func getObjectiveHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetObjectiveArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetObjective(ctx, a.ObjectiveID)
	})
}

func listKeyResultsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetObjectiveArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.ListKeyResults(ctx, a.ObjectiveID)
	})
}

func getKeyResultHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetKeyResultArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetKeyResult(ctx, a.ObjectiveID, a.KeyResultID)
	})
}

func manageObjectiveHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageObjectiveArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Name}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.TimeFrame != "" {
				payload["time_frame"] = a.TimeFrame
			}
			return client.CreateObjective(ctx, payload)
		case "update":
			payload := make(map[string]any)
			if a.Name != "" {
				payload["name"] = a.Name
			}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			return client.UpdateObjective(ctx, a.ObjectiveID, payload)
		case "delete":
			return client.DeleteObjective(ctx, a.ObjectiveID)
		}
		return nil, nil
	})
}

func manageKeyResultHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageKeyResultArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Name}
			if a.TargetValue != "" {
				payload["target_value"] = a.TargetValue
			}
			if a.CurrentValue != "" {
				payload["current_value"] = a.CurrentValue
			}
			return client.CreateKeyResult(ctx, a.ObjectiveID, payload)
		case "update":
			payload := make(map[string]any)
			if a.Name != "" {
				payload["name"] = a.Name
			}
			if a.CurrentValue != "" {
				payload["current_value"] = a.CurrentValue
			}
			return client.UpdateKeyResult(ctx, a.ObjectiveID, a.KeyResultID, payload)
		case "delete":
			return client.DeleteKeyResult(ctx, a.ObjectiveID, a.KeyResultID)
		}
		return nil, nil
	})
}
