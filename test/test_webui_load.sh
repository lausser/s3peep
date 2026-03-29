#!/bin/bash

# Test: Web UI loads HTML content
# Tests that s3peep web UI returns HTML content with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running Web UI load tests..."

generate_s3peep_config
configure_mc
wait_for_s3peep

# Surface under test: http

log_info "Test: Web UI returns strict HTML response"
http_request GET "${S3PEEP_ENDPOINT}/"
assert_http_status "$HTTP_STATUS" "200" "Web UI root should return HTTP 200" || exit 1
assert_content_type_contains "$HTTP_HEADERS" "text/html" "Web UI root should return text/html" || exit 1
assert_contains "$HTTP_BODY" "<!DOCTYPE html>" "Web UI should return HTML document doctype" || exit 1
assert_contains "$HTTP_BODY" "<title>S3 File Browser</title>" "Web UI should return exact page title" || exit 1
assert_contains "$HTTP_BODY" "id=\"content\"" "Web UI should include content container" || exit 1
assert_contains "$HTTP_BODY" "loadFiles" "Web UI should include file loading script" || exit 1
assert_not_contains "$HTTP_BODY" "Internal Server Error" "Web UI success page must not be a generic server error page" || exit 1
log_pass "Web UI root response is strict"

cleanup

log_info "Web UI load tests passed"
exit 0
