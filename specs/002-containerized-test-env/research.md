# Research: Full Containerized Test Environment

## Decision 1: Keep shell-based end-to-end tests and restrict them to public s3peep interfaces
- Decision: Shell tests validate only public s3peep CLI commands and HTTP endpoints; `mc` remains limited to fixture setup and environment inspection.
- Rationale: The repository already exposes stable CLI (`s3peep init`, `s3peep profile add/list/switch/remove`, `s3peep serve`) and HTTP surfaces (`/api/buckets`, `/api/list`, `/api/get`). `test/TODO.md` shows current false positives come from treating MinIO client commands as proof of s3peep behavior.
- Alternatives considered: Continue using `mc` in PASS assertions; rejected because it only proves MinIO works, not s3peep integration.

## Decision 2: Add mandatory preflight checks before fixture seeding or test discovery
- Decision: The test runner performs a blocking preflight that verifies the s3peep binary is executable, MinIO is reachable and authenticated, the config directory is writable, and required tools (`bash`, `curl`, `jq`, `mc`) are present.
- Rationale: `run_tests.sh` currently waits for services but does not distinguish prerequisite failures from feature failures. The spec now requires preflight failure within 30 seconds with a clear diagnostic.
- Alternatives considered: Let individual tests fail independently; rejected because it produces noisy, misleading failures and violates the new success criteria.

## Decision 3: Redefine PASS semantics around strict observable assertions
- Decision: A shell test may report PASS only when it verifies an explicit s3peep observable such as exit code, HTTP status, JSON field, response body, or config file state.
- Rationale: Existing tests accept warnings, `|| true`, or any non-empty response. That behavior directly contradicts FR-022, FR-025, and `test/TODO.md`.
- Alternatives considered: Keep weak tests as informational; rejected because the feature goal is to restore trust in the suite rather than preserve misleading green output.

## Decision 4: Treat invalid config handling as two classes of behavior
- Decision: Cover both malformed JSON and semantically invalid configuration through CLI/server loading paths, while documenting that missing config still resolves to an empty config by design.
- Rationale: `internal/config/config.go` currently distinguishes missing files from malformed JSON but performs minimal semantic validation. The plan should target real current behaviors and the failure paths already exposed via CLI and server startup.
- Alternatives considered: Require full schema validation during planning; rejected because that would expand scope beyond the clarified spec and current interfaces.

## Decision 5: Use current HTTP/API routes as the contract boundary for S3 and web coverage
- Decision: API regressions focus on `GET /api/buckets`, `POST /api/buckets`, `GET /api/list`, `GET /api/get`, and basic `GET /` UI loading.
- Rationale: These routes are implemented in `internal/handlers/api.go` and are the public surfaces already consumed by the shell tests and web UI.
- Alternatives considered: Add browser automation or new endpoints during planning; rejected because the spec explicitly limits work to existing behavior and shell-based coverage.

## Decision 6: Add security regression coverage for path traversal-like keys and XSS-like filenames
- Decision: The plan requires tests for path traversal rejection and XSS-safe filename handling across already-supported s3peep behaviors, including API responses and download-oriented surfaces.
- Rationale: The spec now calls these out explicitly, and `test/TODO.md` identifies them as critical missing tests. Client-side rendering already escapes names in `web/app.js`, but header and request validation behavior still need explicit coverage and contract definition.
- Alternatives considered: Defer security cases to future hardening; rejected because these are high-risk gaps in currently exposed interfaces.

## Decision 7: Keep fixture reset deterministic and scenario-driven
- Decision: Continue using named fixture scenarios (`default`, `empty`, `nested`, `largefiles`) with explicit reset semantics before mutation-sensitive tests.
- Rationale: The repo already has scenario scripts and a central `fixtures/setup.sh`. The main risk is state leakage, not lack of fixture capability.
- Alternatives considered: Replace fixtures with per-test ad hoc setup; rejected because it would increase duplication and make the suite harder to extend.

## Decision 8: Document permission-sensitive config tests as a known container constraint
- Decision: Plan permission-based config tests around the actual testrunner user model and call out root/non-root behavior as an implementation concern.
- Rationale: `test/TODO.md` already notes that unreadable/unwritable tests may behave differently when containers run as root. This is a real constraint but does not block planning.
- Alternatives considered: Remove permission tests from the plan; rejected because the spec still requires error-handling coverage, and the constraint can be managed explicitly in implementation.
