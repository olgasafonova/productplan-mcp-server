package evals

import (
	"encoding/json"
	"testing"
)

// MockSelector implements ToolSelector for testing
type MockSelector struct {
	responses map[string]struct {
		tool string
		args map[string]interface{}
	}
}

func NewMockSelector() *MockSelector {
	return &MockSelector{
		responses: make(map[string]struct {
			tool string
			args map[string]interface{}
		}),
	}
}

func (m *MockSelector) SetResponse(prompt, tool string, args map[string]interface{}) {
	m.responses[prompt] = struct {
		tool string
		args map[string]interface{}
	}{tool: tool, args: args}
}

func (m *MockSelector) SelectTool(prompt string) (string, map[string]interface{}, error) {
	if resp, ok := m.responses[prompt]; ok {
		return resp.tool, resp.args, nil
	}
	return "", nil, nil
}

func TestLoadToolSelectionSuite(t *testing.T) {
	suite, err := LoadToolSelectionSuite("tool_selection.json")
	if err != nil {
		t.Fatalf("Failed to load tool selection suite: %v", err)
	}

	if suite.Name == "" {
		t.Error("Suite name should not be empty")
	}

	if len(suite.Tests) == 0 {
		t.Error("Suite should have tests")
	}

	// Check first test has required fields
	test := suite.Tests[0]
	if test.ID == "" {
		t.Error("Test ID should not be empty")
	}
	if test.Prompt == "" {
		t.Error("Test prompt should not be empty")
	}
	if test.ExpectedTool == "" {
		t.Error("Test expected_tool should not be empty")
	}
}

func TestLoadConfusionPairSuite(t *testing.T) {
	suite, err := LoadConfusionPairSuite("confusion_pairs.json")
	if err != nil {
		t.Fatalf("Failed to load confusion pairs suite: %v", err)
	}

	if suite.Name == "" {
		t.Error("Suite name should not be empty")
	}

	if len(suite.Pairs) == 0 {
		t.Error("Suite should have pairs")
	}

	// Check first pair has required fields
	pair := suite.Pairs[0]
	if len(pair.Tools) < 2 {
		t.Error("Pair should have at least 2 tools")
	}
	if len(pair.Tests) == 0 {
		t.Error("Pair should have tests")
	}
}

func TestLoadArgumentSuite(t *testing.T) {
	suite, err := LoadArgumentSuite("argument_correctness.json")
	if err != nil {
		t.Fatalf("Failed to load argument suite: %v", err)
	}

	if suite.Name == "" {
		t.Error("Suite name should not be empty")
	}

	if len(suite.Tests) == 0 {
		t.Error("Suite should have tests")
	}

	// Check first test has required fields
	test := suite.Tests[0]
	if test.ID == "" {
		t.Error("Test ID should not be empty")
	}
	if test.Tool == "" {
		t.Error("Test tool should not be empty")
	}
	if test.Prompt == "" {
		t.Error("Test prompt should not be empty")
	}
}

func TestEvaluateToolSelection(t *testing.T) {
	suite := &ToolSelectionSuite{
		Name: "Test Suite",
		Tests: []ToolSelectionTest{
			{
				ID:           "test-1",
				Prompt:       "Show my roadmaps",
				ExpectedTool: "list_roadmaps",
				Category:     "roadmaps",
				Difficulty:   "easy",
			},
			{
				ID:           "test-2",
				Prompt:       "Create a bar",
				ExpectedTool: "manage_bar",
				Category:     "create",
				Difficulty:   "easy",
			},
		},
	}

	selector := NewMockSelector()
	selector.SetResponse("Show my roadmaps", "list_roadmaps", nil)
	selector.SetResponse("Create a bar", "manage_bar", nil)

	metrics, results := EvaluateToolSelection(suite, selector)

	if metrics.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", metrics.TotalTests)
	}
	if metrics.PassedTests != 2 {
		t.Errorf("Expected 2 passed tests, got %d", metrics.PassedTests)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestEvaluateToolSelectionWithFailure(t *testing.T) {
	suite := &ToolSelectionSuite{
		Name: "Test Suite",
		Tests: []ToolSelectionTest{
			{
				ID:           "test-1",
				Prompt:       "Show my roadmaps",
				ExpectedTool: "list_roadmaps",
				Category:     "roadmaps",
				Difficulty:   "easy",
			},
		},
	}

	selector := NewMockSelector()
	selector.SetResponse("Show my roadmaps", "get_roadmap", nil) // Wrong tool

	metrics, results := EvaluateToolSelection(suite, selector)

	if metrics.PassedTests != 0 {
		t.Errorf("Expected 0 passed tests, got %d", metrics.PassedTests)
	}
	if metrics.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", metrics.FailedTests)
	}
	if results[0].Passed {
		t.Error("Expected test to fail")
	}
}

func TestEvaluateConfusionPairs(t *testing.T) {
	suite := &ConfusionPairSuite{
		Name: "Test Suite",
		Pairs: []ConfusionPair{
			{
				Tools:       []string{"list_roadmaps", "get_roadmap"},
				Distinction: "list returns all, get returns one",
				Tests: []ConfusionPairTest{
					{
						Prompt:       "Show all roadmaps",
						ExpectedTool: "list_roadmaps",
						Rationale:    "Listing all roadmaps",
					},
					{
						Prompt:       "Get roadmap 123",
						ExpectedTool: "get_roadmap",
						Rationale:    "Getting specific roadmap",
					},
				},
			},
		},
	}

	selector := NewMockSelector()
	selector.SetResponse("Show all roadmaps", "list_roadmaps", nil)
	selector.SetResponse("Get roadmap 123", "get_roadmap", nil)

	metrics, results := EvaluateConfusionPairs(suite, selector)

	if metrics.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", metrics.TotalTests)
	}
	if metrics.PassedTests != 2 {
		t.Errorf("Expected 2 passed tests, got %d", metrics.PassedTests)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestEvaluateArguments(t *testing.T) {
	suite := &ArgumentSuite{
		Name: "Test Suite",
		Tests: []ArgumentTest{
			{
				ID:     "test-1",
				Tool:   "manage_bar",
				Prompt: "Create bar named Test on roadmap 123",
				ExpectedArgs: map[string]interface{}{
					"name":       "Test",
					"roadmap_id": "123",
				},
				RequiredArgs: []string{"action", "roadmap_id"},
				Category:     "create",
			},
		},
	}

	selector := NewMockSelector()
	selector.SetResponse("Create bar named Test on roadmap 123", "manage_bar", map[string]interface{}{
		"action":     "create",
		"roadmap_id": "123",
		"name":       "Test",
	})

	metrics, results := EvaluateArguments(suite, selector)

	if metrics.TotalTests != 1 {
		t.Errorf("Expected 1 total test, got %d", metrics.TotalTests)
	}
	if metrics.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", metrics.PassedTests)
	}
	if !results[0].Passed {
		t.Errorf("Expected test to pass, got: missing=%v, wrong=%v", results[0].MissingArgs, results[0].WrongArgs)
	}
}

func TestEvaluateArgumentsWithMissingArg(t *testing.T) {
	suite := &ArgumentSuite{
		Name: "Test Suite",
		Tests: []ArgumentTest{
			{
				ID:           "test-1",
				Tool:         "manage_bar",
				Prompt:       "Create bar",
				ExpectedArgs: map[string]interface{}{},
				RequiredArgs: []string{"action", "roadmap_id"},
				Category:     "create",
			},
		},
	}

	selector := NewMockSelector()
	selector.SetResponse("Create bar", "manage_bar", map[string]interface{}{
		"action": "create",
		// Missing "roadmap_id"
	})

	metrics, results := EvaluateArguments(suite, selector)

	if metrics.PassedTests != 0 {
		t.Errorf("Expected 0 passed tests, got %d", metrics.PassedTests)
	}
	if len(results[0].MissingArgs) == 0 {
		t.Error("Expected missing args")
	}
}

func TestCompareValues(t *testing.T) {
	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
		want     bool
	}{
		{"equal strings", "hello", "hello", true},
		{"different strings", "hello", "world", false},
		{"equal ints", 42, 42, true},
		{"int vs float64", 42, float64(42), true},
		{"different numbers", 42, float64(43), false},
		{"equal slices", []string{"a", "b"}, []string{"a", "b"}, true},
		{"different slices", []string{"a", "b"}, []string{"a", "c"}, false},
		{"nil values", nil, nil, true},
		{"nil vs value", nil, "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareValues(tt.expected, tt.actual)
			if got != tt.want {
				t.Errorf("compareValues(%v, %v) = %v, want %v", tt.expected, tt.actual, got, tt.want)
			}
		})
	}
}

func TestFormatMetrics(t *testing.T) {
	metrics := &EvalMetrics{
		TotalTests:  10,
		PassedTests: 8,
		FailedTests: 2,
		Accuracy:    0.8,
		ByCategory: map[string]*CategoryMetrics{
			"roadmaps": {Total: 5, Passed: 4, Failed: 1},
			"create":   {Total: 5, Passed: 4, Failed: 1},
		},
		FailedDetails: []string{"[test-1] prompt: error"},
	}

	output := FormatMetrics(metrics, "Test Suite")

	if output == "" {
		t.Error("FormatMetrics should return non-empty string")
	}
	if !contains(output, "Test Suite") {
		t.Error("Output should contain suite name")
	}
	if !contains(output, "80.0%") {
		t.Error("Output should contain accuracy percentage")
	}
}

func TestLoadAllEvals(t *testing.T) {
	toolSelection, confusionPairs, arguments, err := LoadAllEvals(".")
	if err != nil {
		t.Fatalf("Failed to load all evals: %v", err)
	}

	if toolSelection == nil {
		t.Fatal("Tool selection suite should not be nil")
	}
	if confusionPairs == nil {
		t.Fatal("Confusion pairs suite should not be nil")
	}
	if arguments == nil {
		t.Fatal("Arguments suite should not be nil")
	}

	// Count total tests
	total := len(toolSelection.Tests)
	for _, pair := range confusionPairs.Pairs {
		total += len(pair.Tests)
	}
	total += len(arguments.Tests)

	t.Logf("Loaded %d total evaluation tests", total)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestExportMetricsJSON(t *testing.T) {
	metrics := &EvalMetrics{
		TotalTests:  10,
		PassedTests: 8,
		FailedTests: 2,
		Accuracy:    0.8,
		ByCategory: map[string]*CategoryMetrics{
			"roadmaps": {Total: 5, Passed: 5, Failed: 0},
		},
	}

	data, err := ExportMetricsJSON(metrics, "Test Suite")
	if err != nil {
		t.Fatalf("failed to export JSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if result["suite"] != "Test Suite" {
		t.Error("expected suite name in JSON output")
	}
	if result["total_tests"].(float64) != 10 {
		t.Error("expected total_tests=10 in JSON output")
	}
	if result["accuracy"].(float64) != 0.8 {
		t.Error("expected accuracy=0.8 in JSON output")
	}
}

func TestExportMetricsJSON_WithFailedDetails(t *testing.T) {
	metrics := &EvalMetrics{
		TotalTests:    5,
		PassedTests:   3,
		FailedTests:   2,
		Accuracy:      0.6,
		ByCategory:    map[string]*CategoryMetrics{},
		FailedDetails: []string{"[test-1] failed", "[test-2] failed"},
	}

	data, err := ExportMetricsJSON(metrics, "Test Suite")
	if err != nil {
		t.Fatalf("failed to export JSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	details, ok := result["failed_details"].([]interface{})
	if !ok {
		t.Fatal("expected failed_details in JSON output")
	}
	if len(details) != 2 {
		t.Errorf("expected 2 failed details, got %d", len(details))
	}
}

func TestGenerateCombinedReport(t *testing.T) {
	toolSelection := &EvalMetrics{TotalTests: 10, PassedTests: 9, Accuracy: 0.9}
	confusionPairs := &EvalMetrics{TotalTests: 20, PassedTests: 18, Accuracy: 0.9}
	arguments := &EvalMetrics{TotalTests: 15, PassedTests: 12, Accuracy: 0.8}

	report := GenerateCombinedReport(toolSelection, confusionPairs, arguments, 0.8)

	if report == nil {
		t.Fatal("expected non-nil report")
	}

	// 39/45 = 0.8666...
	expectedOverall := float64(39) / float64(45)
	if report.OverallAccuracy != expectedOverall {
		t.Errorf("expected overall accuracy %.4f, got %.4f", expectedOverall, report.OverallAccuracy)
	}

	if !report.PassThreshold {
		t.Error("expected to pass 80% threshold")
	}

	if len(report.Summary) != 3 {
		t.Errorf("expected 3 summary entries, got %d", len(report.Summary))
	}
}

func TestGenerateCombinedReport_FailThreshold(t *testing.T) {
	toolSelection := &EvalMetrics{TotalTests: 10, PassedTests: 5, Accuracy: 0.5}
	confusionPairs := &EvalMetrics{TotalTests: 10, PassedTests: 5, Accuracy: 0.5}
	arguments := &EvalMetrics{TotalTests: 10, PassedTests: 5, Accuracy: 0.5}

	report := GenerateCombinedReport(toolSelection, confusionPairs, arguments, 0.8)

	if report.PassThreshold {
		t.Error("expected to fail 80% threshold with 50% accuracy")
	}
}

func TestExportCombinedReportJSON(t *testing.T) {
	report := &CombinedReport{
		Timestamp:       "0",
		OverallAccuracy: 0.85,
		PassThreshold:   true,
		Summary: map[string]*CategoryTotal{
			"tool_selection": {Total: 10, Passed: 9, Rate: 0.9},
			"confusion_pairs": {Total: 20, Passed: 17, Rate: 0.85},
			"arguments":       {Total: 15, Passed: 12, Rate: 0.8},
		},
	}

	data, err := ExportCombinedReportJSON(report)
	if err != nil {
		t.Fatalf("failed to export combined report: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result["pass_threshold"] != true {
		t.Error("expected pass_threshold=true in JSON output")
	}
	if result["overall_accuracy"].(float64) != 0.85 {
		t.Error("expected overall_accuracy=0.85 in JSON output")
	}
}

func TestToolSelectionSuiteDifficultyLevels(t *testing.T) {
	suite, err := LoadToolSelectionSuite("tool_selection.json")
	if err != nil {
		t.Fatalf("failed to load suite: %v", err)
	}

	difficulties := make(map[string]int)
	for _, test := range suite.Tests {
		difficulties[test.Difficulty]++
	}

	if difficulties["easy"] == 0 {
		t.Error("expected at least one easy test")
	}
	if difficulties["medium"] == 0 {
		t.Error("expected at least one medium test")
	}
	if difficulties["hard"] == 0 {
		t.Error("expected at least one hard test")
	}

	t.Logf("Difficulty distribution: easy=%d, medium=%d, hard=%d",
		difficulties["easy"], difficulties["medium"], difficulties["hard"])
}

func TestConfusionPairTestCounts(t *testing.T) {
	suite, err := LoadConfusionPairSuite("confusion_pairs.json")
	if err != nil {
		t.Fatalf("failed to load suite: %v", err)
	}

	totalTests := 0
	for _, pair := range suite.Pairs {
		totalTests += len(pair.Tests)
	}

	// We enhanced to have at least 15 pairs with 4 tests each = 60
	if totalTests < 40 {
		t.Errorf("expected at least 40 confusion pair tests, got %d", totalTests)
	}

	t.Logf("Confusion pairs: %d pairs, %d tests", len(suite.Pairs), totalTests)
}

func TestArgumentSuiteCategories(t *testing.T) {
	suite, err := LoadArgumentSuite("argument_correctness.json")
	if err != nil {
		t.Fatalf("failed to load suite: %v", err)
	}

	categories := make(map[string]int)
	for _, test := range suite.Tests {
		categories[test.Category]++
	}

	if categories["create"] == 0 {
		t.Error("expected create category tests")
	}
	if categories["update"] == 0 {
		t.Error("expected update category tests")
	}
	if categories["delete"] == 0 {
		t.Error("expected delete category tests")
	}

	t.Logf("Argument categories: create=%d, update=%d, delete=%d",
		categories["create"], categories["update"], categories["delete"])
}
