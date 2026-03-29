#!/bin/bash

# Test: Web UI folder navigation
# Tests prefix-based folder navigation with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running Web UI folder navigation tests..."

TEST_BUCKET="webui-navigate-test-$$"

cleanup_local() {
    mc rb local/"${TEST_BUCKET}" --force >/dev/null 2>&1 || true
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config
configure_mc
wait_for_s3peep

mc mb local/"$TEST_BUCKET" >/dev/null 2>&1
printf 'deep file' | mc pipe local/"$TEST_BUCKET"/level1/level2/level3/deep-file.txt >/dev/null 2>&1
printf 'mid file' | mc pipe local/"$TEST_BUCKET"/level1/mid-file.txt >/dev/null 2>&1
printf 'top file' | mc pipe local/"$TEST_BUCKET"/top-file.txt >/dev/null 2>&1

http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$TEST_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "Folder navigation test bucket selection should succeed" || exit 1

# Surface under test: http

log_info "Test: Navigate into level1 folder"
http_request GET "${S3PEEP_ENDPOINT}/api/list?prefix=level1/"
assert_http_status "$HTTP_STATUS" "200" "level1/ navigation should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.key == "level1/level2/" and .is_folder == true)' >/dev/null 2>&1 || { log_fail "level1/ must include level1/level2/"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "mid-file.txt" and .is_folder == false)' >/dev/null 2>&1 || { log_fail "level1/ must include mid-file.txt"; exit 1; }
log_pass "level1 navigation is strict"

log_info "Test: Navigate into level2 folder"
http_request GET "${S3PEEP_ENDPOINT}/api/list?prefix=level1/level2/"
assert_http_status "$HTTP_STATUS" "200" "level2/ navigation should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.key == "level1/level2/level3/" and .is_folder == true)' >/dev/null 2>&1 || { log_fail "level2/ must include level1/level2/level3/"; exit 1; }
log_pass "level2 navigation is strict"

log_info "Test: Navigate into level3 folder"
http_request GET "${S3PEEP_ENDPOINT}/api/list?prefix=level1/level2/level3/"
assert_http_status "$HTTP_STATUS" "200" "level3/ navigation should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "deep-file.txt" and .is_folder == false)' >/dev/null 2>&1 || { log_fail "level3/ must include deep-file.txt"; exit 1; }
log_pass "level3 navigation is strict"

log_info "Test: Navigate back to parent folder"
http_request GET "${S3PEEP_ENDPOINT}/api/list?prefix=level1/level2/"
assert_http_status "$HTTP_STATUS" "200" "Returning to level2/ should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "deep-file.txt")' >/dev/null 2>&1 && { log_fail "level2/ must not show deep-file.txt directly"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.key == "level1/level2/level3/" and .is_folder == true)' >/dev/null 2>&1 || { log_fail "level2/ must still include level1/level2/level3/ after back navigation"; exit 1; }
log_pass "Parent navigation is strict"

log_info "Web UI folder navigation tests passed"
exit 0
