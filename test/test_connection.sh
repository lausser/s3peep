#!/bin/bash

# Test: s3peep connection
# Tests that s3peep can connect to MinIO (success and failure scenarios)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running connection tests..."

# Surface under test: http

# Test 1: Successful connection with valid credentials
log_info "Test: Successful connection with valid credentials"
generate_s3peep_config
http_request GET "${S3PEEP_ENDPOINT}/"
assert_http_status "$HTTP_STATUS" "200" "Web UI should return HTTP 200 for valid running server" || exit 1
assert_content_type_contains "$HTTP_HEADERS" "text/html" "Web UI should return HTML content type" || exit 1
assert_contains "$HTTP_BODY" "S3 File Browser" "Web UI should contain application title" || exit 1
printf '%s' "$HTTP_BODY" > /tmp/test_connection_body.html
log_pass "Successful connection exposes expected web UI"

# Test 2: s3peep web UI returns content
log_info "Test: s3peep web UI returns content"
assert_contains "$(cat /tmp/test_connection_body.html)" "id=\"content\"" "Web UI should include content container markup" || exit 1
assert_contains "$(cat /tmp/test_connection_body.html)" "Loading..." "Web UI should include loading placeholder text" || exit 1
assert_contains "$(cat /tmp/test_connection_body.html)" "loadFiles" "Web UI should include file loading script" || exit 1
log_pass "Web UI contains required contract markers"

# Test 3: Unsupported method fails strictly
log_info "Test: Unsupported method on buckets endpoint"
http_request PUT "${S3PEEP_ENDPOINT}/api/buckets"
assert_http_status "$HTTP_STATUS" "405" "Unsupported method should return HTTP 405" || exit 1
log_pass "Unsupported HTTP method fails strictly"

# Cleanup
cleanup
rm -f /tmp/test_connection_body.html

log_info "Connection tests passed"
exit 0
