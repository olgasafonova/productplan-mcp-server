package api

import (
	"encoding/json"
	"fmt"
)

// defaultListCap bounds the number of items any list tool returns by default,
// so a large collection does not blow the caller's context (HG-2 cost-lens).
// Responses carry total alongside count, plus truncated=true when clipped, so
// the caller can tell "all of them" from "the first 50 of many".
const defaultListCap = 50

// capList truncates items to defaultListCap, reporting the original total and
// whether anything was dropped.
func capList(items []map[string]any) (capped []map[string]any, total int, truncated bool) {
	total = len(items)
	if total > defaultListCap {
		return items[:defaultListCap], total, true
	}
	return items, total, false
}

// pickKeys copies the named keys from src into a fresh map.
// Missing keys are recorded as nil to preserve JSON null serialisation.
func pickKeys(src map[string]any, keys ...string) map[string]any {
	out := make(map[string]any, len(keys))
	for _, k := range keys {
		out[k] = src[k]
	}
	return out
}

// unmarshalList handles both bare-array and {"results": [...], "paging": {...}}
// envelopes. The paging block is returned (nil for bare arrays) so callers can
// report a merge that GetList stopped short at maxListPages.
// Returns nil and ok=false if neither shape decodes.
func unmarshalList(data json.RawMessage) ([]map[string]any, *paging, bool) {
	var list []map[string]any
	if err := json.Unmarshal(data, &list); err == nil {
		return list, nil, true
	}
	var wrapper struct {
		Results []map[string]any `json:"results"`
		Paging  *paging          `json:"paging"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil {
		return wrapper.Results, wrapper.Paging, true
	}
	return nil, nil, false
}

// markIncomplete records on payload that the upstream collection holds more
// records than were fetched, so neither the count nor the total is the whole
// story. Formatters in internal/tools lift the note into the summary.
func markIncomplete(payload map[string]any, pg *paging) {
	if !pg.incomplete() {
		return
	}
	payload["incomplete"] = true
	payload["record_count"] = pg.RecordCount
	payload["note"] = fmt.Sprintf("Only the first %d of %d pages were fetched (%d records upstream); narrow with filters to see the rest",
		pg.PagesFetched, pg.PageCount, pg.RecordCount)
}

// formatList projects each item via project, wraps the slice under collectionKey,
// adds count, and optionally a hint. Returns the original bytes if unmarshalling fails.
func formatList(data json.RawMessage, collectionKey string, hint string, project func(map[string]any) map[string]any) json.RawMessage {
	items, pg, ok := unmarshalList(data)
	if !ok {
		return data
	}

	capped, total, truncated := capList(items)
	results := make([]map[string]any, 0, len(capped))
	for _, item := range capped {
		results = append(results, project(item))
	}

	payload := map[string]any{
		"count":       len(results),
		"total":       total,
		collectionKey: results,
	}
	if truncated {
		payload["truncated"] = true
	}
	if hint != "" {
		payload["hint"] = hint
	}
	markIncomplete(payload, pg)

	output, _ := json.Marshal(payload)
	return output
}

// FormatRoadmapList formats roadmap list with counts and hints.
func FormatRoadmapList(data json.RawMessage) json.RawMessage {
	return formatList(data, "roadmaps", "Use get_roadmap_bars with a roadmap id to see its items",
		func(rm map[string]any) map[string]any {
			return pickKeys(rm, "id", "name", "updated_at")
		})
}

// FormatLanes formats lane list. The live GET /roadmaps/{id}/lanes result
// carries id, name, description, position, created_at and updated_at; it has
// no color field, so projecting one only ever produced "color": null.
func FormatLanes(data json.RawMessage) json.RawMessage {
	return formatList(data, "lanes", "",
		func(lane map[string]any) map[string]any {
			return pickKeys(lane, "id", "name", "description", "position")
		})
}

// FormatMilestones formats milestone list.
func FormatMilestones(data json.RawMessage) json.RawMessage {
	return formatList(data, "milestones", "",
		func(m map[string]any) map[string]any {
			return pickKeys(m, "id", "name", "date")
		})
}

// FormatObjectives formats objective list with hints.
func FormatObjectives(data json.RawMessage) json.RawMessage {
	return formatList(data, "objectives", "Use get_objective with an id for full details including key results",
		func(obj map[string]any) map[string]any {
			return pickKeys(obj, "id", "name", "status", "time_frame")
		})
}

// FormatIdeas formats idea list.
func FormatIdeas(data json.RawMessage) json.RawMessage {
	return formatList(data, "ideas", "",
		func(idea map[string]any) map[string]any {
			return pickKeys(idea, "id", "name", "channel", "opportunities_count")
		})
}

// FormatOpportunities formats opportunity list.
func FormatOpportunities(data json.RawMessage) json.RawMessage {
	return formatList(data, "opportunities", "",
		func(opp map[string]any) map[string]any {
			return pickKeys(opp, "id", "problem_statement", "workflow_status", "ideas_count")
		})
}

// FormatLaunches formats launch list.
func FormatLaunches(data json.RawMessage) json.RawMessage {
	return formatList(data, "launches", "",
		func(launch map[string]any) map[string]any {
			return pickKeys(launch, "id", "name", "date", "status")
		})
}
