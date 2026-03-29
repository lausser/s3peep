# s3peep Containerized Test Environment

## Overview

This directory contains the strict end-to-end test environment for `s3peep`.

The environment is fully containerized and consists of:

- `minio`: S3-compatible storage used as the backing service
- `s3peep`: the application under test
- `testrunner`: the Bash-based test harness that executes `test_*.sh`

The test suite is designed to be strict.
A test only passes when it proves real `s3peep` behavior with explicit assertions.

## Non-Negotiable Rules

- Do not run `test/run_tests.sh` directly on the host.
- Run the shell suite only inside the `testrunner` container.
- `mc` is infrastructure only. It may set up fixtures or inspect state, but it does not prove `s3peep` behavior.
- A required assertion must never be downgraded to a warning.
- `|| true`, generic HTML matches, redirect tolerance, and “non-empty response means success” are forbidden as proof.
- Every shell test must verify a public `s3peep` surface: `cli`, `http`, or both.

## Prerequisites

- `podman`
- `podman-compose`

No additional host tools are required for the shell suite.

## Correct Way To Run The Suite

Primary one-shot command from the repository root:

```bash
podman-compose -f test/docker-compose.yml up -d --build
podman-compose -f test/docker-compose.yml run --rm testrunner /app/run_tests.sh
```

Primary one-shot command from inside `test/`:

```bash
podman-compose up -d --build
podman-compose run --rm testrunner /app/run_tests.sh
```

Use `run --rm`, not `exec`, as the default documented workflow.
`run --rm` creates a fresh test runner container and does not depend on `testrunner` already being in a running state.

Use `exec` only when you intentionally want to reuse an already-running `testrunner` container.

This is the supported execution path because the suite expects container-only paths and tools such as:

- `/home/s3peep/s3peep`
- `/app/test-config`
- `mc`
- service DNS names `minio` and `s3peep`

## Wrong Way To Run The Suite

Do not do this on the host:

```bash
bash test/run_tests.sh
```

Why it fails:

- the script expects the `s3peep` binary at `/home/s3peep/s3peep`
- the script expects MinIO and `s3peep` to be reachable via container-network hostnames
- the script expects container-local tools and paths

If you run it on the host, preflight should fail immediately. That is correct behavior.

## Common Commands

Start or rebuild the environment:

```bash
cd test
podman-compose up -d --build
```

Run the full shell suite:

```bash
cd test
podman-compose run --rm testrunner /app/run_tests.sh
```

Run preflight only:

```bash
cd test
podman-compose run --rm testrunner /app/run_tests.sh --preflight-only
```

Run one test:

```bash
cd test
podman-compose run --rm testrunner /app/run_tests.sh test_config_add_profile.sh
```

Run several tests:

```bash
cd test
podman-compose run --rm testrunner /app/run_tests.sh test_config_add_profile.sh test_cli_serve_port.sh
```

Stop and clean up:

```bash
cd test
podman-compose down -v
```

## What The Suite Verifies

The current shell suite verifies the documented and currently supported `s3peep` behavior in these areas:

- CLI config creation and profile management
- CLI server startup and `--port` handling
- API bucket listing and bucket selection
- API object listing and file download
- Web UI shell and navigation-driving HTTP contracts
- Security-relevant behavior such as traversal-like keys and XSS-like filenames
- Preflight, fixture setup, and harness strictness

Some tests intentionally capture current semantics even when the product behavior is not ideal yet.
That is still strict testing: the test encodes exact current behavior and fails on drift.

## Current Shell Test Inventory

### CLI and Config

- `test_cli_serve_port.sh`
- `test_config_add_profile.sh`
- `test_config_create.sh`
- `test_config_delete_profile.sh`
- `test_config_invalid_json.sh`
- `test_config_list_profiles.sh`
- `test_config_switch_profile.sh`
- `test_config_unreadable.sh`
- `test_config_unwritable.sh`

### S3 and HTTP API

- `test_connection.sh`
- `test_s3_file_attributes.sh`
- `test_s3_file_download.sh`
- `test_s3_list_buckets.sh`
- `test_s3_list_contents.sh`
- `test_s3_operations.sh`
- `test_s3_select_bucket.sh`

### Security

- `test_security_path_traversal.sh`
- `test_security_xss.sh`

### Web UI Contract Checks

- `test_webui_breadcrumb.sh`
- `test_webui_buckets.sh`
- `test_webui_file_size.sh`
- `test_webui_folder_navigate.sh`
- `test_webui_load.sh`
- `test_webui_select_bucket.sh`

## Browser-Based Verification

Most tests currently validate the HTTP contract directly with `curl` and `jq`.

If browser-level verification is needed for real DOM interaction or rendering behavior, use Chrome DevTools MCP.
Do not replace strict API assertions with vague browser checks.

Use browser-driven checks when you need to prove things like:

- actual DOM updates after interaction
- rendered breadcrumb behavior
- bucket selection via real UI events
- visible text or element states after client-side JavaScript runs

## Fixtures

Fixtures populate MinIO with deterministic test data.

Available scenarios:

| Scenario | Description |
|----------|-------------|
| `default` | Sample files across multiple folders |
| `empty` | Empty bucket with no files |
| `nested` | Deeply nested folder structure |
| `largefiles` | Large files for size-related checks |

Apply a scenario manually:

```bash
cd test
podman-compose exec testrunner /app/fixtures/setup.sh nested
```

## Adding New Tests

Rules for new shell tests:

1. Create a file matching `test_*.sh`
2. Source `helpers.sh`
3. Declare the public surface under test in the file comments or structure
4. Use strict helpers such as:
   - `assert_success`
   - `assert_failure`
   - `assert_equals`
   - `assert_contains`
   - `assert_not_contains`
   - `assert_http_status`
   - `assert_header_contains`
   - `assert_file_contains_json_value`
5. Use `http_request` for HTTP assertions instead of ad hoc `curl` parsing where possible
6. Clean up any created buckets, files, or temporary data
7. Run the test immediately after creating or editing it

Minimum quality bar for a new test:

- it proves real `s3peep` behavior
- it has exact success criteria
- it has at least one meaningful failure-path assertion when applicable
- it does not pass on warnings or loose heuristics

## Go Unit Tests

Go unit tests are separate from the shell suite.

Run them with a Go toolchain available:

```bash
go test ./internal/...
```

If Go is not installed on the host, run them in a suitable build container.

## File Structure

```text
test/
├── docker-compose.yml
├── Dockerfile.s3peep
├── Dockerfile.testrunner
├── run_tests.sh
├── helpers.sh
├── fixtures/
│   ├── setup.sh
│   ├── sample-files/
│   └── scenarios/
├── test_*.sh
└── README.md
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MINIO_ENDPOINT` | `http://minio:9000` | MinIO endpoint inside the container network |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO secret key |
| `S3PEEP_ENDPOINT` | `http://s3peep:8080` | Base URL for the running `s3peep` app inside the container network |
| `S3PEEP_CONFIG_DIR` | `/app/test-config` | Writable config directory inside the testrunner container |
| `S3PEEP_BIN` | `/home/s3peep/s3peep` | `s3peep` binary path inside the testrunner container |

## Troubleshooting

| Problem | What To Check |
|---------|---------------|
| Preflight fails on missing binary | You are probably running outside the `testrunner` container or the image is stale |
| `minio` not reachable | Run `podman-compose ps` and `podman-compose logs minio` |
| `s3peep` not reachable | Run `podman-compose logs s3peep` |
| A new test cannot find services | Confirm it is running inside the `testrunner` container on the `test_default` network |
| `exec testrunner` says container state improper | The `testrunner` container is not running; use `podman-compose run --rm testrunner /app/run_tests.sh` instead |
| Host run of `run_tests.sh` fails immediately | That is expected; use `podman-compose run --rm testrunner /app/run_tests.sh` instead |
