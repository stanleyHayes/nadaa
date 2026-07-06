# Incident Service

The incident service owns citizen disaster reports, media references, verification workflow, duplicate candidates, agency assignments, and incident timelines.

Current NADAA-030/NADAA-033/NADAA-041/NADAA-042/NADAA-043 endpoints:

- `GET /healthz`
- `POST /api/v1/incidents`
- `GET /api/v1/incidents`
- `GET /api/v1/incidents/{id}/duplicates`
- `POST /api/v1/incidents/{id}/verify`
- `PATCH /api/v1/incidents/{id}/status`
- `POST /api/v1/incidents/{id}/merge`
- `POST /api/v1/incidents/{id}/assignments`
- `GET /api/v1/incidents/audit`
- `POST /api/v1/media/uploads`
- `GET /api/v1/media`

## Verification And Status Workflow

Authority workflow endpoints use explicit local-development headers:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

`POST /api/v1/incidents/{id}/verify` moves `reported` or `under_review` incidents to `verified`, stores verifier metadata, and records an `incident.verified` audit event. Verification roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`.

`PATCH /api/v1/incidents/{id}/status` supports `reported`, `under_review`, `verified`, `assigned`, `response_en_route`, `on_scene`, `contained`, `recovery_ongoing`, `closed`, and `false_report`. The service enforces valid transitions, treats `closed` and `false_report` as terminal, and requires `resolutionNotes` for `closed` and `false_report`.

`POST /api/v1/incidents/{id}/assignments` assigns a verified incident to a response agency, stores active assignment metadata, appends an `incident.assigned` timeline event, and records an `incident.assigned` audit event. Assignment roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`; agency admins can assign only to their own agency.

`GET /api/v1/incidents?assignedToMe=true` filters the incident feed to active assignments for the request actor agency. `assignedAgencyId=<agency-id>` is available for authority readers that need to inspect a specific agency queue.

`GET /api/v1/incidents/audit?limit=50` returns latest incident workflow audit events with before/after snapshots for `system_admin`, `agency_admin`, and `nadmo_officer`.

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

`GET /api/v1/incidents/{id}/duplicates` returns the selected incident with full records for open duplicate candidates so the dashboard can compare reports side by side.

`POST /api/v1/incidents/{id}/merge` accepts `duplicateIncidentIds` plus a required `note`. The primary incident remains the operational record, merged duplicates are closed with `mergedIntoId`, `mergedBy`, `mergedAt`, and `mergeReason`, and the service appends timeline plus audit events for traceability.

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

Run a live local workflow smoke after starting the service on `:8084`:

```bash
pnpm smoke:incident-workflow
pnpm smoke:incident-assignment
pnpm smoke:incident-merge
```

## Notes

The current implementation uses an in-memory store to lock in the public API contract, validation behavior, duplicate candidate baseline, duplicate merge contract, incident workflow contract, agency assignment contract, and timeline event shape. PostGIS persistence, media upload storage, and dispatch-service extraction land in later stories.
