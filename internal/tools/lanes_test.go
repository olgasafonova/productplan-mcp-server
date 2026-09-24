package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

// laneRecorder answers every request with {} and records write bodies.
func laneRecorder(t *testing.T) (*api.Client, *[]recordedRequest) {
	t.Helper()
	var got []recordedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		got = append(got, recordedRequest{Method: r.Method, Path: r.URL.Path, Body: body})
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	client, err := api.New(api.Config{Token: "t", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	return client, &got
}

// manage_lane sends exactly the documented LaneCreate / lane PATCH fields:
// name, description, position (productplan.readme.io, fetched 24-09-2026).
func TestManageLane_SendsDocumentedFields(t *testing.T) {
	for name, tc := range map[string]struct {
		args       map[string]any
		wantMethod string
		wantPath   string
		wantBody   map[string]any
	}{
		"create all fields": {
			args:       map[string]any{"action": "create", "roadmap_id": "7", "name": "Backend", "description": "Server work", "position": 2},
			wantMethod: http.MethodPost, wantPath: "/roadmaps/7/lanes",
			wantBody: map[string]any{"name": "Backend", "description": "Server work", "position": float64(2)},
		},
		"create name only": {
			args:       map[string]any{"action": "create", "roadmap_id": "7", "name": "Backend"},
			wantMethod: http.MethodPost, wantPath: "/roadmaps/7/lanes",
			wantBody: map[string]any{"name": "Backend"},
		},
		"update position only": {
			args:       map[string]any{"action": "update", "roadmap_id": "7", "lane_id": "9", "position": 1},
			wantMethod: http.MethodPatch, wantPath: "/roadmaps/7/lanes/9",
			wantBody: map[string]any{"position": float64(1)},
		},
		"update description": {
			args:       map[string]any{"action": "update", "roadmap_id": "7", "lane_id": "9", "description": "Moved"},
			wantMethod: http.MethodPatch, wantPath: "/roadmaps/7/lanes/9",
			wantBody: map[string]any{"description": "Moved"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			client, got := laneRecorder(t)
			if _, err := manageLaneHandler(client).Handle(context.Background(), tc.args); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(*got) != 1 {
				t.Fatalf("want 1 request, got %d", len(*got))
			}
			req := (*got)[0]
			if req.Method != tc.wantMethod || req.Path != tc.wantPath {
				t.Errorf("request = %s %s, want %s %s", req.Method, req.Path, tc.wantMethod, tc.wantPath)
			}
			if !reflect.DeepEqual(req.Body, tc.wantBody) {
				t.Errorf("body = %v, want %v", req.Body, tc.wantBody)
			}
		})
	}
}

// color stays declared (removing it would break callers, Article XII) but
// is rejected with an explanation instead of being sent and ignored.
func TestManageLane_RejectsColor(t *testing.T) {
	client, got := laneRecorder(t)
	_, err := manageLaneHandler(client).Handle(context.Background(),
		map[string]any{"action": "create", "roadmap_id": "7", "name": "Backend", "color": "#FF0000"})
	if err == nil || !strings.Contains(err.Error(), "color is not a ProductPlan lane field") {
		t.Fatalf("want the color explanation, got %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("a request was sent despite the rejection: %v", *got)
	}
	for _, tool := range roadmapManageTools() {
		if tool.Name == "manage_lane" {
			if _, ok := tool.InputSchema.Properties["color"]; !ok {
				t.Error("color must stay declared on manage_lane")
			}
			for _, arg := range []string{"description", "position"} {
				if _, ok := tool.InputSchema.Properties[arg]; !ok {
					t.Errorf("manage_lane does not declare %s", arg)
				}
			}
		}
	}
}
