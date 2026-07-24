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

// manageObjectiveHandler creates, updates, or deletes objectives.
// Validation guarantees a non-empty name on create; time_frame is only
// settable at creation, matching the ProductPlan API.
func manageObjectiveHandler(client *api.Client) mcp.Handler {
	ops := topLevelOps{resource: "objective", create: client.CreateObjective, update: client.UpdateObjective, delete: client.DeleteObjective}
	return manageHandler(ops, func(a ManageObjectiveArgs) manageRequest {
		return manageRequest{
			action: a.Action, id: a.ObjectiveID,
			createPayload: buildPayload(nil,
				fieldCheck{a.Name, "name"}, fieldCheck{a.Description, "description"}, fieldCheck{a.TimeFrame, "time_frame"}),
			updatePayload: buildPayload(nil, fieldCheck{a.Name, "name"}, fieldCheck{a.Description, "description"}),
		}
	})
}

// manageKeyResultHandler creates, updates, or deletes key results under an
// objective. The create payload always names the key result; target_value
// is only settable at creation, matching the ProductPlan API.
func manageKeyResultHandler(client *api.Client) mcp.Handler {
	ops := parentScopedOps{resource: "key result", create: client.CreateKeyResult, update: client.UpdateKeyResult, delete: client.DeleteKeyResult}
	return manageHandler(ops, func(a ManageKeyResultArgs) manageRequest {
		return manageRequest{
			action: a.Action, parentID: a.ObjectiveID, id: a.KeyResultID,
			createPayload: buildPayload(map[string]any{"name": a.Name},
				fieldCheck{a.TargetValue, "target_value"}, fieldCheck{a.CurrentValue, "current_value"}),
			updatePayload: buildPayload(nil, fieldCheck{a.Name, "name"}, fieldCheck{a.CurrentValue, "current_value"}),
		}
	})
}
