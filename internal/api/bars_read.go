package api

import (
	"context"
	"encoding/json"
	"slices"
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
	return barListing{lanes: lanes, filter: f, serverFiltered: !q.IsZero(), warnings: warnings}.format(bars), nil
}

// FormatBarsWithContext projects bars and enriches them from the lane list.
// An unparseable lane list only loses the enrichment, never the bars.
func FormatBarsWithContext(bars json.RawMessage, lanes json.RawMessage) json.RawMessage {
	return barListing{lanes: lanes}.format(bars)
}

// barListing is everything besides the bars that shapes a bar list
// response: the lane list for enrichment, the client-side filter, whether a
// q[...] filter was sent (so an empty result reads "matched the filters"
// rather than "no bars exist"), and warnings to surface.
type barListing struct {
	lanes          json.RawMessage
	filter         BarFilter
	serverFiltered bool
	warnings       []string
}

// format filters, caps and projects bars.
func (l barListing) format(bars json.RawMessage) json.RawMessage {
	items, pg, ok := unmarshalList(bars)
	if !ok {
		return bars
	}
	idx := newLaneIndex(l.lanes)
	capped, total, truncated := capList(l.filter.apply(items, idx))
	payload := map[string]any{
		"count": len(capped),
		"total": total,
		"bars":  projectBars(capped, idx),
	}
	l.annotate(payload, truncated, len(items))
	markIncomplete(payload, pg)

	output, _ := json.Marshal(payload)
	return output
}

// annotate adds the truncated, filtered, scanned and warnings markers.
func (l barListing) annotate(payload map[string]any, truncated bool, scanned int) {
	if truncated {
		payload["truncated"] = true
	}
	if l.serverFiltered || !l.filter.IsZero() {
		payload["filtered"] = true
	}
	if !l.filter.IsZero() {
		payload["scanned"] = scanned
	}
	if len(l.warnings) > 0 {
		payload["warnings"] = l.warnings
	}
}

// apply returns the bars the filter matches (all of them for a zero filter).
func (f BarFilter) apply(items []map[string]any, idx laneIndex) []map[string]any {
	if f.IsZero() {
		return items
	}
	matched := make([]map[string]any, 0, len(items))
	for _, bar := range items {
		if f.matches(bar, idx) {
			matched = append(matched, bar)
		}
	}
	return matched
}

// projectBars projects each bar to the fields agents act on.
func projectBars(bars []map[string]any, idx laneIndex) []map[string]any {
	results := make([]map[string]any, 0, len(bars))
	for _, bar := range bars {
		results = append(results, projectBar(bar, idx))
	}
	return results
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
	return f.laneMatches(bar, idx) && f.legendMatches(bar) && f.tagMatches(bar)
}

// laneMatches accepts the bar's lane by name or by numeric lane id.
func (f BarFilter) laneMatches(bar map[string]any, idx laneIndex) bool {
	if f.Lane == "" {
		return true
	}
	name, id := barLane(bar, idx)
	return strings.EqualFold(name, f.Lane) || laneIDString(id) == strings.TrimSpace(f.Lane)
}

// laneIDString renders a JSON-number lane id, or "" for anything else.
func laneIDString(id any) string {
	if n, ok := id.(float64); ok {
		return strconv.FormatFloat(n, 'f', -1, 64)
	}
	return ""
}

func (f BarFilter) legendMatches(bar map[string]any) bool {
	return f.Legend == "" || strings.EqualFold(barLegend(bar), f.Legend)
}

func (f BarFilter) tagMatches(bar map[string]any) bool {
	if f.Tag == "" {
		return true
	}
	return slices.ContainsFunc(barTags(bar), func(t string) bool { return strings.EqualFold(t, f.Tag) })
}
