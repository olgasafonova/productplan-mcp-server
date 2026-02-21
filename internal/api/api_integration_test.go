//go:build integration

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

// TestAPIEndpoints verifies that all known API endpoints are still reachable.
// Run with: go test -tags integration -run TestAPIEndpoints -v ./internal/api/
// Requires PRODUCTPLAN_API_TOKEN environment variable.
func TestAPIEndpoints(t *testing.T) {
	token := os.Getenv("PRODUCTPLAN_API_TOKEN")
	if token == "" {
		t.Fatal("PRODUCTPLAN_API_TOKEN environment variable is required")
	}

	client, err := New(Config{
		Token:  token,
		Logger: logging.Nop(),
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Load endpoint snapshot
	snapshotData, err := os.ReadFile("../../testdata/api-endpoints.json")
	if err != nil {
		t.Fatalf("failed to read api-endpoints.json: %v", err)
	}

	var snapshot struct {
		Endpoints []struct {
			Method      string `json:"method"`
			Path        string `json:"path"`
			Description string `json:"description"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(snapshotData, &snapshot); err != nil {
		t.Fatalf("failed to parse api-endpoints.json: %v", err)
	}

	// Test only GET endpoints with list paths (no IDs needed)
	listEndpoints := []struct {
		name string
		fn   func(ctx context.Context) (json.RawMessage, error)
	}{
		{"GET /roadmaps", func(ctx context.Context) (json.RawMessage, error) { return client.ListRoadmaps(ctx) }},
		{"GET /strategy/objectives", func(ctx context.Context) (json.RawMessage, error) { return client.ListObjectives(ctx) }},
		{"GET /discovery/ideas", func(ctx context.Context) (json.RawMessage, error) { return client.ListIdeas(ctx) }},
		{"GET /discovery/ideas/customers", func(ctx context.Context) (json.RawMessage, error) { return client.ListAllCustomers(ctx) }},
		{"GET /discovery/ideas/tags", func(ctx context.Context) (json.RawMessage, error) { return client.ListAllTags(ctx) }},
		{"GET /discovery/opportunities", func(ctx context.Context) (json.RawMessage, error) { return client.ListOpportunities(ctx) }},
		{"GET /discovery/idea_forms", func(ctx context.Context) (json.RawMessage, error) { return client.ListIdeaForms(ctx) }},
		{"GET /launches", func(ctx context.Context) (json.RawMessage, error) { return client.ListLaunches(ctx) }},
		{"GET /users", func(ctx context.Context) (json.RawMessage, error) { return client.ListUsers(ctx) }},
		{"GET /teams", func(ctx context.Context) (json.RawMessage, error) { return client.ListTeams(ctx) }},
		{"GET /status", func(ctx context.Context) (json.RawMessage, error) { return client.CheckStatus(ctx) }},
	}

	for _, ep := range listEndpoints {
		t.Run(ep.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := ep.fn(ctx)
			if err != nil {
				t.Errorf("endpoint %s failed: %v", ep.name, err)
			}
		})
	}

	// Also verify authentication works by checking the status endpoint directly
	t.Run("auth_header", func(t *testing.T) {
		req, err := http.NewRequest("GET", "https://app.productplan.com/api/v2/status", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})
}
