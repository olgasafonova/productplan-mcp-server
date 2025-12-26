package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "debug"},
		{LevelInfo, "info"},
		{LevelWarn, "warn"},
		{LevelError, "error"},
		{Level(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  Level
	}{
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"error", LevelError},
		{"invalid", LevelInfo}, // default to info
		{"", LevelInfo},
	}

	for _, tt := range tests {
		if got := ParseLevel(tt.input); got != tt.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestFieldConstructors(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		key   string
		value any
	}{
		{"RequestID", RequestID("abc123"), "req_id", "abc123"},
		{"Operation", Operation("get_roadmap"), "op", "get_roadmap"},
		{"Duration", Duration(100 * time.Millisecond), "dur_ms", int64(100)},
		{"Status", Status("ok"), "status", "ok"},
		{"Error", Error(errors.New("test error")), "error", "test error"},
		{"Tool", Tool("list_roadmaps"), "tool", "list_roadmaps"},
		{"Endpoint", Endpoint("/api/v2/roadmaps"), "endpoint", "/api/v2/roadmaps"},
		{"StatusCode", StatusCode(200), "status_code", 200},
		{"Count", Count(42), "count", 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.Key != tt.key {
				t.Errorf("got key %q, want %q", tt.field.Key, tt.key)
			}
			if tt.field.Value != tt.value {
				t.Errorf("got value %v, want %v", tt.field.Value, tt.value)
			}
		})
	}
}

func TestJSONLoggerLevels(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  Level
		callLevel string
		shouldLog bool
	}{
		{"debug at debug level", LevelDebug, "debug", true},
		{"info at debug level", LevelDebug, "info", true},
		{"debug at info level", LevelInfo, "debug", false},
		{"info at info level", LevelInfo, "info", true},
		{"warn at info level", LevelInfo, "warn", true},
		{"error at warn level", LevelWarn, "error", true},
		{"warn at error level", LevelError, "warn", false},
		{"error at error level", LevelError, "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewWithWriter(&buf, tt.logLevel)

			switch tt.callLevel {
			case "debug":
				logger.Debug("test message")
			case "info":
				logger.Info("test message")
			case "warn":
				logger.Warn("test message")
			case "error":
				logger.Error("test message")
			}

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("expected log output = %v, got output = %v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestJSONLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelDebug)

	logger.Info("test message", F("key1", "value1"), Count(10))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	// Verify required fields
	if entry["level"] != "info" {
		t.Errorf("expected level 'info', got %v", entry["level"])
	}
	if entry["msg"] != "test message" {
		t.Errorf("expected msg 'test message', got %v", entry["msg"])
	}
	if entry["key1"] != "value1" {
		t.Errorf("expected key1 'value1', got %v", entry["key1"])
	}
	if entry["count"] != float64(10) { // JSON numbers are float64
		t.Errorf("expected count 10, got %v", entry["count"])
	}
	if _, ok := entry["ts"]; !ok {
		t.Error("expected timestamp field 'ts'")
	}
}

func TestJSONLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelDebug)

	// Create child logger with base fields
	child := logger.WithFields(RequestID("req-123"), Operation("test_op"))

	child.Info("child message", Status("ok"))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if entry["req_id"] != "req-123" {
		t.Errorf("expected req_id 'req-123', got %v", entry["req_id"])
	}
	if entry["op"] != "test_op" {
		t.Errorf("expected op 'test_op', got %v", entry["op"])
	}
	if entry["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", entry["status"])
	}
}

func TestJSONLoggerWithRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelDebug)

	child := logger.WithRequestID("my-request-id")
	child.Info("request log")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if entry["req_id"] != "my-request-id" {
		t.Errorf("expected req_id 'my-request-id', got %v", entry["req_id"])
	}
}

func TestJSONLoggerFieldOverride(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelDebug)

	// Base field
	child := logger.WithFields(F("key", "base"))

	// Override with call-specific field
	child.Info("message", F("key", "override"))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	// Call-specific field should override base
	if entry["key"] != "override" {
		t.Errorf("expected key 'override', got %v", entry["key"])
	}
}

func TestJSONLoggerTimestamp(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelDebug)

	before := time.Now().UTC()
	logger.Info("timestamp test")
	after := time.Now().UTC()

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	ts, ok := entry["ts"].(string)
	if !ok {
		t.Fatal("expected timestamp to be string")
	}

	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t.Fatalf("failed to parse timestamp: %v", err)
	}

	if parsed.Before(before) || parsed.After(after) {
		t.Errorf("timestamp %v not between %v and %v", parsed, before, after)
	}
}

func TestJSONLoggerSetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelInfo)

	// Debug should not log at Info level
	logger.Debug("should not appear")
	if buf.Len() > 0 {
		t.Error("debug should not log at info level")
	}

	// Change to Debug level
	logger.SetLevel(LevelDebug)

	if logger.GetLevel() != LevelDebug {
		t.Error("GetLevel should return LevelDebug")
	}

	logger.Debug("should appear now")
	if buf.Len() == 0 {
		t.Error("debug should log after level change")
	}
}

func TestNopLogger(t *testing.T) {
	logger := Nop()

	// These should not panic
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")

	child := logger.WithFields(F("key", "value"))
	child.Info("test")

	child2 := logger.WithRequestID("req-id")
	child2.Info("test")
}

func TestNew(t *testing.T) {
	logger := New(LevelInfo)
	if logger == nil {
		t.Fatal("New should return non-nil logger")
	}
	if logger.level != LevelInfo {
		t.Errorf("expected level %v, got %v", LevelInfo, logger.level)
	}
}

func TestJSONLoggerNewlineTerminated(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelDebug)

	logger.Info("message1")
	logger.Info("message2")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	// Each line should be valid JSON
	for i, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestFField(t *testing.T) {
	f := F("custom_key", 123)
	if f.Key != "custom_key" {
		t.Errorf("expected key 'custom_key', got %q", f.Key)
	}
	if f.Value != 123 {
		t.Errorf("expected value 123, got %v", f.Value)
	}
}

func BenchmarkJSONLoggerInfo(b *testing.B) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.Info("benchmark message", F("iteration", i), Status("ok"))
	}
}

func BenchmarkJSONLoggerWithFields(b *testing.B) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelInfo)
	child := logger.WithFields(RequestID("benchmark-req"), Operation("benchmark"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		child.Info("benchmark message", F("iteration", i))
	}
}
