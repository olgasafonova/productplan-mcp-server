package productplan

import (
	"net/http"
	"strings"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   APIError
		expected string
	}{
		{
			name:     "basic error",
			apiErr:   APIError{StatusCode: 404, Message: "Not Found"},
			expected: "ProductPlan API error 404: Not Found",
		},
		{
			name:     "error with details",
			apiErr:   APIError{StatusCode: 400, Message: "Bad Request", Details: "Invalid ID format"},
			expected: "ProductPlan API error 400: Bad Request - Invalid ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.apiErr.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIError_IsRateLimited(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"rate limited", 429, true},
		{"not found", 404, false},
		{"server error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &APIError{StatusCode: tt.statusCode}
			if got := e.IsRateLimited(); got != tt.want {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIError_IsNotFound(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"not found", 404, true},
		{"rate limited", 429, false},
		{"ok", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &APIError{StatusCode: tt.statusCode}
			if got := e.IsNotFound(); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIError_IsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"rate limited", 429, true},
		{"server error", 500, true},
		{"bad gateway", 502, true},
		{"not found", 404, false},
		{"bad request", 400, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &APIError{StatusCode: tt.statusCode}
			if got := e.IsRetryable(); got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIError_Suggestion(t *testing.T) {
	tests := []struct {
		name       string
		apiErr     APIError
		wantEmpty  bool
		wantPrefix string
	}{
		{
			name:       "rate limited with retry",
			apiErr:     APIError{StatusCode: 429, RetryAfter: 30},
			wantPrefix: "Rate limited. Wait 30 seconds",
		},
		{
			name:       "rate limited no retry",
			apiErr:     APIError{StatusCode: 429},
			wantPrefix: "Rate limited. Wait 60 seconds",
		},
		{
			name:       "not found",
			apiErr:     APIError{StatusCode: 404},
			wantPrefix: "Resource not found",
		},
		{
			name:       "unauthorized",
			apiErr:     APIError{StatusCode: 401},
			wantPrefix: "Invalid or expired API token",
		},
		{
			name:      "unknown",
			apiErr:    APIError{StatusCode: 418}, // I'm a teapot
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiErr.Suggestion()
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("Suggestion() = %v, want empty", got)
				}
				return
			}
			if len(got) < len(tt.wantPrefix) || got[:len(tt.wantPrefix)] != tt.wantPrefix {
				t.Errorf("Suggestion() = %v, want prefix %v", got, tt.wantPrefix)
			}
		})
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		headers    map[string]string
		wantCode   int
		wantMsg    string
	}{
		{
			name:       "json error",
			statusCode: 400,
			body:       `{"error": "Invalid parameter", "code": "INVALID_PARAM"}`,
			wantCode:   400,
			wantMsg:    "Invalid parameter",
		},
		{
			name:       "plain text error",
			statusCode: 500,
			body:       "Internal Server Error",
			wantCode:   500,
			wantMsg:    "Internal Server Error",
		},
		{
			name:       "rate limit with retry-after",
			statusCode: 429,
			body:       `{"message": "Too Many Requests"}`,
			headers:    map[string]string{"Retry-After": "60"},
			wantCode:   429,
			wantMsg:    "Too Many Requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     make(http.Header),
			}
			for k, v := range tt.headers {
				resp.Header.Set(k, v)
			}

			got := ParseAPIError(resp, []byte(tt.body))

			if got.StatusCode != tt.wantCode {
				t.Errorf("StatusCode = %v, want %v", got.StatusCode, tt.wantCode)
			}
			if got.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", got.Message, tt.wantMsg)
			}
		})
	}
}

func TestParseAPIError_NonJSONBodyIsSanitized(t *testing.T) {
	// HG-2: when the API returns a non-JSON body, the raw payload must not
	// reach the MCP caller verbatim. ParseAPIError must truncate at the first
	// newline and cap total length so multi-line HTML error pages or stack
	// traces don't flow through Error().
	tests := []struct {
		name           string
		body           string
		wantDetails    string
		wantContains   string // substring check when exact size depends on length
		mustNotContain string
	}{
		{
			name:        "single short line passes through",
			body:        "Internal Server Error",
			wantDetails: "Internal Server Error",
		},
		{
			name:        "multi-line body strips at first newline",
			body:        "Internal Server Error\n<html><body>...stack trace...</body></html>",
			wantDetails: "Internal Server Error",
		},
		{
			name:        "carriage-return-newline body strips at first delimiter",
			body:        "Bad Gateway\r\nUpstream connection refused",
			wantDetails: "Bad Gateway",
		},
		{
			name:           "long body is capped",
			body:           strings.Repeat("X", 500),
			wantContains:   "...",
			mustNotContain: strings.Repeat("X", 300),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: 500, Header: make(http.Header)}
			got := ParseAPIError(resp, []byte(tt.body))

			if tt.wantDetails != "" && got.Details != tt.wantDetails {
				t.Errorf("Details = %q, want %q", got.Details, tt.wantDetails)
			}
			if tt.wantContains != "" && !strings.Contains(got.Details, tt.wantContains) {
				t.Errorf("Details = %q, expected to contain %q", got.Details, tt.wantContains)
			}
			if tt.mustNotContain != "" && strings.Contains(got.Details, tt.mustNotContain) {
				t.Errorf("Details = %q, must not contain a 300-char run of input", got.Details)
			}
			if len(got.Details) > maxClientFacingDetailLen+3 {
				t.Errorf("Details length = %d, exceeds cap %d (+3 for ellipsis)", len(got.Details), maxClientFacingDetailLen)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("roadmap_id", "is required")

	if err.Field != "roadmap_id" {
		t.Errorf("Field = %v, want roadmap_id", err.Field)
	}

	expected := "validation error for 'roadmap_id': is required"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}
