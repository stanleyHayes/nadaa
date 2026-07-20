# NADAA

NADAA is the Ghana National Disaster Alert and Response Platform.

Slogan: **Be Aware. Be Prepared. Be Safe.**

The platform helps citizens, NADMO, district assemblies, dispatchers, and response agencies prepare for, report, monitor, respond to, and recover from disasters. The implementation covers flood risk, citizen reporting, authority incident command, approved alerts, emergency guidance, shelter visibility, hospital capacity, relief distribution points, and recovery logistics.

## Repository Layout

```text
apps/
  marketing-web/          Public website for platform overview, services, platform lanes, benefits, and contact
  citizen-web/            Citizen PWA for alerts, risk checks, reports, guides, and shelters
  dispatcher-web/         Dispatcher command console for incident triage, workflow, assignments, alerts, and maps
  agency-web/             Agency operations portal for assigned incidents, capacity context, and relief logistics
  admin-web/              Governance console for agencies, users, roles, MFA, audit, data sources, and alert rules
  citizen-mobile/         Expo citizen mobile foundation for alerts, reports, guides, shelters, and community tasks
  dispatcher-mobile/      Expo dispatcher mobile triage foundation for shift-friendly incident actions
  authority-dashboard/    Compatibility shell for the original authority dashboard while role-specific apps are split out
services/
  auth-service/
  incident-service/
  alert-service/
  risk-service/
  guide-service/
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

Run the public marketing website:

```bash
pnpm dev:marketing
```

The marketing site runs on port `5172` and summarizes NADAA's about story, features, platform lanes, services, benefits, research context, and contact paths. It uses the real NADAA logo, brand sheet, and Outfit typography.

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
It also uses `VITE_ML_API_URL` for reviewed flood predictions and `VITE_SHELTER_API_URL` for hospital capacity filters, defaulting to the local ML and shelter service URLs.
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

Run the ML service:

```bash
cd services/ml-service
go run .
```

Set `NADAA_ML_API_URL=http://127.0.0.1:8094/api/v1` on risk-service to include baseline flood predictions in area risk responses. Set `VITE_ML_API_URL=http://127.0.0.1:8094/api/v1` for dispatcher-web to use live predictions in the ML review panel. With ML service running on `:8094`, verify model serving with `pnpm smoke:ml`; with alert-service also running on `:8089`, verify reviewed draft traceability with `pnpm smoke:ml-review`.

Run the Go auth service:

```bash
cd services/auth-service
NADAA_AUTH_MOCK_OTP=123456 NADAA_AUTH_EXPOSE_DEV_OTP=true go run .
```

For local agency-user testing, seed an in-memory bootstrap admin with environment variables such as `NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL`, `NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD`, and `NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_SECRET` (a base32 TOTP seed). Do not commit real bootstrap credentials.

Run the Go incident service:

```bash
cd services/incident-service
go run .
```

With the incident service running on `:8084`, verify the status workflow with `pnpm smoke:incident-workflow`, abuse/false-report review with `pnpm smoke:incident-abuse`, agency assignment with `pnpm smoke:incident-assignment`, duplicate merge review with `pnpm smoke:incident-merge`, and community volunteer coordination with `pnpm smoke:community-volunteers`.

Set `NADAA_ALLOWED_ORIGINS` to a comma-separated list of approved browser origins when testing APIs outside local fixture development. Run `pnpm security:scan` before UAT or after changes to API CORS/header handling, Dockerfiles, env templates, or security docs.

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
The notification service also exposes Phase 2 SMS/USSD and WhatsApp sandbox webhooks for inclusive access plus reviewed multilingual voice-alert delivery. Set `NADAA_INCIDENT_SERVICE_URL=http://127.0.0.1:8084/api/v1` when inbound SMS/USSD/WhatsApp reports should be submitted to incident-service; otherwise they remain queued in notification-service. Verify the SMS/USSD flow with `pnpm smoke:sms-ussd`, the WhatsApp chatbot with `pnpm smoke:whatsapp`, and voice assets/delivery logs with `pnpm smoke:voice-alerts`.

The Phase 2 citizen mobile foundation lives in `apps/citizen-mobile`. It is an Expo/React Native app shell with NADAA brand assets, current alerts, risk lookup, incident report drafts, community volunteer assignments, offline guides, shelter/recovery support, session handling, permission copy, and push registration scaffolding. Run `pnpm --filter @nadaa/citizen-mobile typecheck` and `pnpm smoke:citizen-mobile`; use `pnpm dev:citizen-mobile` when running the Expo toolchain locally.

Run the Go guide service:

```bash
cd services/guide-service
go run .
```

The guide service listens on `:8086` by default and exposes `GET /api/v1/guides` for emergency preparedness and response content. With the citizen app on `:5173` and guide-service on `:8086`, verify offline-first guide integration with `pnpm smoke:citizen-guides`.

Run the Go shelter service:

```bash
cd services/shelter-service
go run .
```

The shelter service listens on `:8093` by default and exposes shelter/recovery lookup, relief point management, stock history, hospital capacity, and donation/aid coordination endpoints. With shelter-service running on `:8093`, verify nearby shelters, protected occupancy updates, relief point list/nearby/create/update/stock-history behavior, aid request/pledge coordination, hospital capacity filters, manual capacity updates, and fixture imports with `pnpm smoke:shelter`, `pnpm smoke:relief`, and `pnpm smoke:aid`.

Run the Go missing person service:

```bash
cd services/missing-person-service
go run ./cmd/server
```

The missing person service listens on `:8101` by default and exposes private intake, authority review, public approved search, closure/reunification, and audit endpoints. Citizen and authority web apps use `VITE_MISSING_PERSON_API_URL`, defaulting to `http://localhost:8101/api/v1`. Verify the privacy workflow with `pnpm smoke:missing-person`.

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
- [UAT Plan](docs/uat.md)
- [Release Readiness](docs/release-readiness.md)
- [User Guide And Training](docs/user-guide.md)
- [Beta Monitoring](docs/beta-monitoring.md)
- [Hypercare](docs/hypercare.md)
- [Database](database/README.md)
- [Project Dashboard Contract](docs/project-dashboard/README.md)

## Source Documents

- `spec.md`
- `AI_Native_Software_Engineering_Operations_Manual.docx`
- `AI_Development_Workflow_Training_Manual.docx`
