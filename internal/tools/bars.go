package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// setIfNotEmpty adds a key-value pair to the payload if the value is not empty.
func setIfNotEmpty(payload map[string]any, key, value string) {
	if value != "" {
		payload[key] = value
	}
}

// setIfNotNil adds a key-value pair to the payload if the pointer is not nil.
func setIfNotNil[T any](payload map[string]any, key string, value *T) {
	if value != nil {
		payload[key] = *value
	}
}

// setIfNotEmptySlice adds a key-value pair to the payload if the slice is not empty.
func setIfNotEmptySlice[T any](payload map[string]any, key string, value []T) {
	if len(value) > 0 {
		payload[key] = value
	}
}

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
			addBarOptionalFields(payload, a)
			return client.CreateBar(ctx, payload)
		case "update":
			payload := make(map[string]any)
			setIfNotEmpty(payload, "name", a.Name)
			setIfNotEmpty(payload, "lane_id", a.LaneID)
			addBarOptionalFields(payload, a)
			return client.UpdateBar(ctx, a.BarID, payload)
		case "delete":
			return client.DeleteBar(ctx, a.BarID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}
	})
}

// addBarOptionalFields adds optional bar fields to the payload.
func addBarOptionalFields(payload map[string]any, a ManageBarArgs) {
	setIfNotEmpty(payload, "start_date", a.StartDate)
	setIfNotEmpty(payload, "end_date", a.EndDate)
	setIfNotEmpty(payload, "description", a.Description)
	setIfNotEmpty(payload, "legend_id", a.LegendID)
	setIfNotEmpty(payload, "parent_id", a.ParentID)
	setIfNotEmpty(payload, "strategic_value", a.StrategicValue)
	setIfNotEmpty(payload, "notes", a.Notes)
	setIfNotNil(payload, "percent_done", a.PercentDone)
	setIfNotNil(payload, "container", a.Container)
	setIfNotNil(payload, "parked", a.Parked)
	setIfNotNil(payload, "effort", a.Effort)
	setIfNotEmptySlice(payload, "tags", a.Tags)
	setIfNotEmptySlice(payload, "custom_text_fields", a.CustomTextFields)
	setIfNotEmptySlice(payload, "custom_dropdown_fields", a.CustomDropdownFields)
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
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}
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
			setIfNotEmpty(payload, "url", a.URL)
			setIfNotEmpty(payload, "name", a.Name)
			return client.UpdateBarLink(ctx, a.BarID, a.LinkID, payload)
		case "delete":
			return client.DeleteBarLink(ctx, a.BarID, a.LinkID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}
	})
}
