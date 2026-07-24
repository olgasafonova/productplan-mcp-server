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
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBar(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "bar", a.BarID)
	})
}

func getBarChildrenHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBarChildren(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "child bar")
	})
}

func getBarCommentsHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBarComments(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "comment")
	})
}

func getBarConnectionsHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBarConnections(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "connection")
	})
}

func getBarLinksHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBarLinks(ctx, a.BarID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "link")
	})
}

func manageBarHandler(client *api.Client) mcp.Handler {
	return typedHandler[ManageBarArgs](func(ctx context.Context, a ManageBarArgs) (json.RawMessage, error) {
		var data json.RawMessage
		var err error

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

// barSubresourceOps bundles the client calls for connections and links on
// a bar, which support only create and delete. Unlike the other ops
// bundles, unsupported actions are rejected with an error.
type barSubresourceOps struct {
	resource string
	create   func(ctx context.Context, barID string, payload map[string]any) (json.RawMessage, error)
	delete   func(ctx context.Context, barID, id string) (json.RawMessage, error)
}

// run dispatches the requested action to the matching client call and
// formats the result.
func (o barSubresourceOps) run(ctx context.Context, req manageRequest) (json.RawMessage, error) {
	var data json.RawMessage
	var err error
	switch req.action {
	case "create":
		data, err = o.create(ctx, req.parentID, req.createPayload)
	case "delete":
		data, err = o.delete(ctx, req.parentID, req.id)
	default:
		return nil, fmt.Errorf("unknown action: %s", req.action)
	}
	if err != nil {
		return nil, err
	}
	return FormatAction(data, req.action, o.resource, req.id)
}

// manageBarConnectionHandler creates or deletes dependency connections
// between bars.
func manageBarConnectionHandler(client *api.Client) mcp.Handler {
	ops := barSubresourceOps{resource: "connection", create: client.CreateBarConnection, delete: client.DeleteBarConnection}
	return manageHandler(ops, func(a ManageBarConnectionArgs) manageRequest {
		return manageRequest{action: a.Action, parentID: a.BarID, id: a.ConnectionID,
			createPayload: map[string]any{"target_bar_id": a.TargetBarID}}
	})
}

// manageBarLinkHandler creates or deletes external links attached to a bar.
// The create payload always carries url and name, matching the ProductPlan
// API contract.
func manageBarLinkHandler(client *api.Client) mcp.Handler {
	ops := barSubresourceOps{resource: "link", create: client.CreateBarLink, delete: client.DeleteBarLink}
	return manageHandler(ops, func(a ManageBarLinkArgs) manageRequest {
		return manageRequest{action: a.Action, parentID: a.BarID, id: a.LinkID,
			createPayload: map[string]any{"url": a.URL, "name": a.Name}}
	})
}
