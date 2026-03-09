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
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetIdeaArgs](args)
		if err != nil {
			return nil, err
		}
		if err = a.Validate(); err != nil {
			return nil, err
		}
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
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetOpportunityArgs](args)
		if err != nil {
			return nil, err
		}
		if err = a.Validate(); err != nil {
			return nil, err
		}
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
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetIdeaFormArgs](args)
		if err != nil {
			return nil, err
		}
		if err = a.Validate(); err != nil {
			return nil, err
		}
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

func manageIdeaHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageIdeaArgs](args)
		if err != nil {
			return nil, err
		}
		if err = a.Validate(); err != nil {
			return nil, err
		}

		var data json.RawMessage

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Title}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.Status != "" {
				payload["status"] = a.Status
			}
			data, err = client.CreateIdea(ctx, payload)
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
			data, err = client.UpdateIdea(ctx, a.IdeaID, payload)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "idea", a.IdeaID)
	})
}

func manageOpportunityHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageOpportunityArgs](args)
		if err != nil {
			return nil, err
		}
		if err = a.Validate(); err != nil {
			return nil, err
		}

		var data json.RawMessage

		switch a.Action {
		case "create":
			payload := map[string]any{"problem_statement": a.ProblemStatement}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.WorkflowStatus != "" {
				payload["workflow_status"] = a.WorkflowStatus
			}
			data, err = client.CreateOpportunity(ctx, payload)
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
			data, err = client.UpdateOpportunity(ctx, a.OpportunityID, payload)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "opportunity", a.OpportunityID)
	})
}
