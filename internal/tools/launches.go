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
		data, err := client.ListLaunches(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "launch")
	})
}

func getLaunchHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetLaunchArgs](func(ctx context.Context, a GetLaunchArgs) (json.RawMessage, error) {
		data, err := client.GetLaunch(ctx, a.LaunchID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "launch", a.LaunchID)
	})
}

func manageLaunchHandler(client *api.Client) mcp.Handler {
	return typedHandler[ManageLaunchArgs](func(ctx context.Context, a ManageLaunchArgs) (json.RawMessage, error) {
		var data json.RawMessage
		var err error

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Name}
			setIfNotEmpty(payload, "date", a.Date)
			setIfNotEmpty(payload, "description", a.Description)
			data, err = client.CreateLaunch(ctx, payload)
		case "update":
			payload := make(map[string]any)
			setIfNotEmpty(payload, "name", a.Name)
			setIfNotEmpty(payload, "date", a.Date)
			setIfNotEmpty(payload, "description", a.Description)
			data, err = client.UpdateLaunch(ctx, a.LaunchID, payload)
		case "delete":
			data, err = client.DeleteLaunch(ctx, a.LaunchID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "launch", a.LaunchID)
	})
}

func getLaunchSectionsHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetLaunchArgs](func(ctx context.Context, a GetLaunchArgs) (json.RawMessage, error) {
		data, err := client.GetLaunchSections(ctx, a.LaunchID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "section")
	})
}

func getLaunchSectionHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetLaunchSectionArgs](func(ctx context.Context, a GetLaunchSectionArgs) (json.RawMessage, error) {
		data, err := client.GetLaunchSection(ctx, a.LaunchID, a.SectionID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "section", a.SectionID)
	})
}

func manageLaunchSectionHandler(client *api.Client) mcp.Handler {
	return typedHandler[ManageLaunchSectionArgs](func(ctx context.Context, a ManageLaunchSectionArgs) (json.RawMessage, error) {
		var data json.RawMessage
		var err error

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Name}
			data, err = client.CreateLaunchSection(ctx, a.LaunchID, payload)
		case "update":
			payload := make(map[string]any)
			setIfNotEmpty(payload, "name", a.Name)
			data, err = client.UpdateLaunchSection(ctx, a.LaunchID, a.SectionID, payload)
		case "delete":
			data, err = client.DeleteLaunchSection(ctx, a.LaunchID, a.SectionID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "section", a.SectionID)
	})
}

func getLaunchTasksHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetLaunchArgs](func(ctx context.Context, a GetLaunchArgs) (json.RawMessage, error) {
		data, err := client.GetLaunchTasks(ctx, a.LaunchID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "task")
	})
}

func getLaunchTaskHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetLaunchTaskArgs](func(ctx context.Context, a GetLaunchTaskArgs) (json.RawMessage, error) {
		data, err := client.GetLaunchTask(ctx, a.LaunchID, a.TaskID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "task", a.TaskID)
	})
}

func manageLaunchTaskHandler(client *api.Client) mcp.Handler {
	return typedHandler[ManageLaunchTaskArgs](func(ctx context.Context, a ManageLaunchTaskArgs) (json.RawMessage, error) {
		var data json.RawMessage
		var err error

		switch a.Action {
		case "create":
			payload := map[string]any{
				"name":       a.Name,
				"section_id": a.SectionID,
			}
			setIfNotEmpty(payload, "description", a.Description)
			setIfNotEmpty(payload, "due_date", a.DueDate)
			setIfNotEmpty(payload, "assigned_user_id", a.AssignedUserID)
			setIfNotEmpty(payload, "status", a.Status)
			data, err = client.CreateLaunchTask(ctx, a.LaunchID, payload)
		case "update":
			payload := make(map[string]any)
			setIfNotEmpty(payload, "name", a.Name)
			setIfNotEmpty(payload, "description", a.Description)
			setIfNotEmpty(payload, "due_date", a.DueDate)
			setIfNotEmpty(payload, "assigned_user_id", a.AssignedUserID)
			setIfNotEmpty(payload, "status", a.Status)
			data, err = client.UpdateLaunchTask(ctx, a.LaunchID, a.TaskID, payload)
		case "delete":
			data, err = client.DeleteLaunchTask(ctx, a.LaunchID, a.TaskID)
		default:
			return nil, fmt.Errorf("unknown action: %s", a.Action)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "task", a.TaskID)
	})
}
