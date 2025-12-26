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
	CacheStats   *CacheStats       `json:"cache,omitempty"`
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
	cache       *Cache
	apiChecker  func(ctx context.Context) (int64, error) // Returns latency in ms
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(version string, rateLimiter *AdaptiveRateLimiter, cache *Cache) *HealthChecker {
	return &HealthChecker{
		version:     version,
		rateLimiter: rateLimiter,
		cache:       cache,
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

	// Check rate limiter status
	if h.rateLimiter != nil {
		state := h.rateLimiter.State()
		remaining := h.rateLimiter.RemainingPercent()

		report.RateLimit = &RateLimitHealth{
			Limit:     state.Limit,
			Remaining: state.Remaining,
			Percent:   remaining,
		}

		status := HealthOK
		message := "Rate limits healthy"

		if remaining < 10 {
			status = HealthDegraded
			message = "Rate limit nearly exhausted"
			if report.Status == HealthOK {
				report.Status = HealthDegraded
			}
		}

		report.Components = append(report.Components, ComponentHealth{
			Name:    "rate_limiter",
			Status:  status,
			Message: message,
		})
	}

	// Check cache status
	if h.cache != nil {
		stats := h.cache.Stats()
		report.CacheStats = &stats

		report.Components = append(report.Components, ComponentHealth{
			Name:    "cache",
			Status:  HealthOK,
			Message: "Cache operational",
		})
	}

	// Deep check: verify API connectivity
	if deep && h.apiChecker != nil {
		latency, err := h.apiChecker(ctx)

		if err != nil {
			report.Components = append(report.Components, ComponentHealth{
				Name:    "api",
				Status:  HealthDown,
				Message: err.Error(),
				Latency: latency,
			})
			report.Status = HealthDown
		} else {
			status := HealthOK
			message := "API responding"

			if latency > 5000 {
				status = HealthDegraded
				message = "API response slow"
				if report.Status == HealthOK {
					report.Status = HealthDegraded
				}
			}

			report.Components = append(report.Components, ComponentHealth{
				Name:    "api",
				Status:  status,
				Message: message,
				Latency: latency,
			})
		}
	}

	report.ResponseTime = time.Since(start).Milliseconds()
	return report
}

// ToJSON converts the health report to JSON.
func (r *HealthReport) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
