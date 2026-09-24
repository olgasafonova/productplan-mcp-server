package mcp

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
