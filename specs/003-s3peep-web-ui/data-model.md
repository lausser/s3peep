# Data Model: S3 File Browser

**Feature**: S3 File Browser Beautiful Web UI  
**Date**: March 29, 2026

## Entities

### Bucket

Represents an S3 bucket accessible to the user.

| Field | Type | Description |
|-------|------|-------------|
| name | string | Bucket name (unique identifier) |
| creation_date | timestamp | When bucket was created |
| region | string | AWS region where bucket resides |

**Validation**:
- Name follows S3 bucket naming rules
- Name is unique within the account

### FileObject

Represents an S3 object (file or folder prefix).

| Field | Type | Description |
|-------|------|-------------|
| key | string | Full S3 key path (e.g., "documents/report.pdf") |
| name | string | Display name (last segment of key) |
| size | int64 | Size in bytes (0 for folders) |
| last_modified | timestamp | Last modification time |
| is_folder | boolean | True if this is a folder prefix |
| file_type | enum | Classification: image, document, archive, video, audio, code, other |

**Validation**:
- Key must be valid S3 object key
- Size >= 0
- For folders: size = 0, ends with "/"

**State Transitions**:
- FileObject doesn't have state transitions (S3 is the source of truth)
- UI maintains temporary states: selected, highlighted, filtered

### UploadTask

Represents an active file upload operation.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique upload ID (UUID) |
| file_name | string | Original filename |
| target_key | string | Destination S3 key |
| progress_percentage | int | 0-100 upload progress |
| status | enum | pending, uploading, completed, failed |
| error_message | string | Error details if failed |
| started_at | timestamp | Upload start time |
| completed_at | timestamp | Upload completion time (if applicable) |

**State Machine**:
```
pending → uploading → completed
              ↓
            failed
```

**Validation**:
- progress_percentage: 0-100
- file_name: not empty
- target_key: valid S3 key

### SessionToken

Represents the authentication session.

| Field | Type | Description |
|-------|------|-------------|
| token | string | Cryptographically secure random string (32+ bytes, base64) |
| created_at | timestamp | When token was generated |
| expires_at | timestamp | Token expiration (24 hours from creation) |
| profile_name | string | Active profile name |

**Validation**:
- token: minimum 256 bits entropy
- expires_at: created_at + 24 hours
- profile_name: must exist in config

**Lifecycle**:
- Generated on server startup
- Stored in browser sessionStorage
- Expires after 24 hours or browser close

### SearchFilter

Represents current filter criteria for bucket/file lists.

| Field | Type | Description |
|-------|------|-------------|
| query_text | string | Search/filter text |
| file_type_filter | enum[] | Selected file types to show |
| date_range_start | timestamp | Filter files modified after |
| date_range_end | timestamp | Filter files modified before |

**Behavior**:
- Empty query_text shows all items
- Multiple file_type_filter values: OR logic
- Date range: inclusive of start, exclusive of end

### PaginationState

Represents pagination state for large folders.

| Field | Type | Description |
|-------|------|-------------|
| current_page | int | Current page number (1-based) |
| page_size | int | Items per page (25, 50, 100, or 250) |
| total_items | int | Total objects in current prefix (if known) |
| continuation_tokens | map[int,string] | Page number to S3 continuation token |

**Validation**:
- current_page: >= 1
- page_size: one of [25, 50, 100, 250]
- continuation_tokens: maps page number to S3 token

## Entity Relationships

```
SessionToken (1) ───> Profile (1)
                     └──> Bucket (N)
                          └──> FileObject (N)
                               └──> UploadTask (0..1) [if uploading]

SearchFilter (1) ───> Filters ───> FileObject (N) [current page]

PaginationState (1) ───> Controls ───> FileObject (N) [subset]
```

## API Request/Response Models

### ListBucketsRequest
- No parameters

### ListBucketsResponse
```json
{
  "buckets": [
    {
      "name": "my-bucket",
      "creation_date": "2024-01-15T10:30:00Z",
      "region": "us-east-1"
    }
  ]
}
```

### ListObjectsRequest
```json
{
  "bucket": "my-bucket",
  "prefix": "documents/",
  "continuation_token": "...", // optional, for pagination
  "max_keys": 100
}
```

### ListObjectsResponse
```json
{
  "objects": [
    {
      "key": "documents/report.pdf",
      "name": "report.pdf",
      "size": 1024000,
      "last_modified": "2024-03-20T14:22:00Z",
      "is_folder": false,
      "file_type": "document"
    }
  ],
  "is_truncated": true,
  "next_continuation_token": "...",
  "prefix": "documents/",
  "common_prefixes": ["documents/2024/"]
}
```

### UploadRequest
```json
{
  "bucket": "my-bucket",
  "key": "uploads/file.txt",
  "content_type": "text/plain"
}
```

### UploadResponse
```json
{
  "upload_id": "uuid",
  "status": "uploading",
  "progress_url": "/:token/api/upload/uuid/progress"
}
```

### DeleteRequest
```json
{
  "bucket": "my-bucket",
  "keys": ["file1.txt", "file2.txt"]
}
```

### DeleteResponse
```json
{
  "deleted": ["file1.txt"],
  "failed": [
    {
      "key": "file2.txt",
      "error": "Permission denied"
    }
  ]
}
```

### CreateFolderRequest
```json
{
  "bucket": "my-bucket",
  "folder_path": "newfolder/"
}
```

### ErrorResponse
```json
{
  "error": "Invalid bucket name",
  "code": "INVALID_BUCKET",
  "details": "Bucket name must be 3-63 characters..."
}
```

## Client-Side State

### AppState (sessionStorage + memory)

```typescript
interface AppState {
  // Authentication
  token: string;
  tokenExpiresAt: Date;
  
  // Navigation
  currentView: 'buckets' | 'files';
  selectedBucket: string | null;
  currentPath: string;
  
  // Filtering
  bucketFilter: SearchFilter;
  fileFilter: SearchFilter;
  
  // Pagination
  pagination: PaginationState;
  
  // Uploads
  activeUploads: UploadTask[];
  
  // Selection
  selectedFiles: string[]; // file keys
  
  // UI State
  theme: 'light' | 'dark';
  isLoading: boolean;
}
```

### Persistent State (sessionStorage)
- token
- theme preference
- selectedBucket (if default)

### Transient State (memory only)
- file lists (refetched on navigation)
- upload tasks
- filter text
- pagination position

## Validation Rules

### S3 Key Validation
- Length: 1-1024 bytes (UTF-8)
- Cannot contain: \x00, \x01-\x1F (control chars)
- Cannot start/end with "/" (except folders which must end with "/")
- Folder names must end with "/"

### Bucket Name Validation
- Length: 3-63 characters
- Lowercase letters, numbers, hyphens
- Cannot start/end with hyphen
- Cannot contain consecutive periods
- Cannot be formatted as IP address

### File Type Mapping
```javascript
const fileTypeMap = {
  image: ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg', '.bmp'],
  document: ['.pdf', '.doc', '.docx', '.txt', '.md', '.csv', '.xls', '.xlsx'],
  archive: ['.zip', '.tar', '.gz', '.bz2', '.7z', '.rar'],
  video: ['.mp4', '.avi', '.mov', '.mkv', '.webm'],
  audio: ['.mp3', '.wav', '.flac', '.aac', '.ogg'],
  code: ['.js', '.py', '.go', '.java', '.cpp', '.html', '.css', '.json', '.xml', '.yaml'],
  other: [] // default
};
```

## Indexing Strategy

**None required** - S3 is the source of truth. Client-side filtering is sufficient for current page (100 items max).

## Caching Strategy

**None required** - Always fetch fresh data from S3. This ensures consistency and avoids stale data issues.

Exception: sessionStorage for token persistence only.
