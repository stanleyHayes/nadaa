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
pnpm dev:marketing
pnpm dev:citizen
pnpm dev:citizen-mobile
pnpm dev:authority
pnpm dev:dispatcher
pnpm dev:admin
pnpm dev:agency
```

`pnpm dev:citizen-mobile` starts the Expo/React Native citizen mobile foundation from `apps/citizen-mobile`. Copy `apps/citizen-mobile/.env.example` when pointing the native app at local services. The repository-level CI validates the mobile contracts with `pnpm --filter @nadaa/citizen-mobile typecheck`; native simulator/device validation requires the local Expo toolchain.

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

Notification-service runtime logs are structured for staging visibility:

- `INFO` logs cover startup, alert/feed reads, delivery attempts, voice asset generation/review/delivery, SMS/USSD/WhatsApp receipt, menu and conversation decisions, report creation, incident handoff success, and stored access logs/reports.
- `WARN` logs cover invalid requests, unsupported SMS/USSD/WhatsApp selections, unapproved voice delivery attempts, provider errors, alert-service fixture fallback, skipped delivery, and queued reports when incident-service is not configured or unavailable.
- `ERROR` logs cover service startup failure, missing provider wiring, response encoding failures, and incident handoff request/response construction errors.
- Logs use `phoneRef`, provider IDs, command names, path depth, counts, statuses, and report/log IDs. Do not add raw phone numbers, full message bodies, media captions, or full citizen report details to runtime logs.

Run guide service:

```bash
cd services/guide-service
go run .
```

Run shelter service:

```bash
cd services/shelter-service
go run .
```

Run road closure service:

```bash
cd services/road-closure-service
go run .
```

Shelter-service also owns the Phase 2 hospital capacity tracker and relief distribution endpoints. Runtime logs use `INFO` for shelter/recovery, relief point, and hospital capacity reads, creates, updates, stock-history reads, and fixture imports; `WARN` for invalid coordinates, failed validation, unauthorized authority context, missing records, and stale/fixture workflow problems; and `ERROR` for response encoding failures. Do not log private patient details, hospital staff personal data, or sensitive beneficiary details.

Run integration service:

```bash
cd services/integration-service
go run .
```

Run ML service:

```bash
cd services/ml-service
go run .
```

Set `NADAA_ML_MODEL_DIR` only when the baseline model artifacts are outside the default repository paths. Set `NADAA_ML_API_URL=http://127.0.0.1:8094/api/v1` on risk-service when risk responses should include ML decision support.

Set `VITE_ML_API_URL=http://127.0.0.1:8094/api/v1` for dispatcher-web when the ML review panel should use the live ML service instead of fixture predictions.

Set `VITE_SHELTER_API_URL=http://127.0.0.1:8093/api/v1` for dispatcher-web and agency-web when the hospital capacity, shelter, and relief point panels should use live shelter-service data instead of fixture cards.

Set `NADAA_ROAD_CLOSURE_SERVICE_URL=http://127.0.0.1:8095` on integration-service when road closure adapter imports should reach a non-default `road-closure-service` endpoint.

Set `NADAA_IMPORT_SCHEDULER_ENABLED=true` only when the weather/hydrology fixture importer should run on a timer. Override the default interval with `NADAA_IMPORT_SCHEDULER_INTERVAL`, for example `15m`.

Set `NADAA_ALLOWED_ORIGINS` to a comma-separated allowlist when testing browser apps against local or deployed APIs. Leave it empty only for local fixture development where wildcard CORS is acceptable.

Run checks:

```bash
pnpm validate:docs
pnpm security:scan
pnpm features:flood
pnpm validate:features
pnpm ml:flood:train
pnpm validate:ml
pnpm typecheck
pnpm build
pnpm go:test
pnpm smoke:web
pnpm smoke:marketing
pnpm smoke:citizen-mobile
pnpm smoke:citizen-guides
pnpm smoke:alert
pnpm smoke:alert-geofence
pnpm smoke:notification
pnpm smoke:sms-ussd
pnpm smoke:whatsapp
pnpm smoke:incident-abuse
pnpm smoke:incident-assignment
pnpm smoke:incident-merge
pnpm smoke:incident-workflow
pnpm smoke:ml
pnpm smoke:ml-review
pnpm smoke:risk
pnpm smoke:guide
pnpm smoke:shelter
pnpm smoke:relief
pnpm smoke:road-closure
pnpm smoke:integration
```

Run staging smoke checks against configured URLs:

```bash
STAGING_MARKETING_URL=http://127.0.0.1:5172 \
STAGING_CITIZEN_URL=http://127.0.0.1:5173 \
STAGING_AUTHORITY_URL=http://127.0.0.1:5174 \
STAGING_DISPATCHER_URL=http://127.0.0.1:5175 \
STAGING_ADMIN_URL=http://127.0.0.1:5176 \
STAGING_AGENCY_URL=http://127.0.0.1:5177 \
STAGING_NOTIFICATION_SERVICE_URL=http://127.0.0.1:8090 \
STAGING_ML_SERVICE_URL=http://127.0.0.1:8094 \
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
- `NADAA_ALLOWED_ORIGINS`
- `NADAA_ALERT_ADDR`
- `NADAA_NOTIFICATION_ADDR`
- `NADAA_INCIDENT_SERVICE_URL`
- `NADAA_SHELTER_ADDR`
- `NADAA_ALERT_SERVICE_URL`
- `SMS_PROVIDER`
- `SMS_API_KEY`
- `NADAA_SMS_ENABLED`
- `WHATSAPP_PROVIDER`
- `WHATSAPP_API_KEY`
- `PUSH_PROVIDER`
- `PUSH_API_KEY`
- `NADAA_PUSH_ENABLED`
- `VOICE_PROVIDER`
- `VOICE_API_KEY`
- `NADAA_VOICE_ENABLED`
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
- Security scan for API hardening, container runtime users, env-file guardrails, and security review documentation.
- Workspace lint checks.
- TypeScript type checks.
- Workspace tests.
- App and package builds.
- Go tests for `alert-service`, `auth-service`, `incident-service`, `guide-service`, `integration-service`, `ml-service`, `notification-service`, `risk-service`, and `shelter-service`.
- Docker build validation for marketing web, citizen web, authority dashboard, dispatcher web, admin web, agency web, alert service, auth service, incident service, guide service, integration service, ML service, notification service, risk service, and shelter service images.

`Staging Smoke` runs manually against the GitHub `staging` environment:

- `STAGING_MARKETING_URL` must serve the NADAA Marketing website.
- `STAGING_CITIZEN_URL` must serve the NADAA Citizen app.
- `STAGING_AUTHORITY_URL` must serve the NADAA Authority Dashboard.
- `STAGING_DISPATCHER_URL` must serve the NADAA Dispatch Command app.
- `STAGING_ADMIN_URL` must serve the NADAA Admin Console app.
- `STAGING_AGENCY_URL` must serve the NADAA Agency Operations app.
- Optional service URLs are checked at `/healthz` when configured.

SMS/USSD/WhatsApp inclusive access uses notification-service sandbox webhooks in development. Voice alerts use the notification-service reviewed asset workflow and `NADAA_VOICE_ENABLED=true` by default for sandbox `mock_voice` delivery logs. Set `NADAA_INCIDENT_SERVICE_URL=https://incident.<env>/api/v1` when inbound SMS/USSD/WhatsApp reports should be handed off directly to incident-service; leave it empty to keep reports queued in notification-service for sandbox/manual handoff.

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
- Completed [Release Readiness](release-readiness.md), [Beta Monitoring](beta-monitoring.md), and [Hypercare](hypercare.md) checklists.

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
- Volunteer task assignment, update, and escalation volume.
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

Use [UAT Plan](uat.md), [Release Readiness](release-readiness.md), [User Guide And Training](user-guide.md), [Beta Monitoring](beta-monitoring.md), and [Hypercare](hypercare.md) as the Sprint 7 readiness pack. 8. Hypercare.
