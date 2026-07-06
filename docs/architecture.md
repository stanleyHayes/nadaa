# Architecture

NADAA is organized as a monorepo with citizen-facing apps, authority apps, backend services, shared packages, infrastructure, and project documentation.

## Current Foundation

- `apps/citizen-web` - React/Vite citizen PWA starter.
- `apps/authority-dashboard` - React/Vite authority dashboard starter.
- `packages/brand` - brand constants, slogan, palette, and feature pillars.
- `packages/shared-types` - shared TypeScript domain contracts.
- `services/risk-service` - first Go service with `GET /healthz` and `GET /api/v1/risk`.
- `infra/docker/docker-compose.yml` - local PostGIS, Redis, and MinIO.

## Target Service Map

- Auth Service: citizens, agency users, admins, RBAC, and MFA.
- Risk Service: geospatial risk lookup, shelters, safety recommendations, and ML prediction integration.
- Incident Service: citizen reports, media metadata, deduplication, severity, and dispatch handoff.
- Alert Service: authority alert creation, approval, geofenced targeting, and audit trail.
- Dispatch Service: agency assignment, responder status, and incident timelines.
- Notification Service: push, SMS, WhatsApp, voice, email, and future cell broadcast.
- Integration Service: weather, hydrology, agency systems, hospital capacity, and road closures.
- ML Service: flood risk prediction, explainability, model serving, triage, and future simulation.

## Safety Principles

- Public alerts require human approval.
- ML output is decision support and must include confidence, model version, and explanation factors.
- Authority actions are role-protected and audited.
- Citizen location and identity data are minimized and protected.

