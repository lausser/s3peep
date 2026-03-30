# Implementation Plan: S3 File Browser Beautiful Web UI

**Branch**: `003-s3peep-web-ui` | **Date**: March 29, 2026 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `/specs/003-s3peep-web-ui/spec.md`

## Summary

Build a beautiful, modern web UI for s3peep that makes S3 file browsing effortless. The implementation adds real-time filtering, drag-and-drop uploads, keyboard shortcuts, and a polished UI while maintaining the existing Go backend architecture. Key innovations: token-based auth for single-user local access, S3 continuation token pagination mapping, and client-side page filtering for instant results.

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: 
- aws-sdk-go-v2 v1.41.5 (S3 operations)
- Standard library net/http (embedded server)
  
**Storage**: S3-compatible object storage via AWS SDK  
**Testing**: Go testing + containerized MinIO for integration tests  
**Target Platform**: Linux/macOS/Windows (desktop browsers)  
**Project Type**: CLI tool with embedded web server  
**Performance Goals**: 
- Page load: < 2 seconds
- Filter response: < 500ms for 100 items
- Upload throughput: Maximize S3 upload speed

**Constraints**: 
- Localhost-only binding (security)
- Single-user tool (no multi-user support)
- No external auth providers
- File size limit: 5GB

**Scale/Scope**: 
- Buckets: Unlimited (as per S3)
- Files per bucket: Unlimited with pagination
- Page size: 25/50/100/250 configurable
- Concurrent uploads: 5 max

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

✅ **No constitution file exists** - proceeding with standard implementation guidelines.

Guidelines applied:
- Keep dependencies minimal (no new external deps)
- Follow existing Go project structure
- Maintain backward compatibility with CLI
- Security: localhost-only, token-based auth
- Performance: pagination, client-side filtering

## Project Structure

### Documentation (this feature)

```text
specs/003-s3peep-web-ui/
├── plan.md              # This file
├── research.md          # Phase 0: Technical decisions and rationale
├── data-model.md        # Phase 1: Entities, fields, relationships
├── quickstart.md        # Phase 1: User guide
├── contracts/           # Phase 1: API contracts
│   └── api.md
└── tasks.md             # Phase 2: Task breakdown (TBD)
```

### Source Code (repository root)

```text
s3peep/
├── cmd/s3peep/
│   └── main.go              # Entry point, token generation
├── internal/
│   ├── config/
│   │   ├── config.go        # Config loading/saving
│   │   ├── profile.go       # Profile management
│   │   └── cli.go           # CLI commands
│   ├── handlers/
│   │   ├── api.go           # HTTP handlers (UPDATE)
│   │   ├── auth.go          # Token validation middleware (NEW)
│   │   └── assets.go        # Static asset serving (NEW)
│   ├── s3/
│   │   └── client.go        # S3 operations (UPDATE - pagination)
│   └── web/
│       ├── static/
│       │   ├── js/
│       │   │   ├── app.js       # Main app logic (NEW)
│       │   │   ├── api.js       # API client (NEW)
│       │   │   ├── components.js # UI components (NEW)
│       │   │   └── state.js     # State management (NEW)
│       │   └── css/
│       │       └── styles.css   # CSS with variables (NEW)
│       └── templates/
│           └── index.html       # HTML template (NEW)
├── web/                     # Legacy web assets (DEPRECATED)
│   └── ...
└── test/
    └── ...
```

**Structure Decision**: 
- Keep single-project structure (Go CLI tool)
- Add web/ subdirectory under internal/ for better organization
- Migrate from inline HTML in handlers to separate template files
- Separate JS/CSS for maintainability and caching

## Complexity Tracking

> **No complexity violations identified**

The implementation stays within single-project scope:
- No frontend framework (React/Vue) - vanilla JS is sufficient
- No database - S3 is the source of truth
- No external services - embedded server only
- Minimal new dependencies - using existing stack

## Implementation Phases

### Phase 0: Research ✅ COMPLETE

**Output**: [research.md](research.md)

**Key Decisions**:
1. Token-based auth with sessionStorage persistence
2. S3 continuation token pagination with page number mapping
3. Client-side filtering only (current page, not deep search)
4. Vanilla JavaScript for frontend (no build pipeline)
5. Multipart upload for files > 100MB
6. File type detection by extension

### Phase 1: Design ✅ COMPLETE

**Output**:
- [data-model.md](data-model.md) - Entities and API models
- [contracts/api.md](contracts/api.md) - HTTP API specification
- [quickstart.md](quickstart.md) - User documentation

**Data Model**:
- Bucket: name, creation_date, region
- FileObject: key, name, size, last_modified, is_folder, file_type
- UploadTask: id, file_name, progress, status, error_message
- SessionToken: token, created_at, expires_at, profile_name
- SearchFilter: query_text, file_type_filter, date_range
- PaginationState: current_page, page_size, continuation_tokens

**API Endpoints**:
- `GET /api/buckets` - List buckets
- `POST /api/buckets/select` - Select bucket
- `GET /api/buckets/{bucket}/objects` - List objects
- `GET /api/buckets/{bucket}/download` - Download file
- `POST /api/buckets/{bucket}/upload` - Upload file
- `PUT /api/buckets/{bucket}/folders` - Create folder
- `DELETE /api/buckets/{bucket}/objects` - Delete objects
- `GET /api/profile` - Get profile info

**UI Components**:
- Bucket view with filter input and list
- File browser with breadcrumb, filter, file list, pagination
- Upload modal with drag-drop and conflict resolution
- Delete confirmation modal
- New folder modal
- Error toast notifications
- Loading skeletons

### Phase 2: Implementation Tasks (TBD)

Task breakdown will be generated in `tasks.md` via `/speckit.tasks` command.

**High-level tasks**:
1. Backend
   - Token generation and validation middleware
   - Update S3 client with pagination support
   - Implement new API endpoints
   - Add static asset serving

2. Frontend
   - Create HTML template with semantic structure
   - Implement CSS with theme variables
   - Build JavaScript modules (api, state, components)
   - Add keyboard shortcuts and accessibility
   - Implement drag-drop upload

3. Testing
   - Unit tests for handlers
   - Integration tests with MinIO
   - E2E tests for critical flows
   - Accessibility audit

4. Documentation
   - Update README
   - Add inline code comments
   - Create troubleshooting guide

## Design Patterns

### Authentication
- Token generated on startup (32 bytes, base64)
- Token embedded in URL path: `/:token/api/...`
- Token stored in sessionStorage for page refresh persistence
- Middleware validates token on every request
- No cookies or JWT - simple and secure for localhost

### Pagination
- S3 ListObjectsV2 uses continuation tokens
- Map page numbers (1, 2, 3...) to tokens in memory
- Store token-to-page mapping: `map[pageNumber]continuationToken`
- First page has empty token
- Navigation: First (empty), Previous (page-1), Next (page+1), Last (not supported - S3 doesn't provide total count)

### State Management (Frontend)
- **Global state**: sessionStorage for token, theme preference
- **Ephemeral state**: JavaScript objects for file lists, uploads, filters
- **No state library**: Simple event-driven architecture
- State updates trigger UI re-renders

### Error Handling
- Backend: Return JSON error responses with code and message
- Frontend: Toast notifications for errors, inline validation
- Network errors: Retry with exponential backoff (3 attempts)
- S3 errors: Pass through with user-friendly messages

## Security Considerations

1. **Localhost binding**: Server binds to 127.0.0.1 only
2. **Token entropy**: 256-bit cryptographically secure random
3. **No HTTPS needed**: localhost only (browser handles as secure)
4. **Session duration**: 24 hours max
5. **No secrets in logs**: Access keys never logged
6. **CORS**: Allow all origins (localhost tool, no CSRF risk)

## Performance Optimizations

1. **Pagination**: Limit to 100 items default (configurable)
2. **Debouncing**: 300ms debounce on filter input
3. **Lazy loading**: Load file lists only when needed
4. **No caching**: Always fetch fresh from S3 (ensures consistency)
5. **Multipart upload**: Chunk large files for resumability
6. **Client-side filtering**: Instant results without server round-trip

## Browser Compatibility

- **Modern browsers**: Chrome 100+, Firefox 100+, Safari 15+, Edge 100+
- **Required features**:
  - Fetch API
  - CSS Grid/Flexbox
  - CSS Custom Properties
  - ES6 (const/let, arrow functions, async/await)
  - sessionStorage
  - Drag and Drop API

## Testing Strategy

**Unit Tests**:
- Handler functions with mocked S3 client
- Token generation and validation
- Pagination logic

**Integration Tests**:
- Full HTTP API with MinIO container
- Upload/download/delete operations
- Pagination with real S3 responses

**E2E Tests**:
- Critical user flows:
  1. Start server → Open UI → List buckets → Select bucket
  2. Filter buckets → Select → Filter files → Download
  3. Upload file → Handle conflict → Verify upload
  4. Delete file → Confirm → Verify deletion

**Accessibility Tests**:
- Lighthouse CI audit
- Keyboard navigation
- Screen reader compatibility

## Deployment Notes

**Not applicable** - This is a CLI tool, not a deployed service.

Users install via:
```bash
go install github.com/lausser/s3peep/cmd/s3peep
```

Or build from source:
```bash
go build -o s3peep ./cmd/s3peep
```

## Future Enhancements (Out of Scope)

- Copy/Move files between folders
- Cross-bucket deep search
- File previews (thumbnails)
- Image gallery view
- Multi-user support
- Remote access with HTTPS
- Upload resumption across sessions
- S3 Inventory integration for large buckets

## References

- **Spec**: [spec.md](spec.md)
- **Research**: [research.md](research.md)
- **Data Model**: [data-model.md](data-model.md)
- **API Contract**: [contracts/api.md](contracts/api.md)
- **User Guide**: [quickstart.md](quickstart.md)

---

**Status**: Phase 1 Complete | **Next**: `/speckit.tasks` to generate Phase 2 task breakdown
