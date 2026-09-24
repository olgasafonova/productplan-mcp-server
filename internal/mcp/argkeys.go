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
	var unknown []string
	for k := range args {
		if _, ok := t.InputSchema.Properties[k]; !ok {
			unknown = append(unknown, k)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	slices.Sort(unknown)
	valid := t.argumentNames()

	described := make([]string, len(unknown))
	for i, k := range unknown {
		described[i] = fmt.Sprintf("%q", k)
		if s, ok := closestName(k, valid); ok {
			described[i] += fmt.Sprintf(" (did you mean %q?)", s)
		}
	}
	noun := "argument"
	if len(unknown) > 1 {
		noun = "arguments"
	}
	return fmt.Errorf("unknown %s for %s: %s. %s", noun, t.Name, strings.Join(described, ", "), validArgumentsHint(t.Name, valid))
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
func validArgumentsHint(tool string, valid []string) string {
	if len(valid) == 0 {
		return tool + " takes no arguments"
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

// editDistance is the Levenshtein distance between a and b, by byte (tool
// argument keys are ASCII snake_case).
func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+substitutionCost(a[i-1], b[j-1]))
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

func substitutionCost(x, y byte) int {
	if x == y {
		return 0
	}
	return 1
}
