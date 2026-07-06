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
