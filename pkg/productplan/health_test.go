package productplan

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewHealthChecker(t *testing.T) {
	checker := NewHealthChecker("1.0.0", nil, nil)

	if checker == nil {
		t.Fatal("expected non-nil health checker")
	}

	if checker.version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", checker.version)
	}
}

func TestNewHealthCheckerWithComponents(t *testing.T) {
	rateLimiter := NewAdaptiveRateLimiter(DefaultRateLimiterConfig())
	cache := NewCache(CacheConfig{MaxEntries: 1000})

	checker := NewHealthChecker("2.0.0", rateLimiter, cache)

	if checker.rateLimiter == nil {
		t.Error("expected rate limiter to be set")
	}
	if checker.cache == nil {
		t.Error("expected cache to be set")
	}
}

func TestSetAPIChecker(t *testing.T) {
	checker := NewHealthChecker("1.0.0", nil, nil)

	called := false
	checker.SetAPIChecker(func(ctx context.Context) (int64, error) {
		called = true
		return 100, nil
	})

	if checker.apiChecker == nil {
		t.Fatal("expected apiChecker to be set")
	}

	// Verify the function works
	_, _ = checker.apiChecker(context.Background())
	if !called {
		t.Error("expected apiChecker function to be callable")
	}
}

func TestCheck_Basic(t *testing.T) {
	checker := NewHealthChecker("1.0.0", nil, nil)

	report := checker.Check(context.Background(), false)

	if report.Status != HealthOK {
		t.Errorf("expected status OK, got %s", report.Status)
	}
	if report.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", report.Version)
	}
	if report.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
	if len(report.Components) != 0 {
		t.Errorf("expected 0 components, got %d", len(report.Components))
	}
}

func TestCheck_WithRateLimiter_Healthy(t *testing.T) {
	rateLimiter := NewAdaptiveRateLimiter(DefaultRateLimiterConfig())
	checker := NewHealthChecker("1.0.0", rateLimiter, nil)

	report := checker.Check(context.Background(), false)

	if report.Status != HealthOK {
		t.Errorf("expected status OK, got %s", report.Status)
	}
	if report.RateLimit == nil {
		t.Fatal("expected rate limit health to be set")
	}
	if report.RateLimit.Percent != 100.0 {
		t.Errorf("expected 100%% remaining, got %f", report.RateLimit.Percent)
	}

	// Check component
	if len(report.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(report.Components))
	}
	if report.Components[0].Name != "rate_limiter" {
		t.Errorf("expected rate_limiter component, got %s", report.Components[0].Name)
	}
	if report.Components[0].Status != HealthOK {
		t.Errorf("expected OK status, got %s", report.Components[0].Status)
	}
}

func TestCheck_WithRateLimiter_Degraded(t *testing.T) {
	config := DefaultRateLimiterConfig()
	rateLimiter := NewAdaptiveRateLimiter(config)

	// Simulate nearly exhausted rate limit (5% remaining)
	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("X-RateLimit-Limit", "100")
	resp.Header.Set("X-RateLimit-Remaining", "5")
	rateLimiter.UpdateFromResponse(resp)

	checker := NewHealthChecker("1.0.0", rateLimiter, nil)
	report := checker.Check(context.Background(), false)

	if report.Status != HealthDegraded {
		t.Errorf("expected status Degraded, got %s", report.Status)
	}
	if report.RateLimit.Percent != 5.0 {
		t.Errorf("expected 5%% remaining, got %f", report.RateLimit.Percent)
	}

	// Check component status
	if report.Components[0].Status != HealthDegraded {
		t.Errorf("expected component status Degraded, got %s", report.Components[0].Status)
	}
}

func TestCheck_WithCache(t *testing.T) {
	cache := NewCache(CacheConfig{MaxEntries: 1000})
	checker := NewHealthChecker("1.0.0", nil, cache)

	report := checker.Check(context.Background(), false)

	if report.CacheStats == nil {
		t.Fatal("expected cache stats to be set")
	}
	if len(report.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(report.Components))
	}
	if report.Components[0].Name != "cache" {
		t.Errorf("expected cache component, got %s", report.Components[0].Name)
	}
	if report.Components[0].Status != HealthOK {
		t.Errorf("expected OK status, got %s", report.Components[0].Status)
	}
}

func TestCheck_Deep_APIHealthy(t *testing.T) {
	checker := NewHealthChecker("1.0.0", nil, nil)
	checker.SetAPIChecker(func(ctx context.Context) (int64, error) {
		return 150, nil // 150ms latency
	})

	report := checker.Check(context.Background(), true)

	if report.Status != HealthOK {
		t.Errorf("expected status OK, got %s", report.Status)
	}
	if len(report.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(report.Components))
	}
	if report.Components[0].Name != "api" {
		t.Errorf("expected api component, got %s", report.Components[0].Name)
	}
	if report.Components[0].Latency != 150 {
		t.Errorf("expected latency 150, got %d", report.Components[0].Latency)
	}
}

func TestCheck_Deep_APISlow(t *testing.T) {
	checker := NewHealthChecker("1.0.0", nil, nil)
	checker.SetAPIChecker(func(ctx context.Context) (int64, error) {
		return 6000, nil // 6 seconds - over 5000ms threshold
	})

	report := checker.Check(context.Background(), true)

	if report.Status != HealthDegraded {
		t.Errorf("expected status Degraded, got %s", report.Status)
	}
	if report.Components[0].Status != HealthDegraded {
		t.Errorf("expected component status Degraded, got %s", report.Components[0].Status)
	}
	if report.Components[0].Message != "API response slow" {
		t.Errorf("unexpected message: %s", report.Components[0].Message)
	}
}

func TestCheck_Deep_APIError(t *testing.T) {
	checker := NewHealthChecker("1.0.0", nil, nil)
	checker.SetAPIChecker(func(ctx context.Context) (int64, error) {
		return 0, errors.New("connection refused")
	})

	report := checker.Check(context.Background(), true)

	if report.Status != HealthDown {
		t.Errorf("expected status Down, got %s", report.Status)
	}
	if report.Components[0].Status != HealthDown {
		t.Errorf("expected component status Down, got %s", report.Components[0].Status)
	}
	if report.Components[0].Message != "connection refused" {
		t.Errorf("expected error message, got: %s", report.Components[0].Message)
	}
}

func TestCheck_Deep_WithoutChecker(t *testing.T) {
	checker := NewHealthChecker("1.0.0", nil, nil)
	// No API checker set

	report := checker.Check(context.Background(), true)

	// Should still work, just no API component
	if report.Status != HealthOK {
		t.Errorf("expected status OK, got %s", report.Status)
	}
	if len(report.Components) != 0 {
		t.Errorf("expected 0 components, got %d", len(report.Components))
	}
}

func TestCheck_AllComponents(t *testing.T) {
	rateLimiter := NewAdaptiveRateLimiter(DefaultRateLimiterConfig())
	cache := NewCache(CacheConfig{MaxEntries: 1000})
	checker := NewHealthChecker("1.0.0", rateLimiter, cache)
	checker.SetAPIChecker(func(ctx context.Context) (int64, error) {
		return 100, nil
	})

	report := checker.Check(context.Background(), true)

	if report.Status != HealthOK {
		t.Errorf("expected status OK, got %s", report.Status)
	}
	if len(report.Components) != 3 {
		t.Errorf("expected 3 components, got %d", len(report.Components))
	}

	// Verify all components present
	names := make(map[string]bool)
	for _, c := range report.Components {
		names[c.Name] = true
	}

	expected := []string{"rate_limiter", "cache", "api"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected component %s not found", name)
		}
	}
}

func TestCheck_ResponseTime(t *testing.T) {
	checker := NewHealthChecker("1.0.0", nil, nil)

	report := checker.Check(context.Background(), false)

	// Response time should be >= 0
	if report.ResponseTime < 0 {
		t.Errorf("expected non-negative response time, got %d", report.ResponseTime)
	}
}

func TestHealthReport_ToJSON(t *testing.T) {
	report := HealthReport{
		Status:    HealthOK,
		Version:   "1.0.0",
		Timestamp: time.Now(),
		Components: []ComponentHealth{
			{Name: "test", Status: HealthOK, Message: "test message"},
		},
		ResponseTime: 50,
	}

	data, err := report.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check fields
	if parsed["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", parsed["status"])
	}
	if parsed["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", parsed["version"])
	}
}

func TestHealthReport_ToJSON_WithRateLimit(t *testing.T) {
	report := HealthReport{
		Status:    HealthOK,
		Version:   "1.0.0",
		Timestamp: time.Now(),
		RateLimit: &RateLimitHealth{
			Limit:     100,
			Remaining: 50,
			Percent:   50.0,
		},
	}

	data, err := report.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	rateLimit, ok := parsed["rate_limit"].(map[string]any)
	if !ok {
		t.Fatal("expected rate_limit object in JSON")
	}
	if rateLimit["limit"].(float64) != 100 {
		t.Errorf("expected limit 100, got %v", rateLimit["limit"])
	}
}

