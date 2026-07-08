# NADAA Architecture Guide

This guide is for engineers, operators, and technical maintainers of the National Disaster Alert & Response Platform (NADAA). It explains the architecture, technology choices, component responsibilities, data flows, deployment topology, security model, and development workflow.

---

## Table of contents

1. [Overview and goals](#overview-and-goals)
2. [High-level architecture](#high-level-architecture)
3. [Technology stack](#technology-stack)
4. [Frontend applications](#frontend-applications)
5. [Backend services](#backend-services)
6. [Shared packages](#shared-packages)
7. [Data layer](#data-layer)
8. [Integration architecture](#integration-architecture)
9. [ML pipeline](#ml-pipeline)
10. [Security architecture](#security-architecture)
11. [API conventions](#api-conventions)
12. [Deployment topology](#deployment-topology)
13. [Development workflow](#development-workflow)
14. [Operational runbooks](#operational-runbooks)
15. [Future extensions](#future-extensions)

---

## Overview and goals

NADAA is an AI-assisted, human-approved emergency preparedness, warning, response, and recovery platform for Ghana. The architecture is designed around these goals:

- **Citizen-first and mobile-first.** Public workflows must work on low-bandwidth devices and degrade gracefully when connectivity is poor.
- **Human authority over public warnings.** ML predictions and automated signals are decision support only; a public alert cannot be published without human approval.
- **Role-based access and auditability.** Every sensitive authority action is attributable through signed tokens, RBAC, MFA, and append-only audit logs.
- **Hazard-extensible.** Flood risk is the first hazard, but the data model and service boundaries support fire, road crash, storm, disease outbreak, and others.
- **Replaceable integrations.** Partner data sources and delivery providers are isolated behind contract-first adapters.
- **Geospatial by default.** Locations, target areas, shelters, and risk cells are stored and queried with PostGIS.

---

## High-level architecture

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Client layer                                    │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐        │
│  │ Citizen Web  │ │ Dispatcher   │ │ Agency Web   │ │ Admin Web    │        │
│  │ (PWA)        │ │ Web          │ │              │ │              │        │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘        │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐                         │
│  │ Marketing    │ │ Authority    │ │ Mobile apps  │                         │
│  │ Web          │ │ Dashboard    │ │ (Expo/RN)    │                         │
│  └──────────────┘ └──────────────┘ └──────────────┘                         │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                              API gateway / edge
                                       │
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Service layer                                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │ auth-       │ │ incident-   │ │ alert-      │ │ risk-       │            │
│  │ service     │ │ service     │ │ service     │ │ service     │            │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │ guide-      │ │ shelter-    │ │ notification│ │ integration │            │
│  │ service     │ │ service     │ │ service     │ │ service     │            │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘            │
│  ┌─────────────┐                                                             │
│  │ ml-service  │                                                             │
│  └─────────────┘                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
┌─────────────────────────────────────────────────────────────────────────────┐
│                             Data layer                                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │ PostgreSQL  │ │ Redis       │ │ MinIO       │ │ Object      │            │
│  │ + PostGIS   │ │ cache/queue │ │ object      │ │ storage     │            │
│  │             │ │             │ │ storage     │ │ adapters    │            │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘            │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Technology stack

### Frontend

- **Framework:** React with Vite.
- **Language:** TypeScript.
- **UI library:** MUI (Material UI) with Emotion for styling.
- **State management:** React hooks and context; server state fetched via API clients.
- **Maps:** Leaflet / Mapbox GL for command and citizen maps.
- **Mobile:** Expo / React Native for citizen-mobile and dispatcher-mobile.

### Backend

- **Language:** Go 1.25.
- **Standard library only.** No external Go dependencies in services.
- **HTTP server:** `net/http` with graceful shutdown and sensible timeouts.
- **Authentication:** Custom signed bearer tokens with HMAC-SHA256.

### Data and infrastructure

- **Primary database:** PostgreSQL with PostGIS extension.
- **Cache / queue:** Redis.
- **Object storage:** MinIO (local/S3-compatible) for media and model artifacts.
- **Containerization:** Docker and Docker Compose locally; Kubernetes planned for production.
- **CI/CD:** GitHub Actions.
- **Package manager:** pnpm workspaces.

### ML

- **Language:** Python for training pipelines.
- **Model artifact format:** JSON metadata and sample predictions.
- **Serving:** Go `ml-service` loads artifacts in-memory and exposes a REST API.

---

## Frontend applications

All web apps live under `apps/` and share brand, types, and configuration through workspace packages.

| App                   | Path                       | Purpose                                                                                                                     |
| --------------------- | -------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `marketing-web`       | `apps/marketing-web`       | Public marketing and stakeholder conversion site.                                                                           |
| `citizen-web`         | `apps/citizen-web`         | Citizen PWA for risk checks, reporting, alerts, guides, shelters, and recovery support.                                     |
| `dispatcher-web`      | `apps/dispatcher-web`      | Incident command console: triage, verification, assignment, duplicate review, abuse review, timelines, maps, and ML review. |
| `agency-web`          | `apps/agency-web`          | Agency-scoped portal for assigned incidents, responder updates, shelter/hospital capacity, and relief management.           |
| `admin-web`           | `apps/admin-web`           | Governance console for agencies, users, roles, MFA, audit logs, data sources, and alert rules.                              |
| `authority-dashboard` | `apps/authority-dashboard` | Compatibility shell retained while workflows move to dispatcher, agency, and admin apps.                                    |
| `citizen-mobile`      | `apps/citizen-mobile`      | Expo/React Native citizen app with push alerts, offline guides, GPS/media reporting, and shelter lookup.                    |
| `dispatcher-mobile`   | `apps/dispatcher-mobile`   | Expo/React Native triage app for shift use.                                                                                 |

### Frontend modularity

Each web app follows the same layout:

- `src/App.tsx` is a thin entrypoint.
- `src/app/` holds app-wide configuration, theme, routing, and session helpers.
- `src/features/<feature>/` holds domain-specific components, data hooks, types, and utilities.
- Large features are split into `data.ts`, `types.ts`, `utils.ts`, and focused `*.tsx` components.

---

## Backend services

All services live under `services/`. Each service is a Go module with the following layout:

```text
services/<service>/
├── cmd/server/main.go          # entry point with graceful shutdown
├── internal/config/config.go   # environment-based configuration
├── internal/models/models.go   # request/response/domain types
├── internal/store/store.go     # persistence interface and in-memory implementation
├── internal/utils/utils.go     # service-specific helpers
└── internal/handlers/          # HTTP handlers, routes, middleware, tests
```

| Service                | Responsibility                                                                                                                    | Key endpoints                                                                                                                               |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `auth-service`         | Citizen and agency authentication, mock OTP/MFA, signed bearer tokens, RBAC, audit events.                                        | `POST /api/v1/auth/citizens/register`, `POST /api/v1/auth/citizens/login`, `POST /api/v1/auth/agency/login`, `GET /api/v1/audit/logs`       |
| `incident-service`     | Incident intake, validation, rate limiting, media references, verification, assignment, timeline, duplicate review, abuse review. | `POST /api/v1/incidents`, `POST /api/v1/incidents/{id}/verify`, `POST /api/v1/incidents/{id}/assign`, `POST /api/v1/incidents/{id}/merge`   |
| `alert-service`        | Alert draft, submit, approve, reject, emergency override, target geometry, audit events.                                          | `POST /api/v1/alerts`, `POST /api/v1/alerts/{id}/submit`, `POST /api/v1/alerts/{id}/approve`, `POST /api/v1/alerts/{id}/emergency-override` |
| `risk-service`         | Area risk lookup, nearby shelters/facilities, recommended actions, ML decision-support enrichment.                                | `GET /api/v1/risk`                                                                                                                          |
| `ml-service`           | Flood-risk prediction serving from versioned model artifacts.                                                                     | `POST /api/v1/ml/flood/predictions`, `GET /api/v1/ml/prediction-logs`                                                                       |
| `guide-service`        | Emergency guidance catalog with hazard/stage/language filtering and offline availability metadata.                                | `GET /api/v1/guides`                                                                                                                        |
| `shelter-service`      | Shelter and recovery support lookup, protected capacity updates, relief points, aid requests/pledges, hospital capacity.          | `GET /api/v1/shelters`, `POST /api/v1/relief-points`, `POST /api/v1/aid-requests`                                                           |
| `notification-service` | Citizen alert feed, mock and provider-based push/SMS/USSD/WhatsApp/voice delivery, delivery logs.                                 | `GET /api/v1/notifications/alerts`, `POST /api/v1/notifications/delivery-attempts`                                                          |
| `integration-service`  | Contract-first adapters for weather, hydrology, road closures, hospital capacity, utility outages, incident/alert sync.           | `POST /api/v1/integrations/observations`, `POST /api/v1/integrations/road-closures/imports`                                                 |
| `road-closure-service` | Road closure list/create/update and adapter-import.                                                                               | `GET /api/v1/road-closures`, `POST /api/v1/road-closures`                                                                                   |

> **Note:** `dispatch-service` is currently a placeholder directory and is not implemented as a separate service. Dispatch workflows live in `dispatcher-web` and `incident-service`.

### Graceful shutdown

Every service uses an `http.Server` with:

- `ReadTimeout: 10s`
- `WriteTimeout: 30s`
- `IdleTimeout: 120s`
- Signal-based shutdown on `SIGINT` / `SIGTERM` with a 10-second context timeout.

Example pattern in `cmd/server/main.go`:

```go
srv := &http.Server{
    Addr:         cfg.Addr,
    Handler:      handler,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  120 * time.Second,
}

go func() {
    log.Printf("%s listening on %s", serviceName, cfg.Addr)
    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        log.Fatalf("server error: %v", err)
    }
}()

sig := make(chan os.Signal, 1)
signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
<-sig

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    log.Printf("shutdown error: %v", err)
}
```

---

## Shared packages

Workspace packages live under `packages/`:

| Package        | Path                    | Purpose                                                               |
| -------------- | ----------------------- | --------------------------------------------------------------------- |
| `brand`        | `packages/brand`        | Brand constants: slogan, palette, feature pillars, typography tokens. |
| `shared-types` | `packages/shared-types` | Shared TypeScript domain contracts used across apps and services.     |
| `config`       | `packages/config`       | Shared configuration helpers and environment validation.              |

A cross-service shared Go package was intentionally deferred to a separate story. Each Go service currently keeps its own `internal/utils` helpers, which is intentional until the interfaces stabilize.

---

## Data layer

### PostgreSQL + PostGIS

- Core operational data: incidents, alerts, agencies, users, audit logs, shelters, relief points, road closures, aid coordination.
- Geospatial indexing for locations, target geometries, and risk cells.
- Migration files live in `database/migrations/`.
- Seed data lives in `database/seeds/`.

### Redis

- Session/cache layer.
- Queue backend for notification delivery retries.
- Rate-limit counters for public incident intake.

### MinIO / object storage

- Private media: incident photos, videos, voice alert assets.
- Model artifacts for `ml-service`.
- Short-lived signed URLs for authorized media access.

---

## Integration architecture

Integrations are contract-first and adapter-based. See `docs/integrations.md` for the full contract matrix.

Key principles:

- Every import preserves `source`, `observedAt`/`updatedAt`, validity window, and source constraints.
- Every outbound sync includes a `correlationId` for idempotency.
- Adapter failures are retryable and dead-lettered but must not block manual incident response or human-approved alerts.
- Credentials live in environment secrets or a secret manager, never in source control.

Inbound integrations:

- Weather and hydrology observations (`integration-service`).
- Road closure imports (`integration-service` forwards to `road-closure-service`).
- Hospital capacity feeds (`shelter-service`).
- Relief inventory updates (`shelter-service`).
- Utility outage feeds (`integration-service`).

Outbound integrations:

- Incident and alert sync to NADMO and partner agencies.
- SMS/USSD/WhatsApp/voice delivery through notification providers.

---

## ML pipeline

1. **Feature pipeline:** `scripts/build-flood-risk-features.mjs` generates a versioned 44-column feature set from fixture inputs.
2. **Training:** Python scripts train a baseline logistic-regression model and produce model artifacts.
3. **Artifacts:** `data/flood-risk/models/` contains model metadata and sample predictions.
4. **Serving:** `ml-service` loads artifacts in-memory and serves predictions through a REST API.
5. **Decision support:** `risk-service` enriches risk responses with ML predictions, explanation factors, and safety flags (`humanReviewRequired: true`, `autoPublishAllowed: false`).
6. **Human review:** Dispatcher-web ML review panel lets authorized users create alert drafts from predictions. Drafts still require human approval.

---

## Security architecture

### Authentication

- **Citizens:** phone-based registration/login with mock OTP in development; production integrates with an OTP provider.
- **Agency users:** email/password login with mandatory MFA setup before first use.
- **Tokens:** custom signed bearer tokens with HMAC-SHA256. Tokens include user ID, user type, role, agency ID, MFA state, and expiry.

### Authorization

- Role-based access control (RBAC) for all authority endpoints.
- Initial roles: `citizen`, `agency_viewer`, `dispatcher`, `responder`, `nadmo_officer`, `district_officer`, `agency_admin`, `system_admin`.
- Agency admins are scoped to their own agency.
- Separation of duties: non-system approvers cannot approve their own alert draft.

### Audit

- Sensitive actions create append-only audit records.
- Minimum fields: actor user ID, actor agency ID, actor role, action, target type, target ID, request ID, IP address, user agent, before/after snapshots, created at.
- Audit retention: at least 24 months in production unless policy requires longer.

### Runtime HTTP hardening

- CORS controlled by `NADAA_ALLOWED_ORIGINS` (comma-separated allowlist).
- Wildcard CORS is only acceptable for local development and fixture testing.
- Defensive headers on all API responses:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `Referrer-Policy: no-referrer`
  - `Content-Security-Policy`
  - `Strict-Transport-Security`
  - `Cache-Control: no-store`

### Secret handling

- Secrets are loaded from environment variables or a secret manager.
- Never commit API keys, database credentials, provider tokens, or OTP secrets.

---

## API conventions

- Base path: `/api/v1`.
- Request and response bodies are JSON unless media upload explicitly uses signed URLs or multipart form data.
- Error shape:

```json
{
  "error": {
    "code": "invalid_coordinates",
    "message": "lat and lng query parameters are required"
  }
}
```

- List endpoints support `limit`, `cursor`, `sort`, and filter-specific query parameters.
- Authority write endpoints preserve actor, role, agency, MFA-completed, and request-id context through headers or token claims.

---

## Deployment topology

### Local development

- `infra/docker/docker-compose.yml` starts PostgreSQL/PostGIS, Redis, and MinIO.
- Each app and service can be started individually with `pnpm dev:*` or `go run ./cmd/server`.
- `NADAA_ALLOWED_ORIGINS` can be left empty for wildcard CORS during local development.

### Staging

- Seeded non-production data.
- Sandbox/test notification providers.
- Explicit CORS allowlist for staging app origins.
- Manual staging smoke workflow in GitHub Actions.

### Beta

- Production-like approved data.
- Limited real providers.
- Selected users and agencies.

### Production

- National service with real providers.
- Strict CORS allowlist.
- Non-root container runtimes.
- Secret manager for credentials.
- Backup, monitoring, and hypercare runbooks.

---

## Development workflow

### Branch and commit conventions

- Branch: `feature/NADAA-123-short-name`
- Commit: `NADAA-123 implement short name`
- PR: `NADAA-123 Short Name`

### Verification

Run the full verification suite before opening a PR:

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm build
pnpm go:test
pnpm validate:docs
pnpm security:scan
pnpm audit --audit-level high
pnpm exec prettier --check .
git diff --check
```

### Go linting

A top-level `.golangci.yml` configures linters for the Go services. Run it per service:

```bash
cd services/<service>
golangci-lint run ./...
```

Current enabled linters: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`, `revive`, `gocognit`, `misspell`, `gosec`.

### Testing

- Go services: `go test ./...` from each service directory.
- Frontend: `pnpm test` runs type checks across workspace packages and apps.
- Smoke tests: `pnpm smoke:<scenario>` scripts exercise live local or staging endpoints.

---

## Operational runbooks

### Start local infrastructure

```bash
docker compose -f infra/docker/docker-compose.yml up -d
```

### Start a Go service

```bash
cd services/<service>
go run ./cmd/server
```

### Start a web app

```bash
pnpm dev:<app>
```

### Health checks

Each service exposes `GET /healthz` returning:

```json
{ "status": "ok", "service": "<service-name>" }
```

### Graceful shutdown

Services respond to `SIGINT` and `SIGTERM`. In container orchestration, configure a termination grace period of at least 15 seconds.

### Logs

- Go services log startup, request handling errors, and shutdown errors through the standard `log` package.
- `notification-service` and `shelter-service` use structured `INFO`, `WARN`, and `ERROR` logs.
- Logs must not contain raw phone numbers, full message bodies, media captions, passwords, OTPs, tokens, or private media.

### Security scan

```bash
pnpm security:scan
```

### Dependency audit

```bash
pnpm audit --audit-level high
```

---

## Future extensions

Planned capabilities that will extend the architecture:

- Real-time flood simulation and predictive ambulance/fire positioning.
- Computer vision for flood/fire image verification.
- Cell broadcast and telecom integration for mass alerts.
- National open disaster data portal.
- School emergency preparedness module.
- Drone and satellite image ingestion.
- Cross-service shared Go package to reduce duplicated helpers.
- bcrypt/argon2 credential hashing and JWT token migration.

---

## Directory structure reference

```text
nadaa/
├── apps/                      # React/Vite and React Native applications
├── packages/                  # Shared workspace packages
├── services/                  # Go backend services
├── database/                  # Migrations and seeds
├── data/                      # Flood-risk datasets and model artifacts
├── infra/                     # Docker, Kubernetes, Terraform, staging
├── scripts/                   # Smoke tests, validation, and build scripts
├── docs/                      # Product, architecture, API, security, and user guides
├── .github/workflows/         # CI/CD pipelines
├── .golangci.yml              # Go lint configuration
├── pnpm-workspace.yaml        # pnpm workspace definition
└── tsconfig.base.json         # Shared TypeScript configuration
```
