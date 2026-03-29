#!/bin/bash

# Large files scenario: Create large test files

log_info "Loading largefiles scenario..."

# Create large files (1MB, 5MB, 10MB)
dd if=/dev/zero of=/tmp/large-1mb.bin bs=1M count=1 2>/dev/null
dd if=/dev/zero of=/tmp/large-5mb.bin bs=1M count=5 2>/dev/null
dd if=/dev/zero of=/tmp/large-10mb.bin bs=1M count=10 2>/dev/null

mc cp /tmp/large-1mb.bin local/test-bucket/
mc cp /tmp/large-5mb.bin local/test-bucket/
mc cp /tmp/large-10mb.bin local/test-bucket/

rm -f /tmp/large-*.bin

log_info "Largefiles scenario loaded: 3 files (1MB, 5MB, 10MB)"
