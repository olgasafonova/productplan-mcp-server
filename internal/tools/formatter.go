// Package tools provides response formatting utilities for MCP tool handlers.
package tools

import (
	"encoding/json"
	"fmt"
)

// FormattedResponse wraps API responses with AI-friendly summaries.
type FormattedResponse struct {
	Summary string          `json:"summary"`
	Data    json.RawMessage `json:"data"`
}

// defaultListCap bounds the number of items a count-only list response returns
// by default, so a large collection (e.g. a long comment thread) does not blow
// the caller's context (HG-2 cost-lens). When clipped, the summary reports the
// true total so the caller knows to refine.
const defaultListCap = 50

// FormatList creates a response with a count summary, capping the returned
// items at defaultListCap. It accepts three list shapes:
//
//   - a bare JSON array;
//   - the ProductPlan {"results": [...], "paging": {...}} envelope, whose
//     results become the data (so the cap and the zero-result message apply
//     to it as well);
//   - a payload already projected by internal/api's formatters
//     ({"count", "total", <items>, ...}), which is kept as-is and summarised
//     from its own count/total/truncated/incomplete/warnings fields.
//
// Anything else (a single object) is returned unchanged.
func FormatList(data json.RawMessage, itemType ItemType) (json.RawMessage, error) {
	return FormatFilteredList(data, itemType, false)
}

// FormatFilteredList is FormatList for a list the caller narrowed with
// filters: an empty result then reads "No <items> matched the filters", so
// the agent can tell "nothing exists" from "my filter excluded everything".
func FormatFilteredList(data json.RawMessage, itemType ItemType, filtered bool) (json.RawMessage, error) {
	return listView{itemType: itemType, filtered: filtered}.format(data)
}

// listView carries what a list summary needs to know about the list: the
// item noun and whether the caller filtered it.
type listView struct {
	itemType ItemType
	filtered bool
}

// format dispatches on the three accepted list shapes.
func (v listView) format(data json.RawMessage) (json.RawMessage, error) {
	var items []any
	if err := json.Unmarshal(data, &items); err == nil {
		return v.formatArray(data, items, nil)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return data, nil
	}
	if raw, results, ok := envelopeResults(obj); ok {
		return v.formatArray(raw, results, envelopeNotes(obj["paging"]))
	}
	if isProjectedList(obj) {
		return v.formatProjected(data)
	}
	return data, nil
}

// envelopeResults extracts the results array of a {"results": [...]}
// envelope, reporting false when obj is not one.
func envelopeResults(obj map[string]json.RawMessage) (json.RawMessage, []any, bool) {
	raw, ok := obj["results"]
	if !ok {
		return nil, nil, false
	}
	var items []any
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, nil, false
	}
	return raw, items, true
}

// isProjectedList reports whether obj is an internal/api projected list,
// recognised by carrying both count and total.
func isProjectedList(obj map[string]json.RawMessage) bool {
	_, hasCount := obj["count"]
	_, hasTotal := obj["total"]
	return hasCount && hasTotal
}

// envelopeNotes reports a paged envelope that api.GetList stopped short of
// the last page, so a raw-envelope list is as honest as a projected one.
func envelopeNotes(raw json.RawMessage) []string {
	var pg struct {
		RecordCount  int `json:"record_count"`
		PageCount    int `json:"page_count"`
		PagesFetched int `json:"pages_fetched"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &pg) != nil {
		return nil
	}
	if pg.PagesFetched == 0 || pg.PagesFetched >= pg.PageCount {
		return nil
	}
	return []string{fmt.Sprintf("Only the first %d of %d pages were fetched (%d records upstream)", pg.PagesFetched, pg.PageCount, pg.RecordCount)}
}

// projectedList is the summary-relevant subset of an internal/api formatter
// payload.
type projectedList struct {
	Count      int      `json:"count"`
	Total      int      `json:"total"`
	Truncated  bool     `json:"truncated"`
	Incomplete bool     `json:"incomplete"`
	Note       string   `json:"note"`
	Warnings   []string `json:"warnings"`
	Filtered   bool     `json:"filtered"`
}

// notes lists the incompleteness note and warnings to append to a summary.
func (p projectedList) notes() []string {
	var notes []string
	if p.Incomplete && p.Note != "" {
		notes = append(notes, p.Note)
	}
	for _, w := range p.Warnings {
		notes = append(notes, "Warning: "+w)
	}
	return notes
}

// listCounts is what a list summary sentence is built from.
type listCounts struct {
	count, total int
	truncated    bool
	filtered     bool
}

// summary renders the one-line list summary: clipped, empty (filtered or
// not), or found.
func (v listView) summary(c listCounts) string {
	switch {
	case c.truncated:
		return fmt.Sprintf("Showing first %d of %d %s (refine to narrow)", c.count, c.total, v.itemType.plural(c.total))
	case c.count == 0 && c.filtered:
		return fmt.Sprintf("No %s matched the filters", v.itemType.plural(0))
	case c.count == 0:
		return fmt.Sprintf("No %s found", v.itemType.plural(0))
	}
	return fmt.Sprintf("Found %d %s", c.count, v.itemType.plural(c.count))
}

// formatProjected wraps an already-projected payload, deriving the summary
// from its own fields. The payload itself is not re-capped: the api layer
// capped it at the same defaultListCap and reports truncated/total.
func (v listView) formatProjected(data json.RawMessage) (json.RawMessage, error) {
	var p projectedList
	if err := json.Unmarshal(data, &p); err != nil {
		return data, nil
	}
	counts := listCounts{count: p.Count, total: p.Total, truncated: p.Truncated, filtered: p.Filtered || v.filtered}
	return v.respond(data, counts, p.notes())
}

// respond marshals the list response: the summary sentence with any notes
// appended, and the data.
func (v listView) respond(data json.RawMessage, c listCounts, notes []string) (json.RawMessage, error) {
	summary := v.summary(c)
	for _, n := range notes {
		summary += ". " + n
	}
	return json.Marshal(FormattedResponse{Summary: summary, Data: data})
}

// formatArray caps a decoded array and builds its summary.
func (v listView) formatArray(data json.RawMessage, items []any, notes []string) (json.RawMessage, error) {
	total := len(items)
	truncated := total > defaultListCap
	if truncated {
		items = items[:defaultListCap]
		// Re-marshal the capped slice so Data carries only what we report.
		if capped, err := json.Marshal(items); err == nil {
			data = capped
		}
	}
	return v.respond(data, listCounts{count: len(items), total: total, truncated: truncated, filtered: v.filtered}, notes)
}

// FormatItem creates a response with item context.
func FormatItem(data json.RawMessage, itemType ItemType, id string) (json.RawMessage, error) {
	summary := fmt.Sprintf("%s %s retrieved successfully", itemType.capitalized(), id)

	return json.Marshal(FormattedResponse{
		Summary: summary,
		Data:    data,
	})
}

// FormatAction creates a response confirming a CRUD action.
func FormatAction(data json.RawMessage, action string, itemType ItemType, id string) (json.RawMessage, error) {
	return json.Marshal(FormattedResponse{
		Summary: itemType.actionSummary(action, id),
		Data:    data,
	})
}
