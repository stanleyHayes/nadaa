# Incident Service

The incident service owns citizen disaster reports, media references, verification workflow, duplicate candidates, and incident timelines.

Current NADAA-030/NADAA-033 endpoints:

- `GET /healthz`
- `POST /api/v1/incidents`
- `GET /api/v1/incidents`
- `POST /api/v1/media/uploads`
- `GET /api/v1/media`

## Media Upload Flow

`POST /api/v1/media/uploads` creates private media metadata and returns a controlled development upload target. Incident reports can reference returned media IDs. Known media IDs are marked `linked` when the incident is created.

Supported content types and limits:

- Images: `image/jpeg`, `image/png`, `image/webp`, up to 10 MB.
- Video: `video/mp4`, `video/quicktime`, up to 100 MB.
- Audio: `audio/mpeg`, `audio/mp4`, `audio/wav`, up to 25 MB.

## Duplicate Candidate Baseline

When a report is created, the service compares it against existing same-hazard reports. Candidates are scored with:

- Location distance within 750 meters.
- Report time within 3 hours.
- Description token similarity.

The top candidates are stored on incident records and returned by `POST /api/v1/incidents` and `GET /api/v1/incidents`. This is a dispatcher review aid only: the service does not automatically merge, hide, delete, or downgrade any report.

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

The current implementation uses an in-memory store to lock in the public API contract and validation behavior. PostGIS persistence, media upload storage, duplicate merge review, and verification workflow are planned in later stories.
