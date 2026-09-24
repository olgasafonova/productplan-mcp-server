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
		if err := productplan.Field(fmt.Sprintf("%s[%d].%s", field, i, idName)).RequireID(id); err != nil {
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
	m.inheritScalars(set)
	m.inheritLists(set)
	m.inheritChoices(set)
	return m
}

// inheritScalars fills each unset string and pointer field from set.
func (f *BarFields) inheritScalars(set *BarFields) {
	f.Name = firstString(f.Name, set.Name)
	f.Description = firstString(f.Description, set.Description)
	f.StartsOn = firstString(f.StartsOn, set.StartsOn)
	f.EndsOn = firstString(f.EndsOn, set.EndsOn)
	f.StrategicValue = firstString(f.StrategicValue, set.StrategicValue)
	f.Notes = firstString(f.Notes, set.Notes)
	f.ContainerBarID = firstString(f.ContainerBarID, set.ContainerBarID)
	f.LegendID = firstString(f.LegendID, set.LegendID)
	f.ParentID = firstString(f.ParentID, set.ParentID)
	f.PercentDone = firstPtr(f.PercentDone, set.PercentDone)
	f.Effort = firstPtr(f.Effort, set.Effort)
	f.IsContainer = firstBool(f.IsContainer, set.IsContainer)
	f.Container = firstBool(f.Container, set.Container)
	f.Parked = firstBool(f.Parked, set.Parked)
}

// inheritLists takes set's list when the item's is nil; an item's explicit
// [] still wins, so it can clear a shared list.
func (f *BarFields) inheritLists(set *BarFields) {
	f.Tags = firstNonNil(f.Tags, set.Tags)
	f.CustomTextFields = firstNonNil(f.CustomTextFields, set.CustomTextFields)
	f.CustomDropdownFields = firstNonNil(f.CustomDropdownFields, set.CustomDropdownFields)
}

// inheritChoices handles lane and legend, which are each one choice: an
// item that picks either form replaces the shared choice entirely.
func (f *BarFields) inheritChoices(set *BarFields) {
	if f.Lane == "" && f.LaneID == "" {
		f.Lane, f.LaneID = set.Lane, set.LaneID
	}
	if f.Legend == nil && !f.ClearLegend {
		f.Legend, f.ClearLegend = set.Legend, set.ClearLegend
	}
}

// firstNonNil returns item unless it is nil, then fallback.
func firstNonNil[T any](item, fallback []T) []T {
	if item == nil {
		return fallback
	}
	return item
}

func firstPtr[T any](values ...*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}
