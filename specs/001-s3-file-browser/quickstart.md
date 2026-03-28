# Quickstart: S3 File Browser

## Prerequisites

- Go 1.21+
- Docker (for containerized usage)

## Installation

### From Source

```bash
git clone <repo>
cd s3peep
go build -o s3peep ./cmd/s3peep
./s3peep --help
```

### Using Docker

```bash
docker build -t s3peep .
docker run -p 8080:8080 -v ~/.config/s3peep:/home/s3peep/.config/s3peep s3peep
```

## Configuration

The application stores profiles in `~/.config/s3peep/config.json`:

```json
{
  "active_profile": "my-space",
  "profiles": [
    {
      "name": "my-space",
      "region": "us-east-1",
      "access_key_id": "YOUR_ACCESS_KEY",
      "secret_access_key": "YOUR_SECRET_KEY",
      "endpoint_url": "https://object.storage.example.com"
    }
  ]
}
```

### Managing Profiles

```bash
# Add a new profile
s3peep profile add --name my-space --region us-east-1 \
  --access-key YOUR_KEY --secret YOUR_SECRET \
  --endpoint https://object.storage.example.com

# Switch profile
s3peep profile switch --name my-space

# List profiles
s3peep profile list
```

## Usage

### Start the web interface

```bash
s3peep serve --port 8080
```

Then open http://localhost:8080 in your browser.

### Command-line operations

```bash
# List files
s3peep ls /path/in/bucket

# Download a file
s3peep get path/to/file.txt

# Upload a file
s3peep put local-file.txt /destination/
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| S3PEEP_CONFIG | Custom config file path |
| S3PEEP_PORT | Default HTTP server port |
| S3PEEP_PROFILE | Profile to use by default |
