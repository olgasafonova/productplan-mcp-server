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
		data, err := client.GetBar(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "bar", a.BarID)
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
		data, err := client.GetBarChildren(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "child bar")
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
		data, err := client.GetBarComments(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "comment")
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
		data, err := client.GetBarConnections(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "connection")
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
		data, err := client.GetBarLinks(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "link")
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

		var data json.RawMessage

		switch a.Action {
		case "create":
			payload := map[string]any{
				"roadmap_id": a.RoadmapID,
				"lane_id":    a.LaneID,
				"name":       a.Name,
			}
			addBarOptionalFields(payload, a)
			data, err = client.CreateBar(ctx, payload)
		case "update":
			payload := make(map[string]any)
			setIfNotEmpty(payload, "name", a.Name)
			setIfNotEmpty(payload, "lane_id", a.LaneID)
			addBarOptionalFields(payload, a)
			data, err = client.UpdateBar(ctx, a.BarID, payload)
		case "delete":
			data, err = client.DeleteBar(ctx, a.BarID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "bar", a.BarID)
	})
}

// addBarOptionalFields adds optional bar fields to the payload.
func addBarOptionalFields(payload map[string]any, a ManageBarArgs) {
	setIfNotEmpty(payload, "starts_on", a.StartsOn)
	setIfNotEmpty(payload, "ends_on", a.EndsOn)
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

func manageBarConnectionHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageBarConnectionArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		var data json.RawMessage

		switch a.Action {
		case "create":
			payload := map[string]any{"target_bar_id": a.TargetBarID}
			data, err = client.CreateBarConnection(ctx, a.BarID, payload)
		case "delete":
			data, err = client.DeleteBarConnection(ctx, a.BarID, a.ConnectionID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "connection", a.ConnectionID)
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

		var data json.RawMessage

		switch a.Action {
		case "create":
			payload := map[string]any{
				"url":  a.URL,
				"name": a.Name,
			}
			data, err = client.CreateBarLink(ctx, a.BarID, payload)
		case "delete":
			data, err = client.DeleteBarLink(ctx, a.BarID, a.LinkID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "link", a.LinkID)
	})
}
