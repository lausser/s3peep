# Feature Specification: S3 File Browser

**Feature Branch**: `001-s3-file-browser`  
**Created**: 2026-03-28  
**Status**: Draft  
**Input**: User description: "initial spec for a simple file browser/downloader/uploader for S3-compatible storage"

## Clarifications

### Session 2026-03-28

- Q: Credential Storage Security → A: JSON config file at ~/.config/s3peep/config.json with profiles array, active_profile field, supporting add/switch profiles at runtime
- Q: UI Approach → A: User flexible on native GUI or web-based UI (to be determined in planning)
- Q: Technical Constraints → A: Keep external dependencies slim (except Docker), easily dockerizable

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Browse S3 Files and Folders (Priority: P1)

As a user, I want to navigate through my S3-compatible storage to view files and folder structures so that I can locate the content I need.

**Why this priority**: Browsing is the foundational capability - users cannot download or organize files without first seeing what exists.

**Independent Test**: Can be tested by connecting to any S3-compatible bucket and verifying all files/folders are displayed with correct hierarchy.

**Acceptance Scenarios**:

1. **Given** a connected S3 bucket with nested folders and files, **When** I navigate to the bucket, **Then** I see the complete folder structure with all files listed
2. **Given** I am viewing a folder with many files, **When** I scroll through the file list, **Then** files display with their names, sizes, and last modified dates
3. **Given** I am in a subfolder, **When** I navigate to a parent folder, **Then** I return to the previous level with the correct context preserved

---

### User Story 2 - Download Files from S3 (Priority: P1)

As a user, I want to download files from S3 to my local system so that I can access content offline or transfer it elsewhere.

**Why this priority**: Downloading is a core utility function - users need to extract their data from S3 storage.

**Independent Test**: Can be tested by selecting a file in S3 and verifying it downloads correctly with intact content.

**Acceptance Scenarios**:

1. **Given** I have selected a file in S3, **When** I initiate a download, **Then** the file transfers to my local system with its original filename preserved
2. **Given** I am downloading a large file, **When** the download is in progress, **Then** I can see progress indication and the download completes successfully
3. **Given** I cancel a download in progress, **When** I check my system, **Then** no partial file remains

---

### User Story 3 - Upload Files to S3 (Priority: P1)

As a user, I want to upload files from my local system to S3 so that I can store and backup my data.

**Why this priority**: Uploading is essential for adding new content to storage - without it, users can only read existing data.

**Independent Test**: Can be tested by uploading a file from local system and verifying it appears in the target S3 location.

**Acceptance Scenarios**:

1. **Given** I am in a target folder in S3, **When** I upload a local file, **Then** the file appears in S3 with its original filename and content
2. **Given** I am uploading a file with the same name as an existing file, **When** the upload completes, **Then** I can choose to overwrite or keep both versions
3. **Given** I am uploading a large file, **When** the upload is in progress, **Then** I can see progress indication and the upload completes successfully

---

### User Story 4 - Connect to S3-Compatible Storage (Priority: P1)

As a user, I want to configure and establish a connection to my S3-compatible storage so that I can access my files.

**Why this priority**: Without connection configuration, users cannot access any S3 functionality.

**Independent Test**: Can be tested by entering credentials and verifying successful connection to any S3-compatible service.

**Acceptance Scenarios**:

1. **Given** I have my S3 credentials (endpoint, access key, secret key, region), **When** I create a profile and connect, **Then** I can access my bucket content
2. **Given** I have previously created a profile, **When** I select that profile and connect, **Then** I can reconnect quickly without re-entering credentials
3. **Given** I have multiple profiles configured, **When** I switch to a different profile, **Then** I can access a different S3-compatible service
4. **Given** I enter invalid credentials, **When** I attempt to connect, **Then** I receive a clear error message explaining the issue

---

### Edge Cases

- What happens when network connection is lost during file transfer?
- How does the system handle special characters in file names (unicode, spaces, symbols)?
- What happens when attempting to upload a file that exceeds storage quota?
- How does the system handle very large files (multi-gigabyte)?
- What happens when the S3 bucket is empty?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a hierarchical view of all folders and files in the connected S3 bucket
- **FR-002**: System MUST show file metadata including name, size, and last modified date
- **FR-003**: System MUST allow users to navigate into folders and return to parent levels
- **FR-004**: System MUST enable users to download any file from S3 to their local system
- **FR-005**: System MUST preserve original filenames when downloading files
- **FR-006**: System MUST enable users to upload files from their local system to S3
- **FR-007**: System MUST allow users to configure connection settings (endpoint, credentials, bucket)
- **FR-007a**: System MUST store credentials in JSON config file at ~/.config/s3peep/config.json
- **FR-007b**: System MUST allow users to create named profiles
- **FR-007c**: System MUST allow users to switch between profiles at runtime
- **FR-008**: System MUST support any S3-compatible storage service (AWS S3, MinIO, DigitalOcean Spaces, etc.)
- **FR-009**: System MUST display clear error messages when operations fail
- **FR-010**: System MUST indicate progress during file transfer operations

### Key Entities

- **Profile**: Represents a stored S3 connection configuration (name, region, access_key_id, secret_access_key, endpoint_url)
- **S3Connection**: Represents the active connection to an S3-compatible service using a profile
- **Bucket**: Represents an S3 bucket containing files and folders
- **FileObject**: Represents a file in S3 with metadata (key/name, size, last modified, storage class)
- **Folder**: Represents a logical grouping of files (S3 uses prefix-based folders)
- **TransferOperation**: Represents an in-progress or completed file transfer (upload/download) with status and progress

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully browse and view all files in any S3 bucket they have access to
- **SC-002**: Users can download any file from S3 with the original filename preserved
- **SC-003**: Users can upload files to S3 and see them appear in the correct location
- **SC-004**: Users can connect to any S3-compatible service (not just AWS S3)
- **SC-005**: 95% of file transfers (downloads and uploads) complete successfully without errors
- **SC-006**: Users understand connection errors and can retry with corrected credentials

## Assumptions

- Users have valid S3-compatible storage credentials (access key, secret key)
- Users have network connectivity to access their S3 storage
- The application will be used on a system with sufficient disk space for downloads
- Users are familiar with basic file operations (browsing, downloading, uploading)
- S3 bucket permissions allow the operations users need to perform
- Credentials are stored in plain JSON file (user accepts this security model)
- External dependencies kept minimal for slim deployment (Docker例外)
- Application must be easily dockerizable
