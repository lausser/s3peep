# API Contract: S3 File Browser

**Base URL**: `http://localhost:8080/:token`  
**Authentication**: Token in URL path (validated server-side)

## Endpoints

### List Buckets

**GET** `/api/buckets`

Returns all buckets accessible to the current profile.

**Response**:
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

**Errors**:
- `403 Forbidden`: Invalid or expired token
- `500 Internal Server Error`: S3 connection failed

---

### Select Bucket

**POST** `/api/buckets/select`

Sets the active bucket in the profile.

**Request**:
```json
{
  "bucket": "my-bucket"
}
```

**Response**:
```json
{
  "status": "ok",
  "bucket": "my-bucket"
}
```

**Errors**:
- `400 Bad Request`: Bucket name missing or invalid
- `403 Forbidden`: Invalid token
- `404 Not Found`: Bucket not found or no access

---

### List Objects

**GET** `/api/buckets/{bucket}/objects`

Lists objects (files and folders) in a bucket.

**Query Parameters**:
- `prefix` (string, optional): Path prefix to list under (e.g., "documents/")
- `continuation_token` (string, optional): S3 continuation token for pagination
- `max_keys` (integer, optional): Maximum items per page (25, 50, 100, 250). Default: 100

**Response**:
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
    },
    {
      "key": "documents/2024/",
      "name": "2024/",
      "size": 0,
      "last_modified": "2024-01-01T00:00:00Z",
      "is_folder": true,
      "file_type": "other"
    }
  ],
  "is_truncated": true,
  "next_continuation_token": "eyJDb250aW51YXRpb25Ub2tlbiI6...",
  "prefix": "documents/",
  "common_prefixes": ["documents/2024/", "documents/2023/"]
}
```

**Errors**:
- `400 Bad Request`: Invalid max_keys value
- `403 Forbidden`: Invalid token or no bucket access
- `404 Not Found`: Bucket not found

---

### Download File

**GET** `/api/buckets/{bucket}/download`

Downloads a file from S3.

**Query Parameters**:
- `key` (string, required): Full S3 key of the file

**Response**: Binary file data with headers:
- `Content-Type: application/octet-stream`
- `Content-Disposition: attachment; filename="original-filename"`

**Errors**:
- `400 Bad Request`: Key parameter missing
- `403 Forbidden`: Invalid token or no access
- `404 Not Found`: File not found

---

### Upload File

**POST** `/api/buckets/{bucket}/upload`

Uploads a file to S3.

**Request**: `multipart/form-data`
- `file` (file, required): File to upload
- `key` (string, required): Target S3 key (including folder path)
- `overwrite` (boolean, optional): If true, overwrite existing. Default: false

**Response**:
```json
{
  "upload_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "key": "documents/uploaded-file.pdf",
  "size": 1024000
}
```

**Errors**:
- `400 Bad Request`: File or key missing
- `403 Forbidden`: Invalid token or no write access
- `409 Conflict`: File exists and overwrite=false
- `413 Payload Too Large`: File exceeds 5GB limit
- `500 Internal Server Error`: Upload failed

---

### Create Folder

**PUT** `/api/buckets/{bucket}/folders`

Creates a folder (S3 prefix) in the bucket.

**Request**:
```json
{
  "folder_path": "newfolder/"
}
```

**Response**:
```json
{
  "status": "created",
  "key": "newfolder/"
}
```

**Errors**:
- `400 Bad Request`: Invalid folder path
- `403 Forbidden`: Invalid token or no write access
- `409 Conflict`: Folder already exists

---

### Delete Objects

**DELETE** `/api/buckets/{bucket}/objects`

Deletes one or more files or folders.

**Request**:
```json
{
  "keys": ["file1.txt", "folder/", "file2.pdf"]
}
```

**Response**:
```json
{
  "deleted": ["file1.txt", "file2.pdf"],
  "failed": [
    {
      "key": "folder/",
      "error": "Folder not empty",
      "code": "FOLDER_NOT_EMPTY"
    }
  ]
}
```

**Notes**:
- Deleting a folder attempts to delete all contents recursively
- Non-empty folders may fail with "FOLDER_NOT_EMPTY"

**Errors**:
- `400 Bad Request`: Keys array missing or empty
- `403 Forbidden`: Invalid token or no delete access
- `500 Internal Server Error`: Partial failure (check response)

---

### Get Profile Info

**GET** `/api/profile`

Returns current profile information.

**Response**:
```json
{
  "name": "my-profile",
  "region": "us-east-1",
  "bucket": "my-default-bucket",
  "endpoint_url": "https://s3.amazonaws.com"
}
```

**Note**: Sensitive fields (access_key_id, secret_access_key) are NOT included.

**Errors**:
- `403 Forbidden`: Invalid token

---

### Get Upload Progress

**GET** `/api/upload/{upload_id}/progress`

Returns upload progress for multipart uploads.

**Response**:
```json
{
  "upload_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "uploading",
  "progress_percentage": 45,
  "bytes_uploaded": 461373440,
  "bytes_total": 1024000000,
  "speed_mbps": 12.5,
  "eta_seconds": 45
}
```

**Status values**: `pending`, `uploading`, `completed`, `failed`

**Errors**:
- `404 Not Found`: Upload ID not found

---

## Error Response Format

All errors return JSON with this structure:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": "Optional additional details"
}
```

**Error Codes**:
- `INVALID_TOKEN`: Authentication failed
- `TOKEN_EXPIRED`: Session expired, restart server
- `BUCKET_NOT_FOUND`: Bucket doesn't exist or no access
- `OBJECT_NOT_FOUND`: File/folder not found
- `ACCESS_DENIED`: No permission for operation
- `INVALID_KEY`: Malformed S3 key
- `FILE_TOO_LARGE`: Exceeds 5GB limit
- `UPLOAD_FAILED`: Generic upload error
- `DELETE_FAILED`: Generic delete error

## Authentication

All requests must include the token in the URL path:

```
http://localhost:8080/abc123def456/api/buckets
                └──────┬──────┘
                    token
```

The token is generated on server startup and printed to stdout. Store it in sessionStorage for persistence across page refreshes.

**Token Validation**:
- Server validates token on every request
- Invalid/expired tokens return `403 Forbidden`
- Tokens expire after 24 hours or on server restart

## CORS

The server allows cross-origin requests from any origin (localhost-only tool):

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

## Rate Limiting

No explicit rate limiting. S3 API limits apply:
- List operations: ~100-200 requests/sec per bucket
- Upload: Standard S3 limits
- Download: Standard S3 limits

The UI implements client-side debouncing for filter inputs (300ms).
