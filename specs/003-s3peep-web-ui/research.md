# Research: S3 File Browser Web UI

**Feature**: S3 File Browser Beautiful Web UI  
**Date**: March 29, 2026  
**Branch**: 003-s3peep-web-ui

## Technical Context Analysis

### Current Stack
- **Language**: Go 1.24
- **AWS SDK**: aws-sdk-go-v2 v1.41.5
- **Web Framework**: Standard library `net/http` (embedded)
- **Frontend**: Vanilla JavaScript + CSS (currently inline in Go code)
- **Storage**: S3-compatible storage via aws-sdk-go-v2

### Existing Architecture
The current s3peep implementation has:
- CLI profile management (add/list/switch/remove profiles)
- Embedded HTTP server with basic HTML/CSS/JS serving
- S3 client wrapper around aws-sdk-go-v2
- File browser UI with bucket selection dropdown

## Decisions & Rationale

### 1. Token-Based Authentication
**Decision**: Use cryptographically secure random token in URL path + sessionStorage persistence

**Rationale**:
- Simple and secure for localhost-only development tool
- No external auth dependencies (OAuth, SSO)
- Token in URL prevents accidental sharing via browser history
- sessionStorage allows page refresh without re-entry

**Alternatives considered**:
- Cookie-based sessions: Overkill for single-user tool
- HTTP Basic Auth: Would require browser popup, worse UX
- JWT tokens: Adds unnecessary complexity

### 2. Pagination Strategy
**Decision**: Client-side pagination with S3 continuation token mapping

**Rationale**:
- S3 ListObjectsV2 uses continuation tokens, not offsets
- Mapping page numbers to tokens provides familiar UX
- Default 100 items balances performance and usability
- Page size options (25/50/100/250) give user control

**Implementation approach**:
- Store token-to-page mapping in memory
- First/Previous/Next/Last navigation
- Changing page size resets to page 1

### 3. File Filtering Scope
**Decision**: Filter current page only, not deep search

**Rationale**:
- S3 API doesn't support server-side name filtering
- Deep search would require listing ALL objects first
- Too slow/expensive for large buckets (10k+ objects)
- Users navigate pages and filter within each page

**Future enhancement**: Could add "Search this bucket" button that lists all objects progressively

### 4. Frontend Architecture
**Decision**: Keep vanilla JavaScript, enhance existing codebase

**Rationale**:
- Current implementation uses inline HTML/CSS/JS in Go
- Adding React/Vue would require build pipeline
- Vanilla JS is sufficient for this use case
- Can modernize with ES6 modules and fetch API

**Structure**:
- Separate JS modules: api.js, components.js, state.js
- CSS with CSS variables for theming (light/dark)
- Component-based architecture without framework

### 5. File Type Detection
**Decision**: Extension-based detection on client-side

**Rationale**:
- S3 doesn't store MIME type metadata reliably
- Extension mapping covers 95% of use cases
- Simple and fast: no magic number detection needed

**File type categories**:
- image: jpg, jpeg, png, gif, webp, svg, bmp
- document: pdf, doc, docx, txt, md, csv
- archive: zip, tar, gz, bz2, 7z, rar
- video: mp4, avi, mov, mkv, webm
- audio: mp3, wav, flac, aac, ogg
- code: js, py, go, java, cpp, html, css, json
- other: everything else

### 6. Multipart Upload Threshold
**Decision**: Use multipart uploads for files > 100MB

**Rationale**:
- S3 has 5GB limit for single PUT
- Multipart allows resumable uploads
- 100MB threshold balances complexity vs benefit
- aws-sdk-go-v2 has built-in multipart support

### 7. Keyboard Shortcuts
**Decision**: Implement common shortcuts without libraries

**Shortcuts**:
- `/` - Focus filter input
- `Esc` - Clear filter / close modal
- `Ctrl/Cmd+K` - Quick navigation (jump to bucket)
- Arrow keys - Navigate list
- `Enter` - Open selected item
- `Delete` - Delete selected (with confirmation)

**Rationale**: Native implementation keeps bundle small

## API Endpoints Needed

### Authentication
- All endpoints require token in URL path: `/:token/*`
- Middleware validates token against session

### Buckets
- `GET /:token/api/buckets` - List all buckets
- `POST /:token/api/buckets/select` - Select active bucket

### Files
- `GET /:token/api/buckets/:bucket/objects?prefix=&page=` - List objects with pagination
- `GET /:token/api/buckets/:bucket/download?key=` - Download file
- `POST /:token/api/buckets/:bucket/upload` - Upload file (multipart)
- `DELETE /:token/api/buckets/:bucket/objects` - Delete object(s)
- `PUT /:token/api/buckets/:bucket/folders` - Create folder

### Profile
- `GET /:token/api/profile` - Get current profile info
- `POST /:token/api/profile/switch` - Switch profile

## UI Components Needed

### Bucket View (Homepage)
- Filter input (auto-focused on `/`)
- Bucket list (click to navigate)
- Empty state with "Clear filter" button
- Keyboard shortcut hints

### File Browser View
- Breadcrumb navigation
- Filter input with path prefix
- File list with:
  - File type icons
  - Name
  - Size (human readable)
  - Last modified date
  - Checkbox for multi-select
- Pagination controls
- Action bar (Upload, New Folder, Delete, Refresh)
- Drag-and-drop upload zone

### Modals
- Upload conflict resolution
- Delete confirmation
- New folder creation
- Error messages

### Loading States
- Skeleton loaders for file list
- Progress bars for uploads
- Spinner for async operations

## Theming

**Light theme** (default):
- Background: #ffffff
- Surface: #f5f5f5
- Primary: #0066cc
- Text: #333333
- Border: #dddddd

**Dark theme**:
- Background: #1a1a1a
- Surface: #2d2d2d
- Primary: #4da3ff
- Text: #e0e0e0
- Border: #444444

**Implementation**: CSS custom properties with `data-theme` attribute

## Performance Targets

- Initial bucket list: < 2 seconds
- File filtering: < 500ms for 100 items
- Page navigation: < 1 second
- Upload start: < 3 seconds (including multipart init)

## Accessibility

- WCAG 2.1 AA compliance
- Keyboard navigation
- ARIA labels on interactive elements
- Focus indicators
- Color contrast 4.5:1 minimum
- Lighthouse score 90+

## Browser Support

- Chrome 100+
- Firefox 100+
- Safari 15+
- Edge 100+

## Risk Mitigation

1. **Large bucket listing**: Pagination with continuation tokens
2. **Slow S3 responses**: Loading states and timeout handling
3. **Browser compatibility**: Feature detection, graceful degradation
4. **Memory usage**: Don't cache entire bucket listings
5. **Token exposure**: URL token + sessionStorage, not localStorage

## Open Questions Resolved

1. **Search scope**: Current page only (not deep search)
2. **Default bucket**: Auto-navigate when profile has bucket configured
3. **Token expiration**: 24 hours from creation
4. **File conflict**: Modal with Replace/Rename/Skip options
5. **Copy/Move**: Out of scope for v1

## Dependencies to Add

None - using existing Go standard library and aws-sdk-go-v2

## Files to Create/Modify

### Backend (Go)
- `internal/handlers/api.go` - Update with new endpoints
- `internal/handlers/auth.go` - Token validation middleware
- `internal/s3/client.go` - Add pagination support
- `cmd/s3peep/main.go` - Token generation on startup

### Frontend (JS/CSS/HTML)
- `web/static/js/app.js` - Main application logic
- `web/static/js/api.js` - API client
- `web/static/js/components.js` - UI components
- `web/static/js/state.js` - State management
- `web/static/css/styles.css` - Styles with CSS variables
- `web/templates/index.html` - Main HTML template

## Testing Strategy

- Unit tests for Go handlers
- Integration tests with MinIO (S3-compatible)
- E2E tests for critical user flows
- Accessibility audit with Lighthouse
- Cross-browser testing
