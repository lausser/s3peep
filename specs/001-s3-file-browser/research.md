# Research: S3 File Browser

## Decision: Language Selection

**Chosen**: Go 1.21+

**Rationale**:
- Single binary deployment (no runtime dependencies)
- Excellent S3 support via aws-sdk-go-v2
- Easy Docker containerization (minimal base image)
- Fast execution
- User constraints: slim dependencies, dockerizable

**Alternatives considered**:
- Python: Rich S3 libraries but requires runtime installation
- Node.js: Good S3 support but larger container images
- Rust: Similar benefits to Go but steeper learning curve

---

## Decision: UI Approach

**Chosen**: Web-based (embedded HTTP server)

**Rationale**:
- User explicitly offered "opens a port so that the UI can be a normal web browser"
- Cross-platform without native UI code
- Easier to maintain and extend
- Can be embedded in single Go binary

**Implementation**: Go HTTP server with embedded frontend (HTML/CSS/JS)

---

## Decision: S3 SDK

**Chosen**: aws-sdk-go-v2

**Rationale**:
- Official AWS SDK
- Supports all S3-compatible services (MinIO, DigitalOcean Spaces, etc.)
- Well-maintained
- Works with Go 1.21+

---

## Decision: Project Structure

**Chosen**: Single Go project with embedded web UI

```
src/
├── cmd/s3peep/       # Main entry point
├── internal/
│   ├── config/       # Profile management
│   ├── s3/           # S3 client operations
│   ├── handlers/     # HTTP handlers
│   └── ui/           # Embedded web assets
├── web/              # Web UI source (for development)
└── tests/            # Integration tests
```

---

## Key Implementation Details

- Config stored at ~/.config/s3peep/config.json (per user requirement)
- Single executable with embedded web assets
- Docker: Multi-stage build for minimal image size
- S3 operations: ListObjectsV2, GetObject, PutObject
