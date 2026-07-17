# Integration Service

The integration service owns contract discovery and mock adapters for agency APIs, weather and hydrology imports, hospital capacity feeds, road closure data, utility outages, and outbound incident/alert sync.

Current endpoints:

- `GET /healthz`
- `GET /api/v1/integrations/contracts`
- `GET /api/v1/integrations/mock/weather-hydrology/observations`
- `POST /api/v1/integrations/weather-hydrology/import-jobs`
- `GET /api/v1/integrations/weather-hydrology/import-jobs`
- `POST /api/v1/integrations/weather-hydrology/import-jobs/{id}/retry`
- `GET /api/v1/integrations/weather-hydrology/observations`
- `POST /api/v1/integrations/mock/sync-events`
- `GET /api/v1/integrations/mock/sync-events`
- `POST /api/v1/integrations/road-closures/imports`
- `GET /api/v1/integrations/road-closures/imports`

## Road Closure Imports

`POST /api/v1/integrations/road-closures/imports` validates an inbound road closure record (including the WKT `LINESTRING` geometry) locally, then forwards it to the `road-closure-service` adapter endpoint. The import is also recorded locally for observability. Forwarding uses `NADAA_ROAD_CLOSURE_SERVICE_URL` (default `http://localhost:8095`) and passes through the caller's `Authorization` bearer token so the downstream service verifies and attributes the real actor; downstream client errors (400/401/403/404) are surfaced to the caller with the downstream error code. When `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` (local dev and smoke tests only), legacy `X-NADAA-Actor-*` authority headers are forwarded instead.

## Contracts

`GET /api/v1/integrations/contracts` supports optional `domain`, `direction`, and `partner` filters.

Contracts include:

- Data owner and source of truth.
- Expected cadence and freshness window.
- Authentication mode and required headers.
- Expected payload fields and PII classification.
- Retry, dead-letter, and manual fallback behavior.

## Mock Adapters

The starter mock adapters return fixture rainfall and water-level observations, and accept incident or alert sync events for development. They do not call official partner systems and do not require secrets.

## Weather And Hydrology Imports

`POST /api/v1/integrations/weather-hydrology/import-jobs` imports fixture rainfall and water-level observations into an in-memory store aligned to the `weather_observations` table. Each imported record keeps source, station, timestamp, location, validity window, normalized rainfall/water-level fields, source record, import job ID, and metadata.

Import jobs are logged with status, trigger, attempts, retryability, imported/failed counts, error, and next retry time. Failed jobs can be retried with `POST /api/v1/integrations/weather-hydrology/import-jobs/{id}/retry`.

Set `NADAA_IMPORT_SCHEDULER_ENABLED=true` to enable the scheduled importer hook. Override the default 15-minute interval with `NADAA_IMPORT_SCHEDULER_INTERVAL`, for example `5m`.

## Run

```bash
go run .
```

The service listens on `:8088` by default. Override with `NADAA_INTEGRATION_ADDR`.

## Test

```bash
go test ./...
```

## Related Stories

- NADAA-080
- NADAA-081
- NADAA-121
- NADAA-131
