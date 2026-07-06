# Auth Service

The auth service owns citizen authentication, future agency users, MFA, role-based access control, and sessions.

Current NADAA-010 endpoints:

- `GET /healthz`
- `POST /api/v1/auth/citizens/register`
- `POST /api/v1/auth/citizens/login`
- `GET /api/v1/auth/me`

## Development OTP

Set `NADAA_AUTH_MOCK_OTP=123456` to make the service use a fixed OTP in development.

Set `NADAA_AUTH_EXPOSE_DEV_OTP=true` only for local development or automated tests if the registration response should include `devOtp`.

## Run

```bash
go run .
```

The service listens on `:8080` by default. Override with `NADAA_AUTH_ADDR`.

## Test

```bash
go test ./...
```

## Notes

The current implementation uses an in-memory store so the API contract and tests can unblock downstream incident reporting work. The PostGIS user schema already exists; persistence can be wired in a later hardening slice without changing the citizen auth API contract.

