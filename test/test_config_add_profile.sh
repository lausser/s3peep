#!/bin/bash

# Test: s3peep config add profile
# Tests that s3peep can add profiles via CLI with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running config add profile tests..."

TEST_CONFIG="${S3PEEP_CONFIG_DIR}/test_add_profile.json"
rm -f "${TEST_CONFIG}"

cleanup() {
    log_info "Cleaning up test resources..."
    rm -f "${TEST_CONFIG}" /tmp/test_config_add_profile.out /tmp/test_config_add_profile.err /tmp/test_config_add_duplicate.out /tmp/test_config_add_duplicate.err /tmp/test_config_add_missing.out /tmp/test_config_add_missing.err
}
trap cleanup EXIT

# Surface under test: cli

# Test 1: Add profile via CLI
log_info "Test: Add profile via CLI"
generate_s3peep_config "${TEST_CONFIG}" "initial-profile" "initial-bucket" "${MINIO_ENDPOINT}"

CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile add \
    --name "test-profile" \
    --region "us-west-2" \
    --access-key "testkey123" \
    --secret "testsecret456" \
    --bucket "test-bucket" \
    --endpoint "${MINIO_ENDPOINT}" > /tmp/test_config_add_profile.out 2> /tmp/test_config_add_profile.err
exit_code=$?

assert_success "$exit_code" "Profile add should succeed" || exit 1
assert_contains "$(cat /tmp/test_config_add_profile.out)" "Profile 'test-profile' added successfully" "CLI should report successful profile creation" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles | length' '2' "Config should contain two profiles after adding" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.active_profile' 'test-profile' "Added profile should become active" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles[1].name' 'test-profile' "Added profile name should persist" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles[1].region' 'us-west-2' "Added profile region should persist" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles[1].bucket' 'test-bucket' "Added profile bucket should persist" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles[1].endpoint_url' "$MINIO_ENDPOINT" "Added profile endpoint should persist" || exit 1
log_pass "Profile add persists exact config changes"

# Test 2: Duplicate profile must fail
log_info "Test: Add duplicate profile fails"
set +e
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile add \
    --name "test-profile" \
    --region "eu-central-1" \
    --access-key "anotherkey" \
    --secret "anothersecret" \
    --bucket "another-bucket" \
    --endpoint "${MINIO_ENDPOINT}" > /tmp/test_config_add_duplicate.out 2> /tmp/test_config_add_duplicate.err
exit_code=$?
set -e

assert_failure "$exit_code" "Duplicate profile add should fail" || exit 1
assert_contains "$(cat /tmp/test_config_add_duplicate.err /tmp/test_config_add_duplicate.out 2>/dev/null)" "already exists" "Duplicate profile failure should explain the conflict" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles | length' '2' "Duplicate failure must not change profile count" || exit 1
log_pass "Duplicate profile rejection is strict"

# Test 3: Missing required flags must fail
log_info "Test: Add profile with missing required flags fails"
set +e
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile add \
    --name "incomplete-profile" \
    --region "us-east-1" \
    --access-key "testkey" > /tmp/test_config_add_missing.out 2> /tmp/test_config_add_missing.err
exit_code=$?
set -e

assert_failure "$exit_code" "Missing required flags should fail" || exit 1
assert_contains "$(cat /tmp/test_config_add_missing.err /tmp/test_config_add_missing.out 2>/dev/null)" "required flags" "CLI should explain missing required flags" || exit 1
assert_file_contains_json_value "${TEST_CONFIG}" '.profiles | length' '2' "Failed add must not change profile count" || exit 1
log_pass "Missing required flags are rejected strictly"

log_info "Config add profile tests passed"
exit 0
