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

| Variable                      | Default  | Description                                                |
| ----------------------------- | -------- | ---------------------------------------------------------- |
| `PORT`                        | `:8097`  | HTTP listen address                                        |
| `NADAA_AUTH_TOKEN_SECRET`     | _(none)_ | HMAC secret verifying NADAA bearer tokens                  |
| `NADAA_AUTH_ALLOW_MOCK_ACTORS`| `false`  | Allow legacy `X-NADAA-Actor-*` headers (local dev/smoke)   |
| `NADAA_ALLOWED_ORIGINS`       | `*`      | Comma-separated CORS origins                               |
| `NADAA_ENV`                   | _(none)_ | `development` also allows localhost CORS origins           |

## Authority access

All preparedness endpoints require a verified bearer token
(`Authorization: Bearer nadaa...` issued by auth-service). Actor identity,
role, agency, district, and MFA status come from the token claims. The legacy
`X-NADAA-Actor-*` / `X-NADAA-MFA-Completed` headers are honored only when
`NADAA_AUTH_ALLOW_MOCK_ACTORS=true` (local development and smoke tests); the
service refuses to start with that setting unless `NADAA_ENV=development`.
`X-NADAA-Request-ID` remains an optional request-correlation header.

`system_admin` and `nadmo_officer` users can view and manage schools across
all districts. Other authority roles (`district_officer`, `dispatcher`,
`agency_admin`, `agency_viewer`) require a district claim and are restricted
to that district. `agency_viewer` is read-only: it cannot create or update
schools, drills, or readiness checks.

## Tests

```bash
go test ./...
```
