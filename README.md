# NADAA

NADAA is the Ghana National Disaster Alert and Response Platform.

Slogan: **Be Aware. Be Prepared. Be Safe.**

The platform helps citizens, NADMO, district assemblies, dispatchers, and response agencies prepare for, report, monitor, respond to, and recover from disasters. The first implementation phase focuses on flood risk, citizen reporting, authority incident command, approved alerts, emergency guidance, and shelter visibility.

## Repository Layout

```text
apps/
  citizen-web/            Citizen PWA for alerts, risk checks, reports, guides, and shelters
  dispatcher-web/         Dispatcher command console for incident triage, workflow, assignments, alerts, and maps
  admin-web/              Governance console for agencies, users, roles, MFA, audit, data sources, and alert rules
  authority-dashboard/    Compatibility shell for the original authority dashboard while role-specific apps are split out
services/
  auth-service/
  incident-service/
  alert-service/
  risk-service/
  guide-service/
  dispatch-service/
  notification-service/
  integration-service/
  ml-service/
packages/
  brand/                  NADAA colors, slogan, feature pillars, and brand constants
  shared-types/           Shared TypeScript domain contracts
  config/                 Shared tool configuration
infra/
  docker/
  kubernetes/
  terraform/
docs/
```

## Getting Started

Install dependencies:

```bash
pnpm install
```

Run the citizen web app:

```bash
pnpm dev:citizen
```

The citizen app uses `VITE_RISK_API_URL`, `VITE_INCIDENT_API_URL`, `VITE_NOTIFICATION_API_URL`, and `VITE_GUIDE_API_URL`, defaulting to local service URLs for risk, incident, notification, and guide APIs.
Copy `apps/citizen-web/.env.example` if you need different service URLs.

Run the authority dashboard:

```bash
pnpm dev:authority
```

The authority dashboard uses `VITE_INCIDENT_API_URL`, defaulting to `http://localhost:8084/api/v1`, for incident map and status workflow actions.
Copy `apps/authority-dashboard/.env.example` if you need a different incident service URL.

Run the dispatcher command console:

```bash
pnpm dev:dispatcher
```

The dispatcher web app uses `VITE_INCIDENT_API_URL` and `VITE_ALERT_API_URL`, defaulting to local incident and alert service URLs. It runs on port `5175` and is the target app for dispatcher incident command workflows.
Copy `apps/dispatcher-web/.env.example` if you need different service URLs.

Run the admin governance console:

```bash
pnpm dev:admin
```

The admin web app uses `VITE_AUTH_API_URL`, `VITE_INTEGRATION_API_URL`, and `VITE_ALERT_API_URL`, defaulting to local auth, integration, and alert service URLs. It runs on port `5176` and keeps system administration separate from dispatcher operations.
Copy `apps/admin-web/.env.example` if you need different service URLs.

Run the web apps:

```bash
pnpm dev
```

Run the Go risk service:

```bash
cd services/risk-service
go run .
```

Run the Go auth service:

```bash
cd services/auth-service
NADAA_AUTH_MOCK_OTP=123456 NADAA_AUTH_EXPOSE_DEV_OTP=true go run .
```

For local agency-user testing, seed an in-memory bootstrap admin with environment variables such as `NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL`, `NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD`, and `NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE`. Do not commit real bootstrap credentials.

Run the Go incident service:

```bash
cd services/incident-service
go run .
```

With the incident service running on `:8084`, verify the status workflow with `pnpm smoke:incident-workflow`, abuse/false-report review with `pnpm smoke:incident-abuse`, agency assignment with `pnpm smoke:incident-assignment`, and duplicate merge review with `pnpm smoke:incident-merge`.

Run the Go alert service:

```bash
cd services/alert-service
go run .
```

The authority dashboard uses `VITE_ALERT_API_URL`, defaulting to `http://localhost:8089/api/v1`, for alert draft, geofenced targeting, and approval workflows. With alert-service running on `:8089`, verify alert workflow with `pnpm smoke:alert` and geofenced targeting with `pnpm smoke:alert-geofence`.

Run the Go notification service:

```bash
cd services/notification-service
go run .
```

The citizen app uses `VITE_NOTIFICATION_API_URL`, defaulting to `http://localhost:8090/api/v1`, for current/expired alert feed data. The notification service uses alert-service when available and fixture fallback in development. With notification-service running on `:8090`, verify feed delivery logs with `pnpm smoke:notification`.

Run the Go guide service:

```bash
cd services/guide-service
go run .
```

The guide service listens on `:8086` by default and exposes `GET /api/v1/guides` for emergency preparedness and response content. With the citizen app on `:5173` and guide-service on `:8086`, verify offline-first guide integration with `pnpm smoke:citizen-guides`.

Run the Go integration service:

```bash
cd services/integration-service
go run .
```

The integration service listens on `:8088` by default and exposes integration contracts, mock weather/hydrology adapters, weather/hydrology import jobs, imported observation status, and mock incident/alert sync adapters.

## Project Coordination

Use `agent_plan.md` as the living project board. Before starting work, agents should claim a row in the Active Work Board, update the Master Story Tracker, and record handoff notes when finished or blocked.

## Documentation

- [Product Scope](docs/product.md)
- [Architecture](docs/architecture.md)
- [API](docs/api.md)
- [Security](docs/security.md)
- [ML](docs/ml.md)
- [Integrations](docs/integrations.md)
- [Deployment](docs/deployment.md)
- [QA Strategy](docs/qa.md)
- [Database](database/README.md)
- [Project Dashboard Contract](docs/project-dashboard/README.md)

## Source Documents

- `spec.md`
- `AI_Native_Software_Engineering_Operations_Manual.docx`
- `AI_Development_Workflow_Training_Manual.docx`
