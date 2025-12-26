package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func getBarHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		barID, err := h.RequiredString("bar_id")
		if err != nil {
			return nil, err
		}
		return client.GetBar(ctx, barID)
	})
}

func getBarChildrenHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		barID, err := h.RequiredString("bar_id")
		if err != nil {
			return nil, err
		}
		return client.GetBarChildren(ctx, barID)
	})
}

func getBarCommentsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		barID, err := h.RequiredString("bar_id")
		if err != nil {
			return nil, err
		}
		return client.GetBarComments(ctx, barID)
	})
}

func getBarConnectionsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		barID, err := h.RequiredString("bar_id")
		if err != nil {
			return nil, err
		}
		return client.GetBarConnections(ctx, barID)
	})
}

func getBarLinksHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		barID, err := h.RequiredString("bar_id")
		if err != nil {
			return nil, err
		}
		return client.GetBarLinks(ctx, barID)
	})
}

func manageBarHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}

		switch action {
		case "create":
			data := map[string]any{
				"roadmap_id": h.String("roadmap_id"),
				"lane_id":    h.String("lane_id"),
				"name":       h.String("name"),
			}
			if sd := h.String("start_date"); sd != "" {
				data["start_date"] = sd
			}
			if ed := h.String("end_date"); ed != "" {
				data["end_date"] = ed
			}
			if desc := h.String("description"); desc != "" {
				data["description"] = desc
			}
			return client.CreateBar(ctx, data)
		case "update":
			data := h.BuildData("name", "start_date", "end_date", "description")
			return client.UpdateBar(ctx, h.String("bar_id"), data)
		case "delete":
			return client.DeleteBar(ctx, h.String("bar_id"))
		}
		return nil, nil
	})
}

func manageBarCommentHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		barID, err := h.RequiredString("bar_id")
		if err != nil {
			return nil, err
		}
		body, err := h.RequiredString("body")
		if err != nil {
			return nil, err
		}
		data := map[string]any{"body": body}
		return client.CreateBarComment(ctx, barID, data)
	})
}

func manageBarConnectionHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}
		barID, err := h.RequiredString("bar_id")
		if err != nil {
			return nil, err
		}

		switch action {
		case "create":
			data := map[string]any{"target_bar_id": h.String("target_bar_id")}
			return client.CreateBarConnection(ctx, barID, data)
		case "delete":
			return client.DeleteBarConnection(ctx, barID, h.String("connection_id"))
		}
		return nil, nil
	})
}

func manageBarLinkHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}
		barID, err := h.RequiredString("bar_id")
		if err != nil {
			return nil, err
		}

		switch action {
		case "create":
			data := map[string]any{
				"url":  h.String("url"),
				"name": h.String("name"),
			}
			return client.CreateBarLink(ctx, barID, data)
		case "update":
			data := h.BuildData("url", "name")
			return client.UpdateBarLink(ctx, barID, h.String("link_id"), data)
		case "delete":
			return client.DeleteBarLink(ctx, barID, h.String("link_id"))
		}
		return nil, nil
	})
}
