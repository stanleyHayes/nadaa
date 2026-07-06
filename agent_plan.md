# Agent Plan: Ghana Emergency Preparedness, Warning & Disaster Response Platform

## Source Documents

- `spec.md` - national disaster preparedness and response product specification.
- `AI_Native_Software_Engineering_Operations_Manual.docx` - SDLC, governance, Jira workflow, QA, release, and AI operating model.
- `AI_Development_Workflow_Training_Manual.docx` - Jira/GitHub/dashboard standards, definition of done, security standards, and implementation roadmap.

## Product Goal

Build an emergency preparedness and disaster intelligence platform for Ghana that helps citizens, NADMO, district assemblies, dispatchers, and response agencies prepare for, report, monitor, respond to, and recover from disasters.

The MVP should focus on:

1. Flood risk checker.
2. Citizen incident reporting.
3. Authority incident command dashboard.
4. Alert creation and delivery workflow.
5. Emergency guidance and shelter visibility.
6. Baseline flood risk scoring model with human-approved alert recommendations.

## Delivery Assumptions

- Sprint length: 2 weeks.
- Initial hazard priority: floods, while keeping the data model extensible for fire, road crash, storm, disease outbreak, tidal wave, and other hazards.
- Initial clients/users: citizens, emergency dispatchers, NADMO/district officers, police/fire/ambulance/rescue teams, and system admins.
- Recommended architecture from the spec is accepted unless revised during solution design:
  - Frontend: React, TypeScript, MUI, Mapbox GL or Leaflet, PWA support.
  - Mobile: React Native later or after web/PWA validation.
  - Backend: Golang, hexagonal architecture, REST + WebSocket.
  - Data: PostgreSQL/PostGIS, MongoDB for flexible media/report metadata if needed, Redis for cache/queues/rate limits.
  - ML: Python/FastAPI, MLflow, Airflow or Prefect later.
  - Infra: Docker first, then Kubernetes/Swarm when deployment maturity requires it.

## Operating Rules For Agents

- Keep changes aligned with `spec.md` and update this plan as scope changes.
- Use Jira-shaped work items: Epic -> Story -> Task/Subtask.
- Every story should include user story, business value, acceptance criteria, technical notes, definition of done, estimate, and dependencies before implementation.
- Keep branch, commit, and PR naming compatible with the manuals:
  - Branch: `feature/NADAA-123-short-name`
  - Commit: `NADAA-123 implement short name`
  - PR: `NADAA-123 Short Name`
- Do not commit secrets, API tokens, database credentials, GitHub secrets, Jira credentials, or client data.
- ML predictions must not automatically send public alerts without human approval.
- Authority users require stronger controls than citizen users: role-based access, MFA, audit logs, and approval workflows for mass alerts.

## Status Workflow

Use this workflow for stories and implementation tasks:

`Backlog -> Ready -> In Progress -> Code Review -> QA Testing -> QA Passed -> Staging -> UAT -> Beta -> Production -> Support -> Closed`

Use `Blocked` when a dependency, data access issue, product decision, or external integration prevents progress.

## Multi-Agent Coordination Board

Use this section as the shared control table when multiple agents are working in the repository.

Status values:

- `Todo` - ready or waiting to be picked up.
- `In Progress` - actively owned by one agent.
- `Blocked` - cannot continue without a decision, dependency, credential, data source, or external integration.
- `Review` - implementation is done and waiting for review, QA, or user acceptance.
- `Done` - accepted and no further work remains for the item.

Coordination rules:

- Before starting work, update the relevant row to `In Progress`, add the owner, branch/worktree if applicable, and note the intended scope.
- Keep one active owner per story unless the row is explicitly split into subtasks.
- If a story touches shared files such as migrations, generated types, shared UI, auth middleware, or deployment config, note that in `Notes` before editing.
- When blocked, write the exact blocker and the next needed decision.
- When complete, move the row to `Done`, add verification notes, and append a short entry to the plan ledger.

### Active Work Board

| ID        | Phase/Sprint   | Work Item                                              | Status | Owner | Branch/PR | Dependencies         | Last Update | Notes                                                                                                                                                                                                                      |
| --------- | -------------- | ------------------------------------------------------ | ------ | ----- | --------- | -------------------- | ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| NADAA-001 | MVP Sprint 0   | Create Repository And Monorepo Foundation              | Done   | Codex | main      | None                 | 2026-07-06  | Monorepo foundation, brand assets, starter apps, risk service, infra, docs, and Git remote are in place.                                                                                                                   |
| NADAA-002 | MVP Sprint 0   | Define Product, API, Security, ML, And Deployment Docs | Done   | Codex | main      | NADAA-001            | 2026-07-06  | Product, architecture, API, security, ML, deployment, data ownership, and integration assumptions documented.                                                                                                              |
| NADAA-003 | MVP Sprint 0   | Create Delivery Dashboard Data Contract                | Done   | Codex | main      | NADAA-001            | 2026-07-06  | Dashboard schema, sample records, README, and validation script added under `docs/project-dashboard/`.                                                                                                                     |
| NADAA-100 | MVP Sprint 0   | Build Test Strategy And QA Matrix                      | Done   | Codex | main      | NADAA-002            | 2026-07-06  | QA strategy, MVP test matrix, release checklist, UAT outline, severity model, and web smoke script added.                                                                                                                  |
| NADAA-020 | MVP Sprint 1   | Set Up PostGIS And Core Geospatial Models              | Done   | Codex | main      | NADAA-001            | 2026-07-06  | Core PostGIS schema, geospatial indexes, seed data, database docs, configurable compose ports, and asset validation added.                                                                                                 |
| NADAA-010 | MVP Sprint 1   | Implement Citizen Authentication                       | Done   | Codex | main      | NADAA-001            | 2026-07-06  | Auth service citizen register/login/profile API, mock OTP flow, signed bearer token, shared auth types, docs, and tests added.                                                                                             |
| NADAA-011 | MVP Sprint 1   | Implement Agency Users, Roles, And MFA                 | Done   | Codex | main      | NADAA-010            | 2026-07-06  | Auth-service now supports agency user creation, authority role catalog, mock MFA setup/verification, agency login, MFA-aware tokens, and RBAC denial tests.                                                                |
| NADAA-012 | MVP Sprint 1   | Implement Audit Logging Foundation                     | Done   | Codex | main      | NADAA-011            | 2026-07-06  | Auth-service now records audit events for citizen/agency auth, admin user creation, MFA setup/verify, RBAC denial, and system-admin audit reads with metadata and sanitized snapshots.                                     |
| NADAA-030 | MVP Sprint 2   | Implement Incident Reporting API                       | Done   | Codex | main      | NADAA-020, NADAA-010 | 2026-07-06  | Incident-service report intake API, validation, anonymous/contact-permission behavior, media references, priority review flag, rate limiting, shared types, docs, and tests added.                                         |
| NADAA-031 | MVP Sprint 2   | Implement Media Upload Flow                            | Done   | Codex | main      | NADAA-030            | 2026-07-06  | Controlled media upload initiation, private metadata, content-type and size validation, incident media linkage, shared types, docs, and tests added.                                                                       |
| NADAA-032 | MVP Sprint 2   | Build Citizen Incident Reporting UI                    | Done   | Codex | main      | NADAA-030, NADAA-031 | 2026-07-06  | Citizen report form now supports GPS/manual coordinates, hazard, urgency, people affected, injuries, anonymous/contact controls, accessibility needs, media initiation, offline retry messaging, and success/error states. |
| NADAA-033 | MVP Sprint 2   | Add Incident Deduplication Baseline                    | Done   | Codex | main      | NADAA-030            | 2026-07-06  | Incident-service now stores same-hazard duplicate candidates scored by distance, time window, and description similarity without merging or deleting reports.                                                              |
| NADAA-040 | MVP Sprint 3   | Build Incident Command Map                             | Done   | Codex | main      | NADAA-011, NADAA-030 | 2026-07-06  | Authority dashboard now has a Leaflet incident command map, API-backed incident feed with fixture fallback, map/list sync, filters, selected-incident detail, role-protected framing, and loading/empty/error states.      |
| NADAA-101 | MVP Sprint 1/7 | Set Up CI/CD And Staging Environment                   | Done   | Codex | main      | NADAA-001            | 2026-07-06  | GitHub Actions CI, manual staging smoke workflow, Docker build validation, staging env template, and staging runbook added; registry push/deploy credentials remain environment-owned.                                     |
| NADAA-021 | MVP Sprint 5   | Implement Area Risk API                                | Done   | Codex | main      | NADAA-020            | 2026-07-06  | Risk service now returns fixture-backed geospatial risk lookup with low/high/severe flood scoring, nearby shelters, nearby facilities, recommended actions, validation, docs, and tests.                                   |
| NADAA-022 | MVP Sprint 5   | Build Citizen Risk Checker UI                          | Done   | Codex | main      | NADAA-021            | 2026-07-06  | Citizen risk surface now uses the risk API with area presets, coordinate entry, GPS lookup, loading/error/permission/empty states, shelters, facilities, recommended actions, and smoke coverage.                          |
| NADAA-060 | MVP Sprint 5   | Implement Emergency Guide Content Model                | Done   | Codex | main      | NADAA-020            | 2026-07-06  | Guide-service API, guide content fixtures, seed expansion, shared guide types, docs, Docker/CI wiring, staging smoke wiring, and lookup tests added.                                                                       |
| NADAA-080 | MVP Sprint 6   | Define Agency Integration Contracts                    | Done   | Codex | main      | NADAA-002            | 2026-07-06  | Integration matrix, inbound weather/hydrology contracts, outbound incident/alert sync contracts, mock integration service, shared types, docs, CI/staging wiring, smoke script, Dockerfile, and tests added.               |

### Agent Handoff Log

| Date       | Agent | Item                | Status      | Handoff Notes                                                                                                                                                                                                                                                                                                    |
| ---------- | ----- | ------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-07-06 | Codex | NADAA-080           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, live `pnpm smoke:integration` on `localhost:8088`, local `pnpm smoke:staging`, integration-service Docker build, and `git diff --check`.                                              |
| 2026-07-06 | Codex | NADAA-080           | In Progress | Claimed integration matrix, inbound weather/hydrology contract, outbound incident/alert sync contract, mock integration service, docs, CI wiring, and tests.                                                                                                                                                     |
| 2026-07-06 | Codex | NADAA-060           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, live `pnpm smoke:guide` on `localhost:8087`, live `pnpm smoke:risk` on `localhost:8081`, local `pnpm smoke:staging`, `git diff --check`, and disposable PostGIS migration/seed apply. |
| 2026-07-06 | Codex | NADAA-060           | In Progress | Claimed guide-service API, guide content fixtures, seed expansion, shared guide types, docs, and lookup tests.                                                                                                                                                                                                   |
| 2026-07-06 | Codex | NADAA-040           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live incident command API smoke on `localhost:8084`.                                                                                                                              |
| 2026-07-06 | Codex | NADAA-040           | In Progress | Claimed authority-dashboard incident command map, API-backed list fallback, map/list sync, filters, role-protected framing, and UI states.                                                                                                                                                                       |
| 2026-07-06 | Codex | NADAA-012           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live auth/audit HTTP smoke on `localhost:8082`.                                                                                                                                   |
| 2026-07-06 | Codex | NADAA-012           | In Progress | Claimed auth-service audit event model, in-memory audit store/helper, auth/admin event wiring, tests, shared types, and retention docs.                                                                                                                                                                          |
| 2026-07-06 | Codex | NADAA-011           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live agency auth HTTP smoke on `localhost:8082`.                                                                                                                                  |
| 2026-07-06 | Codex | NADAA-011           | In Progress | Claimed agency user creation, role catalog, mock MFA setup/verification, and RBAC middleware/tests in auth-service; shared auth contracts and docs may be updated.                                                                                                                                               |
| 2026-07-06 | Codex | NADAA-022           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live `pnpm smoke:risk` on `localhost:8081`.                                                                                                                                       |
| 2026-07-06 | Codex | NADAA-022           | In Progress | Claimed citizen risk checker UI; depends on local/staging `VITE_RISK_API_URL` or default risk-service URL.                                                                                                                                                                                                       |
| 2026-07-06 | Codex | NADAA-021           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live risk HTTP smoke on `localhost:8081`.                                                                                                                                         |
| 2026-07-06 | Codex | NADAA-021           | In Progress | Claimed area risk API; using seed-aligned fixtures until PostGIS persistence is wired into services.                                                                                                                                                                                                             |
| 2026-07-06 | Codex | NADAA-101           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, local `pnpm smoke:staging`, and local Docker builds for five deployable images.                                                                                                       |
| 2026-07-06 | Codex | NADAA-101           | In Progress | Claimed CI/CD and staging foundation; credentials and provider deployment remain environment-owned, not committed.                                                                                                                                                                                               |
| 2026-07-06 | Codex | NADAA-033           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live duplicate candidate HTTP smoke on `localhost:8084`.                                                                                                                                                    |
| 2026-07-06 | Codex | NADAA-033           | In Progress | Claimed incident deduplication baseline; focusing on candidate scoring/storage without merging or deleting reports.                                                                                                                                                                                              |
| 2026-07-06 | Codex | NADAA-032           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live citizen incident UI/API smoke on `localhost:8084`.                                                                                                                                                     |
| 2026-07-06 | Codex | NADAA-032           | In Progress | Claimed citizen incident reporting UI; integrating report submission with media upload initiation and existing incident-service API.                                                                                                                                                                             |
| 2026-07-06 | Codex | NADAA-031           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live media upload/link smoke on `localhost:8084`.                                                                                                                                                           |
| 2026-07-06 | Codex | NADAA-031           | In Progress | Claimed media upload flow; using controlled dev upload URLs and private in-memory metadata for this slice.                                                                                                                                                                                                       |
| 2026-07-06 | Codex | NADAA-030           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live incident HTTP smoke on `localhost:8084`.                                                                                                                                                               |
| 2026-07-06 | Codex | NADAA-030           | In Progress | Claimed incident reporting API; media upload storage and deduplication remain NADAA-031/NADAA-033.                                                                                                                                                                                                               |
| 2026-07-06 | Codex | NADAA-010           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live auth HTTP smoke on `localhost:8082`.                                                                                                                                                                   |
| 2026-07-06 | Codex | NADAA-010           | In Progress | Claimed citizen authentication service slice; agency MFA remains NADAA-011.                                                                                                                                                                                                                                      |
| 2026-07-06 | Codex | NADAA-100/NADAA-020 | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and PostGIS migration/seed on `localhost:55432`.                                                                                                                                                                |
| 2026-07-06 | Codex | NADAA-100/NADAA-020 | In Progress | Claimed QA matrix and PostGIS geospatial foundation.                                                                                                                                                                                                                                                             |
| 2026-07-06 | Codex | NADAA-002/NADAA-003 | Done        | Verified `pnpm validate:dashboard`, `pnpm typecheck`, `pnpm build`, and `pnpm go:test`.                                                                                                                                                                                                                          |
| 2026-07-06 | Codex | NADAA-002/NADAA-003 | In Progress | Claimed documentation expansion and delivery dashboard contract.                                                                                                                                                                                                                                                 |
| 2026-07-06 | Codex | NADAA-001           | Done        | Verified `pnpm typecheck`, `pnpm build`, `pnpm go:test`, and app HTTP smoke checks on ports 5173 and 5174.                                                                                                                                                                                                       |
| 2026-07-06 | Codex | NADAA-001           | In Progress | Claimed repository foundation and initial scaffold.                                                                                                                                                                                                                                                              |
| 2026-07-06 | Codex | agent_plan.md       | Done        | Initial plan created and expanded for multi-agent coordination.                                                                                                                                                                                                                                                  |

## Definition Of Done

A feature is done only when:

- Code is implemented and reviewed.
- Unit/integration tests pass for the changed surface.
- Security-sensitive paths have access control and audit coverage.
- User-facing flows have empty, loading, error, and success states.
- Documentation is updated where workflow, API, deployment, or operating behavior changes.
- Jira/project status and this plan are updated.
- The feature has been validated in the appropriate environment.

## MVP Architecture Workstreams

- Apps:
  - `apps/citizen-web` - citizen PWA for alerts, risk checks, incident reporting, shelters, and emergency guidance.
  - `apps/authority-dashboard` - incident command, verification, alerts, assignment, maps, and audit views.
  - `apps/mobile-app` - phase 2 unless mobile is prioritized earlier.
- Services:
  - `services/auth-service`
  - `services/incident-service`
  - `services/alert-service`
  - `services/risk-service`
  - `services/dispatch-service`
  - `services/notification-service`
  - `services/integration-service`
  - `services/ml-service`
- Packages:
  - `packages/shared-types`
  - `packages/ui`
  - `packages/config`
- Infrastructure and docs:
  - `infra/docker`
  - `infra/kubernetes`
  - `infra/terraform`
  - `docs/architecture.md`
  - `docs/api.md`
  - `docs/ml.md`
  - `docs/security.md`
  - `docs/deployment.md`

## Epic Roadmap

### EPIC 0: Project Foundation And Governance

Goal: establish the repository, engineering standards, delivery governance, and project documentation needed before feature development accelerates.

Stories:

#### NADAA-001: Create Repository And Monorepo Foundation

- User story: As an engineering team, we need a consistent project structure so apps, services, packages, docs, and infra can evolve without confusion.
- Business value: Reduces delivery friction and keeps all agents working from the same architecture.
- Acceptance criteria:
  - Suggested repo structure from `spec.md` exists.
  - Shared config package is available.
  - README explains local setup and service layout.
  - `AGENTS.md` and `CLAUDE.md` exist with project-specific operating rules.
- Tasks:
  - Choose monorepo tooling.
  - Create app/service/package/infra/docs directories.
  - Add shared linting, formatting, TypeScript, and Go conventions.
  - Add initial README, `AGENTS.md`, and `CLAUDE.md`.
- Estimate: 5 points.
- Dependencies: none.

#### NADAA-002: Define Product, API, Security, ML, And Deployment Docs

- User story: As a stakeholder, I need clear implementation documents so delivery aligns with the approved product vision.
- Business value: Creates traceability from requirements to implementation.
- Acceptance criteria:
  - Architecture, API, ML, security, and deployment docs exist.
  - MVP scope is separated from phase 2 and phase 3 scope.
  - Data ownership and integration assumptions are documented.
- Tasks:
  - Convert relevant parts of `spec.md` into `docs/architecture.md`.
  - Draft `docs/api.md` with core API contracts.
  - Draft `docs/security.md` with RBAC, MFA, alert approval, audit, privacy, retention, and media requirements.
  - Draft `docs/ml.md` with flood model inputs, outputs, safety constraints, and evaluation plan.
  - Draft `docs/deployment.md` with local, staging, and production environments.
- Estimate: 5 points.
- Dependencies: NADAA-001.

#### NADAA-003: Create Delivery Dashboard Data Contract

- User story: As a project manager, I need progress and traceability data so stories, branches, PRs, estimates, and delivery states can be tracked consistently.
- Business value: Enables the dashboard required by the operating manuals.
- Acceptance criteria:
  - Dashboard fields match the manuals: client, project, epic, story, Jira key, branch, PR, status, assignee, estimates, actual effort, progress, and last updated.
  - JSON/schema contract exists for future dashboard sync.
- Tasks:
  - Define project tracking schema.
  - Document Jira/GitHub mapping.
  - Add sample records for MVP epics.
- Estimate: 3 points.
- Dependencies: NADAA-001.

### EPIC 1: Identity, Access Control, And Agency Model

Goal: provide secure identity for citizens, agency users, dispatchers, responders, and admins.

Stories:

#### NADAA-010: Implement Citizen Authentication

- User story: As a citizen, I want to register and log in with my phone number so I can report incidents and receive relevant alerts.
- Business value: Enables trusted citizen participation while preserving broad access.
- Acceptance criteria:
  - Citizen can register/login using phone-based flow.
  - User profile stores phone, name, preferred language, home location, and contact permission.
  - Anonymous incident reporting remains possible where allowed.
- Tasks:
  - Define user schema and migration.
  - Implement registration/login API.
  - Add session/token handling.
  - Add citizen profile endpoints.
  - Add tests for registration, login, duplicate phone, and invalid credentials.
- Estimate: 8 points.
- Dependencies: NADAA-001.

#### NADAA-011: Implement Agency Users, Roles, And MFA

- User story: As an authority user, I need secure role-based access with MFA so sensitive incident, alert, and assignment workflows are protected.
- Business value: Protects public safety workflows and reduces risk of unauthorized alerts.
- Acceptance criteria:
  - Agency users belong to agencies and roles.
  - Roles cover admin, NADMO officer, district officer, dispatcher, agency responder, and viewer.
  - MFA is required for authority users.
  - Access is denied for unauthorized roles.
- Tasks:
  - Define agency and role schema.
  - Implement agency user invitation or creation flow.
  - Implement MFA setup and verification.
  - Add RBAC middleware.
  - Add authorization tests per role.
- Estimate: 13 points.
- Dependencies: NADAA-010.

#### NADAA-012: Implement Audit Logging Foundation

- User story: As a system admin, I need audit logs for sensitive actions so incident, alert, assignment, and admin activity is traceable.
- Business value: Supports accountability, investigation, and safety governance.
- Acceptance criteria:
  - Login, alert, incident verification, assignment, and admin changes are auditable.
  - Audit records include actor, action, target, timestamp, IP/device metadata where available, and before/after context where appropriate.
- Tasks:
  - Define audit log schema.
  - Add audit logging middleware/helper.
  - Wire audit events into auth and role changes.
  - Document audit retention assumptions.
- Estimate: 5 points.
- Dependencies: NADAA-011.

### EPIC 2: Geospatial Data And Risk Checker

Goal: let citizens and authorities search or select an area and understand hazard risk, nearby facilities, and recommended actions.

Stories:

#### NADAA-020: Set Up PostGIS And Core Geospatial Models

- User story: As the platform, I need geospatial storage for incidents, risk zones, shelters, agencies, and target alert areas.
- Business value: Geospatial capability is central to alerts, response, and risk checking.
- Acceptance criteria:
  - PostgreSQL/PostGIS is available locally.
  - Core tables exist for users, agencies, incidents, alerts, risk zones, shelters, emergency guides, and ML predictions.
  - Geometry fields are indexed.
- Tasks:
  - Add database service to local Docker setup.
  - Create migrations for core tables.
  - Add geospatial indexes.
  - Add seed data for Ghana regions/district samples, shelters, and agencies.
- Estimate: 8 points.
- Dependencies: NADAA-001.

#### NADAA-021: Implement Area Risk API

- User story: As a citizen, I want to check the risk level for an area so I know how to prepare or respond.
- Business value: Delivers one of the core MVP promises.
- Acceptance criteria:
  - `GET /api/v1/risk?lat={lat}&lng={lng}` returns location, overall risk, hazard-specific risks, nearby shelters, and recommended actions.
  - Risk levels support low, moderate, high, severe, and emergency.
  - Flood risk is prioritized for MVP.
  - API handles invalid coordinates and unsupported areas.
- Tasks:
  - Implement risk service endpoint.
  - Query nearby shelters and facilities.
  - Add rule-based baseline flood scoring.
  - Add response DTOs in shared types.
  - Add API tests for low, high, severe, and invalid-coordinate cases.
- Estimate: 13 points.
- Dependencies: NADAA-020.

#### NADAA-022: Build Citizen Risk Checker UI

- User story: As a citizen, I want to search or use my current location to view nearby risks and safety advice.
- Business value: Gives citizens actionable preparedness information.
- Acceptance criteria:
  - Citizen can search/select an area or use device location.
  - UI shows overall risk, flood risk, nearby shelters, nearby emergency facilities, and recommended actions.
  - Loading, permission denied, empty, and error states are handled.
  - View is usable on mobile-sized screens.
- Tasks:
  - Build risk checker page.
  - Integrate map and location permission.
  - Render risk cards and facility list.
  - Add responsive layout and accessibility labels.
  - Add smoke/e2e coverage for risk lookup.
- Estimate: 13 points.
- Dependencies: NADAA-021.

### EPIC 3: Citizen Incident Reporting

Goal: allow citizens to report disasters and emergencies with GPS, media, affected people, injuries, urgency, accessibility needs, anonymous option, and contact permission.

Stories:

#### NADAA-030: Implement Incident Reporting API

- User story: As a citizen, I want to submit a disaster report so authorities can verify and respond.
- Business value: Creates the primary intake channel for emergency intelligence.
- Acceptance criteria:
  - `POST /api/v1/incidents` accepts hazard type, GPS location, description, people affected, injuries, urgency, anonymous flag, contact permission, accessibility needs, and media references.
  - Incident status starts as reported.
  - Input validation prevents invalid coordinates, unsupported hazards, and unsafe payloads.
  - Life-threatening reports can be flagged for priority review.
- Tasks:
  - Implement incident schema and service.
  - Implement create incident endpoint.
  - Add validation and rate limiting.
  - Add media metadata support.
  - Add tests for valid report, anonymous report, invalid location, and high-urgency report.
- Estimate: 13 points.
- Dependencies: NADAA-020, NADAA-010.

#### NADAA-031: Implement Media Upload Flow

- User story: As a citizen, I want to attach photos, videos, or audio so responders can assess the situation.
- Business value: Improves verification and triage quality.
- Acceptance criteria:
  - Supported media can be uploaded through signed or controlled upload flow.
  - Media is linked to an incident.
  - File type and size restrictions are enforced.
  - Stored media is not public by default.
- Tasks:
  - Select object storage provider/configuration.
  - Implement upload initiation endpoint.
  - Implement media metadata persistence.
  - Add validation and error handling.
  - Document media storage and privacy behavior.
- Estimate: 8 points.
- Dependencies: NADAA-030.

#### NADAA-032: Build Citizen Incident Reporting UI

- User story: As a citizen, I want a simple mobile-first form to report an incident quickly during an emergency.
- Business value: Converts the incident API into a usable public safety workflow.
- Acceptance criteria:
  - Form supports hazard type, location, description, media, people affected, injuries, urgency, anonymous option, contact permission, and accessibility needs.
  - Device GPS can populate location.
  - User sees submission success, reference number, and emergency contact reminder.
  - Offline/interrupted submission behavior is defined.
- Tasks:
  - Build incident report page.
  - Add GPS capture and map preview.
  - Add media attachment controls.
  - Add validation and progressive disclosure for urgent fields.
  - Add success and failure states.
- Estimate: 13 points.
- Dependencies: NADAA-030, NADAA-031.

#### NADAA-033: Add Incident Deduplication Baseline

- User story: As a dispatcher, I need duplicate reports to be grouped so the dashboard is not flooded with repeated incident submissions.
- Business value: Reduces dispatcher workload during high-volume events.
- Acceptance criteria:
  - New reports are compared by hazard type, distance, time window, and description similarity.
  - Potential duplicates are surfaced for authority review.
  - No report is silently deleted.
- Tasks:
  - Implement duplicate candidate query.
  - Add deduplication score.
  - Store duplicate linkage candidates.
  - Add tests around same-location and nearby-location reports.
- Estimate: 8 points.
- Dependencies: NADAA-030.

### EPIC 4: Authority Incident Command Dashboard

Goal: give authorized agencies a live incident map and workflow for verification, assignment, status tracking, notes, escalation, duplicate merging, and closure.

Stories:

#### NADAA-040: Build Incident Command Map

- User story: As a dispatcher, I want a live map of incidents so I can monitor emergencies by place, severity, time, and hazard.
- Business value: Gives authorities operational visibility.
- Acceptance criteria:
  - Dashboard map displays incidents with severity and status.
  - Filters include hazard, region/district, severity, status, and time.
  - Incident list and map remain synchronized.
  - Only authorized users can access the dashboard.
- Tasks:
  - Build authority dashboard shell.
  - Implement incident list API.
  - Add map markers and filters.
  - Add role-protected routes.
  - Add loading, empty, and error states.
- Estimate: 13 points.
- Dependencies: NADAA-011, NADAA-030.

#### NADAA-041: Implement Verification And Status Workflow

- User story: As a dispatcher, I want to verify reports and move incidents through response statuses so everyone sees current operational state.
- Business value: Converts raw reports into managed incidents.
- Acceptance criteria:
  - Supported statuses: reported, under review, verified, assigned, response en route, on scene, contained, recovery ongoing, closed, false report.
  - Status changes are audited.
  - Resolution notes are required for closed and false report statuses.
- Tasks:
  - Implement status transition rules.
  - Add verification endpoint.
  - Add status update UI.
  - Add audit events.
  - Add tests for valid and invalid transitions.
- Estimate: 13 points.
- Dependencies: NADAA-040, NADAA-012.

#### NADAA-042: Implement Agency Assignment And Incident Timeline

- User story: As a dispatcher, I want to assign incidents to agencies and track timeline events so response coordination is visible.
- Business value: Reduces response delays and improves accountability.
- Acceptance criteria:
  - Incidents can be assigned to police, fire, ambulance, NADMO, district assembly, or other configured agencies.
  - Timeline records report creation, verification, assignment, notes, status changes, and closure.
  - Assigned agency users can see their assigned incidents.
- Tasks:
  - Implement assignment model.
  - Implement timeline model.
  - Add assignment UI.
  - Add assigned incident view for agency users.
  - Add tests for assignment permissions.
- Estimate: 13 points.
- Dependencies: NADAA-041.

#### NADAA-043: Implement Duplicate Merge Review

- User story: As a dispatcher, I want to review and merge duplicate reports so one real event has a single operational record.
- Business value: Improves dashboard clarity and incident response coordination.
- Acceptance criteria:
  - Potential duplicate reports can be reviewed side by side.
  - Dispatcher can merge reports into a primary incident.
  - Merged reports remain traceable.
- Tasks:
  - Add duplicate review endpoint.
  - Add merge endpoint.
  - Add side-by-side UI.
  - Add audit records for merge actions.
- Estimate: 8 points.
- Dependencies: NADAA-033, NADAA-041.

### EPIC 5: Alerts And Notification Delivery

Goal: enable authorities to create, approve, target, deliver, and audit public safety alerts.

Stories:

#### NADAA-050: Implement Alert Creation And Approval Workflow

- User story: As a NADMO or district officer, I want to draft and submit alerts for approval so warnings are accurate and controlled.
- Business value: Supports public warning while reducing false or unauthorized alerts.
- Acceptance criteria:
  - Alerts include title, hazard type, severity, message, affected area, start time, expiry time, recommended action, evacuation flag, shelter links, issuing agency, and approval status.
  - Mass alerts require approval before delivery.
  - Emergency override is role-protected and audited.
- Tasks:
  - Implement alert schema and workflow statuses.
  - Implement create/update/submit/approve/reject endpoints.
  - Add authority alert form.
  - Add approval queue.
  - Add audit events and tests.
- Estimate: 13 points.
- Dependencies: NADAA-011, NADAA-012, NADAA-020.

#### NADAA-051: Implement Geofenced Alert Targeting

- User story: As an authority user, I want to target alerts nationally, regionally, by district, radius, or custom geometry so only affected populations receive them.
- Business value: Improves relevance and reduces alert fatigue.
- Acceptance criteria:
  - Target types include national, region, district, radius, community, and custom geometry.
  - Target geometry is stored and queryable.
  - Preview shows approximate affected area before approval.
- Tasks:
  - Implement target geometry handling.
  - Add district/radius selectors.
  - Add affected area preview.
  - Add tests for target selection.
- Estimate: 8 points.
- Dependencies: NADAA-050.

#### NADAA-052: Implement In-App Alert Feed And Push/SMS Abstraction

- User story: As a citizen, I want to receive warnings in the app and through fallback channels when available.
- Business value: Turns authority alerts into citizen reach.
- Acceptance criteria:
  - Citizen app shows current and expired alerts.
  - Notification service has provider abstraction for push and SMS.
  - Delivery attempts are logged.
  - SMS can be disabled or mocked in development.
- Tasks:
  - Build alert feed API.
  - Build citizen alert feed UI.
  - Add notification provider interface.
  - Implement development/mock provider.
  - Add delivery log schema.
- Estimate: 13 points.
- Dependencies: NADAA-050, NADAA-051.

### EPIC 6: Emergency Guidance, Shelters, And Recovery Support

Goal: provide offline-first preparedness, emergency, and recovery guidance plus shelter and emergency facility visibility.

Stories:

#### NADAA-060: Implement Emergency Guide Content Model

- User story: As a citizen, I want emergency guidance before, during, and after disasters so I know what to do.
- Business value: Improves public preparedness and reduces preventable harm.
- Acceptance criteria:
  - Guides support hazard type, stage, title, body, language, and offline availability.
  - Initial guides cover floods, fire safety, road crash response, electrical hazard safety, disease prevention, safe evacuation, emergency bag checklist, family emergency planning, and contacting 112.
- Tasks:
  - Implement guide schema and seed content.
  - Add guide API.
  - Add admin/editor notes for future CMS.
  - Add tests for guide lookup by hazard/stage/language.
- Estimate: 8 points.
- Dependencies: NADAA-020.

#### NADAA-061: Build Offline-First Citizen Guidance UI

- User story: As a citizen, I want guidance available even when internet access is poor so I can use it during emergencies.
- Business value: Makes preparedness content reliable under disaster conditions.
- Acceptance criteria:
  - Citizen can browse emergency guides.
  - Key guidance is cached for offline use.
  - UI is mobile-first and accessible.
  - The 112 emergency contact is visible where appropriate.
- Tasks:
  - Build guidance pages.
  - Add PWA caching for selected guides.
  - Add language-ready content structure.
  - Add offline state handling.
- Estimate: 8 points.
- Dependencies: NADAA-060.

#### NADAA-062: Implement Shelter And Recovery Support Module

- User story: As an affected citizen, I want to find shelters, relief points, medical support, and recovery guidance after an event.
- Business value: Supports recovery and reduces confusion after disasters.
- Acceptance criteria:
  - Shelters include location, capacity, occupancy, contact, and facilities.
  - Citizen app can show nearby shelters and recovery support locations.
  - Authority users can update shelter capacity and occupancy.
- Tasks:
  - Implement shelter API.
  - Seed initial shelter/facility data.
  - Add citizen shelter map/list.
  - Add authority shelter update view.
  - Add tests for nearby shelter lookup.
- Estimate: 13 points.
- Dependencies: NADAA-020, NADAA-022.

### EPIC 7: Flood ML Risk MVP

Goal: deliver a human-supervised flood risk model that produces explainable risk predictions for maps, APIs, and alert recommendations.

Stories:

#### NADAA-070: Create Flood Risk Dataset And Feature Pipeline

- User story: As a data/ML engineer, I need a repeatable feature pipeline so flood prediction can improve over time.
- Business value: Establishes the foundation for the platform's disaster intelligence promise.
- Acceptance criteria:
  - Pipeline ingests or prepares rainfall, elevation, slope, distance to river/drain, historical reports, land use, population density, and district/community geometry where available.
  - Dataset limitations and missing sources are documented.
  - Feature outputs are versioned.
- Tasks:
  - Identify available public/open datasets.
  - Define feature schema.
  - Build initial feature generation script or pipeline.
  - Document external agency data dependencies.
  - Add sample data for development.
- Estimate: 13 points.
- Dependencies: NADAA-020, NADAA-002.

#### NADAA-071: Train Baseline Flood Risk Model

- User story: As an authority user, I want flood probability and severity estimates so warnings and preparedness decisions are data-informed.
- Business value: Delivers first ML capability while remaining explainable and reviewable.
- Acceptance criteria:
  - Baseline model uses logistic regression, random forest, or XGBoost.
  - Outputs include probability, severity, expected onset where available, confidence level, model version, and explanation factors.
  - Evaluation metrics and limitations are documented.
- Tasks:
  - Train baseline model.
  - Track experiment/version metadata.
  - Generate sample predictions.
  - Document false positive/false negative review process.
  - Add model evaluation report.
- Estimate: 13 points.
- Dependencies: NADAA-070.

#### NADAA-072: Serve ML Predictions Through Risk API

- User story: As the platform, I need ML flood predictions available to the risk service and authority dashboard.
- Business value: Makes model output usable in product workflows.
- Acceptance criteria:
  - ML service exposes prediction endpoint.
  - Risk service can combine rule-based and ML risk outputs.
  - Predictions are logged with model version.
  - Predictions cannot automatically publish public alerts.
- Tasks:
  - Create FastAPI ML service.
  - Add model loading and prediction endpoint.
  - Add service-to-service contract.
  - Store predictions in `ml_predictions`.
  - Add integration tests with fixture predictions.
- Estimate: 13 points.
- Dependencies: NADAA-071, NADAA-021.

#### NADAA-073: Add Authority ML Review View

- User story: As a NADMO or district officer, I want to review model recommendations and explanations before issuing alerts.
- Business value: Keeps public safety decisions human-approved.
- Acceptance criteria:
  - Authority dashboard displays flood probability, severity, confidence, and explanation factors.
  - User can create an alert draft from a prediction.
  - Alert still requires approval workflow.
- Tasks:
  - Build ML prediction map/list.
  - Add explanation panel.
  - Add "create alert draft" action.
  - Add audit logging for draft creation.
- Estimate: 8 points.
- Dependencies: NADAA-072, NADAA-050.

### EPIC 8: Integrations And External Data

Goal: prepare the platform to exchange data with Ghana emergency actors and external weather/hydrology/geospatial sources.

Stories:

#### NADAA-080: Define Agency Integration Contracts

- User story: As an integration engineer, I need clear contracts for NADMO, GMet, Ghana Hydrological Authority, police, fire, ambulance, district assemblies, hospitals, and utilities.
- Business value: Enables future official integrations without blocking MVP.
- Acceptance criteria:
  - Contracts document data ownership, frequency, expected payloads, authentication, and failure behavior.
  - Mock adapters exist for development.
- Tasks:
  - Draft integration matrix.
  - Define inbound weather/hydrology adapter contract.
  - Define outbound incident/alert sync contract.
  - Add mock integration service.
- Estimate: 8 points.
- Dependencies: NADAA-002.

#### NADAA-081: Implement Weather And Hydrology Import Skeleton

- User story: As the risk system, I need weather and hydrology inputs so flood risk can respond to changing conditions.
- Business value: Supports rainfall-based alert trigger and model improvement.
- Acceptance criteria:
  - Import job can ingest fixture rainfall/water-level data.
  - Imported observations are stored with source, timestamp, location, and validity window.
  - Failed imports are logged and retryable.
- Tasks:
  - Define observation schema.
  - Implement fixture importer.
  - Add scheduled job hook.
  - Add import status logging.
- Estimate: 8 points.
- Dependencies: NADAA-020, NADAA-080.

### EPIC 9: Security, Privacy, Safety, And Abuse Controls

Goal: protect users, prevent misuse, and meet safety expectations for emergency workflows.

Stories:

#### NADAA-090: Implement Location Privacy And Anonymous Reporting Controls

- User story: As a citizen, I want privacy controls so I can report emergencies without unnecessary exposure.
- Business value: Builds trust and encourages reporting.
- Acceptance criteria:
  - Anonymous reports hide citizen identity from standard authority views where policy allows.
  - Contact permission is explicit.
  - Location data use is disclosed in product copy.
- Tasks:
  - Implement privacy fields.
  - Adjust incident views for anonymous reports.
  - Add backend authorization checks.
  - Document privacy behavior.
- Estimate: 8 points.
- Dependencies: NADAA-030, NADAA-040.

#### NADAA-091: Implement Abuse, Spam, And False Report Handling

- User story: As a dispatcher, I need tools to reduce spam and identify false reports without suppressing urgent legitimate reports.
- Business value: Preserves operational trust during emergencies.
- Acceptance criteria:
  - Rate limits apply to incident submissions.
  - Suspicious report signals are visible to dispatchers.
  - False report closure requires a reason.
  - Life-threatening reports are not automatically blocked only by suspicion score.
- Tasks:
  - Add rate limits.
  - Add suspicious report fields.
  - Add false report workflow.
  - Add tests for rate limiting and false report closure.
- Estimate: 8 points.
- Dependencies: NADAA-030, NADAA-041.

#### NADAA-092: Security Review And Hardening

- User story: As an engineering lead, I need security validation before UAT so public safety features are safe to test with real users.
- Business value: Reduces production and public harm risk.
- Acceptance criteria:
  - Auth, RBAC, MFA, alert approval, media storage, audit logging, and rate limits are reviewed.
  - Dependency and container scans are run.
  - Critical and high findings are resolved or explicitly accepted.
- Tasks:
  - Perform threat model pass.
  - Run dependency scans.
  - Review environment variable handling.
  - Review public alert approval paths.
  - Document residual risks.
- Estimate: 8 points.
- Dependencies: MVP feature completion.

### EPIC 10: QA, Release, UAT, And Operations

Goal: move the MVP through QA, staging, UAT, beta, production, and hypercare using the manuals' process.

Stories:

#### NADAA-100: Build Test Strategy And QA Matrix

- User story: As a QA engineer, I need a test plan so MVP flows can be validated consistently.
- Business value: Improves release confidence.
- Acceptance criteria:
  - Test matrix covers citizen risk check, incident report, media upload, authority verification, assignment, alert approval, alert feed, shelter lookup, guide offline access, and ML review.
  - Regression checklist exists for release candidates.
- Tasks:
  - Define test levels and ownership.
  - Add core e2e scenarios.
  - Add API test checklist.
  - Add manual test scripts for emergency workflows.
- Estimate: 8 points.
- Dependencies: NADAA-002.

#### NADAA-101: Set Up CI/CD And Staging Environment

- User story: As an engineering team, we need automated validation and staging deployment so releases are repeatable.
- Business value: Reduces deployment risk and supports UAT.
- Acceptance criteria:
  - CI runs lint, tests, type checks, and builds for changed apps/services.
  - Staging deployment is documented.
  - Smoke tests run against staging.
- Tasks:
  - Add GitHub Actions workflows.
  - Add Docker build pipeline.
  - Add environment variable templates.
  - Add staging smoke test script.
- Estimate: 13 points.
- Dependencies: NADAA-001.

#### NADAA-102: Conduct UAT, Beta, And Production Readiness

- User story: As a stakeholder, I need structured UAT and release sign-off so the MVP is accepted before launch.
- Business value: Aligns product delivery with official approval gates.
- Acceptance criteria:
  - UAT feedback is captured as defects or enhancements.
  - Beta metrics are monitored.
  - Release notes, user guide, training material, and acceptance checklist are prepared.
- Tasks:
  - Prepare UAT scripts.
  - Create release notes template.
  - Create user guide/training outline.
  - Define beta metrics dashboard.
  - Prepare hypercare support checklist.
- Estimate: 8 points.
- Dependencies: NADAA-100, NADAA-101, MVP feature completion.

## Sprint Plan

### Sprint 0: Discovery, Planning, And Foundation

Objective: convert documents into executable delivery assets and prepare the repository.

Stories:

- NADAA-001 Create Repository And Monorepo Foundation.
- NADAA-002 Define Product, API, Security, ML, And Deployment Docs.
- NADAA-003 Create Delivery Dashboard Data Contract.
- NADAA-100 Build Test Strategy And QA Matrix.

Deliverables:

- Monorepo scaffold.
- Project docs.
- Initial API/security/ML/deployment plans.
- QA matrix.
- First sprint backlog ready for implementation.

### Sprint 1: Core Platform, Auth, And Geospatial Base

Objective: create the core runtime foundation needed by all MVP features.

Stories:

- NADAA-010 Implement Citizen Authentication.
- NADAA-011 Implement Agency Users, Roles, And MFA.
- NADAA-012 Implement Audit Logging Foundation.
- NADAA-020 Set Up PostGIS And Core Geospatial Models.
- NADAA-101 Set Up CI/CD And Staging Environment, first pass.

Deliverables:

- Auth and RBAC foundation.
- Agency model.
- Audit logging base.
- Core database migrations with geospatial indexes.
- Local Docker/database setup.

### Sprint 2: Citizen Reporting MVP

Objective: let citizens report emergencies and give dispatchers incoming report data.

Stories:

- NADAA-030 Implement Incident Reporting API.
- NADAA-031 Implement Media Upload Flow.
- NADAA-032 Build Citizen Incident Reporting UI.
- NADAA-033 Add Incident Deduplication Baseline.
- NADAA-090 Implement Location Privacy And Anonymous Reporting Controls, first pass.

Deliverables:

- Citizen incident report flow.
- Media attachment support.
- Anonymous/contact-permission behavior.
- Duplicate candidate detection.

### Sprint 3: Authority Incident Command MVP

Objective: enable incident review, verification, assignment, and operational tracking.

Stories:

- NADAA-040 Build Incident Command Map.
- NADAA-041 Implement Verification And Status Workflow.
- NADAA-042 Implement Agency Assignment And Incident Timeline.
- NADAA-043 Implement Duplicate Merge Review.
- NADAA-091 Implement Abuse, Spam, And False Report Handling, first pass.

Deliverables:

- Authority dashboard shell.
- Live incident map/list.
- Verification and status workflow.
- Agency assignment and timeline.
- Duplicate merge workflow.

### Sprint 4: Alerts And Public Warning

Objective: let authorities create, approve, target, deliver, and audit alerts.

Stories:

- NADAA-050 Implement Alert Creation And Approval Workflow.
- NADAA-051 Implement Geofenced Alert Targeting.
- NADAA-052 Implement In-App Alert Feed And Push/SMS Abstraction.

Deliverables:

- Alert draft/approval workflow.
- Targeted district/radius/geometry alerts.
- Citizen alert feed.
- Notification delivery abstraction and logs.

### Sprint 5: Risk Checker, Guides, Shelters, And Recovery

Objective: deliver citizen preparedness and recovery workflows.

Stories:

- NADAA-021 Implement Area Risk API.
- NADAA-022 Build Citizen Risk Checker UI.
- NADAA-060 Implement Emergency Guide Content Model.
- NADAA-061 Build Offline-First Citizen Guidance UI.
- NADAA-062 Implement Shelter And Recovery Support Module.

Deliverables:

- Area risk API and citizen UI.
- Emergency guide content and offline PWA behavior.
- Shelter and recovery map/list.

### Sprint 6: Flood ML MVP And Integration Skeletons

Objective: add baseline flood intelligence and prepare official data integration paths.

Stories:

- NADAA-070 Create Flood Risk Dataset And Feature Pipeline.
- NADAA-071 Train Baseline Flood Risk Model.
- NADAA-072 Serve ML Predictions Through Risk API.
- NADAA-073 Add Authority ML Review View.
- NADAA-080 Define Agency Integration Contracts.
- NADAA-081 Implement Weather And Hydrology Import Skeleton.

Deliverables:

- Baseline flood model and documented limitations.
- ML prediction service.
- Human-reviewed ML recommendation view.
- Mock/fixture weather and hydrology imports.
- Integration contracts for official agencies.

### Sprint 7: Hardening, UAT, Beta, And Release Readiness

Objective: prepare the MVP for stakeholder testing and controlled release.

Stories:

- NADAA-092 Security Review And Hardening.
- NADAA-100 Build Test Strategy And QA Matrix, final pass.
- NADAA-101 Set Up CI/CD And Staging Environment, final pass.
- NADAA-102 Conduct UAT, Beta, And Production Readiness.

Deliverables:

- Security review and risk register.
- Staging release candidate.
- UAT scripts and sign-off package.
- Beta monitoring plan.
- Release notes, user guide, and hypercare checklist.

## Phase 2 Detailed Plan

Phase 2 extends the MVP from web-first emergency reporting and alerts into inclusive access, field coordination, recovery logistics, and richer operational data.

### EPIC 11: Inclusive Warning And Access Channels

Goal: reach citizens who do not have smartphones, stable internet, or high literacy, while keeping all public warning workflows auditable and authority-approved.

#### NADAA-110: SMS/USSD Emergency Access

- Outcome: citizens can check alerts, report basic incidents, and receive safety instructions through SMS/USSD.
- Acceptance criteria:
  - USSD menu supports language selection, current alerts, report emergency, shelter lookup, and 112 guidance.
  - SMS fallback can send alert summaries and basic report confirmations.
  - Messages are linked to citizen profiles when consent and phone identity are available.
  - Delivery failures and provider errors are logged.
- Tasks:
  - Select SMS/USSD provider abstraction.
  - Define USSD menu tree and message templates.
  - Implement inbound SMS/USSD webhook handling.
  - Map SMS/USSD reports into incident intake.
  - Add provider sandbox tests.
- Estimate: 13 points.
- Dependencies: NADAA-010, NADAA-030, NADAA-052.

#### NADAA-111: WhatsApp Emergency Chatbot

- Outcome: citizens can use WhatsApp to ask for alerts, report incidents, find shelters, and receive emergency guidance.
- Acceptance criteria:
  - Chatbot supports alerts, risk check, incident report, shelter lookup, emergency guides, and escalation to 112 guidance.
  - Conversation state handles incomplete reports and location sharing.
  - Authority-approved alert content is reused without manual retyping.
- Tasks:
  - Define WhatsApp Business API provider contract.
  - Build conversation intents and state machine.
  - Implement location and media ingestion.
  - Add incident handoff into incident service.
  - Add transcript and privacy retention rules.
- Estimate: 13 points.
- Dependencies: NADAA-030, NADAA-052, NADAA-060.

#### NADAA-112: Multilingual Voice Alerts

- Outcome: approved alerts can be delivered as voice messages in English, Twi, Ga, Ewe, Dagbani, and Hausa.
- Acceptance criteria:
  - Alert approval workflow can request voice variants.
  - Voice messages are generated, reviewed, approved, and delivered through supported channels.
  - Low-literacy safety messaging remains concise and action-oriented.
- Tasks:
  - Define multilingual message templates.
  - Select text-to-speech or recorded-audio workflow.
  - Add voice asset review step.
  - Add voice delivery logs.
  - Add accessibility QA checklist.
- Estimate: 13 points.
- Dependencies: NADAA-050, NADAA-052.

### EPIC 12: Community Response And Health Operations

Goal: help emergency actors coordinate volunteers, hospital capacity, relief distribution, and donated support during and after events.

#### NADAA-120: Community Volunteer App

- Outcome: verified volunteers can receive assignments, update status, and report field observations.
- Acceptance criteria:
  - Volunteers can register, be verified, and join a district/community response group.
  - Authorities can assign volunteer tasks.
  - Volunteer updates appear in incident timelines.
- Tasks:
  - Define volunteer profile and verification model.
  - Build volunteer assignment API.
  - Build mobile/PWA volunteer task view.
  - Add safety and escalation rules.
  - Add audit events for volunteer assignment.
- Estimate: 13 points.
- Dependencies: NADAA-011, NADAA-042.

#### NADAA-121: Hospital Capacity Tracker

- Outcome: emergency teams can see hospital availability for beds, emergency units, ambulances, and special services.
- Acceptance criteria:
  - Hospital users can update capacity manually or through integration adapters.
  - Dispatchers can filter hospitals by distance, service type, and capacity.
  - Capacity changes are timestamped and source-tracked.
- Tasks:
  - Define hospital/facility capacity schema.
  - Build hospital update workflow.
  - Add dispatcher capacity map/list.
  - Add stale-data warnings.
  - Add fixture integration adapter.
- Estimate: 13 points.
- Dependencies: NADAA-020, NADAA-080.

#### NADAA-122: Relief Distribution Tracking

- Outcome: authorities can publish and manage relief distribution points, stock levels, eligibility notes, and schedules.
- Acceptance criteria:
  - Relief points have location, operating hours, contact, stock categories, and eligibility guidance.
  - Citizens can view nearby relief distribution points.
  - Authorities can update stock and status.
- Tasks:
  - Add relief point schema and API.
  - Build authority management view.
  - Build citizen relief point map/list.
  - Add stock/status history.
  - Add tests for nearby relief lookup.
- Estimate: 8 points.
- Dependencies: NADAA-062.

#### NADAA-123: Donation And Aid Coordination

- Outcome: agencies can coordinate donations and aid requests without overwhelming incident operations.
- Acceptance criteria:
  - Agencies can create aid requests by category, location, priority, and receiving organization.
  - Donors can view approved needs and pledge support.
  - Pledges are tracked separately from emergency incident status.
- Tasks:
  - Define aid request and pledge models.
  - Build authority approval flow.
  - Build public/partner aid listing.
  - Add reporting export.
  - Add anti-fraud review notes.
- Estimate: 8 points.
- Dependencies: NADAA-122, NADAA-012.

### EPIC 13: Mobility, Recovery, And Case Management

Goal: guide safer movement during emergencies and help affected people recover documents, property evidence, and family connections.

#### NADAA-130: Evacuation Route Planner

- Outcome: citizens and responders can see safer evacuation routes based on hazards, road closures, shelters, and terrain.
- Acceptance criteria:
  - Route planner avoids known closed roads and severe hazard zones.
  - Routes can target shelters or higher-ground waypoints.
  - The UI explains that route guidance may change and emergency instructions take priority.
- Tasks:
  - Select routing engine or provider.
  - Import road network and hazard avoidance layers.
  - Build route API.
  - Add citizen and authority map views.
  - Add route QA scenarios.
- Estimate: 13 points.
- Dependencies: NADAA-021, NADAA-062, NADAA-131.

#### NADAA-131: Road Closure Integration

- Outcome: authorities can add, import, and publish road closure information for maps, alerts, and route planning.
- Acceptance criteria:
  - Road closures include geometry, reason, severity, source, start time, expected expiry, and status.
  - Road closures can be created manually or imported from integration adapters.
  - Road closures appear in citizen and authority map layers.
- Tasks:
  - Define road closure schema.
  - Build road closure management endpoints.
  - Add map layer rendering.
  - Add integration adapter contract.
  - Add tests for active/expired closures.
- Estimate: 8 points.
- Dependencies: NADAA-080, NADAA-040.

#### NADAA-132: Missing Persons Module

- Outcome: authorities can manage missing persons reports linked to disasters while protecting sensitive personal data.
- Acceptance criteria:
  - Missing persons records support reporter details, last seen location/time, description, photo, related incident, and status.
  - Public visibility is controlled by authority approval.
  - Reunification or closure records are audited.
- Tasks:
  - Define missing persons schema.
  - Build intake and authority review flow.
  - Add public approved listing/search.
  - Add privacy and consent controls.
  - Add audit and closure workflow.
- Estimate: 13 points.
- Dependencies: NADAA-030, NADAA-090.

#### NADAA-133: Insurance And Property Damage Claim Export

- Outcome: affected citizens can export verified incident and damage information for insurance, district assembly, or relief processes.
- Acceptance criteria:
  - Citizens can submit property damage details linked to an incident.
  - Authorities can verify damage reports where required.
  - Export includes incident reference, location, damage notes, media references, and verification status.
- Tasks:
  - Define damage report schema.
  - Build citizen damage report flow.
  - Build authority verification view.
  - Generate PDF/CSV export.
  - Add privacy and retention documentation.
- Estimate: 8 points.
- Dependencies: NADAA-030, NADAA-041, NADAA-031.

### EPIC 14: Remote Sensing And Evidence Ingestion

Goal: bring drone and satellite evidence into incident verification, flood mapping, and recovery analysis.

#### NADAA-140: Drone And Satellite Image Ingestion

- Outcome: analysts can upload or import drone/satellite imagery and link it to incidents, risk zones, or ML workflows.
- Acceptance criteria:
  - Imagery metadata includes source, capture time, geometry/coverage, resolution, license, and related incident/risk zone.
  - Imagery can be reviewed in authority dashboard map layers.
  - Large files are stored outside the primary database.
- Tasks:
  - Define imagery metadata schema.
  - Build upload/import workflow.
  - Add map overlay viewer.
  - Add storage lifecycle rules.
  - Document licensing and privacy constraints.
- Estimate: 13 points.
- Dependencies: NADAA-031, NADAA-040, NADAA-070.

### Phase 2 Sprint Plan

#### Sprint 8: Inclusive Channel Foundation

- NADAA-110 SMS/USSD Emergency Access.
- NADAA-111 WhatsApp Emergency Chatbot, first pass.

Deliverables:

- SMS/USSD provider abstraction.
- USSD menu and inbound report flow.
- WhatsApp conversation skeleton.

#### Sprint 9: Multilingual Alerts And Chat Completion

- NADAA-111 WhatsApp Emergency Chatbot, completion.
- NADAA-112 Multilingual Voice Alerts.

Deliverables:

- WhatsApp incident/report/shelter flows.
- Voice alert review and delivery workflow.
- Multilingual alert template library.

#### Sprint 10: Field And Health Operations

- NADAA-120 Community Volunteer App.
- NADAA-121 Hospital Capacity Tracker.

Deliverables:

- Volunteer registration, verification, and assignment.
- Hospital capacity map/list and update workflow.

#### Sprint 11: Relief, Aid, And Mobility Data

- NADAA-122 Relief Distribution Tracking.
- NADAA-123 Donation And Aid Coordination.
- NADAA-131 Road Closure Integration.

Deliverables:

- Relief point management.
- Aid request and pledge tracking.
- Road closure map layer and management workflow.

#### Sprint 12: Recovery, Routing, And Remote Sensing

- NADAA-130 Evacuation Route Planner.
- NADAA-132 Missing Persons Module.
- NADAA-133 Insurance And Property Damage Claim Export.
- NADAA-140 Drone And Satellite Image Ingestion.

Deliverables:

- Evacuation route planner.
- Missing persons case workflow.
- Damage claim export.
- Drone/satellite imagery ingestion and map overlay.

## Phase 3 Detailed Plan

Phase 3 moves the platform toward national-scale disaster intelligence, advanced AI support, school/community preparedness, open data, and telecom-grade public warning.

### EPIC 15: Advanced AI, Simulation, And Resource Forecasting

Goal: improve prediction, triage, verification, and resource placement while preserving human oversight for public safety actions.

#### NADAA-150: Real-Time Flood Simulation

- Outcome: authorities can view flood scenario simulations using rainfall, hydrology, elevation, drainage, and terrain inputs.
- Acceptance criteria:
  - Simulation outputs include affected areas, expected depth/severity bands where available, timestamps, assumptions, and confidence.
  - Simulations are labeled as decision support, not automatic alert instructions.
  - Results can feed alert drafting only through approval workflow.
- Tasks:
  - Select simulation approach and required datasets.
  - Build simulation job runner.
  - Store versioned simulation outputs.
  - Render simulation map layers.
  - Add validation and limitations report.
- Estimate: 21 points.
- Dependencies: NADAA-070, NADAA-072, NADAA-081, NADAA-140.

#### NADAA-151: AI Incident Triage

- Outcome: dispatchers receive triage suggestions for severity, duplicate likelihood, affected population, and suggested agency routing.
- Acceptance criteria:
  - AI suggestions are explainable and editable by dispatchers.
  - Suggestions do not automatically verify, close, or assign incidents without human action.
  - All model suggestions and overrides are logged for review.
- Tasks:
  - Define triage labels and training/evaluation data.
  - Build triage model or rules-plus-model service.
  - Add dispatcher suggestion UI.
  - Add override logging.
  - Add bias/error review process.
- Estimate: 13 points.
- Dependencies: NADAA-033, NADAA-041, NADAA-091.

#### NADAA-152: Computer Vision For Flood And Fire Image Verification

- Outcome: uploaded images can be analyzed for flood/fire evidence to support, not replace, human verification.
- Acceptance criteria:
  - CV output includes evidence labels, confidence, model version, and limitations.
  - Low-confidence or sensitive images are routed to human review.
  - CV results are not exposed publicly by default.
- Tasks:
  - Define image labeling taxonomy.
  - Build model evaluation dataset.
  - Add CV inference endpoint.
  - Add evidence panel in authority dashboard.
  - Add privacy and retention review.
- Estimate: 13 points.
- Dependencies: NADAA-031, NADAA-040, NADAA-140.

#### NADAA-153: Predictive Ambulance And Fire Station Positioning

- Outcome: agencies can see demand forecasts and suggested staging positions before high-risk periods.
- Acceptance criteria:
  - Forecasts consider historical incidents, weather/risk predictions, response times, and agency capacity.
  - Suggested positions include confidence and operational constraints.
  - Final deployment decisions remain with agency leadership.
- Tasks:
  - Define resource demand features.
  - Build forecasting baseline.
  - Add map view for suggested staging.
  - Add scenario comparison.
  - Add performance monitoring.
- Estimate: 13 points.
- Dependencies: NADAA-042, NADAA-071, NADAA-121.

### EPIC 16: Preparedness, Education, And Public Data

Goal: support proactive disaster readiness across schools, communities, media campaigns, researchers, and policy stakeholders.

#### NADAA-160: School Emergency Preparedness Module

- Outcome: schools can manage emergency plans, drills, contacts, risk checks, and preparedness resources.
- Acceptance criteria:
  - School profiles include location, student population, emergency contacts, hazards, evacuation points, and drill history.
  - District officers can view school readiness by district.
  - School guidance is age-appropriate and multilingual-ready.
- Tasks:
  - Define school profile schema.
  - Build school admin workflow.
  - Add district readiness dashboard.
  - Add drill tracking.
  - Add school guidance content.
- Estimate: 13 points.
- Dependencies: NADAA-060, NADAA-062, NADAA-021.

#### NADAA-161: Public Disaster Education Campaigns

- Outcome: agencies can publish preparedness campaigns tied to hazards, seasons, regions, and population groups.
- Acceptance criteria:
  - Campaigns support articles, checklists, media, languages, target regions, and publishing windows.
  - Campaign metrics include reach and engagement.
  - Campaign content can link to emergency guides and alerts.
- Tasks:
  - Define campaign content model.
  - Build publishing workflow.
  - Add public campaign pages.
  - Add analytics events.
  - Add seasonal campaign templates.
- Estimate: 8 points.
- Dependencies: NADAA-060, NADAA-052.

#### NADAA-162: National Open Disaster Data Portal

- Outcome: approved, anonymized disaster data can be shared for public awareness, research, planning, and accountability.
- Acceptance criteria:
  - Portal exposes approved datasets with metadata, license, update frequency, and privacy review status.
  - Data is anonymized and aggregated where needed.
  - Downloads and APIs are rate-limited and audited.
- Tasks:
  - Define open data governance policy.
  - Build dataset catalog.
  - Add export/API layer.
  - Add privacy/anonymization jobs.
  - Add documentation for researchers and public users.
- Estimate: 13 points.
- Dependencies: NADAA-012, NADAA-090, NADAA-100.

### EPIC 17: National-Scale Alerting And Operations

Goal: prepare the platform for national telecom-grade warning, resilience, observability, and operational scale.

#### NADAA-163: Telecom Cell Broadcast Integration

- Outcome: approved emergency alerts can be routed to telecom cell broadcast channels when official agreements and infrastructure are available.
- Acceptance criteria:
  - Cell broadcast adapter is isolated behind the notification service.
  - Alert templates meet telecom and national emergency requirements.
  - Emergency override and approval controls remain enforced.
  - End-to-end tests can run against simulator/sandbox environments.
- Tasks:
  - Define telecom integration requirements.
  - Build adapter interface and sandbox simulator.
  - Add cell broadcast alert preview.
  - Add delivery/audit logs.
  - Prepare compliance and operational runbook.
- Estimate: 21 points.
- Dependencies: NADAA-050, NADAA-051, NADAA-052, NADAA-092.

#### NADAA-170: National-Scale Resilience And Operations Hardening

- Outcome: the platform is ready for high-volume national events with clear incident management and support operations.
- Acceptance criteria:
  - Load tests cover report spikes, alert delivery, dashboard maps, and notification queues.
  - Observability dashboards cover latency, errors, queue depth, provider delivery, and geospatial query performance.
  - Disaster recovery and failover runbooks exist.
- Tasks:
  - Define national event load profiles.
  - Run load and queue tests.
  - Add observability dashboards.
  - Create incident response runbooks.
  - Validate backup and restore.
- Estimate: 21 points.
- Dependencies: NADAA-101, NADAA-163.

### Phase 3 Sprint Plan

#### Sprint 13: Simulation And AI Triage

- NADAA-150 Real-Time Flood Simulation, first pass.
- NADAA-151 AI Incident Triage.

Deliverables:

- Simulation architecture and first map output.
- Dispatcher AI triage suggestions with override logging.

#### Sprint 14: Vision And Predictive Resource Planning

- NADAA-150 Real-Time Flood Simulation, completion.
- NADAA-152 Computer Vision For Flood And Fire Image Verification.
- NADAA-153 Predictive Ambulance And Fire Station Positioning.

Deliverables:

- Versioned flood simulation layers.
- CV evidence panel.
- Resource demand/staging forecast view.

#### Sprint 15: Preparedness And Open Data

- NADAA-160 School Emergency Preparedness Module.
- NADAA-161 Public Disaster Education Campaigns.
- NADAA-162 National Open Disaster Data Portal, first pass.

Deliverables:

- School readiness workflows.
- Campaign publishing.
- Open data governance and catalog foundation.

#### Sprint 16: National Alerting And Operational Scale

- NADAA-162 National Open Disaster Data Portal, completion.
- NADAA-163 Telecom Cell Broadcast Integration.
- NADAA-170 National-Scale Resilience And Operations Hardening.

Deliverables:

- Public open data API/export layer.
- Cell broadcast adapter and simulator.
- Load testing, observability, backup/restore, and operations runbooks.

## Master Story Tracker

Use this table for cross-agent status tracking. Keep the `Active Work Board` focused on current work, and update this master tracker when stories move between `Todo`, `In Progress`, `Blocked`, `Review`, and `Done`.

### MVP Tracker

| ID        | Phase | Sprint     | Story                                                       | Status | Owner      | Branch/PR | Notes                                                                                                                                                                                                        |
| --------- | ----- | ---------- | ----------------------------------------------------------- | ------ | ---------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| NADAA-001 | MVP   | Sprint 0   | Create Repository And Monorepo Foundation                   | Done   | Codex      | main      | Monorepo foundation, starter apps, risk service, infra, and docs created.                                                                                                                                    |
| NADAA-002 | MVP   | Sprint 0   | Define Product, API, Security, ML, And Deployment Docs      | Done   | Codex      | main      | Docs expanded and linked from README.                                                                                                                                                                        |
| NADAA-003 | MVP   | Sprint 0   | Create Delivery Dashboard Data Contract                     | Done   | Codex      | main      | Schema, sample records, and validation script added.                                                                                                                                                         |
| NADAA-010 | MVP   | Sprint 1   | Implement Citizen Authentication                            | Done   | Codex      | main      | Citizen register/login/profile API and tests added in auth-service.                                                                                                                                          |
| NADAA-011 | MVP   | Sprint 1   | Implement Agency Users, Roles, And MFA                      | Done   | Codex      | main      | Agency user creation, authority role catalog, mock MFA setup/verification, agency login, MFA-aware tokens, shared types, docs, and RBAC tests added.                                                         |
| NADAA-012 | MVP   | Sprint 1   | Implement Audit Logging Foundation                          | Done   | Codex      | main      | Auth-service audit model, helper, auth/admin event wiring, system-admin audit read endpoint, shared types, retention docs, and tests added.                                                                  |
| NADAA-020 | MVP   | Sprint 1   | Set Up PostGIS And Core Geospatial Models                   | Done   | Codex      | main      | Migration and seed verified against local PostGIS on port 55432.                                                                                                                                             |
| NADAA-021 | MVP   | Sprint 5   | Implement Area Risk API                                     | Done   | Codex      | main      | Fixture-backed API returns low/high/severe flood scoring, nearby shelters, nearby facilities, recommended actions, and coordinate validation.                                                                |
| NADAA-022 | MVP   | Sprint 5   | Build Citizen Risk Checker UI                               | Done   | Codex      | main      | Risk checker UI integrates the risk API, area presets, coordinate entry, GPS lookup, shelters, facilities, recommended actions, loading/error/permission/empty states, and risk smoke coverage.              |
| NADAA-030 | MVP   | Sprint 2   | Implement Incident Reporting API                            | Done   | Codex      | main      | Incident intake API and tests added in incident-service.                                                                                                                                                     |
| NADAA-031 | MVP   | Sprint 2   | Implement Media Upload Flow                                 | Done   | Codex      | main      | Controlled upload initiation and private media linkage added in incident-service.                                                                                                                            |
| NADAA-032 | MVP   | Sprint 2   | Build Citizen Incident Reporting UI                         | Done   | Codex      | main      | Citizen incident form integrated with media upload initiation, GPS/manual location, validation, offline retry messaging, and API success/error states.                                                       |
| NADAA-033 | MVP   | Sprint 2   | Add Incident Deduplication Baseline                         | Done   | Codex      | main      | Same-hazard duplicate candidates are scored by location distance, report time, and description similarity; reports remain reviewable and are never automatically merged or deleted.                          |
| NADAA-040 | MVP   | Sprint 3   | Build Incident Command Map                                  | Done   | Codex      | main      | Leaflet command map, API-backed incident feed with fixture fallback, map/list sync, filters, selected-incident detail, role-protected framing, docs, and UI states added.                                    |
| NADAA-041 | MVP   | Sprint 3   | Implement Verification And Status Workflow                  | Todo   | Unassigned | TBD       | Depends on NADAA-040, NADAA-012.                                                                                                                                                                             |
| NADAA-042 | MVP   | Sprint 3   | Implement Agency Assignment And Incident Timeline           | Todo   | Unassigned | TBD       | Depends on NADAA-041.                                                                                                                                                                                        |
| NADAA-043 | MVP   | Sprint 3   | Implement Duplicate Merge Review                            | Todo   | Unassigned | TBD       | Depends on NADAA-033, NADAA-041.                                                                                                                                                                             |
| NADAA-050 | MVP   | Sprint 4   | Implement Alert Creation And Approval Workflow              | Todo   | Unassigned | TBD       | Depends on NADAA-011, NADAA-012, NADAA-020.                                                                                                                                                                  |
| NADAA-051 | MVP   | Sprint 4   | Implement Geofenced Alert Targeting                         | Todo   | Unassigned | TBD       | Depends on NADAA-050.                                                                                                                                                                                        |
| NADAA-052 | MVP   | Sprint 4   | Implement In-App Alert Feed And Push/SMS Abstraction        | Todo   | Unassigned | TBD       | Depends on NADAA-050, NADAA-051.                                                                                                                                                                             |
| NADAA-060 | MVP   | Sprint 5   | Implement Emergency Guide Content Model                     | Done   | Codex      | main      | Guide-service API, guide content fixtures, seed expansion, shared guide types, docs, Docker/CI wiring, staging smoke wiring, and lookup tests added.                                                         |
| NADAA-061 | MVP   | Sprint 5   | Build Offline-First Citizen Guidance UI                     | Todo   | Unassigned | TBD       | Depends on NADAA-060.                                                                                                                                                                                        |
| NADAA-062 | MVP   | Sprint 5   | Implement Shelter And Recovery Support Module               | Todo   | Unassigned | TBD       | Depends on NADAA-020, NADAA-022.                                                                                                                                                                             |
| NADAA-070 | MVP   | Sprint 6   | Create Flood Risk Dataset And Feature Pipeline              | Todo   | Unassigned | TBD       | Depends on NADAA-020, NADAA-002.                                                                                                                                                                             |
| NADAA-071 | MVP   | Sprint 6   | Train Baseline Flood Risk Model                             | Todo   | Unassigned | TBD       | Depends on NADAA-070.                                                                                                                                                                                        |
| NADAA-072 | MVP   | Sprint 6   | Serve ML Predictions Through Risk API                       | Todo   | Unassigned | TBD       | Depends on NADAA-071, NADAA-021.                                                                                                                                                                             |
| NADAA-073 | MVP   | Sprint 6   | Add Authority ML Review View                                | Todo   | Unassigned | TBD       | Depends on NADAA-072, NADAA-050.                                                                                                                                                                             |
| NADAA-080 | MVP   | Sprint 6   | Define Agency Integration Contracts                         | Done   | Codex      | main      | Integration matrix, inbound weather/hydrology contracts, outbound incident/alert sync contracts, mock integration service, shared types, docs, CI/staging wiring, smoke script, Dockerfile, and tests added. |
| NADAA-081 | MVP   | Sprint 6   | Implement Weather And Hydrology Import Skeleton             | Todo   | Unassigned | TBD       | Depends on NADAA-020, NADAA-080.                                                                                                                                                                             |
| NADAA-090 | MVP   | Sprint 2   | Implement Location Privacy And Anonymous Reporting Controls | Todo   | Unassigned | TBD       | Depends on NADAA-030, NADAA-040.                                                                                                                                                                             |
| NADAA-091 | MVP   | Sprint 3   | Implement Abuse, Spam, And False Report Handling            | Todo   | Unassigned | TBD       | Depends on NADAA-030, NADAA-041.                                                                                                                                                                             |
| NADAA-092 | MVP   | Sprint 7   | Security Review And Hardening                               | Todo   | Unassigned | TBD       | Depends on MVP feature completion.                                                                                                                                                                           |
| NADAA-100 | MVP   | Sprint 0/7 | Build Test Strategy And QA Matrix                           | Done   | Codex      | main      | Sprint 0 QA matrix complete; final pass remains in Sprint 7.                                                                                                                                                 |
| NADAA-101 | MVP   | Sprint 1/7 | Set Up CI/CD And Staging Environment                        | Done   | Codex      | main      | CI workflow, Docker build validation, manual staging smoke workflow/script, staging env template, and runbook added; registry push/deploy credentials remain environment-owned.                              |
| NADAA-102 | MVP   | Sprint 7   | Conduct UAT, Beta, And Production Readiness                 | Todo   | Unassigned | TBD       | Depends on NADAA-100, NADAA-101, MVP feature completion.                                                                                                                                                     |

### Phase 2 Tracker

| ID        | Phase   | Sprint     | Story                                      | Status | Owner      | Branch/PR | Notes                                  |
| --------- | ------- | ---------- | ------------------------------------------ | ------ | ---------- | --------- | -------------------------------------- |
| NADAA-110 | Phase 2 | Sprint 8   | SMS/USSD Emergency Access                  | Todo   | Unassigned | TBD       | Inclusive access channel.              |
| NADAA-111 | Phase 2 | Sprint 8/9 | WhatsApp Emergency Chatbot                 | Todo   | Unassigned | TBD       | Spans two sprints.                     |
| NADAA-112 | Phase 2 | Sprint 9   | Multilingual Voice Alerts                  | Todo   | Unassigned | TBD       | English, Twi, Ga, Ewe, Dagbani, Hausa. |
| NADAA-120 | Phase 2 | Sprint 10  | Community Volunteer App                    | Todo   | Unassigned | TBD       | Field coordination.                    |
| NADAA-121 | Phase 2 | Sprint 10  | Hospital Capacity Tracker                  | Todo   | Unassigned | TBD       | Facility operations.                   |
| NADAA-122 | Phase 2 | Sprint 11  | Relief Distribution Tracking               | Todo   | Unassigned | TBD       | Recovery logistics.                    |
| NADAA-123 | Phase 2 | Sprint 11  | Donation And Aid Coordination              | Todo   | Unassigned | TBD       | Partner/donor workflow.                |
| NADAA-130 | Phase 2 | Sprint 12  | Evacuation Route Planner                   | Todo   | Unassigned | TBD       | Depends on road closures and shelters. |
| NADAA-131 | Phase 2 | Sprint 11  | Road Closure Integration                   | Todo   | Unassigned | TBD       | Feeds route planner.                   |
| NADAA-132 | Phase 2 | Sprint 12  | Missing Persons Module                     | Todo   | Unassigned | TBD       | Sensitive data workflow.               |
| NADAA-133 | Phase 2 | Sprint 12  | Insurance And Property Damage Claim Export | Todo   | Unassigned | TBD       | Recovery case export.                  |
| NADAA-140 | Phase 2 | Sprint 12  | Drone And Satellite Image Ingestion        | Todo   | Unassigned | TBD       | Supports verification and ML.          |

### Phase 3 Tracker

| ID        | Phase   | Sprint       | Story                                                 | Status | Owner      | Branch/PR | Notes                                       |
| --------- | ------- | ------------ | ----------------------------------------------------- | ------ | ---------- | --------- | ------------------------------------------- |
| NADAA-150 | Phase 3 | Sprint 13/14 | Real-Time Flood Simulation                            | Todo   | Unassigned | TBD       | Spans two sprints.                          |
| NADAA-151 | Phase 3 | Sprint 13    | AI Incident Triage                                    | Todo   | Unassigned | TBD       | Human-supervised only.                      |
| NADAA-152 | Phase 3 | Sprint 14    | Computer Vision For Flood And Fire Image Verification | Todo   | Unassigned | TBD       | Supports verification, not auto-escalation. |
| NADAA-153 | Phase 3 | Sprint 14    | Predictive Ambulance And Fire Station Positioning     | Todo   | Unassigned | TBD       | Agency decision support.                    |
| NADAA-160 | Phase 3 | Sprint 15    | School Emergency Preparedness Module                  | Todo   | Unassigned | TBD       | District readiness.                         |
| NADAA-161 | Phase 3 | Sprint 15    | Public Disaster Education Campaigns                   | Todo   | Unassigned | TBD       | Seasonal campaigns.                         |
| NADAA-162 | Phase 3 | Sprint 15/16 | National Open Disaster Data Portal                    | Todo   | Unassigned | TBD       | Requires privacy governance.                |
| NADAA-163 | Phase 3 | Sprint 16    | Telecom Cell Broadcast Integration                    | Todo   | Unassigned | TBD       | Depends on official telecom path.           |
| NADAA-170 | Phase 3 | Sprint 16    | National-Scale Resilience And Operations Hardening    | Todo   | Unassigned | TBD       | Load, observability, DR.                    |

## Initial Ready Queue

Start here:

1. NADAA-050 Implement Alert Creation And Approval Workflow.
2. NADAA-041 Implement Verification And Status Workflow.
3. NADAA-042 Implement Agency Assignment And Incident Timeline.
4. NADAA-061 Build Offline-First Citizen Guidance UI.
5. NADAA-081 Implement Weather And Hydrology Import Skeleton.

## Key Risks And Early Decisions

- Official agency data access may lag development. Use fixture/mock adapters early and isolate integration contracts.
- Flood model quality depends on data availability. Start with transparent rule-based scoring and baseline ML before advanced models.
- Public alert misuse is high risk. Require RBAC, MFA, approval workflow, audit logs, and emergency override controls.
- Citizen reporting can attract spam or false reports. Use rate limits, duplicate detection, suspicious report flags, and human review.
- Offline and low-literacy needs are important. Keep PWA/offline guidance in MVP and plan voice/USSD/WhatsApp for phase 2.
- Geospatial targeting must be correct. Prioritize PostGIS indexes, district boundaries, geometry validation, and map QA.

## Success Metrics

- Time from citizen report to authority verification.
- Time from verification to agency assignment.
- Alert delivery success rate.
- Number of citizens reached.
- Flood prediction accuracy and calibration.
- False alert rate.
- Response time reduction.
- Duplicate reports merged.
- Number of shelters/facilities discoverable.
- User trust and feedback score.

## Plan Ledger

| Date       | Update                                                                                                                                                                                                                                                                    | Owner | Status   |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----- | -------- |
| 2026-07-06 | Completed NADAA-040 with authority-dashboard Leaflet incident command map, API-backed incident feed with fixture fallback, map/list synchronization, hazard/district/severity/status/time filters, selected-incident detail, role-protected framing, docs, and UI states. | Codex | Complete |
| 2026-07-06 | Completed NADAA-012 with auth-service audit event model, in-memory audit store/helper, system-admin audit read endpoint, auth/admin event wiring, metadata capture, sanitized before/after snapshots, shared audit types, retention docs, and tests.                      | Codex | Complete |
| 2026-07-06 | Completed NADAA-011 with auth-service agency user creation, authority role catalog, mock MFA setup/verification, agency login, MFA-aware bearer tokens, shared agency auth types, API/security docs, and RBAC tests.                                                      | Codex | Complete |
| 2026-07-06 | Completed NADAA-022 with citizen risk checker API integration, area presets, coordinate entry, GPS lookup, shelters/facilities rendering, loading/error/permission/empty states, and risk smoke coverage.                                                                 | Codex | Complete |
| 2026-07-06 | Completed NADAA-021 with fixture-backed risk-service area lookup, baseline flood scoring, nearby shelter/facility aggregation, coordinate validation, shared response types, API docs, and tests.                                                                         | Codex | Complete |
| 2026-07-06 | Completed NADAA-101 with GitHub Actions CI, Docker build validation for deployable apps/services, manual staging smoke workflow, staging smoke script, staging environment template, and staging runbook.                                                                 | Codex | Complete |
| 2026-07-06 | Completed NADAA-031 with controlled media upload initiation, private metadata, content-type and size validation, incident media linkage, shared media types, API docs, and tests.                                                                                         | Codex | Complete |
| 2026-07-06 | Completed NADAA-030 with incident-service report intake API, validation, anonymous/contact-permission behavior, media references, priority review flag, rate limiting, shared types, API docs, and tests.                                                                 | Codex | Complete |
| 2026-07-06 | Completed NADAA-010 with auth-service citizen registration, mock OTP login, signed bearer token profile access, shared auth types, API docs, and tests.                                                                                                                   | Codex | Complete |
| 2026-07-06 | Completed NADAA-100 and NADAA-020 with QA strategy, MVP test matrix, smoke script, PostGIS schema, geospatial indexes, seed data, database docs, configurable local ports, and database validation.                                                                       | Codex | Complete |
| 2026-07-06 | Completed NADAA-002 and NADAA-003 with expanded product/API/security/ML/deployment docs, dashboard tracking schema, sample records, and validation script.                                                                                                                | Codex | Complete |
| 2026-07-06 | Completed NADAA-001 initial monorepo foundation with React/Vite citizen and authority apps, shared brand/types packages, Go risk service, local infra, docs, dependency lockfile, and Git remote setup.                                                                   | Codex | Complete |
| 2026-07-06 | Expanded Phase 2 and Phase 3 into detailed epics, stories, tasks, sprint plans, and master status trackers.                                                                                                                                                               | Codex | Complete |
| 2026-07-06 | Created initial agent plan from repository documents.                                                                                                                                                                                                                     | Codex | Complete |
