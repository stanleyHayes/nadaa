# Road Closure Service

Manages road closure records as a geospatial context layer for dispatcher maps, agency operations, citizen guidance, and future route planning.

## Local run

```bash
cd services/road-closure-service
go run .
```

The service listens on `:8095` by default. Override with `NADAA_ROAD_CLOSURE_ADDR`.

## Endpoints

- `GET /healthz`
- `GET /api/v1/road-closures?status=&lat=&lng=&radius=&bbox=&limit=&includeExpired=`
- `POST /api/v1/road-closures`
- `PATCH /api/v1/road-closures/{id}`
- `POST /api/v1/road-closures/imports/adapter`

Authority endpoints require `X-NADAA-Actor-ID`, `X-NADAA-Actor-Role`, `X-NADAA-Agency-ID`, `X-NADAA-MFA-Completed`, and `X-NADAA-Request-ID`. Allowed roles for create/update/import: `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher`.

## Geometry

Manual create/update accepts a GeoJSON `LineString` object. The adapter import endpoint accepts a WKT `LINESTRING(...)` string.

## Status values

`active`, `scheduled`, `lifted`, `cancelled`.

## Severity values

`low`, `moderate`, `high`, `severe`, `emergency`.

## Tests

```bash
go test ./...
```
