package tools

import (
	"context"
	"encoding/json"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func listRoadmapsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		return client.ListRoadmaps(ctx)
	})
}

func getRoadmapHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		roadmapID, err := h.RequiredString("roadmap_id")
		if err != nil {
			return nil, err
		}
		return client.GetRoadmap(ctx, roadmapID)
	})
}

func getRoadmapBarsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		roadmapID, err := h.RequiredString("roadmap_id")
		if err != nil {
			return nil, err
		}
		return client.GetRoadmapBars(ctx, roadmapID)
	})
}

func getRoadmapLanesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		roadmapID, err := h.RequiredString("roadmap_id")
		if err != nil {
			return nil, err
		}
		return client.GetRoadmapLanes(ctx, roadmapID)
	})
}

func getRoadmapMilestonesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		roadmapID, err := h.RequiredString("roadmap_id")
		if err != nil {
			return nil, err
		}
		return client.GetRoadmapMilestones(ctx, roadmapID)
	})
}

func manageLaneHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}
		roadmapID, err := h.RequiredString("roadmap_id")
		if err != nil {
			return nil, err
		}

		switch action {
		case "create":
			data := map[string]any{"name": h.String("name")}
			if c := h.String("color"); c != "" {
				data["color"] = c
			}
			return client.CreateLane(ctx, roadmapID, data)
		case "update":
			data := h.BuildData("name", "color")
			return client.UpdateLane(ctx, roadmapID, h.String("lane_id"), data)
		case "delete":
			return client.DeleteLane(ctx, roadmapID, h.String("lane_id"))
		}
		return nil, nil
	})
}

func manageMilestoneHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		action, err := h.RequiredString("action")
		if err != nil {
			return nil, err
		}
		roadmapID, err := h.RequiredString("roadmap_id")
		if err != nil {
			return nil, err
		}

		switch action {
		case "create":
			data := map[string]any{
				"name": h.String("name"),
				"date": h.String("date"),
			}
			return client.CreateMilestone(ctx, roadmapID, data)
		case "update":
			data := h.BuildData("name", "date")
			return client.UpdateMilestone(ctx, roadmapID, h.String("milestone_id"), data)
		case "delete":
			return client.DeleteMilestone(ctx, roadmapID, h.String("milestone_id"))
		}
		return nil, nil
	})
}
