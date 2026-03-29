# Implementation Plan: Full Containerized Test Environment

**Branch**: `002-containerized-test-env` | **Date**: 2026-03-29 | **Spec**: `/home/lausser/git/s3peep/specs/002-containerized-test-env/spec.md`
**Input**: Feature specification from `/home/lausser/git/s3peep/specs/002-containerized-test-env/spec.md`

## Summary

Restore trust in the containerized test environment by replacing fake or weak PASS shell tests with strict end-to-end coverage over public s3peep CLI and HTTP interfaces, adding mandatory preflight sanity checks, and documenting fixtures/contracts so the MinIO-backed podman-compose environment is reliable, diagnosable, and extensible.

## Technical Context

**Language/Version**: Go 1.24 for `s3peep`; Bash for containerized shell tests  
**Primary Dependencies**: aws-sdk-go-v2, embedded stdlib HTTP server, podman/podman-compose, MinIO, MinIO client `mc`, `curl`, `jq`  
**Storage**: File-based JSON config at `~/.config/s3peep/config.json`; MinIO object storage with persistent test volume  
**Testing**: Go `testing` package for unit tests; Bash `test_*.sh` end-to-end suite in containers  
**Target Platform**: Linux/macOS/Windows developer machines running podman and podman-compose; Linux containers at runtime  
**Project Type**: Go CLI + embedded web server with containerized integration test harness  
**Performance Goals**: Environment startup within 3 minutes; shell suite completion within 2 minutes; preflight failure within 30 seconds when prerequisites are broken  
**Constraints**: No host software beyond podman and podman-compose; shell PASS results must come only from strict assertions over public CLI/HTTP behavior; MinIO client may be used for fixture setup/inspection only  
**Scale/Scope**: One local test environment with services `minio`, `s3peep`, and `testrunner`; regression coverage limited to currently exposed CLI and HTTP interfaces, including high-risk security and failure paths

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Correctness over cleverness: PASS - The plan centers on replacing misleading assertions with explicit verification and keeping MinIO as setup-only infrastructure.
- Smallest change that works: PASS - Scope is limited to spec 002 artifacts and tests for already-supported behavior; future product features remain out of scope.
- Leverage existing patterns: PASS - Reuses the existing `test/` harness, fixture scenarios, Go unit tests, and current CLI/API surfaces.
- Prove it works: PASS - Plan includes preflight validation, targeted regression coverage, and containerized suite verification criteria.
- Explicit uncertainty: PASS - No unresolved technical-context items remain; permission-sensitive config tests are tracked as a design constraint, not a blocker.

## Project Structure

### Documentation (this feature)

```text
specs/002-containerized-test-env/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── cli.md
│   ├── fixtures.md
│   ├── http.md
│   ├── pass-criteria.md
│   ├── preflight.md
│   ├── security.md
│   └── test-runner.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
└── s3peep/
    └── main.go

internal/
├── config/
│   ├── cli.go
│   ├── cli_test.go
│   ├── config.go
│   ├── config_test.go
│   ├── profile.go
│   └── profile_test.go
├── handlers/
│   ├── api.go
│   └── api_test.go
└── s3/
    ├── client.go
    └── client_test.go

web/
├── app.js
├── index.html
└── styles.css

test/
├── docker-compose.yml
├── Dockerfile.s3peep
├── Dockerfile.testrunner
├── helpers.sh
├── fixtures/
│   ├── sample-files/
│   ├── scenarios/
│   └── setup.sh
├── run_tests.sh
├── README.md
├── TODO.md
└── test_*.sh
```

**Structure Decision**: Keep the existing single-project Go layout. Implement spec 002 by strengthening the existing `test/` container harness and its corresponding CLI/API code paths rather than introducing new packages or services.

## Phase 0: Research

### Research Goals

1. Confirm best practices for strict shell-based E2E assertions against a Go CLI/web app backed by MinIO in podman-compose.
2. Define mandatory preflight checks and failure diagnostics for the testrunner.
3. Resolve security and failure-path expectations for invalid config, connection failures, path traversal rejection, and XSS-safe filename handling based on current repo interfaces.
4. Capture fixture/reset and PASS-criteria rules needed to eliminate misleading green results.

### Research Output

- `/home/lausser/git/s3peep/specs/002-containerized-test-env/research.md`

## Phase 1: Design & Contracts

### Design Goals

1. Model the test environment, preflight checks, config states, fixture scenarios, and security cases in `data-model.md`.
2. Define executable contracts for CLI, HTTP, fixtures, preflight, PASS criteria, security expectations, and test-runner behavior under `contracts/`.
3. Produce `quickstart.md` with start, preflight, run, reseed, and cleanup flows aligned with podman-compose usage and failure diagnostics.
4. Update agent context so future implementation work reflects Go, MinIO, podman-compose, and shell-based E2E testing.

### Planned Artifacts

- `/home/lausser/git/s3peep/specs/002-containerized-test-env/data-model.md`
- `/home/lausser/git/s3peep/specs/002-containerized-test-env/quickstart.md`
- `/home/lausser/git/s3peep/specs/002-containerized-test-env/contracts/cli.md`
- `/home/lausser/git/s3peep/specs/002-containerized-test-env/contracts/fixtures.md`
- `/home/lausser/git/s3peep/specs/002-containerized-test-env/contracts/http.md`
- `/home/lausser/git/s3peep/specs/002-containerized-test-env/contracts/pass-criteria.md`
- `/home/lausser/git/s3peep/specs/002-containerized-test-env/contracts/preflight.md`
- `/home/lausser/git/s3peep/specs/002-containerized-test-env/contracts/security.md`
- `/home/lausser/git/s3peep/specs/002-containerized-test-env/contracts/test-runner.md`

## Phase 2: Implementation Planning Approach

1. Add a blocking preflight step to the test runner before fixture seeding or test discovery.
2. Rewrite or reclassify misleading PASS shell tests so they validate only s3peep CLI/HTTP behavior.
3. Add missing high-risk regressions for invalid config, connection failures, path traversal rejection, and XSS-safe filename handling.
4. Keep fixture setup isolated and deterministic so mutated MinIO state cannot create false PASS results.
5. Verify updated documentation and contracts match the real repository layout and public interfaces.

## Post-Design Constitution Check

- Correctness over cleverness: PASS - Contracts require observable CLI/HTTP assertions and eliminate MinIO-only proof.
- Smallest change that works: PASS - Design keeps the current repo layout and narrows new coverage to already-exposed behavior.
- Leverage existing patterns: PASS - Contracts formalize `test_*.sh`, `fixtures/`, and current API endpoints rather than inventing new testing infrastructure.
- Prove it works: PASS - Design requires preflight, deterministic fixtures, strict PASS criteria, and regression coverage tied to spec success criteria.
- Explicit uncertainty: PASS - Remaining risk is documented as implementation detail (notably permission-sensitive filesystem tests in containers), with no unresolved clarification blocking planning.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
