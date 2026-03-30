# Quickstart Guide: S3 File Browser Web UI

**Feature**: S3 File Browser Beautiful Web UI  
**Branch**: 003-s3peep-web-ui

## Prerequisites

- Go 1.24 or later
- S3-compatible storage (AWS S3, MinIO, etc.)
- Modern web browser (Chrome 100+, Firefox 100+, Safari 15+, Edge 100+)

## Installation

```bash
# Clone or navigate to the repository
cd /home/lausser/git/s3peep

# Build the binary
go build -o s3peep ./cmd/s3peep

# Or install directly
go install ./cmd/s3peep
```

## Configuration

### 1. Initialize Configuration

```bash
# Create default config file
./s3peep init
```

This creates `~/.config/s3peep/config.json`

### 2. Add S3 Profile

```bash
# Add a profile
./s3peep profile add \
  --name my-profile \
  --region us-east-1 \
  --access-key YOUR_ACCESS_KEY \
  --secret YOUR_SECRET_KEY \
  --bucket my-default-bucket \
  --endpoint https://s3.amazonaws.com  # optional, for S3-compatible services
```

### 3. List and Switch Profiles

```bash
# List all profiles
./s3peep profile list

# Switch to a profile
./s3peep profile switch --name my-profile
```

## Running the Web UI

### Start the Server

```bash
# Start on default port 8080
./s3peep serve

# Start on custom port
./s3peep serve --port 9000
```

### Access the Web UI

On startup, the server prints:

```
Starting server on port 8080...
Access the web UI at: http://localhost:8080/abc123def456...
```

**Copy the full URL** (including the token) and paste it into your browser.

## Using the Web UI

### Homepage - Bucket Selection

1. **Filter buckets**: Type in the search box at the top
   - Matching buckets appear instantly
   - Press `/` to focus the filter quickly
   - Press `Esc` to clear the filter

2. **Select a bucket**: Click on a bucket name to view its files

3. **Default bucket**: If your profile has a default bucket, you'll be taken directly to it

### File Browser

1. **Navigate folders**: Click folder names to enter them
2. **Go back**: Click breadcrumb links (Home > folder > subfolder)
3. **Download files**: Click on file names
4. **Filter files**: Type in the filter box to search current page

### Upload Files

**Method 1 - Drag & Drop**:
1. Drag files from your desktop into the browser window
2. Files upload to the current folder

**Method 2 - File Picker**:
1. Click the "Upload" button
2. Select files from the dialog
3. Click "Open" to upload

**Upload Conflicts**:
If a file with the same name exists, you'll see a dialog with options:
- **Replace**: Overwrite existing file
- **Keep both**: Auto-rename new file with timestamp
- **Skip**: Don't upload this file
- **Apply to all**: Use same choice for remaining files

### Create Folders

1. Click "New Folder" button
2. Enter folder name
3. Click "Create"

Folder names must:
- Be unique in current location
- Not contain special characters
- End with `/` (added automatically)

### Delete Files/Folders

1. Select items using checkboxes
2. Click "Delete" button
3. Confirm deletion

**Warning**: Deleting a folder deletes all contents recursively!

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `/` | Focus filter input |
| `Esc` | Clear filter / close modal |
| `Ctrl/Cmd + K` | Quick navigation |
| `↑/↓` | Navigate list |
| `Enter` | Open selected item |
| `Delete` | Delete selected (with confirmation) |

### Switching Themes

Click the theme toggle in the header to switch between:
- Light theme
- Dark theme

## Troubleshooting

### "Session expired" error

The token expires after 24 hours or when you close the browser. Restart the server to get a new URL.

```bash
# Restart the server
./s3peep serve
```

### "Permission denied" error

Your S3 credentials don't have permission for this operation. Check:
- Profile credentials are correct
- IAM policy allows the operation
- Bucket name is correct

### Can't see expected buckets

Only buckets visible to your access key are shown. Check:
- You're using the correct profile
- The profile has the right credentials
- The S3 endpoint is correct (for non-AWS S3)

### Slow loading

Large buckets may take time to list. The UI shows:
- Skeleton loaders while loading
- Pagination for folders with many files

Adjust page size (25/50/100/250 items) to balance speed vs navigation.

### Upload fails

Check:
- File size under 5GB
- Stable internet connection
- S3 bucket has write permissions
- For large files (>100MB), multipart upload is used automatically

## Configuration File Format

```json
{
  "active_profile": "my-profile",
  "profiles": [
    {
      "name": "my-profile",
      "region": "us-east-1",
      "access_key_id": "AKIA...",
      "secret_access_key": "...",
      "endpoint_url": "https://s3.amazonaws.com",
      "bucket": "my-default-bucket"
    }
  ]
}
```

## Security Notes

- **Token in URL**: The token authenticates you. Keep the URL private.
- **Localhost only**: Server binds to 127.0.0.1 (localhost) only
- **No HTTPS**: Not needed for localhost, but consider proxy for remote access
- **Session storage**: Token stored in browser sessionStorage, cleared on close
- **No password**: Single-user tool, no authentication beyond the token

## Development

### Run Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/handlers
```

### Project Structure

```
s3peep/
├── cmd/s3peep/          # Main entry point
├── internal/
│   ├── config/          # Profile management
│   ├── handlers/        # HTTP handlers
│   └── s3/              # S3 client wrapper
├── web/
│   ├── static/          # CSS, JS files
│   └── templates/       # HTML templates
└── specs/               # Feature specs and plans
```

### API Endpoints

All endpoints require token in URL: `/:token/...`

- `GET /:token/api/buckets` - List buckets
- `GET /:token/api/buckets/:bucket/objects` - List objects
- `GET /:token/api/buckets/:bucket/download` - Download file
- `POST /:token/api/buckets/:bucket/upload` - Upload file
- `DELETE /:token/api/buckets/:bucket/objects` - Delete objects
- `PUT /:token/api/buckets/:bucket/folders` - Create folder

## Getting Help

- Check the spec: `specs/003-s3peep-web-ui/spec.md`
- Review the data model: `specs/003-s3peep-web-ui/data-model.md`
- Run with debug logging: `LOG_LEVEL=debug ./s3peep serve`
