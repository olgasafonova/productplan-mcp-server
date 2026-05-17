package productplan

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiterConfig configures the adaptive rate limiter.
type RateLimiterConfig struct {
	// SlowdownThreshold is the percentage of remaining requests at which to start slowing down (0.2 = 20%).
	SlowdownThreshold float64

	// MinDelay is the minimum delay between requests when slowing down.
	MinDelay time.Duration

	// MaxDelay is the maximum delay between requests.
	MaxDelay time.Duration

	// DefaultLimit is the assumed rate limit if headers are not present.
	DefaultLimit int

	// ResetBuffer is the extra time to add after a reset window.
	ResetBuffer time.Duration
}

// DefaultRateLimiterConfig returns sensible defaults.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		SlowdownThreshold: 0.2, // Slow down at 20% remaining
		MinDelay:          100 * time.Millisecond,
		MaxDelay:          5 * time.Second,
		DefaultLimit:      100,
		ResetBuffer:       time.Second,
	}
}

// RateLimitState tracks the current rate limit status.
type RateLimitState struct {
	Limit     int       // Total requests allowed
	Remaining int       // Requests remaining
	ResetAt   time.Time // When the window resets
}

// AdaptiveRateLimiter tracks rate limits and slows down proactively.
type AdaptiveRateLimiter struct {
	config RateLimiterConfig
	state  RateLimitState
	mu     sync.RWMutex
}

// NewAdaptiveRateLimiter creates a new rate limiter with the given config.
func NewAdaptiveRateLimiter(config RateLimiterConfig) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		config: config,
		state: RateLimitState{
			Limit:     config.DefaultLimit,
			Remaining: config.DefaultLimit,
		},
	}
}

// parseIntHeader applies fn to the integer value of header, ignoring empty
// or unparseable values. Centralises the "if header non-empty and parses" idiom.
func parseIntHeader(resp *http.Response, header string, fn func(int)) {
	v := resp.Header.Get(header)
	if v == "" {
		return
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return
	}
	fn(n)
}

// parseUnixHeader applies fn to the time.Time parsed from a Unix-timestamp header.
func parseUnixHeader(resp *http.Response, header string, fn func(time.Time)) {
	v := resp.Header.Get(header)
	if v == "" {
		return
	}
	ts, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return
	}
	fn(time.Unix(ts, 0))
}

// UpdateFromResponse updates the rate limit state from response headers.
// It honours both the common X-RateLimit-* family and the IETF RateLimit-* family.
func (r *AdaptiveRateLimiter) UpdateFromResponse(resp *http.Response) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// X-RateLimit-* headers (common format)
	parseIntHeader(resp, "X-RateLimit-Limit", func(n int) { r.state.Limit = n })
	parseIntHeader(resp, "X-RateLimit-Remaining", func(n int) { r.state.Remaining = n })
	parseUnixHeader(resp, "X-RateLimit-Reset", func(t time.Time) { r.state.ResetAt = t })

	// RateLimit-* headers (IETF standard)
	parseIntHeader(resp, "RateLimit-Limit", func(n int) { r.state.Limit = n })
	parseIntHeader(resp, "RateLimit-Remaining", func(n int) { r.state.Remaining = n })
}

// Wait blocks until it's safe to make the next request.
// Returns the delay that was applied.
func (r *AdaptiveRateLimiter) Wait() time.Duration {
	r.mu.RLock()
	state := r.state
	config := r.config
	r.mu.RUnlock()

	// If we're past the reset time, no delay needed
	if !state.ResetAt.IsZero() && time.Now().After(state.ResetAt.Add(config.ResetBuffer)) {
		return 0
	}

	// Calculate remaining percentage
	if state.Limit == 0 {
		return 0
	}

	remainingPct := float64(state.Remaining) / float64(state.Limit)

	// If above threshold, no delay
	if remainingPct > config.SlowdownThreshold {
		return 0
	}

	// Calculate delay based on how close we are to exhaustion
	// As remaining approaches 0, delay approaches MaxDelay
	delayRatio := 1.0 - (remainingPct / config.SlowdownThreshold)
	delay := config.MinDelay + time.Duration(float64(config.MaxDelay-config.MinDelay)*delayRatio)

	// Apply delay
	time.Sleep(delay)
	return delay
}

// ShouldRetry returns true if the request should be retried after a rate limit error.
func (r *AdaptiveRateLimiter) ShouldRetry(resp *http.Response) bool {
	if resp.StatusCode != 429 {
		return false
	}

	// Check if there's a Retry-After header
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			// Only retry if wait time is reasonable (under 60 seconds)
			return seconds <= 60
		}
	}

	return true
}

// GetRetryDelay returns how long to wait before retrying after a 429.
func (r *AdaptiveRateLimiter) GetRetryDelay(resp *http.Response) time.Duration {
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return r.config.MaxDelay
}

// State returns the current rate limit state (for debugging/monitoring).
func (r *AdaptiveRateLimiter) State() RateLimitState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

// RemainingPercent returns the percentage of requests remaining.
func (r *AdaptiveRateLimiter) RemainingPercent() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.state.Limit == 0 {
		return 100.0
	}
	return float64(r.state.Remaining) / float64(r.state.Limit) * 100.0
}
