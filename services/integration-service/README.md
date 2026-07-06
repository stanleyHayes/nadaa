# Integration Service

The integration service owns contract discovery and mock adapters for agency APIs, weather and hydrology imports, hospital capacity feeds, road closure data, utility outages, and outbound incident/alert sync.

Current NADAA-080 endpoints:

- `GET /healthz`
- `GET /api/v1/integrations/contracts`
- `GET /api/v1/integrations/mock/weather-hydrology/observations`
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
