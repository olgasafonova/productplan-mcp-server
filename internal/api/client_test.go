package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty token",
			cfg:     Config{Token: ""},
			wantErr: true,
			errMsg:  "API token is required",
		},
		{
			name:    "valid token",
			cfg:     Config{Token: "test-token"},
			wantErr: false,
		},
		{
			name: "custom base URL",
			cfg: Config{
				Token:   "test-token",
				BaseURL: "https://custom.example.com/api",
			},
			wantErr: false,
		},
		{
			name: "invalid base URL",
			cfg: Config{
				Token:   "test-token",
				BaseURL: "://invalid",
			},
			wantErr: true,
			errMsg:  "invalid base URL",
		},
		{
			name: "with custom timeout",
			cfg: Config{
				Token:   "test-token",
				Timeout: 60 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "with logger",
			cfg: Config{
				Token:  "test-token",
				Logger: logging.Nop(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("expected non-nil client")
			}
		})
	}
}

func TestNewSimple(t *testing.T) {
	client, err := NewSimple("test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.token != "test-token" {
		t.Errorf("expected token 'test-token', got %q", client.token)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("my-token")
	if cfg.Token != "my-token" {
		t.Errorf("expected token 'my-token', got %q", cfg.Token)
	}
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("expected BaseURL %q, got %q", DefaultBaseURL, cfg.BaseURL)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("expected Timeout %v, got %v", DefaultTimeout, cfg.Timeout)
	}
}

func TestClientRequest(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		endpoint   string
		body       any
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name:       "successful GET",
			method:     http.MethodGet,
			endpoint:   "/test",
			statusCode: 200,
			response:   `{"id": 1, "name": "test"}`,
			wantErr:    false,
		},
		{
			name:       "successful POST",
			method:     http.MethodPost,
			endpoint:   "/create",
			body:       map[string]string{"name": "new item"},
			statusCode: 201,
			response:   `{"id": 2, "name": "new item"}`,
			wantErr:    false,
		},
		{
			name:       "204 no content",
			method:     http.MethodDelete,
			endpoint:   "/delete/1",
			statusCode: 204,
			response:   "",
			wantErr:    false,
		},
		{
			name:       "404 not found",
			method:     http.MethodGet,
			endpoint:   "/notfound",
			statusCode: 404,
			response:   `{"error": "not found"}`,
			wantErr:    true,
		},
		{
			name:       "500 server error",
			method:     http.MethodGet,
			endpoint:   "/error",
			statusCode: 500,
			response:   `{"error": "internal error"}`,
			wantErr:    true,
		},
		{
			name:       "401 unauthorized",
			method:     http.MethodGet,
			endpoint:   "/protected",
			statusCode: 401,
			response:   `{"error": "unauthorized"}`,
			wantErr:    true,
		},
		{
			name:       "429 rate limited",
			method:     http.MethodGet,
			endpoint:   "/ratelimited",
			statusCode: 429,
			response:   `{"error": "too many requests"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify headers
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-token" {
					t.Errorf("expected Bearer token, got %q", auth)
				}
				ct := r.Header.Get("Content-Type")
				if ct != "application/json" {
					t.Errorf("expected application/json, got %q", ct)
				}

				// Verify method
				if r.Method != tt.method {
					t.Errorf("expected method %s, got %s", tt.method, r.Method)
				}

				w.WriteHeader(tt.statusCode)
				if tt.response != "" {
					w.Write([]byte(tt.response))
				}
			}))
			defer server.Close()

			client, err := New(Config{
				Token:   "test-token",
				BaseURL: server.URL,
				Logger:  logging.Nop(),
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			result, err := client.Request(context.Background(), tt.method, tt.endpoint, tt.body)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.statusCode == 204 {
				// Check for success response
				var resp map[string]any
				if err := json.Unmarshal(result, &resp); err != nil {
					t.Errorf("failed to unmarshal 204 response: %v", err)
				}
				if resp["success"] != true {
					t.Error("expected success: true for 204 response")
				}
			} else if string(result) != tt.response {
				t.Errorf("expected response %q, got %q", tt.response, string(result))
			}
		})
	}
}

func TestClientHTTPMethods(t *testing.T) {
	var receivedMethod string
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		if r.Body != nil {
			receivedBody, _ = json.Marshal(r.Body)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	client, _ := New(Config{
		Token:   "test-token",
		BaseURL: server.URL,
		Logger:  logging.Nop(),
	})

	ctx := context.Background()

	t.Run("Get", func(t *testing.T) {
		_, err := client.Get(ctx, "/test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodGet {
			t.Errorf("expected GET, got %s", receivedMethod)
		}
	})

	t.Run("Post", func(t *testing.T) {
		_, err := client.Post(ctx, "/test", map[string]string{"key": "value"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodPost {
			t.Errorf("expected POST, got %s", receivedMethod)
		}
	})

	t.Run("Patch", func(t *testing.T) {
		_, err := client.Patch(ctx, "/test", map[string]string{"key": "updated"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", receivedMethod)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		_, err := client.Delete(ctx, "/test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", receivedMethod)
		}
	})

	// Suppress unused variable warning
	_ = receivedBody
}

func TestClientCache(t *testing.T) {
	client, _ := NewSimple("test-token")
	cache := client.Cache()
	if cache == nil {
		t.Error("expected non-nil cache")
	}
}

func TestClientRateLimiter(t *testing.T) {
	client, _ := NewSimple("test-token")
	limiter := client.RateLimiter()
	if limiter == nil {
		t.Error("expected non-nil rate limiter")
	}
}

func TestClientSetLogger(t *testing.T) {
	client, _ := NewSimple("test-token")
	logger := logging.Nop()
	client.SetLogger(logger)
	// No assertion needed; just verify it doesn't panic
}

func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer server.Close()

	client, _ := New(Config{
		Token:   "test-token",
		BaseURL: server.URL,
		Logger:  logging.Nop(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, "/test")
	if err == nil {
		t.Error("expected error due to cancelled context")
	}
}

func TestClientRequestBodyMarshalError(t *testing.T) {
	client, _ := NewSimple("test-token")

	// Create an unmarshalable value (channel)
	ch := make(chan int)

	_, err := client.Post(context.Background(), "/test", ch)
	if err == nil {
		t.Error("expected marshal error")
	}
	if !strings.Contains(err.Error(), "marshal") {
		t.Errorf("expected marshal error, got: %v", err)
	}
}

func BenchmarkClientRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client, _ := New(Config{
		Token:   "test-token",
		BaseURL: server.URL,
		Logger:  logging.Nop(),
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get(ctx, "/test")
	}
}
