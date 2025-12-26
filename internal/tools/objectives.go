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
		h := mcp.NewArgHelper(args)
		objectiveID, err := h.RequiredString("objective_id")
		if err != nil {
			return nil, err
		}
		return client.GetObjective(ctx, objectiveID)
	})
}

func listKeyResultsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		objectiveID, err := h.RequiredString("objective_id")
		if err != nil {
			return nil, err
		}
		return client.ListKeyResults(ctx, objectiveID)
	})
}

func manageObjectiveHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}

		switch action {
		case "create":
			data := map[string]any{"name": h.String("name")}
			if desc := h.String("description"); desc != "" {
				data["description"] = desc
			}
			if tf := h.String("time_frame"); tf != "" {
				data["time_frame"] = tf
			}
			return client.CreateObjective(ctx, data)
		case "update":
			data := h.BuildData("name", "description")
			return client.UpdateObjective(ctx, h.String("objective_id"), data)
		case "delete":
			return client.DeleteObjective(ctx, h.String("objective_id"))
		}
		return nil, nil
	})
}

func manageKeyResultHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}
		objectiveID, err := h.RequiredString("objective_id")
		if err != nil {
			return nil, err
		}

		switch action {
		case "create":
			data := map[string]any{"name": h.String("name")}
			if tv := h.String("target_value"); tv != "" {
				data["target_value"] = tv
			}
			if cv := h.String("current_value"); cv != "" {
				data["current_value"] = cv
			}
			return client.CreateKeyResult(ctx, objectiveID, data)
		case "update":
			data := h.BuildData("name", "current_value")
			return client.UpdateKeyResult(ctx, objectiveID, h.String("key_result_id"), data)
		case "delete":
			return client.DeleteKeyResult(ctx, objectiveID, h.String("key_result_id"))
		}
		return nil, nil
	})
}
