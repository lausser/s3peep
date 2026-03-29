#!/bin/bash

# Shared test helper functions for s3peep test environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Environment variables (set by docker-compose)
MINIO_ENDPOINT="${MINIO_ENDPOINT:-http://minio:9000}"
MINIO_ACCESS_KEY="${MINIO_ACCESS_KEY:-minioadmin}"
MINIO_SECRET_KEY="${MINIO_SECRET_KEY:-minioadmin}"
S3PEEP_ENDPOINT="${S3PEEP_ENDPOINT:-http://s3peep:8080}"
S3PEEP_CONFIG_DIR="${S3PEEP_CONFIG_DIR:-/app/test-config}"
S3PEEP_CONFIG="${S3PEEP_CONFIG_DIR}/config.json"
S3PEEP_BIN="${S3PEEP_BIN:-/home/s3peep/s3peep}"

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Wait for MinIO to be ready
wait_for_minio() {
    local max_attempts="${1:-30}"
    local attempt=0
    
    log_info "Waiting for MinIO to be ready at ${MINIO_ENDPOINT}..."
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -sf "${MINIO_ENDPOINT}/minio/health/live" > /dev/null 2>&1; then
            log_info "MinIO is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done
    
    log_error "MinIO did not become ready within ${max_attempts} seconds"
    return 1
}

# Wait for s3peep to be ready
wait_for_s3peep() {
    local max_attempts="${1:-30}"
    local attempt=0
    
    log_info "Waiting for s3peep to be ready at ${S3PEEP_ENDPOINT}..."
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -sf "${S3PEEP_ENDPOINT}/" > /dev/null 2>&1; then
            log_info "s3peep is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done
    
    log_error "s3peep did not become ready within ${max_attempts} seconds"
    return 1
}

# Configure mc alias for MinIO
configure_mc() {
    mc alias set local "${MINIO_ENDPOINT}" "${MINIO_ACCESS_KEY}" "${MINIO_SECRET_KEY}" > /dev/null 2>&1
}

# Assert functions
assert_equals() {
    local actual="$1"
    local expected="$2"
    local message="${3:-Values should be equal}"
    
    if [ "$actual" = "$expected" ]; then
        return 0
    else
        log_error "$message"
        log_error "  Expected: $expected"
        log_error "  Actual: $actual"
        return 1
    fi
}

assert_success() {
    local exit_code="$1"
    local message="${2:-Command should succeed}"
    
    if [ "$exit_code" -eq 0 ]; then
        return 0
    else
        log_error "$message (exit code: $exit_code)"
        return 1
    fi
}

assert_failure() {
    local exit_code="$1"
    local message="${2:-Command should fail}"
    
    if [ "$exit_code" -ne 0 ]; then
        return 0
    else
        log_error "$message (expected failure but got success)"
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-String should contain substring}"
    
    if echo "$haystack" | grep -q "$needle"; then
        return 0
    else
        log_error "$message"
        log_error "  Looking for: $needle"
        return 1
    fi
}

assert_not_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-String should not contain substring}"

    if echo "$haystack" | grep -q "$needle"; then
        log_error "$message"
        log_error "  Unexpected: $needle"
        return 1
    fi
    return 0
}

assert_http_status() {
    local actual="$1"
    local expected="$2"
    local message="${3:-Unexpected HTTP status}"
    assert_equals "$actual" "$expected" "$message"
}

assert_content_type_contains() {
    local actual="$1"
    local expected="$2"
    local message="${3:-Unexpected content type}"

    if [[ "$actual" == *"$expected"* ]]; then
        return 0
    fi

    log_error "$message"
    log_error "  Expected content type containing: $expected"
    log_error "  Actual content type: $actual"
    return 1
}

assert_file_contains_json_value() {
    local filepath="$1"
    local jq_filter="$2"
    local expected="$3"
    local message="${4:-Unexpected JSON value}"
    local actual

    actual=$(jq -r "$jq_filter" "$filepath") || return 1
    assert_equals "$actual" "$expected" "$message"
}

assert_header_contains() {
    local headers="$1"
    local expected="$2"
    local message="${3:-Expected header not found}"
    assert_contains "$headers" "$expected" "$message"
}

fail_if_warning_output() {
    local output="$1"
    local message="${2:-Warnings are not allowed in strict tests}"

    if echo "$output" | grep -q "\[WARN\]"; then
        log_error "$message"
        return 1
    fi
    return 0
}

assert_file_exists() {
    local filepath="$1"
    local message="${2:-File should exist}"
    
    if [ -f "$filepath" ]; then
        return 0
    else
        log_error "$message: $filepath"
        return 1
    fi
}

# Generate s3peep config with MinIO credentials
generate_s3peep_config() {
    local config_path="${1:-${S3PEEP_CONFIG}}"
    local profile_name="${2:-test}"
    local bucket="${3:-test-bucket}"
    local endpoint="${4:-${MINIO_ENDPOINT}}"
    
    mkdir -p "$(dirname "$config_path")"
    
    cat > "$config_path" <<EOF
{
  "active_profile": "${profile_name}",
  "profiles": [
    {
      "name": "${profile_name}",
      "region": "us-east-1",
      "access_key_id": "${MINIO_ACCESS_KEY}",
      "secret_access_key": "${MINIO_SECRET_KEY}",
      "bucket": "${bucket}",
      "endpoint_url": "${endpoint}"
    }
  ]
}
EOF
}

# Generate invalid s3peep config (wrong credentials)
generate_invalid_config() {
    local config_path="${1:-${S3PEEP_CONFIG}}"
    local profile_name="${2:-test}"
    local bucket="${3:-test-bucket}"
    
    mkdir -p "$(dirname "$config_path")"
    
    cat > "$config_path" <<EOF
{
  "active_profile": "${profile_name}",
  "profiles": [
    {
      "name": "${profile_name}",
      "region": "us-east-1",
      "access_key_id": "wrongkey",
      "secret_access_key": "wrongsecret",
      "bucket": "${bucket}",
      "endpoint_url": "${MINIO_ENDPOINT}"
    }
  ]
}
EOF
}

# Run s3peep command and capture output/exit code
run_s3peep() {
    local args="$*"
    local output
    local exit_code
    
    output=$(curl -sf "${S3PEEP_ENDPOINT}/${args}" 2>&1)
    exit_code=$?
    
    echo "$output"
    return $exit_code
}

# Run s3peep CLI command via container exec (if available)
run_s3peep_cli() {
    local config_path="${S3PEEP_CONFIG}"
    CONFIG="$config_path" "$S3PEEP_BIN" "$@"
}

http_request() {
    local method="$1"
    local url="$2"
    local body="${3:-}"
    local response_body
    local response_headers
    local status_code
    local tmp_body
    local tmp_headers

    tmp_body=$(mktemp)
    tmp_headers=$(mktemp)

    if [ -n "$body" ]; then
        status_code=$(curl -sS -X "$method" -D "$tmp_headers" -o "$tmp_body" -w "%{http_code}" -H "Content-Type: application/json" --data "$body" "$url")
    else
        status_code=$(curl -sS -X "$method" -D "$tmp_headers" -o "$tmp_body" -w "%{http_code}" "$url")
    fi

    response_body=$(cat "$tmp_body")
    response_headers=$(cat "$tmp_headers")
    rm -f "$tmp_body" "$tmp_headers"

    HTTP_STATUS="$status_code"
    HTTP_BODY="$response_body"
    HTTP_HEADERS="$response_headers"
}

preflight_check() {
    log_info "Running preflight checks..."

    command -v bash >/dev/null 2>&1 || { log_fail "Missing required tool: bash"; return 1; }
    command -v curl >/dev/null 2>&1 || { log_fail "Missing required tool: curl"; return 1; }
    command -v jq >/dev/null 2>&1 || { log_fail "Missing required tool: jq"; return 1; }
    command -v mc >/dev/null 2>&1 || { log_fail "Missing required tool: mc"; return 1; }

    if [ ! -x "$S3PEEP_BIN" ]; then
        log_fail "s3peep binary is missing or not executable: $S3PEEP_BIN"
        return 1
    fi

    mkdir -p "$S3PEEP_CONFIG_DIR" || { log_fail "Config directory could not be created: $S3PEEP_CONFIG_DIR"; return 1; }
    test_file="$S3PEEP_CONFIG_DIR/.preflight-write-test"
    if ! touch "$test_file" 2>/dev/null; then
        log_fail "Config directory is not writable: $S3PEEP_CONFIG_DIR"
        return 1
    fi
    rm -f "$test_file"

    wait_for_minio || return 1
    configure_mc || { log_fail "Failed to configure MinIO client alias"; return 1; }
    mc ls local/ >/dev/null 2>&1 || { log_fail "MinIO is not reachable with configured credentials"; return 1; }
    wait_for_s3peep || return 1

    log_info "Preflight checks passed"
    return 0
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test resources..."
    # Remove test config if it exists
    rm -f "${S3PEEP_CONFIG}"
}

# Run a single test file
run_test() {
    local test_file="$1"
    local test_name=$(basename "$test_file")
    
    if [ ! -f "$test_file" ]; then
        log_error "Test file not found: $test_file"
        return 1
    fi
    
    if [ ! -x "$test_file" ]; then
        chmod +x "$test_file"
    fi
    
    local start_time=$(date +%s)
    
    if "$test_file"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log_pass "$test_name (${duration}s)"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log_fail "$test_name (${duration}s)"
        return 1
    fi
}
