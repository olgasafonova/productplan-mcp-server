package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func listObjectivesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.ListObjectives(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "objective")
	})
}

func getObjectiveHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetObjectiveArgs](func(ctx context.Context, a GetObjectiveArgs) (json.RawMessage, error) {
		data, err := client.GetObjective(ctx, a.ObjectiveID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "objective", a.ObjectiveID)
	})
}

func listKeyResultsHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetObjectiveArgs](func(ctx context.Context, a GetObjectiveArgs) (json.RawMessage, error) {
		data, err := client.ListKeyResults(ctx, a.ObjectiveID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "key result")
	})
}

func getKeyResultHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetKeyResultArgs](func(ctx context.Context, a GetKeyResultArgs) (json.RawMessage, error) {
		data, err := client.GetKeyResult(ctx, a.ObjectiveID, a.KeyResultID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "key result", a.KeyResultID)
	})
}

func manageObjectiveHandler(client *api.Client) mcp.Handler {
	return typedHandler[ManageObjectiveArgs](func(ctx context.Context, a ManageObjectiveArgs) (json.RawMessage, error) {
		var data json.RawMessage
		var err error

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Name}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.TimeFrame != "" {
				payload["time_frame"] = a.TimeFrame
			}
			data, err = client.CreateObjective(ctx, payload)
		case "update":
			payload := make(map[string]any)
			if a.Name != "" {
				payload["name"] = a.Name
			}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			data, err = client.UpdateObjective(ctx, a.ObjectiveID, payload)
		case "delete":
			data, err = client.DeleteObjective(ctx, a.ObjectiveID)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "objective", a.ObjectiveID)
	})
}

func manageKeyResultHandler(client *api.Client) mcp.Handler {
	return typedHandler[ManageKeyResultArgs](func(ctx context.Context, a ManageKeyResultArgs) (json.RawMessage, error) {
		var data json.RawMessage
		var err error

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Name}
			if a.TargetValue != "" {
				payload["target_value"] = a.TargetValue
			}
			if a.CurrentValue != "" {
				payload["current_value"] = a.CurrentValue
			}
			data, err = client.CreateKeyResult(ctx, a.ObjectiveID, payload)
		case "update":
			payload := make(map[string]any)
			if a.Name != "" {
				payload["name"] = a.Name
			}
			if a.CurrentValue != "" {
				payload["current_value"] = a.CurrentValue
			}
			data, err = client.UpdateKeyResult(ctx, a.ObjectiveID, a.KeyResultID, payload)
		case "delete":
			data, err = client.DeleteKeyResult(ctx, a.ObjectiveID, a.KeyResultID)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "key result", a.KeyResultID)
	})
}
