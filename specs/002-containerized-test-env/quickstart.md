# Quickstart: Full Containerized Test Environment

## Prerequisites

- `podman`
- `podman-compose`

No additional host tools are required. All test tooling runs inside containers.

## Start the environment

```bash
cd /home/lausser/git/s3peep/test
podman-compose up -d --build
```

Expected result:
- `minio`, `s3peep`, and `testrunner` containers start successfully
- MinIO becomes reachable at the internal service address `http://minio:9000`
- The s3peep web UI is exposed on the mapped host port, default `http://localhost:8080`

## Run preflight checks

```bash
cd /home/lausser/git/s3peep/test
podman-compose exec testrunner /app/run_tests.sh --preflight-only
```

Expected result:
- The runner verifies the s3peep binary, MinIO reachability, writable config path, and required in-container tools
- If a prerequisite is broken, the command fails within 30 seconds with a clear diagnostic

## Run the full suite

```bash
cd /home/lausser/git/s3peep/test
podman-compose exec testrunner /app/run_tests.sh
```

Expected result:
- Preflight passes before any fixture seeding or feature test execution
- Default fixtures are applied
- All discovered `test_*.sh` files run and report strict PASS/FAIL results

## Run a single test

```bash
cd /home/lausser/git/s3peep/test
podman-compose exec testrunner /app/run_tests.sh test_connection.sh
```

Use this to debug one CLI or HTTP behavior without running the full suite.

## Apply a fixture scenario

```bash
cd /home/lausser/git/s3peep/test
podman-compose exec testrunner /app/fixtures/setup.sh nested
```

Available scenarios:
- `default`
- `empty`
- `nested`
- `largefiles`

Fixture scripts may use `mc` for setup and inspection, but feature tests must validate s3peep through its public CLI and HTTP interfaces.

## Stop and clean up

```bash
cd /home/lausser/git/s3peep/test
podman-compose down -v
```

Expected result:
- Containers stop cleanly
- Volumes and networks created for the environment are removed

## Failure examples

- Port conflict: startup fails and identifies the conflicting local port
- MinIO unreachable: preflight fails before running feature tests
- Config path not writable: preflight fails with a writable-directory diagnostic
- Missing in-container tool: preflight fails naming the missing executable
