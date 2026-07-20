# Incident Service

The incident service owns citizen disaster reports, media references, verification workflow, duplicate candidates, abuse/spam review, agency assignments, and incident timelines.

Current NADAA-030/NADAA-033/NADAA-041/NADAA-042/NADAA-043/NADAA-091 endpoints:

- `GET /healthz`
- `POST /api/v1/incidents`
- `GET /api/v1/incidents`
- `GET /api/v1/incidents/{id}`
- `GET /api/v1/incidents/{id}/duplicates`
- `POST /api/v1/incidents/{id}/verify`
- `PATCH /api/v1/incidents/{id}/status`
- `POST /api/v1/incidents/{id}/merge`
- `POST /api/v1/incidents/{id}/abuse-review`
- `POST /api/v1/incidents/{id}/assignments`
- `GET /api/v1/incidents/audit`
- `POST /api/v1/media/uploads`
- `PUT /api/v1/media/{id}/content`
- `GET /api/v1/media/{id}/content`
- `GET /api/v1/media`

## Authentication

Authority endpoints require a signed NADAA bearer token (`Authorization: Bearer nadaa.<payload>.<sig>`) issued by auth-service and verified with the shared `NADAA_AUTH_TOKEN_SECRET`. The actor context (user id, role, agency id, district, MFA) comes from verified token claims only.

Volunteer registration (`POST /api/v1/volunteers`) is citizen self-service: it requires a verified citizen bearer token (`typ=citizen`) and derives `citizenUserId` from the token subject — a mismatched body value is rejected with `403`. Registration is idempotent per citizen: re-registering returns the existing profile with `200` instead of creating a duplicate.

For local development and smoke tests, the legacy `X-NADAA-*` actor headers below are honored only when `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` (which itself is rejected at startup unless `NADAA_ENV=development`):

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

Volunteer task endpoints (`GET /api/v1/volunteers/{id}/tasks`, `PATCH /api/v1/volunteer-tasks/{id}/status`, `POST /api/v1/volunteer-tasks/{id}/observations`) accept either a verified agency token or a verified citizen token whose `sub` matches the volunteer's registered `citizenUserId`. The status and observation mutations require a volunteer-task workflow role for agency actors (`system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher`, `responder` — read-only `agency_viewer` is rejected); the task list remains readable by any incident read role. Timeline and audit events are attributed to the verified actor, never to the client-supplied `volunteerId`, and escalating observations honor the same terminal-state transition guard as status updates.

## Verification And Status Workflow

`POST /api/v1/incidents/{id}/verify` moves `reported` or `under_review` incidents to `verified`, stores verifier metadata, and records an `incident.verified` audit event. Verification roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`. `PATCH /api/v1/incidents/{id}/status` rejects transitions to `verified` for other workflow roles (for example `responder`).

`PATCH /api/v1/incidents/{id}/status` supports `reported`, `under_review`, `verified`, `assigned`, `response_en_route`, `on_scene`, `contained`, `recovery_ongoing`, `closed`, and `false_report`. The service enforces valid transitions, treats `closed` and `false_report` as terminal, and requires `resolutionNotes` for `closed` and `false_report`. Transitions to `false_report` are restricted to the abuse-review roles (`system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher`); other workflow roles such as `responder` receive `403`, matching the narrower gate on `POST /api/v1/incidents/{id}/abuse-review`.

`POST /api/v1/incidents/{id}/abuse-review` lets dispatchers clear suspicious report signals, keep a report under monitoring, or mark a report false with required `resolutionNotes`. The endpoint records `abuseReviewDecision`, reviewer metadata, timeline events, and audit events. Suspicious scores never block report creation; life-threatening reports remain live and retain priority review.

`POST /api/v1/incidents/{id}/assignments` assigns a verified incident to a response agency, stores active assignment metadata, appends an `incident.assigned` timeline event, and records an `incident.assigned` audit event. Assignment roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`; agency admins can assign only to their own agency.

`GET /api/v1/incidents?assignedToMe=true` filters the incident feed to active assignments for the request actor agency. `assignedAgencyId=<agency-id>` is available for authority readers that need to inspect a specific agency queue.

`GET /api/v1/incidents/audit?limit=50` returns latest incident workflow audit events with before/after snapshots for `system_admin`, `agency_admin`, and `nadmo_officer`.

`GET /api/v1/incidents/{id}` returns a single incident with the same privacy split as the list: reporter identity and contact are hidden unless the caller holds a reporter-contact role and the reporter granted contact permission.

## Media Upload Flow

`POST /api/v1/media/uploads` creates private media metadata and returns an absolute `uploadUrl` built from `NADAA_INCIDENT_PUBLIC_BASE_URL` pointing at the content endpoint. When the caller carries a verified bearer token, its subject becomes the uploader of record. Clients then `PUT /api/v1/media/{id}/content` with the raw bytes (requires the initiating uploader's verified token or an MFA-verified authority token); bytes are stored under `NADAA_INCIDENT_MEDIA_STORAGE_PATH`, capped at the content type's size limit, and the record moves to `uploaded`. `GET /api/v1/media/{id}/content` returns the stored bytes to the same uploader or authority readers. Incident reports can reference returned media IDs. Known media IDs are marked `linked` when the incident is created. `GET /api/v1/media` lists media metadata and requires an authority reader role.

Supported content types and limits:

- Images: `image/jpeg`, `image/png`, `image/webp`, up to 10 MB.
- Video: `video/mp4`, `video/quicktime`, up to 100 MB.
- Audio: `audio/mpeg`, `audio/mp4`, `audio/wav`, up to 25 MB.

## Duplicate Candidate Baseline

When a report is created, the service compares it against existing same-hazard reports. Candidates are scored with:

- Location distance within 750 meters.
- Report time within 3 hours.
- Description token similarity.

`location` is optional on incident reports (channels without GPS, such as USSD, omit it) and round-trips as `null`; when supplied, `0,0` is rejected with `invalid_location`. Reports without a location are never distance-scored, so locationless pairs produce no duplicate candidates.

The top candidates are stored on incident records and returned by `POST /api/v1/incidents` and `GET /api/v1/incidents`. This is a dispatcher review aid only: the service does not automatically merge, hide, delete, or downgrade any report.

`GET /api/v1/incidents/{id}/duplicates` returns the selected incident with full records for open duplicate candidates so the dashboard can compare reports side by side.

`POST /api/v1/incidents/{id}/merge` accepts `duplicateIncidentIds` plus a required `note`. The primary incident remains the operational record, merged duplicates are closed with `mergedIntoId`, `mergedBy`, `mergedAt`, and `mergeReason`, and the service appends timeline plus audit events for traceability.

## Abuse And False Report Handling

When a report is created, the service adds transparent `abuseSignals` for suspicious content such as public links, promotional wording, repeated language, very low detail, or a burst of reports from the same retained reporter identity. Signals roll up into `abuseScore`; scores at or above `0.55` set `abuseReviewRequired`.

These signals are dispatcher review aids only. The service never rejects or hides a report solely because of the suspicion score, and life-threatening reports remain in the live queue with `priorityReview`.

## Run

```bash
go run .
```

The service listens on `:8084` by default. Override with `NADAA_INCIDENT_ADDR`.

Other environment variables:

- `NADAA_AUTH_TOKEN_SECRET` — shared HMAC secret used to verify NADAA bearer tokens; when empty, authority requests are rejected unless mock actors are enabled.
- `NADAA_INTERNAL_SERVICE_TOKEN` — shared service-to-service token; when set, a matching `X-NADAA-Service-Token` header grants read-only incident access (no reporter contact disclosure); when unset the header is ignored.
- `NADAA_AUTH_ALLOW_MOCK_ACTORS`, default `false` — when `true`, honor legacy `X-NADAA-Actor-*` headers (local development and smoke tests only; startup fails unless `NADAA_ENV=development`).
- `NADAA_TRUST_PROXY_HEADERS`, default `false` — when `true`, rate limiting uses `X-Forwarded-For`/`X-Real-Ip` (only set behind a trusted reverse proxy).
- `NADAA_ENV` — when `development`, localhost origins are allowed alongside the `NADAA_ALLOWED_ORIGINS` allowlist.
- `NADAA_INCIDENT_MEDIA_STORAGE_PATH`, default `./uploads/media` — directory where uploaded media bytes are stored.
- `NADAA_INCIDENT_PUBLIC_BASE_URL`, default `http://localhost:8084` — externally reachable base URL used to build absolute media `uploadUrl` values.

## Rate Limiting

The starter service uses an in-memory per-client limiter keyed by client IP (proxy headers only when trusted). Upload initiation and volunteer registration share the limiter with incident creation. JSON request bodies are capped at 1 MiB.

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
pnpm smoke:incident-abuse
pnpm smoke:incident-assignment
pnpm smoke:incident-merge
```

## Notes

The current implementation uses an in-memory store to lock in the public API contract, validation behavior, duplicate candidate baseline, duplicate merge contract, incident workflow contract, abuse review contract, agency assignment contract, and timeline event shape. Media bytes are written to disk under `NADAA_INCIDENT_MEDIA_STORAGE_PATH`; PostGIS persistence lands in later stories. Dispatch workflows (assignment, timelines, response status) are served by this service and the dispatcher apps; there is no separate dispatch-service.
