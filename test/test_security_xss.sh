#!/bin/bash

# Test: XSS-like filename handling
# Tests that script-like filenames are returned as data and not transformed into executable content by the API contract

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running XSS filename security tests..."

XSS_BUCKET="xss-test-bucket-$$"
XSS_KEY='<script>alert(1)</script>.txt'

cleanup_local() {
    mc rb local/"${XSS_BUCKET}" --force >/dev/null 2>&1 || true
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config
configure_mc
wait_for_s3peep

mc mb local/"${XSS_BUCKET}" >/dev/null 2>&1
printf 'xss payload file' | mc pipe local/"${XSS_BUCKET}"/"${XSS_KEY}" >/dev/null 2>&1

# Surface under test: http

log_info "Test: XSS-like filename is returned in JSON data only"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"${XSS_BUCKET}\"}"
assert_http_status "$HTTP_STATUS" "200" "Bucket selection should succeed for XSS test bucket" || exit 1

http_request GET "${S3PEEP_ENDPOINT}/api/list"
assert_http_status "$HTTP_STATUS" "200" "List endpoint should return HTTP 200 for XSS test bucket" || exit 1
assert_contains "$HTTP_HEADERS" "application/json" "List endpoint should return JSON for XSS test bucket" || exit 1
echo "$HTTP_BODY" | jq empty >/dev/null 2>&1 || { log_fail "List endpoint did not return valid JSON for XSS test bucket"; exit 1; }
assert_not_contains "$HTTP_BODY" '<script>' "JSON list should not emit raw script tags" || exit 1
assert_contains "$HTTP_BODY" '\\u003cscript' "JSON list should escape script-like filename content" || exit 1
log_pass "Script-like filename is escaped in JSON response"

log_info "Test: Download headers do not execute script-like filenames"
encoded_key='%3Cscript%3Ealert(1)%3C%2Fscript%3E.txt'
http_request GET "${S3PEEP_ENDPOINT}/api/get?key=${encoded_key}"
assert_http_status "$HTTP_STATUS" "200" "Download of escaped script-like filename should succeed for stored object" || exit 1
assert_header_contains "$HTTP_HEADERS" 'Content-Disposition: attachment; filename=' "Download should force attachment for escaped script-like filename" || exit 1
assert_contains "$HTTP_BODY" 'xss payload file' "Download should return stored object content for escaped script-like filename" || exit 1
log_pass "Script-like filename is served as attachment data"

log_info "XSS filename security tests passed"
exit 0
