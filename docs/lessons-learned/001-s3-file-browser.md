# Lessons Learned: S3 File Browser (001-s3-file-browser)

## Project Overview
A simple file browser/downloader/uploader for S3-compatible storage with a web UI.

## Technology Stack
- **Language**: Go 1.21+
- **S3 SDK**: aws-sdk-go-v2
- **UI**: Embedded web server (stdlib HTTP)
- **Config**: JSON file (~/.config/s3peep/config.json)

## Key Decisions

### 1. Web-based UI vs Native GUI
- Chose web-based UI (embedded HTTP server) for cross-platform compatibility
- User was flexible: "I leave it up to you if you create a native graphical interface or if the tool opens a port so that the UI can be a normal web browser"
- Advantage: Single binary, no platform-specific UI code

### 2. Config Storage
- Stored as JSON at ~/.config/s3peep/config.json
- Supports multiple profiles for different S3 accounts
- Active profile enables quick switching
- Credentials stored in plain JSON (user accepted this security model)

### 3. Bucket Selection
- Made bucket optional - browser lists all buckets when no bucket selected
- Allows dynamic bucket selection via dropdown in UI

## Technical Challenges & Solutions

### Challenge 1: AWS SDK Endpoint Resolution
**Problem**: The `s3.WithEndpointResolverWithOptions` function no longer exists in newer versions of aws-sdk-go-v2.

**Solution**: Use `BaseEndpoint` option instead:
```go
opts = append(opts, func(o *s3.Options) {
    o.BaseEndpoint = &profile.EndpointURL
})
```

### Challenge 2: Go Flag Parsing
**Problem**: Standard `flag.Parse()` stops at first non-flag argument (e.g., "init", "profile"), so `--config` after the subcommand didn't work.

**Solution**: Use `flag.CommandLine.Parse(os.Args[1:])` to parse all flags, then use `flag.Args()` to get commands.

### Challenge 3: JSON Field Serialization
**Problem**: Frontend received `undefined` values for bucket list.

**Solution**: Add explicit JSON tags to struct fields:
```go
type Bucket struct {
    Name string `json:"name"`
}
```

### Challenge 4: Profile Auto-activation
**Problem**: Adding a new profile didn't auto-activate it.

**Solution**: Modified `AddProfile` to set `cfg.ActiveProfile = profile.Name` after adding.

## Failures

1. **Frontend JSON Parsing**: Initially didn't add JSON tags to S3 response structs, causing frontend to show "undefined"
2. **AWS SDK Version**: Upgraded Go to 1.24 because aws-sdk-go-v2 required newer Go version
3. **Initial UI Design**: Tried to show files when no bucket selected - fixed by adding bucket selector dropdown
4. **Flag Parsing**: First attempt at config flag didn't work due to Go's flag package behavior

## Project Structure
```
cmd/s3peep/main.go       - Entry point
internal/config/         - Profile management
internal/s3/             - S3 client operations
internal/handlers/       - HTTP handlers + embedded UI
web/                     - Development UI files (not embedded)
```

## Build Commands
```bash
# Build binary
go build -o s3peep ./cmd/s3peep

# Run
./s3peep init
./s3peep profile add --name myspace --region us-east-1 --access-key KEY --secret SECRET
./s3peep serve --port 8080

# Docker
docker build -t s3peep .
```

## Open Issues
- Frontend UI still has bugs with bucket selection (shows "undefined" in some cases)
- Upload functionality not implemented
- No error handling for network failures during file transfer

## Summary
The project is a working CLI tool with profile management and basic S3 browsing capability. Main challenges were around AWS SDK API changes and proper JSON serialization for the frontend.
