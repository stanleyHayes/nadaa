# Donation Service

The donation service coordinates donors, aid catalog items, aid requests, and pledges for the NADAA platform.

## Current endpoints

Public endpoints:

- `GET /healthz`
- `GET /api/v1/aid-catalog`
- `GET /api/v1/aid-requests?status=&category=&region=&priority=`
- `POST /api/v1/donors`
- `POST /api/v1/aid-requests/{id}/pledges`

Authority endpoints (require `X-NADAA-Actor-ID`, `X-NADAA-Actor-Role`, `X-NADAA-Agency-ID`, `X-NADAA-MFA-Completed: true`, `X-NADAA-Request-ID`):

- `GET /api/v1/donors?type=&q=`
- `POST /api/v1/donors`
- `GET /api/v1/donors/{id}`
- `PATCH /api/v1/donors/{id}`
- `POST /api/v1/aid-requests`
- `GET /api/v1/aid-requests/{id}`
- `PATCH /api/v1/aid-requests/{id}`
- `GET /api/v1/aid-requests/{id}/pledges`
- `GET /api/v1/pledges?status=`
- `PATCH /api/v1/pledges/{id}`
- `POST /api/v1/aid-requests/{id}/allocate`

Allowed authority roles are `system_admin`, `agency_admin`, `agency_viewer`, `nadmo_officer`, `district_officer`, `dispatcher`, and `ngo`.

## Run

```bash
go run .
```

The service listens on `:8100` by default. Override with `NADAA_DONATION_ADDR`.

## Test

```bash
go test ./...
```

## Environment variables

| Variable                | Default | Description                                                 |
| ----------------------- | ------- | ----------------------------------------------------------- |
| `PORT`                  | `:8100` | HTTP listen address.                                        |
| `NADAA_ALLOWED_ORIGINS` | `*`     | Comma-separated allowed CORS origins. Use `*` to allow all. |

## Notes

The current implementation uses an in-memory fixture store so citizen and authority surfaces can integrate against a stable API contract. Persistence can replace the store later without changing the endpoint shapes.
