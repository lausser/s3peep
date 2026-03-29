# Preflight Contract

Mandatory checks before any fixture setup or feature test:
- s3peep binary exists and is executable
- MinIO is reachable with configured credentials
- config directory is writable
- required in-container tools are installed: `bash`, `curl`, `jq`, `mc`
- s3peep HTTP endpoint is reachable when running HTTP-facing tests

Contract rules:
- Preflight aborts the suite on the first missing prerequisite.
- Failure output must name the failing prerequisite.
- Broken prerequisite states must fail within 30 seconds.
