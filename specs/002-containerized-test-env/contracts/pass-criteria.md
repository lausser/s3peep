# PASS Criteria Contract

A shell test may report PASS only when all of the following are true:
- It exercises a public s3peep CLI command or HTTP endpoint.
- It asserts explicit observables such as exit code, status code, JSON fields, headers, stdout/stderr, or config-file state.
- It fails on unmet expectations rather than downgrading them to warnings.

Forbidden PASS patterns:
- Passing because a MinIO client command succeeded
- Passing because a response body is merely non-empty
- Passing after `|| true` suppresses a failure
- Passing after `log_warn` on a required assertion
