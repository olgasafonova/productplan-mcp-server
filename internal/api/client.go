// Package api provides the ProductPlan API client.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
	"github.com/olgasafonova/productplan-mcp-server/pkg/productplan"
)

// safeSeg validates an ID arg as a URL-safe path segment and returns the
// PathEscape-d form ready for interpolation. Used by every endpoint method
// that interpolates user-supplied IDs into a path; without it, an
// adversarial caller could send `bar_id="../../strategy/objectives/SECRET"`
// and pivot a manage_bar action to a different resource.
//
// PathEscape on a validator-approved ID is a no-op today (the regex restricts
// to URL-safe chars), but it stays as belt-and-braces against future regex
// loosening.
func safeSeg(field, value string) (string, error) {
	if err := productplan.RequireID(field, value); err != nil {
		return "", err
	}
	return url.PathEscape(strings.TrimSpace(value)), nil
}

// safeSegPair is a convenience for the recurring "validate two IDs, then
// interpolate" pattern used by every update/delete of a sub-resource. It
// returns the two escaped segments or the first validation error encountered.
// The order of fields in the call site is the order returned, which keeps the
// URL composition local and obvious at the call site.
func safeSegPair(field1, value1, field2, value2 string) (string, string, error) {
	seg1, err := safeSeg(field1, value1)
	if err != nil {
		return "", "", err
	}
	seg2, err := safeSeg(field2, value2)
	if err != nil {
		return "", "", err
	}
	return seg1, seg2, nil
}

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
	Logger  logging.Logger
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
	cache       *productplan.Cache
	logger      logging.Logger
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
		baseURL:     baseURL,
		token:       cfg.Token,
		httpClient:  &http.Client{Timeout: timeout},
		rateLimiter: productplan.NewAdaptiveRateLimiter(productplan.DefaultRateLimiterConfig()),
		cache:       productplan.NewCache(productplan.DefaultCacheConfig()),
		logger:      logger,
	}, nil
}

// NewSimple creates a client with just a token (uses defaults).
func NewSimple(token string) (*Client, error) {
	return New(DefaultConfig(token))
}

// buildRequest constructs an HTTP request with auth and content-type headers attached.
func (c *Client) buildRequest(ctx context.Context, method, endpoint string, body any) (*http.Request, error) {
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
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
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
			return nil, fmt.Errorf("%s. %s", apiErr.Error(), suggestion)
		}
		return nil, apiErr
	}
	if resp.StatusCode == 204 {
		return json.RawMessage(`{"success": true}`), nil
	}
	return respBody, nil
}

// Request performs an HTTP request to the API.
func (c *Client) Request(ctx context.Context, method, endpoint string, body any) (json.RawMessage, error) {
	start := time.Now()

	if c.rateLimiter != nil {
		c.rateLimiter.Wait()
	}

	req, err := c.buildRequest(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("API request",
		logging.Endpoint(endpoint),
		logging.F("method", method),
	)

	resp, err := c.httpClient.Do(req) // #nosec G704 -- URL is the configured ProductPlan API endpoint, not user-controlled
	if err != nil {
		c.logger.Error("API request failed",
			logging.Endpoint(endpoint),
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
		logging.Endpoint(endpoint),
		logging.StatusCode(resp.StatusCode),
		logging.Duration(time.Since(start)),
	)

	return handleResponse(resp, respBody)
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, endpoint string) (json.RawMessage, error) {
	return c.Request(ctx, http.MethodGet, endpoint, nil)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, endpoint string, body any) (json.RawMessage, error) {
	return c.Request(ctx, http.MethodPost, endpoint, body)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, endpoint string, body any) (json.RawMessage, error) {
	return c.Request(ctx, http.MethodPatch, endpoint, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, endpoint string) (json.RawMessage, error) {
	return c.Request(ctx, http.MethodDelete, endpoint, nil)
}

// Cache returns the client's cache for external use.
func (c *Client) Cache() *productplan.Cache {
	return c.cache
}

// RateLimiter returns the client's rate limiter for external use.
func (c *Client) RateLimiter() *productplan.AdaptiveRateLimiter {
	return c.rateLimiter
}

// SetLogger sets the logger for the client.
func (c *Client) SetLogger(logger logging.Logger) {
	c.logger = logger
}
