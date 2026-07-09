# Damage Claim Service

The damage-claim-service manages insurance and property damage claim intake, verification, updates, closure, and export for the NADAA platform.

## Current endpoints

- `GET /health`
- `POST /claims` — citizen intake (public)
- `GET /claims?status=&verificationStatus=&incidentId=&q=` — authority list
- `GET /claims/{id}` — authority detail
- `PATCH /claims/{id}` — authority/citizen update of description/amount/photos
- `POST /claims/{id}/verify` — authority verification
- `POST /claims/{id}/close` — authority close with reason
- `GET /claims/{id}/export?format=csv|pdf` — export single claim

Authority endpoints require NADAA authority headers:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

Allowed authority roles are `system_admin`, `nadmo_officer`, `district_officer`, `dispatcher`, `police`, `insurance_officer`, `fire`, and `ambulance`.

## Run

```bash
go run ./cmd/server
```

The service listens on `:8098` by default. Override with `PORT`.

## Test

```bash
go test ./...
```

```bash
go build ./cmd/server
```

## Environment variables

| Variable                | Default                 | Description                                                                              |
| ----------------------- | ----------------------- | ---------------------------------------------------------------------------------------- |
| `PORT`                  | `:8098`                 | HTTP listen address.                                                                     |
| `NADAA_ALLOWED_ORIGINS` | `*`                     | Comma-separated CORS origin allowlist. Use `*` or leave empty for any origin.            |
| `INCIDENT_SERVICE_URL`  | `http://localhost:8081` | Base URL for incident-service lookups to enrich claims with incident reference/location. |

## Notes

The current implementation uses an in-memory fixture store so citizen and authority surfaces can integrate against a stable API contract. Persistence and event sourcing can replace the store later without changing endpoint shapes.
