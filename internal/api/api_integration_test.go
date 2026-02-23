//go:build integration

package api

import (
	"context"
	"encoding/json"
	"fmt"
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
		t.Skip("PRODUCTPLAN_API_TOKEN not set, skipping integration tests")
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

// TestAPIEndpointDiscovery probes candidate endpoints to detect new API additions.
// Candidates are common REST patterns that ProductPlan might add. Any that return
// 200 are reported so we can add them to the MCP server.
func TestAPIEndpointDiscovery(t *testing.T) {
	token := os.Getenv("PRODUCTPLAN_API_TOKEN")
	if token == "" {
		t.Skip("PRODUCTPLAN_API_TOKEN not set, skipping discovery tests")
	}

	// Load candidates
	candidateData, err := os.ReadFile("../../testdata/api-candidates.json")
	if err != nil {
		t.Fatalf("failed to read api-candidates.json: %v", err)
	}

	var candidates struct {
		Candidates []string `json:"candidates"`
	}
	if err := json.Unmarshal(candidateData, &candidates); err != nil {
		t.Fatalf("failed to parse api-candidates.json: %v", err)
	}

	// Load known endpoints to skip
	snapshotData, err := os.ReadFile("../../testdata/api-endpoints.json")
	if err != nil {
		t.Fatalf("failed to read api-endpoints.json: %v", err)
	}

	var snapshot struct {
		Endpoints []struct {
			Path string `json:"path"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(snapshotData, &snapshot); err != nil {
		t.Fatalf("failed to parse api-endpoints.json: %v", err)
	}

	known := make(map[string]bool)
	for _, ep := range snapshot.Endpoints {
		known[ep.Path] = true
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	baseURL := DefaultBaseURL

	var discovered []string

	for _, path := range candidates.Candidates {
		if known[path] {
			continue
		}

		req, err := http.NewRequest("GET", baseURL+path, nil)
		if err != nil {
			t.Logf("skip %s: %v", path, err)
			continue
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Logf("skip %s: %v", path, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			discovered = append(discovered, path)
			t.Logf("DISCOVERED: GET %s returned 200", path)
		}
	}

	if len(discovered) > 0 {
		msg := fmt.Sprintf("Found %d new endpoint(s) not in api-endpoints.json:\n", len(discovered))
		for _, ep := range discovered {
			msg += fmt.Sprintf("  GET %s\n", ep)
		}
		msg += "Add these to testdata/api-endpoints.json and implement in the MCP server."
		t.Error(msg)
	}
}
