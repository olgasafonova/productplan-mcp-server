package productplan

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	if config.SlowdownThreshold != 0.2 {
		t.Errorf("expected SlowdownThreshold 0.2, got %f", config.SlowdownThreshold)
	}
	if config.MinDelay != 100*time.Millisecond {
		t.Errorf("expected MinDelay 100ms, got %v", config.MinDelay)
	}
	if config.MaxDelay != 5*time.Second {
		t.Errorf("expected MaxDelay 5s, got %v", config.MaxDelay)
	}
	if config.DefaultLimit != 100 {
		t.Errorf("expected DefaultLimit 100, got %d", config.DefaultLimit)
	}
	if config.ResetBuffer != time.Second {
		t.Errorf("expected ResetBuffer 1s, got %v", config.ResetBuffer)
	}
}

func TestNewAdaptiveRateLimiter(t *testing.T) {
	config := DefaultRateLimiterConfig()
	limiter := NewAdaptiveRateLimiter(config)

	if limiter == nil {
		t.Fatal("expected non-nil limiter")
	}

	state := limiter.State()
	if state.Limit != config.DefaultLimit {
		t.Errorf("expected limit %d, got %d", config.DefaultLimit, state.Limit)
	}
	if state.Remaining != config.DefaultLimit {
		t.Errorf("expected remaining %d, got %d", config.DefaultLimit, state.Remaining)
	}
}

func TestUpdateFromResponse_XRateLimitHeaders(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(DefaultRateLimiterConfig())

	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("X-RateLimit-Limit", "500")
	resp.Header.Set("X-RateLimit-Remaining", "450")
	resp.Header.Set("X-RateLimit-Reset", "1735315200")

	limiter.UpdateFromResponse(resp)

	state := limiter.State()
	if state.Limit != 500 {
		t.Errorf("expected limit 500, got %d", state.Limit)
	}
	if state.Remaining != 450 {
		t.Errorf("expected remaining 450, got %d", state.Remaining)
	}
	if state.ResetAt.Unix() != 1735315200 {
		t.Errorf("expected reset at 1735315200, got %d", state.ResetAt.Unix())
	}
}

func TestUpdateFromResponse_IETFHeaders(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(DefaultRateLimiterConfig())

	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("RateLimit-Limit", "200")
	resp.Header.Set("RateLimit-Remaining", "180")

	limiter.UpdateFromResponse(resp)

	state := limiter.State()
	if state.Limit != 200 {
		t.Errorf("expected limit 200, got %d", state.Limit)
	}
	if state.Remaining != 180 {
		t.Errorf("expected remaining 180, got %d", state.Remaining)
	}
}

func TestUpdateFromResponse_InvalidHeaders(t *testing.T) {
	config := DefaultRateLimiterConfig()
	limiter := NewAdaptiveRateLimiter(config)

	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("X-RateLimit-Limit", "not-a-number")
	resp.Header.Set("X-RateLimit-Remaining", "also-not-a-number")

	limiter.UpdateFromResponse(resp)

	state := limiter.State()
	// Should keep default values when headers are invalid
	if state.Limit != config.DefaultLimit {
		t.Errorf("expected limit %d, got %d", config.DefaultLimit, state.Limit)
	}
}

func TestRemainingPercent(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(DefaultRateLimiterConfig())

	// Initial state should be 100%
	pct := limiter.RemainingPercent()
	if pct != 100.0 {
		t.Errorf("expected 100%%, got %f", pct)
	}

	// Update with half remaining
	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("X-RateLimit-Limit", "100")
	resp.Header.Set("X-RateLimit-Remaining", "50")
	limiter.UpdateFromResponse(resp)

	pct = limiter.RemainingPercent()
	if pct != 50.0 {
		t.Errorf("expected 50%%, got %f", pct)
	}
}

func TestRemainingPercent_ZeroLimit(t *testing.T) {
	config := DefaultRateLimiterConfig()
	config.DefaultLimit = 0
	limiter := NewAdaptiveRateLimiter(config)

	// With zero limit, should return 100%
	pct := limiter.RemainingPercent()
	if pct != 100.0 {
		t.Errorf("expected 100%% for zero limit, got %f", pct)
	}
}

func TestWait_NoDelayAboveThreshold(t *testing.T) {
	config := DefaultRateLimiterConfig()
	config.SlowdownThreshold = 0.2 // 20%
	limiter := NewAdaptiveRateLimiter(config)

	// Set state to 50% remaining (well above 20% threshold)
	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("X-RateLimit-Limit", "100")
	resp.Header.Set("X-RateLimit-Remaining", "50")
	limiter.UpdateFromResponse(resp)

	delay := limiter.Wait()
	if delay != 0 {
		t.Errorf("expected no delay above threshold, got %v", delay)
	}
}

func TestWait_ZeroLimit(t *testing.T) {
	config := DefaultRateLimiterConfig()
	config.DefaultLimit = 0
	limiter := NewAdaptiveRateLimiter(config)

	delay := limiter.Wait()
	if delay != 0 {
		t.Errorf("expected no delay with zero limit, got %v", delay)
	}
}

func TestWait_PastResetTime(t *testing.T) {
	config := DefaultRateLimiterConfig()
	limiter := NewAdaptiveRateLimiter(config)

	// Set state with reset time in the past
	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("X-RateLimit-Limit", "100")
	resp.Header.Set("X-RateLimit-Remaining", "5") // Below threshold
	resp.Header.Set("X-RateLimit-Reset", "1000")
	limiter.UpdateFromResponse(resp)

	delay := limiter.Wait()
	if delay != 0 {
		t.Errorf("expected no delay after reset time, got %v", delay)
	}
}

func TestShouldRetry(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(DefaultRateLimiterConfig())

	tests := []struct {
		name       string
		statusCode int
		retryAfter string
		expected   bool
	}{
		{"non-429", 200, "", false},
		{"429 without header", 429, "", true},
		{"429 with short wait", 429, "30", true},
		{"429 with long wait", 429, "120", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tc.statusCode,
				Header:     http.Header{},
			}
			if tc.retryAfter != "" {
				resp.Header.Set("Retry-After", tc.retryAfter)
			}

			result := limiter.ShouldRetry(resp)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetRetryDelay(t *testing.T) {
	config := DefaultRateLimiterConfig()
	limiter := NewAdaptiveRateLimiter(config)

	tests := []struct {
		name       string
		retryAfter string
		expected   time.Duration
	}{
		{"with header", "10", 10 * time.Second},
		{"without header", "", config.MaxDelay},
		{"invalid header", "not-a-number", config.MaxDelay},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: 429,
				Header:     http.Header{},
			}
			if tc.retryAfter != "" {
				resp.Header.Set("Retry-After", tc.retryAfter)
			}

			delay := limiter.GetRetryDelay(resp)
			if delay != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, delay)
			}
		})
	}
}

func TestState_Concurrent(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(DefaultRateLimiterConfig())

	done := make(chan bool)

	// Multiple goroutines reading state
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = limiter.State()
				_ = limiter.RemainingPercent()
			}
			done <- true
		}()
	}

	// One goroutine writing state
	go func() {
		for j := 0; j < 100; j++ {
			resp := &http.Response{Header: make(http.Header)}
			resp.Header.Set("X-RateLimit-Limit", "100")
			resp.Header.Set("X-RateLimit-Remaining", "50")
			limiter.UpdateFromResponse(resp)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 11; i++ {
		<-done
	}
}
