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

Authority endpoints require a NADAA bearer token issued by auth-service:

- `Authorization: Bearer nadaa.<payload>.<sig>` — claims supply the actor id (`sub`), role (`role`), agency (`agencyId`), district (`district`), and MFA status (`mfa`).

For local development and smoke tests only, setting `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` makes the service fall back to the legacy headers `X-NADAA-Actor-ID`, `X-NADAA-Actor-Role`, `X-NADAA-Agency-ID`, and `X-NADAA-MFA-Completed: true` when no valid bearer token is presented. `X-NADAA-Request-ID` is always honored as a tracing header.

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

| Variable                       | Default                          | Description                                                                                          |
| ------------------------------ | -------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `PORT`                         | `:8098`                          | HTTP listen address.                                                                                 |
| `NADAA_ALLOWED_ORIGINS`        | `*`                              | Comma-separated CORS origin allowlist. Use `*` or leave empty for any origin.                        |
| `NADAA_ENV`                    | _(empty)_                        | When `development`, localhost/127.0.0.1 origins bypass the configured CORS allowlist.                |
| `INCIDENT_SERVICE_URL`         | `http://localhost:8084/api/v1`   | Base URL for incident-service lookups to enrich claims with incident reference/location.             |
| `NADAA_AUTH_TOKEN_SECRET`      | _(empty)_                        | HMAC-SHA256 key verifying NADAA bearer tokens. Empty → authority requests are rejected (401).        |
| `NADAA_INTERNAL_SERVICE_TOKEN` | _(empty)_                        | Sent as `X-NADAA-Service-Token` on incident lookups when the caller presents no bearer token.        |
| `NADAA_AUTH_ALLOW_MOCK_ACTORS` | `false`                          | When `true`, legacy `X-NADAA-Actor-*` headers are honored for local development and smoke tests.     |

## Notes

The current implementation uses an in-memory fixture store so citizen and authority surfaces can integrate against a stable API contract. Persistence and event sourcing can replace the store later without changing endpoint shapes.
