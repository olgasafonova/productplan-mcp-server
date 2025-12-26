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
	baseURL     *url.URL
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

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
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
		baseURL:     parsed,
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

// Request performs an HTTP request to the API.
func (c *Client) Request(ctx context.Context, method, endpoint string, body any) (json.RawMessage, error) {
	start := time.Now()

	// Wait if rate limited
	if c.rateLimiter != nil {
		c.rateLimiter.Wait()
	}

	// Build URL
	reqURL := c.baseURL.ResolveReference(&url.URL{Path: endpoint})

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	c.logger.Debug("API request",
		logging.Endpoint(endpoint),
		logging.F("method", method),
	)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("API request failed",
			logging.Endpoint(endpoint),
			logging.Error(err),
			logging.Duration(time.Since(start)),
		)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Update rate limiter from response headers
	if c.rateLimiter != nil {
		c.rateLimiter.UpdateFromResponse(resp)
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Debug("API response",
		logging.Endpoint(endpoint),
		logging.StatusCode(resp.StatusCode),
		logging.Duration(time.Since(start)),
	)

	// Handle errors
	if resp.StatusCode >= 400 {
		apiErr := productplan.ParseAPIError(resp, respBody)
		if suggestion := apiErr.Suggestion(); suggestion != "" {
			return nil, fmt.Errorf("%s. %s", apiErr.Error(), suggestion)
		}
		return nil, apiErr
	}

	// Handle 204 No Content
	if resp.StatusCode == 204 {
		return json.RawMessage(`{"success": true}`), nil
	}

	return respBody, nil
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
