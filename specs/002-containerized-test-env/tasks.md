# Tasks: Full Containerized Test Environment

**Feature**: Full Containerized Test Environment  
**Branch**: `002-containerized-test-env`  
**Input**: Refined `spec.md`, `plan.md`, `data-model.md`, contracts, and `test/TODO.md`

## Phase 1: Preserve Existing Environment Foundations

These tasks recognize work that already exists and identify what remains usable as a base.

- [X] T001 Keep the `test/` directory structure and container-oriented layout in place
- [X] T002 Keep `test/docker-compose.yml` as the starting orchestration for `minio`, `s3peep`, and `testrunner`
- [X] T003 Keep `test/Dockerfile.s3peep` and `test/Dockerfile.testrunner` as the base images for the environment
- [X] T004 Keep scenario-based fixture structure under `test/fixtures/`
- [X] T005 Keep `test/run_tests.sh`, `test/helpers.sh`, and existing `test_*.sh` files as the harness to be repaired rather than replaced wholesale

## Phase 2: Trustworthiness Gates (Blocking)

These tasks must be completed before any existing PASS result can be trusted.

- [X] T006 Add a mandatory preflight mode to `test/run_tests.sh` that fails before fixture setup or feature execution when prerequisites are broken
- [X] T007 Add preflight checks for executable s3peep binary, MinIO reachability, config-directory writability, required in-container tools, and service readiness
- [X] T008 Harden `test/helpers.sh` so required assertions cannot be downgraded by `|| true`, warning-only outcomes, or broad success matches
- [X] T009 Add exact assertion helpers for HTTP status, content type, JSON fields, headers, and CLI exit codes
- [X] T010 Add harness validation coverage for helper behavior, fixture behavior, test discovery order, and failure propagation
- [X] T011 Add test-isolation rules for config state, bucket state, and process state so repeated runs are deterministic

## Phase 3: Eliminate False PASS Results (P1)

**Goal**: No existing PASS remains fake, weak, ambiguous, or helper-only.

**Independent Test**: Audit all current PASS tests and verify each one proves a public s3peep CLI or HTTP behavior through exact assertions.

- [X] T012 [P1] Rewrite `test/test_s3_operations.sh` so every PASS comes from s3peep HTTP behavior, not `mc`
- [X] T013 [P1] Rewrite `test/test_connection.sh` so it validates real s3peep connection behavior and fails on wrong status, wrong output, or warning-based logic
- [X] T014 [P1] Rewrite `test/test_config_create.sh` so it validates actual s3peep CLI config creation and subsequent load/use behavior
- [X] T015 [P1] Audit all existing PASS shell tests and remove broad matches, redirect tolerance, non-empty-response acceptance, and warning-based required assertions
- [X] T016 [P1] Mark or remove any test that cannot clearly declare a public s3peep surface under test

## Phase 4: Strict CLI Coverage (P1)

**Goal**: Every currently supported CLI surface included in spec 002 has exact success and negative-path coverage.

**Independent Test**: Run CLI-focused tests and verify exact exit codes, stdout/stderr expectations, and config-file state changes.

- [X] T017 [P1] Tighten `test/test_config_add_profile.sh` for exact success and failure assertions
- [X] T018 [P1] Tighten `test/test_config_list_profiles.sh` for exact empty-state and populated-state assertions
- [X] T019 [P1] Tighten `test/test_config_switch_profile.sh` for exact success and failure assertions
- [X] T020 [P1] Tighten `test/test_config_delete_profile.sh` for exact success and failure assertions
- [X] T021 [P1] Decide whether `s3peep init` is in scope for strict coverage and add or tighten tests accordingly
- [X] T022 [P1] Ensure every CLI test declares the `cli` surface and verifies exact exit code plus output or persisted config state

## Phase 5: Strict HTTP and Web Coverage (P1)

**Goal**: Every currently supported HTTP surface included in spec 002 has exact success and negative-path coverage.

**Independent Test**: Run HTTP-focused tests and verify exact status codes, JSON structure, HTML markers, content types, and download headers.

- [X] T023 [P1] Tighten `test/test_s3_list_buckets.sh` for exact `GET /api/buckets` success and unsupported-method coverage
- [X] T024 [P1] Tighten `test/test_s3_select_bucket.sh` for exact `POST /api/buckets` success, malformed JSON, missing-field, and invalid-bucket coverage
- [X] T025 [P1] Tighten `test/test_s3_list_contents.sh` for exact `GET /api/list` success, empty-state, prefix, and invalid-request coverage
- [X] T026 [P1] Tighten `test/test_s3_file_download.sh` for exact `GET /api/get` success, missing-key, invalid-key, unsupported-method, and header coverage
- [X] T027 [P1] Tighten `test/test_webui_load.sh` for exact `GET /` status, content type, and required HTML marker coverage
- [X] T028 [P1] Audit `test/test_webui_buckets.sh`, `test/test_webui_select_bucket.sh`, `test/test_webui_file_size.sh`, `test/test_webui_folder_navigate.sh`, and `test/test_webui_breadcrumb.sh` so they do not claim DOM behavior when only API behavior is validated
- [X] T029 [P1] Ensure every HTTP-facing test declares the `http` surface and asserts exact contract-level observables

## Phase 6: Security and Failure-Path Regressions (P1)

**Goal**: High-risk already-supported behaviors are covered with strict regressions.

**Independent Test**: Intentionally provide malicious or invalid inputs and verify deterministic rejection or safe handling through s3peep outputs.

- [X] T030 [P1] Add `test/test_security_path_traversal.sh` for traversal-like `key` inputs against supported list/download flows
- [X] T031 [P1] Add `test/test_security_xss.sh` for script-like filenames and verify safe handling in returned or rendered surfaces
- [X] T032 [P1] Add special-character filename coverage for quotes, spaces, angle brackets, and URL-sensitive characters
- [X] T033 [P1] Add explicit connection-failure coverage for wrong credentials and broken endpoint scenarios through real s3peep behavior
- [X] T034 [P1] Tighten `test/test_config_invalid_json.sh` and related config-error tests for malformed JSON, semantically invalid config, and other strict invalid-config cases supported by the environment
- [X] T035 [P1] Add regression checks proving generic error pages, malformed JSON, or incorrect headers cannot pass as success

## Phase 7: Isolation, Repeatability, and Fixture Integrity (P2)

**Goal**: The suite produces the same result from the same starting state and cannot hide order-dependent behavior.

**Independent Test**: Run the full suite twice from a clean environment and verify identical results with no hidden shared-state dependencies.

- [X] T036 [P2] Add explicit config-state cleanup or isolation between mutation-sensitive tests
- [X] T037 [P2] Add explicit bucket reseeding or isolated bucket usage between mutation-sensitive tests
- [X] T038 [P2] Add isolation verification to prove tests do not depend on execution order
- [X] T039 [P2] Verify each fixture scenario remains correct, deterministic, and recoverable after interruption or partial failure
- [X] T040 [P2] Verify the full suite can run twice in succession with identical results

## Phase 8: Documentation and Quality Bar Enforcement (P2)

**Goal**: The ruthless quality bar is visible in the repo and guides future test additions.

**Independent Test**: Review documentation and confirm a new contributor can tell what counts as real coverage versus infrastructure checks.

- [X] T041 [P2] Rewrite `test/README.md` to document the ruthless PASS policy and banned fake-coverage patterns
- [X] T042 [P2] Document which checks are infrastructure validation versus product verification
- [X] T043 [P2] Add a lightweight template or checklist for new tests covering declared surface, success path, negative path, exact assertions, and cleanup/isolation
- [X] T044 [P2] Separate future-feature or unsupported tests from spec 002 deliverables so they do not dilute the current quality bar

## Phase 9: Cross-Cutting Verification

- [X] T045 Run preflight-only mode and verify it fails within 30 seconds with clear diagnostics when a prerequisite is intentionally broken
- [X] T046 Run the strict shell suite and verify no PASS depends on `mc`, warnings, redirects, or non-empty responses
- [X] T047 Verify every currently supported CLI and HTTP surface in scope has at least one success-path and one negative-path test
- [X] T048 Verify web UI, JSON API, and download tests reject generic error pages, wrong content types, malformed JSON, and incorrect headers as success
- [X] T049 Verify the environment still starts within spec limits and cleanup removes containers, networks, and volumes without orphaned resources

## Results

- Spec 002 is complete and verified through the containerized one-shot workflow documented in `test/README.md`
- The shell suite now enforces strict PASS criteria, mandatory preflight, deterministic fixture setup, and exact CLI/HTTP assertions
- CLI contract bugs and S3 compatibility issues found during strict testing were fixed in production code
- Package-level unit tests were added and repaired so `internal/config`, `internal/handlers`, and `internal/s3` are all covered and passing
- Lessons learned were recorded in `docs/lessons-learned/002-containerized-test-env.md`

## Dependencies

```text
Phase 1 -> Phase 2 -> Phase 3
                  -> Phase 4
                  -> Phase 5
Phase 3, 4, 5 -> Phase 6 -> Phase 7 -> Phase 8 -> Phase 9
```

- Phase 2 blocks all trustworthy coverage work
- Phase 3 must finish before any legacy PASS result can be accepted
- Phases 4 and 5 can proceed in parallel once Phase 2 is complete
- Phase 6 depends on hardened helpers plus stable CLI/HTTP test patterns from Phases 4 and 5
- Phase 7 depends on mutation-sensitive tests existing
- Phase 8 can begin once the enforcement rules are settled, but should finish before final verification
- Phase 9 depends on all earlier phases

## Parallel Execution Opportunities

- T006 and T007 can progress together on runner and preflight internals
- T008 and T009 can progress together inside `test/helpers.sh`
- T012, T013, and T014 can run in parallel because they target different legacy false-PASS tests
- T017 through T020 can run in parallel as separate CLI test files
- T023 through T027 can run in parallel as separate HTTP/UI test files
- T030 through T035 can be split across separate security and failure-path test files after strict helper patterns are in place
- T036 and T037 can run in parallel for config and bucket isolation
- T041 through T044 can run in parallel as documentation updates

## Implementation Strategy

**First repair slice**
- Complete Phase 2 and Phase 3 first
- Outcome: existing green tests become trustworthy or are removed/reclassified

**Second repair slice**
- Complete Phases 4 and 5
- Outcome: strict success and negative-path coverage for current public CLI and HTTP surfaces

**Third repair slice**
- Complete Phases 6 and 7
- Outcome: high-risk regressions, isolation, and repeatability are enforced

**Finalization slice**
- Complete Phases 8 and 9
- Outcome: documentation, enforcement, and verification all match the stricter spec 002 definition of done
