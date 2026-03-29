#!/bin/bash

# Fixture setup script - populates MinIO with test data using mc
# Usage: ./setup.sh [scenario]
# Scenarios: default, empty, nested, largefiles

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# Default scenario
SCENARIO="${1:-default}"

log_info "Setting up fixtures: ${SCENARIO}"

# Wait for MinIO
wait_for_minio

# Configure mc
configure_mc

# Create test bucket (delete if exists)
mc rb --force local/test-bucket 2>/dev/null || true
mc mb local/test-bucket 2>/dev/null || true

# Run scenario-specific setup
case "$SCENARIO" in
    default)
        if [ -f "${SCRIPT_DIR}/scenarios/default.sh" ]; then
            source "${SCRIPT_DIR}/scenarios/default.sh"
        fi
        ;;
    empty)
        if [ -f "${SCRIPT_DIR}/scenarios/empty.sh" ]; then
            source "${SCRIPT_DIR}/scenarios/empty.sh"
        fi
        ;;
    nested)
        if [ -f "${SCRIPT_DIR}/scenarios/nested.sh" ]; then
            source "${SCRIPT_DIR}/scenarios/nested.sh"
        fi
        ;;
    largefiles)
        if [ -f "${SCRIPT_DIR}/scenarios/largefiles.sh" ]; then
            source "${SCRIPT_DIR}/scenarios/largefiles.sh"
        fi
        ;;
    *)
        log_error "Unknown scenario: ${SCENARIO}"
        log_info "Available scenarios: default, empty, nested, largefiles"
        exit 1
        ;;
esac

log_info "Fixture setup complete: ${SCENARIO}"
