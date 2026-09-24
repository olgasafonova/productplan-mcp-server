package productplan

import (
	"errors"
	"strings"
	"testing"
)

func TestFieldRequireID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantMsg string // "" means valid
	}{
		{"numeric", "123", ""},
		{"token", "wbihRzEYTdOOOLXTeyc8", ""},
		{"underscore and hyphen", "a_b-c", ""},
		{"surrounding space is trimmed", "  42  ", ""},
		{"exactly max length", strings.Repeat("a", MaxIDLen), ""},
		{"empty", "", "is required - get it from the corresponding list_* tool"},
		{"whitespace only", "   ", "is required - get it from the corresponding list_* tool"},
		{"too long", strings.Repeat("a", MaxIDLen+1), "is too long"},
		{"path traversal", "../x", "contains invalid characters"},
		{"query", "1?x=y", "contains invalid characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Field("roadmap_id").RequireID(tt.value)
			if tt.wantMsg == "" {
				if err != nil {
					t.Fatalf("RequireID(%q) = %v, want nil", tt.value, err)
				}
				return
			}
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("RequireID(%q) = %v, want *ValidationError", tt.value, err)
			}
			if ve.Field != "roadmap_id" {
				t.Errorf("Field = %q, want roadmap_id", ve.Field)
			}
			if !strings.HasPrefix(ve.Message, tt.wantMsg) {
				t.Errorf("Message = %q, want prefix %q", ve.Message, tt.wantMsg)
			}
		})
	}
}

// TestFieldRequireID_ErrorText pins the exact wording agents and evals see.
func TestFieldRequireID_ErrorText(t *testing.T) {
	err := Field("bar_id").RequireID("a/b")
	want := "validation error for 'bar_id': contains invalid characters (only letters, digits, underscore, hyphen are allowed)"
	if err == nil || err.Error() != want {
		t.Fatalf("got %v, want %q", err, want)
	}
}
