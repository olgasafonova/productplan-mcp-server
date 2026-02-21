package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func listIdeasHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListIdeas(ctx)
	})
}

func getIdeaHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetIdeaArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetIdea(ctx, a.IdeaID)
	})
}

func listOpportunitiesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListOpportunities(ctx)
	})
}

func getOpportunityHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetOpportunityArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetOpportunity(ctx, a.OpportunityID)
	})
}

func listIdeaFormsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListIdeaForms(ctx)
	})
}

func getIdeaFormHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetIdeaFormArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetIdeaForm(ctx, a.FormID)
	})
}

func listAllCustomersHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListAllCustomers(ctx)
	})
}

func listAllTagsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListAllTags(ctx)
	})
}

func manageIdeaHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageIdeaArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Title}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.Status != "" {
				payload["status"] = a.Status
			}
			return client.CreateIdea(ctx, payload)
		case "update":
			payload := make(map[string]any)
			if a.Title != "" {
				payload["name"] = a.Title
			}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.Status != "" {
				payload["status"] = a.Status
			}
			return client.UpdateIdea(ctx, a.IdeaID, payload)
		}
		return nil, nil
	})
}

func manageOpportunityHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageOpportunityArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{"problem_statement": a.ProblemStatement}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.WorkflowStatus != "" {
				payload["workflow_status"] = a.WorkflowStatus
			}
			return client.CreateOpportunity(ctx, payload)
		case "update":
			payload := make(map[string]any)
			if a.ProblemStatement != "" {
				payload["problem_statement"] = a.ProblemStatement
			}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.WorkflowStatus != "" {
				payload["workflow_status"] = a.WorkflowStatus
			}
			return client.UpdateOpportunity(ctx, a.OpportunityID, payload)
		}
		return nil, nil
	})
}
