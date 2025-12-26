package api

import "encoding/json"

// FormatRoadmapList formats roadmap list with counts and hints.
func FormatRoadmapList(data json.RawMessage) json.RawMessage {
	var roadmaps []map[string]any
	if err := json.Unmarshal(data, &roadmaps); err != nil {
		return data
	}

	results := make([]map[string]any, 0, len(roadmaps))
	for _, rm := range roadmaps {
		results = append(results, map[string]any{
			"id":         rm["id"],
			"name":       rm["name"],
			"updated_at": rm["updated_at"],
		})
	}

	output, _ := json.Marshal(map[string]any{
		"count":    len(results),
		"roadmaps": results,
		"hint":     "Use get_roadmap_bars with a roadmap id to see its items",
	})
	return output
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

	// Build lane lookup
	laneLookup := make(map[float64]string)
	for _, lane := range laneList {
		if id, ok := lane["id"].(float64); ok {
			if name, ok := lane["name"].(string); ok {
				laneLookup[id] = name
			}
		}
	}

	results := make([]map[string]any, 0, len(barList))
	for _, bar := range barList {
		laneID, _ := bar["lane_id"].(float64)
		laneName := laneLookup[laneID]
		if laneName == "" {
			laneName = "Unknown"
		}

		results = append(results, map[string]any{
			"id":         bar["id"],
			"name":       bar["name"],
			"start_date": bar["start_date"],
			"end_date":   bar["end_date"],
			"lane_id":    bar["lane_id"],
			"lane_name":  laneName,
		})
	}

	output, _ := json.Marshal(map[string]any{
		"count": len(results),
		"bars":  results,
	})
	return output
}

// FormatLanes formats lane list.
func FormatLanes(data json.RawMessage) json.RawMessage {
	var lanes []map[string]any
	if err := json.Unmarshal(data, &lanes); err != nil {
		return data
	}

	results := make([]map[string]any, 0, len(lanes))
	for _, lane := range lanes {
		results = append(results, map[string]any{
			"id":    lane["id"],
			"name":  lane["name"],
			"color": lane["color"],
		})
	}

	output, _ := json.Marshal(map[string]any{
		"count": len(results),
		"lanes": results,
	})
	return output
}

// FormatMilestones formats milestone list.
func FormatMilestones(data json.RawMessage) json.RawMessage {
	var milestones []map[string]any
	if err := json.Unmarshal(data, &milestones); err != nil {
		return data
	}

	results := make([]map[string]any, 0, len(milestones))
	for _, m := range milestones {
		results = append(results, map[string]any{
			"id":   m["id"],
			"name": m["name"],
			"date": m["date"],
		})
	}

	output, _ := json.Marshal(map[string]any{
		"count":      len(results),
		"milestones": results,
	})
	return output
}

// FormatObjectives formats objective list with hints.
func FormatObjectives(data json.RawMessage) json.RawMessage {
	var objectives []map[string]any
	if err := json.Unmarshal(data, &objectives); err != nil {
		return data
	}

	results := make([]map[string]any, 0, len(objectives))
	for _, obj := range objectives {
		results = append(results, map[string]any{
			"id":         obj["id"],
			"name":       obj["name"],
			"status":     obj["status"],
			"time_frame": obj["time_frame"],
		})
	}

	output, _ := json.Marshal(map[string]any{
		"count":      len(results),
		"objectives": results,
		"hint":       "Use get_objective with an id for full details including key results",
	})
	return output
}

// FormatIdeas formats idea list.
func FormatIdeas(data json.RawMessage) json.RawMessage {
	var wrapper struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		// Try as array
		var ideas []map[string]any
		if err := json.Unmarshal(data, &ideas); err != nil {
			return data
		}
		wrapper.Results = ideas
	}

	results := make([]map[string]any, 0, len(wrapper.Results))
	for _, idea := range wrapper.Results {
		results = append(results, map[string]any{
			"id":                  idea["id"],
			"name":                idea["name"],
			"channel":             idea["channel"],
			"opportunities_count": idea["opportunities_count"],
		})
	}

	output, _ := json.Marshal(map[string]any{
		"count": len(results),
		"ideas": results,
	})
	return output
}

// FormatOpportunities formats opportunity list.
func FormatOpportunities(data json.RawMessage) json.RawMessage {
	var wrapper struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		// Try as array
		var opportunities []map[string]any
		if err := json.Unmarshal(data, &opportunities); err != nil {
			return data
		}
		wrapper.Results = opportunities
	}

	results := make([]map[string]any, 0, len(wrapper.Results))
	for _, opp := range wrapper.Results {
		results = append(results, map[string]any{
			"id":                opp["id"],
			"problem_statement": opp["problem_statement"],
			"workflow_status":   opp["workflow_status"],
			"ideas_count":       opp["ideas_count"],
		})
	}

	output, _ := json.Marshal(map[string]any{
		"count":         len(results),
		"opportunities": results,
	})
	return output
}

// FormatLaunches formats launch list.
func FormatLaunches(data json.RawMessage) json.RawMessage {
	var launches []map[string]any
	if err := json.Unmarshal(data, &launches); err != nil {
		return data
	}

	results := make([]map[string]any, 0, len(launches))
	for _, launch := range launches {
		results = append(results, map[string]any{
			"id":     launch["id"],
			"name":   launch["name"],
			"date":   launch["date"],
			"status": launch["status"],
		})
	}

	output, _ := json.Marshal(map[string]any{
		"count":    len(results),
		"launches": results,
	})
	return output
}
