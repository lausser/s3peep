#!/bin/bash

# Test: s3peep config invalid JSON handling
# Tests that s3peep handles invalid JSON in config file gracefully and strictly

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running config invalid JSON tests..."

TEST_CONFIG="${S3PEEP_CONFIG_DIR}/test_invalid_json.json"
rm -f "${TEST_CONFIG}"

cleanup() {
    log_info "Cleaning up test resources..."
    rm -f "${TEST_CONFIG}" /tmp/test_config_invalid_json_1.out /tmp/test_config_invalid_json_1.err /tmp/test_config_invalid_json_2.out /tmp/test_config_invalid_json_2.err /tmp/test_config_invalid_json_3.out /tmp/test_config_invalid_json_3.err
}
trap cleanup EXIT

# Surface under test: cli

log_info "Test: Completely invalid JSON"
echo "this is not json at all" > "${TEST_CONFIG}"
set +e
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile list > /tmp/test_config_invalid_json_1.out 2> /tmp/test_config_invalid_json_1.err
exit_code=$?
set -e
assert_failure "$exit_code" "Completely invalid JSON should fail" || exit 1
assert_contains "$(cat /tmp/test_config_invalid_json_1.err /tmp/test_config_invalid_json_1.out 2>/dev/null)" "failed to load config" "CLI should explain invalid config load failure" || exit 1
log_pass "Completely invalid JSON is rejected strictly"

log_info "Test: Partially invalid JSON"
cat > "${TEST_CONFIG}" <<EOF
{
  "active_profile": "test",
  "profiles": [
    {
      "name": "test"
    }
  ]
}
EOF
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile list > /tmp/test_config_invalid_json_2.out 2> /tmp/test_config_invalid_json_2.err
exit_code=$?
assert_success "$exit_code" "Partially filled but syntactically valid JSON currently loads" || exit 1
assert_contains "$(cat /tmp/test_config_invalid_json_2.out)" "test (active)" "CLI should reflect currently loaded profile even with missing optional fields" || exit 1
log_pass "Current semantics for syntactically valid partial config are captured"

log_info "Test: Valid JSON with wrong data types"
cat > "${TEST_CONFIG}" <<EOF
{
  "active_profile": 123,
  "profiles": [
    {
      "name": "test",
      "region": "us-east-1",
      "access_key_id": "testkey",
      "secret_access_key": "testsecret",
      "bucket": "test-bucket"
    }
  ]
}
EOF
set +e
CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile list > /tmp/test_config_invalid_json_3.out 2> /tmp/test_config_invalid_json_3.err
exit_code=$?
set -e
assert_failure "$exit_code" "Wrong typed JSON should fail" || exit 1
assert_contains "$(cat /tmp/test_config_invalid_json_3.err /tmp/test_config_invalid_json_3.out 2>/dev/null)" "cannot unmarshal number into Go struct field Config.active_profile" "CLI should expose typed JSON parsing failure" || exit 1
log_pass "Wrong data types are rejected strictly"

log_info "Config invalid JSON tests passed"
exit 0
