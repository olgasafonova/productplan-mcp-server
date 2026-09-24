// Package tools provides response formatting utilities for MCP tool handlers.
package tools

import (
	"encoding/json"
	"fmt"
	"strings"
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
func FormatList(data json.RawMessage, itemType string) (json.RawMessage, error) {
	return FormatFilteredList(data, itemType, false)
}

// FormatFilteredList is FormatList for a list the caller narrowed with
// filters: an empty result then reads "No <items> matched the filters", so
// the agent can tell "nothing exists" from "my filter excluded everything".
func FormatFilteredList(data json.RawMessage, itemType string, filtered bool) (json.RawMessage, error) {
	var items []any
	if err := json.Unmarshal(data, &items); err == nil {
		return formatArray(data, items, itemType, filtered, nil)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return data, nil
	}
	if raw, ok := obj["results"]; ok {
		if err := json.Unmarshal(raw, &items); err == nil {
			return formatArray(raw, items, itemType, filtered, envelopeNotes(obj["paging"]))
		}
	}
	if _, hasCount := obj["count"]; hasCount {
		if _, hasTotal := obj["total"]; hasTotal {
			return formatProjected(data, itemType, filtered)
		}
	}
	return data, nil
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

// formatProjected wraps an already-projected payload, deriving the summary
// from its own fields. The payload itself is not re-capped: the api layer
// capped it at the same defaultListCap and reports truncated/total.
func formatProjected(data json.RawMessage, itemType string, filtered bool) (json.RawMessage, error) {
	var p projectedList
	if err := json.Unmarshal(data, &p); err != nil {
		return data, nil
	}
	var summary string
	switch {
	case p.Truncated:
		summary = fmt.Sprintf("Showing first %d of %d %s (refine to narrow)", p.Count, p.Total, pluralize(itemType, p.Total))
	case p.Count == 0 && (p.Filtered || filtered):
		summary = fmt.Sprintf("No %s matched the filters", pluralize(itemType, 0))
	case p.Count == 0:
		summary = fmt.Sprintf("No %s found", pluralize(itemType, 0))
	default:
		summary = fmt.Sprintf("Found %d %s", p.Count, pluralize(itemType, p.Count))
	}
	var notes []string
	if p.Incomplete && p.Note != "" {
		notes = append(notes, p.Note)
	}
	for _, w := range p.Warnings {
		notes = append(notes, "Warning: "+w)
	}
	return json.Marshal(FormattedResponse{Summary: appendNotes(summary, notes), Data: data})
}

// appendNotes joins notes onto a summary sentence.
func appendNotes(summary string, notes []string) string {
	for _, n := range notes {
		summary += ". " + n
	}
	return summary
}

// formatArray caps a decoded array and builds its summary.
func formatArray(data json.RawMessage, items []any, itemType string, filtered bool, notes []string) (json.RawMessage, error) {
	total := len(items)
	truncated := false
	if total > defaultListCap {
		items = items[:defaultListCap]
		truncated = true
		// Re-marshal the capped slice so Data carries only what we report.
		if capped, err := json.Marshal(items); err == nil {
			data = capped
		}
	}

	count := len(items)
	summary := fmt.Sprintf("Found %d %s", count, pluralize(itemType, count))
	switch {
	case truncated:
		summary = fmt.Sprintf("Showing first %d of %d %s (refine to narrow)", count, total, pluralize(itemType, total))
	case count == 0 && filtered:
		summary = fmt.Sprintf("No %s matched the filters", pluralize(itemType, 0))
	case count == 0:
		summary = fmt.Sprintf("No %s found", pluralize(itemType, 0))
	}

	return json.Marshal(FormattedResponse{
		Summary: appendNotes(summary, notes),
		Data:    data,
	})
}

// FormatItem creates a response with item context.
func FormatItem(data json.RawMessage, itemType, id string) (json.RawMessage, error) {
	summary := fmt.Sprintf("%s %s retrieved successfully", capitalize(itemType), id)

	return json.Marshal(FormattedResponse{
		Summary: summary,
		Data:    data,
	})
}

// FormatAction creates a response confirming a CRUD action.
func FormatAction(data json.RawMessage, action, itemType, id string) (json.RawMessage, error) {
	var summary string
	switch action {
	case "create":
		summary = fmt.Sprintf("%s created successfully", capitalize(itemType))
	case "update":
		summary = fmt.Sprintf("%s %s updated successfully", capitalize(itemType), id)
	case "delete":
		summary = fmt.Sprintf("%s %s deleted successfully", capitalize(itemType), id)
	default:
		summary = fmt.Sprintf("%s action completed for %s", capitalize(action), itemType)
	}

	return json.Marshal(FormattedResponse{
		Summary: summary,
		Data:    data,
	})
}

// pluralize returns the plural of word for count != 1. It covers the
// English rules the item types here need: "opportunity" -> "opportunities",
// "launch" -> "launches"; everything else takes "s".
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	switch {
	case strings.HasSuffix(word, "y") && len(word) > 1 && !strings.ContainsRune("aeiou", rune(word[len(word)-2])):
		return word[:len(word)-1] + "ies"
	case strings.HasSuffix(word, "ch"), strings.HasSuffix(word, "sh"),
		strings.HasSuffix(word, "s"), strings.HasSuffix(word, "x"):
		return word + "es"
	}
	return word + "s"
}

// capitalize makes the first letter uppercase.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}
