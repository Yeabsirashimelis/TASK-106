#!/usr/bin/env bash
# =============================================================================
# Unit Test Runner
# Executes all Go unit tests in the unit_tests/ directory
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "=============================================="
echo "  UNIT TESTS"
echo "=============================================="
echo ""

# Run unit tests with verbose output and count results
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting unit tests..."
echo ""

if go test -v -count=1 ./unit_tests/... 2>&1 | tee /tmp/unit_test_output.txt; then
    UNIT_EXIT=0
else
    UNIT_EXIT=1
fi

echo ""
echo "=============================================="
echo "  UNIT TEST SUMMARY"
echo "=============================================="

TOTAL=$(grep -c "^--- " /tmp/unit_test_output.txt 2>/dev/null || echo "0")
PASSED=$(grep -c "^--- PASS" /tmp/unit_test_output.txt 2>/dev/null || echo "0")
FAILED=$(grep -c "^--- FAIL" /tmp/unit_test_output.txt 2>/dev/null || echo "0")

echo "  Total:  $TOTAL"
echo "  Passed: $PASSED"
echo "  Failed: $FAILED"
echo "=============================================="

if [ "$UNIT_EXIT" -ne 0 ]; then
    echo "  STATUS: FAILED"
    echo "=============================================="
    exit 1
else
    echo "  STATUS: ALL PASSED"
    echo "=============================================="
    exit 0
fi
