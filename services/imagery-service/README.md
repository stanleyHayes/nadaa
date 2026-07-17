# Imagery Service

The imagery service manages drone, satellite, and other aerial imagery ingestion, metadata, file storage, footprint GeoJSON, lifecycle/retention, and authority downloads for the NADAA platform.

## Current endpoints

- `GET /healthz`
- `POST /api/v1/imagery` — authority multipart upload
- `GET /api/v1/imagery?source=&status=&relatedIncidentId=&relatedRiskZoneId=&q=` — authority list
- `GET /api/v1/imagery/{id}` — authority detail
- `GET /api/v1/imagery/{id}/download` — authority file download
- `DELETE /api/v1/imagery/{id}` — authority delete
- `POST /api/v1/imagery/{id}/expire` — authority mark expired
- `POST /api/v1/imagery/lifecycle/run` — authority run retention lifecycle
- `GET /api/v1/imagery/geojson` — public active footprint FeatureCollection

Authority endpoints require a valid `Authorization: Bearer nadaa.<payload>.<sig>` token issued by auth-service (verified with `NADAA_AUTH_TOKEN_SECRET`); the token's role must be an allowed authority role with MFA completed. For local development and smoke tests, setting `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` also honors the legacy headers:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

Allowed authority roles are `system_admin`, `nadmo_officer`, `district_officer`, `dispatcher`, `police`, `fire`, `ambulance`, `rescue`, and `analyst`.

## Upload fields

`POST /api/v1/imagery` accepts `multipart/form-data` with a required `file` part and the following fields:

- `source` — `drone`, `satellite`, or `other` (required)
- `captureTime` — ISO 8601 timestamp (required)
- `geometry` — GeoJSON Polygon JSON string (required)
- `coverageAreaKm2` — non-negative number (required)
- `resolutionMeters` — non-negative number (required)
- `license` — optional license string
- `relatedIncidentId` — optional incident reference
- `relatedRiskZoneId` — optional risk-zone reference
- `mlWorkflowId` — optional ML workflow reference

Uploaded files must be images (`image/*`) and may not exceed 20 MB. Files are stored as `{storagePath}/{id}-{originalFilename}`.

## Run

```bash
go run ./cmd/server
```

The service listens on `:8099` by default. Override with `PORT`.

## Test

```bash
go test ./...
```

## Environment variables

- `PORT` — HTTP listen address (default `:8099`)
- `NADAA_ALLOWED_ORIGINS` — comma-separated CORS allowlist; `*` or empty allows all origins
- `IMAGERY_STORAGE_PATH` — directory for uploaded files (default `./uploads`)
- `DEFAULT_RETENTION_DAYS` — retention period in days (default `90`)
- `NADAA_AUTH_TOKEN_SECRET` — HMAC secret used to verify auth-service bearer tokens; when empty (and mock actors off) authority endpoints return 401
- `NADAA_AUTH_ALLOW_MOCK_ACTORS` — `true` honors legacy `X-NADAA-Actor-*` headers (local dev/smoke tests only)
- `NADAA_IMAGERY_PUBLIC_BASE_URL` — externally reachable base URL (e.g. `https://imagery.example.com`) used for geojson `downloadUrl`s; falls back to the request scheme/host when unset
- `NADAA_ENV` — `development` allows localhost/127.0.0.1 CORS origins alongside the allowlist

Expired records are expired automatically by an hourly retention lifecycle tick (in addition to the manual `POST /api/v1/imagery/lifecycle/run` endpoint).

## Notes

The current implementation uses an in-memory fixture store plus local disk file storage so authority and public surfaces can integrate against a stable API contract. PostGIS persistence and object-store file backends can replace the store later without changing endpoint shapes.
