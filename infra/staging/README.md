# Staging

The staging environment is the QA/UAT target for NADAA. It should use seeded non-production data, sandbox notification providers, staging-only secrets, and locked-down access for project stakeholders.

## GitHub Environment Variables

Create a GitHub environment named `staging` and set these environment variables for the `Staging Smoke` workflow:

- `STAGING_CITIZEN_URL`
- `STAGING_AUTHORITY_URL`
- `STAGING_DISPATCHER_URL`
- `STAGING_ADMIN_URL`
- `STAGING_AUTH_SERVICE_URL`
- `STAGING_INCIDENT_SERVICE_URL`
- `STAGING_ALERT_SERVICE_URL`
- `STAGING_GUIDE_SERVICE_URL`
- `STAGING_INTEGRATION_SERVICE_URL`
- `STAGING_NOTIFICATION_SERVICE_URL`
- `STAGING_RISK_SERVICE_URL`

The web URLs are required. Service URLs are optional until those services are deployed publicly, but when set they must expose `/healthz`.

## Environment Template

Use `infra/staging/staging.env.example` as the staging configuration checklist. Do not commit real values. Secrets must live in the hosting provider, GitHub environment secrets, or a secret manager.

## First Deployment Shape

The current CI pipeline builds:

- `nadaa/citizen-web`
- `nadaa/authority-dashboard`
- `nadaa/dispatcher-web`
- `nadaa/admin-web`
- `nadaa/auth-service`
- `nadaa/incident-service`
- `nadaa/alert-service`
- `nadaa/guide-service`
- `nadaa/integration-service`
- `nadaa/notification-service`
- `nadaa/risk-service`

The pipeline validates Docker builds but does not push images yet. Once a registry is selected, add registry login and push steps guarded by the `staging` environment.

## Smoke Test

Run staging smoke locally after exporting staging URLs:

```bash
pnpm smoke:staging
```

Run the same check in GitHub through the `Staging Smoke` workflow.
