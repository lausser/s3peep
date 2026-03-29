#!/bin/bash

# Test: Web UI select bucket updates file list
# Tests the HTTP workflow behind bucket selection with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running Web UI select bucket tests..."

BUCKET1="webui-select-bucket1-$$"
BUCKET2="webui-select-bucket2-$$"

cleanup_local() {
    mc rb local/"${BUCKET1}" --force >/dev/null 2>&1 || true
    mc rb local/"${BUCKET2}" --force >/dev/null 2>&1 || true
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config
configure_mc
wait_for_s3peep

printf 'bucket1 file' | mc pipe local/"$BUCKET1"/file-in-bucket1.txt >/dev/null 2>&1 || mc mb local/"$BUCKET1" >/dev/null 2>&1
mc mb local/"$BUCKET1" >/dev/null 2>&1 || true
mc mb local/"$BUCKET2" >/dev/null 2>&1 || true
printf 'bucket1 file' | mc pipe local/"$BUCKET1"/file-in-bucket1.txt >/dev/null 2>&1
printf 'bucket2 file' | mc pipe local/"$BUCKET2"/file-in-bucket2.txt >/dev/null 2>&1
printf 'nested one' | mc pipe local/"$BUCKET1"/folder-in-bucket1/nested.txt >/dev/null 2>&1
printf 'nested two' | mc pipe local/"$BUCKET2"/folder-in-bucket2/nested.txt >/dev/null 2>&1

# Surface under test: http

log_info "Test: Select first bucket and verify list contents"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$BUCKET1\"}"
assert_http_status "$HTTP_STATUS" "200" "Bucket 1 selection should succeed" || exit 1
assert_contains "$HTTP_BODY" "\"bucket\":\"$BUCKET1\"" "Bucket 1 selection response should echo selected bucket" || exit 1

http_request GET "${S3PEEP_ENDPOINT}/api/list"
assert_http_status "$HTTP_STATUS" "200" "Bucket 1 listing should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "file-in-bucket1.txt")' >/dev/null 2>&1 || { log_fail "Bucket 1 listing must include file-in-bucket1.txt"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "folder-in-bucket1/" and .is_folder == true)' >/dev/null 2>&1 || { log_fail "Bucket 1 listing must include folder-in-bucket1/"; exit 1; }
log_pass "Bucket 1 selection updates list contents strictly"

log_info "Test: Select second bucket and verify list contents change"
http_request POST "${S3PEEP_ENDPOINT}/api/buckets" "{\"bucket\":\"$BUCKET2\"}"
assert_http_status "$HTTP_STATUS" "200" "Bucket 2 selection should succeed" || exit 1
assert_contains "$HTTP_BODY" "\"bucket\":\"$BUCKET2\"" "Bucket 2 selection response should echo selected bucket" || exit 1

http_request GET "${S3PEEP_ENDPOINT}/api/list"
assert_http_status "$HTTP_STATUS" "200" "Bucket 2 listing should succeed" || exit 1
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "file-in-bucket2.txt")' >/dev/null 2>&1 || { log_fail "Bucket 2 listing must include file-in-bucket2.txt"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "folder-in-bucket2/" and .is_folder == true)' >/dev/null 2>&1 || { log_fail "Bucket 2 listing must include folder-in-bucket2/"; exit 1; }
echo "$HTTP_BODY" | jq -e '.[] | select(.name == "file-in-bucket1.txt")' >/dev/null 2>&1 && { log_fail "Bucket 2 listing must not include bucket 1 file"; exit 1; }
log_pass "Bucket switch updates list contents strictly"

log_info "Web UI select bucket tests passed"
exit 0
