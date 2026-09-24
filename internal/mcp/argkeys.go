package mcp

import (
	"fmt"
	"slices"
	"strings"
)

// maxSuggestDistance is the largest edit distance at which an unknown
// argument key is offered a "did you mean" suggestion.
const maxSuggestDistance = 2

// CheckArgumentKeys rejects argument keys the tool's InputSchema does not
// declare. The error names every unknown key, suggests the closest declared
// key when one is within maxSuggestDistance edits, and lists the valid keys,
// so an agent can correct the call in one retry.
func (t Tool) CheckArgumentKeys(args map[string]any) error {
	unknown := t.unknownKeys(args)
	if len(unknown) == 0 {
		return nil
	}
	valid := t.argumentNames()
	described := make([]string, len(unknown))
	for i, k := range unknown {
		described[i] = describeUnknown(k, valid)
	}
	noun := "argument"
	if len(unknown) > 1 {
		noun = "arguments"
	}
	return fmt.Errorf("unknown %s for %s: %s. %s", noun, t.Name, strings.Join(described, ", "), t.validArgumentsHint(valid))
}

// unknownKeys returns the keys of args the schema does not declare, sorted.
func (t Tool) unknownKeys(args map[string]any) []string {
	var unknown []string
	for k := range args {
		if _, ok := t.InputSchema.Properties[k]; !ok {
			unknown = append(unknown, k)
		}
	}
	slices.Sort(unknown)
	return unknown
}

// describeUnknown quotes an unknown key, with a suggestion when one is close.
func describeUnknown(key string, valid []string) string {
	if s, ok := closestName(key, valid); ok {
		return fmt.Sprintf("%q (did you mean %q?)", key, s)
	}
	return fmt.Sprintf("%q", key)
}

// argumentNames returns the declared argument keys, sorted.
func (t Tool) argumentNames() []string {
	names := make([]string, 0, len(t.InputSchema.Properties))
	for k := range t.InputSchema.Properties {
		names = append(names, k)
	}
	slices.Sort(names)
	return names
}

// validArgumentsHint lists the valid keys, or says the tool takes none.
func (t Tool) validArgumentsHint(valid []string) string {
	if len(valid) == 0 {
		return t.Name + " takes no arguments"
	}
	return "Valid arguments: " + strings.Join(valid, ", ")
}

// closestName returns the candidate nearest to key by edit distance, if it
// is within maxSuggestDistance. Ties go to the alphabetically first.
func closestName(key string, candidates []string) (string, bool) {
	best, bestDist := "", maxSuggestDistance+1
	for _, c := range candidates {
		if d := editDistance(key, c); d < bestDist {
			best, bestDist = c, d
		}
	}
	return best, bestDist <= maxSuggestDistance
}
