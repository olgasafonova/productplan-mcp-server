package api

import (
	"context"
	"strings"
	"testing"
)

// TestSafeSeg_RejectsInjection is the regression test for the path-injection
// finding (Carlini scan, productplan Finding 2). Each payload must be
// rejected before the request reaches the API.
func TestSafeSeg_RejectsInjection(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{"path traversal", "../../strategy/objectives/SECRET"},
		{"sub-resource pivot", "X/comments/Y"},
		{"query injection", "123?expand=*&force=true"},
		{"hash anchor", "valid#anchor"},
		{"semicolon param", "valid;param"},
		{"ampersand", "valid&query=injected"},
		{"newline injection", "valid\nX-Inject: 1"},
		{"null byte", "valid\x00.attacker"},
		{"empty", ""},
		{"only whitespace", "   "},
		{"oversize", strings.Repeat("a", 101)},
		{"forward slash", "a/b"},
		{"percent encoding", "valid%2Fextra"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			seg, err := safeSeg("test_id", c.id)
			if err == nil {
				t.Errorf("safeSeg(%q) returned nil error; expected rejection. got seg=%q", c.id, seg)
			}
		})
	}
}

// TestSafeSeg_AcceptsRealisticIDs guards the success path. ProductPlan IDs
// in the wild are short alphanumeric tokens, sometimes with hyphens or
// underscores. None of these should be rejected by the validator.
func TestSafeSeg_AcceptsRealisticIDs(t *testing.T) {
	realisticIDs := []string{
		"wbihRzEYTdOOOLXTeyc8",
		"abc123",
		"user_id_123",
		"hyphen-id-456",
		"BAR-789",
		"123",
		"a",
	}

	for _, id := range realisticIDs {
		t.Run(id, func(t *testing.T) {
			seg, err := safeSeg("test_id", id)
			if err != nil {
				t.Errorf("safeSeg(%q) returned %v; expected nil for realistic ID", id, err)
			}
			if seg == "" {
				t.Errorf("safeSeg(%q) returned empty string", id)
			}
		})
	}
}

// TestEndpoint_RejectsInjectionBeforeHTTP exercises the integrated path
// through one representative endpoint method. Asserts the validator rejects
// the malicious bar_id before any HTTP layer is touched (the Client has no
// transport configured, so any request that reaches the wire would error
// differently).
func TestEndpoint_RejectsInjectionBeforeHTTP(t *testing.T) {
	c := &Client{baseURL: "https://example.invalid"}

	maliciousID := "../../strategy/objectives/SECRET"
	_, err := c.GetBar(context.Background(), maliciousID)

	if err == nil {
		t.Fatal("GetBar accepted path-injection payload; expected validator rejection")
	}
	if !strings.Contains(err.Error(), "bar_id") {
		t.Errorf("error did not mention bar_id: %v", err)
	}
	if strings.Contains(err.Error(), "SECRET") {
		t.Errorf("error leaked the injected payload: %v", err)
	}
}

// TestEndpoint_RejectsInjectionInSecondID covers the multi-ID case (e.g.,
// DeleteKeyResult takes both objective_id and key_result_id). A malicious
// caller might supply a valid first ID and a malicious second.
func TestEndpoint_RejectsInjectionInSecondID(t *testing.T) {
	c := &Client{baseURL: "https://example.invalid"}

	_, err := c.DeleteKeyResult(context.Background(), "validObjective", "../launches/X")

	if err == nil {
		t.Fatal("DeleteKeyResult accepted injection in key_result_id; expected validator rejection")
	}
	if !strings.Contains(err.Error(), "key_result_id") {
		t.Errorf("error did not mention key_result_id: %v", err)
	}
}
