#!/bin/bash

# Test: s3peep API file attributes
# Tests that s3peep returns correct file attributes via list endpoint with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running S3 file attributes API tests..."

TEST_BUCKET="file-attributes-test-bucket-$$"
TEST_CONTENT="Hello, World! This is a test file for checking attributes."

cleanup_local() {
    mc rb local/"${TEST_BUCKET}" --force >/dev/null 2>&1 || true
    rm -f /tmp/test-file.txt
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config
configure_mc
wait_for_s3peep

mc mb local/"$TEST_BUCKET" >/dev/null 2>&1
printf '%s' "$TEST_CONTENT" > /tmp/test-file.txt
mc cp /tmp/test-file.txt local/"$TEST_BUCKET"/test-file.txt >/dev/null 2>&1
printf 'nested' | mc pipe local/"$TEST_BUCKET"/test-folder/nested.txt >/dev/null 2>&1

http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$TEST_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "Attributes test bucket selection should succeed" || exit 1

# Surface under test: http

log_info "Test: Root list returns correct file attributes"
http_request GET "${S3PEEP_ENDPOINT}/api/list"
assert_http_status "$HTTP_STATUS" "200" "Attribute list should return HTTP 200" || exit 1
echo "$HTTP_BODY" | jq empty >/dev/null 2>&1 || { log_fail "Attribute list did not return valid JSON"; exit 1; }
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "test-file.txt") | .name')" "test-file.txt" "File name attribute must match" || exit 1
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "test-file.txt") | .size')" "58" "File size attribute must match exact content size" || exit 1
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "test-file.txt") | .is_folder')" "false" "File must not be marked as folder" || exit 1
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "test-file.txt") | .key')" "test-file.txt" "File key attribute must match" || exit 1
assert_not_contains "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "test-file.txt") | .last_modified')" "0001-01-01T00:00:00Z" "File last_modified must be populated" || exit 1
log_pass "File attributes are strict"

log_info "Test: Root list returns correct folder attributes"
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "test-folder/") | .name')" "test-folder/" "Folder name attribute must include trailing slash" || exit 1
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "test-folder/") | .is_folder')" "true" "Folder must be marked as folder" || exit 1
log_pass "Folder attributes are strict"

log_info "S3 file attributes API tests passed"
exit 0
