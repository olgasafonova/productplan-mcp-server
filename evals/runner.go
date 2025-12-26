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

// ToolSelectionSuite contains tests for verifying correct tool selection.
type ToolSelectionSuite struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Version     string              `json:"version"`
	Tests       []ToolSelectionTest `json:"tests"`
}

// ToolSelectionTest represents a single tool selection test case.
type ToolSelectionTest struct {
	ID           string `json:"id"`
	Prompt       string `json:"prompt"`
	ExpectedTool string `json:"expected_tool"`
	Category     string `json:"category"`
	Difficulty   string `json:"difficulty"`
}

// ConfusionPairSuite contains tests for distinguishing commonly confused tools.
type ConfusionPairSuite struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	Pairs       []ConfusionPair `json:"pairs"`
}

// ConfusionPair represents a pair of commonly confused tools.
type ConfusionPair struct {
	Tools       []string            `json:"tools"`
	Distinction string              `json:"distinction"`
	Tests       []ConfusionPairTest `json:"tests"`
}

// ConfusionPairTest represents a test case for confusion pairs.
type ConfusionPairTest struct {
	Prompt       string `json:"prompt"`
	ExpectedTool string `json:"expected_tool"`
	Rationale    string `json:"rationale"`
}

// ArgumentSuite contains tests for verifying argument extraction.
type ArgumentSuite struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Tests       []ArgumentTest `json:"tests"`
}

// ArgumentTest represents a test case for argument extraction.
type ArgumentTest struct {
	ID           string                 `json:"id"`
	Tool         string                 `json:"tool"`
	Prompt       string                 `json:"prompt"`
	ExpectedArgs map[string]interface{} `json:"expected_args"`
	RequiredArgs []string               `json:"required_args"`
	Category     string                 `json:"category"`
}

// ToolSelector is the interface that LLM-based selectors must implement.
type ToolSelector interface {
	SelectTool(prompt string) (toolName string, args map[string]interface{}, err error)
}

// EvalMetrics contains evaluation results and statistics.
type EvalMetrics struct {
	TotalTests    int                         `json:"total_tests"`
	PassedTests   int                         `json:"passed_tests"`
	FailedTests   int                         `json:"failed_tests"`
	Accuracy      float64                     `json:"accuracy"`
	ByCategory    map[string]*CategoryMetrics `json:"by_category"`
	FailedDetails []string                    `json:"failed_details"`
}

// CategoryMetrics contains per-category evaluation results.
type CategoryMetrics struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// ToolSelectionResult contains the result of a single tool selection test.
type ToolSelectionResult struct {
	TestID       string `json:"test_id"`
	Prompt       string `json:"prompt"`
	Expected     string `json:"expected"`
	Actual       string `json:"actual"`
	Passed       bool   `json:"passed"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// ConfusionPairResult contains the result of a confusion pair test.
type ConfusionPairResult struct {
	Tools        []string `json:"tools"`
	Prompt       string   `json:"prompt"`
	Expected     string   `json:"expected"`
	Actual       string   `json:"actual"`
	Passed       bool     `json:"passed"`
	ErrorMessage string   `json:"error_message,omitempty"`
}

// ArgumentResult contains the result of an argument extraction test.
type ArgumentResult struct {
	TestID      string   `json:"test_id"`
	Tool        string   `json:"tool"`
	Passed      bool     `json:"passed"`
	MissingArgs []string `json:"missing_args,omitempty"`
	WrongArgs   []string `json:"wrong_args,omitempty"`
}

// LoadToolSelectionSuite loads a tool selection test suite from a JSON file.
func LoadToolSelectionSuite(path string) (*ToolSelectionSuite, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is controlled by eval framework
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var suite ToolSelectionSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &suite, nil
}

// LoadConfusionPairSuite loads a confusion pair test suite from a JSON file.
func LoadConfusionPairSuite(path string) (*ConfusionPairSuite, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is controlled by eval framework
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var suite ConfusionPairSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &suite, nil
}

// LoadArgumentSuite loads an argument test suite from a JSON file.
func LoadArgumentSuite(path string) (*ArgumentSuite, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is controlled by eval framework
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var suite ArgumentSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
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
			result := ConfusionPairResult{
				Tools:    pair.Tools,
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
			metrics.ByCategory[pairKey].Total++

			if result.Passed {
				metrics.PassedTests++
				metrics.ByCategory[pairKey].Passed++
			} else {
				metrics.FailedTests++
				metrics.ByCategory[pairKey].Failed++
				metrics.FailedDetails = append(metrics.FailedDetails,
					fmt.Sprintf("[%s] %s: expected %s, got %s", pairKey, test.Prompt, test.ExpectedTool, result.Actual))
			}
		}
	}

	if metrics.TotalTests > 0 {
		metrics.Accuracy = float64(metrics.PassedTests) / float64(metrics.TotalTests)
	}

	return metrics, results
}

// EvaluateArguments runs argument extraction tests and returns metrics.
func EvaluateArguments(suite *ArgumentSuite, selector ToolSelector) (*EvalMetrics, []ArgumentResult) {
	metrics := &EvalMetrics{
		ByCategory: make(map[string]*CategoryMetrics),
	}
	var results []ArgumentResult

	for _, test := range suite.Tests {
		result := ArgumentResult{
			TestID: test.ID,
			Tool:   test.Tool,
		}

		_, args, err := selector.SelectTool(test.Prompt)
		if err != nil {
			result.Passed = false
			result.MissingArgs = test.RequiredArgs
		} else {
			// Check required args
			for _, reqArg := range test.RequiredArgs {
				if _, ok := args[reqArg]; !ok {
					result.MissingArgs = append(result.MissingArgs, reqArg)
				}
			}

			// Check expected values
			for key, expectedVal := range test.ExpectedArgs {
				if actualVal, ok := args[key]; ok {
					if !compareValues(expectedVal, actualVal) {
						result.WrongArgs = append(result.WrongArgs,
							fmt.Sprintf("%s: expected %v, got %v", key, expectedVal, actualVal))
					}
				}
			}

			result.Passed = len(result.MissingArgs) == 0 && len(result.WrongArgs) == 0
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
			details := fmt.Sprintf("[%s] %s", test.ID, test.Prompt)
			if len(result.MissingArgs) > 0 {
				details += fmt.Sprintf(" missing: %v", result.MissingArgs)
			}
			if len(result.WrongArgs) > 0 {
				details += fmt.Sprintf(" wrong: %v", result.WrongArgs)
			}
			metrics.FailedDetails = append(metrics.FailedDetails, details)
		}
	}

	if metrics.TotalTests > 0 {
		metrics.Accuracy = float64(metrics.PassedTests) / float64(metrics.TotalTests)
	}

	return metrics, results
}

// compareValues compares two values for equality, handling type differences.
func compareValues(expected, actual interface{}) bool {
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}

	// Handle numeric comparisons (JSON numbers are float64)
	switch e := expected.(type) {
	case int:
		if a, ok := actual.(float64); ok {
			return float64(e) == a
		}
	case float64:
		if a, ok := actual.(float64); ok {
			return e == a
		}
	}

	// Use reflect for deep equality
	return reflect.DeepEqual(expected, actual)
}

// FormatMetrics returns a formatted string representation of metrics.
func FormatMetrics(metrics *EvalMetrics, suiteName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n=== %s Results ===\n", suiteName))
	sb.WriteString(fmt.Sprintf("Total: %d | Passed: %d | Failed: %d\n",
		metrics.TotalTests, metrics.PassedTests, metrics.FailedTests))
	sb.WriteString(fmt.Sprintf("Accuracy: %.1f%%\n", metrics.Accuracy*100))

	if len(metrics.ByCategory) > 0 {
		sb.WriteString("\nBy Category:\n")
		for cat, catMetrics := range metrics.ByCategory {
			accuracy := float64(0)
			if catMetrics.Total > 0 {
				accuracy = float64(catMetrics.Passed) / float64(catMetrics.Total) * 100
			}
			sb.WriteString(fmt.Sprintf("  %s: %d/%d (%.1f%%)\n",
				cat, catMetrics.Passed, catMetrics.Total, accuracy))
		}
	}

	if len(metrics.FailedDetails) > 0 {
		sb.WriteString("\nFailed Tests:\n")
		for _, detail := range metrics.FailedDetails {
			sb.WriteString(fmt.Sprintf("  - %s\n", detail))
		}
	}

	return sb.String()
}
