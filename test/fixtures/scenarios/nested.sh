#!/bin/bash

# Nested scenario: Create deeply nested folder structure

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SAMPLE_DIR="${SCRIPT_DIR}/../sample-files"

log_info "Loading nested scenario..."

# Create deeply nested structure
mc mb local/test-bucket/level1/level2/level3/level4/level5 2>/dev/null || true

# Add files at different levels
mc cp "${SAMPLE_DIR}/hello.txt" local/test-bucket/
mc cp "${SAMPLE_DIR}/hello.txt" local/test-bucket/level1/
mc cp "${SAMPLE_DIR}/hello.txt" local/test-bucket/level1/level2/
mc cp "${SAMPLE_DIR}/hello.txt" local/test-bucket/level1/level2/level3/
mc cp "${SAMPLE_DIR}/hello.txt" local/test-bucket/level1/level2/level3/level4/
mc cp "${SAMPLE_DIR}/hello.txt" local/test-bucket/level1/level2/level3/level4/level5/

log_info "Nested scenario loaded: 6 files across 6 levels"
