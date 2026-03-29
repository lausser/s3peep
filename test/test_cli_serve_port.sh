#!/bin/bash

# Test: s3peep serve honors --port flag
# Tests the exact documented CLI serve invocation form

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

log_info "Running CLI serve port tests..."

PORT=19090
cleanup_local() {
    if [ -n "${SERVER_PID:-}" ]; then
        kill "${SERVER_PID}" >/dev/null 2>&1 || true
        wait "${SERVER_PID}" 2>/dev/null || true
    fi
    rm -f /tmp/test_cli_serve_port.out /tmp/test_cli_serve_port.err
    cleanup
}
trap cleanup_local EXIT

generate_s3peep_config

# Surface under test: cli

log_info "Test: Exact serve invocation honors requested port"
CONFIG="${S3PEEP_CONFIG}" "${S3PEEP_BIN}" serve --port "${PORT}" > /tmp/test_cli_serve_port.out 2> /tmp/test_cli_serve_port.err &
SERVER_PID=$!

for _ in $(seq 1 20); do
    if curl -fsS "http://127.0.0.1:${PORT}/" >/dev/null 2>&1; then
        break
    fi
    sleep 1
done

assert_contains "$(cat /tmp/test_cli_serve_port.out)" "Starting server on port ${PORT}..." "Serve output must report the requested port" || exit 1
http_request GET "http://127.0.0.1:${PORT}/"
assert_http_status "$HTTP_STATUS" "200" "Serve command must bind and respond on the requested port" || exit 1
log_pass "Serve command honors exact documented --port flag"

log_info "CLI serve port tests passed"
exit 0
