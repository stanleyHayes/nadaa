# Auth Service

The auth service owns citizen authentication, agency users, MFA, role-based access control, and sessions.

Current endpoints:

- `GET /healthz`
- `POST /api/v1/auth/citizens/register`
- `POST /api/v1/auth/citizens/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/agency-users`
- `POST /api/v1/auth/agency-users/{id}/mfa/setup`
- `POST /api/v1/auth/agency-users/{id}/mfa/verify`
- `POST /api/v1/auth/agency/login`

## Agency Users And MFA

Agency-user creation requires a bearer token for a `system_admin` or `agency_admin` user that has completed MFA. New agency users receive a development temporary password, must set up and verify mock MFA, and cannot log in until MFA is enabled.

The in-memory service seeds the default fixture agency:

- `00000000-0000-0000-0000-000000000101` - NADMO Accra Metro

For local development, you can seed a bootstrap system admin by setting:

```bash
NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL=admin@nadaa.local
NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD=change-me-locally
NADAA_AUTH_BOOTSTRAP_ADMIN_PHONE=+233200000001
NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE=123456
```

Bootstrap credentials must be provided by environment variables only. Do not commit real passwords or MFA codes.

## Development OTP

Set `NADAA_AUTH_MOCK_OTP=123456` to make the service use a fixed OTP in development.

Set `NADAA_AUTH_EXPOSE_DEV_OTP=true` only for local development or automated tests if the registration response should include `devOtp` or MFA setup should include `devCode`.

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
