package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"
	"testing"
)

// Article II: every recovered panic gets its own reference, present in both
// the log line and the caller's error, and the panic value never reaches
// the caller.
func TestRegistryCall_PanicCarriesReference(t *testing.T) {
	logs := captureSlog(t)
	r := NewRegistry()
	r.RegisterFunc(Tool{Name: "explodes"}, func(context.Context, map[string]any) (json.RawMessage, error) {
		panic("secret-token-xyz")
	})

	_, err1 := r.Call(context.Background(), "explodes", nil)
	_, err2 := r.Call(context.Background(), "explodes", nil)
	if err1 == nil || err2 == nil {
		t.Fatal("panic produced no error")
	}
	ref1, ref2 := panicRefIn(t, err1.Error()), panicRefIn(t, err2.Error())
	if ref1 == ref2 {
		t.Errorf("two panics share ref %s", ref1)
	}
	for _, ref := range []string{ref1, ref2} {
		if !strings.Contains(logs.String(), `"ref":"`+ref+`"`) {
			t.Errorf("ref %s missing from log", ref)
		}
	}
	if strings.Contains(err1.Error(), "secret-token-xyz") {
		t.Errorf("panic value leaked: %q", err1)
	}
	if want := "internal error in explodes (ref " + ref1 + ")"; err1.Error() != want {
		t.Errorf("error = %q, want %q", err1, want)
	}
}

var panicRefPattern = regexp.MustCompile(`\(ref ([0-9a-f]{16})\)`)

// panicRefIn extracts the 16-hex-digit reference from an internal error.
func panicRefIn(t *testing.T, msg string) string {
	t.Helper()
	m := panicRefPattern.FindStringSubmatch(msg)
	if m == nil {
		t.Fatalf("no (ref <16 hex>) in %q", msg)
	}
	return m[1]
}

// captureSlog routes the default slog logger to a JSON buffer for the test.
// Tests using it must not run in parallel.
func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}
