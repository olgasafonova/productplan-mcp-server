package tools

import (
	"context"
	"encoding/json"
	"fmt"
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
	return typedHandler[GetRoadmapArgs](func(ctx context.Context, a GetRoadmapArgs) (json.RawMessage, error) {
		data, err := client.GetRoadmap(ctx, a.RoadmapID)
		if err != nil {
			return nil, err
		}
		return FormatItem(data, "roadmap", a.RoadmapID)
	})
}

func getRoadmapBarsHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetRoadmapArgs](func(ctx context.Context, a GetRoadmapArgs) (json.RawMessage, error) {
		data, err := client.GetRoadmapBars(ctx, a.RoadmapID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "bar")
	})
}

func getRoadmapLanesHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetRoadmapArgs](func(ctx context.Context, a GetRoadmapArgs) (json.RawMessage, error) {
		data, err := client.GetRoadmapLanes(ctx, a.RoadmapID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "lane")
	})
}

func getRoadmapMilestonesHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetRoadmapArgs](func(ctx context.Context, a GetRoadmapArgs) (json.RawMessage, error) {
		data, err := client.GetRoadmapMilestones(ctx, a.RoadmapID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "milestone")
	})
}

func getRoadmapLegendsHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetRoadmapArgs](func(ctx context.Context, a GetRoadmapArgs) (json.RawMessage, error) {
		data, err := client.GetRoadmapLegends(ctx, a.RoadmapID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "legend")
	})
}

func getRoadmapCommentsHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetRoadmapArgs](func(ctx context.Context, a GetRoadmapArgs) (json.RawMessage, error) {
		data, err := client.GetRoadmapComments(ctx, a.RoadmapID)
		if err != nil {
			return nil, err
		}
		return FormatList(data, "comment")
	})
}

func manageLaneHandler(client *api.Client) mcp.Handler {
	return typedHandler[ManageLaneArgs](func(ctx context.Context, a ManageLaneArgs) (json.RawMessage, error) {
		var data json.RawMessage
		var err error

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
	return typedHandler[ManageMilestoneArgs](func(ctx context.Context, a ManageMilestoneArgs) (json.RawMessage, error) {
		var data json.RawMessage
		var err error

		switch a.Action {
		case "create":
			payload := map[string]any{
				"title": a.Title,
				"date":  a.Date,
			}
			data, err = client.CreateMilestone(ctx, a.RoadmapID, payload)
		case "update":
			payload := make(map[string]any)
			if a.Title != "" {
				payload["title"] = a.Title
			}
			if a.Date != "" {
				payload["date"] = a.Date
			}
			data, err = client.UpdateMilestone(ctx, a.RoadmapID, a.MilestoneID, payload)
		case "delete":
			data, err = client.DeleteMilestone(ctx, a.RoadmapID, a.MilestoneID)
		}

		if err != nil {
			return nil, err
		}
		return FormatAction(data, a.Action, "milestone", a.MilestoneID)
	})
}

// getRoadmapCompleteHandler fetches roadmap details, bars, lanes, and milestones in parallel.
// Returns partial results with per-section error reporting instead of failing on first error.
func getRoadmapCompleteHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetRoadmapArgs](func(ctx context.Context, a GetRoadmapArgs) (json.RawMessage, error) {
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

		// If roadmap itself fails, the whole request is invalid
		if roadmapErr != nil {
			return nil, roadmapErr
		}

		// Collect per-section errors instead of failing on first error
		var sectionErrors []map[string]string
		if barsErr != nil {
			sectionErrors = append(sectionErrors, map[string]string{"section": "bars", "error": barsErr.Error()})
		}
		if lanesErr != nil {
			sectionErrors = append(sectionErrors, map[string]string{"section": "lanes", "error": lanesErr.Error()})
		}
		if milestonesErr != nil {
			sectionErrors = append(sectionErrors, map[string]string{"section": "milestones", "error": milestonesErr.Error()})
		}

		// Build result with partial data
		result := map[string]any{
			"roadmap": json.RawMessage(roadmap),
		}

		// Include sections that succeeded
		if barsErr == nil {
			result["bars"] = json.RawMessage(bars)
		}
		if lanesErr == nil {
			result["lanes"] = json.RawMessage(lanes)
		}
		if milestonesErr == nil {
			result["milestones"] = json.RawMessage(milestones)
		}

		// Always include errors array (empty if all succeeded)
		result["errors"] = sectionErrors

		data, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		summary := fmt.Sprintf("Roadmap %s retrieved", roadmapID)
		if len(sectionErrors) > 0 {
			summary = fmt.Sprintf("Roadmap %s retrieved with %d section error(s)", roadmapID, len(sectionErrors))
		}

		return json.Marshal(FormattedResponse{
			Summary: summary,
			Data:    data,
		})
	})
}
