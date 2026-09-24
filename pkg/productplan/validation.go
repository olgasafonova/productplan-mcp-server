package productplan

import (
	"regexp"
	"strings"
)

const (
	// MaxIDLen is the maximum allowed length for any ID field. ProductPlan
	// IDs are short alphanumeric tokens; anything longer is suspicious.
	MaxIDLen = 100
)

// validIDPattern matches safe ProductPlan IDs. ProductPlan uses short
// alphanumeric tokens with optional underscore/hyphen. Any character outside
// this set is rejected because it can pivot the URL via Go's URL parser
// (e.g. `?` splits into a query, `..` traverses, `#` becomes a fragment).
var validIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Field names a caller-supplied argument as it appears in validation
// errors: "bar_id", or a position inside a list such as "bar_ids[3].bar_id".
// Validation hangs off the field so every error names what the caller must
// fix.
type Field string

// RequireID validates that value is a usable ID for f: non-empty after
// trimming, at most MaxIDLen long, and made only of letters, digits,
// underscore and hyphen. This is the chokepoint that prevents path
// injection: the typed IDs in internal/api run every user-supplied ID
// through it before the ID can become part of a URL.
func (f Field) RequireID(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return f.invalid("is required - get it from the corresponding list_* tool")
	}
	if len(trimmed) > MaxIDLen {
		return f.invalid("is too long")
	}
	if !validIDPattern.MatchString(trimmed) {
		return f.invalid("contains invalid characters (only letters, digits, underscore, hyphen are allowed)")
	}
	return nil
}

// invalid builds the ValidationError for f.
func (f Field) invalid(message string) *ValidationError {
	return NewValidationError(string(f), message)
}
