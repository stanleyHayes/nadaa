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

Run incident service:

```bash
cd services/incident-service
go run .
```

Run checks:

```bash
pnpm validate:docs
pnpm typecheck
pnpm build
pnpm go:test
```

## Environment Matrix

| Environment | Purpose | Data | Notifications | Access |
| --- | --- | --- | --- | --- |
| Local | Developer implementation | Fixtures only | Mock providers | Developer machine |
| Staging | QA, UAT, demos | Seeded non-production data | Sandbox/test providers | Project team and approved stakeholders |
| Beta | Limited controlled rollout | Production-like approved data | Limited real providers | Selected users/agencies |
| Production | National service | Approved production data | Real providers | Public and agencies |

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
- `SMS_PROVIDER`
- `SMS_API_KEY`
- `WHATSAPP_PROVIDER`
- `WHATSAPP_API_KEY`
- `PUSH_PROVIDER`
- `PUSH_API_KEY`
- `EMAIL_PROVIDER`
- `EMAIL_API_KEY`
- `ML_SERVICE_URL`

Do not commit real values. Use `.env.example` when a service needs a template.

## CI/CD Expectations

CI should run:

- TypeScript type checks.
- App builds.
- Go tests.
- Future API tests.
- Future linting/format checks.
- Future container builds.

Staging deployment should require:

- Passing CI.
- Review approval.
- Environment variable validation.
- Database migration plan.
- Smoke test pass.

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
