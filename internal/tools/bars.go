package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func getBarHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetBarArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetBar(ctx, a.BarID)
	})
}

func getBarChildrenHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetBarArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetBarChildren(ctx, a.BarID)
	})
}

func getBarCommentsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetBarArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetBarComments(ctx, a.BarID)
	})
}

func getBarConnectionsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetBarArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetBarConnections(ctx, a.BarID)
	})
}

func getBarLinksHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetBarArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetBarLinks(ctx, a.BarID)
	})
}

func manageBarHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageBarArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{
				"roadmap_id": a.RoadmapID,
				"lane_id":    a.LaneID,
				"name":       a.Name,
			}
			if a.StartDate != "" {
				payload["start_date"] = a.StartDate
			}
			if a.EndDate != "" {
				payload["end_date"] = a.EndDate
			}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			return client.CreateBar(ctx, payload)
		case "update":
			payload := make(map[string]any)
			if a.Name != "" {
				payload["name"] = a.Name
			}
			if a.StartDate != "" {
				payload["start_date"] = a.StartDate
			}
			if a.EndDate != "" {
				payload["end_date"] = a.EndDate
			}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			return client.UpdateBar(ctx, a.BarID, payload)
		case "delete":
			return client.DeleteBar(ctx, a.BarID)
		}
		return nil, nil
	})
}

func manageBarCommentHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageBarCommentArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		payload := map[string]any{"body": a.Body}
		return client.CreateBarComment(ctx, a.BarID, payload)
	})
}

func manageBarConnectionHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageBarConnectionArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{"target_bar_id": a.TargetBarID}
			return client.CreateBarConnection(ctx, a.BarID, payload)
		case "delete":
			return client.DeleteBarConnection(ctx, a.BarID, a.ConnectionID)
		}
		return nil, nil
	})
}

func manageBarLinkHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageBarLinkArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{
				"url":  a.URL,
				"name": a.Name,
			}
			return client.CreateBarLink(ctx, a.BarID, payload)
		case "update":
			payload := make(map[string]any)
			if a.URL != "" {
				payload["url"] = a.URL
			}
			if a.Name != "" {
				payload["name"] = a.Name
			}
			return client.UpdateBarLink(ctx, a.BarID, a.LinkID, payload)
		case "delete":
			return client.DeleteBarLink(ctx, a.BarID, a.LinkID)
		}
		return nil, nil
	})
}
