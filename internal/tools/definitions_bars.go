package tools

import (
	"maps"

	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// manageBarDescription is the manage_bar contract. The field semantics it
// states were probed against the live API on 24-09-2026.
const manageBarDescription = `Create, update, or delete a bar on a roadmap.

USE WHEN: "Add feature", "Update dates", "Delete item", "Change color", "Nest this bar under that one"
For many bars at once (e.g. "color these 30 bars"), use bulk_update_bars, bulk_create_bars, or bulk_delete_bars instead.
Actions: create (roadmap_id + name + lane or lane_id), update (bar_id + any fields), delete (bar_id)
Color a bar with legend (a legend NAME from get_roadmap_legends); legend:"" or clear_legend:true removes the color. Omitted fields are never sent, so an update only touches what you pass.
New bars are parked (off the timeline) by ProductPlan's default. When you give starts_on and ends_on on create and leave parked unset, the bar defaults to parked:false so it lands on the timeline; a nested bar inherits its container's parked state.
Names (legend, lane, custom field labels, dropdown values) are checked against the roadmap before writing, matched case-insensitively, and sent in canonical spelling.
Returns: create gives the new bar's id plus the bar read back; update gives the exact fields sent.
FAILS WHEN: create without roadmap_id, name, or a lane; update/delete without bar_id; a legend, lane, or custom field name is not on the roadmap (the error lists the valid ones); container_bar_id on a bar without both dates, or a parked state that differs from the container's. legend_id and effort are rejected with an explanation (legend_id would wipe the bar's color). A bar cannot be un-nested via the API once container_bar_id is set. WARNING: delete is permanent and cannot be undone.`

// barFieldProperties are the writable bar fields shared by manage_bar and
// the bulk bar tools.
func barFieldProperties() map[string]mcp.Property {
	return map[string]mcp.Property{
		"name":                   {Type: "string", Description: "Bar name"},
		"lane":                   {Type: "string", Description: "Lane NAME (see get_roadmap lanes). Use this or lane_id"},
		"lane_id":                {Type: "string", Description: "Lane ID from get_roadmap_lanes. Use this or lane"},
		"legend":                 {Type: "string", Description: "Legend NAME that colors the bar (from get_roadmap_legends). Empty string clears the color"},
		"clear_legend":           {Type: "boolean", Description: "True to remove the bar's color (same as legend:\"\")"},
		"starts_on":              {Type: "string", Description: "Start date YYYY-MM-DD", Pattern: `^\d{4}-\d{2}-\d{2}$`, Examples: []any{"2025-03-15"}},
		"ends_on":                {Type: "string", Description: "End date YYYY-MM-DD", Pattern: `^\d{4}-\d{2}-\d{2}$`, Examples: []any{"2025-06-30"}},
		"description":            {Type: "string", Description: "Description (markdown)"},
		"percent_done":           {Type: "integer", Description: "Progress 0-100", Minimum: floatPtr(0), Maximum: floatPtr(100)},
		"is_container":           {Type: "boolean", Description: "True to make the bar a container for child bars"},
		"container_bar_id":       {Type: "string", Description: "Numeric ID of the container bar to nest under. Requires starts_on and ends_on on the bar; cannot be removed once set"},
		"parked":                 {Type: "boolean", Description: "True parks the bar (off the timeline, kept on the roadmap). Must match the container's when nested"},
		"strategic_value":        {Type: "string", Description: "Free-text strategic importance note"},
		"notes":                  {Type: "string", Description: "Additional notes"},
		"tags":                   {Type: "array", Description: "Tag strings; REPLACES the bar's tag list. [] clears it", Items: &mcp.Property{Type: "string", Description: "Tag name"}},
		"custom_text_fields":     {Type: "array", Description: "[{label,value}] using labels from get_roadmap custom_text_fields", Items: &mcp.Property{Type: "object", Description: "{\"label\": \"<field label>\", \"value\": \"<text>\"}"}},
		"custom_dropdown_fields": {Type: "array", Description: "[{label,value}]; value must be one of the field's allowed_values from get_roadmap", Items: &mcp.Property{Type: "object", Description: "{\"label\": \"<field label>\", \"value\": \"<allowed value>\"}"}},
	}
}

// manageBarProperties adds manage_bar's action and ID arguments, plus the
// deprecated aliases that older agents may still send, to the shared fields.
func manageBarProperties() map[string]mcp.Property {
	props := map[string]mcp.Property{
		"action":     {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
		"bar_id":     {Type: "string", Description: "Bar ID (for update/delete)"},
		"roadmap_id": {Type: "string", Description: "Roadmap ID (required for create; optional on update to skip looking up the bar's roadmap)"},
		"legend_id":  {Type: "string", Description: "DEPRECATED and rejected: the API reads legend_id as 'clear the color'. Use legend (name)"},
		"parent_id":  {Type: "string", Description: "DEPRECATED alias for container_bar_id"},
		"container":  {Type: "boolean", Description: "DEPRECATED alias for is_container"},
		"effort":     {Type: "integer", Description: "DEPRECATED and rejected: not a ProductPlan bar field (the API ignores it). Use a custom field"},
	}
	maps.Copy(props, barFieldProperties())
	return props
}
