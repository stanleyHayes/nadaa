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
- `GET /api/v1/audit/logs`

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

## Audit Logs

The service records in-memory audit events for citizen registration/login, agency login, agency-user creation, MFA setup/verification, RBAC denial, and audit-log viewing. `GET /api/v1/audit/logs` is restricted to `system_admin` users with MFA completed.

Audit records include actor, action, target, timestamp, request id, IP address, user agent, and sanitized before/after context where appropriate. They must not include OTPs, MFA codes, temporary passwords, bearer tokens, or provider secrets.

## Development OTP

Set `NADAA_AUTH_MOCK_OTP=123456` to make the service use a fixed OTP in development.

Set `NADAA_AUTH_EXPOSE_DEV_OTP=true` only for local development or automated tests if the registration response should include `devOtp` or MFA setup should include `devCode`.

## Mock actor headers (dev only)

Set `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` to let agency endpoints accept the shared `X-NADAA-Actor-ID` / `X-NADAA-Agency-ID` / `X-NADAA-Actor-Role` / `X-NADAA-MFA-Completed` headers instead of a verified session token, matching the mock-auth scheme the other services use. This trusts client-supplied role headers and therefore bypasses real authentication, so it is **off by default and must never be enabled in production** — leave it unset there and require signed tokens.

## Token signing secret (required)

`NADAA_AUTH_TOKEN_SECRET` signs every citizen and agency bearer token (HMAC-SHA256). It is **required**: the service refuses to start if it is empty, equals the placeholder `dev-secret-change-me`, or is shorter than 32 bytes — because a build shipping with the placeholder can have its tokens forged and its RBAC bypassed. Set a strong random value in every real environment. For local development only, set `NADAA_AUTH_ALLOW_INSECURE_SECRET=true` to bypass this check (the dev backend script already does).

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

The current implementation uses an in-memory store so the API contract and tests can unblock downstream incident reporting and authority-dashboard work. The PostGIS user and audit schemas already exist; persistence can be wired in a later hardening slice without changing the API contract.
