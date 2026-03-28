# Tasks: S3 File Browser

**Feature**: 001-s3-file-browser  
**Generated**: 2026-03-28

## Implementation Strategy

MVP Scope: User Story 4 (Connect) + User Story 1 (Browse) = functional file browser  
Delivery: Incremental by user story

## Phase 1: Setup

- [X] T001 Initialize Go module with go mod init
- [X] T002 Create project directory structure per plan.md
- [X] T003 Add aws-sdk-go-v2 dependencies to go.mod
- [X] T004 Create Dockerfile with multi-stage build
- [X] T005 Create .gitignore for Go project

## Phase 2: Foundational

- [X] T006 Implement config package in internal/config/config.go (load/save config)
- [X] T007 Implement Profile struct with validation in internal/config/profile.go
- [X] T008 Implement S3 client wrapper in internal/s3/client.go
- [X] T009 Create main.go entry point with CLI flag parsing in cmd/s3peep/main.go
- [X] T010 Implement embedded HTTP server stub in internal/handlers/api.go
- [X] T011 Verify project compiles successfully

## Phase 3: User Story 4 - Connect to S3-Compatible Storage

**Goal**: Users can create, list, switch profiles and connect to S3  
**Independent Test**: Can add profile, switch profile, connect to any S3-compatible service  
**Dependencies**: Phase 2 complete

- [X] T012 [P] [US4] Add profile management CLI commands in internal/config/cli.go
- [X] T013 [US4] Implement profile add command
- [X] T014 [US4] Implement profile list command
- [X] T015 [US4] Implement profile switch command
- [X] T016 [US4] Add S3 connection test endpoint in internal/s3/client.go
- [X] T017 [US4] Add connection status to HTTP API in internal/handlers/api.go
- [X] T018 [US4] Test profile CRUD operations manually

## Phase 4: User Story 1 - Browse S3 Files and Folders

**Goal**: Users can navigate folder hierarchy and see file metadata  
**Independent Test**: Can list bucket contents, navigate into folders, return to parent  
**Dependencies**: US4 (need connection)

- [X] T019 [P] [US1] Implement ListObjects in internal/s3/client.go
- [X] T020 [US1] Add list endpoint to HTTP API in internal/handlers/api.go
- [X] T021 [US1] Create basic web UI in web/index.html
- [X] T022 [US1] Implement folder navigation in web/app.js
- [X] T023 [US1] Display file metadata (name, size, date) in UI
- [ ] T024 [US1] Test file browsing with real S3 bucket

## Phase 5: User Story 2 - Download Files from S3

**Goal**: Users can download files with progress indication  
**Independent Test**: Can download any file, see progress, cancel download  
**Dependencies**: US1 (need browsing to select file)

- [ ] T025 [P] [US2] Implement GetObject in internal/s3/client.go
- [ ] T026 [US2] Add download endpoint to HTTP API with streaming
- [ ] T027 [US2] Add download button to web UI
- [ ] T028 [US2] Implement progress tracking for downloads
- [ ] T029 [US2] Test large file download

## Phase 6: User Story 3 - Upload Files to S3

**Goal**: Users can upload files with progress and overwrite handling  
**Independent Test**: Can upload file, see progress, handle name conflicts  
**Dependencies**: US1 (need browsing to select destination)

- [ ] T030 [P] [US3] Implement PutObject in internal/s3/client.go
- [ ] T031 [US3] Add upload endpoint to HTTP API
- [ ] T032 [US3] Add upload functionality to web UI
- [ ] T033 [US3] Implement progress tracking for uploads
- [ ] T034 [US3] Handle overwrite confirmation dialog
- [ ] T035 [US3] Test large file upload

## Phase 7: Polish & Cross-Cutting

- [ ] T036 Implement error handling and user-friendly error messages
- [ ] T037 Add empty state handling (empty bucket message)
- [ ] T038 Embed web assets in Go binary using embed directive
- [ ] T039 Final build verification and Docker image test
- [ ] T040 Update quickstart.md with final usage instructions

## Dependencies Graph

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational)
    ↓
Phase 3 (US4: Connect) → Phase 4 (US1: Browse) → Phase 5 (US2: Download)
                         ↓
                       Phase 6 (US3: Upload)
                         ↓
                    Phase 7 (Polish)
```

## Parallel Opportunities

| Task | Parallel With | Reason |
|------|---------------|--------|
| T012 | T019, T025, T030 | Different components (config vs s3 operations) |
| T020 | T026, T031 | Different API endpoints |
| T021 | T027, T032 | Different UI components |

## Task Count

- Total: 40 tasks
- Phase 1 (Setup): 5 tasks
- Phase 2 (Foundational): 6 tasks
- Phase 3 (US4): 7 tasks
- Phase 4 (US1): 6 tasks
- Phase 5 (US2): 5 tasks
- Phase 6 (US3): 6 tasks
- Phase 7 (Polish): 5 tasks
