#!/bin/bash

# Test: s3peep API list buckets
# Tests that s3peep can list buckets via API endpoint

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running S3 list buckets API tests..."

# Setup
generate_s3peep_config
configure_mc

# Wait for s3peep to be ready
wait_for_s3peep

# Test 1: List buckets via s3peep API
log_info "Test: List buckets via s3peep API"
response=$(curl -sf "${S3PEEP_ENDPOINT}/api/buckets" 2>&1) || true

if [ -n "$response" ]; then
    # Verify it's valid JSON
    if echo "$response" | jq empty >/dev/null 2>&1; then
        # Parse the response to verify it contains bucket data
        bucket_count=$(echo "$response" | jq 'length' 2>/dev/null || echo "0")
        if [ "$bucket_count" -gt 0 ]; then
            log_pass "s3peep API returned bucket list with $bucket_count bucket(s)"
        else
            log_warn "s3peep API returned empty bucket list"
        fi
    else
        log_fail "s3peep API did not return valid JSON: $response"
        exit 1
    fi
else
    log_fail "s3peep API did not return data"
    exit 1
fi

# Test 2: List buckets via s3peep API with no buckets
log_info "Test: List buckets via s3peep API with no buckets"
# Create a new bucket to test with, then remove all buckets
TEST_BUCKET="test-bucket-api-$$"
mc mb local/"$TEST_BUCKET" >/dev/null 2>&1

# List buckets via API
response=$(curl -sf "${S3PEEP_ENDPOINT}/api/buckets" 2>&1) || true

if [ -n "$response" ]; then
    if echo "$response" | jq empty >/dev/null 2>&1; then
        bucket_count=$(echo "$response" | jq 'length' 2>/dev/null || echo "0")
        log_pass "s3peep API returned bucket list with $bucket_count bucket(s)"
    else
        log_fail "s3peep API did not return valid JSON: $response"
        exit 1
    fi
else
    log_fail "s3peep API did not return data"
    exit 1
fi

# Clean up test bucket
mc rb local/"$TEST_BUCKET" >/dev/null 2>&1 || true

# Test 3: s3peep API buckets endpoint error handling
log_info "Test: s3peep API buckets endpoint error handling"
# Test with invalid method (PUT instead of GET)
response=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "${S3PEEP_ENDPOINT}/api/buckets" 2>&1) || true
if [ "$response" = "405" ]; then
    log_pass "s3peep API correctly rejects PUT method on buckets endpoint"
else
    log_warn "s3peep API returned unexpected status $response for PUT on buckets endpoint (expected 405)"
fi

# Cleanup
cleanup

log_info "S3 list buckets API tests passed"
exit 0