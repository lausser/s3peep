#!/bin/bash

# Test: s3peep config delete profile
# Tests that s3peep can remove profiles via CLI with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running config delete profile tests..."

TEST_CONFIG="${S3PEEP_CONFIG_DIR}/test_delete_profile.json"
rm -f "${TEST_CONFIG}"

cleanup() {
    log_info "Cleaning up test resources..."
    rm -f "${TEST_CONFIG}" /tmp/test_config_delete_non_active.out /tmp/test_config_delete_non_active.err /tmp/test_config_delete_active.out /tmp/test_config_delete_active.err /tmp/test_config_delete_missing.out /tmp/test_config_delete_missing.err /tmp/test_config_delete_unknown.out /tmp/test_config_delete_unknown.err
}
trap cleanup EXIT

# Surface under test: cli

log_info "Test: Delete non-active profile via CLI"
generate_s3peep_config "${TEST_CONFIG}" "profile-to-keep" "keep-bucket" "${MINIO_ENDPOINT}"
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile add \
    --name "profile-to-delete" \
    --region "us-west-2" \
    --access-key "deletekey123" \
    --secret "deletesecret456" \
    --bucket "delete-bucket" \
    --endpoint "${MINIO_ENDPOINT}" > /dev/null 2>&1
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile switch --name "profile-to-keep" > /dev/null 2>&1

CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile remove --name "profile-to-delete" > /tmp/test_config_delete_non_active.out 2> /tmp/test_config_delete_non_active.err
exit_code=$?
assert_success "$exit_code" "Removing non-active profile should succeed" || exit 1
assert_contains "$(cat /tmp/test_config_delete_non_active.out)" "Profile 'profile-to-delete' removed" "CLI should confirm removed non-active profile" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles | length' '1' "Deleting non-active profile should leave one profile" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.active_profile' 'profile-to-keep' "Deleting non-active profile must preserve active profile" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles[0].name' 'profile-to-keep' "Remaining profile should be profile-to-keep" || exit 1
log_pass "Non-active profile deletion is strict"

log_info "Test: Delete active profile via CLI"
generate_s3peep_config "${TEST_CONFIG}" "profile-to-delete" "delete-bucket" "${MINIO_ENDPOINT}"
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile add \
    --name "profile-to-keep" \
    --region "us-east-1" \
    --access-key "keepkey123" \
    --secret "keepsecret456" \
    --bucket "keep-bucket" \
    --endpoint "${MINIO_ENDPOINT}" > /dev/null 2>&1
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile switch --name "profile-to-delete" > /dev/null 2>&1

CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile remove --name "profile-to-delete" > /tmp/test_config_delete_active.out 2> /tmp/test_config_delete_active.err
exit_code=$?
assert_success "$exit_code" "Removing active profile should succeed" || exit 1
assert_contains "$(cat /tmp/test_config_delete_active.out)" "Profile 'profile-to-delete' removed" "CLI should confirm removed active profile" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles | length' '1' "Deleting active profile should leave one profile" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.active_profile' '' "Deleting active profile should clear active profile" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles[0].name' 'profile-to-keep' "Remaining profile should be profile-to-keep" || exit 1
log_pass "Active profile deletion is strict"

log_info "Test: Delete non-existent profile fails"
set +e
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile remove --name "non-existent-profile" > /tmp/test_config_delete_unknown.out 2> /tmp/test_config_delete_unknown.err
exit_code=$?
set -e
assert_failure "$exit_code" "Deleting unknown profile should fail" || exit 1
assert_contains "$(cat /tmp/test_config_delete_unknown.err /tmp/test_config_delete_unknown.out 2>/dev/null)" "not found" "CLI should explain missing profile deletion" || exit 1
log_pass "Unknown profile deletion is rejected strictly"

log_info "Test: Delete profile with missing name flag fails"
set +e
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile remove > /tmp/test_config_delete_missing.out 2> /tmp/test_config_delete_missing.err
exit_code=$?
set -e
assert_failure "$exit_code" "Deleting without name should fail" || exit 1
assert_contains "$(cat /tmp/test_config_delete_missing.err /tmp/test_config_delete_missing.out 2>/dev/null)" "profile name is required" "CLI should explain missing delete name" || exit 1
log_pass "Missing delete name is rejected strictly"

log_info "Config delete profile tests passed"
exit 0
