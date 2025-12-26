package productplan

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
	"time"
)

// RequestID represents a unique identifier for a request.
type RequestID string

// requestIDKey is the context key for request IDs.
type requestIDKey struct{}

var requestCounter uint64

// NewRequestID generates a new unique request ID.
// Format: timestamp(hex)-counter-random
func NewRequestID() RequestID {
	// Increment counter atomically
	count := atomic.AddUint64(&requestCounter, 1)

	// Generate 4 random bytes
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes) // Ignore error; crypto/rand.Read only fails on OS entropy exhaustion

	// Combine timestamp, counter, and random
	ts := time.Now().UnixMilli() & 0xFFFFFFFF // Lower 32 bits
	return RequestID(hex.EncodeToString([]byte{
		byte(ts >> 24), byte(ts >> 16), byte(ts >> 8), byte(ts),
		byte(count >> 8), byte(count),
	}) + hex.EncodeToString(randomBytes))
}

// Short returns a shortened version of the request ID (first 8 chars).
func (r RequestID) Short() string {
	if len(r) >= 8 {
		return string(r)[:8]
	}
	return string(r)
}

// String returns the full request ID.
func (r RequestID) String() string {
	return string(r)
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, id RequestID) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// GetRequestID retrieves the request ID from context.
// Returns empty string if not set.
func GetRequestID(ctx context.Context) RequestID {
	if id, ok := ctx.Value(requestIDKey{}).(RequestID); ok {
		return id
	}
	return ""
}

// EnsureRequestID gets existing request ID or creates a new one.
func EnsureRequestID(ctx context.Context) (context.Context, RequestID) {
	if id := GetRequestID(ctx); id != "" {
		return ctx, id
	}
	id := NewRequestID()
	return WithRequestID(ctx, id), id
}

// RequestTrace tracks timing and metadata for a request.
type RequestTrace struct {
	RequestID  RequestID     `json:"request_id"`
	Operation  string        `json:"operation"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time,omitempty"`
	Duration   time.Duration `json:"duration_ms,omitempty"`
	StatusCode int           `json:"status_code,omitempty"`
	Error      string        `json:"error,omitempty"`
	Retries    int           `json:"retries,omitempty"`
}

// NewRequestTrace creates a new trace for an operation.
func NewRequestTrace(ctx context.Context, operation string) *RequestTrace {
	return &RequestTrace{
		RequestID: GetRequestID(ctx),
		Operation: operation,
		StartTime: time.Now(),
	}
}

// Complete marks the trace as finished.
func (t *RequestTrace) Complete(statusCode int, err error) {
	t.EndTime = time.Now()
	t.Duration = t.EndTime.Sub(t.StartTime)
	t.StatusCode = statusCode
	if err != nil {
		t.Error = err.Error()
	}
}

// WithRetries sets the retry count.
func (t *RequestTrace) WithRetries(count int) *RequestTrace {
	t.Retries = count
	return t
}

// TracingEnabled controls whether tracing is active.
var TracingEnabled = false

// RequestTracer collects request traces.
type RequestTracer struct {
	traces    []*RequestTrace
	maxTraces int
}

// NewRequestTracer creates a tracer that keeps the last N traces.
func NewRequestTracer(maxTraces int) *RequestTracer {
	return &RequestTracer{
		traces:    make([]*RequestTrace, 0, maxTraces),
		maxTraces: maxTraces,
	}
}

// Add adds a trace, evicting oldest if at capacity.
func (rt *RequestTracer) Add(trace *RequestTrace) {
	if len(rt.traces) >= rt.maxTraces {
		rt.traces = rt.traces[1:]
	}
	rt.traces = append(rt.traces, trace)
}

// Recent returns the N most recent traces.
func (rt *RequestTracer) Recent(n int) []*RequestTrace {
	if n > len(rt.traces) {
		n = len(rt.traces)
	}
	start := len(rt.traces) - n
	result := make([]*RequestTrace, n)
	copy(result, rt.traces[start:])
	return result
}

// Clear removes all traces.
func (rt *RequestTracer) Clear() {
	rt.traces = rt.traces[:0]
}

// Stats returns aggregate statistics about traced requests.
func (rt *RequestTracer) Stats() TraceStats {
	stats := TraceStats{
		TotalRequests: len(rt.traces),
	}

	if len(rt.traces) == 0 {
		return stats
	}

	var totalDuration time.Duration
	for _, t := range rt.traces {
		totalDuration += t.Duration
		if t.Error != "" {
			stats.Errors++
		}
		stats.TotalRetries += t.Retries
	}

	stats.AvgDuration = totalDuration / time.Duration(len(rt.traces))
	return stats
}

// TraceStats contains aggregate statistics.
type TraceStats struct {
	TotalRequests int           `json:"total_requests"`
	Errors        int           `json:"errors"`
	TotalRetries  int           `json:"total_retries"`
	AvgDuration   time.Duration `json:"avg_duration_ms"`
}
