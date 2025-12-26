package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

// testServer creates a mock server that returns predefined responses
func testServer(t *testing.T, responses map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		key := method + " " + path
		if resp, ok := responses[key]; ok {
			w.WriteHeader(200)
			w.Write([]byte(resp))
			return
		}

		// Try path-only match for GET requests
		if resp, ok := responses[path]; ok {
			w.WriteHeader(200)
			w.Write([]byte(resp))
			return
		}

		t.Logf("unhandled request: %s %s", method, path)
		w.WriteHeader(404)
	}))
}

func testClient(t *testing.T, server *httptest.Server) *Client {
	client, err := New(Config{
		Token:   "test-token",
		BaseURL: server.URL,
		Logger:  logging.Nop(),
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

// ============================================================================
// Roadmaps Tests
// ============================================================================

func TestListRoadmaps(t *testing.T) {
	server := testServer(t, map[string]string{
		"/roadmaps": `[{"id": 1, "name": "Test Roadmap", "updated_at": "2024-01-01"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.ListRoadmaps(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", parsed["count"])
	}
}

func TestGetRoadmap(t *testing.T) {
	server := testServer(t, map[string]string{
		"/roadmaps/123": `{"id": 123, "name": "My Roadmap"}`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.GetRoadmap(context.Background(), "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["id"] != float64(123) {
		t.Errorf("expected id 123, got %v", parsed["id"])
	}
}

func TestGetRoadmapBars(t *testing.T) {
	server := testServer(t, map[string]string{
		"/roadmaps/1/bars":  `[{"id": 10, "name": "Bar", "lane_id": 100}]`,
		"/roadmaps/1/lanes": `[{"id": 100, "name": "Engineering"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.GetRoadmapBars(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Bars []map[string]any `json:"bars"`
	}
	json.Unmarshal(result, &parsed)

	if len(parsed.Bars) != 1 {
		t.Fatalf("expected 1 bar, got %d", len(parsed.Bars))
	}
	if parsed.Bars[0]["lane_name"] != "Engineering" {
		t.Errorf("expected lane_name 'Engineering', got %v", parsed.Bars[0]["lane_name"])
	}
}

func TestGetRoadmapLanes(t *testing.T) {
	server := testServer(t, map[string]string{
		"/roadmaps/1/lanes": `[{"id": 1, "name": "Dev", "color": "#FF0000"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.GetRoadmapLanes(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", parsed["count"])
	}
}

func TestGetRoadmapMilestones(t *testing.T) {
	server := testServer(t, map[string]string{
		"/roadmaps/1/milestones": `[{"id": 1, "name": "Launch", "date": "2024-06-01"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.GetRoadmapMilestones(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", parsed["count"])
	}
}

// ============================================================================
// Bars Tests
// ============================================================================

func TestGetBar(t *testing.T) {
	server := testServer(t, map[string]string{
		"/bars/42": `{"id": 42, "name": "Feature Bar"}`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.GetBar(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["id"] != float64(42) {
		t.Errorf("expected id 42, got %v", parsed["id"])
	}
}

func TestCreateBar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/bars") {
			t.Errorf("expected /bars path, got %s", r.URL.Path)
		}
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 99, "name": "New Bar"}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	result, err := client.CreateBar(context.Background(), map[string]any{
		"name":       "New Bar",
		"roadmap_id": 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["id"] != float64(99) {
		t.Errorf("expected id 99, got %v", parsed["id"])
	}
}

func TestUpdateBar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"id": 42, "name": "Updated Bar"}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	result, err := client.UpdateBar(context.Background(), "42", map[string]any{
		"name": "Updated Bar",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["name"] != "Updated Bar" {
		t.Errorf("expected name 'Updated Bar', got %v", parsed["name"])
	}
}

func TestDeleteBar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.DeleteBar(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetBarChildren(t *testing.T) {
	server := testServer(t, map[string]string{
		"/bars/1/child_bars": `[{"id": 10}, {"id": 11}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.GetBarChildren(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []any
	json.Unmarshal(result, &parsed)
	if len(parsed) != 2 {
		t.Errorf("expected 2 children, got %d", len(parsed))
	}
}

// ============================================================================
// Bar Comments Tests
// ============================================================================

func TestGetBarComments(t *testing.T) {
	server := testServer(t, map[string]string{
		"/bars/1/comments": `[{"id": 1, "text": "Comment"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetBarComments(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateBarComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1, "text": "New comment"}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.CreateBarComment(context.Background(), "1", map[string]any{"text": "New comment"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Bar Connections Tests
// ============================================================================

func TestGetBarConnections(t *testing.T) {
	server := testServer(t, map[string]string{
		"/bars/1/connections": `[{"id": 1, "target_id": 2}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetBarConnections(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateBarConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.CreateBarConnection(context.Background(), "1", map[string]any{"target_id": 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteBarConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/connections/") {
			t.Errorf("expected connections path, got %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.DeleteBarConnection(context.Background(), "1", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Bar Links Tests
// ============================================================================

func TestGetBarLinks(t *testing.T) {
	server := testServer(t, map[string]string{
		"/bars/1/links": `[{"id": 1, "url": "https://example.com"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetBarLinks(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateBarLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.CreateBarLink(context.Background(), "1", map[string]any{"url": "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateBarLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.UpdateBarLink(context.Background(), "1", "10", map[string]any{"url": "https://updated.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteBarLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.DeleteBarLink(context.Background(), "1", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Lanes Tests
// ============================================================================

func TestCreateLane(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1, "name": "New Lane"}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.CreateLane(context.Background(), "1", map[string]any{"name": "New Lane"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateLane(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.UpdateLane(context.Background(), "1", "10", map[string]any{"name": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteLane(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.DeleteLane(context.Background(), "1", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Milestones Tests
// ============================================================================

func TestCreateMilestone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.CreateMilestone(context.Background(), "1", map[string]any{"name": "Launch"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateMilestone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.UpdateMilestone(context.Background(), "1", "10", map[string]any{"name": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMilestone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.DeleteMilestone(context.Background(), "1", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Objectives Tests
// ============================================================================

func TestListObjectives(t *testing.T) {
	server := testServer(t, map[string]string{
		"/strategy/objectives": `[{"id": 1, "name": "OKR", "status": "on_track"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.ListObjectives(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", parsed["count"])
	}
}

func TestGetObjective(t *testing.T) {
	server := testServer(t, map[string]string{
		"/strategy/objectives/1": `{"id": 1, "name": "Objective"}`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetObjective(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateObjective(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.CreateObjective(context.Background(), map[string]any{"name": "New OKR"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateObjective(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.UpdateObjective(context.Background(), "1", map[string]any{"status": "at_risk"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteObjective(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.DeleteObjective(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Key Results Tests
// ============================================================================

func TestListKeyResults(t *testing.T) {
	server := testServer(t, map[string]string{
		"/strategy/objectives/1/key_results": `[{"id": 1}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.ListKeyResults(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateKeyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.CreateKeyResult(context.Background(), "1", map[string]any{"name": "KR"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateKeyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.UpdateKeyResult(context.Background(), "1", "10", map[string]any{"progress": 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteKeyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.DeleteKeyResult(context.Background(), "1", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Ideas Tests
// ============================================================================

func TestListIdeas(t *testing.T) {
	server := testServer(t, map[string]string{
		"/discovery/ideas": `{"results": [{"id": 1, "name": "Idea"}]}`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.ListIdeas(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", parsed["count"])
	}
}

func TestGetIdea(t *testing.T) {
	server := testServer(t, map[string]string{
		"/discovery/ideas/1": `{"id": 1}`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetIdea(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateIdea(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.CreateIdea(context.Background(), map[string]any{"name": "New Idea"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateIdea(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.UpdateIdea(context.Background(), "1", map[string]any{"name": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Idea Customers Tests
// ============================================================================

func TestGetIdeaCustomers(t *testing.T) {
	server := testServer(t, map[string]string{
		"/discovery/ideas/1/customers": `[{"id": 1}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetIdeaCustomers(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddIdeaCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.AddIdeaCustomer(context.Background(), "1", map[string]any{"customer_id": 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveIdeaCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.RemoveIdeaCustomer(context.Background(), "1", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Idea Tags Tests
// ============================================================================

func TestGetIdeaTags(t *testing.T) {
	server := testServer(t, map[string]string{
		"/discovery/ideas/1/tags": `[{"id": 1}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetIdeaTags(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddIdeaTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.AddIdeaTag(context.Background(), "1", map[string]any{"tag_id": 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveIdeaTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.RemoveIdeaTag(context.Background(), "1", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Opportunities Tests
// ============================================================================

func TestListOpportunities(t *testing.T) {
	server := testServer(t, map[string]string{
		"/discovery/opportunities": `{"results": [{"id": 1}]}`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.ListOpportunities(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", parsed["count"])
	}
}

func TestGetOpportunity(t *testing.T) {
	server := testServer(t, map[string]string{
		"/discovery/opportunities/1": `{"id": 1}`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetOpportunity(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateOpportunity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.CreateOpportunity(context.Background(), map[string]any{"problem_statement": "Problem"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateOpportunity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.UpdateOpportunity(context.Background(), "1", map[string]any{"status": "validated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteOpportunity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.DeleteOpportunity(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Idea Forms Tests
// ============================================================================

func TestListIdeaForms(t *testing.T) {
	server := testServer(t, map[string]string{
		"/discovery/idea_forms": `[{"id": 1}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.ListIdeaForms(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetIdeaForm(t *testing.T) {
	server := testServer(t, map[string]string{
		"/discovery/idea_forms/1": `{"id": 1}`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetIdeaForm(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Launches Tests
// ============================================================================

func TestListLaunches(t *testing.T) {
	server := testServer(t, map[string]string{
		"/launches": `[{"id": 1, "name": "v1.0", "status": "planned"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.ListLaunches(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", parsed["count"])
	}
}

func TestGetLaunch(t *testing.T) {
	server := testServer(t, map[string]string{
		"/launches/1": `{"id": 1}`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.GetLaunch(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// Admin Tests
// ============================================================================

func TestListUsers(t *testing.T) {
	server := testServer(t, map[string]string{
		"/users": `[{"id": 1, "email": "user@example.com"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTeams(t *testing.T) {
	server := testServer(t, map[string]string{
		"/teams": `[{"id": 1, "name": "Engineering"}]`,
	})
	defer server.Close()

	client := testClient(t, server)
	_, err := client.ListTeams(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckStatus(t *testing.T) {
	server := testServer(t, map[string]string{
		"/status": `{"status": "ok"}`,
	})
	defer server.Close()

	client := testClient(t, server)
	result, err := client.CheckStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	json.Unmarshal(result, &parsed)
	if parsed["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", parsed["status"])
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestEndpointError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	client := testClient(t, server)
	_, err := client.ListRoadmaps(context.Background())
	if err == nil {
		t.Error("expected error for 500 response")
	}
}
