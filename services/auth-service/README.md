# Auth Service

The auth service owns citizen authentication, agency users, MFA, role-based access control, and sessions.

Current endpoints:

- `GET /healthz`
- `POST /api/v1/auth/citizens/register`
- `POST /api/v1/auth/citizens/login/otp`
- `POST /api/v1/auth/citizens/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/agency-users`
- `GET /api/v1/auth/agency-users`
- `POST /api/v1/auth/agency-users/{id}/mfa/setup`
- `POST /api/v1/auth/agency-users/{id}/mfa/verify`
- `POST /api/v1/auth/agency/login`
- `POST /api/v1/auth/agency/password`
- `GET /api/v1/auth/agencies`
- `GET /api/v1/audit/logs`
- `POST /api/v1/audit/logs`

## Citizen Login

`POST /api/v1/auth/citizens/register` issues the first OTP challenge and keeps returning `409 phone_already_registered` for an existing phone. Returning citizens request a fresh challenge with `POST /api/v1/auth/citizens/login/otp` (`{"phone": "+233..."}` → `{phone, challengeId, otpDelivery}`), then exchange it at `POST /api/v1/auth/citizens/login` (`{phone, otp}`) for a 24h bearer token.

## Agency Login, MFA, And Agencies

Agency users log in at `POST /api/v1/auth/agency/login` (`{email, password, mfaCode}`) for a 12h bearer token whose claims include `agencyId`, `district`, and `mfa`. Before MFA is enabled the login returns `403 mfa_setup_required`; complete `POST /api/v1/auth/agency-users/{id}/mfa/setup` then `.../mfa/verify` first. `GET /api/v1/auth/agencies` returns the agency directory (`{agencies: [...]}`) to `system_admin` tokens with MFA completed.

`GET /api/v1/auth/agency-users` returns the agency user directory (`{users: [{id, name, email, role, agencyId, mfaEnabled, lockedUntil?, createdAt}]}`) to `system_admin` and `agency_admin` tokens with MFA completed; agency admins see only their own agency. The directory never exposes password hashes or MFA secrets.

`POST /api/v1/auth/agency/password` (`{currentPassword, newPassword}`) lets any verified agency user with MFA completed change their password. The current password is verified through the same brute-force lockout machinery as login (`429 too_many_attempts`, `401 invalid_credentials`); the new password must be at least 12 characters (`400 weak_password`). Success returns `{ok: true}` and records an audit event.

## Brute-Force Protection

Credential verification paths (citizen OTP verify, agency login, agency MFA verify, agency password change) track consecutive failures per phone/email/user. After 5 failures the account key locks for 15 minutes and requests fail with `429 too_many_attempts`; a successful verify resets the counter.

## Password Hashing

Agency passwords are stored as salted PBKDF2-SHA256 hashes (`pbkdf2$<iterations>$<salt_b64>$<hash_b64>`, 210000 iterations). Verification still accepts legacy unsalted SHA-256 digests so hashes written by older builds keep working.

## Agency Users And MFA

Agency-user creation requires a bearer token for a `system_admin` or `agency_admin` user that has completed MFA. New agency users receive a development temporary password, must enroll and verify TOTP MFA, and cannot log in until MFA is enabled.

Agency MFA is real TOTP (RFC 6238: HMAC-SHA1, 30-second step, 6-digit codes, ±1 step window for clock skew) — there are no static or reusable MFA codes in any environment. `POST /api/v1/auth/agency-users/{id}/mfa/setup` stores a pending base32 secret and returns `{secret, otpauthUrl}` so the user can enroll an authenticator app; `.../mfa/verify` validates a code from that authenticator against the pending secret and enables MFA; login validates the current TOTP window.

The in-memory service seeds the default fixture agency:

- `00000000-0000-0000-0000-000000000101` - NADMO Accra Metro

For local development, you can seed a bootstrap system admin by setting:

```bash
NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL=admin@nadaa.local
NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD=change-me-locally
NADAA_AUTH_BOOTSTRAP_ADMIN_PHONE=+233200000001
NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_SECRET=JBSWY3DPEHPK3PXP
```

Bootstrap credentials must be provided by environment variables only. Do not commit real passwords or MFA secrets. The password must be at least 12 characters. The MFA secret must be a base32-encoded TOTP seed; when it is omitted the service generates a random one and logs its otpauth URL once at startup so the operator can enroll an authenticator — there is no constant fallback. The retired `NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE` variable is ignored with a startup warning. If the bootstrap credentials are configured but seeding fails, the service refuses to start.

## Audit Logs

The service records in-memory audit events for citizen registration/login, agency login, agency-user creation, MFA setup/verification, password change, RBAC denial, and audit-log viewing. `GET /api/v1/audit/logs` is restricted to `system_admin` users with MFA completed.

`POST /api/v1/audit/logs` ingests audit events forwarded by other services. It requires `X-NADAA-Service-Token` matching `NADAA_INTERNAL_SERVICE_TOKEN` (`401 invalid_service_token` otherwise, and the endpoint stays closed when the token is unset). The body is `{eventType, actorId?, actorRole?, resourceType?, resourceId?, summary?, metadata?}`; `eventType` must be non-empty (`400 invalid_event`). Success persists the event and returns `201 {id}`.

Audit records include actor, action, target, timestamp, request id, IP address, user agent, and sanitized before/after context where appropriate. They must not include OTPs, MFA codes or secrets, temporary passwords, bearer tokens, or provider secrets.

## Development OTP

Set `NADAA_AUTH_MOCK_OTP=123456` to make the service use a fixed OTP in development.

Set `NADAA_AUTH_EXPOSE_DEV_OTP=true` only for local development or automated tests if the registration response should include `devOtp` or MFA setup should include `devCode`. For MFA the `devCode` is simply the current TOTP code of the enrollment secret; it is emitted only when `NADAA_ENV=development` is also set.

Both settings — and `NADAA_AUTH_ALLOW_MOCK_ACTORS` — are rejected at startup unless `NADAA_ENV=development`, so they can never leak into a deployed environment.

## Mock actor headers (dev only)

Set `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` (with `NADAA_ENV=development`) to let agency endpoints accept the shared `X-NADAA-Actor-ID` / `X-NADAA-Agency-ID` / `X-NADAA-Actor-Role` / `X-NADAA-MFA-Completed` headers instead of a verified session token, matching the mock-auth scheme the other services use. This trusts client-supplied role headers and therefore bypasses real authentication, so it is **off by default and must never be enabled in production** — leave it unset there and require signed tokens.

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
