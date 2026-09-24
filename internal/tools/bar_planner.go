package tools

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

// memo fetches each key at most once per planner, even under concurrent
// callers, so a bulk call costs one roadmap GET per distinct roadmap.
type memo[T any] struct {
	mu      sync.Mutex
	entries map[string]*memoEntry[T]
}

type memoEntry[T any] struct {
	once sync.Once
	val  T
	err  error
}

func (m *memo[T]) get(key string, fetch func() (T, error)) (T, error) {
	m.mu.Lock()
	if m.entries == nil {
		m.entries = make(map[string]*memoEntry[T])
	}
	e, ok := m.entries[key]
	if !ok {
		e = &memoEntry[T]{}
		m.entries[key] = e
	}
	m.mu.Unlock()
	e.once.Do(func() { e.val, e.err = fetch() })
	return e.val, e.err
}

// barPlanner turns caller-facing bar fields into validated API payloads.
// It lives for one tool call; every read goes through client GETs, so a
// client-level read cache benefits it without changes here.
type barPlanner struct {
	client  *api.Client
	schemas memo[*api.BarWriteSchema]
	bars    memo[*api.BarSummary]
}

func newBarPlanner(client *api.Client) *barPlanner {
	return &barPlanner{client: client}
}

func (pl *barPlanner) schema(ctx context.Context, roadmapID string) (*api.BarWriteSchema, error) {
	return pl.schemas.get(roadmapID, func() (*api.BarWriteSchema, error) {
		return pl.client.GetBarWriteSchema(ctx, roadmapID)
	})
}

func (pl *barPlanner) bar(ctx context.Context, barID string) (*api.BarSummary, error) {
	return pl.bars.get(barID, func() (*api.BarSummary, error) {
		return pl.client.GetBarSummary(ctx, barID)
	})
}

// barOp is one intended bar write.
type barOp struct {
	create    bool
	barID     string // update only
	roadmapID string // required for create; optional hint for update
	fields    BarFields
}

// prepare validates op and returns the exact payload to send. It fetches
// the roadmap only when names need checking, and the bar or its container
// only when a nesting pre-check needs them.
func (pl *barPlanner) prepare(ctx context.Context, op barOp) (map[string]any, error) {
	p, err := pl.basePayload(ctx, op)
	if err != nil {
		return nil, err
	}
	err = firstError(
		func() error { return pl.resolveNames(ctx, op, p) },
		func() error { return pl.checkNesting(ctx, op, p) },
		func() error { return pl.completeCreate(ctx, op, p) },
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// basePayload translates op's fields, explaining a legend_id rejection with
// the roadmap's legends and refusing an update that changes nothing.
func (pl *barPlanner) basePayload(ctx context.Context, op barOp) (map[string]any, error) {
	p, err := op.fields.payload()
	if errors.Is(err, errLegendID) {
		return nil, pl.legendIDError(ctx, op)
	}
	if err != nil {
		return nil, err
	}
	if !op.create && len(p) == 0 {
		return nil, errors.New("nothing to update: pass at least one bar field")
	}
	return p, nil
}

// completeCreate adds the create-only roadmap_id and parked default.
func (pl *barPlanner) completeCreate(ctx context.Context, op barOp, p map[string]any) error {
	if !op.create {
		return nil
	}
	p["roadmap_id"] = numericID(op.roadmapID)
	return pl.defaultParked(ctx, op, p)
}

// roadmapFor returns the roadmap a write targets: the caller's roadmap_id,
// or for an update without one, the bar's own roadmap.
func (pl *barPlanner) roadmapFor(ctx context.Context, op barOp) (string, error) {
	if op.roadmapID != "" {
		return op.roadmapID, nil
	}
	bar, err := pl.bar(ctx, op.barID)
	if err != nil {
		return "", fmt.Errorf("could not look up bar %s to find its roadmap: %w", op.barID, err)
	}
	if bar.RoadmapID == "" {
		return "", fmt.Errorf("bar %s response carried no roadmap ID; pass roadmap_id explicitly", op.barID)
	}
	return bar.RoadmapID, nil
}

// resolveNames checks legend, lane, and custom field names against the
// roadmap, when the payload carries any.
func (pl *barPlanner) resolveNames(ctx context.Context, op barOp, p map[string]any) error {
	if !needsSchema(p) {
		return nil
	}
	roadmapID, err := pl.roadmapFor(ctx, op)
	if err != nil {
		return err
	}
	schema, err := pl.schema(ctx, roadmapID)
	if err != nil {
		return fmt.Errorf("could not load roadmap %s to validate names: %w", roadmapID, err)
	}
	return resolveAgainstSchema(p, schema)
}

// legendIDError explains the legend_id rejection with the roadmap's valid
// legend names, falling back to a pointer at get_roadmap_legends.
func (pl *barPlanner) legendIDError(ctx context.Context, op barOp) error {
	roadmapID, err := pl.roadmapFor(ctx, op)
	if err == nil {
		var schema *api.BarWriteSchema
		if schema, err = pl.schema(ctx, roadmapID); err == nil {
			return fmt.Errorf("%w. Valid legends on roadmap %s: %s", errLegendID, roadmapID, quoteList(schema.Legends))
		}
	}
	return fmt.Errorf("%w. Call get_roadmap_legends for the valid names (lookup failed: %v)", errLegendID, err)
}

// checkNesting pre-checks container_bar_id against the rules the API
// enforces with 422s: the child needs both dates, cannot nest in itself,
// and its parked state must equal the container's.
func (pl *barPlanner) checkNesting(ctx context.Context, op barOp, p map[string]any) error {
	parentID, ok := containerID(p)
	if !ok {
		return nil
	}
	if !op.create && parentID == op.barID {
		return fmt.Errorf("bar %s cannot be its own container", op.barID)
	}
	if err := pl.checkNestingDates(ctx, op, p); err != nil {
		return err
	}
	if op.create || p["parked"] != nil {
		return nil // create inherits the container's parked state in defaultParked
	}
	return pl.checkParkedMatches(ctx, op.barID, parentID)
}

// containerID returns the payload's container_bar_id as a string.
func containerID(p map[string]any) (string, bool) {
	raw, ok := p["container_bar_id"]
	if !ok {
		return "", false
	}
	return fmt.Sprint(raw), true
}

// checkNestingDates requires a nested bar to end up with both dates, from
// the payload or, for an update, from the bar as it stands.
func (pl *barPlanner) checkNestingDates(ctx context.Context, op barOp, p map[string]any) error {
	starts, ends := p["starts_on"] != nil, p["ends_on"] != nil
	if starts && ends {
		return nil
	}
	if op.create {
		return errors.New("container_bar_id requires starts_on and ends_on: ProductPlan rejects nesting an undated bar (422)")
	}
	bar, err := pl.bar(ctx, op.barID)
	if err != nil {
		return fmt.Errorf("could not look up bar %s to check its dates for nesting: %w", op.barID, err)
	}
	if lacksDate(starts, bar.StartsOn) || lacksDate(ends, bar.EndsOn) {
		return fmt.Errorf("container_bar_id requires bar %s to have starts_on and ends_on; it has none, so pass both dates in the same call", op.barID)
	}
	return nil
}

// lacksDate reports a date the payload does not set and the bar lacks.
func lacksDate(inPayload bool, existing string) bool {
	return !inPayload && existing == ""
}

// checkParkedMatches fails early when an update would nest a bar under a
// container whose parked state differs, which the API rejects with 422
// "Parked must match parent".
func (pl *barPlanner) checkParkedMatches(ctx context.Context, barID, parentID string) error {
	parent, err := pl.bar(ctx, parentID)
	if err != nil {
		return fmt.Errorf("could not look up container bar %s: %w", parentID, err)
	}
	child, err := pl.bar(ctx, barID)
	if err != nil {
		return fmt.Errorf("could not look up bar %s: %w", barID, err)
	}
	if boolsConflict(parent.Parked, child.Parked) {
		return fmt.Errorf("bar %s is parked=%t but container %s is parked=%t, and ProductPlan requires them to match; pass parked:%t in the same call",
			barID, *child.Parked, parentID, *parent.Parked, *parent.Parked)
	}
	return nil
}

// defaultParked applies create-time parked defaults when the caller did not
// set parked. ProductPlan parks every new bar by default; a dated bar
// defaults to parked:false so it lands on the timeline, and a nested bar
// inherits its container's parked state because the API requires a match.
func (pl *barPlanner) defaultParked(ctx context.Context, op barOp, p map[string]any) error {
	if op.fields.Parked != nil {
		return nil
	}
	if parentID, ok := containerID(p); ok {
		parent, err := pl.bar(ctx, parentID)
		if err != nil {
			return fmt.Errorf("could not look up container bar %s: %w", parentID, err)
		}
		if parent.Parked != nil {
			p["parked"] = *parent.Parked
		}
		return nil
	}
	if p["starts_on"] != nil && p["ends_on"] != nil {
		p["parked"] = false
	}
	return nil
}

// needsSchema reports whether a payload carries names that must be checked
// against the roadmap.
func needsSchema(p map[string]any) bool {
	if _, ok := p["legend"].(string); ok {
		return true
	}
	for _, k := range []string{"lane", "custom_text_fields", "custom_dropdown_fields"} {
		if _, ok := p[k]; ok {
			return true
		}
	}
	return false
}

// explainWriteError adds the known ProductPlan 422 causes for the fields a
// failed payload carried, since the API's messages rarely name the fix.
func explainWriteError(err error, p map[string]any) error {
	if err == nil || !strings.Contains(err.Error(), "API error 422") {
		return err
	}
	var hints []string
	if _, ok := p["container_bar_id"]; ok {
		hints = append(hints, "container_bar_id needs the bar to have starts_on and ends_on, and its parked must equal the container's")
	}
	if _, ok := p["parked"]; ok {
		hints = append(hints, "a nested bar's parked must match its container's (\"Parked must match parent\")")
	}
	if _, ok := p["custom_dropdown_fields"]; ok {
		hints = append(hints, "dropdown values must be one of the field's allowed_values (see get_roadmap)")
	}
	if len(hints) == 0 {
		return err
	}
	return fmt.Errorf("%w. Likely cause: %s", err, strings.Join(hints, "; "))
}

// maxListedNames bounds how many valid names an error message lists.
const maxListedNames = 40

// quoteList renders names as a quoted, comma-separated list, capped at
// maxListedNames with the remainder counted.
func quoteList(names []string) string {
	if len(names) == 0 {
		return "(none defined)"
	}
	shown := names
	if len(shown) > maxListedNames {
		shown = shown[:maxListedNames]
	}
	quoted := make([]string, len(shown))
	for i, n := range shown {
		quoted[i] = strconv.Quote(n)
	}
	out := strings.Join(quoted, ", ")
	if extra := len(names) - len(shown); extra > 0 {
		out += fmt.Sprintf(" (and %d more)", extra)
	}
	return out
}
