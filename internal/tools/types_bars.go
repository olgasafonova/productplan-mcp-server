package tools

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CustomFieldValue is one custom field assignment on a bar. The API
// requires {label, value}; `name` is accepted as a deprecated alias for
// label because the schema advertised {name,value} before 24-09-2026, and
// the API rejects that shape with 422.
type CustomFieldValue struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value"`
}

// BarFields holds the writable bar fields shared by manage_bar, bulk item
// entries, and the bulk `set` object. Pointer, slice-nil, and empty-string
// zero values all mean "omitted", and omitted fields are never sent.
type BarFields struct {
	Name                 string             `json:"name,omitempty"`
	Description          string             `json:"description,omitempty"`
	StartsOn             string             `json:"starts_on,omitempty"`
	EndsOn               string             `json:"ends_on,omitempty"`
	StrategicValue       string             `json:"strategic_value,omitempty"`
	Notes                string             `json:"notes,omitempty"`
	PercentDone          *int               `json:"percent_done,omitempty"`
	Tags                 []string           `json:"tags,omitempty"`
	Lane                 string             `json:"lane,omitempty"`
	LaneID               string             `json:"lane_id,omitempty"`
	Legend               *string            `json:"legend,omitempty"`
	ClearLegend          bool               `json:"clear_legend,omitempty"`
	IsContainer          *bool              `json:"is_container,omitempty"`
	Parked               *bool              `json:"parked,omitempty"`
	ContainerBarID       string             `json:"container_bar_id,omitempty"`
	CustomTextFields     []CustomFieldValue `json:"custom_text_fields,omitempty"`
	CustomDropdownFields []CustomFieldValue `json:"custom_dropdown_fields,omitempty"`

	// Deprecated inputs, still accepted so older agents get a precise
	// answer instead of a silent no-op. legend_id and effort are rejected;
	// parent_id and container are mapped to the documented fields.
	LegendID  string `json:"legend_id,omitempty"`
	ParentID  string `json:"parent_id,omitempty"`
	Container *bool  `json:"container,omitempty"`
	Effort    *int   `json:"effort,omitempty"`
}

// ManageBarArgs holds arguments for bar management operations.
type ManageBarArgs struct {
	Action    string `json:"action"`
	BarID     string `json:"bar_id,omitempty"`
	RoadmapID string `json:"roadmap_id,omitempty"`
	BarFields
}

// Validate checks required fields based on action.
func (a ManageBarArgs) Validate() error {
	if err := (fieldCheck{a.Action, "action"}).require(); err != nil {
		return err
	}
	switch a.Action {
	case "create":
		if err := requireAllForAction("create",
			fieldCheck{a.RoadmapID, "roadmap_id"},
			fieldCheck{a.Name, "name"},
		); err != nil {
			return err
		}
		return a.requireLane()
	case "update", "delete":
		return fieldCheck{a.BarID, "bar_id"}.requireFor(a.Action)
	}
	return nil
}

// requireLane enforces that a create names its lane one way or the other.
func (f BarFields) requireLane() error {
	if f.Lane == "" && f.LaneID == "" {
		return errors.New("required parameter missing: lane (lane name from get_roadmap, or lane_id from get_roadmap_lanes) is required for create")
	}
	return nil
}

// errLegendID marks a legend_id argument. It is enriched with the roadmap's
// valid legend names by the handler, which has the client to fetch them.
var errLegendID = errors.New("legend_id is not supported: ProductPlan treats legend_id as a request to CLEAR the bar's color, so it is never forwarded. Pass `legend` with the legend name instead")

var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// payload translates the fields into the documented BarCreate/BarUpdate
// contract, rejecting inputs the API would ignore, misread, or 422 on.
// Names (legend, lane, custom field labels) are checked against the
// roadmap separately by resolveAgainstSchema.
func (f BarFields) payload() (map[string]any, error) {
	if err := f.checkDeprecated(); err != nil {
		return nil, err
	}
	if err := f.checkValues(); err != nil {
		return nil, err
	}
	p := make(map[string]any)
	setIfNotEmpty(p, "name", f.Name)
	setIfNotEmpty(p, "description", f.Description)
	setIfNotEmpty(p, "starts_on", f.StartsOn)
	setIfNotEmpty(p, "ends_on", f.EndsOn)
	setIfNotEmpty(p, "strategic_value", f.StrategicValue)
	setIfNotEmpty(p, "notes", f.Notes)
	setIfNotEmpty(p, "lane", f.Lane)
	setIfNotNil(p, "percent_done", f.PercentDone)
	setIfNotNil(p, "parked", f.Parked)
	if f.Tags != nil {
		p["tags"] = f.Tags // an explicit [] clears the tag list
	}
	if f.LaneID != "" {
		p["lane_id"] = numericID(f.LaneID)
	}
	f.putLegend(p)
	if isContainer := firstBool(f.IsContainer, f.Container); isContainer != nil {
		p["is_container"] = *isContainer
	}
	if parent := firstString(f.ContainerBarID, f.ParentID); parent != "" {
		p["container_bar_id"] = numericID(parent)
	}
	if f.CustomTextFields != nil {
		p["custom_text_fields"] = customFieldPayload(f.CustomTextFields)
	}
	if f.CustomDropdownFields != nil {
		p["custom_dropdown_fields"] = customFieldPayload(f.CustomDropdownFields)
	}
	return p, nil
}

// putLegend sets legend to the name, or to JSON null when the caller asked
// to clear it (legend:"" or clear_legend:true). Omitted means not sent.
func (f BarFields) putLegend(p map[string]any) {
	switch {
	case f.ClearLegend, f.Legend != nil && *f.Legend == "":
		p["legend"] = nil
	case f.Legend != nil:
		p["legend"] = *f.Legend
	}
}

// checkDeprecated rejects deprecated inputs that cannot be mapped and
// catches contradictory aliases.
func (f BarFields) checkDeprecated() error {
	if f.LegendID != "" {
		return errLegendID
	}
	if f.Effort != nil {
		return errors.New("effort is not a ProductPlan bar field and the API silently ignores it, so it is rejected rather than dropped. Record effort in a custom field (custom_text_fields or custom_dropdown_fields; get_roadmap lists the roadmap's field definitions)")
	}
	if f.ParentID != "" && f.ContainerBarID != "" && f.ParentID != f.ContainerBarID {
		return errors.New("parent_id and container_bar_id disagree; pass container_bar_id only (parent_id is its deprecated alias)")
	}
	if f.Container != nil && f.IsContainer != nil && *f.Container != *f.IsContainer {
		return errors.New("container and is_container disagree; pass is_container only (container is its deprecated alias)")
	}
	return nil
}

// checkValues validates values that need no roadmap context.
func (f BarFields) checkValues() error {
	if f.Lane != "" && f.LaneID != "" {
		return errors.New("pass either lane (name) or lane_id, not both")
	}
	if f.ClearLegend && f.Legend != nil && *f.Legend != "" {
		return fmt.Errorf("clear_legend:true contradicts legend:%q; pass one or the other", *f.Legend)
	}
	for _, d := range []fieldCheck{{f.StartsOn, "starts_on"}, {f.EndsOn, "ends_on"}} {
		if d.value != "" && !datePattern.MatchString(d.value) {
			return fmt.Errorf("%s must be YYYY-MM-DD, got %q", d.name, d.value)
		}
	}
	if f.StartsOn != "" && f.EndsOn != "" && f.StartsOn > f.EndsOn {
		return fmt.Errorf("starts_on %s is after ends_on %s", f.StartsOn, f.EndsOn)
	}
	if f.PercentDone != nil && (*f.PercentDone < 0 || *f.PercentDone > 100) {
		return fmt.Errorf("percent_done must be 0-100, got %d", *f.PercentDone)
	}
	for _, id := range []string{f.ContainerBarID, f.ParentID} {
		if id != "" {
			if _, err := strconv.ParseInt(id, 10, 64); err != nil {
				return fmt.Errorf("container_bar_id must be a numeric bar ID, got %q", id)
			}
		}
	}
	return checkCustomLabels(f.CustomTextFields, f.CustomDropdownFields)
}

// checkCustomLabels requires every custom field entry to carry a label.
func checkCustomLabels(groups ...[]CustomFieldValue) error {
	for _, group := range groups {
		for _, cf := range group {
			if cf.Label == "" && cf.Name == "" {
				return errors.New("every custom field entry needs a label: {\"label\":\"<field label>\",\"value\":\"<value>\"}")
			}
		}
	}
	return nil
}

// customFieldPayload renders custom field entries as the API's
// [{label,value}] shape, mapping the deprecated `name` onto label.
func customFieldPayload(fields []CustomFieldValue) []map[string]any {
	out := make([]map[string]any, 0, len(fields))
	for _, cf := range fields {
		out = append(out, map[string]any{"label": firstString(cf.Label, cf.Name), "value": cf.Value})
	}
	return out
}

// numericID sends a numeric ID as a JSON integer (the documented type) and
// leaves anything else as the string the caller gave.
func numericID(id string) any {
	if n, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64); err == nil {
		return n
	}
	return id
}

// firstString returns the first non-empty string.
func firstString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// firstBool returns the first non-nil bool pointer.
func firstBool(values ...*bool) *bool {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}
