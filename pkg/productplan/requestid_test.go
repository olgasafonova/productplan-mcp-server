package productplan

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

func TestNewRequestID_FormatAndUniqueness(t *testing.T) {
	seen := make(map[RequestID]bool)

	for i := 0; i < 100; i++ {
		id := NewRequestID()

		// timestamp(4) + counter(2) + random(4) bytes, hex-encoded
		if len(id) != 20 {
			t.Fatalf("expected 20 hex chars, got %d (%q)", len(id), id)
		}
		if _, err := hex.DecodeString(string(id)); err != nil {
			t.Fatalf("id %q is not valid hex: %v", id, err)
		}
		if seen[id] {
			t.Fatalf("duplicate request ID %q after %d iterations", id, i)
		}
		seen[id] = true
	}
}

func TestRequestID_Short(t *testing.T) {
	tests := []struct {
		name string
		id   RequestID
		want string
	}{
		{"longer than 8", RequestID("0123456789abcdef0123"), "01234567"},
		{"exactly 8", RequestID("01234567"), "01234567"},
		{"shorter than 8", RequestID("0123"), "0123"},
		{"empty", RequestID(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Short(); got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestRequestID_String(t *testing.T) {
	id := RequestID("0123456789abcdef0123")
	if got := id.String(); got != "0123456789abcdef0123" {
		t.Errorf("expected full ID, got %q", got)
	}
}

func TestWithRequestID_RoundTrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), RequestID("abc123"))

	if got := GetRequestID(ctx); got != RequestID("abc123") {
		t.Errorf("expected abc123, got %q", got)
	}
}

func TestGetRequestID_Unset(t *testing.T) {
	if got := GetRequestID(context.Background()); got != "" {
		t.Errorf("expected empty ID on bare context, got %q", got)
	}
}

func TestEnsureRequestID_CreatesWhenMissing(t *testing.T) {
	ctx, id := EnsureRequestID(context.Background())

	if id == "" {
		t.Fatal("expected a generated ID")
	}
	if got := GetRequestID(ctx); got != id {
		t.Errorf("returned ID %q not stored in context (got %q)", id, got)
	}
}

func TestEnsureRequestID_PreservesExisting(t *testing.T) {
	original := RequestID("preexisting")
	ctx, id := EnsureRequestID(WithRequestID(context.Background(), original))

	if id != original {
		t.Errorf("expected existing ID %q, got %q", original, id)
	}
	if got := GetRequestID(ctx); got != original {
		t.Errorf("context ID changed to %q", got)
	}
}

func TestNewRequestTrace(t *testing.T) {
	ctx := WithRequestID(context.Background(), RequestID("trace-me"))
	trace := NewRequestTrace(ctx, "list_roadmaps")

	if trace.RequestID != RequestID("trace-me") {
		t.Errorf("expected request ID trace-me, got %q", trace.RequestID)
	}
	if trace.Operation != "list_roadmaps" {
		t.Errorf("expected operation list_roadmaps, got %q", trace.Operation)
	}
	if trace.StartTime.IsZero() {
		t.Error("expected StartTime to be set")
	}
}

func TestRequestTrace_Complete(t *testing.T) {
	trace := NewRequestTrace(context.Background(), "get_bar")
	trace.Complete(200, nil)

	if trace.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", trace.StatusCode)
	}
	if trace.EndTime.IsZero() {
		t.Error("expected EndTime to be set")
	}
	if trace.Duration < 0 {
		t.Errorf("expected non-negative duration, got %v", trace.Duration)
	}
	if trace.Error != "" {
		t.Errorf("expected empty error, got %q", trace.Error)
	}
}

func TestRequestTrace_CompleteWithError(t *testing.T) {
	trace := NewRequestTrace(context.Background(), "get_bar")
	trace.Complete(500, errors.New("upstream exploded"))

	if trace.Error != "upstream exploded" {
		t.Errorf("expected error text to be recorded, got %q", trace.Error)
	}
	if trace.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", trace.StatusCode)
	}
}

func TestRequestTrace_WithRetries(t *testing.T) {
	trace := NewRequestTrace(context.Background(), "get_bar").WithRetries(3)

	if trace.Retries != 3 {
		t.Errorf("expected 3 retries, got %d", trace.Retries)
	}
}

func TestRequestTracer_AddEvictsOldest(t *testing.T) {
	tracer := NewRequestTracer(2)

	tracer.Add(&RequestTrace{Operation: "first"})
	tracer.Add(&RequestTrace{Operation: "second"})
	tracer.Add(&RequestTrace{Operation: "third"})

	recent := tracer.Recent(10)
	if len(recent) != 2 {
		t.Fatalf("expected 2 traces retained, got %d", len(recent))
	}
	if recent[0].Operation != "second" || recent[1].Operation != "third" {
		t.Errorf("expected [second third], got [%s %s]", recent[0].Operation, recent[1].Operation)
	}
}

func TestRequestTracer_RecentClampsToLength(t *testing.T) {
	tracer := NewRequestTracer(10)
	tracer.Add(&RequestTrace{Operation: "only"})

	if got := len(tracer.Recent(5)); got != 1 {
		t.Errorf("expected 1 trace, got %d", got)
	}
	if got := len(tracer.Recent(0)); got != 0 {
		t.Errorf("expected 0 traces, got %d", got)
	}
}

func TestRequestTracer_Clear(t *testing.T) {
	tracer := NewRequestTracer(4)
	tracer.Add(&RequestTrace{Operation: "one"})
	tracer.Clear()

	if got := len(tracer.Recent(4)); got != 0 {
		t.Errorf("expected no traces after Clear, got %d", got)
	}
}

func TestRequestTracer_StatsEmpty(t *testing.T) {
	stats := NewRequestTracer(4).Stats()

	if stats.TotalRequests != 0 {
		t.Errorf("expected 0 requests, got %d", stats.TotalRequests)
	}
	if stats.AvgDuration != 0 {
		t.Errorf("expected zero average duration, got %v", stats.AvgDuration)
	}
}

func TestRequestTracer_StatsAggregates(t *testing.T) {
	tracer := NewRequestTracer(4)
	tracer.Add(&RequestTrace{Duration: 100 * time.Millisecond, Retries: 1})
	tracer.Add(&RequestTrace{Duration: 300 * time.Millisecond, Retries: 2, Error: "boom"})

	stats := tracer.Stats()

	if stats.TotalRequests != 2 {
		t.Errorf("expected 2 requests, got %d", stats.TotalRequests)
	}
	if stats.Errors != 1 {
		t.Errorf("expected 1 error, got %d", stats.Errors)
	}
	if stats.TotalRetries != 3 {
		t.Errorf("expected 3 retries, got %d", stats.TotalRetries)
	}
	if stats.AvgDuration != 200*time.Millisecond {
		t.Errorf("expected 200ms average, got %v", stats.AvgDuration)
	}
}
