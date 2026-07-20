# campaign-service

Public disaster education campaign backend for NADAA.

## Scope

- Stores preparedness campaigns tied to hazards, regions, languages, and publishing windows.
- Serves public campaign listings and detail pages.
- Provides seasonal campaign templates.
- Tracks mocked reach and engagement metrics (first pass).
- Links campaigns to emergency guides and alerts.

## API

Base path: `/api/v1`

### Public endpoints

- `GET /healthz`
- `GET /api/v1/campaigns?region=&language=&hazard=&status=`
- `GET /api/v1/campaigns/{id}`
- `GET /api/v1/campaigns/{id}/metrics`
- `GET /api/v1/campaign-templates`

### Authority endpoints

- `POST /api/v1/campaigns`
- `PUT /api/v1/campaigns/{id}`

Authority endpoints require an auth-service bearer token (`Authorization: Bearer nadaa.<payload>.<sig>`) whose claims carry an allowed role, an agency id, and `mfa: true`. Allowed roles: `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher`. For local development and smoke tests only, setting `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` also honors the legacy `X-NADAA-Actor-ID`, `X-NADAA-Actor-Role`, `X-NADAA-Agency-ID`, and `X-NADAA-MFA-Completed: true` headers. The service refuses to start when `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` is set without `NADAA_ENV=development`.

A create request without an explicit `status` defaults to `draft`. Public callers may only filter by `status=published`; any other status filter returns `403 forbidden`.

## Configuration

| Environment variable          | Default | Description                                                  |
| ----------------------------- | ------- | ------------------------------------------------------------ |
| `NADAA_CAMPAIGN_ADDR`         | `:8103` | HTTP listen address                                          |
| `NADAA_ALLOWED_ORIGINS`       | `*`     | Comma-separated CORS allowlist                               |
| `NADAA_AUTH_TOKEN_SECRET`     | (empty) | HMAC-SHA256 key verifying auth-service bearer tokens         |
| `NADAA_AUTH_ALLOW_MOCK_ACTORS`| `false` | Honor legacy X-NADAA-Actor-* headers (local dev/smoke only)  |
| `NADAA_ENV`                   | (empty) | `development` also echoes localhost CORS origins             |

## Run locally

```bash
cd services/campaign-service
go run ./cmd/server
```

The service listens on `:8103` by default.

## Test

```bash
cd services/campaign-service
go test ./...
go vet ./...
```

## Docker

```bash
docker build -f services/campaign-service/Dockerfile -t nadaa/campaign-service:local .
```
