// Package productplan provides API client utilities for ProductPlan MCP server.
package productplan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// APIError represents a structured error from the ProductPlan API.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	RetryAfter int    `json:"retry_after,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("ProductPlan API error %d: %s - %s", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("ProductPlan API error %d: %s", e.StatusCode, e.Message)
}

// IsRateLimited returns true if the error is due to rate limiting.
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == 429
}

// IsNotFound returns true if the resource was not found.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsUnauthorized returns true if the request was unauthorized.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// IsForbidden returns true if access is forbidden.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == 403
}

// IsServerError returns true if the error is a server-side error.
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500
}

// IsRetryable returns true if the request can be retried.
func (e *APIError) IsRetryable() bool {
	return e.IsRateLimited() || e.IsServerError()
}

// Suggestion returns actionable guidance for handling this error.
func (e *APIError) Suggestion() string {
	switch {
	case e.IsRateLimited():
		if e.RetryAfter > 0 {
			return fmt.Sprintf("Rate limited. Wait %d seconds before retrying.", e.RetryAfter)
		}
		return "Rate limited. Wait 60 seconds before retrying."

	case e.IsNotFound():
		return "Resource not found. Verify the ID is correct using the list_* tools."

	case e.IsUnauthorized():
		return "Invalid or expired API token. Check PRODUCTPLAN_API_TOKEN environment variable."

	case e.IsForbidden():
		return "Access denied. Your API token may not have permission for this operation."

	case e.StatusCode == 400:
		return "Invalid request. Check required parameters and their formats."

	case e.StatusCode == 422:
		return "Validation error. Check the field values match expected formats."

	case e.IsServerError():
		return "ProductPlan server error. Try again in a few moments."

	default:
		return ""
	}
}

// maxClientFacingDetailLen caps the body excerpt that may reach an MCP caller
// when the API returns a non-JSON error response. The full body is still
// available to operators via server-side request logging.
const maxClientFacingDetailLen = 200

// sanitizeBodyForCaller produces a short, single-line excerpt of a non-JSON
// response body, safe to surface to the MCP caller. It strips at the first
// newline (HTML error pages and stack traces are common) and caps total
// length so the agent sees a compact error rather than an unbounded payload.
// Per HG-2 in code-review-prompts.md: raw response bodies must not reach the
// MCP caller verbatim.
func sanitizeBodyForCaller(body []byte) string {
	excerpt := strings.TrimSpace(string(body))
	if i := strings.IndexAny(excerpt, "\r\n"); i >= 0 {
		excerpt = excerpt[:i]
	}
	if len(excerpt) > maxClientFacingDetailLen {
		excerpt = excerpt[:maxClientFacingDetailLen] + "..."
	}
	return excerpt
}

// apiErrorBody mirrors the conventional JSON error shape returned by the API.
type apiErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Details string `json:"details"`
}

// applyErrorBody copies non-empty fields from the parsed body onto apiErr.
// Message prefers `message` over `error` (latter wins last-write semantics).
func applyErrorBody(apiErr *APIError, errBody apiErrorBody) {
	if errBody.Error != "" {
		apiErr.Message = errBody.Error
	}
	if errBody.Message != "" {
		apiErr.Message = errBody.Message
	}
	if errBody.Code != "" {
		apiErr.Code = errBody.Code
	}
	if errBody.Details != "" {
		apiErr.Details = errBody.Details
	}
}

// applyRetryAfter parses the Retry-After header (integer seconds form) into apiErr.
func applyRetryAfter(apiErr *APIError, resp *http.Response) {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return
	}
	seconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		return
	}
	apiErr.RetryAfter = seconds
}

// ParseAPIError creates an APIError from an HTTP response.
func ParseAPIError(resp *http.Response, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Message:    http.StatusText(resp.StatusCode),
	}

	var errBody apiErrorBody
	if err := json.Unmarshal(body, &errBody); err == nil {
		applyErrorBody(apiErr, errBody)
	} else if len(body) > 0 {
		apiErr.Details = sanitizeBodyForCaller(body)
	}

	applyRetryAfter(apiErr, resp)
	return apiErr
}

// ValidationError represents an input validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a validation error.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}
