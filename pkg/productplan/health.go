package productplan

import (
	"context"
	"encoding/json"
	"time"
)

// HealthStatus represents the overall health status.
type HealthStatus string

const (
	HealthOK       HealthStatus = "ok"
	HealthDegraded HealthStatus = "degraded"
	HealthDown     HealthStatus = "down"
)

// ComponentHealth represents the health of a single component.
type ComponentHealth struct {
	Name    string       `json:"name"`
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Latency int64        `json:"latency_ms,omitempty"`
}

// HealthReport contains the health status of all components.
type HealthReport struct {
	Status       HealthStatus      `json:"status"`
	Version      string            `json:"version"`
	Timestamp    time.Time         `json:"timestamp"`
	Components   []ComponentHealth `json:"components"`
	RateLimit    *RateLimitHealth  `json:"rate_limit,omitempty"`
	ResponseTime int64             `json:"response_time_ms"`
}

// RateLimitHealth reports rate limiting status.
type RateLimitHealth struct {
	Limit     int     `json:"limit"`
	Remaining int     `json:"remaining"`
	Percent   float64 `json:"remaining_percent"`
}

// HealthChecker performs health checks.
type HealthChecker struct {
	version     string
	rateLimiter *AdaptiveRateLimiter
	apiChecker  func(ctx context.Context) (int64, error) // Returns latency in ms
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(version string, rateLimiter *AdaptiveRateLimiter) *HealthChecker {
	return &HealthChecker{
		version:     version,
		rateLimiter: rateLimiter,
	}
}

// SetAPIChecker sets the function to check API health.
func (h *HealthChecker) SetAPIChecker(checker func(ctx context.Context) (int64, error)) {
	h.apiChecker = checker
}

// Check performs a health check.
// If deep is true, it will also check the API connectivity.
func (h *HealthChecker) Check(ctx context.Context, deep bool) HealthReport {
	start := time.Now()

	report := HealthReport{
		Status:     HealthOK,
		Version:    h.version,
		Timestamp:  time.Now(),
		Components: make([]ComponentHealth, 0),
	}

	h.checkRateLimiter(&report)
	if deep {
		h.checkAPI(ctx, &report)
	}

	report.ResponseTime = time.Since(start).Milliseconds()
	return report
}

// checkRateLimiter adds rate limiter health to the report.
func (h *HealthChecker) checkRateLimiter(report *HealthReport) {
	if h.rateLimiter == nil {
		return
	}
	state := h.rateLimiter.State()
	remaining := h.rateLimiter.RemainingPercent()

	report.RateLimit = &RateLimitHealth{
		Limit:     state.Limit,
		Remaining: state.Remaining,
		Percent:   remaining,
	}

	status, message := rateLimiterStatus(remaining)
	if status == HealthDegraded && report.Status == HealthOK {
		report.Status = HealthDegraded
	}

	report.Components = append(report.Components, ComponentHealth{
		Name:    "rate_limiter",
		Status:  status,
		Message: message,
	})
}

// rateLimiterStatus returns the status and message for a given remaining percentage.
func rateLimiterStatus(remaining float64) (HealthStatus, string) {
	if remaining < 10 {
		return HealthDegraded, "Rate limit nearly exhausted"
	}
	return HealthOK, "Rate limits healthy"
}

// checkAPI adds API connectivity health to the report.
func (h *HealthChecker) checkAPI(ctx context.Context, report *HealthReport) {
	if h.apiChecker == nil {
		return
	}
	latency, err := h.apiChecker(ctx)
	if err != nil {
		report.Components = append(report.Components, ComponentHealth{
			Name:    "api",
			Status:  HealthDown,
			Message: err.Error(),
			Latency: latency,
		})
		report.Status = HealthDown
		return
	}

	status, message := apiStatus(latency)
	if status == HealthDegraded && report.Status == HealthOK {
		report.Status = HealthDegraded
	}

	report.Components = append(report.Components, ComponentHealth{
		Name:    "api",
		Status:  status,
		Message: message,
		Latency: latency,
	})
}

// apiStatus returns the status and message based on API latency in ms.
func apiStatus(latency int64) (HealthStatus, string) {
	if latency > 5000 {
		return HealthDegraded, "API response slow"
	}
	return HealthOK, "API responding"
}

// ToJSON converts the health report to JSON.
func (r *HealthReport) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
