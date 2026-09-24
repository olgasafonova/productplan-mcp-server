package tools

import (
	"errors"
	"fmt"

	"github.com/olgasafonova/productplan-mcp-server/pkg/productplan"
)

// maxBulkItems caps one bulk call. 100 writes at bulkConcurrency 4 stays
// well inside the API's rate window and a single response stays readable.
const maxBulkItems = 100

// BulkUpdateItem is one bar in a bulk_update_bars call.
type BulkUpdateItem struct {
	BarID     string `json:"bar_id"`
	RoadmapID string `json:"roadmap_id,omitempty"`
	BarFields
}

// BulkUpdateBarsArgs holds arguments for bulk_update_bars.
type BulkUpdateBarsArgs struct {
	Items     []BulkUpdateItem `json:"items"`
	Set       *BarFields       `json:"set,omitempty"`
	RoadmapID string           `json:"roadmap_id,omitempty"`
	DryRun    bool             `json:"dry_run,omitempty"`
}

// Validate checks item count and bar IDs.
func (a BulkUpdateBarsArgs) Validate() error {
	ids := make([]string, len(a.Items))
	for i, it := range a.Items {
		ids[i] = it.BarID
	}
	return validateBulkIDs("items", "bar_id", ids)
}

// BulkCreateBarsArgs holds arguments for bulk_create_bars.
type BulkCreateBarsArgs struct {
	RoadmapID string      `json:"roadmap_id"`
	Items     []BarFields `json:"items"`
	Set       *BarFields  `json:"set,omitempty"`
	DryRun    bool        `json:"dry_run,omitempty"`
}

// Validate checks the roadmap, item count, and each item's name and lane
// after `set` is applied.
func (a BulkCreateBarsArgs) Validate() error {
	if err := (fieldCheck{a.RoadmapID, "roadmap_id"}).require(); err != nil {
		return err
	}
	if err := checkBulkCount("items", len(a.Items)); err != nil {
		return err
	}
	for i, it := range a.Items {
		merged := mergeBarFields(a.Set, it)
		if merged.Name == "" {
			return fmt.Errorf("items[%d]: name is required", i)
		}
		if err := merged.requireLane(); err != nil {
			return fmt.Errorf("items[%d]: %w", i, err)
		}
	}
	return nil
}

// BulkDeleteBarsArgs holds arguments for bulk_delete_bars.
type BulkDeleteBarsArgs struct {
	BarIDs  []string `json:"bar_ids"`
	Confirm bool     `json:"confirm,omitempty"`
	DryRun  bool     `json:"dry_run,omitempty"`
}

// Validate checks IDs and requires confirm:true for a real delete.
func (a BulkDeleteBarsArgs) Validate() error {
	if err := validateBulkIDs("bar_ids", "bar_id", a.BarIDs); err != nil {
		return err
	}
	if !a.Confirm && !a.DryRun {
		return errors.New("bulk_delete_bars is permanent: pass confirm:true to delete, or dry_run:true to preview which bars would be deleted")
	}
	return nil
}

func checkBulkCount(field string, n int) error {
	switch {
	case n == 0:
		return fmt.Errorf("required parameter missing: %s (at least one entry)", field)
	case n > maxBulkItems:
		return fmt.Errorf("%s has %d entries; the limit is %d per call, so split the work into batches", field, n, maxBulkItems)
	}
	return nil
}

// validateBulkIDs requires 1..maxBulkItems well-formed, distinct IDs.
// Duplicates are refused because concurrent writes to one bar race.
func validateBulkIDs(field, idName string, ids []string) error {
	if err := checkBulkCount(field, len(ids)); err != nil {
		return err
	}
	seen := make(map[string]int, len(ids))
	for i, id := range ids {
		if err := productplan.RequireID(fmt.Sprintf("%s[%d].%s", field, i, idName), id); err != nil {
			return err
		}
		if j, dup := seen[id]; dup {
			return fmt.Errorf("%s[%d] repeats %s %s from %s[%d]; each bar may appear once per call", field, i, idName, id, field, j)
		}
		seen[id] = i
	}
	return nil
}

// mergeBarFields overlays an item's fields on the shared `set`: any field
// the item sets wins, anything it omits falls back to `set`.
func mergeBarFields(set *BarFields, item BarFields) BarFields {
	if set == nil {
		return item
	}
	m := item
	m.Name = firstString(item.Name, set.Name)
	m.Description = firstString(item.Description, set.Description)
	m.StartsOn = firstString(item.StartsOn, set.StartsOn)
	m.EndsOn = firstString(item.EndsOn, set.EndsOn)
	m.StrategicValue = firstString(item.StrategicValue, set.StrategicValue)
	m.Notes = firstString(item.Notes, set.Notes)
	m.ContainerBarID = firstString(item.ContainerBarID, set.ContainerBarID)
	m.LegendID = firstString(item.LegendID, set.LegendID)
	m.ParentID = firstString(item.ParentID, set.ParentID)
	m.PercentDone = firstPtr(item.PercentDone, set.PercentDone)
	m.Effort = firstPtr(item.Effort, set.Effort)
	m.IsContainer = firstBool(item.IsContainer, set.IsContainer)
	m.Container = firstBool(item.Container, set.Container)
	m.Parked = firstBool(item.Parked, set.Parked)
	if item.Tags == nil {
		m.Tags = set.Tags
	}
	if item.CustomTextFields == nil {
		m.CustomTextFields = set.CustomTextFields
	}
	if item.CustomDropdownFields == nil {
		m.CustomDropdownFields = set.CustomDropdownFields
	}
	// Lane and legend are each one choice: an item that picks either form
	// replaces the shared choice entirely.
	if item.Lane == "" && item.LaneID == "" {
		m.Lane, m.LaneID = set.Lane, set.LaneID
	}
	if item.Legend == nil && !item.ClearLegend {
		m.Legend, m.ClearLegend = set.Legend, set.ClearLegend
	}
	return m
}

func firstPtr[T any](values ...*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}
