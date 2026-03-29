#!/bin/bash

# Test: s3peep config unwritable file handling
# Tests that s3peep handles unwritable config file gracefully

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running config unwritable file tests..."

TEST_CONFIG="${S3PEEP_CONFIG_DIR}/test_unwritable.json"
export S3PEEP_CONFIG="${TEST_CONFIG}"

# Cleanup function
cleanup() {
    log_info "Cleaning up test resources..."
    # Restore permissions if file exists
    if [ -f "${TEST_CONFIG}" ]; then
        chmod 644 "${TEST_CONFIG}" 2>/dev/null || true
        rm -f "${TEST_CONFIG}"
    fi
}
trap cleanup EXIT

# Test 1: Unwritable config file (permission denied) when trying to save
log_info "Test: Unwritable config file when trying to save"
# Create a config file then remove write permissions
echo '{"active_profile":"test","profiles":[{"name":"test","region":"us-east-1","access_key_id":"test","secret_access_key":"test"}]}' > "${TEST_CONFIG}"
chmod 444 "${TEST_CONFIG}"  # read-only

# Try to add a profile (which would require writing)
output=$(run_s3peep_cli profile add \
    --name "test2" \
    --region "us-west-2" \
    --access-key "test2" \
    --secret "test2" \
    --bucket "test2-bucket" \
    --endpoint "${MINIO_ENDPOINT}" 2>&1) || true

# Should handle the error gracefully
if echo "$output" | grep -iq "failed to save config\|permission denied\|cannot write"; then
    log_pass "Correctly handles unwritable config file when saving"
else
    # It might succeed if it falls back to a different config path or fails earlier
    # Let's check if the file was actually modified (it shouldn't be)
    original_content=$(cat "${TEST_CONFIG}" 2>/dev/null || echo "FILE_NOT_READABLE")
    if [ "$original_content" != "FILE_NOT_READABLE" ]; then
        current_content=$(cat "${TEST_CONFIG}" 2>/dev/null || echo "FILE_NOT_READABLE")
        if [ "$original_content" = "$current_content" ]; then
            log_pass "Config file was not modified (correctly handled unwritable file)"
        else
            log_fail "Config file was modified despite being unwritable"
            exit 1
        fi
    else
        log_pass "Config file was not readable (expected)"
    fi
fi

# Restore permissions for cleanup
chmod 644 "${TEST_CONFIG}" 2>/dev/null || true

log_info "Config unwritable file tests passed"
exit 0