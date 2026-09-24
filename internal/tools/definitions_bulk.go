package tools

import (
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// bulkBarTools returns the bulk bar tool definitions.
//
// The bulk_ prefix is not one annotateTool understands, so each tool
// declares its annotations explicitly (CONSTITUTION Article I):
//
//   - bulk_update_bars: DestructiveHint=true. An update overwrites fields
//     and can clear a legend or replace a tag list, so it is not additive.
//   - bulk_create_bars: DestructiveHint=false. It only adds bars.
//   - bulk_delete_bars: DestructiveHint=true. Deletes are permanent.
//
// All three are OpenWorldHint=true (they write to an external service).
// IdempotentHint is left unset on all three, as on every tool in this
// server (Article VIII); bulk_update_bars is idempotent in practice, but
// claiming it is an amendment to that article, not a per-tool choice.
func bulkBarTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "bulk_update_bars",
			Description: `Update many bars in one call: recolor, move lanes, retag, reschedule, set progress.

USE WHEN: "Color all these bars Committed", "Move these 12 bars to the Backend lane", "Set percent_done on every Q3 bar"
For a single bar, use manage_bar instead.
Put values shared by every bar in set (e.g. set:{"legend":"Committed"}) and per-bar values in items; an item's own fields override set. Up to 100 items; each bar_id may appear once.
Every item is validated before anything is written: legend, lane, custom field labels, and dropdown values are checked against each bar's roadmap (one roadmap fetch per distinct roadmap; pass roadmap_id to skip looking up each bar's roadmap). One invalid item means nothing is sent, and the error lists every problem with the valid options.
dry_run:true returns the exact PATCH payload per bar without writing.
Writes run 4 at a time under the client's rate limiter. Returns per-item {index, bar_id, ok, error} and a summary like "Updated 28 of 30 bars; 2 failed".
The result is an error only when no bar was updated. A partial failure is a normal result whose summary counts the failures; retry only the failed items.
FAILS WHEN: items empty or over 100, a bar_id missing or repeated, a name not on the roadmap, legend_id or effort passed (see manage_bar).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"items": {Type: "array", Description: "Bars to update: [{bar_id, ...fields}]. Fields are the same as manage_bar update: name, lane, lane_id, legend, clear_legend, starts_on, ends_on, description, percent_done, is_container, container_bar_id, parked, strategic_value, notes, tags, custom_text_fields, custom_dropdown_fields. Optional per-item roadmap_id.",
						Items: &mcp.Property{Type: "object", Description: "{\"bar_id\": \"123\", \"legend\": \"Committed\"}"}},
					"set":        {Type: "object", Description: "Fields applied to every item unless the item sets them itself, e.g. {\"legend\":\"Committed\"}. Same fields as items"},
					"roadmap_id": {Type: "string", Description: "Roadmap all bars are on; skips one bar lookup per item when names need validating"},
					"dry_run":    {Type: "boolean", Description: "True to validate and return the payloads without writing"},
				},
				Required: []string{"items"},
			},
			Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true), OpenWorldHint: boolPtr(true)},
		},
		{
			Name: "bulk_create_bars",
			Description: `Create many bars on one roadmap in one call.

USE WHEN: "Add these 20 features to the roadmap", "Import this list as bars", "Create a bar per epic"
For a single bar, use manage_bar instead.
Each item needs name and a lane (lane name or lane_id), from the item or from set. Put shared values in set (e.g. set:{"lane":"Backend","legend":"Exploring"}). Up to 100 items.
All items are validated against the roadmap before anything is written; one invalid item means nothing is sent. dry_run:true returns the exact POST payloads without writing.
Parking follows manage_bar: a bar given starts_on and ends_on and no parked lands on the timeline (parked:false); an undated bar is parked by ProductPlan's default; a nested bar inherits its container's parked state.
Writes run 4 at a time under the client's rate limiter. Returns per-item {index, bar_id, name, ok, error} and a summary like "Created 18 of 20 bars; 2 failed".
The result is an error only when no bar was created. On a partial failure, retry only the failed items: resending the whole call would duplicate the bars that were created.
FAILS WHEN: roadmap_id missing, items empty or over 100, an item without name or lane, a name not on the roadmap.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"roadmap_id": {Type: "string", Description: "Roadmap to create the bars on"},
					"items": {Type: "array", Description: "Bars to create: [{name, lane or lane_id, ...fields}]. Fields are the same as manage_bar create.",
						Items: &mcp.Property{Type: "object", Description: "{\"name\": \"SSO\", \"lane\": \"Backend\", \"starts_on\": \"2026-01-05\", \"ends_on\": \"2026-02-27\"}"}},
					"set":     {Type: "object", Description: "Fields applied to every item unless the item sets them itself"},
					"dry_run": {Type: "boolean", Description: "True to validate and return the payloads without writing"},
				},
				Required: []string{"roadmap_id", "items"},
			},
			Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(false), OpenWorldHint: boolPtr(true)},
		},
		{
			Name: "bulk_delete_bars",
			Description: `Permanently delete many bars in one call, verifying each delete.

USE WHEN: "Delete these bars", "Remove all the test bars I just made", "Clean up the parked duplicates"
For a single bar, use manage_bar action=delete instead.
Requires confirm:true. dry_run:true looks each bar up and returns its name without deleting, so the list can be checked first. Up to 100 bar_ids, each once.
Each delete is verified by reading the bar back and expecting 404; a bar that still exists after a successful-looking delete is reported as failed.
Deletes run 4 at a time under the client's rate limiter. Returns per-item {index, bar_id, ok, error} and a summary like "Deleted 9 of 10 bars; 1 failed".
The result is an error only when no bar was deleted.
FAILS WHEN: confirm is not true (and dry_run is not set), bar_ids empty, over 100, malformed, or repeated. WARNING: delete is permanent and cannot be undone.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"bar_ids": {Type: "array", Description: "Bar IDs to delete", Items: &mcp.Property{Type: "string", Description: "Bar ID"}},
					"confirm": {Type: "boolean", Description: "Must be true to delete"},
					"dry_run": {Type: "boolean", Description: "True to look the bars up and list them without deleting"},
				},
				Required: []string{"bar_ids"},
			},
			Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true), OpenWorldHint: boolPtr(true)},
		},
	}
}
