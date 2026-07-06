# Shelter Service

The shelter service owns shelter capacity, nearby shelter lookup, and recovery support locations for the NADAA MVP.

Current NADAA-062 endpoints:

- `GET /healthz`
- `GET /api/v1/shelters`
- `GET /api/v1/shelters/nearby?lat=5.6037&lng=-0.1870`
- `GET /api/v1/recovery-support/nearby?lat=5.6037&lng=-0.1870`
- `PATCH /api/v1/shelters/{id}/occupancy`

Occupancy updates require authority headers:

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

## Notes

The current implementation uses an in-memory fixture store so citizen and authority surfaces can integrate against a stable API contract. PostGIS persistence and district-owned shelter feeds can replace the store later without changing the endpoint shapes.
