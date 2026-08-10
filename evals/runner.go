// Package evals provides evaluation framework for testing MCP tool selection accuracy.
package evals

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// loadJSON reads a file at path and decodes it into out. Returns wrapped
// errors so the caller can attach context-specific messages.
func loadJSON(path string, out any) error {
	data, err := os.ReadFile(path) // #nosec G304 -- path is controlled by eval framework
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	return nil
}

// LoadToolSelectionSuite loads a tool selection test suite from a JSON file.
func LoadToolSelectionSuite(path string) (*ToolSelectionSuite, error) {
	var suite ToolSelectionSuite
	if err := loadJSON(path, &suite); err != nil {
		return nil, err
	}
	return &suite, nil
}

// LoadConfusionPairSuite loads a confusion pair test suite from a JSON file.
func LoadConfusionPairSuite(path string) (*ConfusionPairSuite, error) {
	var suite ConfusionPairSuite
	if err := loadJSON(path, &suite); err != nil {
		return nil, err
	}
	return &suite, nil
}

// LoadArgumentSuite loads an argument test suite from a JSON file.
func LoadArgumentSuite(path string) (*ArgumentSuite, error) {
	var suite ArgumentSuite
	if err := loadJSON(path, &suite); err != nil {
		return nil, err
	}
	return &suite, nil
}

// LoadAllEvals loads all evaluation suites from a directory.
func LoadAllEvals(dir string) (*ToolSelectionSuite, *ConfusionPairSuite, *ArgumentSuite, error) {
	toolSelection, err := LoadToolSelectionSuite(filepath.Join(dir, "tool_selection.json"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("tool_selection: %w", err)
	}

	confusionPairs, err := LoadConfusionPairSuite(filepath.Join(dir, "confusion_pairs.json"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("confusion_pairs: %w", err)
	}

	arguments, err := LoadArgumentSuite(filepath.Join(dir, "argument_correctness.json"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("argument_correctness: %w", err)
	}

	return toolSelection, confusionPairs, arguments, nil
}

// EvaluateToolSelection runs tool selection tests and returns metrics.
func EvaluateToolSelection(suite *ToolSelectionSuite, selector ToolSelector) (*EvalMetrics, []ToolSelectionResult) {
	metrics := &EvalMetrics{
		ByCategory: make(map[string]*CategoryMetrics),
	}
	var results []ToolSelectionResult

	for _, test := range suite.Tests {
		result := ToolSelectionResult{
			TestID:   test.ID,
			Prompt:   test.Prompt,
			Expected: test.ExpectedTool,
		}

		toolName, _, err := selector.SelectTool(test.Prompt)
		if err != nil {
			result.ErrorMessage = err.Error()
			result.Passed = false
		} else {
			result.Actual = toolName
			result.Passed = toolName == test.ExpectedTool
		}

		results = append(results, result)
		metrics.TotalTests++

		// Update category metrics
		if _, ok := metrics.ByCategory[test.Category]; !ok {
			metrics.ByCategory[test.Category] = &CategoryMetrics{}
		}
		metrics.ByCategory[test.Category].Total++

		if result.Passed {
			metrics.PassedTests++
			metrics.ByCategory[test.Category].Passed++
		} else {
			metrics.FailedTests++
			metrics.ByCategory[test.Category].Failed++
			metrics.FailedDetails = append(metrics.FailedDetails,
				fmt.Sprintf("[%s] %s: expected %s, got %s", test.ID, test.Prompt, test.ExpectedTool, result.Actual))
		}
	}

	if metrics.TotalTests > 0 {
		metrics.Accuracy = float64(metrics.PassedTests) / float64(metrics.TotalTests)
	}

	return metrics, results
}

// EvaluateConfusionPairs runs confusion pair tests and returns metrics.
// runConfusionPairTest executes one prompt/expected-tool case against the selector.
func runConfusionPairTest(pair ConfusionPair, test ConfusionPairTest, selector ToolSelector) ConfusionPairResult {
	result := ConfusionPairResult{
		Tools:    pair.Tools,
		Prompt:   test.Prompt,
		Expected: test.ExpectedTool,
	}
	toolName, _, err := selector.SelectTool(test.Prompt)
	if err != nil {
		result.ErrorMessage = err.Error()
		result.Passed = false
		return result
	}
	result.Actual = toolName
	result.Passed = toolName == test.ExpectedTool
	return result
}

// recordConfusionPairResult updates the per-pair metrics aggregate.
func recordConfusionPairResult(metrics *EvalMetrics, pairKey string, test ConfusionPairTest, result ConfusionPairResult) {
	metrics.TotalTests++
	metrics.ByCategory[pairKey].Total++
	if result.Passed {
		metrics.PassedTests++
		metrics.ByCategory[pairKey].Passed++
		return
	}
	metrics.FailedTests++
	metrics.ByCategory[pairKey].Failed++
	metrics.FailedDetails = append(metrics.FailedDetails,
		fmt.Sprintf("[%s] %s: expected %s, got %s", pairKey, test.Prompt, test.ExpectedTool, result.Actual))
}

func EvaluateConfusionPairs(suite *ConfusionPairSuite, selector ToolSelector) (*EvalMetrics, []ConfusionPairResult) {
	metrics := &EvalMetrics{
		ByCategory: make(map[string]*CategoryMetrics),
	}
	var results []ConfusionPairResult

	for _, pair := range suite.Pairs {
		pairKey := strings.Join(pair.Tools, "_vs_")
		if _, ok := metrics.ByCategory[pairKey]; !ok {
			metrics.ByCategory[pairKey] = &CategoryMetrics{}
		}

		for _, test := range pair.Tests {
			result := runConfusionPairTest(pair, test, selector)
			results = append(results, result)
			recordConfusionPairResult(metrics, pairKey, test, result)
		}
	}

	if metrics.TotalTests > 0 {
		metrics.Accuracy = float64(metrics.PassedTests) / float64(metrics.TotalTests)
	}

	return metrics, results
}

// runArgumentTest executes a single argument-extraction test and returns its result.
func runArgumentTest(test ArgumentTest, selector ToolSelector) ArgumentResult {
	result := ArgumentResult{
		TestID: test.ID,
		Tool:   test.Tool,
	}

	_, args, err := selector.SelectTool(test.Prompt)
	if err != nil {
		result.Passed = false
		result.MissingArgs = test.RequiredArgs
		return result
	}

	result.MissingArgs = findMissingArgs(test.RequiredArgs, args)
	result.WrongArgs = findWrongArgs(test.ExpectedArgs, args)
	result.Passed = len(result.MissingArgs) == 0 && len(result.WrongArgs) == 0
	return result
}

// findMissingArgs returns required arg names that are absent from args.
func findMissingArgs(required []string, args map[string]any) []string {
	var missing []string
	for _, reqArg := range required {
		if _, ok := args[reqArg]; !ok {
			missing = append(missing, reqArg)
		}
	}
	return missing
}

// findWrongArgs returns "key: expected X, got Y" descriptions for every arg
// where args[key] is present but does not compare equal to the expected value.
func findWrongArgs(expected map[string]any, args map[string]any) []string {
	var wrong []string
	for key, expectedVal := range expected {
		actualVal, ok := args[key]
		if !ok {
			continue
		}
		if !compareValues(expectedVal, actualVal) {
			wrong = append(wrong, fmt.Sprintf("%s: expected %v, got %v", key, expectedVal, actualVal))
		}
	}
	return wrong
}

// recordArgumentResult updates the metrics aggregate for a single test outcome.
func recordArgumentResult(metrics *EvalMetrics, test ArgumentTest, result ArgumentResult) {
	metrics.TotalTests++
	if _, ok := metrics.ByCategory[test.Category]; !ok {
		metrics.ByCategory[test.Category] = &CategoryMetrics{}
	}
	cat := metrics.ByCategory[test.Category]
	cat.Total++

	if result.Passed {
		metrics.PassedTests++
		cat.Passed++
		return
	}
	metrics.FailedTests++
	cat.Failed++
	metrics.FailedDetails = append(metrics.FailedDetails, formatArgumentFailure(test, result))
}

// formatArgumentFailure renders a one-line failure summary for the metrics log.
func formatArgumentFailure(test ArgumentTest, result ArgumentResult) string {
	details := fmt.Sprintf("[%s] %s", test.ID, test.Prompt)
	if len(result.MissingArgs) > 0 {
		details += fmt.Sprintf(" missing: %v", result.MissingArgs)
	}
	if len(result.WrongArgs) > 0 {
		details += fmt.Sprintf(" wrong: %v", result.WrongArgs)
	}
	return details
}

// EvaluateArguments runs argument extraction tests and returns metrics.
func EvaluateArguments(suite *ArgumentSuite, selector ToolSelector) (*EvalMetrics, []ArgumentResult) {
	metrics := &EvalMetrics{
		ByCategory: make(map[string]*CategoryMetrics),
	}
	var results []ArgumentResult

	for _, test := range suite.Tests {
		result := runArgumentTest(test, selector)
		results = append(results, result)
		recordArgumentResult(metrics, test, result)
	}

	if metrics.TotalTests > 0 {
		metrics.Accuracy = float64(metrics.PassedTests) / float64(metrics.TotalTests)
	}

	return metrics, results
}

// asFloat normalises the numeric types JSON comparisons produce (int from
// fixtures, float64 from decoded JSON) to float64.
func asFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}

// compareValues compares two values for equality, handling type differences.
func compareValues(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == nil && actual == nil
	}

	// Handle numeric comparisons (JSON numbers are float64)
	if e, ok := asFloat(expected); ok {
		if a, ok := actual.(float64); ok {
			return e == a
		}
	}

	// Use reflect for deep equality
	return reflect.DeepEqual(expected, actual)
}

// writeCategoryBreakdown appends the "By Category" block to sb. No-op for empty input.
func writeCategoryBreakdown(sb *strings.Builder, byCategory map[string]*CategoryMetrics) {
	if len(byCategory) == 0 {
		return
	}
	sb.WriteString("\nBy Category:\n")
	for cat, catMetrics := range byCategory {
		accuracy := float64(0)
		if catMetrics.Total > 0 {
			accuracy = float64(catMetrics.Passed) / float64(catMetrics.Total) * 100
		}
		fmt.Fprintf(sb, "  %s: %d/%d (%.1f%%)\n", cat, catMetrics.Passed, catMetrics.Total, accuracy)
	}
}

// writeFailedDetails appends the "Failed Tests" block to sb. No-op for empty input.
func writeFailedDetails(sb *strings.Builder, failedDetails []string) {
	if len(failedDetails) == 0 {
		return
	}
	sb.WriteString("\nFailed Tests:\n")
	for _, detail := range failedDetails {
		fmt.Fprintf(sb, "  - %s\n", detail)
	}
}

// FormatMetrics returns a formatted string representation of metrics.
func FormatMetrics(metrics *EvalMetrics, suiteName string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "\n=== %s Results ===\n", suiteName)
	fmt.Fprintf(&sb, "Total: %d | Passed: %d | Failed: %d\n",
		metrics.TotalTests, metrics.PassedTests, metrics.FailedTests)
	fmt.Fprintf(&sb, "Accuracy: %.1f%%\n", metrics.Accuracy*100)

	writeCategoryBreakdown(&sb, metrics.ByCategory)
	writeFailedDetails(&sb, metrics.FailedDetails)

	return sb.String()
}

// ExportMetricsJSON exports metrics to JSON for CI integration.
func ExportMetricsJSON(metrics *EvalMetrics, suiteName string) ([]byte, error) {
	report := map[string]any{
		"suite":        suiteName,
		"total_tests":  metrics.TotalTests,
		"passed_tests": metrics.PassedTests,
		"failed_tests": metrics.FailedTests,
		"accuracy":     metrics.Accuracy,
		"by_category":  metrics.ByCategory,
	}

	if len(metrics.FailedDetails) > 0 {
		report["failed_details"] = metrics.FailedDetails
	}

	return json.MarshalIndent(report, "", "  ")
}

// GenerateCombinedReport creates a unified report from all evaluation results.
func GenerateCombinedReport(
	toolSelection *EvalMetrics,
	confusionPairs *EvalMetrics,
	arguments *EvalMetrics,
	threshold float64,
) *CombinedReport {
	totalTests := toolSelection.TotalTests + confusionPairs.TotalTests + arguments.TotalTests
	totalPassed := toolSelection.PassedTests + confusionPairs.PassedTests + arguments.PassedTests

	overallAccuracy := float64(0)
	if totalTests > 0 {
		overallAccuracy = float64(totalPassed) / float64(totalTests)
	}

	return &CombinedReport{
		Timestamp:       fmt.Sprintf("%d", currentTimestamp()),
		ToolSelection:   toolSelection,
		ConfusionPairs:  confusionPairs,
		Arguments:       arguments,
		OverallAccuracy: overallAccuracy,
		PassThreshold:   overallAccuracy >= threshold,
		Summary: map[string]*CategoryTotal{
			"tool_selection": {
				Total:  toolSelection.TotalTests,
				Passed: toolSelection.PassedTests,
				Rate:   toolSelection.Accuracy,
			},
			"confusion_pairs": {
				Total:  confusionPairs.TotalTests,
				Passed: confusionPairs.PassedTests,
				Rate:   confusionPairs.Accuracy,
			},
			"arguments": {
				Total:  arguments.TotalTests,
				Passed: arguments.PassedTests,
				Rate:   arguments.Accuracy,
			},
		},
	}
}

// ExportCombinedReportJSON exports the combined report to JSON.
func ExportCombinedReportJSON(report *CombinedReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// currentTimestamp returns Unix timestamp (allows testing override).
var currentTimestamp = func() int64 {
	return 0 // Would use time.Now().Unix() in production
}
