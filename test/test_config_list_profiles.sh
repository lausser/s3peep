#!/bin/bash

# Test: s3peep config list profiles
# Tests that s3peep can list profiles via CLI with strict assertions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running config list profiles tests..."

TEST_CONFIG="${S3PEEP_CONFIG_DIR}/test_list_profiles.json"
rm -f "${TEST_CONFIG}"

cleanup() {
    log_info "Cleaning up test resources..."
    rm -f "${TEST_CONFIG}" /tmp/test_config_list_empty.out /tmp/test_config_list_empty.err /tmp/test_config_list_full.out /tmp/test_config_list_full.err
}
trap cleanup EXIT

# Surface under test: cli

# Test 1: List profiles when none exist
log_info "Test: List profiles when none exist"
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" init > /dev/null 2>&1
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile list > /tmp/test_config_list_empty.out 2> /tmp/test_config_list_empty.err
exit_code=$?

assert_success "$exit_code" "Profile list should succeed for empty config" || exit 1
assert_equals "$(cat /tmp/test_config_list_empty.out)" "No profiles configured" "CLI should report empty profile list exactly" || exit 1
log_pass "Empty profile list output is strict"

# Test 2: List profiles when some exist
log_info "Test: List profiles when some exist"
cat > "${TEST_CONFIG}" <<EOF
{
  "active_profile": "active-profile",
  "profiles": [
    {
      "name": "inactive-profile",
      "region": "us-east-1",
      "access_key_id": "inactivekey",
      "secret_access_key": "inactiversecret",
      "bucket": "inactive-bucket"
    },
    {
      "name": "active-profile",
      "region": "us-west-2",
      "access_key_id": "activekey",
      "secret_access_key": "activesecret",
      "bucket": "active-bucket",
      "endpoint_url": "http://custom-endpoint:9000"
    }
  ]
}
EOF

CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile list > /tmp/test_config_list_full.out 2> /tmp/test_config_list_full.err
exit_code=$?

assert_success "$exit_code" "Profile list should succeed for populated config" || exit 1
output="$(cat /tmp/test_config_list_full.out)"
assert_contains "$output" "Profiles:" "Profile list should include header" || exit 1
assert_contains "$output" "inactive-profile" "Profile list should include inactive profile" || exit 1
assert_contains "$output" "active-profile (active)" "Profile list should mark active profile" || exit 1
assert_contains "$output" "Endpoint: http://custom-endpoint:9000" "Profile list should show custom endpoint" || exit 1
assert_contains "$output" "Region: us-east-1, Bucket: inactive-bucket" "Profile list should show inactive profile details" || exit 1
assert_contains "$output" "Region: us-west-2, Bucket: active-bucket" "Profile list should show active profile details" || exit 1
log_pass "Populated profile list output is strict"

log_info "Config list profiles tests passed"
exit 0
