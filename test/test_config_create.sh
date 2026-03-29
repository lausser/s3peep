#!/bin/bash

# Test: s3peep config creation
# Tests that s3peep can create and use configuration files

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running config creation tests..."

TEST_CONFIG="${S3PEEP_CONFIG_DIR}/test_config.json"
rm -f "$TEST_CONFIG"

# Surface under test: cli

# Test 1: Create default config via CLI
log_info "Test: Create default config via CLI"
CONFIG="$TEST_CONFIG" "$S3PEEP_BIN" init > /tmp/test_config_create.out 2> /tmp/test_config_create.err
exit_code=$?
sleep 1
ls -l "$S3PEEP_CONFIG_DIR" > /tmp/test_config_ls.out 2>&1 || true
assert_success "$exit_code" "s3peep init should succeed" || exit 1

assert_file_exists "$TEST_CONFIG" "Config file should be created by s3peep init" || exit 1
jq empty "$TEST_CONFIG" >/dev/null 2>&1 || { log_fail "Config created by s3peep init is not valid JSON"; exit 1; }
assert_file_contains_json_value "$TEST_CONFIG" '.active_profile' '' "Default config should not set an active profile" || exit 1
assert_file_contains_json_value "$TEST_CONFIG" '.profiles | length' '0' "Default config should contain no profiles" || exit 1
assert_contains "$(cat /tmp/test_config_create.out)" "Created config at" "s3peep init should report created config path" || exit 1
log_pass "CLI created valid default config"

# Test 2: Add valid profile via CLI and verify persisted config
log_info "Test: Add valid profile via CLI"
CONFIG="$TEST_CONFIG" "$S3PEEP_BIN" profile add --name strict --region us-east-1 --access-key "$MINIO_ACCESS_KEY" --secret "$MINIO_SECRET_KEY" --bucket test-bucket --endpoint "$MINIO_ENDPOINT" > /tmp/test_config_add.out 2> /tmp/test_config_add.err
exit_code=$?
sleep 1
assert_success "$exit_code" "s3peep profile add should succeed" || exit 1
assert_file_contains_json_value "$TEST_CONFIG" '.active_profile' 'strict' "Added profile should become active" || exit 1
assert_file_contains_json_value "$TEST_CONFIG" '.profiles[0].name' 'strict' "Persisted profile name should match" || exit 1
assert_file_contains_json_value "$TEST_CONFIG" '.profiles[0].endpoint_url' "$MINIO_ENDPOINT" "Persisted endpoint should match MinIO" || exit 1
assert_contains "$(cat /tmp/test_config_add.out)" "added successfully" "CLI should report successful profile creation" || exit 1
log_pass "CLI persisted valid profile"

# Test 3: Invalid config must fail to load in serve mode
log_info "Test: Invalid config fails during serve startup"
cat > "$TEST_CONFIG" <<'EOF'
{ invalid json
EOF
set +e
timeout 5 env CONFIG="$TEST_CONFIG" "$S3PEEP_BIN" serve --port 18080 > /tmp/test_config_invalid.out 2> /tmp/test_config_invalid.err
exit_code=$?
set -e
assert_failure "$exit_code" "s3peep serve should fail with invalid config" || exit 1
assert_contains "$(cat /tmp/test_config_invalid.err /tmp/test_config_invalid.out 2>/dev/null)" "Failed to load config\|invalid character" "Serve failure should explain config parsing failure" || exit 1
log_pass "Invalid config causes strict startup failure"

# Cleanup
rm -f "$TEST_CONFIG" /tmp/test_config_create.out /tmp/test_config_create.err /tmp/test_config_add.out /tmp/test_config_add.err /tmp/test_config_invalid.out /tmp/test_config_invalid.err /tmp/test_config_ls.out

log_info "Config creation tests passed"
exit 0
