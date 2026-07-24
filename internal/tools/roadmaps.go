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

// manageLaneHandler creates, updates, or deletes lanes on a roadmap.
// The create payload always names the lane; updates send only set fields.
func manageLaneHandler(client *api.Client) mcp.Handler {
	ops := parentScopedOps{resource: "lane", create: client.CreateLane, update: client.UpdateLane, delete: client.DeleteLane}
	return manageHandler(ops, func(a ManageLaneArgs) manageRequest {
		create := buildPayload(map[string]any{"name": a.Name}, fieldCheck{a.Color, "color"})
		update := buildPayload(nil, fieldCheck{a.Name, "name"}, fieldCheck{a.Color, "color"})
		return manageRequest{action: a.Action, parentID: a.RoadmapID, id: a.LaneID, createPayload: create, updatePayload: update}
	})
}

// manageMilestoneHandler creates, updates, or deletes roadmap milestones.
// The create payload always carries title and date, even when empty, to
// match the ProductPlan API contract; updates send only set fields.
func manageMilestoneHandler(client *api.Client) mcp.Handler {
	ops := parentScopedOps{resource: "milestone", create: client.CreateMilestone, update: client.UpdateMilestone, delete: client.DeleteMilestone}
	return manageHandler(ops, func(a ManageMilestoneArgs) manageRequest {
		return manageRequest{
			action: a.Action, parentID: a.RoadmapID, id: a.MilestoneID,
			createPayload: map[string]any{"title": a.Title, "date": a.Date},
			updatePayload: buildPayload(nil, fieldCheck{a.Title, "title"}, fieldCheck{a.Date, "date"}),
		}
	})
}

// roadmapSection is one optional part of a get_roadmap_complete response,
// fetched in parallel with the others.
type roadmapSection struct {
	name  string
	fetch func(ctx context.Context, roadmapID string) (json.RawMessage, error)
	data  json.RawMessage
	err   error
}

// getRoadmapCompleteHandler fetches roadmap details, bars, lanes, and milestones in parallel.
// Returns partial results with per-section error reporting instead of failing on first error.
func getRoadmapCompleteHandler(client *api.Client) mcp.Handler {
	return typedHandler[GetRoadmapArgs](func(ctx context.Context, a GetRoadmapArgs) (json.RawMessage, error) {
		roadmapID := a.RoadmapID
		sections := []*roadmapSection{
			{name: "bars", fetch: client.GetRoadmapBars},
			{name: "lanes", fetch: client.GetRoadmapLanes},
			{name: "milestones", fetch: client.GetRoadmapMilestones},
		}

		// Fetch the roadmap and every section in parallel.
		var wg sync.WaitGroup
		var roadmap json.RawMessage
		var roadmapErr error
		wg.Add(len(sections) + 1)
		go func() {
			defer wg.Done()
			roadmap, roadmapErr = client.GetRoadmap(ctx, roadmapID)
		}()
		for _, s := range sections {
			go func() {
				defer wg.Done()
				s.data, s.err = s.fetch(ctx, roadmapID)
			}()
		}
		wg.Wait()

		// If the roadmap itself fails, the whole request is invalid.
		if roadmapErr != nil {
			return nil, roadmapErr
		}

		// Include sections that succeeded; collect errors for the rest
		// instead of failing on the first one.
		result := map[string]any{"roadmap": roadmap}
		var sectionErrors []map[string]string
		for _, s := range sections {
			if s.err != nil {
				sectionErrors = append(sectionErrors, map[string]string{"section": s.name, "error": s.err.Error()})
				continue
			}
			result[s.name] = s.data
		}

		// Always include the errors array (empty if all succeeded).
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
