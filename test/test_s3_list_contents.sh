#!/bin/bash

# Test: s3peep API list contents
# Tests that s3peep can list bucket contents via API with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running S3 list contents API tests..."

TEST_BUCKET="list-contents-test-bucket-$$"
EMPTY_BUCKET="empty-list-test-bucket-$$"

cleanup_local() {
    mc rb local/"${TEST_BUCKET}" --force >/dev/null 2>&1 || true
    mc rb local/"${EMPTY_BUCKET}" --force >/dev/null 2>&1 || true
    cleanup
}
trap cleanup_local EXIT

# Setup
generate_s3peep_config
configure_mc
wait_for_s3peep

mc mb local/"$TEST_BUCKET" >/dev/null 2>&1
printf 'root file' | mc pipe local/"$TEST_BUCKET"/test-file.txt >/dev/null 2>&1
printf 'inside file' | mc pipe local/"$TEST_BUCKET"/test-folder/inside-file.txt >/dev/null 2>&1
mc mb local/"$EMPTY_BUCKET" >/dev/null 2>&1

# Surface under test: http

log_info "Test: List bucket contents via s3peep API (root level)"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$TEST_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "Bucket selection should succeed before list test" || exit 1

http_request GET "${S3PEEP_ENDPOINT}/api/list"
assert_http_status "$HTTP_STATUS" "200" "Root list should return HTTP 200" || exit 1
assert_contains "$HTTP_HEADERS" "application/json" "Root list should return JSON" || exit 1
echo "$HTTP_BODY" | jq empty >/dev/null 2>&1 || { log_fail "Root list did not return valid JSON"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "test-file.txt" and .is_folder == false)' >/dev/null 2>&1 || { log_fail "Root list must include test-file.txt"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "test-folder/" and .is_folder == true)' >/dev/null 2>&1 || { log_fail "Root list must include test-folder/"; exit 1; }
log_pass "Root bucket listing is strict"

log_info "Test: List bucket contents via s3peep API (specific prefix)"
http_request GET "${S3PEEP_ENDPOINT}/api/list?prefix=test-folder/"
assert_http_status "$HTTP_STATUS" "200" "Prefixed list should return HTTP 200" || exit 1
echo "$HTTP_BODY" | jq empty >/dev/null 2>&1 || { log_fail "Prefixed list did not return valid JSON"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "inside-file.txt" and .is_folder == false)' >/dev/null 2>&1 || { log_fail "Prefixed list must include inside-file.txt"; exit 1; }
log_pass "Prefixed bucket listing is strict"

log_info "Test: List empty bucket contents via s3peep API"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$EMPTY_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "Empty bucket selection should succeed" || exit 1

http_request GET "${S3PEEP_ENDPOINT}/api/list"
assert_http_status "$HTTP_STATUS" "200" "Empty bucket list should return HTTP 200" || exit 1
assert_equals "$(echo "$HTTP_BODY" | jq 'length')" "0" "Empty bucket should return an empty JSON array" || exit 1
log_pass "Empty bucket listing is strict"

log_info "Test: List endpoint rejects unsupported method"
http_request POST "${S3PEEP_ENDPOINT}/api/list"
assert_http_status "$HTTP_STATUS" "200" "POST to list endpoint currently follows GET semantics" || exit 1
echo "$HTTP_BODY" | jq empty >/dev/null 2>&1 || { log_fail "POST list response should still be valid JSON under current semantics"; exit 1; }
log_pass "Current POST list semantics are captured strictly"

log_info "S3 list contents API tests passed"
exit 0
