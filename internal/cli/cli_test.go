package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

func setupTestCLI(t *testing.T, response any) (*CLI, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	client, err := api.New(api.Config{
		Token:   "test-token",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	output := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	cli := New(client, Config{
		Version: "test",
		Output:  output,
		Error:   errOut,
	})

	return cli, server
}

func TestCLI_Run_NoArgs(t *testing.T) {
	cli, server := setupTestCLI(t, map[string]any{})
	defer server.Close()

	code := cli.Run([]string{})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestCLI_Run_UnknownCommand(t *testing.T) {
	cli, server := setupTestCLI(t, map[string]any{})
	defer server.Close()

	code := cli.Run([]string{"unknown"})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestCLI_Run_Roadmaps(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		response any
	}{
		{"list", []string{"roadmaps"}, []map[string]any{{"id": "1", "name": "Roadmap 1"}}},
		{"get", []string{"roadmaps", "123"}, map[string]any{"id": "123", "name": "Roadmap"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, server := setupTestCLI(t, tt.response)
			defer server.Close()

			code := cli.Run(tt.args)
			if code != 0 {
				t.Errorf("expected exit code 0, got %d", code)
			}
		})
	}
}

func TestCLI_Run_Bars(t *testing.T) {
	cli, server := setupTestCLI(t, []map[string]any{{"id": "1"}})
	defer server.Close()

	// Without roadmap_id
	code := cli.Run([]string{"bars"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing roadmap_id, got %d", code)
	}

	// With roadmap_id
	code = cli.Run([]string{"bars", "123"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestCLI_Run_Lanes(t *testing.T) {
	cli, server := setupTestCLI(t, []map[string]any{{"id": "1"}})
	defer server.Close()

	// Without roadmap_id
	code := cli.Run([]string{"lanes"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing roadmap_id, got %d", code)
	}

	// With roadmap_id
	code = cli.Run([]string{"lanes", "123"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestCLI_Run_Milestones(t *testing.T) {
	cli, server := setupTestCLI(t, []map[string]any{{"id": "1"}})
	defer server.Close()

	// Without roadmap_id
	code := cli.Run([]string{"milestones"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing roadmap_id, got %d", code)
	}

	// With roadmap_id
	code = cli.Run([]string{"milestones", "123"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestCLI_Run_Objectives(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"list", []string{"objectives"}},
		{"get", []string{"objectives", "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, server := setupTestCLI(t, map[string]any{"id": "1"})
			defer server.Close()

			code := cli.Run(tt.args)
			if code != 0 {
				t.Errorf("expected exit code 0, got %d", code)
			}
		})
	}
}

func TestCLI_Run_KeyResults(t *testing.T) {
	cli, server := setupTestCLI(t, []map[string]any{{"id": "1"}})
	defer server.Close()

	// Without objective_id
	code := cli.Run([]string{"key-results"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing objective_id, got %d", code)
	}

	// With objective_id
	code = cli.Run([]string{"key-results", "123"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestCLI_Run_Ideas(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"list", []string{"ideas"}},
		{"get", []string{"ideas", "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, server := setupTestCLI(t, map[string]any{"id": "1"})
			defer server.Close()

			code := cli.Run(tt.args)
			if code != 0 {
				t.Errorf("expected exit code 0, got %d", code)
			}
		})
	}
}

func TestCLI_Run_Launches(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"list", []string{"launches"}},
		{"get", []string{"launches", "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, server := setupTestCLI(t, map[string]any{"id": "1"})
			defer server.Close()

			code := cli.Run(tt.args)
			if code != 0 {
				t.Errorf("expected exit code 0, got %d", code)
			}
		})
	}
}

func TestCLI_Run_Opportunities(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"list", []string{"opportunities"}},
		{"get", []string{"opportunities", "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, server := setupTestCLI(t, map[string]any{"id": "1"})
			defer server.Close()

			code := cli.Run(tt.args)
			if code != 0 {
				t.Errorf("expected exit code 0, got %d", code)
			}
		})
	}
}

func TestCLI_Run_Status(t *testing.T) {
	cli, server := setupTestCLI(t, map[string]any{"status": "ok"})
	defer server.Close()

	code := cli.Run([]string{"status"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestCLI_PrintUsage(t *testing.T) {
	output := &bytes.Buffer{}
	cli := &CLI{
		output:  output,
		version: "1.2.3",
	}

	cli.PrintUsage()

	usage := output.String()
	if !strings.Contains(usage, "ProductPlan CLI") {
		t.Error("usage should contain 'ProductPlan CLI'")
	}
	if !strings.Contains(usage, "1.2.3") {
		t.Error("usage should contain version")
	}
	if !strings.Contains(usage, "roadmaps") {
		t.Error("usage should contain 'roadmaps' command")
	}
}

func TestCLI_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := api.New(api.Config{
		Token:   "test-token",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	errOut := &bytes.Buffer{}
	cli := New(client, Config{
		Version: "test",
		Error:   errOut,
	})

	code := cli.Run([]string{"status"})
	if code != 1 {
		t.Errorf("expected exit code 1 for API error, got %d", code)
	}
	if !strings.Contains(errOut.String(), "Error") {
		t.Error("expected error message in stderr")
	}
}
