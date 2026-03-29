#!/bin/bash

# Default scenario: Populate test bucket with sample files

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SAMPLE_DIR="${SCRIPT_DIR}/../sample-files"

log_info "Loading default scenario..."

# Create folder structure
mc mb local/test-bucket/documents 2>/dev/null || true
mc mb local/test-bucket/images 2>/dev/null || true
mc mb local/test-bucket/data 2>/dev/null || true

# Upload sample files
mc cp "${SAMPLE_DIR}/hello.txt" local/test-bucket/documents/
mc cp "${SAMPLE_DIR}/readme.md" local/test-bucket/documents/
mc cp "${SAMPLE_DIR}/config.json" local/test-bucket/data/
mc cp "${SAMPLE_DIR}/data.csv" local/test-bucket/data/

log_info "Default scenario loaded: 4 files across 3 folders"
