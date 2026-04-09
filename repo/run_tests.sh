#!/usr/bin/env bash
# =============================================================================
# Unified Test Runner
# Runs all unit tests and API functional tests with summary output
# Usage: ./run_tests.sh [--unit-only | --api-only | --no-docker]
# =============================================================================
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

RUN_UNIT=true
RUN_API=true
MANAGE_DOCKER=true
UNIT_EXIT=0
API_EXIT=0

for arg in "$@"; do
    case "$arg" in
        --unit-only)  RUN_API=false ;;
        --api-only)   RUN_UNIT=false ;;
        --no-docker)  MANAGE_DOCKER=false ;;
    esac
done

echo "============================================================"
echo "  COLLEGIATE ATHLETICS & LEARNING OPS - TEST SUITE"
echo "  $(date '+%Y-%m-%d %H:%M:%S')"
echo "============================================================"
echo ""

# ─────────────────────────────────────────────────────────────────
# PHASE 1: UNIT TESTS (no Docker needed)
# ─────────────────────────────────────────────────────────────────
if [ "$RUN_UNIT" = true ]; then
    echo "============================================================"
    echo "  PHASE 1: UNIT TESTS"
    echo "============================================================"
    echo ""

    if bash unit_tests/run_unit_tests.sh; then
        UNIT_EXIT=0
    else
        UNIT_EXIT=1
    fi
    echo ""
fi

# ─────────────────────────────────────────────────────────────────
# PHASE 2: API FUNCTIONAL TESTS (requires Docker)
# ─────────────────────────────────────────────────────────────────
if [ "$RUN_API" = true ]; then
    echo "============================================================"
    echo "  PHASE 2: API FUNCTIONAL TESTS"
    echo "============================================================"
    echo ""

    # Check prerequisites
    if ! command -v curl &> /dev/null; then
        echo "ERROR: curl is required for API tests"
        API_EXIT=1
    elif ! command -v jq &> /dev/null; then
        echo "ERROR: jq is required for API tests"
        API_EXIT=1
    else
        # Start Docker services if managing docker
        if [ "$MANAGE_DOCKER" = true ]; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting Docker services..."

            # Determine docker compose command
            if command -v docker &> /dev/null && docker compose version &> /dev/null; then
                DC="docker compose"
            elif command -v docker-compose &> /dev/null; then
                DC="docker-compose"
            else
                echo "ERROR: docker compose is required for API tests"
                API_EXIT=1
            fi

            if [ "$API_EXIT" -eq 0 ]; then
                # Build and start
                $DC build --quiet 2>&1 | tail -5
                $DC up -d 2>&1

                echo "[$(date '+%Y-%m-%d %H:%M:%S')] Waiting for services to be healthy..."
                sleep 5

                # Run API tests
                if bash API_tests/run_api_tests.sh; then
                    API_EXIT=0
                else
                    API_EXIT=1
                fi

                # Cleanup
                echo ""
                echo "[$(date '+%Y-%m-%d %H:%M:%S')] Stopping Docker services..."
                $DC down -v --remove-orphans 2>&1 | tail -3
            fi
        else
            # --no-docker: assume services are already running
            if bash API_tests/run_api_tests.sh; then
                API_EXIT=0
            else
                API_EXIT=1
            fi
        fi
    fi
    echo ""
fi

# ─────────────────────────────────────────────────────────────────
# FINAL SUMMARY
# ─────────────────────────────────────────────────────────────────
echo ""
echo "============================================================"
echo "  FINAL TEST SUMMARY"
echo "============================================================"

if [ "$RUN_UNIT" = true ]; then
    if [ "$UNIT_EXIT" -eq 0 ]; then
        echo "  Unit Tests:     PASSED"
    else
        echo "  Unit Tests:     FAILED"
    fi
fi

if [ "$RUN_API" = true ]; then
    if [ "$API_EXIT" -eq 0 ]; then
        echo "  API Tests:      PASSED"
    else
        echo "  API Tests:      FAILED"
    fi
fi

OVERALL_EXIT=$((UNIT_EXIT + API_EXIT))
echo ""
if [ "$OVERALL_EXIT" -eq 0 ]; then
    echo "  OVERALL STATUS: ALL PASSED"
else
    echo "  OVERALL STATUS: FAILED"
fi
echo "============================================================"

exit $OVERALL_EXIT
