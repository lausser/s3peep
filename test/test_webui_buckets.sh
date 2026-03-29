#!/bin/bash

# Test: Web UI bucket loading behavior
# Tests the HTTP contract behind bucket loading used by the web UI

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running Web UI buckets tests..."

TEST_BUCKET="webui-bucket-test-$$"

cleanup_local() {
    mc rb local/"${TEST_BUCKET}" >/dev/null 2>&1 || true
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config
configure_mc
wait_for_s3peep

# Surface under test: http

log_info "Test: Web UI shell includes required loading hooks"
http_request GET "${S3PEEP_ENDPOINT}/"
assert_http_status "$HTTP_STATUS" "200" "Web UI root should load successfully" || exit 1
assert_contains "$HTTP_BODY" "loadFiles" "Web UI shell should contain loadFiles hook" || exit 1
assert_contains "$HTTP_BODY" "id=\"content\"" "Web UI shell should contain content container" || exit 1
log_pass "Web UI shell contains loading hooks"

log_info "Test: Bucket API provides data for UI bucket loading"
http_request GET "${S3PEEP_ENDPOINT}/api/buckets"
assert_http_status "$HTTP_STATUS" "200" "Bucket listing should return HTTP 200" || exit 1
assert_contains "$HTTP_HEADERS" "application/json" "Bucket listing should return JSON" || exit 1
echo "$HTTP_BODY" | jq empty >/dev/null 2>&1 || { log_fail "Bucket listing did not return valid JSON"; exit 1; }
assert_contains "$HTTP_BODY" '"name":"test-bucket"' "Bucket listing should include default fixture bucket" || exit 1
log_pass "Bucket API provides strict data for the UI"

log_info "Test: Bucket selection API supports UI workflow"
mc mb local/"$TEST_BUCKET" >/dev/null 2>&1
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$TEST_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "Bucket selection should return HTTP 200" || exit 1
assert_contains "$HTTP_BODY" '"status":"ok"' "Bucket selection should report ok status" || exit 1
assert_contains "$HTTP_BODY" "\"bucket\":\"$TEST_BUCKET\"" "Bucket selection should echo selected bucket" || exit 1
log_pass "Bucket selection API supports UI workflow strictly"

log_info "Web UI buckets tests passed"
exit 0
