package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func listIdeasHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.ListIdeas(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "idea")
	})
}

func getIdeaHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetIdeaArgs](func(ctx context.Context, a GetIdeaArgs) (json.RawMessage, error) {
		data, err := client.GetIdea(ctx, a.IdeaID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "idea", a.IdeaID)
	})
}

func listOpportunitiesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.ListOpportunities(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "opportunity")
	})
}

func getOpportunityHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetOpportunityArgs](func(ctx context.Context, a GetOpportunityArgs) (json.RawMessage, error) {
		data, err := client.GetOpportunity(ctx, a.OpportunityID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "opportunity", a.OpportunityID)
	})
}

func listIdeaFormsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.ListIdeaForms(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "idea form")
	})
}

func getIdeaFormHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetIdeaFormArgs](func(ctx context.Context, a GetIdeaFormArgs) (json.RawMessage, error) {
		data, err := client.GetIdeaForm(ctx, a.FormID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "idea form", a.FormID)
	})
}

func listAllCustomersHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.ListAllCustomers(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "customer")
	})
}

func listAllTagsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.ListAllTags(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "tag")
	})
}

// manageIdeaHandler creates or updates ideas. Validation guarantees a
// non-empty title on create, so one payload serves both actions.
func manageIdeaHandler(client *api.Client) mcp.Handler {
	ops := topLevelOps{resource: "idea", create: client.CreateIdea, update: client.UpdateIdea}
	return manageHandler(ops, func(a ManageIdeaArgs) manageRequest {
		payload := buildPayload(nil,
			fieldCheck{a.Title, "name"}, fieldCheck{a.Description, "description"}, fieldCheck{a.Status, "status"})
		return manageRequest{action: a.Action, id: a.IdeaID, createPayload: payload, updatePayload: payload}
	})
}

// manageOpportunityHandler creates or updates opportunities. Validation
// guarantees a non-empty problem statement on create, so one payload serves
// both actions.
func manageOpportunityHandler(client *api.Client) mcp.Handler {
	ops := topLevelOps{resource: "opportunity", create: client.CreateOpportunity, update: client.UpdateOpportunity}
	return manageHandler(ops, func(a ManageOpportunityArgs) manageRequest {
		payload := buildPayload(nil, fieldCheck{a.ProblemStatement, "problem_statement"},
			fieldCheck{a.Description, "description"}, fieldCheck{a.WorkflowStatus, "workflow_status"})
		return manageRequest{action: a.Action, id: a.OpportunityID, createPayload: payload, updatePayload: payload}
	})
}
