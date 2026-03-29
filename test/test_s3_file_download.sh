#!/bin/bash

# Test: s3peep API file download
# Tests that s3peep can download files via API with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running S3 file download API tests..."

TEST_BUCKET="file-download-test-bucket-$$"
TEST_CONTENT="Hello, World! This is a test file for download testing."

cleanup_local() {
    mc rb local/"${TEST_BUCKET}" --force >/dev/null 2>&1 || true
    rm -f /tmp/test-download.txt
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config
configure_mc
wait_for_s3peep

mc mb local/"$TEST_BUCKET" >/dev/null 2>&1
printf '%s' "$TEST_CONTENT" > /tmp/test-download.txt
mc cp /tmp/test-download.txt local/"$TEST_BUCKET"/test-download.txt >/dev/null 2>&1

http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$TEST_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "Download test bucket selection should succeed" || exit 1

# Surface under test: http

log_info "Test: Download file via s3peep API"
http_request GET "${S3PEEP_ENDPOINT}/api/get?key=test-download.txt"
assert_http_status "$HTTP_STATUS" "200" "File download should return HTTP 200" || exit 1
assert_header_contains "$HTTP_HEADERS" 'Content-Type: application/octet-stream' "Download should return octet-stream content type" || exit 1
assert_header_contains "$HTTP_HEADERS" 'Content-Disposition: attachment; filename=test-download.txt' "Download should return attachment header" || exit 1
assert_equals "$HTTP_BODY" "$TEST_CONTENT" "Downloaded content must match uploaded content exactly" || exit 1
log_pass "File download response is strict"

log_info "Test: Download non-existent file fails"
http_request GET "${S3PEEP_ENDPOINT}/api/get?key=non-existent-file.txt"
assert_http_status "$HTTP_STATUS" "500" "Missing download should return server error" || exit 1
assert_contains "$HTTP_BODY" "Failed to get object" "Missing download should explain lookup failure" || exit 1
log_pass "Missing file download is strict"

log_info "Test: Download without key fails"
http_request GET "${S3PEEP_ENDPOINT}/api/get"
assert_http_status "$HTTP_STATUS" "400" "Missing key should return HTTP 400" || exit 1
assert_contains "$HTTP_BODY" "key is required" "Missing key should explain required parameter" || exit 1
log_pass "Missing key failure is strict"

log_info "Test: POST to get endpoint follows current behavior"
http_request POST "${S3PEEP_ENDPOINT}/api/get?key=test-download.txt"
assert_http_status "$HTTP_STATUS" "200" "POST to get endpoint currently succeeds" || exit 1
assert_equals "$HTTP_BODY" "$TEST_CONTENT" "POST get response should still match file content under current semantics" || exit 1
log_pass "Current POST get semantics are captured strictly"

log_info "S3 file download API tests passed"
exit 0
