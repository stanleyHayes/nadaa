# Shelter Service

The shelter service owns shelter capacity, nearby shelter lookup, recovery support locations, hospital capacity, and relief distribution point tracking for the NADAA platform.

## Current endpoints

- `GET /healthz`
- `GET /api/v1/shelters`
- `GET /api/v1/shelters/nearby?lat=5.6037&lng=-0.1870`
- `GET /api/v1/recovery-support/nearby?lat=5.6037&lng=-0.1870`
- `PATCH /api/v1/shelters/{id}/occupancy`
- `GET /api/v1/relief-points?status=open&type=food&limit=12`
- `GET /api/v1/relief-points/nearby?lat=5.5600&lng=-0.2000`
- `POST /api/v1/relief-points`
- `PATCH /api/v1/relief-points/{id}`
- `GET /api/v1/relief-points/{id}/stock-history`
- `GET /api/v1/hospitals/capacity?lat=5.5600&lng=-0.2000`
- `PATCH /api/v1/hospitals/{id}/capacity`
- `POST /api/v1/hospitals/capacity/imports/fixture`

Shelter occupancy, relief point create/update, and hospital capacity updates require authority headers:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

Allowed update roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`.

## Run

```bash
go run .
```

The service listens on `:8093` by default. Override with `NADAA_SHELTER_ADDR`.

## Test

```bash
go test ./...
```

## Smoke

```bash
pnpm smoke:shelter
pnpm smoke:relief
```

`smoke:relief` expects the service on port `8093` by default; override with `RELIEF_API_URL`.

## Notes

The current implementation uses an in-memory fixture store so citizen and authority surfaces can integrate against a stable API contract. PostGIS persistence and district-owned shelter/relief/hospital feeds can replace the store later without changing the endpoint shapes.
