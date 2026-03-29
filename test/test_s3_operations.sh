#!/bin/bash

# Test: S3 operations
# Tests bucket listing, file listing, and file access via s3peep

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running S3 operations tests..."

# Setup
generate_s3peep_config
configure_mc
mc ls local/test-bucket >/dev/null 2>&1 || { log_fail "Fixture bucket must exist before strict S3 tests"; exit 1; }

# Surface under test: http

# Test 1: List buckets via s3peep API
log_info "Test: List buckets via s3peep API"
http_request GET "${S3PEEP_ENDPOINT}/api/buckets"
assert_http_status "$HTTP_STATUS" "200" "Bucket listing should return HTTP 200" || exit 1
assert_contains "$HTTP_HEADERS" "application/json" "Bucket listing should return JSON" || exit 1
echo "$HTTP_BODY" | jq empty >/dev/null 2>&1 || { log_fail "Bucket listing did not return valid JSON"; exit 1; }
assert_contains "$HTTP_BODY" '"name":"test-bucket"' "Bucket listing should include test-bucket" || exit 1
log_pass "s3peep lists buckets through API"

# Test 2: Select bucket via s3peep API
log_info "Test: Select bucket via s3peep API"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" '{"bucket":"test-bucket"}'
assert_http_status "$HTTP_STATUS" "200" "Bucket selection should return HTTP 200" || exit 1
assert_contains "$HTTP_BODY" '"bucket":"test-bucket"' "Bucket selection response should confirm selected bucket" || exit 1
log_pass "s3peep selects bucket through API"

# Reselect bucket through config-backed startup path to avoid stale in-memory target ambiguity
generate_s3peep_config

# Test 3: List root objects via s3peep API
log_info "Test: List root objects via s3peep API"
http_request GET "${S3PEEP_ENDPOINT}/api/list?prefix="
if [ "$HTTP_STATUS" != "200" ]; then
    log_error "List response body: $HTTP_BODY"
fi
assert_http_status "$HTTP_STATUS" "200" "Object listing should return HTTP 200" || exit 1
echo "$HTTP_BODY" | jq empty >/dev/null 2>&1 || { log_fail "Object listing did not return valid JSON"; exit 1; }
assert_contains "$HTTP_BODY" '"name":"documents/"' "Object listing should include documents folder" || exit 1
assert_contains "$HTTP_BODY" '"name":"data/"' "Object listing should include data folder" || exit 1
log_pass "s3peep lists bucket contents through API"

# Test 4: Download specific file via s3peep API
log_info "Test: Download specific file via s3peep API"
http_request GET "${S3PEEP_ENDPOINT}/api/get?key=documents/hello.txt"
assert_http_status "$HTTP_STATUS" "200" "File download should return HTTP 200" || exit 1
assert_header_contains "$HTTP_HEADERS" "Content-Disposition: attachment; filename=hello.txt" "Download should set content disposition header" || exit 1
assert_contains "$HTTP_BODY" "Hello" "Downloaded file should contain expected content" || exit 1
log_pass "s3peep downloads files through API"

# Test 5: Invalid object request fails strictly
log_info "Test: Invalid object request fails strictly"
http_request GET "${S3PEEP_ENDPOINT}/api/get?key=missing-file.txt"
assert_http_status "$HTTP_STATUS" "500" "Missing object should return strict failure status" || exit 1
assert_not_contains "$HTTP_BODY" "Hello" "Missing object response must not return file content" || exit 1
log_pass "Invalid object requests fail strictly"

# Cleanup
cleanup

log_info "S3 operations tests passed"
exit 0
