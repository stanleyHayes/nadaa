# Incident Service

The incident service owns citizen disaster reports, media references, verification workflow, duplicate candidates, and incident timelines.

Current NADAA-030 endpoints:

- `GET /healthz`
- `POST /api/v1/incidents`
- `GET /api/v1/incidents`

## Run

```bash
go run .
```

The service listens on `:8084` by default. Override with `NADAA_INCIDENT_ADDR`.

## Rate Limiting

The starter service uses an in-memory per-client limiter.

Environment variables:

- `NADAA_INCIDENT_RATE_LIMIT`, default `60`.
- `NADAA_INCIDENT_RATE_WINDOW_SECONDS`, default `60`.

## Test

```bash
go test ./...
```

## Notes

The current implementation uses an in-memory store to lock in the public API contract and validation behavior. PostGIS persistence, media upload storage, duplicate detection, and verification workflow are planned in later stories.

