# Lessons Learned: Containerized Test Environment (002-containerized-test-env)

## Why This Work Mattered

Spec 002 started as "create a full containerized test environment," but the real work quickly became something more demanding: create a test system that deserves trust. The original test suite looked active, but much of that activity was misleading. It mixed infrastructure verification with product verification, accepted warnings as effectively good enough, and passed tests on evidence that was too weak to support the confidence implied by a green result. The user was explicit about wanting strictness and "no mercy," and that changed the nature of the work. This was no longer just about adding a MinIO container and a few shell scripts. It was about establishing a testing culture in code, documentation, and workflow.

The most important insight is that a containerized test environment is not valuable because it exists. It is valuable only if it makes false confidence difficult. Once that standard is accepted, many normal shortcuts become unacceptable: broad matching, helper-generated state treated as proof, MinIO-only checks masquerading as application validation, incomplete documentation, and unverified command paths. A strict test environment is a product in its own right. It needs its own design, contracts, failure modes, invariants, and regression coverage.

## The Central Failure Mode: False Confidence

The biggest technical and process problem uncovered during this work was not a single bug in `s3peep`. It was the existence of tests that looked meaningful while proving very little. Several shell tests initially passed because they checked whether something happened somewhere, not whether `s3peep` itself delivered the promised behavior. That difference is the difference between testing and theater.

Three examples stand out:

- Some S3 tests proved that `mc` could list or fetch objects from MinIO. That established that MinIO worked, but not that `s3peep` correctly used it.
- Some HTTP tests passed if the response was non-empty, regardless of whether the response was the correct payload, the correct status code, or even a generic error page.
- Some CLI/config tests created JSON files with helper functions and then validated those files, which proved the helper could write JSON, not that the real `s3peep` CLI command could create or load the file correctly.

The lesson is straightforward: a test must prove the contract of the subject under test, not the helpfulness of nearby infrastructure. The more indirect the evidence, the more likely the test will eventually lie.

## Bash Testing Lessons

### Bash is viable for integration tests, but only with discipline

Bash is completely adequate for black-box CLI and HTTP integration testing when the product already exposes shell-friendly surfaces. In this project, `s3peep` is a CLI and HTTP server, and MinIO plus `curl`, `jq`, and `mc` fit naturally into a shell-based flow. The problem was never the use of Bash itself. The problem was the absence of a rigor framework around Bash.

That rigor framework needed several ingredients:

- explicit assertion helpers rather than hand-written conditionals in every test
- clear PASS/FAIL semantics with no warning-based success path
- consistent handling of `stdout`, `stderr`, exit codes, and HTTP response metadata
- fixture setup separated from feature validation
- immediate failure on required assertion mismatches

Once helpers such as `assert_http_status`, `assert_header_contains`, `assert_file_contains_json_value`, and `http_request` were introduced, Bash stopped being a loose scripting medium and became a structured integration test harness. The lesson is not "Bash is bad" or "Bash is good." The lesson is that Bash without strong helper conventions decays into permissive testing very quickly.

### Weak shell idioms are dangerous

Several shell idioms are especially risky in tests:

- `|| true`
- broad `grep` matching on arbitrary output
- checking only whether output is non-empty
- accepting multiple unrelated status codes for the same success path
- logging a warning for a required assertion instead of failing

Each of these patterns hides ambiguity. Ambiguity in a test is not harmless; it is deferred misinformation. One of the strongest lessons from this work is that shell tests need a defined list of forbidden patterns, not just a list of preferred helpers.

### Tests must be explicit about what surface they validate

An important improvement was the discipline of identifying the public `s3peep` surface under test: `cli`, `http`, or occasionally both. This sounds bureaucratic at first, but it has real value. It forces the author to state, before any assertions are written, what is actually being proved. That prevented several tests from drifting into "MinIO works" or "the helper ran" territory while still looking like application tests.

## Dockerized/Containerized Test Environment Lessons

### Containerization solves reproducibility, but only if the workflow is correct

The environment design was sound in principle: all shell tests and supporting tools should run inside a container, with MinIO and `s3peep` reachable through the same network every time. That gives stable paths, stable hostnames, and consistent dependencies. But containerization is not self-executing. The workflow around it must also be correct.

One major documentation failure made this obvious. The README originally suggested `exec` as the primary way to run the test suite. That works only if the target container is already running and in the right state. In real usage, this produced a failure that the documentation should have prevented. The correct primary flow turned out to be the one-shot `run --rm` invocation. The lesson is that documentation for a containerized test system must prefer the least stateful execution path. If a command depends on ambient container state, it is not a good default.

### The test runner itself is part of the product under test

The test harness had to be treated with the same suspicion as the application. `run_tests.sh` originally had issues around argument filtering and wrapper invocation. At different points it:

- misinterpreted `--preflight-only`
- accepted a wrapper-injected runner path as if it were a test name
- allowed a recursion-like nested full-suite run in a one-shot container flow

These are not cosmetic problems. When the test runner misbehaves, every downstream result is harder to interpret. The lesson is that a test harness must have self-validation and must itself be exercised in the same modes that documentation recommends to users.

### Preflight checks are a first-class feature

Before this work, the environment would often discover prerequisites implicitly while already running tests. That leads to noisy failures and wasted debugging time. Adding mandatory preflight changed the experience dramatically. A proper preflight check needs to verify:

- the `s3peep` binary exists and is executable
- MinIO is reachable and authenticated
- the config directory is writable
- required in-container tools are present
- the application endpoint is reachable when HTTP tests depend on it

The lesson is that preflight is not busywork. It is the border between infrastructure failure and feature failure. If that border is blurry, test output becomes misleading.

## The Importance of Testing the Tests Immediately

This project repeatedly reinforced the same rule: every new or modified test must be run immediately. Not later. Not after several related edits. Not after a batch of changes. Immediately.

There were multiple examples where this rule was violated or only partially followed, and each one caused pain:

- a CLI test passed in the container harness while the exact user-facing local command still failed
- the README recommended a workflow that was only partially verified and broke in a real user context
- the one-shot runner flow initially triggered nested behavior because wrapper arguments were not filtered correctly
- a new strict serve-port test initially failed because it targeted the wrong host in the container context

Each of these failures had the same moral structure: a thing was written, it looked plausible, and it was not verified in the exact form users would rely on. The remedy is not subtle. Every created or updated artifact needs immediate proof:

- new test script -> run that test script immediately
- updated test helper -> rerun at least one directly affected test immediately
- updated runner -> rerun the exact workflow it is supposed to support immediately
- updated README command -> execute the exact command immediately

The phrase "test the tests" sounds redundant until a test suite begins producing green results for broken behavior. At that point it becomes obvious that testing the tests is not optional. It is the difference between a trustworthy quality system and a performative one.

## The Hard Lesson About Not Faking Tests

The user explicitly asked for strictness and no mercy. The most uncomfortable but valuable lesson from this work is that false confidence can come not only from the repository's existing tests, but also from the agent improving them. At one point a CLI contract was declared effectively tested, but the exact documented local invocation still failed. That was a serious process error.

The problem was not malicious intent. The problem was accepting an indirect verification path as if it were equivalent to the actual product contract. A test that only proves the helper path is not a strict test for the public CLI. A documentation command that is not executed is not validated documentation. A runner mode that is not exercised through the real wrapper path is not verified infrastructure.

The lesson is severe but clear: if a user-facing behavior is claimed to be tested, the exact user-facing behavior must have been exercised. Not something nearby. Not something structurally similar. The exact thing.

This should become a standing rule for future work in the project:

- if the contract is a CLI command, run that exact command form
- if the contract is a documented README command, run that exact command form
- if the contract is a browser interaction, perform or simulate that interaction at the right layer
- if the contract is a networked containerized workflow, run it through the real container orchestration path

Anything less should be described as partial verification, not verification.

## Bookkeeping: What Is Ready To Test vs What Must Be Tracked

This project also highlighted the importance of bookkeeping. A test plan cannot simply be a wishlist. It must distinguish among several states:

- behavior already implemented and ready for strict testing
- behavior implemented but currently too permissive, so tests must encode current semantics precisely
- behavior not yet implemented, which should be documented as future test debt rather than faked into existence
- infrastructure capabilities that enable testing, such as fixture scenarios or browser tooling

The refined `spec.md`, `tasks.md`, and `test/TODO.md` became useful because they separated these categories. For example:

- some endpoints were strict enough to demand exact negative-path tests immediately
- some behaviors were not ideal but existed, so tests needed to capture the current semantics without pretending the product already behaved better
- some future UI or shutdown scenarios remained outside current implementation and therefore should remain explicit future work rather than becoming vague TODOs hidden in passing summaries

The lesson is that bookkeeping is part of engineering quality. A strict suite requires a strict map of reality. If the map does not distinguish "implemented and tested," "implemented but not ideal," and "not yet implemented," the suite and documentation will drift out of sync.

## Product Bugs Found Because Strict Tests Improved

Stricter tests did not just improve confidence in existing behavior; they found real defects.

### MinIO path-style addressing bug

The stricter S3 operations tests exposed that bucketed object requests were using a virtual-host-style pattern that did not work for the local MinIO setup. This led to a product fix in `internal/s3/client.go`, where path-style addressing had to be enabled for custom endpoints. Without the stricter tests, the environment could have remained "mostly green" while core object operations were actually fragile or broken.

### CLI profile argument handling bug

The documented `./s3peep profile add --name ...` command failed because `main.go` and `internal/config/cli.go` disagreed about argument shape. This was a direct contract bug. The fact that it survived until explicitly exercised by the user is a reminder that CLI contracts must be tested with exact invocation forms.

### `serve --port` parsing bug

The `serve` command printed and used the default port even when `--port` was supplied. This was another direct CLI contract bug, and another example of why a minimal "server starts" test is insufficient. A strict CLI test must verify that the configured port is the one actually bound and reported.

### README workflow bug

The documentation suggested a test execution mode that failed when the runner container was not already in the correct state. This was a usability bug in the test system itself. It reinforced the lesson that documentation is executable product surface and must be tested as such.

### Unit-test architecture and mocking gaps

The later unit-test repairs exposed another important truth: strictness at the shell and integration level is necessary, but not sufficient. The internal Go package tests were still carrying structural problems that made deterministic unit testing harder than it needed to be. Fixing those tests required improving the production code seams themselves.

Several themes emerged:

- `internal/s3/client.go` needed an interface boundary so tests could substitute a controllable S3 implementation instead of depending on the concrete AWS SDK client everywhere.
- `internal/handlers/api.go` needed an interface boundary for its S3-facing dependency so handler tests could be isolated from real SDK behavior.
- some tests had become brittle because they tried to fake too much in the wrong place instead of using real filesystem/config behavior with temporary directories.
- some user-visible correctness issues, such as static asset path handling and `Content-Disposition` quoting, were not being protected by tests despite being part of the observable contract.

The lesson is that production code sometimes needs to become more testable in order for tests to become more trustworthy. Interfaces are not automatically good, but when they create stable seams around external dependencies such as SDK clients or service layers, they can turn unreliable or contorted tests into straightforward ones.

Another lesson is that testability is not just about mocks. In some places, the better answer was to stop over-mocking and use real config save/load operations in temporary directories. Good tests choose the simplest truthful boundary. Sometimes that means an interface. Sometimes it means the real filesystem. What matters is deterministic control with minimal fakery.

Finally, the unit-test fixes reinforced that strict quality work must span all three layers:

- package-level unit tests for pure logic and local error semantics
- handler/service tests for API contracts and boundary behavior
- end-to-end shell tests for real CLI and HTTP workflows in the containerized environment

If any of those layers is neglected, confidence becomes uneven. A green E2E suite with broken package tests is incomplete. A strong package suite with fake end-to-end evidence is also incomplete. Trust comes from alignment across layers.

## Lessons About Environment-Specific Semantics

One subtle but important lesson is that tests must distinguish between bad application behavior and environment-specific semantics. The unreadable-config test is a good example. In the containerized runner context, the process could still read a file that had mode `000`. Treating that as a failure of `s3peep` would have been incorrect. Treating it as if unreadability was truly proven would also have been incorrect. The correct test documents current runner semantics strictly.

This suggests an important testing rule:

- when the environment changes the meaning of an input (permissions, networking, process identity, filesystem visibility), the test must either control for that precisely or explicitly encode the environment-dependent semantics

In other words, a strict test suite is not the same as an idealized test suite. It must be strict about the behavior that truly exists in the environment where it runs.

## Documentation Is Part of the Test System

The README update was not a side task. It was part of making the suite trustworthy. A strict test environment needs documentation that answers, unambiguously:

- what command should be run
- where it should be run
- why host execution is wrong
- which workflow is preferred and why
- how to run a single test vs the full suite vs preflight
- which tools are infrastructure and which outputs count as proof

The lesson here is that documentation is operational code. When it is wrong, users lose time and confidence. When it is incomplete, they create unverified local workarounds. README commands should be treated as test cases, not prose.

## Process Rules That Should Persist

This feature generated several durable process rules that should survive beyond spec 002.

### Rule 1: Exact contract verification beats approximate verification

If the product advertises `./s3peep profile add --name ...`, then that exact command must be tested. If the README says `podman-compose ... run --rm testrunner /app/run_tests.sh`, that exact command must be executed. Testing something close to the real contract is not enough.

### Rule 2: Infrastructure checks are not product checks

MinIO, fixture scripts, helper functions, and container health are prerequisites or support mechanisms. They are not proof that `s3peep` works. Their outputs should never be the final basis for a product PASS.

### Rule 3: Run immediately after every edit

This rule applies to:

- tests
- helpers
- runner logic
- documentation commands
- public CLI contract changes

Immediate execution is not just fast feedback. It prevents a pile-up of interacting assumptions that become difficult to debug later.

### Rule 4: Track unimplemented behavior honestly

If a feature is not yet implemented, that should appear as future test debt, not as a vague or weak passing test. Honest TODOs are vastly better than dishonest green results.

### Rule 5: Prefer the least stateful documented workflow

For containerized systems, `run --rm` is usually a safer default than `exec`, because it reduces dependency on ambient state. Documentation should guide users toward the most deterministic path first.

### Rule 6: Improve production seams when tests reveal friction at dependency boundaries

If a package interacts with heavyweight or external dependencies such as cloud SDK clients, HTTP services, or filesystem-backed state, and the tests are becoming contorted, brittle, or fake, that is often a design signal. Add the smallest interface or boundary needed to make deterministic testing possible. The goal is not abstraction for its own sake. The goal is truthful, maintainable tests.

### Rule 7: Cover all three layers of confidence

Strict testing in this project should always be evaluated across three layers:

- unit tests for internal logic and error handling
- package or handler integration tests for API and service boundaries
- containerized end-to-end tests for real user-facing workflows

Declaring victory at only one layer invites blind spots. The strongest verification story is one where all three layers agree.

## What "Done" Means After This Experience

Before this work, "done" could have been interpreted as "there is a containerized test environment and a bunch of test files exist." That definition is too weak.

After this work, "done" for a containerized test system should mean:

- the documented workflows have been executed successfully as written
- every PASS in the suite corresponds to a real, explicit, user-meaningful contract
- infrastructure failures are separated cleanly from feature failures
- the suite is strict enough to catch CLI and HTTP contract drift quickly
- unit tests for the core packages are healthy enough to validate local logic and dependency-boundary behavior without heroic mocking
- the runner, helpers, and documentation have all been tested as system components
- unimplemented or not-yet-testable features are tracked explicitly rather than blurred into weak checks

This is a higher standard, but it is a better one. It produces fewer green lies, fewer confused debugging sessions, and more confidence that when the suite is green, it means something real.

## Final Reflection

This feature was hard and cumbersome partly because it forced repeated confrontation with a simple truth: a test environment is easy to make noisy and hard to make trustworthy. The pain came from removing ambiguity. That pain was useful. It exposed where the suite was bluffing, where the CLI contracts were broken, where the runner was brittle, and where the documentation had become aspirational instead of operational.

The most valuable insight from this work is that strictness is not a tone or an intention. It is a set of concrete engineering behaviors:

- verify the exact contract
- distrust indirect evidence
- separate infrastructure from product proof
- create testable seams at dependency boundaries when unit tests reveal design friction
- encode current semantics honestly
- rerun immediately after each change
- treat documentation and test harnesses as things that themselves require testing
- keep shell, handler, and package tests aligned so confidence is consistent across layers

If those principles are carried forward, this project will not just have more tests. It will have a more truthful testing system.
