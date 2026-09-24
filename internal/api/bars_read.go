package api

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

// BarFilter narrows a roadmap's bars after they are fetched. ProductPlan's
// q[...] filter cannot match on lane, legend or tag (the bars endpoint only
// filters id, name, starts_on, ends_on, is_container, created_at,
// updated_at), but bulk bar work needs exactly those, so they are applied
// here, client-side, before the defaultListCap cap. All matches are
// case-insensitive and exact; empty fields do not filter.
type BarFilter struct {
	Lane   string // lane name, or lane id
	Legend string // legend name
	Tag    string // one of the bar's tags
}

// IsZero reports whether the filter matches every bar.
func (f BarFilter) IsZero() bool { return f.Lane == "" && f.Legend == "" && f.Tag == "" }

// GetRoadmapBars returns all bars for a roadmap, enriched with lane ids.
func (c *Client) GetRoadmapBars(ctx context.Context, id string) (json.RawMessage, error) {
	return c.GetRoadmapBarsWhere(ctx, id, Query{}, BarFilter{})
}

// GetRoadmapBarsWhere returns the roadmap's bars matching the server-side
// query q and the client-side filter f.
//
// Bars and lanes are fetched concurrently. Lanes only enrich the bars (the
// bar payload already names its lane; the lane list supplies lane_id, which
// manage_bar needs), so a lanes failure does not fail the call: it is logged
// and reported in the payload's warnings, which the tools layer lifts into
// the summary. A bars failure fails the call.
func (c *Client) GetRoadmapBarsWhere(ctx context.Context, id string, q Query, f BarFilter) (json.RawMessage, error) {
	seg, err := safeSeg("roadmap_id", id)
	if err != nil {
		return nil, err
	}

	var (
		wg                sync.WaitGroup
		bars, lanes       json.RawMessage
		barsErr, lanesErr error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		bars, barsErr = c.GetList(ctx, "/roadmaps/"+seg+"/bars", q)
	}()
	go func() {
		defer wg.Done()
		lanes, lanesErr = c.GetList(ctx, "/roadmaps/"+seg+"/lanes", Query{})
	}()
	wg.Wait()

	if barsErr != nil {
		return nil, barsErr
	}
	var warnings []string
	if lanesErr != nil {
		c.logger.Warn("lane lookup failed; returning bars without lane_id enrichment",
			logging.Endpoint("/roadmaps/"+seg+"/lanes"),
			logging.Error(lanesErr),
		)
		warnings = append(warnings, "lane lookup failed, so lane_id may be missing: "+lanesErr.Error())
		lanes = nil
	}
	return formatBars(bars, lanes, f, !q.IsZero(), warnings), nil
}

// FormatBarsWithContext projects bars and enriches them from the lane list.
// An unparseable lane list only loses the enrichment, never the bars.
func FormatBarsWithContext(bars json.RawMessage, lanes json.RawMessage) json.RawMessage {
	return formatBars(bars, lanes, BarFilter{}, false, nil)
}

// formatBars filters, caps and projects bars. serverFiltered records that a
// q[...] filter was sent, so an empty result reads "matched the filters"
// rather than "no bars exist".
func formatBars(bars, lanes json.RawMessage, f BarFilter, serverFiltered bool, warnings []string) json.RawMessage {
	items, pg, ok := unmarshalList(bars)
	if !ok {
		return bars
	}
	idx := newLaneIndex(lanes)

	matched := items
	if !f.IsZero() {
		matched = make([]map[string]any, 0, len(items))
		for _, bar := range items {
			if f.matches(bar, idx) {
				matched = append(matched, bar)
			}
		}
	}

	capped, total, truncated := capList(matched)
	results := make([]map[string]any, 0, len(capped))
	for _, bar := range capped {
		results = append(results, projectBar(bar, idx))
	}

	payload := map[string]any{
		"count": len(results),
		"total": total,
		"bars":  results,
	}
	if truncated {
		payload["truncated"] = true
	}
	if serverFiltered || !f.IsZero() {
		payload["filtered"] = true
	}
	if !f.IsZero() {
		payload["scanned"] = len(items)
	}
	if len(warnings) > 0 {
		payload["warnings"] = warnings
	}
	markIncomplete(payload, pg)

	output, _ := json.Marshal(payload)
	return output
}

// laneIndex joins bars to lanes in both directions.
type laneIndex struct {
	nameByID map[float64]string
	idByName map[string]any // lower-cased name -> id; nil when ambiguous
}

func newLaneIndex(lanes json.RawMessage) laneIndex {
	idx := laneIndex{nameByID: map[float64]string{}, idByName: map[string]any{}}
	list, _, ok := unmarshalList(lanes)
	if !ok {
		return idx
	}
	for _, lane := range list {
		id, idOK := lane["id"].(float64)
		name, nameOK := lane["name"].(string)
		if !idOK || !nameOK {
			continue
		}
		idx.nameByID[id] = name
		key := strings.ToLower(name)
		if _, dup := idx.idByName[key]; dup {
			idx.idByName[key] = nil // two lanes share a name; do not guess
			continue
		}
		idx.idByName[key] = id
	}
	return idx
}

// barLane resolves a bar's lane name and id. The live bars payload names
// the lane as a string; older or alternative shapes carry an object or a
// bare lane_id, so all three are accepted.
func barLane(bar map[string]any, idx laneIndex) (name string, id any) {
	switch v := bar["lane"].(type) {
	case string:
		name = v
	case map[string]any:
		name, _ = v["name"].(string)
		id = v["id"]
	}
	if id == nil {
		id = bar["lane_id"]
	}
	if n, ok := id.(float64); ok && name == "" {
		name = idx.nameByID[n]
	}
	if id == nil && name != "" {
		id = idx.idByName[strings.ToLower(name)]
	}
	return name, id
}

// barLegend returns the legend name from a string or {name|label} object.
func barLegend(bar map[string]any) string {
	switch v := bar["legend"].(type) {
	case string:
		return v
	case map[string]any:
		if n, ok := v["name"].(string); ok {
			return n
		}
		n, _ := v["label"].(string)
		return n
	}
	return ""
}

// barTags returns the bar's string tags.
func barTags(bar map[string]any) []string {
	raw, _ := bar["tags"].([]any)
	tags := make([]string, 0, len(raw))
	for _, t := range raw {
		if s, ok := t.(string); ok {
			tags = append(tags, s)
		}
	}
	return tags
}

// firstPresent returns the first non-nil value among keys.
func firstPresent(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

// projectBar returns the flat view of a bar an agent needs to read a roadmap
// and to act on it with manage_bar: dates, lane (name and id), legend, tags,
// progress and structure. Description and custom fields stay behind get_bar.
func projectBar(bar map[string]any, idx laneIndex) map[string]any {
	laneName, laneID := barLane(bar, idx)
	if laneName == "" {
		laneName = "Unknown"
	}
	return map[string]any{
		"id":           bar["id"],
		"name":         bar["name"],
		"starts_on":    firstPresent(bar, "starts_on", "start_date"),
		"ends_on":      firstPresent(bar, "ends_on", "end_date"),
		"lane_id":      laneID,
		"lane_name":    laneName,
		"legend":       barLegend(bar),
		"tags":         barTags(bar),
		"percent_done": bar["percent_done"],
		"is_container": bar["is_container"],
		"parked":       bar["parked"],
	}
}

// matches applies the client-side filter to one bar.
func (f BarFilter) matches(bar map[string]any, idx laneIndex) bool {
	if f.Lane != "" {
		name, id := barLane(bar, idx)
		idStr := ""
		if n, ok := id.(float64); ok {
			idStr = strconv.FormatFloat(n, 'f', -1, 64)
		}
		if !strings.EqualFold(name, f.Lane) && idStr != strings.TrimSpace(f.Lane) {
			return false
		}
	}
	if f.Legend != "" && !strings.EqualFold(barLegend(bar), f.Legend) {
		return false
	}
	if f.Tag != "" {
		found := false
		for _, t := range barTags(bar) {
			if strings.EqualFold(t, f.Tag) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
