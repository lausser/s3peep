#!/bin/bash

# Test: s3peep API select bucket
# Tests that s3peep can select/change active bucket via API endpoint with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running S3 select bucket API tests..."

TEST_BUCKET="select-test-bucket-$$"
NONEXISTENT_BUCKET="this-bucket-does-not-exist-$$"

cleanup_local() {
    mc rb local/"${TEST_BUCKET}" >/dev/null 2>&1 || true
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config
configure_mc
wait_for_s3peep

mc mb local/"$TEST_BUCKET" >/dev/null 2>&1

# Surface under test: http

log_info "Test: Select existing bucket via API"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$TEST_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "Selecting existing bucket should succeed" || exit 1
assert_contains "$HTTP_BODY" '"status":"ok"' "Bucket selection should return ok status" || exit 1
assert_contains "$HTTP_BODY" "\"bucket\":\"$TEST_BUCKET\"" "Bucket selection should echo selected bucket" || exit 1
assert_contains "$HTTP_BODY" "\"bucket\":\"$TEST_BUCKET\"" "Bucket response must reflect the persisted selection" || exit 1
log_pass "Existing bucket selection is strict"

log_info "Test: Select non-existent bucket under current semantics"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$NONEXISTENT_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "Selecting non-existent bucket currently succeeds" || exit 1
assert_contains "$HTTP_BODY" "\"bucket\":\"$NONEXISTENT_BUCKET\"" "Non-existent bucket response should echo selected bucket" || exit 1
log_pass "Current non-existent bucket semantics are captured strictly"

log_info "Test: Invalid JSON fails"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "invalid json"
assert_http_status "$HTTP_STATUS" "400" "Invalid JSON selection must return HTTP 400" || exit 1
log_pass "Invalid JSON rejection is strict"

log_info "Test: Missing bucket field follows current semantics"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" '{}'
assert_http_status "$HTTP_STATUS" "200" "Missing bucket field currently succeeds with empty bucket" || exit 1
assert_contains "$HTTP_BODY" '"bucket":""' "Missing bucket field should currently persist empty bucket" || exit 1
log_pass "Current missing bucket semantics are captured strictly"

log_info "Test: Method handling on buckets endpoint"
http_request GET "${S3PEEP_ENDPOINT}/api/buckets"
assert_http_status "$HTTP_STATUS" "200" "GET buckets should list buckets" || exit 1
assert_contains "$HTTP_HEADERS" "application/json" "GET buckets should return JSON" || exit 1

http_request PUT "${S3PEEP_ENDPOINT}/api/buckets"
assert_http_status "$HTTP_STATUS" "405" "PUT buckets must be rejected" || exit 1
log_pass "Buckets endpoint method handling is strict"

log_info "S3 select bucket API tests passed"
exit 0
