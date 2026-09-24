package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// decodeLines parses each JSON line written to buf.
func decodeLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("line is not JSON: %q: %v", line, err)
		}
		out = append(out, m)
	}
	return out
}

func TestLoggerOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	NewWithWriter(&buf, slog.LevelInfo).Info("API request",
		slog.String("endpoint", "/roadmaps"), Duration(245*time.Millisecond), Error(errors.New("boom")))

	lines := decodeLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	e := lines[0]
	for key, want := range map[string]any{
		"level": "info", "msg": "API request", "endpoint": "/roadmaps", "dur_ms": float64(245), "error": "boom",
	} {
		if e[key] != want {
			t.Errorf("%s = %#v, want %#v", key, e[key], want)
		}
	}
	if _, ok := e["time"]; ok {
		t.Error(`slog's "time" key leaked; want "ts"`)
	}
}

func TestLoggerTimestampIsUTCRFC3339Nano(t *testing.T) {
	var buf bytes.Buffer
	before := time.Now().UTC()
	NewWithWriter(&buf, slog.LevelInfo).Info("x")
	ts, _ := decodeLines(t, &buf)[0]["ts"].(string)
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t.Fatalf("ts %q is not RFC 3339: %v", ts, err)
	}
	if !strings.HasSuffix(ts, "Z") {
		t.Errorf("ts %q is not UTC", ts)
	}
	if parsed.Before(before.Add(-time.Second)) {
		t.Errorf("ts %q is earlier than the call", ts)
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	l := NewWithWriter(&buf, slog.LevelWarn)
	l.Debug("d")
	l.Info("i")
	l.Warn("w")
	l.Error("e")

	lines := decodeLines(t, &buf)
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want warn and error only: %s", len(lines), buf.String())
	}
	if lines[0]["level"] != "warn" || lines[1]["level"] != "error" {
		t.Errorf("levels = %v, %v; want warn, error", lines[0]["level"], lines[1]["level"])
	}
}

func TestLoggerGroupedAttrsKeepTheirKeys(t *testing.T) {
	var buf bytes.Buffer
	NewWithWriter(&buf, slog.LevelInfo).Info("x", slog.Group("req", slog.String("time", "t1")))
	req, _ := decodeLines(t, &buf)[0]["req"].(map[string]any)
	if req["time"] != "t1" {
		t.Errorf("grouped time attr rewritten: %v", req)
	}
}

func TestLoggerNewlineTerminated(t *testing.T) {
	var buf bytes.Buffer
	l := NewWithWriter(&buf, slog.LevelInfo)
	l.Info("one")
	l.Info("two")
	if got := strings.Count(buf.String(), "\n"); got != 2 {
		t.Errorf("got %d newlines, want 2", got)
	}
}

func TestNop(t *testing.T) {
	l := Nop()
	l.Error("discarded", Error(errors.New("x")))
	if l.Enabled(t.Context(), slog.LevelError) {
		t.Error("Nop logger reports itself enabled")
	}
}

func TestNew(t *testing.T) {
	if New(slog.LevelInfo) == nil {
		t.Fatal("New returned nil")
	}
}
