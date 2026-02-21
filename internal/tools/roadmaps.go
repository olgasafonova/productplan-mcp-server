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
		a, err := ParseArgs[GetRoadmapArgs](args)
		if err != nil {
			return nil, err
		}
		if err = a.Validate(); err != nil {
			return nil, err
		}
		data, err := client.GetRoadmap(ctx, a.RoadmapID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "roadmap", a.RoadmapID)
	})
}

func getRoadmapBarsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetRoadmapArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetRoadmapBars(ctx, a.RoadmapID)
	})
}

func getRoadmapLanesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetRoadmapArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetRoadmapLanes(ctx, a.RoadmapID)
	})
}

func getRoadmapMilestonesHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetRoadmapArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetRoadmapMilestones(ctx, a.RoadmapID)
	})
}

func getRoadmapLegendsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetRoadmapArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetRoadmapLegends(ctx, a.RoadmapID)
	})
}

func getRoadmapCommentsHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetRoadmapArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return client.GetRoadmapComments(ctx, a.RoadmapID)
	})
}

func manageLaneHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageLaneArgs](args)
		if err != nil {
			return nil, err
		}
		if err = a.Validate(); err != nil {
			return nil, err
		}

		var data json.RawMessage

		switch a.Action {
		case "create":
			payload := map[string]any{"name": a.Name}
			if a.Color != "" {
				payload["color"] = a.Color
			}
			data, err = client.CreateLane(ctx, a.RoadmapID, payload)
		case "update":
			payload := make(map[string]any)
			if a.Name != "" {
				payload["name"] = a.Name
			}
			if a.Color != "" {
				payload["color"] = a.Color
			}
			data, err = client.UpdateLane(ctx, a.RoadmapID, a.LaneID, payload)
		case "delete":
			data, err = client.DeleteLane(ctx, a.RoadmapID, a.LaneID)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "lane", a.LaneID)
	})
}

func manageMilestoneHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[ManageMilestoneArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}

		switch a.Action {
		case "create":
			payload := map[string]any{
				"title": a.Title,
				"date":  a.Date,
			}
			return client.CreateMilestone(ctx, a.RoadmapID, payload)
		case "update":
			payload := make(map[string]any)
			if a.Title != "" {
				payload["title"] = a.Title
			}
			if a.Date != "" {
				payload["date"] = a.Date
			}
			return client.UpdateMilestone(ctx, a.RoadmapID, a.MilestoneID, payload)
		case "delete":
			return client.DeleteMilestone(ctx, a.RoadmapID, a.MilestoneID)
		}
		return nil, nil
	})
}

// getRoadmapCompleteHandler fetches roadmap details, bars, lanes, and milestones in parallel.
func getRoadmapCompleteHandler(client *api.Client) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[GetRoadmapArgs](args)
		if err != nil {
			return nil, err
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		roadmapID := a.RoadmapID

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
