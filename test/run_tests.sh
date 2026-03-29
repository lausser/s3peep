#!/bin/bash

# Central test runner - discovers and executes all test_*.sh files

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

# Results tracking
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=""

log_info "Starting s3peep test suite..."

SELF_PATH="${SCRIPT_DIR}/$(basename "$0")"

# Ignore wrapper-injected runner path argument when this script is already the entrypoint
if [ "$1" = "$0" ] || [ "$1" = "$SELF_PATH" ] || [ "$(basename "$1")" = "run_tests.sh" ]; then
    shift
fi

if [ "$1" = "--" ]; then
    shift
fi

# Run mandatory preflight before any feature work
preflight_check

if [ "$1" = "--preflight-only" ]; then
    log_info "Preflight-only mode succeeded"
    exit 0
fi

# Setup default fixtures
log_info "Setting up default fixtures..."
"${SCRIPT_DIR}/fixtures/setup.sh" default

# Generate s3peep config
log_info "Generating s3peep config..."
generate_s3peep_config

# Discover and run tests
if [ $# -gt 0 ]; then
    # Run specific test files
    TEST_FILES=()
    for arg in "$@"; do
        if [ "$arg" = "--preflight-only" ]; then
            continue
        fi
        if [ "$arg" = "$0" ] || [ "$arg" = "$SELF_PATH" ] || [ "$(basename "$arg")" = "run_tests.sh" ]; then
            continue
        fi
        if [ -f "$arg" ]; then
            TEST_FILES+=("$arg")
        elif [ -f "${SCRIPT_DIR}/$arg" ]; then
            TEST_FILES+=("${SCRIPT_DIR}/$arg")
        else
            TEST_FILES+=("$arg")
        fi
    done
    if [ ${#TEST_FILES[@]} -eq 0 ]; then
        log_info "No test files selected after filtering runner arguments"
        exit 0
    fi
else
    # Discover all test_*.sh files
    mapfile -t TEST_FILES < <(find "${SCRIPT_DIR}" -maxdepth 1 -name "test_*.sh" -type f | sort)
fi

for test_file in "${TEST_FILES[@]}"; do
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    test_name=$(basename "$test_file")
    
    if run_test "$test_file"; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS="${FAILED_TESTS} ${test_name}"
    fi
done

# Summary
echo ""
log_info "========================================="
log_info "Test Results: ${TESTS_PASSED}/${TESTS_TOTAL} passed"

if [ $TESTS_FAILED -gt 0 ]; then
    log_fail "Failed tests:${FAILED_TESTS}"
    log_info "========================================="
    exit 1
else
    log_info "All tests passed!"
    log_info "========================================="
    exit 0
fi
