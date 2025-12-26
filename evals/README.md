# ProductPlan MCP Evaluations

This directory contains evaluation test suites for validating LLM tool selection accuracy with the ProductPlan MCP server.

## Test Suites

### 1. Tool Selection (`tool_selection.json`)
Tests that the correct tool is selected for various natural language prompts.

- **69 tests** covering all ProductPlan operations
- Categories: roadmaps, bars, lanes, milestones, objectives, ideas, opportunities, launches, status, create, update, delete
- Difficulty levels: easy, medium, hard

### 2. Confusion Pairs (`confusion_pairs.json`)
Tests for distinguishing between commonly confused tools.

- **10 tool pairs** with disambiguation guidance
- **40 tests** for subtle distinctions
- Examples:
  - `list_roadmaps` vs `get_roadmap`
  - `get_roadmap` vs `get_roadmap_bars`
  - `get_bar_comments` vs `manage_bar_comment`
  - `list_ideas` vs `list_opportunities`

### 3. Argument Correctness (`argument_correctness.json`)
Tests that arguments are correctly extracted from natural language.

- **25 tests** for argument extraction
- Validates required args, expected values
- Covers IDs, dates, names, actions, and nested parameters

## Running Tests

```bash
# Run all tests
go test ./evals/...

# Run with verbose output
go test ./evals/... -v

# Run specific test
go test ./evals/... -run TestLoadToolSelectionSuite
```

## Using the Framework

```go
package main

import (
    "fmt"
    "github.com/olgasafonova/productplan-mcp-server/evals"
)

func main() {
    // Load all test suites
    toolSel, confPairs, args, err := evals.LoadAllEvals("evals/")
    if err != nil {
        panic(err)
    }

    // Create your LLM-based selector
    selector := NewMyLLMSelector()

    // Run evaluations
    metrics, _ := evals.EvaluateToolSelection(toolSel, selector)
    fmt.Println(evals.FormatMetrics(metrics, "Tool Selection"))
}
```

## Implementing a Selector

```go
type ToolSelector interface {
    SelectTool(prompt string) (toolName string, args map[string]interface{}, err error)
}
```

Your selector should:
1. Take a natural language prompt
2. Return the tool name to use
3. Return the arguments to pass
4. Return any errors

## Test Coverage

| Category | Tests | Description |
|----------|-------|-------------|
| Roadmaps | 8 | Roadmap listing and details |
| Bars | 15 | Bar CRUD, children, comments, connections, links |
| Lanes | 3 | Lane listing and management |
| Milestones | 3 | Milestone listing and management |
| Objectives | 9 | OKR objectives and key results |
| Ideas | 10 | Ideas, customers, tags |
| Opportunities | 4 | Opportunity pipeline |
| Launches | 3 | Product launches |
| Status | 2 | API status check |
| Create | 14 | Creating various entities |
| Update | 8 | Updating entities |
| Delete | 5 | Deleting entities |

## Adding New Tests

1. Add test cases to the appropriate JSON file
2. Follow the existing format:
   - `id`: Unique test identifier
   - `prompt`: Natural language input
   - `expected_tool`: Correct tool name
   - `category`: Logical grouping
   - `difficulty`: easy/medium/hard

3. Run tests to validate JSON syntax:
   ```bash
   go test ./evals/... -run TestLoad
   ```
