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
	return queryHandler("list_roadmaps", func(ctx context.Context, _ NoArgs, q api.Query) (json.RawMessage, error) {
		data, err := client.ListRoadmapsWhere(ctx, q)
		if err != nil {
			return nil, err
		}
		return FormatFilteredList(data, "roadmap", !q.IsZero())
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
	return queryHandler("get_roadmap_bars", func(ctx context.Context, a GetRoadmapBarsArgs, q api.Query) (json.RawMessage, error) {
		local := api.BarFilter{Lane: a.Lane, Legend: a.Legend, Tag: a.Tag}
		data, err := client.GetRoadmapBarsWhere(ctx, a.RoadmapID, q, local)
		if err != nil {
			return nil, err
		}
		return FormatFilteredList(data, "bar", !q.IsZero())
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
		out, err := FormatList(data, "legend")
		if err != nil {
			return nil, err
		}
		hint := legendHint
		if string(data) == "[]" {
			hint = noLegendHint
		}
		return appendSummary(out, hint)
	})
}

// legendHint tells the caller how legend names are used, since the API
// exposes names only (no legend IDs or hex colors).
const legendHint = "Pass one of these names as `legend` on manage_bar, bulk_update_bars, or bulk_create_bars to color a bar; legend:\"\" clears it."

// noLegendHint explains an empty legend list instead of pointing at names
// that do not exist.
const noLegendHint = "Bars on this roadmap cannot be colored until a legend is added in the ProductPlan UI"

// appendSummary adds a sentence to a FormattedResponse summary.
func appendSummary(resp json.RawMessage, sentence string) (json.RawMessage, error) {
	var fr FormattedResponse
	if err := json.Unmarshal(resp, &fr); err != nil {
		return nil, fmt.Errorf("failed to decode formatted response: %w", err)
	}
	fr.Summary += ". " + sentence
	return json.Marshal(fr)
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
