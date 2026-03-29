# Feature Specification: Full Containerized Test Environment

**Feature Branch**: `002-containerized-test-env`  
**Created**: 2026-03-28  
**Status**: Completed  
**Input**: User description: "spec 002: create a full containerized test environment"

## Clarifications

### Session 2026-03-28

- Q: Which S3-compatible service should be used for local testing? → A: MinIO
- Q: How should the test framework be organized? → A: End-to-end tests with compiled s3peep binary interacting with MinIO; Go unit tests use separate `_test.go` files per category + shared helpers package
- Q: What should orchestrate the end-to-end tests? → A: Shell scripts (Bash) that invoke the compiled s3peep binary and verify output
- Q: How should the test runner discover and execute test scripts? → A: File naming convention (`test_*.sh`) with central runner looping over matches
- Q: How should test fixtures define MinIO state? → A: Fixture directory with sample files + shell setup script using MinIO client (`mc`)
- Constraint: No extra software may be installed on the host; all tools (including `mc` and test scripts) MUST run in containers

### Session 2026-03-29

- Q: Should spec 002 fulfill `test/TODO.md` by focusing on existing behavior or by expanding into future features? → A: Fix fake/weak tests plus high-risk missing tests for existing behavior
- Q: What surfaces may shell-based end-to-end tests exercise? → A: Only public s3peep surfaces: CLI commands and HTTP endpoints
- Q: Should high-risk security and failure-path coverage be required now? → A: Yes - require tests for already-supported behavior including path traversal, XSS-safe filename handling, invalid config, and connection failure cases
- Q: What qualifies as a passing shell test? → A: It must prove real s3peep behavior with strict assertions; no PASS may rely on warnings, `|| true`, or any non-empty response checks
- Q: Should the suite run preflight sanity checks before tests? → A: Yes - require mandatory preflight checks before test execution
- Finding: The suite must optimize for ruthless trustworthiness; misleading PASS behavior, weak assertions, and helper-only validation are defects in the feature itself
- Finding: Existing shell tests are still too forgiving where they accept redirects, non-empty responses, warning-only outcomes, helper-generated configs, or MinIO-only checks as proof of s3peep behavior
- Finding: Spec 002 must define a test quality contract, not just a list of test cases, so strictness, isolation, negative coverage, and harness validation become mandatory acceptance requirements

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Local S3 Test Environment (Priority: P1)

A developer starts a local test environment with a single command. The environment includes an S3-compatible storage service pre-loaded with sample data, and the s3peep application connected to it. The developer can immediately open a browser and interact with the file browser using realistic test data.

**Why this priority**: Without a local S3 service, developers cannot test s3peep without access to real AWS resources. This blocks all development, testing, and demo activities.

**Independent Test**: Can be fully tested by running the startup command, verifying all containers are healthy, and confirming the web UI loads with sample data visible.

**Acceptance Scenarios**:

1. **Given** podman-compose is installed, **When** developer runs the startup command, **Then** all environment containers start within 60 seconds and become healthy
2. **Given** the environment is running, **When** developer opens the web UI, **Then** sample files and folders are visible and navigable
3. **Given** the environment is running, **When** developer stops the environment, **Then** all containers stop cleanly without orphaned resources

---

### User Story 2 - Test Data Management (Priority: P2)

A developer can seed the test environment with different datasets to simulate various scenarios: empty bucket, small files, deeply nested folders, large files, and mixed content types.

**Why this priority**: Consistent test data enables reproducible testing and eliminates the need to manually upload files for each test session.

**Independent Test**: Can be fully tested by running each seed scenario and verifying the expected file structure appears in the S3-compatible storage.

**Acceptance Scenarios**:

1. **Given** the environment is running, **When** developer runs the default seed, **Then** a predefined set of folders and files of various types appear in the test bucket
2. **Given** the environment is running, **When** developer runs a seed scenario, **Then** the bucket is reset and populated with only the scenario-specific data

---

### User Story 3 - Automated Test Execution (Priority: P3)

A developer runs end-to-end tests that exercise the compiled s3peep binary against MinIO. Tests cover configuration creation, connection scenarios (success and failure), fixture-based bucket/file setup, access verification, and high-risk regressions for already-supported behavior. Shell tests interact only through public s3peep surfaces (CLI commands and HTTP endpoints), not MinIO-only commands except for environment setup. All test scripts and tooling (including MinIO client) run in containers—no host software installation required. Go unit tests follow the separate `_test.go` files per category pattern with shared helpers for internal logic validation.

**Why this priority**: End-to-end tests validate real s3peep behavior against S3, catching integration issues that unit tests miss. The extensible structure allows adding tests for future features without restructuring.

**Independent Test**: Can be fully tested by running the test command and verifying that the compiled s3peep binary executes operations against MinIO and reports pass/fail results.

**Acceptance Scenarios**:

1. **Given** the environment is running, **When** developer runs the test command, **Then** end-to-end tests execute the compiled s3peep binary against MinIO and report results within 2 minutes
2. **Given** tests have completed, **When** developer reviews output, **Then** clear pass/fail status is shown for each test case
3. **Given** a new feature is implemented, **When** developer adds a new test file following the established pattern, **Then** the test is automatically discovered and executed
4. **Given** an existing shell test reports PASS, **When** its assertions are reviewed, **Then** the PASS result is based on a verified CLI or HTTP behavior and not on warnings, `|| true`, or any non-empty response
5. **Given** the suite starts, **When** required prerequisites are missing, **Then** a preflight sanity check fails fast with a clear diagnostic before any feature tests run
6. **Given** a shell test targets a success path, **When** the test completes, **Then** it asserts exact expected status codes, outputs, JSON fields, headers, or config state rather than broad pattern matches or approximate success indicators
7. **Given** a public CLI command or HTTP endpoint is covered by the suite, **When** coverage is reviewed, **Then** the suite includes required negative-path assertions for invalid input, unsupported methods, or failure conditions relevant to that interface
8. **Given** a test mutates config or fixture state, **When** the next test begins, **Then** state is reset or isolated so no earlier test can cause a false PASS or false FAIL

---

### Edge Cases

- If the required ports are already in use, startup must fail fast with a clear error message naming the conflicting port
- If podman is not running or containers are unreachable, preflight checks must fail before the suite begins with a diagnostic that identifies the missing dependency
- If a seed operation is interrupted mid-way, the reset or seed command must leave the bucket in a known recoverable state and the next reset must succeed
- If containers restart during a test session, the runner must surface a deterministic failure instead of reporting a false PASS
- File names containing HTML or script-like content must be handled without enabling script execution in UI or API consumers
- Invalid object keys or request parameters that resemble path traversal must be rejected without exposing host filesystem data
- If an endpoint returns an unexpected redirect, generic error page, or non-empty body with the wrong status code, the related success-path test must fail
- If a helper script can create or inspect state without exercising s3peep, that helper behavior must not be counted as application coverage
- If a required assertion downgrades a failure to a warning, the test must be treated as invalid until rewritten
- If two tests rely on shared mutable config, bucket contents, or process state, the suite must detect and remove that coupling through reseeding or cleanup
- If the test harness itself is broken (missing binary, broken helper, invalid fixture, missing tool), the suite must fail as an infrastructure defect before feature coverage begins
- If a test claims to validate HTML, JSON, or download behavior, it must assert the contract-relevant structure and not merely the presence of some content

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a single-command startup that launches all required services
- **FR-002**: System MUST include a MinIO container as the local S3-compatible storage service
- **FR-003**: System MUST configure s3peep to connect to the local MinIO service automatically without manual credential setup
- **FR-004**: System MUST pre-populate the test bucket with sample data on startup
- **FR-005**: System MUST provide multiple seed scenarios (default populated, empty bucket, deeply nested, large files)
- **FR-006**: System MUST allow stopping and cleaning up the entire environment with a single command
- **FR-007**: System MUST expose the s3peep web UI on a configurable local port
- **FR-008**: System MUST handle port conflicts gracefully with clear error messages
- **FR-009**: System MUST persist test data across container restarts unless explicitly cleaned
- **FR-010**: System MUST provide a command to reset the environment to its initial state
- **FR-011**: System MUST provide end-to-end shell scripts that execute the compiled s3peep binary against MinIO
- **FR-012**: System MUST include tests for configuration creation (valid and invalid configs)
- **FR-013**: System MUST include tests for connection scenarios (successful connection, connection failure with wrong credentials)
- **FR-014**: System MUST include tests for S3 operations using fixtures (bucket creation, file upload, file access)
- **FR-015**: End-to-end test structure MUST be extensible for adding tests for future features
- **FR-016**: Test runner MUST automatically discover test scripts using a file naming convention (e.g., `test_*.sh`)
- **FR-017**: System MUST provide a `fixtures/` directory containing sample files for test data
- **FR-018**: System MUST provide a fixture setup script that uses the MinIO client (`mc`) running in a container to create buckets and upload files
- **FR-019**: Test scripts and tooling MUST run in containers—no host software installation beyond podman and podman-compose is permitted
- **FR-020**: To fulfill `test/TODO.md`, the system MUST prioritize fixing misleading existing shell tests and adding high-risk missing tests for already-supported behavior rather than requiring tests for not-yet-implemented product features
- **FR-021**: Shell-based end-to-end tests MUST verify s3peep only through its public CLI commands and HTTP endpoints; MinIO client commands MAY be used only for fixture setup or environment inspection, not as proof of s3peep behavior
- **FR-022**: Any shell test that reports PASS MUST use strict assertions tied to expected status codes, response bodies, created config state, or other explicit s3peep outputs; tests MUST NOT pass via `log_warn`, `|| true`, or checks that accept any non-empty response
- **FR-023**: The suite MUST include regression coverage for high-risk already-supported behavior, including invalid config handling, connection failure handling, path traversal rejection, and XSS-safe handling of file names rendered or returned by s3peep
- **FR-024**: The test runner MUST execute a mandatory preflight sanity check before feature tests that verifies the s3peep binary is executable, MinIO is reachable, the config directory is writable, and required in-container tools are available
- **FR-025**: Existing fake or misleading PASS tests MUST be rewritten, removed, or reclassified so that no test result claims to validate s3peep behavior unless it actually exercises s3peep
- **FR-026**: Spec 002 MUST treat misleading PASS results, weak assertions, and helper-only validation as feature defects, not documentation issues or acceptable temporary compromises
- **FR-027**: Every shell test MUST declare and verify one or more explicit public s3peep surfaces under test (`cli` and/or `http`) so the subject under test is unambiguous
- **FR-028**: Success-path shell tests MUST assert exact contract-relevant observables, including exact expected HTTP status codes, exact CLI exit codes, required JSON fields, required response headers, and persisted config or fixture state where applicable
- **FR-029**: Broad or approximate success indicators MUST NOT be accepted, including generic content matches, any non-empty response, broad redirect acceptance, or helper-side state inspection without validating s3peep behavior
- **FR-030**: Every currently supported public CLI command and HTTP endpoint covered by spec 002 MUST include negative-path testing for relevant invalid input, unsupported method, missing required parameter, malformed payload, or upstream failure condition
- **FR-031**: Test helpers MUST fail loudly on unmet expectations and MUST NOT suppress required failures; helper functions that only create, inspect, or transport state MUST be treated as infrastructure, not feature verification
- **FR-032**: The suite MUST validate its own infrastructure, including preflight behavior, fixture setup correctness, helper correctness, test discovery, and test isolation, before claiming end-to-end coverage is trustworthy
- **FR-033**: Mutation-sensitive tests MUST use deterministic isolation or reset behavior for config files, bucket contents, and process state so repeated runs produce the same result from the same starting environment
- **FR-034**: Existing tests that currently accept redirects, warnings, helper-generated configs, or MinIO-only checks as proof of behavior MUST be tightened or replaced with exact assertions over s3peep outputs
- **FR-035**: Web UI and HTTP coverage MUST assert contract-specific structure, including expected HTML markers, content types, status codes, JSON structure, and download headers, rather than only checking for the presence of some response content
- **FR-036**: The suite MUST include explicit regression coverage for strict method handling and invalid-request behavior on currently exposed HTTP endpoints

### Key Entities *(include if feature involves data)*

- **Test Environment**: The complete set of running containers (MinIO, s3peep, test runner) that together provide a functional test setup
- **Seed Scenario**: A predefined collection of buckets, folders, and files that represent a specific test condition
- **Test Bucket**: The S3 bucket used by s3peep during testing, pre-populated with seed data
- **Test Fixture**: A script or configuration that sets up specific MinIO state (buckets, files) before a test runs
- **End-to-End Test**: A test that exercises the compiled s3peep binary against MinIO, verifying real integration behavior
- **Test Runner Container**: A container that executes test scripts and tooling (including MinIO client) without requiring host software installation

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can go from zero to a fully running test environment in under 3 minutes
- **SC-002**: The environment starts with sample data visible in the web UI without any manual configuration
- **SC-003**: All seed scenarios can be applied in under 30 seconds each
- **SC-004**: Environment cleanup removes all containers, volumes, and networks with no orphaned resources
- **SC-005**: The environment runs identically on Linux, macOS, and Windows with podman
- **SC-006**: The automated shell test suite contains no PASS result that is based solely on MinIO client behavior, warning-only assertions, `|| true`, or acceptance of any non-empty response
- **SC-007**: Running the suite against a broken prerequisite state fails in preflight within 30 seconds and reports which prerequisite is missing or unreachable
- **SC-008**: The suite includes passing regression tests for invalid config handling, connection failure handling, path traversal rejection, and XSS-safe filename handling against already-supported s3peep behavior
- **SC-009**: Every shell-based PASS result can be traced to an explicit CLI or HTTP assertion over s3peep, with no ambiguous subject-under-test remaining in the suite
- **SC-010**: Every currently supported CLI command and HTTP endpoint included in spec 002 has at least one passing success-path test and at least one passing negative-path test, unless the interface is read-only and has no meaningful negative-path variant beyond unsupported-method handling
- **SC-011**: Running the full suite twice in succession from the same clean environment produces the same pass/fail result without manual cleanup or hidden shared-state dependencies
- **SC-012**: Intentionally breaking a required prerequisite, fixture, helper, or assertion causes the suite to fail deterministically and visibly rather than continuing with misleading coverage output
- **SC-013**: Web UI, JSON API, and file-download tests assert their contract-relevant structure tightly enough that a generic error page, wrong content type, malformed JSON, or incorrect download header cannot pass as success

## Assumptions

- Developers have only podman and podman-compose installed on their machines; no other host software is required
- All tooling (MinIO client, test scripts, test runner) runs in containers
- The local machine has sufficient resources (memory, CPU, disk) to run the S3 service, s3peep, and test containers simultaneously
- Network access is available to pull container images on first run
- The S3-compatible service API is sufficient for s3peep functionality (no advanced AWS-specific features required)
- Test data files are small enough to be bundled with the project (no external download required)
- Spec 002 fulfills `test/TODO.md` only for tests that validate behavior already exposed by current s3peep CLI or HTTP interfaces; tests for future features remain out of scope for this spec
