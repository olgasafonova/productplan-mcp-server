// Package api provides the ProductPlan API client.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
	"github.com/olgasafonova/productplan-mcp-server/pkg/productplan"
)

const (
	// DefaultBaseURL is the ProductPlan API base URL.
	DefaultBaseURL = "https://app.productplan.com/api/v2"

	// DefaultTimeout for HTTP requests.
	DefaultTimeout = 30 * time.Second
)

// Config holds API client configuration.
type Config struct {
	BaseURL string
	Token   string
	Timeout time.Duration
	Logger  *slog.Logger
	// CacheTTL enables the in-process read cache for GETs when positive.
	// Zero (the zero value) disables it; the server sets it from
	// CacheTTLFromEnv.
	CacheTTL time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(token string) Config {
	return Config{
		BaseURL: DefaultBaseURL,
		Token:   token,
		Timeout: DefaultTimeout,
		Logger:  logging.Nop(),
	}
}

// Client is the ProductPlan API client.
type Client struct {
	baseURL     string
	token       string
	httpClient  *http.Client
	rateLimiter *productplan.AdaptiveRateLimiter
	logger      *slog.Logger
	cache       *readCache // nil when disabled
}

// New creates a new API client with the given configuration.
func New(cfg Config) (*Client, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("API token is required")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	logger := cfg.Logger
	if logger == nil {
		logger = logging.Nop()
	}

	return &Client{
		baseURL: baseURL,
		token:   cfg.Token,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: newTransport(),
			// SECURITY: Refuse all redirects. The configured BaseURL
			// (app.productplan.com/api/v2 by default) is the only legitimate
			// target. Without CheckRedirect, Go follows up to 10 3xx responses;
			// a misconfigured deployment combined with an upstream returning
			// Location: http://169.254.169.254/... would pivot the request to
			// internal IPs (cloud metadata, link-local). HG-4 graduated rule
			// (rules/code-review-prompts.md): every HTTP client must set
			// CheckRedirect. Full-refuse is appropriate here because the
			// ProductPlan API has exactly one legitimate upstream.
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		rateLimiter: productplan.NewAdaptiveRateLimiter(productplan.DefaultRateLimiterConfig()),
		logger:      logger,
		cache:       newReadCache(cfg.CacheTTL),
	}, nil
}

// NewSimple creates a client with just a token (uses defaults).
func NewSimple(token string) (*Client, error) {
	return New(DefaultConfig(token))
}

// buildRequest constructs an HTTP request with auth and content-type headers attached.
func (c *Client) buildRequest(ctx context.Context, v verb, path apiPath, body any) (*http.Request, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Build URL by concatenating base URL with endpoint path.
	// ResolveReference strips the base path when endpoint starts with "/",
	// so we use simple string concatenation instead.
	req, err := http.NewRequestWithContext(ctx, string(v), c.baseURL+string(path), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// handleResponse converts an HTTP response body and status into the API contract.
func handleResponse(resp *http.Response, respBody []byte) (json.RawMessage, error) {
	if resp.StatusCode >= 400 {
		apiErr := productplan.ParseAPIError(resp, respBody)
		if suggestion := apiErr.Suggestion(); suggestion != "" {
			// %w keeps the *APIError reachable via errors.As, so callers
			// (IsNotFound, IsRetryable checks) never parse message text.
			return nil, fmt.Errorf("%w. %s", apiErr, suggestion)
		}
		return nil, apiErr
	}
	if resp.StatusCode == 204 {
		return json.RawMessage(`{"success": true}`), nil
	}
	return respBody, nil
}

// request performs an HTTP request to the API. Any verb other than GET or
// HEAD clears the read cache once the request returns, whether or not it
// succeeded: a failed or timed-out write may still have been applied
// upstream, so correct beats clever.
//
// request, get and getList are unexported on purpose: outside this package
// the API is reachable only through the typed endpoint methods, whose paths
// are built from validated IDs (ids.go, path.go).
func (c *Client) request(ctx context.Context, v verb, path apiPath, body any) (json.RawMessage, error) {
	if !v.readOnly() {
		defer c.cache.invalidate()
	}
	start := time.Now()

	if c.rateLimiter != nil {
		c.rateLimiter.Wait()
	}

	req, err := c.buildRequest(ctx, v, path, body)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("API request",
		path.attr(),
		slog.String("method", string(v)),
	)

	resp, err := c.httpClient.Do(req) // #nosec G704 -- URL is the configured ProductPlan API endpoint, not user-controlled
	if err != nil {
		c.logger.Error("API request failed",
			path.attr(),
			logging.Error(err),
			logging.Duration(time.Since(start)),
		)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if c.rateLimiter != nil {
		c.rateLimiter.UpdateFromResponse(resp)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Debug("API response",
		path.attr(),
		slog.Int("status_code", resp.StatusCode),
		logging.Duration(time.Since(start)),
	)

	return handleResponse(resp, respBody)
}

// get performs a GET request, served from the read cache when enabled.
// path (including its query) is the cache key.
func (c *Client) get(ctx context.Context, path apiPath) (json.RawMessage, error) {
	return c.cache.get(ctx, string(path), func(ctx context.Context) (json.RawMessage, error) {
		return c.request(ctx, http.MethodGet, path, nil)
	})
}

// CacheStats reports read-cache counters (zero value with Enabled=false
// when the cache is off). In-process only; touches no network.
func (c *Client) CacheStats() CacheStats {
	return c.cache.stats()
}

// RateLimiter returns the client's rate limiter for external use.
func (c *Client) RateLimiter() *productplan.AdaptiveRateLimiter {
	return c.rateLimiter
}

// SetLogger sets the logger for the client.
func (c *Client) SetLogger(logger *slog.Logger) {
	c.logger = logger
}
