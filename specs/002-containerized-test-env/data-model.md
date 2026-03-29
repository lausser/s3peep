# Data Model: Full Containerized Test Environment

## TestEnvironment
- Purpose: Represents the full local containerized system used for end-to-end validation.
- Fields:
  - `services`: `minio`, `s3peep`, `testrunner`
  - `s3_endpoint`: `http://minio:9000`
  - `app_endpoint`: `http://s3peep:8080`
  - `host_ui_port`: configurable mapped host port, default `8080`
  - `persistent_volume`: named MinIO data volume
  - `config_dir`: writable config path used by CLI/server tests
- Relationships:
  - Owns one active `SeedScenario`
  - Runs zero or more `PreflightCheck` items before any `TestCaseContract`

## PreflightCheck
- Purpose: Blocking prerequisite verification executed before fixture seeding or feature tests.
- Fields:
  - `name`: stable identifier such as `binary-exists` or `minio-reachable`
  - `command_or_probe`: shell command or HTTP/S3 probe used for validation
  - `timeout_seconds`: maximum time before the check fails
  - `failure_message`: deterministic diagnostic shown to the developer
  - `blocking`: always `true` for spec 002 preflight checks
- Invariants:
  - Every configured preflight check must either PASS explicitly or abort the suite
  - Combined preflight runtime must satisfy the 30-second failure budget in the spec

## SeedScenario
- Purpose: Named fixture state applied to MinIO before tests.
- Fields:
  - `name`: `default`, `empty`, `nested`, `largefiles`
  - `bucket_name`: test bucket targeted by the scenario
  - `reset_behavior`: bucket is recreated from a known baseline before scenario content is applied
  - `expected_objects`: list of required keys or object count expectations
- State transitions:
  - `unapplied` -> `reset` -> `seeded` -> `mutated` -> `reset`
- Invariants:
  - Re-running a scenario after interruption must return the bucket to a known state

## ConfigState
- Purpose: Captures config-file conditions relevant to CLI and server tests.
- Fields:
  - `path`: config file location under the test runner's writable directory
  - `status`: `missing`, `empty-default`, `valid`, `malformed-json`, `semantically-invalid`, `unreadable`, `unwritable`
  - `active_profile`: optional profile name
  - `profiles`: zero or more configured profiles
- Invariants:
  - CLI and server tests assert against actual load/save behavior, not helper-generated JSON alone
  - Invalid config tests must distinguish malformed content from permission or path failures

## TestCaseContract
- Purpose: Defines the observable behavior a shell test must prove.
- Fields:
  - `id`: test filename or logical test identifier
  - `surface`: `cli` or `http`
  - `fixture_dependency`: optional `SeedScenario`
  - `preconditions`: required service or config state
  - `expected_exit_code`: for CLI and script-level commands
  - `expected_status_code`: for HTTP requests when applicable
  - `expected_observables`: required JSON fields, headers, output strings, or file-state changes
- Invariants:
  - PASS requires strict assertion over explicit observables
  - No PASS may depend solely on MinIO client output, warning logs, or non-empty responses

## SecurityCase
- Purpose: Represents a high-risk regression input against currently supported interfaces.
- Fields:
  - `name`: `path-traversal`, `xss-filename`, `invalid-config`, `connection-failure`
  - `entrypoint`: CLI command or HTTP endpoint under test
  - `malicious_or_invalid_input`: crafted key, filename, config body, or connection setting
  - `expected_rejection`: client-visible error status, message, or failure exit
  - `must_not_happen`: forbidden side effect such as host file exposure, script execution, or false PASS
- Invariants:
  - Every security case must be testable through a current public interface
  - Contracts must define rejection behavior tightly enough for automated assertions
