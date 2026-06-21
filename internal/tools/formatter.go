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
// items at defaultListCap.
func FormatList(data json.RawMessage, itemType string) (json.RawMessage, error) {
	var items []any
	if err := json.Unmarshal(data, &items); err != nil {
		// Not an array, return as-is
		return data, nil
	}

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
	case count == 0:
		summary = fmt.Sprintf("No %s found", pluralize(itemType, 0))
	}

	return json.Marshal(FormattedResponse{
		Summary: summary,
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

// pluralize adds 's' for count != 1.
func pluralize(word string, count int) string {
	if count == 1 {
		return word
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
