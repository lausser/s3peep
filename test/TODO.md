# s3peep Test TODO - Ruthless Coverage Plan

> **Status**: SPEC 002 ALIGNMENT REQUIRED
> **Priority**: CRITICAL
> **Goal**: No mercy. A test only passes when it proves real s3peep behavior through strict, contract-level assertions.

---

## Core Rule

This file is a working checklist for implementing spec 002's strict test strategy.

Non-negotiable rules:
- A PASS is valid only if the test exercised a public s3peep surface (`cli` or `http`) and asserted explicit observables.
- MinIO tooling such as `mc` is infrastructure only. It may prepare or inspect fixtures, but it never proves s3peep behavior.
- `log_warn`, `|| true`, broad grep matches, redirect tolerance, and "any non-empty response" are forbidden for required assertions.
- Weak tests are not "good enough for now". They are defects.
- Test infrastructure failures are coverage failures and must fail early.

---

## PRIORITY 1: ELIMINATE FALSE PASS RESULTS

### 1. Rewrite fake or misleading PASS tests

- [ ] **CRITICAL** Rewrite `test_s3_operations.sh` so every PASS comes from s3peep HTTP endpoints, not `mc ls`, `mc cat`, or `mc stat`
- [ ] **CRITICAL** Rewrite `test_connection.sh` so it validates actual s3peep connection behavior rather than generic `curl` behavior against arbitrary hosts
- [ ] **CRITICAL** Rewrite `test_config_create.sh` so it validates actual s3peep CLI config creation and subsequent config loading, not helper-generated JSON
- [ ] **HIGH** Audit all existing PASS tests and remove any assertion that can pass on non-empty output, generic HTML, or broad text matches
- [ ] **HIGH** Remove all warning-based required assertions; any required check must fail hard
- [ ] **HIGH** Remove redirect-tolerant success assertions unless the redirect is explicitly the contract under test

### 2. Enforce subject-under-test clarity

- [ ] **HIGH** For every shell test, declare the public s3peep surface under test: `cli`, `http`, or both
- [ ] **HIGH** Reject any test that proves only helper behavior, fixture behavior, or MinIO behavior while claiming s3peep coverage
- [ ] **MEDIUM** Rename misleading helper usage patterns so infrastructure helpers cannot be mistaken for product validation

---

## PRIORITY 2: BUILD THE STRICT ASSERTION CONTRACT

### 3. Tighten assertion semantics in helpers and runner

- [ ] **CRITICAL** Add assertion helpers for exact HTTP status, exact content type, exact JSON values, expected headers, and expected exit codes
- [ ] **CRITICAL** Ensure helper functions never swallow command failures for required assertions
- [ ] **HIGH** Replace fuzzy `grep`-style success checks with contract-level assertions where possible
- [ ] **HIGH** Make the runner fail if a test emits warnings for required checks instead of explicit failures
- [ ] **HIGH** Add a runner-level rule that a test cannot report PASS if it never exercised a declared s3peep surface

### 4. Add harness self-validation

- [ ] **CRITICAL** Create a preflight sanity test or preflight mode that verifies binary availability, MinIO reachability, config writability, required tools, and service readiness
- [ ] **HIGH** Add helper validation coverage so broken helpers fail before feature tests can create false confidence
- [ ] **HIGH** Add fixture sanity coverage to verify each named scenario creates the expected bucket/object state
- [ ] **HIGH** Add test-runner validation to verify discovery order, failure propagation, and exit-code behavior

---

## PRIORITY 3: REQUIRE SUCCESS AND NEGATIVE COVERAGE FOR EVERY PUBLIC SURFACE

### 5. CLI coverage

- [ ] **CRITICAL** Implement strict tests for `s3peep profile add` success and failure cases
- [ ] **CRITICAL** Implement strict tests for `s3peep profile list` success and empty-state behavior
- [ ] **CRITICAL** Implement strict tests for `s3peep profile switch` success and failure cases
- [ ] **CRITICAL** Implement strict tests for `s3peep profile remove` success and failure cases
- [ ] **HIGH** Cover `s3peep init` behavior if it is part of the supported public workflow for spec 002
- [ ] **HIGH** For every covered CLI command, assert exact exit code plus required stdout/stderr or config-file state changes

### 6. HTTP coverage

- [ ] **CRITICAL** Implement strict `GET /api/buckets` success and unsupported-method coverage
- [ ] **CRITICAL** Implement strict `POST /api/buckets` success, malformed JSON, missing-field, and invalid-bucket coverage
- [ ] **CRITICAL** Implement strict `GET /api/list` success, empty-state, prefix, and invalid-request coverage
- [ ] **CRITICAL** Implement strict `GET /api/get` success, missing-key, invalid-key, and unsupported-method coverage
- [ ] **HIGH** Implement strict `GET /` coverage for exact status, content type, and required HTML markers
- [ ] **HIGH** For every covered endpoint, assert exact status codes and structured output rather than broad content presence

---

## PRIORITY 4: SECURITY AND FAILURE-PATH REGRESSIONS

### 7. Security cases that must exist now

- [ ] **CRITICAL** Add `test_security_path_traversal.sh` for traversal-like `key` inputs against download/list surfaces
- [ ] **CRITICAL** Add `test_security_xss.sh` for script-like filenames and verify safe handling in rendered and returned surfaces
- [ ] **HIGH** Add coverage for special characters in file names, including quotes, spaces, angle brackets, and URL-sensitive characters
- [ ] **HIGH** Add coverage for bucket-name validation boundaries that matter to current public interfaces

### 8. Failure handling that must exist now

- [ ] **CRITICAL** Add explicit connection-failure coverage for wrong credentials and broken endpoint scenarios through actual s3peep startup or API behavior
- [ ] **HIGH** Add invalid-config coverage for malformed JSON, semantically invalid config, BOM-prefixed files, and null bytes if supported by the environment
- [ ] **HIGH** Add coverage that verifies unexpected error pages, malformed JSON, or wrong headers cannot pass as success
- [ ] **MEDIUM** Add coverage for interrupted or resettable fixture operations to prove recoverability

---

## PRIORITY 5: ISOLATION, DETERMINISM, AND REPEATABILITY

### 9. Remove shared-state flakiness

- [ ] **CRITICAL** Ensure tests that mutate bucket contents reseed or isolate state before the next test
- [ ] **CRITICAL** Ensure tests that mutate config files clean up or isolate config state before the next test
- [ ] **HIGH** Add explicit isolation verification to prove tests do not depend on execution order
- [ ] **HIGH** Verify the suite can run twice in succession from a clean environment with identical results
- [ ] **MEDIUM** Ensure container restarts or partial failures become deterministic suite failures, not hidden flakiness

---

## PRIORITY 6: WEB UI AND DOWNLOAD CONTRACT STRICTNESS

### 10. Tighten UI/document/download checks

- [ ] **HIGH** Tighten `test_webui_load.sh` so it asserts exact required HTML markers, not broad HTML-like content
- [ ] **HIGH** Add strict static asset checks for CSS and JavaScript responses where those assets are part of the supported UI contract
- [ ] **HIGH** Ensure web UI tests do not claim DOM behavior coverage when they only prove API behavior
- [ ] **HIGH** Tighten download tests to assert exact `Content-Type`, `Content-Disposition`, and returned content bytes where relevant
- [ ] **MEDIUM** Add file-size formatting edge cases only if the UI contract is actually rendered and verifiable through current test tooling

---

## PRIORITY 7: DOCUMENT THE QUALITY BAR IN THE TEST SUITE ITSELF

### 11. Make strictness visible and enforceable

- [ ] **HIGH** Update `test/README.md` so it states the ruthless PASS policy and bans fake coverage patterns
- [ ] **HIGH** Document which existing tests are infrastructure checks versus actual product verification
- [ ] **HIGH** Add a short checklist template for new tests: declared surface, success path, negative path, exact assertions, cleanup/isolation
- [ ] **MEDIUM** Record known unsupported or future-feature tests separately so they do not dilute spec 002's current quality bar

---

## Done Means

Spec 002 is not done when there are "more tests".

It is done when:
- No existing PASS is fake, forgiving, or ambiguous.
- Every claimed s3peep test proves s3peep behavior.
- Required CLI and HTTP surfaces have both success and failure-path coverage.
- Infrastructure defects fail before feature coverage starts.
- Repeated runs are deterministic.
- Generic error pages, warnings, redirects, helper output, and MinIO-only behavior can no longer masquerade as success.
