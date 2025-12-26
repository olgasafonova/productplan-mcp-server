package productplan

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryer_Success(t *testing.T) {
	r := NewRetryer(DefaultRetryConfig())

	callCount := 0
	result, retryResult := r.DoSimple(context.Background(), func() (interface{}, error) {
		callCount++
		return "success", nil
	})

	if result != "success" {
		t.Errorf("Expected 'success', got %v", result)
	}
	if retryResult.Attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", retryResult.Attempts)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestRetryer_RetryOnAPIError(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      0,
	}
	r := NewRetryer(config)

	callCount := 0
	result, retryResult := r.DoSimple(context.Background(), func() (interface{}, error) {
		callCount++
		if callCount < 3 {
			return nil, &APIError{StatusCode: 503, Message: "Service Unavailable"}
		}
		return "success", nil
	})

	if result != "success" {
		t.Errorf("Expected 'success', got %v", result)
	}
	if retryResult.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", retryResult.Attempts)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestRetryer_NoRetryOn404(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      0,
	}
	r := NewRetryer(config)

	callCount := 0
	_, retryResult := r.DoSimple(context.Background(), func() (interface{}, error) {
		callCount++
		return nil, &APIError{StatusCode: 404, Message: "Not Found"}
	})

	if retryResult.Attempts != 1 {
		t.Errorf("Expected 1 attempt (no retry on 404), got %d", retryResult.Attempts)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestRetryer_MaxAttemptsExhausted(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      0,
	}
	r := NewRetryer(config)

	callCount := 0
	_, retryResult := r.DoSimple(context.Background(), func() (interface{}, error) {
		callCount++
		return nil, &APIError{StatusCode: 500, Message: "Server Error"}
	})

	if retryResult.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", retryResult.Attempts)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
	if retryResult.LastError == nil {
		t.Error("Expected error, got nil")
	}
}

func TestRetryer_ContextCancellation(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		Multiplier:  2.0,
		Jitter:      0,
	}
	r := NewRetryer(config)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	callCount := 0
	_, retryResult := r.DoSimple(ctx, func() (interface{}, error) {
		callCount++
		return nil, &APIError{StatusCode: 500, Message: "Server Error"}
	})

	// Should stop early due to context cancellation
	if callCount >= 5 {
		t.Errorf("Expected fewer than 5 calls due to context cancellation, got %d", callCount)
	}
	if !errors.Is(retryResult.LastError, context.DeadlineExceeded) {
		t.Errorf("Expected context deadline error, got %v", retryResult.LastError)
	}
}

func TestRetryer_ExponentialBackoff(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 4,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		Multiplier:  2.0,
		Jitter:      0,
	}
	r := NewRetryer(config)

	// Test delay calculation
	delays := []time.Duration{
		r.calculateDelay(1), // 100ms
		r.calculateDelay(2), // 200ms
		r.calculateDelay(3), // 400ms
		r.calculateDelay(4), // 800ms
	}

	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}

	for i, delay := range delays {
		if delay != expected[i] {
			t.Errorf("Delay %d: expected %v, got %v", i+1, expected[i], delay)
		}
	}
}

func TestRetryer_MaxDelayCapt(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 10,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    500 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      0,
	}
	r := NewRetryer(config)

	// After several iterations, delay should be capped at MaxDelay
	delay := r.calculateDelay(10) // Would be 51200ms without cap
	if delay != 500*time.Millisecond {
		t.Errorf("Expected max delay %v, got %v", 500*time.Millisecond, delay)
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"connection refused", errors.New("connection refused"), true},
		{"timeout", errors.New("request timeout"), true},
		{"EOF", errors.New("unexpected EOF"), true},
		{"regular error", errors.New("some other error"), false},
		{"api error", &APIError{StatusCode: 500}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNetworkError(tt.err); got != tt.expected {
				t.Errorf("isNetworkError() = %v, want %v", got, tt.expected)
			}
		})
	}
}
