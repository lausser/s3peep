#!/bin/bash

# Test: Web UI breadcrumb navigation
# Tests the prefix-driven navigation contract that powers breadcrumb rendering

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running Web UI breadcrumb tests..."

TEST_BUCKET="webui-breadcrumb-test-$$"

cleanup_local() {
    mc rb local/"${TEST_BUCKET}" --force >/dev/null 2>&1 || true
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config
configure_mc
wait_for_s3peep

mc mb local/"$TEST_BUCKET" >/dev/null 2>&1
printf 'deep file' | mc pipe local/"$TEST_BUCKET"/folder1/subfolder2/deepfolder3/deep-file.txt >/dev/null 2>&1
printf 'mid file' | mc pipe local/"$TEST_BUCKET"/folder1/mid-file.txt >/dev/null 2>&1
printf 'top file' | mc pipe local/"$TEST_BUCKET"/top-file.txt >/dev/null 2>&1

http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$TEST_BUCKET\"}"
assert_http_status "$HTTP_STATUS" "200" "Breadcrumb test bucket selection should succeed" || exit 1

# Surface under test: http

log_info "Test: Root navigation data"
http_request GET "${S3PEEP_ENDPOINT}/api/list"
assert_http_status "$HTTP_STATUS" "200" "Root breadcrumb list should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "folder1/" and .is_folder == true)' >/dev/null 2>&1 || { log_fail "Root list must include folder1/"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "top-file.txt" and .is_folder == false)' >/dev/null 2>&1 || { log_fail "Root list must include top-file.txt"; exit 1; }
log_pass "Root breadcrumb data is strict"

log_info "Test: First-level breadcrumb navigation"
http_request GET "${S3PEEP_ENDPOINT}/api/list?prefix=folder1/"
assert_http_status "$HTTP_STATUS" "200" "First-level breadcrumb list should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.key == "folder1/subfolder2/" and .is_folder == true)' >/dev/null 2>&1 || { log_fail "folder1/ list must include folder1/subfolder2/"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "mid-file.txt" and .is_folder == false)' >/dev/null 2>&1 || { log_fail "folder1/ list must include mid-file.txt"; exit 1; }
log_pass "First-level breadcrumb data is strict"

log_info "Test: Deep breadcrumb navigation"
http_request GET "${S3PEEP_ENDPOINT}/api/list?prefix=folder1/subfolder2/deepfolder3/"
assert_http_status "$HTTP_STATUS" "200" "Deep breadcrumb list should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "deep-file.txt" and .is_folder == false)' >/dev/null 2>&1 || { log_fail "Deep breadcrumb list must include deep-file.txt"; exit 1; }
log_pass "Deep breadcrumb data is strict"

log_info "Test: Parent breadcrumb navigation excludes deeper file"
http_request GET "${S3PEEP_ENDPOINT}/api/list?prefix=folder1/subfolder2/"
assert_http_status "$HTTP_STATUS" "200" "Parent breadcrumb list should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.key == "folder1/subfolder2/deepfolder3/" and .is_folder == true)' >/dev/null 2>&1 || { log_fail "Parent breadcrumb list must include folder1/subfolder2/deepfolder3/"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "deep-file.txt")' >/dev/null 2>&1 && { log_fail "Parent breadcrumb list must not include deep-file.txt directly"; exit 1; }
log_pass "Parent breadcrumb navigation is strict"

log_info "Web UI breadcrumb tests passed"
exit 0
