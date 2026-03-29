#!/bin/bash

# Test: s3peep config unreadable file handling
# Tests current unreadable-config behavior strictly in the containerized runner

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running config unreadable file tests..."

TEST_CONFIG="${S3PEEP_CONFIG_DIR}/test_unreadable.json"

cleanup_local() {
    log_info "Cleaning up test resources..."
    if [ -f "${TEST_CONFIG}" ]; then
        chmod 644 "${TEST_CONFIG}" 2>/dev/null || true
        rm -f "${TEST_CONFIG}"
    fi
    rm -f /tmp/test_config_unreadable.out /tmp/test_config_unreadable.err
}
trap cleanup_local EXIT

# Surface under test: cli

log_info "Test: Unreadable config file in root-runner context"
cat > "${TEST_CONFIG}" <<EOF
{
  "active_profile": "test",
  "profiles": [
    {
      "name": "test",
      "region": "us-east-1",
      "access_key_id": "test",
      "secret_access_key": "test"
    }
  ]
}
EOF
chmod 000 "${TEST_CONFIG}"

CONFIG="${TEST_CONFIG}" "${S3PEEP_BIN}" profile list > /tmp/test_config_unreadable.out 2> /tmp/test_config_unreadable.err
exit_code=$?

assert_success "$exit_code" "Unreadable test should reflect current root-runner semantics" || exit 1
output="$(cat /tmp/test_config_unreadable.out /tmp/test_config_unreadable.err 2>/dev/null)"
assert_contains "$output" "Profiles:" "Root-runner context should still read the config successfully" || exit 1
assert_contains "$output" "test (active)" "Unreadable-permissions test should document current root access behavior" || exit 1
log_pass "Current unreadable-config semantics are captured strictly"

log_info "Config unreadable file tests passed"
exit 0
