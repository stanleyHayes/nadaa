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

Authority endpoints require `X-NADAA-Actor-ID`, `X-NADAA-Actor-Role`, `X-NADAA-Agency-ID`, and `X-NADAA-MFA-Completed: true` headers. Allowed roles: `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher`.

## Configuration

| Environment variable    | Default | Description                    |
| ----------------------- | ------- | ------------------------------ |
| `NADAA_CAMPAIGN_ADDR`   | `:8103` | HTTP listen address            |
| `NADAA_ALLOWED_ORIGINS` | `*`     | Comma-separated CORS allowlist |

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
