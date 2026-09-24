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

func getBarHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBar(ctx, api.BarID(a.BarID))
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "bar", a.BarID)
	})
}

func getBarChildrenHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBarChildren(ctx, api.BarID(a.BarID))
		if err != nil {
			return nil, err
		}
		return FormatList(data, "child bar")
	})
}

func getBarCommentsHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBarComments(ctx, api.BarID(a.BarID))
		if err != nil {
			return nil, err
		}
		return FormatList(data, "comment")
	})
}

func getBarConnectionsHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBarConnections(ctx, api.BarID(a.BarID))
		if err != nil {
			return nil, err
		}
		return FormatList(data, "connection")
	})
}

func getBarLinksHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetBarArgs](func(ctx context.Context, a GetBarArgs) (json.RawMessage, error) {
		data, err := client.GetBarLinks(ctx, api.BarID(a.BarID))
		if err != nil {
			return nil, err
		}
		return FormatList(data, "link")
	})
}

// manageBarHandler creates, updates, or deletes a bar. Writes go through
// barPlanner, which translates the fields into the documented contract
// and validates names against the roadmap before anything is sent.
func manageBarHandler(client *api.Client) mcp.Handler {
	return typedHandler[ManageBarArgs](func(ctx context.Context, a ManageBarArgs) (json.RawMessage, error) {
		pl := newBarPlanner(client)
		switch a.Action {
		case "create":
			return createBar(ctx, pl, a)
		case "update":
			return updateBar(ctx, pl, a)
		case "delete":
			data, err := client.DeleteBar(ctx, api.BarID(a.BarID))
			if err != nil {
				return nil, err
			}
			return FormatAction(data, a.Action, "bar", a.BarID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}
	})
}

// createBar posts a new bar, parses its ID from the Location-style
// response, and reads it back so the caller sees what landed.
func createBar(ctx context.Context, pl *barPlanner, a ManageBarArgs) (json.RawMessage, error) {
	p, err := pl.prepare(ctx, barOp{create: true, roadmapID: a.RoadmapID, fields: a.BarFields})
	if err != nil {
		return nil, err
	}
	data, err := pl.client.CreateBar(ctx, p)
	if err != nil {
		return nil, explainWriteError(err, p)
	}
	id, err := api.ParseCreatedID(data)
	if err != nil {
		return nil, fmt.Errorf("bar was created but its ID could not be read (%w); find it with get_roadmap_bars before retrying, or a retry will create a duplicate", err)
	}
	out, err := json.Marshal(readBackCreated(ctx, pl, id, p))
	if err != nil {
		return nil, err
	}
	return json.Marshal(FormattedResponse{Summary: createdSummary(id, a.RoadmapID, p), Data: out})
}

// readBackCreated reports the new bar's ID and sent payload, plus the bar
// as the API now returns it (or why it could not be read).
func readBackCreated(ctx context.Context, pl *barPlanner, id string, p map[string]any) map[string]any {
	result := map[string]any{"id": id, "sent": p}
	if bar, err := pl.client.GetBar(ctx, api.BarID(id)); err != nil {
		result["read_back_error"] = err.Error()
	} else {
		result["bar"] = bar
	}
	return result
}

// createdSummary says where the bar landed and warns when it is parked.
func createdSummary(id, roadmapID string, p map[string]any) string {
	summary := fmt.Sprintf("Bar %s created on roadmap %s", id, roadmapID)
	if parked, _ := p["parked"].(bool); parked || p["parked"] == nil {
		summary += " (parked: not on the timeline until given dates and parked:false)"
	}
	return summary
}

// updateBar patches only the fields the caller set.
func updateBar(ctx context.Context, pl *barPlanner, a ManageBarArgs) (json.RawMessage, error) {
	p, err := pl.prepare(ctx, barOp{barID: a.BarID, roadmapID: a.RoadmapID, fields: a.BarFields})
	if err != nil {
		return nil, err
	}
	if _, err = pl.client.UpdateBar(ctx, api.BarID(a.BarID), p); err != nil {
		return nil, explainWriteError(err, p)
	}
	out, err := json.Marshal(map[string]any{"bar_id": a.BarID, "sent": p})
	if err != nil {
		return nil, err
	}
	return FormatAction(out, "update", "bar", a.BarID)
}

// barSubresourceOps bundles the client calls for connections and links on
// a bar, which support only create and delete. Unlike the other ops
// bundles, unsupported actions are rejected with an error.
type barSubresourceOps[C ~string] struct {
	resource ItemType
	create   func(ctx context.Context, barID api.BarID, payload map[string]any) (json.RawMessage, error)
	delete   func(ctx context.Context, barID api.BarID, id C) (json.RawMessage, error)
}

// run dispatches the requested action to the matching client call and
// formats the result.
func (o barSubresourceOps[C]) run(ctx context.Context, req manageRequest) (json.RawMessage, error) {
	var data json.RawMessage
	var err error
	switch req.action {
	case "create":
		data, err = o.create(ctx, api.BarID(req.parentID), req.createPayload)
	case "delete":
		data, err = o.delete(ctx, api.BarID(req.parentID), C(req.id))
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
	ops := barSubresourceOps[api.ConnectionID]{resource: "connection", create: client.CreateBarConnection, delete: client.DeleteBarConnection}
	return manageHandler(ops, func(a ManageBarConnectionArgs) manageRequest {
		return manageRequest{action: a.Action, parentID: a.BarID, id: a.ConnectionID,
			createPayload: map[string]any{"target_bar_id": a.TargetBarID}}
	})
}

// manageBarLinkHandler creates or deletes external links attached to a bar.
// The create payload always carries url and name, matching the ProductPlan
// API contract.
func manageBarLinkHandler(client *api.Client) mcp.Handler {
	ops := barSubresourceOps[api.LinkID]{resource: "link", create: client.CreateBarLink, delete: client.DeleteBarLink}
	return manageHandler(ops, func(a ManageBarLinkArgs) manageRequest {
		return manageRequest{action: a.Action, parentID: a.BarID, id: a.LinkID,
			createPayload: map[string]any{"url": a.URL, "name": a.Name}}
	})
}
