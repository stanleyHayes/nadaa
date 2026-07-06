# Architecture

NADAA is the Ghana National Disaster Alert and Response Platform. It is designed as an AI-assisted, human-approved emergency preparedness, warning, response, and recovery platform.

## Architecture Goals

- Keep citizen workflows fast, mobile-first, and resilient under poor connectivity.
- Keep authority workflows auditable, role-protected, and operationally clear.
- Treat flood risk as the first hazard while preserving an extensible hazard model.
- Keep public warning decisions human-approved.
- Make agency integrations replaceable through explicit adapters and contracts.
- Keep geospatial data, incident data, media, alert delivery, and ML predictions traceable by source and timestamp.

## Current Foundation

- `apps/citizen-web` - React/Vite citizen PWA starter.
- `apps/authority-dashboard` - React/Vite authority dashboard starter.
- `packages/brand` - brand constants, slogan, palette, and feature pillars.
- `packages/shared-types` - shared TypeScript domain contracts.
- `services/auth-service` - Go citizen and agency authentication starter with mock OTP/MFA, signed bearer tokens, RBAC, and audit events.
- `services/alert-service` - Go alert draft, submission, approval, rejection, emergency override, and audit-event starter.
- `services/incident-service` - Go incident intake starter with validation, rate limiting, media references, and priority review flagging.
- `services/guide-service` - Go emergency guidance starter with hazard/stage/language lookup, offline availability metadata, and seed-aligned content fixtures.
- `services/integration-service` - Go integration contract and mock-adapter starter for agency, weather, hydrology, incident, alert, hospital, road, utility, and shelter data exchange.
- `services/notification-service` - Go citizen alert feed, mock push/SMS provider abstraction, and delivery log starter.
- `services/risk-service` - first Go service with `GET /healthz` and `GET /api/v1/risk`.
- `infra/docker/docker-compose.yml` - local PostGIS, Redis, and MinIO.
- `database/migrations/001_core_geospatial_schema.sql` - core PostGIS schema and indexes.
- `database/seeds/001_ghana_mvp_seed.sql` - development seed data for Ghana MVP fixtures.

## Target Runtime Topology

```text
Citizen Web / Mobile / USSD / WhatsApp
        |
        v
API Gateway / Edge Routing
        |
        +--> Auth Service
        +--> Risk Service -----> PostGIS + ML Service
        +--> Incident Service -> PostGIS + Object Storage
        +--> Alert Service ----> Notification Service
        +--> Dispatch Service -> Agency Users + Timelines
        +--> Guide Service
        +--> Integration Service -> GMet, Hydro, NADMO, Police, Fire, Ambulance, Hospitals

Authority Dashboard
        |
        +--> Incident, Alert, Dispatch, Risk, ML, Integration APIs
```

## Service Boundaries

### Auth Service

Owns citizen accounts, agency users, admins, sessions, MFA, role-based access, and agency membership.

Primary dependencies:

- User database.
- MFA provider or app-based OTP.
- Audit logging.

### Risk Service

Owns area risk lookup, hazard risk scoring, nearby shelters, nearby emergency facilities, recommended actions, and risk API aggregation.

Primary dependencies:

- PostGIS risk zones.
- Shelter/facility records.
- Weather/hydrology observations.
- ML prediction service.

### Incident Service

Owns citizen reports, agency reports, media metadata, duplicate candidates, incident verification, audited status transitions, starter agency assignments, timeline events, severity, and incident record lifecycle.

Primary dependencies:

- PostGIS incidents.
- Object storage for media.
- Dispatch service for future assignment orchestration and responder visibility.
- Audit logging.

### Alert Service

Owns alert drafts, approval workflow, emergency override, targeting metadata, publication status, expiry, audit, and alert history.

Primary dependencies:

- PostGIS target geometry.
- Notification service.
- Auth/RBAC.
- Audit logging.

### Dispatch Service

Owns agency assignment, responder status, response timeline, internal notes, escalation, and closure handoff.

Primary dependencies:

- Agency model.
- Incident service.
- Notification service for responder updates.

### Guide Service

Owns emergency preparedness, response, and recovery guide content, including hazard type, stage, language, offline availability, and stable ordering for citizen offline caching.

Primary dependencies:

- Emergency guide records in PostGIS.
- Future CMS/editor workflow.
- Citizen PWA cache.

### Notification Service

Owns citizen alert feed aggregation, provider abstraction for push and SMS, delivery logs, and the extension path for email, WhatsApp, voice alerts, retries, and future cell broadcast.

Primary dependencies:

- Alert service approved/published alert feed.
- Provider credentials through environment secrets.
- Notification delivery logs.

### Integration Service

Owns external data ingestion and outbound sync with NADMO, GMet, Ghana Hydrological Authority, police, fire, ambulance, hospitals, utilities, and district assemblies.

Primary dependencies:

- Adapter registry.
- Import job logs.
- Source-specific credentials.
- Contract matrix in `docs/integrations.md`.

### ML Service

Owns model serving, flood prediction, prediction metadata, explainability, model versioning, and later simulation/triage/computer vision.

Primary dependencies:

- Feature pipeline outputs.
- MLflow or equivalent model registry.
- Risk service.
- Prediction logs.

## Data Stores

- PostgreSQL + PostGIS: authoritative relational and geospatial store for users, agencies, incidents, alerts, risk zones, shelters, observations, and predictions.
- Redis: caching, rate limits, queues, and short-lived workflow state.
- Object storage: incident media, voice alert assets, imagery, and generated exports.
- Future analytical store: model features, reporting aggregates, and open data exports if PostGIS becomes too operationally loaded.

## MVP Scope Boundary

MVP includes:

- Citizen web/PWA foundation.
- Citizen phone login foundation.
- Risk checker with flood-first scoring.
- Incident reporting with GPS/media metadata.
- Authority incident map, verification, assignment, and timeline.
- Alert creation, approval, and geofenced targeting.
- In-app alert feed and notification abstraction.
- Emergency guidance and shelter map/list.
- Baseline flood risk model and human review view.
- Core docs, QA matrix, CI/CD, and staging readiness.

Phase 2 includes inclusive channels, field coordination, hospital capacity, relief logistics, road closures, route planning, missing persons, damage exports, and remote sensing ingestion.

Phase 3 includes advanced simulation, AI triage, computer vision, predictive resource positioning, school preparedness, campaigns, open data, cell broadcast, and national-scale hardening.

## Data Ownership Assumptions

- Citizen profile and report data is owned by the platform operator and governed by privacy policy and emergency reporting rules.
- Agency user and assignment data is owned by the participating agency or platform operator depending on deployment agreement.
- Weather, hydrology, police, fire, ambulance, hospital, utility, and district assembly data remains owned by the originating institution.
- Imported external data must keep source, license/usage constraints, freshness timestamp, and contact point.
- Open-data exports must be anonymized, aggregated, and approved before publication.

## Integration Assumptions

- Official agency APIs may not be available during MVP.
- MVP should use fixture/mock adapters for weather, hydrology, shelter, facility, and agency data.
- Adapters must isolate source-specific authentication, payload shape, rate limits, and retry behavior.
- Integration failures must not block manual incident response or manual alert approval.

## Safety Principles

- Public alerts require human approval.
- ML output is decision support and must include confidence, model version, and explanation factors.
- Authority actions are role-protected and audited.
- Citizen location and identity data are minimized and protected.
- Suspicious report scoring can inform human review but must not silently suppress urgent life-safety reports.
