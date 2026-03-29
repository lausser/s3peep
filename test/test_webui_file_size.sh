#!/bin/bash

# Test: Web UI file sizes display correctly
# Tests the API contract that the UI uses to render file sizes

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running Web UI file size tests..."

TEST_BUCKET="webui-file-size-test-$$"

cleanup_local() {
    mc rb local/"${TEST_BUCKET}" --force >/dev/null 2>&1 || true
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config
configure_mc
wait_for_s3peep

mc mb local/"$TEST_BUCKET" >/dev/null 2>&1
printf '' | mc pipe local/"$TEST_BUCKET"/zero-byte.txt >/dev/null 2>&1
dd if=/dev/zero bs=100 count=1 2>/dev/null | mc pipe local/"$TEST_BUCKET"/hundred-byte.txt >/dev/null 2>&1
dd if=/dev/zero bs=1000 count=1 2>/dev/null | mc pipe local/"$TEST_BUCKET"/thousand-byte.txt >/dev/null 2>&1
dd if=/dev/zero bs=1500000 count=1 2>/dev/null | mc pipe local/"$TEST_BUCKET"/pointfivemb.txt >/dev/null 2>&1

http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$TEST_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "File size test bucket selection should succeed" || exit 1

# Surface under test: http

log_info "Test: API returns exact file sizes"
http_request GET "${S3PEEP_ENDPOINT}/api/list"
assert_http_status "$HTTP_STATUS" "200" "File size list should return HTTP 200" || exit 1
echo "$HTTP_BODY" | jq empty >/dev/null 2>&1 || { log_fail "File size list did not return valid JSON"; exit 1; }
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "zero-byte.txt") | .size')" "0" "Zero-byte file size must be exact" || exit 1
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "hundred-byte.txt") | .size')" "100" "Hundred-byte file size must be exact" || exit 1
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "thousand-byte.txt") | .size')" "1000" "Thousand-byte file size must be exact" || exit 1
assert_equals "$(echo "$HTTP_BODY" | jq -r '.[] | select(.name == "pointfivemb.txt") | .size')" "1500000" "1.5MB file size must be exact" || exit 1
log_pass "API returns exact file sizes for UI formatting"

log_info "Web UI file size tests passed"
exit 0
