package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func listLaunchesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListLaunches(ctx)
	})
}

func getLaunchHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetLaunchArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetLaunch(ctx, a.LaunchID)
	})
}

func manageLaunchHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageLaunchArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Name}
			setIfNotEmpty(payload, "date", a.Date)
			setIfNotEmpty(payload, "description", a.Description)
			return client.CreateLaunch(ctx, payload)
		case "update":
			payload := make(map[string]any)
			setIfNotEmpty(payload, "name", a.Name)
			setIfNotEmpty(payload, "date", a.Date)
			setIfNotEmpty(payload, "description", a.Description)
			return client.UpdateLaunch(ctx, a.LaunchID, payload)
		case "delete":
			return client.DeleteLaunch(ctx, a.LaunchID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}
	})
}

func getLaunchSectionsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetLaunchArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetLaunchSections(ctx, a.LaunchID)
	})
}

func getLaunchSectionHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetLaunchSectionArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetLaunchSection(ctx, a.LaunchID, a.SectionID)
	})
}

func manageLaunchSectionHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageLaunchSectionArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Name}
			return client.CreateLaunchSection(ctx, a.LaunchID, payload)
		case "update":
			payload := make(map[string]any)
			setIfNotEmpty(payload, "name", a.Name)
			return client.UpdateLaunchSection(ctx, a.LaunchID, a.SectionID, payload)
		case "delete":
			return client.DeleteLaunchSection(ctx, a.LaunchID, a.SectionID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}
	})
}

func getLaunchTasksHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetLaunchArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetLaunchTasks(ctx, a.LaunchID)
	})
}

func getLaunchTaskHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetLaunchTaskArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetLaunchTask(ctx, a.LaunchID, a.TaskID)
	})
}

func manageLaunchTaskHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageLaunchTaskArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{
				"name":       a.Name,
				"section_id": a.SectionID,
			}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.DueDate != "" {
				payload["due_date"] = a.DueDate
			}
			if a.AssigneeID != "" {
				payload["assignee_id"] = a.AssigneeID
			}
			return client.CreateLaunchTask(ctx, a.LaunchID, payload)
		case "update":
			payload := make(map[string]any)
			if a.Name != "" {
				payload["name"] = a.Name
			}
			if a.Description != "" {
				payload["description"] = a.Description
			}
			if a.DueDate != "" {
				payload["due_date"] = a.DueDate
			}
			if a.AssigneeID != "" {
				payload["assignee_id"] = a.AssigneeID
			}
			if a.Completed != nil {
				payload["completed"] = *a.Completed
			}
			return client.UpdateLaunchTask(ctx, a.LaunchID, a.TaskID, payload)
		case "delete":
			return client.DeleteLaunchTask(ctx, a.LaunchID, a.TaskID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}
	})
}
