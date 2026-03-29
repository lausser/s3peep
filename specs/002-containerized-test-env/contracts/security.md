# Security Contract

Required regression scope for current public interfaces:
- Path traversal-like request inputs must be rejected and must not expose host filesystem data.
- File names containing HTML or script-like content must be safely rendered or returned by s3peep.
- Invalid config content must fail deterministically through CLI or server load paths.
- Connection failure cases must surface explicit failure rather than false PASS output.

Contract rules:
- Security tests are limited to already-supported CLI and HTTP behavior.
- Rejection behavior must be asserted through status codes, exit codes, messages, or other explicit outputs.
