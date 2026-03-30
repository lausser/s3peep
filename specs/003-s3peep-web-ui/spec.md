# Feature Specification: S3 File Browser Beautiful Web UI

**Feature Branch**: `003-s3peep-web-ui`  
**Created**: March 29, 2026  
**Status**: Draft  
**Input**: User description: "make the s3peep program usable, create a beautiful web ui"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Browse S3 Buckets (Priority: P1)

As a user, I want to browse available S3 buckets and select one to view its contents, so I can easily navigate my S3 storage.

**Why this priority**: This is the entry point of the application. Without bucket selection, users cannot access any files.

**Independent Test**: Can be fully tested by starting the server and viewing the bucket list. The UI should display all accessible buckets with a filter input, and clicking a bucket should navigate to the file browser.

**Acceptance Scenarios**:

1. **Given** the user has started s3peep server, **When** they open the web UI homepage, **Then** they see a text input field at the top and a list of all buckets visible to their credentials below
2. **Given** the user types text in the bucket filter input, **When** they type characters matching bucket names, **Then** the bucket list dynamically filters to show only matching buckets in real-time
3. **Given** the active profile has a default bucket configured, **When** the user opens the web UI, **Then** the bucket filter input is pre-filled with that bucket name, only that bucket is displayed, and the UI automatically navigates to the bucket view
4. **Given** the user sees filtered bucket results, **When** they click on a bucket name, **Then** the UI navigates to the bucket view showing all files in that bucket
5. **Given** the bucket filter returns no results, **When** the filter is active, **Then** the UI displays an empty state with a "Clear filter" button

---

### User Story 2 - Browse Files in Bucket (Priority: P1)

As a user, I want to view and navigate through files in a selected bucket with filtering capabilities, so I can find and access specific files quickly.

**Why this priority**: This is the primary file browsing functionality. Users need to see what files exist and filter them efficiently.

**Independent Test**: Can be fully tested by selecting a bucket and viewing its contents. The UI should display files with filtering, pagination, and allow clicking to navigate folders or download files.

**Acceptance Scenarios**:

1. **Given** the user has selected a bucket, **When** they view the bucket page, **Then** they see a text input field at the top (showing current path prefix) and a paginated list of all files/folders in that bucket below
2. **Given** the user types text in the file filter input, **When** they type characters matching file names, **Then** the file list dynamically filters to show only matching files on the current page in real-time
3. **Given** the user is viewing files in a folder, **When** they click on a file, **Then** the file downloads to their local machine
4. **Given** the user is in a nested folder structure, **When** they click the breadcrumb navigation, **Then** they can jump back to any parent folder level
5. **Given** the file filter returns no results, **When** the filter is active, **Then** the UI displays an empty state with a "Clear filter" button

---

### User Story 3 - Upload Files and Folders (Priority: P1)

As a user, I want to upload files and folders to S3 through a simple drag-and-drop interface, so I can easily add content to my buckets without using command-line tools.

**Why this priority**: Upload capability is essential for a complete file browser experience. Without it, users can only view and download, making the tool incomplete.

**Independent Test**: Can be fully tested by dragging files from the desktop into the browser window and verifying they appear in the S3 bucket. The upload should show progress indicators and handle both single files and multiple file selections.

**Acceptance Scenarios**:

1. **Given** the user has selected a bucket and navigated to a folder, **When** they drag and drop files onto the browser window, **Then** the files upload to the current location with visible progress indicators
2. **Given** the user clicks the upload button, **When** they select files from the file picker dialog, **Then** the selected files upload successfully to S3
3. **Given** multiple files are being uploaded, **When** one upload fails, **Then** the user sees a clear error message while other uploads continue
4. **Given** a large file is being uploaded, **When** the upload is in progress, **Then** the user sees upload speed, percentage complete, and estimated time remaining

---

### User Story 4 - Delete Files and Folders (Priority: P2)

As a user, I want to delete files and folders with a confirmation dialog, so I can manage my S3 storage and remove unwanted content safely.

**Why this priority**: File management requires the ability to delete. This enables users to maintain clean, organized storage without needing other tools.

**Independent Test**: Can be fully tested by selecting files/folders and clicking delete, then verifying they are removed from S3 after confirming the action.

**Acceptance Scenarios**:

1. **Given** the user selects a file, **When** they click the delete button, **Then** a confirmation dialog appears before deletion proceeds
2. **Given** the user selects multiple files using checkboxes, **When** they click delete, **Then** all selected files are deleted after confirmation
3. **Given** the user attempts to delete a non-empty folder, **When** they confirm deletion, **Then** all contents of the folder are recursively deleted
4. **Given** a deletion operation fails, **When** an error occurs, **Then** the user sees a clear error message explaining what went wrong

---

### User Story 5 - Search and Filter Files (Priority: P2)

As a user with large buckets, I want to search for files by name and filter by file type or date, so I can quickly find specific content without manual browsing.

**Why this priority**: Search becomes essential as storage grows. Without it, users with many files would spend excessive time browsing hierarchies.

**Independent Test**: Can be fully tested by typing search terms in a search box and verifying that matching files are displayed, regardless of which folder they're in.

**Acceptance Scenarios**:

1. **Given** the user types a search term in the search box, **When** they type characters, **Then** the UI filters the current page of files to show only matching names in real-time
2. **Given** search results are displayed, **When** the user clears the search, **Then** the UI returns to the normal folder view
3. **Given** the user applies a file type filter, **When** viewing a folder, **Then** only files of the selected type are displayed
4. **Given** the user applies a date range filter, **When** viewing a folder, **Then** only files modified within that date range are displayed

---

### User Story 6 - Create Folders (Priority: P3)

As a user, I want to create new folders in S3, so I can organize my files into logical structures.

**Why this priority**: Organization is important for maintaining clean storage, but users can work around this by uploading files to existing structures or using prefixes.

**Independent Test**: Can be fully tested by clicking "New Folder", entering a name, and verifying the folder appears in the current directory.

**Acceptance Scenarios**:

1. **Given** the user is viewing a bucket, **When** they click "New Folder" and enter a valid name, **Then** a new folder is created and appears in the file list
2. **Given** the user tries to create a folder with a name that already exists, **When** they submit, **Then** they see an error message and the folder is not created
3. **Given** the user enters an invalid folder name, **When** they submit, **Then** they see validation feedback about what characters are not allowed

---

### Edge Cases

- What happens when the S3 connection is lost during browsing? The UI should display a clear error message and offer a retry option.
- How does the system handle extremely large folders with thousands of files? The UI implements pagination with 100 items per page by default, with controls to adjust page size (25/50/100/250 items). Pagination uses S3 continuation tokens; page numbers are mapped to these tokens for navigation (First, Previous, Next, Last).
- What happens when a user tries to upload a file with the same name as an existing file? The system displays a modal dialog showing both files (existing file size and modification date vs new file) with three options: (1) Replace existing file, (2) Keep both files (auto-rename new file with timestamp suffix), (3) Skip this file. For batch uploads, provide an "Apply to all" checkbox to avoid repeated prompts.
- How are special characters in file names handled? The UI should properly encode/decode special characters and display them correctly.
- What happens when a user attempts to navigate to a bucket they don't have permissions for? The UI should display a permission denied message with helpful guidance.
- How does the system handle very large file uploads? The upload should support resumable multipart uploads for files over 100MB.
- What happens when someone accesses the server without the valid token? The server must return a 403 Forbidden response and not expose any application functionality or data.
- What happens when the user refreshes the browser page? The token is stored in sessionStorage, allowing the page to continue working. If sessionStorage is cleared or expired, the user sees a "Session expired" message with instructions to restart s3peep.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a visually appealing, modern web interface with consistent design language (colors, typography, spacing)
- **FR-002**: The system MUST display all accessible S3 buckets in a clear, selectable list
- **FR-003**: The system MUST show folder contents with file icons, names, sizes, and modification dates in a sortable table or grid view
- **FR-004**: The system MUST support breadcrumb navigation showing the current path and allowing quick navigation to parent folders
- **FR-005**: The system MUST allow users to download individual files by clicking on them
- **FR-006**: The system MUST support drag-and-drop file upload from the desktop to the browser
- **FR-007**: The system MUST support traditional file picker dialog for uploading files
- **FR-008**: The system MUST display upload progress with percentage, speed, and estimated time remaining
- **FR-009**: The system MUST allow deletion of files and folders with a confirmation dialog to prevent accidental deletion
- **FR-010**: The system MUST support multi-select of files using checkboxes for bulk operations
- **FR-011**: The system MUST provide a search functionality that filters files by name on the current page (not deep search across entire bucket)
- **FR-012**: The system MUST support creating new folders with validation for duplicate names and invalid characters
- **FR-013**: The system MUST handle errors gracefully, displaying user-friendly error messages instead of technical stack traces
- **FR-014**: The system MUST be responsive and work on desktop browsers (Chrome, Firefox, Safari, Edge)
- **FR-015**: The system MUST support keyboard navigation (arrow keys, Enter, Escape) for accessibility
- **FR-016**: The system MUST display loading states when fetching data from S3 to indicate activity
- **FR-017**: The system MUST support both light and dark color themes
- **FR-018**: The system MUST show empty state messages when folders have no content
- **FR-019**: The system MUST bind to localhost only (127.0.0.1/::1) and reject external connections
- **FR-020**: The system MUST generate a cryptographically secure random token on startup and require it in all URL paths for access
- **FR-021**: The homepage MUST display a text input field at the top for filtering buckets by name
- **FR-022**: The bucket list MUST filter dynamically as the user types in the bucket filter input
- **FR-023**: If the active profile has a default bucket configured, the bucket filter input MUST be pre-filled with that bucket name on page load and the UI MUST automatically navigate to that bucket view
- **FR-024**: The bucket view MUST display a text input field at the top for filtering files by name
- **FR-025**: The file list MUST filter dynamically as the user types in the file filter input
- **FR-026**: The system MUST store the authentication token in sessionStorage to persist across page refreshes
- **FR-027**: The system MUST display keyboard shortcut hints (e.g., "Press / to focus filter")
- **FR-028**: The system MUST support the following keyboard shortcuts: `/` to focus filter input, `Esc` to clear filter, `Ctrl/Cmd+K` for quick navigation
- **FR-029**: The system MUST display skeleton loaders while fetching data from S3
- **FR-030**: The system MUST show different file type icons for common types (images, documents, archives, etc.)
- **FR-031**: When filter returns no results, the system MUST display an empty state with a "Clear filter" button

### Key Entities *(include if feature involves data)*

- **FileObject**: Represents an S3 object (file or folder prefix) with properties: key (full path), name (display name), size (in bytes), last_modified (timestamp), is_folder (boolean), file_type (enum: image, document, archive, video, audio, code, other)
- **Bucket**: Represents an S3 bucket with properties: name, creation_date, region
- **UploadTask**: Represents an active file upload with properties: id, file_name, progress_percentage, status (pending/uploading/completed/failed), error_message
- **SearchFilter**: Represents current search criteria with properties: query_text, file_type_filter, date_range_start, date_range_end
- **SessionToken**: Represents the authentication token with properties: token (string), created_at (timestamp), expires_at (timestamp, default 24 hours from creation)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can browse and navigate S3 buckets without requiring any command-line interaction or external AWS tools
- **SC-002**: Users can upload files up to 5GB through the web interface with progress visibility
- **SC-003**: Users can complete common file operations (browse, download, upload, delete) within 3 clicks from the main interface
- **SC-004**: The interface achieves a Google Lighthouse accessibility score of 90 or higher
- **SC-005**: Page load time for initial bucket listing is under 2 seconds on a standard broadband connection
- **SC-006**: File filtering returns results within 500ms for pages containing up to 100 objects
- **SC-007**: Users can successfully perform file operations without reading documentation (measured by first-time user testing with 5 users)
- **SC-008**: The interface displays correctly and functions properly on screen sizes from 1280x720 to 4K resolution
- **SC-009**: Error messages are actionable and help users resolve issues without external support in 90% of cases
- **SC-010**: Users can access the application immediately after page refresh without re-entering the URL (token persistence)

## Assumptions

- Users have modern web browsers that support HTML5, CSS3, and ES6+ JavaScript (last 2 versions of major browsers)
- Users have stable internet connectivity sufficient for their S3 operations
- The S3 credentials provided have appropriate permissions for the operations users want to perform
- Mobile support is out of scope for this version; focus is on desktop browser experience
- File previews (image thumbnails, document previews) are out of scope for initial release
- Multi-part upload resumption across browser sessions is out of scope; uploads restart if page is refreshed
- Advanced ACL management and bucket policy editing are out of scope; focus is on object-level operations
- Integration with identity providers (SSO, OAuth) is out of scope; existing profile-based authentication is sufficient
- The existing backend API can be extended to support new operations without breaking changes
- Users have basic familiarity with S3 concepts (buckets, objects, folders as prefixes)
- Copy/Move operations are out of scope for v1
- Cross-bucket deep search (listing all objects to filter) is out of scope due to S3 API limitations

## Clarifications

### Session 2026-03-29

- **Q**: How should the web UI handle security and authentication between browser and server?  
  **A**: Server listens on localhost only. On startup, server generates a cryptographically secure random token (minimum 32 bytes) and prints to stdout the full URL including this token (e.g., `http://localhost:8080/<token>`). All HTTP requests must include this token in the URL path; requests without valid token receive 403 Forbidden. The token is stored in sessionStorage to persist across page refreshes. This provides single-user local access protection suitable for a development tool.

- **Q**: What pagination strategy should be used for large folders?  
  **A**: Pagination with configurable page size (default 100 items). The UI displays page numbers and navigation controls (First, Previous, Next, Last). Users can change page size to 25, 50, 100, or 250 items per page. Since S3 uses continuation tokens, the UI maps page numbers to these tokens for navigation. This provides predictable performance with S3's ListObjectsV2 API while maintaining familiar UX patterns similar to Google Drive and AWS Console.

- **Q**: How should file name conflicts be handled during upload?  
  **A**: Prompt for overwrite confirmation with option to auto-rename. When a file with the same name exists, display a modal dialog showing both files (existing file size, modification date vs new file) with three options: (1) Replace existing file, (2) Keep both files (auto-rename new file with timestamp suffix), (3) Skip this file. For batch uploads, provide "Apply to all" checkbox to avoid repeated prompts.

- **Q**: How should the bucket selection and filtering UI work?  
  **A**: Homepage displays a text input field at the top with a list of all accessible buckets below. As the user types, the bucket list filters in real-time to show only matching buckets. Clicking a bucket navigates to the bucket view. If the profile has a default bucket, the filter input is pre-filled with that bucket name on load, only that bucket is shown, and the UI automatically navigates to the bucket view.

- **Q**: How should file filtering work in the bucket view?  
  **A**: The bucket view displays a text input field at the top (prefixed with current path) for filtering files by name. As the user types, the file list filters in real-time to show only files on the current page whose names contain the typed text. This is client-side filtering of the current page, not a deep search across the entire bucket.

- **Q**: What keyboard shortcuts should be supported?  
  **A**: `/` focuses the filter input, `Esc` clears the current filter, `Ctrl/Cmd+K` opens quick navigation, arrow keys navigate the list, `Enter` opens selected item, `Delete` triggers delete (with confirmation). Shortcut hints should be visible in the UI.

- **Q**: How should empty filter results be handled?  
  **A**: When a filter returns no results, display an empty state message (e.g., "No buckets match 'xyz'") with a prominent "Clear filter" button that resets the filter input and shows all items again.

- **Q**: Should file filtering search across all pages of a bucket?  
  **A**: No, file filtering only filters the current page of results. S3's ListObjectsV2 API doesn't support server-side name filtering, so deep searching would require listing all objects first, which is too slow for large buckets. Users navigate through pages and filter within each page.
