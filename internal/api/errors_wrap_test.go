package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/pkg/productplan"
)

// A 4xx carrying a suggestion must still wrap *productplan.APIError, so
// callers classify it with errors.As instead of parsing message text.
func TestHandleResponse_WrapsAPIErrorWithSuggestion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Bar not found"}`))
	}))
	defer srv.Close()
	c, err := New(Config{Token: "t", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.get(context.Background(), "/bars/1")
	if err == nil {
		t.Fatal("want an error for a 404")
	}
	var apiErr *productplan.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As(*APIError) failed on %q", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", apiErr.StatusCode)
	}
	if !strings.Contains(err.Error(), "Verify the ID") {
		t.Errorf("suggestion missing from %q", err)
	}
	if !IsNotFound(err) {
		t.Error("IsNotFound = false for a wrapped 404")
	}
	if !IsNotFound(fmt.Errorf("outer: %w", err)) {
		t.Error("IsNotFound = false through a second wrap")
	}
}

// IsNotFound is type-based: text that merely looks like a 404 is not one,
// and other statuses are not 404s.
func TestIsNotFound_TypedOnly(t *testing.T) {
	for name, tc := range map[string]struct {
		err  error
		want bool
	}{
		"nil":               {nil, false},
		"plain text 404":    {errors.New("ProductPlan API error 404: nope"), false},
		"typed 404":         {&productplan.APIError{StatusCode: 404}, true},
		"typed 422":         {&productplan.APIError{StatusCode: 422}, false},
		"wrapped typed 404": {fmt.Errorf("x: %w", &productplan.APIError{StatusCode: 404}), true},
	} {
		t.Run(name, func(t *testing.T) {
			if got := IsNotFound(tc.err); got != tc.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
