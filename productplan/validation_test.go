package productplan

import "testing"

func TestRequireNonEmpty(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{"valid value", "name", "Test", false},
		{"empty value", "name", "", true},
		{"whitespace only", "name", "   ", true},
		{"value with spaces", "name", "  Test  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireNonEmpty(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireNonEmpty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequireID(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{"valid ID", "roadmap_id", "123", false},
		{"empty ID", "roadmap_id", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireID(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequireAction(t *testing.T) {
	allowed := []string{"create", "update", "delete"}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid create", "create", false},
		{"valid update", "update", false},
		{"valid delete", "delete", false},
		{"valid uppercase", "CREATE", false},
		{"invalid action", "invalid", true},
		{"empty action", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireAction(tt.value, allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid date", "2024-06-30", false},
		{"empty date", "", false},
		{"invalid format slash", "2024/06/30", true},
		{"invalid format dots", "30.06.2024", true},
		{"incomplete date", "2024-06", true},
		{"too long", "2024-06-300", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDate("date", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateColor(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid uppercase", "#FF5733", false},
		{"valid lowercase", "#ff5733", false},
		{"valid mixed", "#Ff5733", false},
		{"empty", "", false},
		{"missing hash", "FF5733", true},
		{"too short", "#FFF", true},
		{"invalid chars", "#GGGGGG", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColor("color", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateColor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid https", "https://example.com", false},
		{"valid http", "http://example.com", false},
		{"empty", "", true},
		{"no protocol", "example.com", true},
		{"ftp protocol", "ftp://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL("url", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"empty", "", false},
		{"no at", "userexample.com", true},
		{"no domain", "user@", true},
		{"no user", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail("email", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	args := map[string]interface{}{
		"name": "Test",
		"id":   123,
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing string", "name", "Test"},
		{"non-existent key", "missing", ""},
		{"non-string value", "id", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetString(args, tt.key); got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStringSlice(t *testing.T) {
	tests := []struct {
		name string
		args map[string]interface{}
		key  string
		want []string
	}{
		{
			name: "string slice",
			args: map[string]interface{}{"tags": []string{"a", "b"}},
			key:  "tags",
			want: []string{"a", "b"},
		},
		{
			name: "interface slice",
			args: map[string]interface{}{"tags": []interface{}{"a", "b"}},
			key:  "tags",
			want: []string{"a", "b"},
		},
		{
			name: "missing key",
			args: map[string]interface{}{},
			key:  "tags",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringSlice(tt.args, tt.key)
			if len(got) != len(tt.want) {
				t.Errorf("GetStringSlice() len = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetStringSlice()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
