package tools

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

func listRoadmapsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		data, err := client.ListRoadmaps(ctx)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "roadmap")
	})
}

func getRoadmapHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		roadmapID, err := h.RequiredString("roadmap_id")
		if err != nil {
			return nil, err
		}
		data, err := client.GetRoadmap(ctx, roadmapID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "roadmap", roadmapID)
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

		var data json.RawMessage
		laneID := h.String("lane_id")

		switch action {
		case "create":
			payload := map[string]any{"name": h.String("name")}
			if c := h.String("color"); c != "" {
				payload["color"] = c
			}
			data, err = client.CreateLane(ctx, roadmapID, payload)
		case "update":
			payload := h.BuildData("name", "color")
			data, err = client.UpdateLane(ctx, roadmapID, laneID, payload)
		case "delete":
			data, err = client.DeleteLane(ctx, roadmapID, laneID)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, action, "lane", laneID)
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

// getRoadmapCompleteHandler fetches roadmap details, bars, lanes, and milestones in parallel.
func getRoadmapCompleteHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		h := mcp.NewArgHelper(args)
		roadmapID, err := h.RequiredString("roadmap_id")
		if err != nil {
			return nil, err
		}

		// Fetch all data in parallel
		var wg sync.WaitGroup
		var roadmap, bars, lanes, milestones json.RawMessage
		var roadmapErr, barsErr, lanesErr, milestonesErr error

		wg.Add(4)

		go func() {
			defer wg.Done()
			roadmap, roadmapErr = client.GetRoadmap(ctx, roadmapID)
		}()

		go func() {
			defer wg.Done()
			bars, barsErr = client.GetRoadmapBars(ctx, roadmapID)
		}()

		go func() {
			defer wg.Done()
			lanes, lanesErr = client.GetRoadmapLanes(ctx, roadmapID)
		}()

		go func() {
			defer wg.Done()
			milestones, milestonesErr = client.GetRoadmapMilestones(ctx, roadmapID)
		}()

		wg.Wait()

		// Return first error encountered
		if roadmapErr != nil {
			return nil, roadmapErr
		}
		if barsErr != nil {
			return nil, barsErr
		}
		if lanesErr != nil {
			return nil, lanesErr
		}
		if milestonesErr != nil {
			return nil, milestonesErr
		}

		// Combine results into a single response
		result := map[string]json.RawMessage{
			"roadmap":    roadmap,
			"bars":       bars,
			"lanes":      lanes,
			"milestones": milestones,
		}

		return json.Marshal(result)
	})
}
