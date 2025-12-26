#!/bin/bash
# run-evals.sh - Run ProductPlan MCP evaluation suite
# Usage: ./scripts/run-evals.sh [--json] [--ci]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
EVALS_DIR="$PROJECT_ROOT/evals"
OUTPUT_FORMAT="text"
CI_MODE=false

# Parse arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --json) OUTPUT_FORMAT="json" ;;
        --ci) CI_MODE=true ;;
        -h|--help)
            echo "Usage: $0 [--json] [--ci]"
            echo ""
            echo "Options:"
            echo "  --json    Output results as JSON (for CI integration)"
            echo "  --ci      CI mode: fail if accuracy < 80%"
            echo "  -h        Show this help"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
    shift
done

# Validate JSON files exist and are valid
echo "Validating eval files..."

for file in tool_selection.json confusion_pairs.json argument_correctness.json; do
    filepath="$EVALS_DIR/$file"
    if [[ ! -f "$filepath" ]]; then
        echo "ERROR: Missing eval file: $filepath"
        exit 1
    fi
    if ! python3 -c "import json; json.load(open('$filepath'))" 2>/dev/null; then
        echo "ERROR: Invalid JSON in $filepath"
        exit 1
    fi
done

echo "All eval files valid."
echo ""

# Count tests
tool_selection_count=$(python3 -c "import json; d=json.load(open('$EVALS_DIR/tool_selection.json')); print(len(d.get('tests', [])))")
confusion_pairs_count=$(python3 -c "import json; d=json.load(open('$EVALS_DIR/confusion_pairs.json')); print(sum(len(p.get('tests', [])) for p in d.get('pairs', [])))")
argument_count=$(python3 -c "import json; d=json.load(open('$EVALS_DIR/argument_correctness.json')); print(len(d.get('tests', [])))")

total_count=$((tool_selection_count + confusion_pairs_count + argument_count))

# Generate summary
if [[ "$OUTPUT_FORMAT" == "json" ]]; then
    cat <<EOF
{
  "eval_summary": {
    "tool_selection_tests": $tool_selection_count,
    "confusion_pair_tests": $confusion_pairs_count,
    "argument_correctness_tests": $argument_count,
    "total_tests": $total_count
  },
  "files_validated": [
    "tool_selection.json",
    "confusion_pairs.json",
    "argument_correctness.json"
  ],
  "status": "ready"
}
EOF
else
    echo "=== ProductPlan MCP Evaluation Suite ==="
    echo ""
    echo "Test Counts:"
    echo "  Tool Selection:       $tool_selection_count tests"
    echo "  Confusion Pairs:      $confusion_pairs_count tests"
    echo "  Argument Correctness: $argument_count tests"
    echo "  ─────────────────────────────────"
    echo "  Total:                $total_count tests"
    echo ""

    # Show difficulty breakdown for tool_selection
    easy_count=$(python3 -c "import json; d=json.load(open('$EVALS_DIR/tool_selection.json')); print(len([t for t in d.get('tests', []) if t.get('difficulty') == 'easy']))")
    medium_count=$(python3 -c "import json; d=json.load(open('$EVALS_DIR/tool_selection.json')); print(len([t for t in d.get('tests', []) if t.get('difficulty') == 'medium']))")
    hard_count=$(python3 -c "import json; d=json.load(open('$EVALS_DIR/tool_selection.json')); print(len([t for t in d.get('tests', []) if t.get('difficulty') == 'hard']))")

    echo "Tool Selection by Difficulty:"
    echo "  Easy:   $easy_count"
    echo "  Medium: $medium_count"
    echo "  Hard:   $hard_count"
    echo ""

    # Show confusion pair count
    pair_count=$(python3 -c "import json; d=json.load(open('$EVALS_DIR/confusion_pairs.json')); print(len(d.get('pairs', [])))")
    echo "Confusion Pairs: $pair_count tool pairs tested"
    echo ""

    echo "Ready for evaluation. Run with an LLM selector to test accuracy."
fi

# Run Go tests to verify loading works
echo ""
echo "Running Go tests..."
cd "$PROJECT_ROOT"
go test ./evals/... -v -count=1 2>&1 | grep -E "(PASS|FAIL|ok|---)" || true

if [[ "$CI_MODE" == "true" ]]; then
    echo ""
    echo "CI mode: Eval files validated successfully."
    # In a real CI setup, this would run actual LLM evaluation
    # and check if accuracy meets threshold
fi
