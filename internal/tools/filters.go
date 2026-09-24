package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// This file is the single place where friendly list-tool filter arguments
// are mapped to ProductPlan's Ransack predicates (q[attribute_predicate]).
// The tool schemas (filterProperties), the validation (buildQuery) and the
// allowed sort fields all read the same table, so they cannot drift.
//
// Attribute lists come from each endpoint's "Filterable attributes" in the
// API reference (https://productplan.readme.io/llms.txt, fetched
// 24-09-2026). Only attributes that list documents are used. GET
// /strategy/objectives documents label_id, label_name, ... which do not
// correspond to objective fields, so list_objectives gets no filters until
// that is verified against the live API.

type filterKind int

const (
	textFilter filterKind = iota // free text, sent verbatim
	dateFilter                   // YYYY-MM-DD, validated
	boolFilter                   // true/false -> attribute_true=1 / attribute_false=1
)

// filterArg maps one tool argument to one predicate. For boolFilter,
// predicate is the bare attribute; the _true/_false suffix is chosen by value.
type filterArg struct {
	arg       string
	predicate string
	kind      filterKind
	desc      string
}

type listFilterSpec struct {
	tool       string // set by specFor, for error messages
	args       []filterArg
	sortFields []string
}

// listFilters is the arg -> predicate table, per tool.
var listFilters = map[string]listFilterSpec{
	"get_roadmap_bars": {
		args: []filterArg{
			{"name_contains", "name_i_cont", textFilter, "Only bars whose name contains this text (case-insensitive, server-side)"},
			{"starts_after", "starts_on_gteq", dateFilter, "Only bars starting on or after this date, YYYY-MM-DD (server-side)"},
			{"starts_before", "starts_on_lteq", dateFilter, "Only bars starting on or before this date, YYYY-MM-DD (server-side)"},
			{"ends_after", "ends_on_gteq", dateFilter, "Only bars ending on or after this date, YYYY-MM-DD (server-side)"},
			{"ends_before", "ends_on_lteq", dateFilter, "Only bars ending on or before this date, YYYY-MM-DD (server-side)"},
			{"is_container", "is_container", boolFilter, "true for container bars only, false to exclude containers (server-side)"},
		},
		sortFields: []string{"id", "name", "starts_on", "ends_on", "is_container", "created_at", "updated_at"},
	},
	"list_roadmaps": {
		args: []filterArg{
			{"name_contains", "name_i_cont", textFilter, "Only roadmaps whose name contains this text (case-insensitive, server-side)"},
		},
		sortFields: []string{"id", "name", "is_version", "created_at", "updated_at"},
	},
	"list_ideas": {
		args: []filterArg{
			{"name_contains", "name_i_cont", textFilter, "Only ideas whose name contains this text (case-insensitive, server-side)"},
			{"channel", "channel_eq", textFilter, "Only ideas from this channel, exact match (server-side)"},
		},
		sortFields: []string{"id", "name", "description", "channel", "customer", "opportunities_count", "source_name", "source_email", "location_status"},
	},
	"list_opportunities": {
		args: []filterArg{
			{"problem_contains", "problem_statement_i_cont", textFilter, "Only opportunities whose problem statement contains this text (case-insensitive, server-side)"},
			{"workflow_status", "workflow_status_eq", textFilter, "Only opportunities in this workflow status, exact match (server-side)"},
		},
		sortFields: []string{"id", "problem_statement", "workflow_status", "location_status", "description", "user_id", "ideas_count", "bars_count"},
	},
	"list_launches": {
		args: []filterArg{
			{"name_contains", "name_i_cont", textFilter, "Only launches whose name contains this text (case-insensitive, server-side)"},
			{"status", "status_eq", textFilter, "Only launches with this status, exact match (server-side)"},
			{"launch_after", "launch_date_gteq", dateFilter, "Only launches on or after this date, YYYY-MM-DD (server-side)"},
			{"launch_before", "launch_date_lteq", dateFilter, "Only launches on or before this date, YYYY-MM-DD (server-side)"},
		},
		sortFields: []string{"id", "name", "description", "launch_date", "status", "user_id"},
	},
}

const datePattern = `^\d{4}-\d{2}-\d{2}$`

// specFor returns tool's filter spec, reporting false when the tool takes
// no filters.
func specFor(tool string) (listFilterSpec, bool) {
	spec, ok := listFilters[tool]
	spec.tool = tool
	return spec, ok
}

// filterProperties returns the schema properties for tool's filter and sort
// arguments, merged over base (base wins on a name clash, which the tests
// forbid). Tools without a spec get base unchanged.
func filterProperties(tool string, base map[string]mcp.Property) map[string]mcp.Property {
	spec, ok := specFor(tool)
	if !ok {
		return base
	}
	props := make(map[string]mcp.Property, len(base)+len(spec.args)+1)
	for _, fa := range spec.args {
		props[fa.arg] = fa.property()
	}
	props["sort"] = spec.sortProperty()
	for k, v := range base {
		props[k] = v
	}
	return props
}

// property is the schema property advertised for the filter argument.
func (fa filterArg) property() mcp.Property {
	p := mcp.Property{Type: "string", Description: fa.desc}
	switch fa.kind {
	case dateFilter:
		p.Pattern = datePattern
		p.Examples = []any{"2026-01-01"}
	case boolFilter:
		p.Type = "boolean"
	}
	return p
}

// sortProperty is the schema property advertised for the sort argument.
func (spec listFilterSpec) sortProperty() mcp.Property {
	return mcp.Property{
		Type:        "string",
		Description: fmt.Sprintf("Sort as \"field\" or \"field asc|desc\" (server-side). Fields: %s", strings.Join(spec.sortFields, ", ")),
		Pattern:     spec.sortPattern(),
		Examples:    []any{spec.sortFields[1] + " asc"},
	}
}

func (spec listFilterSpec) sortPattern() string {
	quoted := make([]string, len(spec.sortFields))
	for i, f := range spec.sortFields {
		quoted[i] = regexp.QuoteMeta(f)
	}
	return `^(` + strings.Join(quoted, "|") + `)( (asc|desc))?$`
}

// buildQuery validates tool's filter and sort arguments and maps them to an
// api.Query. Absent or empty arguments do not filter. A malformed date, a
// non-boolean is_container, or an unknown sort field is an error that names
// the argument and the accepted values.
func buildQuery(tool string, args map[string]any) (api.Query, error) {
	spec, ok := specFor(tool)
	if !ok {
		return api.Query{}, nil
	}
	preds, err := spec.predicates(args)
	if err != nil {
		return api.Query{}, err
	}
	sortBy, err := spec.sortArg(args)
	if err != nil {
		return api.Query{}, err
	}
	return api.Query{Predicates: preds, Sort: sortBy}, nil
}

// predicates maps the present filter arguments to Ransack predicates.
func (spec listFilterSpec) predicates(args map[string]any) (map[string]string, error) {
	preds := map[string]string{}
	for _, fa := range spec.args {
		raw := args[fa.arg]
		if raw == nil {
			continue
		}
		if err := fa.addPredicate(preds, raw); err != nil {
			return nil, err
		}
	}
	return preds, nil
}

// addPredicate validates raw for the filter's kind and records its predicate.
func (fa filterArg) addPredicate(preds map[string]string, raw any) error {
	v := argValue{name: fa.arg, raw: raw}
	if fa.kind == boolFilter {
		b, err := v.boolean()
		if err != nil {
			return err
		}
		preds[fa.predicate+boolSuffix(b)] = "1"
		return nil
	}
	s, err := v.str()
	if err != nil || s == "" {
		return err
	}
	if err := fa.checkDate(s); err != nil {
		return err
	}
	preds[fa.predicate] = s
	return nil
}

// checkDate requires a dateFilter value to be a real YYYY-MM-DD date.
func (fa filterArg) checkDate(s string) error {
	if fa.kind != dateFilter {
		return nil
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return fmt.Errorf("%s must be a calendar date in YYYY-MM-DD format, got %q", fa.arg, s)
	}
	return nil
}

// boolSuffix is the Ransack predicate suffix for a boolean filter value.
func boolSuffix(b bool) string {
	if b {
		return "_true"
	}
	return "_false"
}

// sortArg validates the optional sort argument; absent or empty means "".
func (spec listFilterSpec) sortArg(args map[string]any) (string, error) {
	raw := args["sort"]
	if raw == nil {
		return "", nil
	}
	return spec.parseSort(raw)
}

func (spec listFilterSpec) parseSort(raw any) (string, error) {
	s, err := argValue{name: "sort", raw: raw}.str()
	if err != nil || s == "" {
		return "", err
	}
	parts := strings.Fields(s)
	field, dir := "", "asc"
	switch len(parts) {
	case 1:
		field = parts[0]
	case 2:
		field, dir = parts[0], strings.ToLower(parts[1])
	default:
		return "", fmt.Errorf("sort must be \"field\" or \"field asc|desc\", got %q", s)
	}
	if !slices.Contains(spec.sortFields, field) {
		sorted := slices.Clone(spec.sortFields)
		sort.Strings(sorted)
		return "", fmt.Errorf("sort field %q is not sortable for %s; allowed: %s", field, spec.tool, strings.Join(sorted, ", "))
	}
	if dir != "asc" && dir != "desc" {
		return "", fmt.Errorf("sort direction must be asc or desc, got %q", parts[1])
	}
	return field + " " + dir, nil
}

// argValue is a raw tool argument together with its name, for error
// messages that say which argument was wrong.
type argValue struct {
	name string
	raw  any
}

func (v argValue) str() (string, error) {
	s, ok := v.raw.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string, got %T", v.name, v.raw)
	}
	return strings.TrimSpace(s), nil
}

func (v argValue) boolean() (bool, error) {
	switch b := v.raw.(type) {
	case bool:
		return b, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(b)) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
	}
	return false, fmt.Errorf("%s must be true or false, got %v", v.name, v.raw)
}

// queryHandler is typedHandler plus the tool's validated filter query.
func queryHandler[T Validatable](tool string, fn func(ctx context.Context, a T, q api.Query) (json.RawMessage, error)) mcp.Handler {
	return typed[T]{fn: func(ctx context.Context, a T, args map[string]any) (json.RawMessage, error) {
		q, err := buildQuery(tool, args)
		if err != nil {
			return nil, err
		}
		return fn(ctx, a, q)
	}}
}

// NoArgs is the arg struct for list tools whose only arguments are filters.
type NoArgs struct{}

// Validate always succeeds.
func (NoArgs) Validate() error { return nil }
