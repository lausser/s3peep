# Tasks: S3 File Browser Beautiful Web UI

**Feature**: S3 File Browser Beautiful Web UI  
**Branch**: `003-s3peep-web-ui`  
**Generated**: March 29, 2026  
**Status**: Ready for Implementation

---

## Overview

This task breakdown implements a beautiful, modern web UI for s3peep with real-time filtering, drag-and-drop uploads, keyboard shortcuts, and polished UX. Tasks are organized by user story priority for independent implementation and testing.

**Total Tasks**: 47  
**User Stories**: 6 (3 P1, 2 P2, 1 P3)  
**Parallel Opportunities**: 12 tasks can be executed in parallel

---

## Phase 1: Setup

**Goal**: Initialize project structure and shared infrastructure

**Independent Test**: Project builds successfully, no runtime errors

- [X] T001 Create directory structure per implementation plan: `internal/web/static/js/`, `internal/web/static/css/`, `internal/web/templates/`
- [X] T002 [P] Initialize Go module dependencies (no new external deps needed - using existing aws-sdk-go-v2)
- [X] T003 [P] Create base HTML template skeleton in `internal/web/templates/index.html` with semantic structure
- [X] T004 [P] Set up CSS foundation with CSS custom properties for light/dark themes in `internal/web/static/css/styles.css`
- [X] T005 Create shared utility functions in `internal/web/static/js/utils.js` (debounce, formatSize, formatDate, escapeHtml)

---

## Phase 2: Foundational Infrastructure

**Goal**: Build blocking prerequisites required by all user stories

**Independent Test**: Token middleware validates requests, API client can make authenticated requests

### Authentication & Security

- [X] T006 Implement token generation in `cmd/s3peep/main.go` (32-byte crypto-random, base64, print to stdout on startup)
- [X] T007 Create token validation middleware in `internal/handlers/auth.go` - validates token from URL path, returns 403 for invalid tokens
- [ ] T008 Store token in sessionStorage and validate on page load in `internal/web/static/js/auth.js`

### API Client Foundation

- [X] T009 Create API client module in `internal/web/static/js/api.js` with functions: request(), handleError(), getTokenFromUrl()

### S3 Client Enhancements

- [X] T010 Add pagination support to S3 client in `internal/s3/client.go` - ListObjectsPaginated(bucket, prefix, continuationToken, maxKeys)

### Error Handling

- [ ] T011 Create error display component in `internal/web/static/js/components/error.js` - toast notifications for API errors

---

## Phase 3: User Story 1 - Browse S3 Buckets (P1)

**Story Goal**: Homepage with bucket list, real-time filtering, and navigation

**Independent Test**: 
1. Start server
2. Open web UI
3. See list of accessible buckets
4. Type in filter - list filters dynamically
5. Click bucket - navigates to bucket view

### Backend

- [X] T012 [US1] Implement GET `/api/buckets` endpoint in `internal/handlers/buckets.go` - returns all buckets visible to profile
- [X] T013 [US1] Implement POST `/api/buckets/select` endpoint in `internal/handlers/buckets.go` - sets active bucket in profile
- [X] T014 [US1] Implement GET `/api/profile` endpoint in `internal/handlers/profile.go` - returns profile info (no sensitive fields)

### Frontend - Bucket List

- [X] T015 [US1] Create bucket list view component in `internal/web/static/js/components/bucket-list.js`
- [X] T016 [US1] Implement bucket filter input in `internal/web/static/js/components/bucket-filter.js` - real-time filtering, debounced (300ms)
- [X] T017 [US1] Create empty state component for "No buckets match" in `internal/web/static/js/components/empty-state.js`

### Frontend - State Management

- [X] T018 [US1] Add bucket list state to `internal/web/static/js/state.js` - buckets[], filterText, selectedBucket

### Integration

- [X] T019 [US1] Wire up bucket list page in `internal/web/static/js/app.js` - render buckets, handle click, navigate to bucket view
- [X] T020 [US1] Handle default bucket navigation in `internal/web/static/js/app.js` - if profile.bucket exists, pre-fill filter and auto-navigate

---

## Phase 4: User Story 2 - Browse Files in Bucket (P1)

**Story Goal**: File browser with filtering, pagination, breadcrumbs, and download

**Independent Test**:
1. Navigate to bucket
2. See file list with icons, sizes, dates
3. Type in filter - files filter dynamically (current page only)
4. Click folder - navigate into it
5. Click breadcrumb - go back
6. Click file - download starts

### Backend

- [X] T021 [US2] Implement GET `/api/buckets/{bucket}/objects` endpoint in `internal/handlers/objects.go` - list files/folders with pagination support
- [X] T022 [US2] Implement GET `/api/buckets/{bucket}/download` endpoint in `internal/handlers/objects.go` - stream file download

### Frontend - File Browser

- [X] T023 [US2] Create file list view component in `internal/web/static/js/components/file-list.js` - render files with icons, sizes, dates
- [X] T024 [US2] Implement file type icon mapping in `internal/web/static/js/utils/file-types.js` - extension to icon class
- [X] T025 [US2] Create breadcrumb navigation component in `internal/web/static/js/components/breadcrumb.js`
- [X] T026 [US2] Implement file filter input in `internal/web/static/js/components/file-filter.js` - client-side filtering of current page
- [X] T027 [US2] Create pagination controls component in `internal/web/static/js/components/pagination.js` - First, Previous, Next, Last, page size selector

### Frontend - Empty States

- [X] T028 [US2] Create empty state for "No files" in `internal/web/static/js/components/empty-state.js`
- [X] T029 [US2] Create empty state for "No files match filter" with Clear Filter button

### Frontend - State Management

- [X] T030 [US2] Add file browser state to `internal/web/static/js/state.js` - currentPath, files[], pagination, filterText

### Integration

- [X] T031 [US2] Wire up file browser in `internal/web/static/js/app.js` - render files, handle navigation, download

---

## Phase 5: User Story 3 - Upload Files and Folders (P1)

**Story Goal**: Drag-and-drop upload with progress indicators and conflict resolution

**Independent Test**:
1. Drag file into browser
2. File uploads with progress bar
3. Upload completes, file appears in list
4. Upload same file again - conflict dialog appears
5. Choose "Replace" - file overwritten
6. Upload large file (>100MB) - multipart upload with progress

### Backend

- [ ] T032 [US3] Implement POST `/api/buckets/{bucket}/upload` endpoint in `internal/handlers/upload.go` - single and multipart upload
- [ ] T033 [US3] Implement GET `/api/upload/{upload_id}/progress` endpoint in `internal/handlers/upload.go` - returns upload progress

### Frontend - Upload

- [ ] T034 [US3] Create drag-and-drop zone component in `internal/web/static/js/components/upload-zone.js`
- [ ] T035 [US3] Implement upload progress component in `internal/web/static/js/components/upload-progress.js` - progress bar, speed, ETA
- [ ] T036 [US3] Create file picker dialog integration in `internal/web/static/js/components/upload-button.js`
- [ ] T037 [US3] Implement conflict resolution modal in `internal/web/static/js/components/conflict-modal.js` - Replace/Auto-rename/Skip options, Apply to all checkbox

### Frontend - State Management

- [ ] T038 [US3] Add upload queue state to `internal/web/static/js/state.js` - UploadTask[], activeUploads

### Integration

- [ ] T039 [US3] Wire up upload flow in `internal/web/static/js/app.js` - handle drop, show progress, refresh file list on completion

---

## Phase 6: User Story 4 - Delete Files and Folders (P2)

**Story Goal**: Multi-select deletion with confirmation dialog

**Independent Test**:
1. Select files using checkboxes
2. Click Delete button
3. Confirmation dialog appears
4. Confirm - files deleted, removed from list
5. Delete folder - confirmation shows "will delete contents recursively"

### Backend

- [ ] T040 [US4] Implement DELETE `/api/buckets/{bucket}/objects` endpoint in `internal/handlers/objects.go` - delete single or multiple objects, recursive folder deletion

### Frontend

- [ ] T041 [US4] Create multi-select checkbox component in `internal/web/static/js/components/file-list.js` - add checkboxes, select all
- [ ] T042 [US4] Create delete confirmation modal in `internal/web/static/js/components/delete-modal.js` - show selected items, recursive warning for folders
- [ ] T043 [US4] Add bulk action bar in `internal/web/static/js/components/bulk-actions.js` - shows when items selected (Delete, Download buttons)

### Integration

- [ ] T044 [US4] Wire up deletion flow in `internal/web/static/js/app.js` - handle delete click, show modal, call API, refresh list

---

## Phase 7: User Story 5 - Search and Filter Files (P2)

**Story Goal**: Enhanced filtering with file type and date filters

**Independent Test**:
1. Open filter panel
2. Select file type "Images" - only images shown
3. Select date range - only files from that range shown
4. Clear filters - all files shown

### Frontend

- [ ] T045 [US5] Create advanced filter panel component in `internal/web/static/js/components/filter-panel.js` - file type checkboxes, date range pickers
- [ ] T046 [US5] Implement combined filtering logic in `internal/web/static/js/utils/filters.js` - name + type + date, OR for multiple types

### Integration

- [ ] T047 [US5] Wire up filter panel in `internal/web/static/js/app.js` - toggle panel, apply filters, clear filters

---

## Phase 8: User Story 6 - Create Folders (P3)

**Story Goal**: Create new folders in S3

**Independent Test**:
1. Click "New Folder" button
2. Enter folder name
3. Folder created, appears in list
4. Try creating duplicate - error message shown

### Backend

- [ ] T048 [US6] Implement PUT `/api/buckets/{bucket}/folders` endpoint in `internal/handlers/objects.go` - creates folder prefix

### Frontend

- [ ] T049 [US6] Create "New Folder" button in toolbar in `internal/web/static/js/components/toolbar.js`
- [ ] T050 [US6] Create new folder modal in `internal/web/static/js/components/folder-modal.js` - input field, validation, duplicate check

### Integration

- [ ] T051 [US6] Wire up folder creation in `internal/web/static/js/app.js` - show modal, call API, refresh list

---

## Phase 9: Polish & Cross-Cutting Concerns

**Goal**: Keyboard shortcuts, accessibility, loading states, theming, final integration

**Independent Test**: Lighthouse score 90+, all keyboard shortcuts work, theme toggle works

### Keyboard Shortcuts

- [ ] T052 [P] Implement keyboard shortcuts in `internal/web/static/js/keyboard.js` - `/` focus filter, `Esc` clear filter, `Ctrl/Cmd+K` quick nav, arrow keys navigate, `Enter` open, `Delete` delete
- [ ] T053 [P] Add keyboard shortcut hints to UI in `internal/web/static/js/components/shortcuts-help.js`

### Loading States

- [ ] T054 [P] Create skeleton loader components in `internal/web/static/js/components/skeleton.js` - for bucket list, file list

### Theming

- [ ] T055 [P] Implement theme toggle in `internal/web/static/js/theme.js` - light/dark switch, persist to sessionStorage
- [ ] T056 [P] Add CSS dark theme variables in `internal/web/static/css/styles.css`

### Accessibility

- [ ] T057 [P] Add ARIA labels to all interactive elements in HTML templates and JS components
- [ ] T058 [P] Ensure focus management (visible focus indicators, trap focus in modals) in CSS and JS
- [ ] T059 [P] Run Lighthouse accessibility audit and fix issues

### Final Integration

- [ ] T060 Update `internal/handlers/api.go` to serve new static files and templates (migrate from inline HTML)
- [ ] T061 Update `cmd/s3peep/main.go` to print full URL with token on startup
- [ ] T062 Create integration tests in `test/integration/api_test.go` - test all endpoints with MinIO
- [ ] T063 Update README.md with new features and usage instructions

---

## Dependency Graph

```
Phase 1 (Setup)
    ├── T001-T005
    └── All subsequent phases depend on this

Phase 2 (Foundational)
    ├── T006-T011 (Auth, API client, S3 pagination)
    └── Required by all user stories

Phase 3 (US1 - Browse Buckets)
    ├── Depends on: Phase 1, Phase 2
    ├── T012-T014 (Backend endpoints)
    ├── T015-T020 (Frontend components)
    └── Can work in parallel with: US2 initial tasks

Phase 4 (US2 - Browse Files)
    ├── Depends on: Phase 1, Phase 2
    ├── T021-T022 (Backend endpoints)
    ├── T023-T031 (Frontend components)
    └── Blocks: US3, US4, US5, US6 (needs file browser)

Phase 5 (US3 - Upload)
    ├── Depends on: Phase 1, Phase 2, US2 (needs file browser context)
    ├── T032-T033 (Backend)
    ├── T034-T039 (Frontend)
    └── Can work in parallel with: US4, US5, US6

Phase 6 (US4 - Delete)
    ├── Depends on: Phase 1, Phase 2, US2
    ├── T040 (Backend)
    ├── T041-T044 (Frontend)
    └── Can work in parallel with: US3, US5, US6

Phase 7 (US5 - Advanced Filter)
    ├── Depends on: Phase 1, Phase 2, US2
    ├── T045-T047 (Frontend only)
    └── Can work in parallel with: US3, US4, US6

Phase 8 (US6 - Create Folder)
    ├── Depends on: Phase 1, Phase 2, US2
    ├── T048 (Backend)
    ├── T049-T051 (Frontend)
    └── Can work in parallel with: US3, US4, US5

Phase 9 (Polish)
    ├── Depends on: All user stories
    └── T052-T063
```

## Parallel Execution Examples

**Example 1: Maximum Parallelism (Early Phase)**
```bash
# Phase 1 + 2 setup tasks can all run in parallel
T002, T003, T004, T005 - All independent setup tasks
T006, T007, T008, T009, T010, T011 - All foundational tasks
```

**Example 2: Backend/Frontend Split (US1)**
```bash
# Backend team works on:
T012, T013, T014 (bucket endpoints)

# Frontend team works in parallel:
T015, T016, T017, T018 (bucket list UI)

# Integration at the end:
T019, T020 (wire everything together)
```

**Example 3: Multiple User Stories in Parallel**
```bash
# After US2 (file browser) is complete, these can run in parallel:
T032-T039 (US3 - Upload)
T040-T044 (US4 - Delete)
T045-T047 (US5 - Advanced Filter)
T048-T051 (US6 - Create Folder)
```

**Example 4: Polish Phase Parallelism**
```bash
# All polish tasks are independent:
T052-T063 (Keyboard, Loading, Theming, Accessibility, Integration)
```

## MVP Scope Recommendation

**For initial release, implement:**

**Phase 1-2**: Setup and Foundational (T001-T011) ✅ REQUIRED  
**Phase 3**: User Story 1 - Browse Buckets (T012-T020) ✅ CORE  
**Phase 4**: User Story 2 - Browse Files (T021-T031) ✅ CORE  
**Phase 5**: User Story 3 - Upload Files (T032-T039) ✅ ESSENTIAL  
**Phase 6**: User Story 4 - Delete Files (T040-T044) ✅ ESSENTIAL  
**Phase 9**: Basic Polish (T060-T063, partial T052, T054) ✅ MINIMUM

**Total MVP Tasks**: 45  
**Deferred to v1.1**: US5 (Advanced Filter), US6 (Create Folders), full accessibility audit, theme toggle

This provides a complete, usable file browser with browse, upload, and delete capabilities.

---

## Implementation Notes

### File Type Icons
Use CSS classes: `.file-icon-image`, `.file-icon-document`, `.file-icon-archive`, `.file-icon-video`, `.file-icon-audio`, `.file-icon-code`, `.file-icon-folder`

### Pagination Strategy
Store continuation tokens in memory: `pageTokens = {1: '', 2: 'token2', 3: 'token3'}`  
First page: empty token  
Next: use token from current page + 1  
Previous: use token from current page - 1  
Last: Not supported (S3 doesn't provide total count)

### State Management Pattern
Event-driven: Components emit events, state updates, components re-render  
No external state library - vanilla JS with event listeners

### Error Handling
Backend: Always return JSON with `{error, code, details}`  
Frontend: Toast notifications for errors, inline validation for forms  
Network: Retry 3x with exponential backoff

### Security Checklist
- [ ] Token validation on every request
- [ ] Localhost binding only (127.0.0.1)
- [ ] No sensitive data in responses
- [ ] CORS allow-all (localhost tool)
- [ ] Input validation on all endpoints

---

**Next Step**: Begin implementation with Phase 1 (Setup) tasks T001-T005
