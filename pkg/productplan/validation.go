package productplan

import (
	"regexp"
	"strings"
)

// RequireNonEmpty validates that a string field is not empty.
func RequireNonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(field, "is required and cannot be empty")
	}
	return nil
}

// RequireID validates that an ID field is non-empty and valid.
func RequireID(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(field, "is required - get it from the corresponding list_* tool")
	}
	return nil
}

// RequireRoadmapID validates a roadmap_id field.
func RequireRoadmapID(value string) error {
	return RequireID("roadmap_id", value)
}

// RequireBarID validates a bar_id field.
func RequireBarID(value string) error {
	return RequireID("bar_id", value)
}

// RequireLaneID validates a lane_id field.
func RequireLaneID(value string) error {
	return RequireID("lane_id", value)
}

// RequireObjectiveID validates an objective_id field.
func RequireObjectiveID(value string) error {
	return RequireID("objective_id", value)
}

// RequireIdeaID validates an idea_id field.
func RequireIdeaID(value string) error {
	return RequireID("idea_id", value)
}

// RequireAction validates an action field with allowed values.
func RequireAction(value string, allowed []string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError("action", "is required - must be one of: "+strings.Join(allowed, ", "))
	}

	v := strings.ToLower(strings.TrimSpace(value))
	for _, a := range allowed {
		if v == strings.ToLower(a) {
			return nil
		}
	}

	return NewValidationError("action", "must be one of: "+strings.Join(allowed, ", "))
}

// ValidateDate checks if a date string is in YYYY-MM-DD format.
func ValidateDate(field, value string) error {
	if value == "" {
		return nil // Optional field
	}

	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !datePattern.MatchString(value) {
		return NewValidationError(field, "must be in YYYY-MM-DD format (e.g., 2024-06-30)")
	}
	return nil
}

// ValidateColor checks if a color string is a valid hex color.
func ValidateColor(field, value string) error {
	if value == "" {
		return nil // Optional field
	}

	colorPattern := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	if !colorPattern.MatchString(value) {
		return NewValidationError(field, "must be a hex color code (e.g., #FF5733)")
	}
	return nil
}

// ValidateURL checks if a URL string is valid.
func ValidateURL(field, value string) error {
	if value == "" {
		return NewValidationError(field, "is required")
	}

	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		return NewValidationError(field, "must be a valid URL starting with http:// or https://")
	}
	return nil
}

// ValidateEmail checks if an email string is valid.
func ValidateEmail(field, value string) error {
	if value == "" {
		return nil // Optional field
	}

	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(value) {
		return NewValidationError(field, "must be a valid email address")
	}
	return nil
}

// GetString safely extracts a string from a map.
func GetString(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetStringSlice safely extracts a string slice from a map.
func GetStringSlice(args map[string]interface{}, key string) []string {
	if v, ok := args[key]; ok {
		switch val := v.(type) {
		case []string:
			return val
		case []interface{}:
			result := make([]string, 0, len(val))
			for _, item := range val {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}
