package api

import "encoding/json"

// pickKeys copies the named keys from src into a fresh map.
// Missing keys are recorded as nil to preserve JSON null serialisation.
func pickKeys(src map[string]any, keys ...string) map[string]any {
	out := make(map[string]any, len(keys))
	for _, k := range keys {
		out[k] = src[k]
	}
	return out
}

// unmarshalList handles both bare-array and {"results": [...]} envelopes.
// Returns nil and ok=false if neither shape decodes.
func unmarshalList(data json.RawMessage) ([]map[string]any, bool) {
	var list []map[string]any
	if err := json.Unmarshal(data, &list); err == nil {
		return list, true
	}
	var wrapper struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil {
		return wrapper.Results, true
	}
	return nil, false
}

// formatList projects each item via project, wraps the slice under collectionKey,
// adds count, and optionally a hint. Returns the original bytes if unmarshalling fails.
func formatList(data json.RawMessage, collectionKey string, hint string, project func(map[string]any) map[string]any) json.RawMessage {
	items, ok := unmarshalList(data)
	if !ok {
		return data
	}

	results := make([]map[string]any, 0, len(items))
	for _, item := range items {
		results = append(results, project(item))
	}

	payload := map[string]any{
		"count":       len(results),
		collectionKey: results,
	}
	if hint != "" {
		payload["hint"] = hint
	}

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

// buildLaneLookup returns a map from lane id to lane name for fast joins.
func buildLaneLookup(laneList []map[string]any) map[float64]string {
	lookup := make(map[float64]string, len(laneList))
	for _, lane := range laneList {
		id, ok := lane["id"].(float64)
		if !ok {
			continue
		}
		name, ok := lane["name"].(string)
		if !ok {
			continue
		}
		lookup[id] = name
	}
	return lookup
}

// projectBar returns a flat projection of a raw bar enriched with lane_name.
func projectBar(bar map[string]any, laneLookup map[float64]string) map[string]any {
	laneID, _ := bar["lane_id"].(float64)
	laneName := laneLookup[laneID]
	if laneName == "" {
		laneName = "Unknown"
	}
	return map[string]any{
		"id":         bar["id"],
		"name":       bar["name"],
		"start_date": bar["start_date"],
		"end_date":   bar["end_date"],
		"lane_id":    bar["lane_id"],
		"lane_name":  laneName,
	}
}

// FormatBarsWithContext enriches bars with lane names.
func FormatBarsWithContext(bars json.RawMessage, lanes json.RawMessage) json.RawMessage {
	var barList []map[string]any
	var laneList []map[string]any

	if err := json.Unmarshal(bars, &barList); err != nil {
		return bars
	}
	if err := json.Unmarshal(lanes, &laneList); err != nil {
		return bars
	}

	laneLookup := buildLaneLookup(laneList)
	results := make([]map[string]any, 0, len(barList))
	for _, bar := range barList {
		results = append(results, projectBar(bar, laneLookup))
	}

	output, _ := json.Marshal(map[string]any{
		"count": len(results),
		"bars":  results,
	})
	return output
}

// FormatLanes formats lane list.
func FormatLanes(data json.RawMessage) json.RawMessage {
	return formatList(data, "lanes", "",
		func(lane map[string]any) map[string]any {
			return pickKeys(lane, "id", "name", "color")
		})
}

// FormatMilestones formats milestone list.
func FormatMilestones(data json.RawMessage) json.RawMessage {
	return formatList(data, "milestones", "",
		func(m map[string]any) map[string]any {
			return pickKeys(m, "id", "name", "date")
		})
}

// FormatLegends formats legend list (bar colors).
func FormatLegends(data json.RawMessage) json.RawMessage {
	return formatList(data, "legends", "Use legend_id when creating or updating bars to set their color",
		func(legend map[string]any) map[string]any {
			return pickKeys(legend, "id", "label", "color")
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
