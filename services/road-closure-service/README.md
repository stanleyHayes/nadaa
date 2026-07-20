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

Authority endpoints require an `Authorization: Bearer nadaa.<payload>.<sig>` token issued by auth-service, verified with `NADAA_AUTH_TOKEN_SECRET`; the actor id, role, agency, district, and MFA flag are taken from the verified claims. For local development and smoke tests, setting `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` makes the service honor legacy `X-NADAA-Actor-ID`, `X-NADAA-Actor-Role`, `X-NADAA-Agency-ID`, `X-NADAA-Actor-District`, and `X-NADAA-MFA-Completed` headers instead; the service refuses to start with that setting unless `NADAA_ENV=development`. Allowed roles for create/update/import: `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher`.

## Geometry

Manual create/update accepts a GeoJSON `LineString` object. The adapter import endpoint accepts a WKT `LINESTRING(...)` string.

## Status values

`active`, `scheduled`, `lifted`, `cancelled`. A `scheduled` closure already
inside its validity window (`validFrom <= now <= validTo`, inclusive) is
treated as active: it matches `status=active` queries and is served with
`status: "active"` so map and route-facing consumers do not miss it. The
stored record is unchanged — it returns to `scheduled` queries once its
window has passed.

## Severity values

`low`, `moderate`, `high`, `severe`, `emergency`.

## Tests

```bash
go test ./...
```
