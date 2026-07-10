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

Authority endpoints require MFA-completed authority headers:

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

## Notes

The current implementation uses an in-memory fixture store plus local disk file storage so authority and public surfaces can integrate against a stable API contract. PostGIS persistence and object-store file backends can replace the store later without changing endpoint shapes.
