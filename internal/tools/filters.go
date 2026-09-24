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

// filterProperties returns the schema properties for tool's filter and sort
// arguments, merged over base (base wins on a name clash, which the tests
// forbid). Tools without a spec get base unchanged.
func filterProperties(tool string, base map[string]mcp.Property) map[string]mcp.Property {
	spec, ok := listFilters[tool]
	if !ok {
		return base
	}
	props := make(map[string]mcp.Property, len(base)+len(spec.args)+1)
	for _, fa := range spec.args {
		p := mcp.Property{Type: "string", Description: fa.desc}
		switch fa.kind {
		case dateFilter:
			p.Pattern = datePattern
			p.Examples = []any{"2026-01-01"}
		case boolFilter:
			p.Type = "boolean"
		}
		props[fa.arg] = p
	}
	props["sort"] = mcp.Property{
		Type:        "string",
		Description: fmt.Sprintf("Sort as \"field\" or \"field asc|desc\" (server-side). Fields: %s", strings.Join(spec.sortFields, ", ")),
		Pattern:     sortPattern(spec.sortFields),
		Examples:    []any{spec.sortFields[1] + " asc"},
	}
	for k, v := range base {
		props[k] = v
	}
	return props
}

func sortPattern(fields []string) string {
	quoted := make([]string, len(fields))
	for i, f := range fields {
		quoted[i] = regexp.QuoteMeta(f)
	}
	return `^(` + strings.Join(quoted, "|") + `)( (asc|desc))?$`
}

// buildQuery validates tool's filter and sort arguments and maps them to an
// api.Query. Absent or empty arguments do not filter. A malformed date, a
// non-boolean is_container, or an unknown sort field is an error that names
// the argument and the accepted values.
func buildQuery(tool string, args map[string]any) (api.Query, error) {
	spec, ok := listFilters[tool]
	if !ok {
		return api.Query{}, nil
	}
	q := api.Query{Predicates: map[string]string{}}
	for _, fa := range spec.args {
		raw, present := args[fa.arg]
		if !present || raw == nil {
			continue
		}
		if err := addPredicate(q.Predicates, fa, raw); err != nil {
			return api.Query{}, err
		}
	}
	if raw, present := args["sort"]; present && raw != nil {
		s, err := parseSort(tool, spec.sortFields, raw)
		if err != nil {
			return api.Query{}, err
		}
		q.Sort = s
	}
	return q, nil
}

func addPredicate(preds map[string]string, fa filterArg, raw any) error {
	switch fa.kind {
	case boolFilter:
		b, err := asBool(fa.arg, raw)
		if err != nil {
			return err
		}
		if b {
			preds[fa.predicate+"_true"] = "1"
		} else {
			preds[fa.predicate+"_false"] = "1"
		}
		return nil
	case dateFilter:
		s, err := asString(fa.arg, raw)
		if err != nil || s == "" {
			return err
		}
		if _, perr := time.Parse("2006-01-02", s); perr != nil {
			return fmt.Errorf("%s must be a calendar date in YYYY-MM-DD format, got %q", fa.arg, s)
		}
		preds[fa.predicate] = s
		return nil
	default:
		s, err := asString(fa.arg, raw)
		if err != nil || s == "" {
			return err
		}
		preds[fa.predicate] = s
		return nil
	}
}

func parseSort(tool string, allowed []string, raw any) (string, error) {
	s, err := asString("sort", raw)
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
	if !slices.Contains(allowed, field) {
		sorted := slices.Clone(allowed)
		sort.Strings(sorted)
		return "", fmt.Errorf("sort field %q is not sortable for %s; allowed: %s", field, tool, strings.Join(sorted, ", "))
	}
	if dir != "asc" && dir != "desc" {
		return "", fmt.Errorf("sort direction must be asc or desc, got %q", parts[1])
	}
	return field + " " + dir, nil
}

func asString(arg string, raw any) (string, error) {
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string, got %T", arg, raw)
	}
	return strings.TrimSpace(s), nil
}

func asBool(arg string, raw any) (bool, error) {
	switch v := raw.(type) {
	case bool:
		return v, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
	}
	return false, fmt.Errorf("%s must be true or false, got %v", arg, raw)
}

// queryHandler is typedHandler plus the tool's validated filter query.
func queryHandler[T Validatable](tool string, fn func(ctx context.Context, a T, q api.Query) (json.RawMessage, error)) mcp.Handler {
	return mcp.HandlerFunc(func(ctx context.Context, args map[string]any) (json.RawMessage, error) {
		a, err := ParseArgs[T](args)
		if err != nil {
			return nil, err
		}
		if err = a.Validate(); err != nil {
			return nil, err
		}
		q, err := buildQuery(tool, args)
		if err != nil {
			return nil, err
		}
		return fn(ctx, a, q)
	})
}

// NoArgs is the arg struct for list tools whose only arguments are filters.
type NoArgs struct{}

// Validate always succeeds.
func (NoArgs) Validate() error { return nil }
