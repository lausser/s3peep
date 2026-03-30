# s3peep Project Constraints

## Build and Test Environment

### Critical Constraints
- **NO Go compiler on host** - Compilation must happen in a golang container
- **CAN run binary on host** - After building in container, binary runs on host
- **Integration tests in containers** - Full test suite runs via podman-compose

### Build Workflow

**Step 1: Build binary in golang container:**
```bash
podman run --rm -v $(pwd):/app -w /app golang:1.24-alpine go build -o s3peep ./cmd/s3peep
```

**Step 2: Run binary on host:**
```bash
./s3peep serve
```

### Integration Test Commands

**Build and start test environment:**
```bash
podman-compose -f test/docker-compose.yml up -d --build
```

**Run full test suite (in testrunner container):**
```bash
podman-compose -f test/docker-compose.yml run --rm testrunner /app/run_tests.sh
```

**Run single test:**
```bash
podman-compose -f test/docker-compose.yml run --rm testrunner /app/run_tests.sh test_name.sh
```

**Stop and cleanup:**
```bash
podman-compose -f test/docker-compose.yml down -v
```

### What NOT To Do
- ❌ `go run ./cmd/s3peep serve` (no Go compiler on host)
- ❌ `go build ./cmd/s3peep` (no Go compiler on host)
- ❌ `bash test/run_tests.sh` on host (must run inside testrunner container)

### Container Architecture (for integration tests)
- `minio`: S3-compatible storage
- `s3peep`: Application under test (built from Dockerfile.s3peep)
- `testrunner`: Bash-based test harness (built from Dockerfile.testrunner)

### Reference
- See `test/README.md` for full documentation
- See `test/docker-compose.yml` for service definitions