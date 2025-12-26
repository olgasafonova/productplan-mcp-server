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
		h := mcp.NewArgHelper(args)
		ideaID, err := h.RequiredString("idea_id")
		if err != nil {
			return nil, err
		}
		return client.GetIdea(ctx, ideaID)
	})
}

func getIdeaCustomersHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		ideaID, err := h.RequiredString("idea_id")
		if err != nil {
			return nil, err
		}
		return client.GetIdeaCustomers(ctx, ideaID)
	})
}

func getIdeaTagsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		ideaID, err := h.RequiredString("idea_id")
		if err != nil {
			return nil, err
		}
		return client.GetIdeaTags(ctx, ideaID)
	})
}

func listOpportunitiesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListOpportunities(ctx)
	})
}

func getOpportunityHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		opportunityID, err := h.RequiredString("opportunity_id")
		if err != nil {
			return nil, err
		}
		return client.GetOpportunity(ctx, opportunityID)
	})
}

func listIdeaFormsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListIdeaForms(ctx)
	})
}

func getIdeaFormHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		formID, err := h.RequiredString("form_id")
		if err != nil {
			return nil, err
		}
		return client.GetIdeaForm(ctx, formID)
	})
}

func manageIdeaHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}

		switch action {
		case "create":
			data := map[string]any{"name": h.String("title")}
			if desc := h.String("description"); desc != "" {
				data["description"] = desc
			}
			if status := h.String("status"); status != "" {
				data["status"] = status
			}
			return client.CreateIdea(ctx, data)
		case "update":
			data := make(map[string]any)
			if n := h.String("title"); n != "" {
				data["name"] = n
			}
			if desc := h.String("description"); desc != "" {
				data["description"] = desc
			}
			if status := h.String("status"); status != "" {
				data["status"] = status
			}
			return client.UpdateIdea(ctx, h.String("idea_id"), data)
		}
		return nil, nil
	})
}

func manageIdeaCustomerHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}
		ideaID, err := h.RequiredString("idea_id")
		if err != nil {
			return nil, err
		}

		switch action {
		case "add":
			data := map[string]any{"customer_id": h.String("customer_id")}
			return client.AddIdeaCustomer(ctx, ideaID, data)
		case "remove":
			return client.RemoveIdeaCustomer(ctx, ideaID, h.String("customer_id"))
		}
		return nil, nil
	})
}

func manageIdeaTagHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}
		ideaID, err := h.RequiredString("idea_id")
		if err != nil {
			return nil, err
		}

		switch action {
		case "add":
			data := map[string]any{"tag_id": h.String("tag_id")}
			return client.AddIdeaTag(ctx, ideaID, data)
		case "remove":
			return client.RemoveIdeaTag(ctx, ideaID, h.String("tag_id"))
		}
		return nil, nil
	})
}

func manageOpportunityHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}

		switch action {
		case "create":
			data := map[string]any{"problem_statement": h.String("problem_statement")}
			if desc := h.String("description"); desc != "" {
				data["description"] = desc
			}
			if status := h.String("workflow_status"); status != "" {
				data["workflow_status"] = status
			}
			return client.CreateOpportunity(ctx, data)
		case "update":
			data := h.BuildData("problem_statement", "description", "workflow_status")
			return client.UpdateOpportunity(ctx, h.String("opportunity_id"), data)
		case "delete":
			return client.DeleteOpportunity(ctx, h.String("opportunity_id"))
		}
		return nil, nil
	})
}
