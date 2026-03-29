#!/bin/bash

# Test: path traversal-like object access
# Tests that unsupported traversal-style keys do not return successful object content

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running path traversal security tests..."

generate_s3peep_config
configure_mc
wait_for_s3peep

# Surface under test: http

log_info "Test: Traversal-like key request fails"
http_request GET "${S3PEEP_ENDPOINT}/api/get?key=../../etc/passwd"
assert_http_status "$HTTP_STATUS" "500" "Traversal-like key must not succeed" || exit 1
assert_not_contains "$HTTP_BODY" "root:x:" "Traversal-like key must not expose host file content" || exit 1
log_pass "Traversal-like key is rejected strictly"

log_info "Test: Missing key parameter fails with bad request"
http_request GET "${S3PEEP_ENDPOINT}/api/get"
assert_http_status "$HTTP_STATUS" "400" "Missing key parameter must return HTTP 400" || exit 1
assert_contains "$HTTP_BODY" "key is required" "Missing key failure should explain required parameter" || exit 1
log_pass "Missing key handling is strict"

cleanup

log_info "Path traversal security tests passed"
exit 0
