# Test Runner Contract

Discovery and execution rules:
- Discover test files by the `test_*.sh` naming convention.
- Run mandatory preflight before fixture seeding or test discovery output is summarized.
- Apply default fixtures before the general suite unless a test explicitly manages its own scenario.
- Summarize totals for passed and failed tests and exit non-zero when any test fails.

Contract rules:
- The runner must not report success if preflight fails.
- The runner must preserve deterministic ordering of discovered test files.
- The runner must surface failed test filenames in the final summary.
