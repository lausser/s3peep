# S3Peep - S3 File Browser

A lightweight web-based file browser for S3-compatible object storage (AWS S3, MinIO, DigitalOcean Spaces, etc.)

## Features

- **Web-based UI** - Access your S3 buckets through a simple web interface
- **Multiple Profiles** - Manage multiple S3 configurations
- **Bucket Browsing** - Navigate through folders and files
- **File Downloads** - Download files directly from the browser
- **Filtering** - Filter buckets and files by name
- **S3-compatible** - Works with AWS S3, MinIO, and other S3-compatible services
- **Secure Token** - Each server instance uses a unique secure token for access

## Installation

### Prerequisites

- Go 1.21+ (for building from source)
- Or Podman/Docker (for building in container)

### Build Instructions

#### Option 1: With Go compiler on the host

```bash
# Clone the repository
git clone https://github.com/lausser/s3peep.git
cd s3peep

# Build the binary
go build -o s3peep ./cmd/s3peep

# Move to a location in your PATH (optional)
mv s3peep /usr/local/bin/
```

#### Option 2: Without Go compiler (using Podman)

```bash
# Clone the repository
git clone https://github.com/lausser/s3peep.git
cd s3peep

# Build using Podman with official Go image
podman run --rm -v "$PWD:/app:Z" -w /app docker.io/golang:1.24 \
    sh -c "go mod tidy && go build -o s3peep ./cmd/s3peep"

# The s3peep binary will be created in the current directory
```

#### Option 3: Without Go compiler (using Docker)

```bash
# Clone the repository
git clone https://github.com/lausser/s3peep.git
cd s3peep

# Build using Docker with official Go image
docker run --rm -v "$PWD:/app" -w /app golang:1.24 \
    sh -c "go mod tidy && go build -o s3peep ./cmd/s3peep"
```

## Configuration

### Initialize Configuration

Create a default configuration file:

```bash
s3peep init
```

This creates `~/.config/s3peep/config.json` with a sample profile.

### Configuration File Location

- Default: `~/.config/s3peep/config.json`
- Custom path: Use `--config FILE` flag
- Environment variable: Set `$CONFIG` environment variable

### Managing Profiles

Profiles store your S3 connection credentials and settings.

#### Add a Profile

```bash
# For AWS S3
s3peep profile add --name production \
    --region us-east-1 \
    --access-key YOUR_ACCESS_KEY \
    --secret YOUR_SECRET_KEY \
    --bucket my-bucket

# For MinIO or other S3-compatible services
s3peep profile add --name minio \
    --region us-east-1 \
    --access-key minioadmin \
    --secret minioadmin \
    --bucket mybucket \
    --endpoint http://localhost:9000
```

**Profile Options:**
- `--name`: Profile name (required)
- `--region`: AWS region (required)
- `--access-key`: Access key ID (required)
- `--secret`: Secret access key (required)
- `--bucket`: Default bucket (optional)
- `--endpoint`: Custom endpoint URL (optional, for MinIO, etc.)

#### List Profiles

```bash
s3peep profile list
```

Output shows all profiles with their configuration:
```
Profiles:
  - production (active)
    Endpoint: https://s3.amazonaws.com
    Region: us-east-1, Bucket: my-bucket
  - minio
    Endpoint: http://localhost:9000
    Region: us-east-1, Bucket: mybucket
```

#### Switch Active Profile

```bash
s3peep profile switch --name production
```

#### Remove a Profile

```bash
s3peep profile remove --name minio
```

## Usage

### Start the Web Server

```bash
# Start on default port 8080
s3peep serve

# Start on custom port
s3peep serve --port 9090

# Start with debug logging
s3peep serve --port 8080 --debug
```

**Server Options:**
- `--port`: HTTP server port (default: 8080)
- `--debug`: Enable debug logging (shows API requests)

### Access the Web UI

When the server starts, it prints an access URL with a unique secure token:

```
╔════════════════════════════════════════════════════════════════╗
║                    S3 File Browser Ready                      ║
╠════════════════════════════════════════════════════════════════╣
║  Access URL: http://localhost:8080/AbCdEfGhIjKlMnOpQrStUvWxYz  ║
╚════════════════════════════════════════════════════════════════╝
```

Open this URL in your browser to access the web interface.

**Security Note:** The token in the URL acts as a simple authentication mechanism. Keep the URL private.

## Command Reference

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `~/.config/s3peep/config.json` |
| `--debug` | Enable debug logging | `false` |

### Commands

#### `s3peep init`
Creates a default configuration file with a sample profile.

#### `s3peep profile add`
Adds a new S3 profile with credentials.

**Required flags:**
- `--name`: Profile name
- `--region`: AWS region
- `--access-key`: Access key ID
- `--secret`: Secret access key

**Optional flags:**
- `--bucket`: Default bucket name
- `--endpoint`: Custom S3 endpoint (for MinIO, etc.)

#### `s3peep profile list`
Lists all configured profiles with their settings.

#### `s3peep profile switch`
Sets the active profile to use.

**Required flags:**
- `--name`: Profile name to activate

#### `s3peep profile remove`
Deletes a profile.

**Required flags:**
- `--name`: Profile name to remove

#### `s3peep serve`
Starts the web server.

**Flags:**
- `--port`: Server port (default: 8080)
- `--debug`: Enable debug logging

## Web Interface Features

### Bucket Selection
- View all accessible buckets
- Filter buckets by name
- Click a bucket to browse its contents

### File Browsing
- Navigate through folders
- Breadcrumb navigation (click to go back)
- Parent folder shortcut (`..`)
- Filter files and folders by name

### File Operations
- **Download**: Click any file to download
- Files download to your browser's default download location

## Development

### Project Structure

```
s3peep/
├── cmd/s3peep/         # Main application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP request handlers
│   └── s3/            # S3 client wrapper
├── test/              # Integration tests
└── go.mod             # Go module definition
```

### Running Tests

```bash
# Run tests in containerized environment
cd test
podman-compose up
```

## Troubleshooting

### "No active profile" Error
Create a profile first:
```bash
s3peep profile add --name myprofile --region us-east-1 --access-key KEY --secret SECRET
```

### "Failed to connect to S3" Error
Check your endpoint URL and credentials:
```bash
s3peep profile list
```

### Downloads Not Working
The Web UI uses a simple anchor tag download method that works with most browsers. Files download directly to your browser's default download folder (usually ~/Downloads).

## License

MIT License - See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and feature requests, please use the GitHub issue tracker.
