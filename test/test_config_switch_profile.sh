#!/bin/bash

# Test: s3peep config switch profile
# Tests that s3peep can switch active profile via CLI with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running config switch profile tests..."

TEST_CONFIG="${S3PEEP_CONFIG_DIR}/test_switch_profile.json"
rm -f "${TEST_CONFIG}"

cleanup() {
    log_info "Cleaning up test resources..."
    rm -f "${TEST_CONFIG}" /tmp/test_config_switch.out /tmp/test_config_switch.err /tmp/test_config_switch_same.out /tmp/test_config_switch_same.err /tmp/test_config_switch_missing.out /tmp/test_config_switch_missing.err /tmp/test_config_switch_unknown.out /tmp/test_config_switch_unknown.err
}
trap cleanup EXIT

# Surface under test: cli

log_info "Test: Switch profile via CLI"
generate_s3peep_config "${TEST_CONFIG}" "original-profile" "original-bucket" "${MINIO_ENDPOINT}"
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile add \
    --name "new-profile" \
    --region "us-west-2" \
    --access-key "newkey123" \
    --secret "newsecret456" \
    --bucket "new-bucket" \
    --endpoint "${MINIO_ENDPOINT}" > /dev/null 2>&1

CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile switch --name "new-profile" > /tmp/test_config_switch.out 2> /tmp/test_config_switch.err
exit_code=$?
assert_success "$exit_code" "Profile switch should succeed" || exit 1
assert_contains "$(cat /tmp/test_config_switch.out)" "Switched to profile 'new-profile'" "CLI should confirm switched profile" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.active_profile' 'new-profile' "Active profile should switch to new profile" || exit 1
log_pass "Profile switch persists exact active profile"

log_info "Test: Switch to same profile"
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile switch --name "new-profile" > /tmp/test_config_switch_same.out 2> /tmp/test_config_switch_same.err
exit_code=$?
assert_success "$exit_code" "Switching to same profile should succeed" || exit 1
assert_contains "$(cat /tmp/test_config_switch_same.out)" "Switched to profile 'new-profile'" "CLI should still confirm same-profile switch" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.active_profile' 'new-profile' "Active profile should remain new-profile" || exit 1
log_pass "Same-profile switch remains deterministic"

log_info "Test: Switch with missing name flag fails"
set +e
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile switch > /tmp/test_config_switch_missing.out 2> /tmp/test_config_switch_missing.err
exit_code=$?
set -e
assert_failure "$exit_code" "Missing profile name should fail" || exit 1
assert_contains "$(cat /tmp/test_config_switch_missing.err /tmp/test_config_switch_missing.out 2>/dev/null)" "profile name is required" "CLI should explain missing switch name" || exit 1
log_pass "Missing switch name is rejected strictly"

log_info "Test: Switch to non-existent profile fails"
set +e
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile switch --name "non-existent-profile" > /tmp/test_config_switch_unknown.out 2> /tmp/test_config_switch_unknown.err
exit_code=$?
set -e
assert_failure "$exit_code" "Unknown profile switch should fail" || exit 1
assert_contains "$(cat /tmp/test_config_switch_unknown.err /tmp/test_config_switch_unknown.out 2>/dev/null)" "not found" "CLI should explain missing profile" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.active_profile' 'new-profile' "Failed switch must not change active profile" || exit 1
log_pass "Unknown profile rejection is strict"

log_info "Config switch profile tests passed"
exit 0
