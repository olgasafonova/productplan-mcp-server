package tools

import (
	"fmt"
	"strings"
)

// ItemType is the human-readable name of a ProductPlan item type ("bar",
// "key result") as used in response summaries.
type ItemType string

// actionSummary renders the confirmation sentence for a CRUD action.
func (n ItemType) actionSummary(action, id string) string {
	switch action {
	case "create":
		return fmt.Sprintf("%s created successfully", n.capitalized())
	case "update":
		return fmt.Sprintf("%s %s updated successfully", n.capitalized(), id)
	case "delete":
		return fmt.Sprintf("%s %s deleted successfully", n.capitalized(), id)
	}
	return fmt.Sprintf("%s action completed for %s", ItemType(action).capitalized(), n)
}

// plural returns the plural of the noun for count != 1. It covers the
// English rules the item types here need: "opportunity" -> "opportunities",
// "launch" -> "launches"; everything else takes "s".
func (n ItemType) plural(count int) string {
	word := string(n)
	if count == 1 {
		return word
	}
	switch {
	case n.endsInConsonantY():
		return word[:len(word)-1] + "ies"
	case n.takesEsPlural():
		return word + "es"
	}
	return word + "s"
}

// endsInConsonantY reports nouns like "opportunity" whose plural is -ies.
func (n ItemType) endsInConsonantY() bool {
	word := string(n)
	if len(word) < 2 || !strings.HasSuffix(word, "y") {
		return false
	}
	return !strings.ContainsRune("aeiou", rune(word[len(word)-2]))
}

// takesEsPlural reports nouns ending in a sibilant ("launch", "box").
func (n ItemType) takesEsPlural() bool {
	for _, suffix := range []string{"ch", "sh", "s", "x"} {
		if strings.HasSuffix(string(n), suffix) {
			return true
		}
	}
	return false
}

// capitalized returns the noun with its first letter uppercased.
func (n ItemType) capitalized() string {
	s := string(n)
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}
