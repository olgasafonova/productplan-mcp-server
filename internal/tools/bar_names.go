package tools

import (
	"errors"
	"fmt"
	"strings"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

// resolveAgainstSchema checks every name in p against the roadmap and
// rewrites case-insensitive matches to the canonical spelling the API
// expects. It reports every problem at once, each with the valid options,
// so the caller can fix a whole payload in one retry.
func resolveAgainstSchema(p map[string]any, s *api.BarWriteSchema) error {
	var problems []string
	if name, ok := p["legend"].(string); ok {
		if canon, found := matchName(s.Legends, name); found {
			p["legend"] = canon
		} else {
			problems = append(problems, fmt.Sprintf("legend %q is not on roadmap %s. Valid legends: %s", name, s.RoadmapID, quoteList(s.Legends)))
		}
	}
	if name, ok := p["lane"].(string); ok {
		if canon, found := matchName(s.Lanes, name); found {
			p["lane"] = canon
		} else {
			problems = append(problems, fmt.Sprintf("lane %q is not on roadmap %s. Valid lanes: %s", name, s.RoadmapID, quoteList(s.Lanes)))
		}
	}
	if fields, ok := p["custom_text_fields"].([]map[string]any); ok {
		problems = append(problems, resolveTextFields(fields, s)...)
	}
	if fields, ok := p["custom_dropdown_fields"].([]map[string]any); ok {
		problems = append(problems, resolveDropdownFields(fields, s)...)
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func resolveTextFields(fields []map[string]any, s *api.BarWriteSchema) []string {
	var problems []string
	for _, f := range fields {
		label, _ := f["label"].(string)
		if canon, found := matchName(s.CustomTextFields, label); found {
			f["label"] = canon
			continue
		}
		problems = append(problems, fmt.Sprintf("custom text field %q is not defined on roadmap %s. Defined text fields: %s",
			label, s.RoadmapID, quoteList(s.CustomTextFields)))
	}
	return problems
}

func resolveDropdownFields(fields []map[string]any, s *api.BarWriteSchema) []string {
	labels := make([]string, len(s.CustomDropdownFields))
	for i, d := range s.CustomDropdownFields {
		labels[i] = d.Label
	}
	var problems []string
	for _, f := range fields {
		label, _ := f["label"].(string)
		idx := indexOfName(labels, label)
		if idx < 0 {
			problems = append(problems, fmt.Sprintf("custom dropdown field %q is not defined on roadmap %s. Defined dropdown fields: %s",
				label, s.RoadmapID, quoteList(labels)))
			continue
		}
		def := s.CustomDropdownFields[idx]
		f["label"] = def.Label
		value, _ := f["value"].(string)
		if canon, found := matchName(def.AllowedValues, value); found {
			f["value"] = canon
			continue
		}
		problems = append(problems, fmt.Sprintf("%q is not an allowed value for dropdown %q. Allowed values: %s",
			value, def.Label, quoteList(def.AllowedValues)))
	}
	return problems
}

// matchName finds name in names, exactly first and then case-insensitively,
// returning the canonical spelling.
func matchName(names []string, name string) (string, bool) {
	if i := indexOfName(names, name); i >= 0 {
		return names[i], true
	}
	return "", false
}

func indexOfName(names []string, name string) int {
	for i, n := range names {
		if n == name {
			return i
		}
	}
	trimmed := strings.TrimSpace(name)
	for i, n := range names {
		if strings.EqualFold(n, trimmed) {
			return i
		}
	}
	return -1
}
