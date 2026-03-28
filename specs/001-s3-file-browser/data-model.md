# Data Model: S3 File Browser

## Entities

### Profile
Represents a stored S3 connection configuration.

| Field | Type | Validation | Description |
|-------|------|-----------|-------------|
| name | string | Required, unique | Profile identifier |
| region | string | Required | S3 region |
| access_key_id | string | Required | AWS access key |
| secret_access_key | string | Required | AWS secret key |
| endpoint_url | string | Optional | Custom S3 endpoint |

### Config
Root configuration file structure (~/.config/s3peep/config.json).

| Field | Type | Description |
|-------|------|-------------|
| active_profile | string | Currently selected profile name |
| profiles | []Profile | List of saved profiles |

### FileObject
Represents a file in S3.

| Field | Type | Description |
|-------|------|-------------|
| key | string | Full S3 key (path + filename) |
| name | string | Filename only |
| size | int64 | File size in bytes |
| last_modified | time.Time | Last modification timestamp |
| is_folder | bool | Whether this is a prefix/folder |

### Bucket
Represents an S3 bucket.

| Field | Type | Description |
|-------|------|-------------|
| name | string | Bucket name |
| region | string | Bucket region |

### TransferOperation
Represents an in-progress or completed file transfer.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique operation ID |
| type | enum | "upload" or "download" |
| local_path | string | Local file path |
| s3_key | string | S3 object key |
| status | enum | "pending", "in_progress", "completed", "failed", "cancelled" |
| progress | float64 | 0.0 to 1.0 progress |
| error | string | Error message if failed |

## State Transitions

### TransferOperation States
```
pending → in_progress → completed
                    → failed
                    → cancelled
```

## Relationships

- Config contains multiple Profiles
- Profile is selected as active_profile in Config
- S3Connection uses active Profile
- S3Connection lists Bucket contents as []FileObject
- TransferOperation references S3Connection for upload/download
