# HTTP Contract

Covered HTTP surfaces for spec 002:
- `GET /`
- `GET /api/buckets`
- `POST /api/buckets`
- `GET /api/list?prefix=`
- `GET /api/get?key=`

Contract rules:
- Tests assert exact status codes for success and negative cases.
- JSON responses must be validated with field-level assertions, not non-empty body checks.
- Download responses must assert required headers and returned content.
- Unsupported methods must fail explicitly rather than being treated as warning-only outcomes.
