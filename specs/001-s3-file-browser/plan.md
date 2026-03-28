# Implementation Plan: S3 File Browser

**Branch**: `001-s3-file-browser` | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-s3-file-browser/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

A web-based file browser/downloader/uploader for S3-compatible storage. Users manage connection profiles stored in JSON config, then browse buckets and transfer files via embedded web UI.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: aws-sdk-go-v2, embedded HTTP server (stdlib)  
**Storage**: JSON config file (~/.config/s3peep/config.json)  
**Testing**: Go testing package (stdlib)  
**Target Platform**: Linux/Windows/macOS (cross-platform)  
**Project Type**: CLI tool with embedded web UI  
**Performance Goals**: Responsive UI (<200ms for file listings), efficient large file transfers (multipart)  
**Constraints**: Slim dependencies, easily dockerizable  
**Scale/Scope**: Single-user desktop tool

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Note**: No constitution file present. Proceeding with project defaults:
- Single project structure (Go binary with embedded web UI)
- Standard Go testing patterns
- Minimal external dependencies

## Project Structure

### Documentation (this feature)

```
specs/001-s3-file-browser/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── tasks.md             # (NOT created by /speckit.plan)
└── checklists/
```

### Source Code (repository root)

```
cmd/s3peep/
└── main.go              # Entry point

internal/
├── config/
│   └── config.go        # Profile management, config file I/O
├── s3/
│   └── client.go        # S3 operations (list, get, put)
├── handlers/
│   └── api.go           # HTTP handlers for web UI
└── ui/
    └── assets.go        # Embedded web assets

web/                     # Web UI source (for development)
├── index.html
├── styles.css
└── app.js

Dockerfile               # Multi-stage build

go.mod
go.sum
```

**Structure Decision**: Single Go project with embedded web UI. All code in Go except web assets in `web/` directory.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
