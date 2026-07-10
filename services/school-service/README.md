# School Service

Emergency preparedness records for Ghanaian schools. Stores school profiles,
drill histories, and readiness checks, with district-scoped access for
authority users.

## Local run

```bash
cd services/school-service
go run ./cmd/server
```

The service listens on `:8097` by default. Override with `PORT`.

## Endpoints

- `GET /health`
- `GET /api/v1/schools` – list school summaries (district filter with `?district=...`)
- `POST /api/v1/schools` – create a school profile
- `GET /api/v1/schools/{id}` – get a school profile
- `PUT /api/v1/schools/{id}` – update a school profile
- `GET /api/v1/schools/{id}/drills` – list drill records
- `POST /api/v1/schools/{id}/drills` – add a drill record
- `GET /api/v1/schools/{id}/readiness` – get the latest readiness check
- `POST /api/v1/schools/{id}/readiness` – submit a readiness check

## Environment variables

| Variable                | Default                 | Description                  |
| ----------------------- | ----------------------- | ---------------------------- |
| `PORT`                  | `:8097`                 | HTTP listen address          |
| `RISK_SERVICE_URL`      | `http://localhost:8082` | Base URL for risk-service    |
| `NADAA_ALLOWED_ORIGINS` | `*`                     | Comma-separated CORS origins |

## Authority headers

All preparedness endpoints require authority headers:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Actor-District` (used to scope district officers to their district)
- `X-NADAA-Request-ID` (optional)

`system_admin` users can view and manage schools across all districts. Other
authority roles are restricted to `X-NADAA-Actor-District`.

## Tests

```bash
go test ./...
```
