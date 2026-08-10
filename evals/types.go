package evals

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

// CombinedReport holds results from all evaluation suites.
type CombinedReport struct {
	Timestamp       string                    `json:"timestamp"`
	ToolSelection   *EvalMetrics              `json:"tool_selection"`
	ConfusionPairs  *EvalMetrics              `json:"confusion_pairs"`
	Arguments       *EvalMetrics              `json:"arguments"`
	OverallAccuracy float64                   `json:"overall_accuracy"`
	PassThreshold   bool                      `json:"pass_threshold"`
	Summary         map[string]*CategoryTotal `json:"summary"`
}

// CategoryTotal aggregates totals across all suites.
type CategoryTotal struct {
	Total  int     `json:"total"`
	Passed int     `json:"passed"`
	Rate   float64 `json:"rate"`
}
