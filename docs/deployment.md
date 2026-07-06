# Deployment

This document defines local, staging, and production deployment expectations for NADAA.

## Local Development

Start local data services:

```bash
docker compose -f infra/docker/docker-compose.yml up -d
```

Install dependencies:

```bash
pnpm install
```

Run apps:

```bash
pnpm dev:citizen
pnpm dev:authority
```

Run risk service:

```bash
cd services/risk-service
go run .
```

Run auth service with a development OTP:

```bash
cd services/auth-service
NADAA_AUTH_MOCK_OTP=123456 NADAA_AUTH_EXPOSE_DEV_OTP=true go run .
```

Optional local-only in-memory agency admin bootstrap:

```bash
NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL=admin@nadaa.local
NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD=change-me-locally
NADAA_AUTH_BOOTSTRAP_ADMIN_PHONE=+233200000001
NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE=123456
```

Run incident service:

```bash
cd services/incident-service
go run .
```

Run alert service:

```bash
cd services/alert-service
go run .
```

Run notification service:

```bash
cd services/notification-service
go run .
```

Run guide service:

```bash
cd services/guide-service
go run .
```

Run integration service:

```bash
cd services/integration-service
go run .
```

Run checks:

```bash
pnpm validate:docs
pnpm typecheck
pnpm build
pnpm go:test
pnpm smoke:web
pnpm smoke:citizen-guides
pnpm smoke:alert
pnpm smoke:alert-geofence
pnpm smoke:notification
pnpm smoke:incident-abuse
pnpm smoke:incident-assignment
pnpm smoke:incident-merge
pnpm smoke:incident-workflow
pnpm smoke:risk
pnpm smoke:guide
pnpm smoke:integration
```

Run staging smoke checks against configured URLs:

```bash
STAGING_CITIZEN_URL=http://127.0.0.1:5173 \
STAGING_AUTHORITY_URL=http://127.0.0.1:5174 \
STAGING_NOTIFICATION_SERVICE_URL=http://127.0.0.1:8090 \
pnpm smoke:staging
```

## Environment Matrix

| Environment | Purpose                    | Data                          | Notifications          | Access                                 |
| ----------- | -------------------------- | ----------------------------- | ---------------------- | -------------------------------------- |
| Local       | Developer implementation   | Fixtures only                 | Mock providers         | Developer machine                      |
| Staging     | QA, UAT, demos             | Seeded non-production data    | Sandbox/test providers | Project team and approved stakeholders |
| Beta        | Limited controlled rollout | Production-like approved data | Limited real providers | Selected users/agencies                |
| Production  | National service           | Approved production data      | Real providers         | Public and agencies                    |

## Required Environment Variables

Initial expected groups:

- `DATABASE_URL`
- `REDIS_URL`
- `OBJECT_STORAGE_ENDPOINT`
- `OBJECT_STORAGE_BUCKET`
- `OBJECT_STORAGE_ACCESS_KEY`
- `OBJECT_STORAGE_SECRET_KEY`
- `JWT_SECRET`
- `MFA_ISSUER`
- `NADAA_AUTH_TOKEN_SECRET`
- `NADAA_ALERT_ADDR`
- `NADAA_NOTIFICATION_ADDR`
- `NADAA_ALERT_SERVICE_URL`
- `SMS_PROVIDER`
- `SMS_API_KEY`
- `NADAA_SMS_ENABLED`
- `WHATSAPP_PROVIDER`
- `WHATSAPP_API_KEY`
- `PUSH_PROVIDER`
- `PUSH_API_KEY`
- `NADAA_PUSH_ENABLED`
- `EMAIL_PROVIDER`
- `EMAIL_API_KEY`
- `ML_SERVICE_URL`
- `NADAA_INTEGRATION_ADDR`

Do not commit real values. Use `.env.example` when a service needs a template. Use `infra/staging/staging.env.example` as the staging environment checklist.

Optional in-memory auth-service bootstrap variables for local/staging fixture environments:

- `NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL`
- `NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD`
- `NADAA_AUTH_BOOTSTRAP_ADMIN_PHONE`
- `NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE`

## CI/CD Expectations

The first-pass GitHub Actions workflows live in `.github/workflows/`.

`CI` runs on pull requests and pushes to `main`:

- Project dashboard and database asset validation.
- Workspace lint checks.
- TypeScript type checks.
- Workspace tests.
- App and package builds.
- Go tests for `alert-service`, `auth-service`, `incident-service`, `guide-service`, `integration-service`, `notification-service`, and `risk-service`.
- Docker build validation for citizen web, authority dashboard, alert service, auth service, incident service, guide service, integration service, notification service, and risk service images.

`Staging Smoke` runs manually against the GitHub `staging` environment:

- `STAGING_CITIZEN_URL` must serve the NADAA Citizen app.
- `STAGING_AUTHORITY_URL` must serve the NADAA Authority Dashboard.
- Optional service URLs are checked at `/healthz` when configured.

Staging deployment should require:

- Passing CI.
- Review approval.
- Environment variable validation.
- Database migration plan.
- Smoke test pass.
- Image registry push and deploy credentials configured in the `staging` environment.

Production deployment should require:

- UAT sign-off.
- Security review for affected surfaces.
- Release notes.
- Rollback plan.
- Monitoring and support coverage.

## Local Infrastructure

`infra/docker/docker-compose.yml` provides:

- PostGIS on port `5432`, or `POSTGRES_PORT` when overridden.
- Redis on port `6379`, or `REDIS_PORT` when overridden.
- MinIO object storage on ports `9000` and `9001`, or `MINIO_API_PORT` and `MINIO_CONSOLE_PORT` when overridden.

These credentials are development-only and must not be reused in staging or production.

## Observability Targets

MVP should capture:

- API latency and error rate.
- Incident report creation rate.
- Report-to-verification time.
- Verification-to-assignment time.
- Alert approval and delivery events.
- Notification provider failures.
- Geospatial query performance.
- ML prediction latency and model version.

## Release Gates

1. Internal QA.
2. Staging deployment.
3. Smoke testing.
4. Security review for sensitive changes.
5. UAT.
6. Beta rollout.
7. Production release.
8. Hypercare.
