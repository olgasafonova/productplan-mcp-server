package productplan

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxAttempts int           // Maximum number of attempts (including initial)
	BaseDelay   time.Duration // Initial delay between retries
	MaxDelay    time.Duration // Maximum delay between retries
	Multiplier  float64       // Multiplier for exponential backoff
	Jitter      float64       // Random jitter factor (0-1)
}

// DefaultRetryConfig returns sensible defaults for API retries.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
		Jitter:      0.1,
	}
}

// Retryer handles retry logic with exponential backoff.
type Retryer struct {
	config RetryConfig
}

// NewRetryer creates a new retryer with the given config.
func NewRetryer(config RetryConfig) *Retryer {
	return &Retryer{config: config}
}

// RetryResult contains the result of a retried operation.
type RetryResult struct {
	Attempts   int           // Number of attempts made
	TotalDelay time.Duration // Total time spent waiting between retries
	LastError  error         // Last error encountered (nil if successful)
}

// Do executes the given function with retries.
// The function should return (result, error, shouldRetry).
// If shouldRetry is false, no retry will be attempted regardless of error.
func (r *Retryer) Do(ctx context.Context, fn func() (interface{}, error, bool)) (interface{}, RetryResult) {
	result := RetryResult{}

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		result.Attempts = attempt

		// Check context before attempting
		if ctx.Err() != nil {
			result.LastError = ctx.Err()
			return nil, result
		}

		// Execute the function
		res, err, shouldRetry := fn()
		if err == nil {
			return res, result
		}

		result.LastError = err

		// Don't retry if not retryable or last attempt
		if !shouldRetry || attempt == r.config.MaxAttempts {
			return nil, result
		}

		// Calculate delay with exponential backoff
		delay := r.calculateDelay(attempt)
		result.TotalDelay += delay

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			result.LastError = ctx.Err()
			return nil, result
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return nil, result
}

// DoSimple executes with automatic retry detection using APIError.
func (r *Retryer) DoSimple(ctx context.Context, fn func() (interface{}, error)) (interface{}, RetryResult) {
	return r.Do(ctx, func() (interface{}, error, bool) {
		res, err := fn()
		if err == nil {
			return res, nil, false
		}

		// Check if error is retryable
		if apiErr, ok := err.(*APIError); ok {
			return nil, err, apiErr.IsRetryable()
		}

		// Network errors are generally retryable
		return nil, err, isNetworkError(err)
	})
}

// calculateDelay computes the delay for a given attempt with jitter.
func (r *Retryer) calculateDelay(attempt int) time.Duration {
	// Exponential backoff: baseDelay * multiplier^(attempt-1)
	delay := float64(r.config.BaseDelay) * math.Pow(r.config.Multiplier, float64(attempt-1))

	// Apply jitter
	if r.config.Jitter > 0 {
		jitter := delay * r.config.Jitter * (rand.Float64()*2 - 1) // +/- jitter%
		delay += jitter
	}

	// Cap at max delay
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	return time.Duration(delay)
}

// isNetworkError checks if an error is likely a network-related error.
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common network error patterns
	errStr := err.Error()
	networkPatterns := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"timeout",
		"temporary failure",
		"EOF",
	}
	for _, pattern := range networkPatterns {
		if contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// contains checks if s contains substr (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchIgnoreCase(s, substr)
}

func searchIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldAt(s, i, substr) {
			return true
		}
	}
	return false
}

func equalFoldAt(s string, start int, substr string) bool {
	for j := 0; j < len(substr); j++ {
		c1 := s[start+j]
		c2 := substr[j]
		if c1 != c2 && toLower(c1) != toLower(c2) {
			return false
		}
	}
	return true
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}
