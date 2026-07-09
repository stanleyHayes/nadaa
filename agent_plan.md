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
  - Mobile: React Native after the web/PWA workflows prove the API contracts, with citizen and dispatcher mobile apps planned as separate products.
  - Backend: Golang, hexagonal architecture, REST + WebSocket.
  - Data: PostgreSQL/PostGIS, MongoDB for flexible media/report metadata if needed, Redis for cache/queues/rate limits.
  - ML: Python/FastAPI, MLflow, Airflow or Prefect later.
  - Infra: Docker first, then Kubernetes/Swarm when deployment maturity requires it.

## Target Platform Portfolio

The build must converge on role-specific products rather than one oversized dashboard. Each platform should share design tokens, auth contracts, API clients, and domain types through `packages/ui`, `packages/config`, and `packages/shared-types`, but keep app shells, routing, and feature composition separate.

| Audience/Role       | Platform | Target App Path          | Primary Phase | Primary Jobs                                                                                         | Build Notes                                                                                      |
| ------------------- | -------- | ------------------------ | ------------- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| Public stakeholders | Web      | `apps/marketing-web`     | MVP hardening | About, platform summary, services, benefits, contact, and stakeholder conversion                     | Public marketing site should explain NADAA clearly without mixing in citizen self-service flows. |
| Citizens            | Web/PWA  | `apps/citizen-web`       | MVP           | Risk checks, alerts, incident reports, media upload, emergency guides, shelters, recovery support    | Current citizen web app is the MVP public surface and should stay mobile-first/PWA-ready.        |
| Citizens            | Mobile   | `apps/citizen-mobile`    | Phase 2       | Push alerts, offline guides, GPS reporting, shelter/recovery support, risk checks                    | React Native app should reuse citizen contracts proven by `apps/citizen-web`.                    |
| Dispatchers         | Web      | `apps/dispatcher-web`    | MVP hardening | Incident command map, verification, assignment, timelines, duplicate review, abuse review, ML review | Current `apps/authority-dashboard` command-center work should be split into this dedicated app.  |
| Dispatchers         | Mobile   | `apps/dispatcher-mobile` | Phase 2       | Triage queue, critical incident review, assignment/status updates, push escalation, offline refresh  | React Native app should focus on field/shift use, not full admin configuration.                  |
| Agency users        | Web      | `apps/agency-web`        | Phase 2       | Assigned incidents, responder updates, agency notes, shelter/capacity updates, relief/road context   | Agency users should see agency-scoped work only, with RBAC and audit coverage.                   |
| System/admin users  | Web      | `apps/admin-web`         | MVP hardening | Agencies, users, roles, MFA support, audit logs, data sources, alert rules, platform configuration   | Admin work must be isolated from dispatcher operations and protected by stricter roles.          |

### Platform Build Board

| Platform          | App Path                 | Status | Owner | Phase/Sprint                   | Driving Stories                                                             | Notes                                                                                                                                                                                                                                                                           |
| ----------------- | ------------------------ | ------ | ----- | ------------------------------ | --------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Marketing web     | `apps/marketing-web`     | Done   | Codex | MVP Sprint 7/Phase 2 Sprint 11 | NADAA-015                                                                   | Public NADAA website completed with about, features, platform lanes, services, benefits, emergency contact, partner contact, research-backed Ghana disaster context, smoke/staging checks, CI Docker entry, docs, dashboard record, and verification.                           |
| Citizen web       | `apps/citizen-web`       | Done   | Codex | MVP Sprints 2, 4, 5            | NADAA-022, NADAA-032, NADAA-052, NADAA-061, NADAA-062                       | MVP citizen web flows now cover risk, reporting, alerts, offline guides, shelter lookup, and recovery support.                                                                                                                                                                  |
| Citizen mobile    | `apps/citizen-mobile`    | Done   | Codex | Phase 2 Sprints 8/10           | NADAA-113, NADAA-120                                                        | Expo/React Native foundation plus Community volunteer tab added with shared contracts, offline primitives, permission copy, smoke checks, and docs.                                                                                                                             |
| Dispatcher web    | `apps/dispatcher-web`    | Done   | Codex | MVP Sprints 3/7                | NADAA-040, NADAA-041, NADAA-042, NADAA-043, NADAA-044, NADAA-073, NADAA-091 | Dedicated dispatcher command app scaffolded, wired into workspace scripts, CI Docker matrix, local/staging smoke checks, and docs.                                                                                                                                              |
| Dispatcher mobile | `apps/dispatcher-mobile` | Done   | Codex | Phase 2 Sprint 10              | NADAA-124                                                                   | Expo/RN triage app with queue, detail, action, capacity, profile tabs; offline/stale indicators and sandbox push. Verified typecheck, build, smoke, lint, docs, dashboard.                                                                                                      |
| Agency web        | `apps/agency-web`        | Done   | Codex | Phase 2 Sprints 10/11          | NADAA-042, NADAA-062, NADAA-121, NADAA-122, NADAA-125                       | `apps/agency-web` React/Vite operations portal with agency session/MFA shell, assigned-incident dashboard, responder status/timeline-note updates, shelter/hospital capacity context, occupancy/capacity update forms, smoke/CI/Docker/docs/dashboard wiring, and verification. |
| Admin web         | `apps/admin-web`         | Done   | Codex | MVP Sprint 7                   | NADAA-014                                                                   | Dedicated governance console scaffolded with admin RBAC shell, agency/user/role, MFA, audit, data-source, alert-rule, scripts, docs, CI, smoke, and Docker wiring.                                                                                                              |

### Platform Delivery Breakdown

Use this table when deciding where a feature belongs. Do not add dispatcher, agency, or admin-only workflows to the citizen apps; do not add system administration to dispatcher or agency apps.

| Platform                 | Product Scope                                                                                                                                               | Must Not Own                                                                                                 | Shared Dependencies                                                                                                                | First Build Story                           | Release Gate                                                                                                                       |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `apps/marketing-web`     | Public information and conversion website for platform overview, about, features, benefits, services, platform lanes, research context, contact, and demos. | Live incident reporting, protected dashboards, admin actions, operational incident data, private user data.  | Shared NADAA brand, public logo/assets, platform copy, product docs, smoke/deployment scripts.                                     | NADAA-015.                                  | Build/typecheck passes, smoke confirms title/content, contact CTAs and emergency 112 guidance are visible, public copy is sourced. |
| `apps/citizen-web`       | Public PWA for alerts, risk checks, reporting, media, guides, shelters, recovery support, consent, and privacy copy.                                        | Dispatcher verification, agency assignment, admin configuration, private audit logs.                         | Auth citizen session, incident API, risk API, alert/feed API, guide API, shelter/recovery API, shared brand/UI.                    | NADAA-062 after existing citizen MVP flows. | Mobile-first web QA, offline guide cache, report submission smoke, alert feed smoke, privacy copy reviewed.                        |
| `apps/citizen-mobile`    | Native citizen app for push alerts, GPS/media reporting, risk checks, offline guides, shelters, recovery support, and report drafts.                        | Authority review, agency responder tools, system configuration.                                              | Stable citizen web contracts, notification push abstraction, shared types/config/theme, mobile permission copy.                    | NADAA-113.                                  | iOS/Android build path documented, location/media/push permission flows tested, offline report draft and guide cache tested.       |
| `apps/dispatcher-web`    | Operational command console for report intake, map/list triage, verification, status, assignment, duplicate review, abuse review, timelines, and ML review. | System-wide user/role management, data-source administration, citizen self-service pages.                    | Incident workflow APIs, assignment/timeline APIs, alert draft link, ML review API, dispatcher RBAC/MFA, shared command components. | NADAA-044.                                  | RBAC/MFA enforced, no admin settings exposed, incident workflow regression smoke passes, map/list/detail flows verified.           |
| `apps/dispatcher-mobile` | Shift-friendly mobile triage for urgent queue review, selected incident details, status updates, assignment handoff, timeline notes, and push escalation.   | Full map administration, bulk configuration, public citizen features.                                        | Dispatcher web contracts, incident workflow APIs, push abstraction, shared auth/config/types.                                      | NADAA-124.                                  | MFA/session handling verified, stale/offline indicators visible, critical push flow tested, status/assignment action smoke passes. |
| `apps/agency-web`        | Agency-scoped portal for assigned incidents, responder updates, operational notes, shelter/capacity updates, relief context, and road/hospital context.     | Cross-agency dispatch control unless role-granted, system administration, public citizen content management. | Agency RBAC, assignment filtering, shelter/recovery APIs, hospital capacity APIs, relief and road closure APIs, audit logging.     | NADAA-125.                                  | Agency scoping tests pass, responder updates audited, unauthorized cross-agency data is denied, capacity/shelter updates verified. |
| `apps/admin-web`         | Governance console for agencies, users, roles, MFA support, audit logs, data sources, alert rules, environment-safe configuration, and platform health.     | Live dispatcher triage, citizen reporting, agency field updates.                                             | Auth admin APIs, audit APIs, integration/data-source APIs, alert-rule APIs, MFA/session enforcement.                               | NADAA-014.                                  | System-admin RBAC/MFA enforced, admin actions audited, no secrets exposed, no unnecessary citizen report identity exposed.         |

### Platform Build Sequencing

1. Keep `apps/marketing-web` as the public explainer/conversion surface so stakeholders can understand NADAA without entering citizen, dispatcher, agency, or admin apps.
2. Keep `apps/citizen-web` as the citizen MVP surface while completing shelters, recovery, privacy, and ML/risk integrations.
3. Build `apps/dispatcher-web` next by extracting operational command features from `apps/authority-dashboard`; keep `authority-dashboard` only as a compatibility shell until scripts/docs move.
4. Build `apps/admin-web` as a separate governance console before adding more user, role, audit, data-source, or alert-rule administration.
5. Build `apps/agency-web` after assigned-incident filtering, shelter/recovery, and hospital/capacity APIs are stable enough for agency-scoped workflows.
6. Build `apps/citizen-mobile` from proven citizen web contracts, with extra attention to offline drafts, push registration, and location/media permission UX.
7. Build `apps/dispatcher-mobile` from proven dispatcher web contracts, limited to urgent triage, status, assignment, and timeline actions.
8. Add each new app to workspace scripts, CI, staging smoke checks, deployment docs, dashboard records, and this plan before marking the platform story `Done`.

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

## Frontend Modularity Rules

- Keep each web app root `src/App.tsx` as a thin entrypoint that delegates to a feature-level app component.
- Put app-wide configuration, theme, and session helpers under `src/app/`.
- Put domain-specific data, types, utilities, and components under `src/features/<feature>/`.
- Do not add new screens, API orchestration, fixtures, or large JSX surfaces directly to root app files.
- If a feature grows beyond one focused concern, split it into `data.ts`, `types.ts`, `utils.ts`, and focused `*.tsx` components before adding more behavior.

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

| ID        | Phase/Sprint                   | Work Item                                                   | Status | Owner | Branch/PR                                                  | Dependencies                                                  | Last Update | Notes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --------- | ------------------------------ | ----------------------------------------------------------- | ------ | ----- | ---------------------------------------------------------- | ------------------------------------------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| NADAA-001 | MVP Sprint 0                   | Create Repository And Monorepo Foundation                   | Done   | Codex | main                                                       | None                                                          | 2026-07-06  | Monorepo foundation, brand assets, starter apps, risk service, infra, docs, and Git remote are in place.                                                                                                                                                                                                                                                                                                                                                                                             |
| NADAA-002 | MVP Sprint 0                   | Define Product, API, Security, ML, And Deployment Docs      | Done   | Codex | main                                                       | NADAA-001                                                     | 2026-07-06  | Product, architecture, API, security, ML, deployment, data ownership, and integration assumptions documented.                                                                                                                                                                                                                                                                                                                                                                                        |
| NADAA-003 | MVP Sprint 0                   | Create Delivery Dashboard Data Contract                     | Done   | Codex | main                                                       | NADAA-001                                                     | 2026-07-06  | Dashboard schema, sample records, README, and validation script added under `docs/project-dashboard/`.                                                                                                                                                                                                                                                                                                                                                                                               |
| NADAA-004 | MVP Sprint 0                   | Define Multi-Platform App Portfolio And Build Lanes         | Done   | Codex | main                                                       | NADAA-001                                                     | 2026-07-06  | Planned citizen web/mobile, dispatcher web/mobile, agency web, and admin web lanes with target paths, phase placement, driving stories, ownership boundaries, shared dependencies, build sequence, and release gates.                                                                                                                                                                                                                                                                                |
| NADAA-100 | MVP Sprint 0                   | Build Test Strategy And QA Matrix                           | Done   | Codex | main                                                       | NADAA-002                                                     | 2026-07-06  | QA strategy, MVP test matrix, release checklist, UAT outline, severity model, and web smoke script added.                                                                                                                                                                                                                                                                                                                                                                                            |
| NADAA-020 | MVP Sprint 1                   | Set Up PostGIS And Core Geospatial Models                   | Done   | Codex | main                                                       | NADAA-001                                                     | 2026-07-06  | Core PostGIS schema, geospatial indexes, seed data, database docs, configurable compose ports, and asset validation added.                                                                                                                                                                                                                                                                                                                                                                           |
| NADAA-010 | MVP Sprint 1                   | Implement Citizen Authentication                            | Done   | Codex | main                                                       | NADAA-001                                                     | 2026-07-06  | Auth service citizen register/login/profile API, mock OTP flow, signed bearer token, shared auth types, docs, and tests added.                                                                                                                                                                                                                                                                                                                                                                       |
| NADAA-011 | MVP Sprint 1                   | Implement Agency Users, Roles, And MFA                      | Done   | Codex | main                                                       | NADAA-010                                                     | 2026-07-06  | Auth-service now supports agency user creation, authority role catalog, mock MFA setup/verification, agency login, MFA-aware tokens, and RBAC denial tests.                                                                                                                                                                                                                                                                                                                                          |
| NADAA-012 | MVP Sprint 1                   | Implement Audit Logging Foundation                          | Done   | Codex | main                                                       | NADAA-011                                                     | 2026-07-06  | Auth-service now records audit events for citizen/agency auth, admin user creation, MFA setup/verify, RBAC denial, and system-admin audit reads with metadata and sanitized snapshots.                                                                                                                                                                                                                                                                                                               |
| NADAA-030 | MVP Sprint 2                   | Implement Incident Reporting API                            | Done   | Codex | main                                                       | NADAA-020, NADAA-010                                          | 2026-07-06  | Incident-service report intake API, validation, anonymous/contact-permission behavior, media references, priority review flag, rate limiting, shared types, docs, and tests added.                                                                                                                                                                                                                                                                                                                   |
| NADAA-031 | MVP Sprint 2                   | Implement Media Upload Flow                                 | Done   | Codex | main                                                       | NADAA-030                                                     | 2026-07-06  | Controlled media upload initiation, private metadata, content-type and size validation, incident media linkage, shared types, docs, and tests added.                                                                                                                                                                                                                                                                                                                                                 |
| NADAA-032 | MVP Sprint 2                   | Build Citizen Incident Reporting UI                         | Done   | Codex | main                                                       | NADAA-030, NADAA-031                                          | 2026-07-06  | Citizen report form now supports GPS/manual coordinates, hazard, urgency, people affected, injuries, anonymous/contact controls, accessibility needs, media initiation, offline retry messaging, and success/error states.                                                                                                                                                                                                                                                                           |
| NADAA-033 | MVP Sprint 2                   | Add Incident Deduplication Baseline                         | Done   | Codex | main                                                       | NADAA-030                                                     | 2026-07-06  | Incident-service now stores same-hazard duplicate candidates scored by distance, time window, and description similarity without merging or deleting reports.                                                                                                                                                                                                                                                                                                                                        |
| NADAA-040 | MVP Sprint 3                   | Build Incident Command Map                                  | Done   | Codex | main                                                       | NADAA-011, NADAA-030                                          | 2026-07-06  | Authority dashboard now has a Leaflet incident command map, API-backed incident feed with fixture fallback, map/list sync, filters, selected-incident detail, role-protected framing, and loading/empty/error states.                                                                                                                                                                                                                                                                                |
| NADAA-101 | MVP Sprint 1/7                 | Set Up CI/CD And Staging Environment                        | Done   | Codex | main                                                       | NADAA-001                                                     | 2026-07-06  | GitHub Actions CI, manual staging smoke workflow, Docker build validation, staging env template, and staging runbook added; registry push/deploy credentials remain environment-owned.                                                                                                                                                                                                                                                                                                               |
| NADAA-021 | MVP Sprint 5                   | Implement Area Risk API                                     | Done   | Codex | main                                                       | NADAA-020                                                     | 2026-07-06  | Risk service now returns fixture-backed geospatial risk lookup with low/high/severe flood scoring, nearby shelters, nearby facilities, recommended actions, validation, docs, and tests.                                                                                                                                                                                                                                                                                                             |
| NADAA-022 | MVP Sprint 5                   | Build Citizen Risk Checker UI                               | Done   | Codex | main                                                       | NADAA-021                                                     | 2026-07-06  | Citizen risk surface now uses the risk API with area presets, coordinate entry, GPS lookup, loading/error/permission/empty states, shelters, facilities, recommended actions, and smoke coverage.                                                                                                                                                                                                                                                                                                    |
| NADAA-060 | MVP Sprint 5                   | Implement Emergency Guide Content Model                     | Done   | Codex | main                                                       | NADAA-020                                                     | 2026-07-06  | Guide-service API, guide content fixtures, seed expansion, shared guide types, docs, Docker/CI wiring, staging smoke wiring, and lookup tests added.                                                                                                                                                                                                                                                                                                                                                 |
| NADAA-061 | MVP Sprint 5                   | Build Offline-First Citizen Guidance UI                     | Done   | Codex | main                                                       | NADAA-060                                                     | 2026-07-06  | Citizen guide browser now integrates guide-service, caches offline guides in localStorage and the service worker, supports hazard/stage/language filters with English fallback, shows a visible 112 CTA, and has docs, smoke, and test coverage.                                                                                                                                                                                                                                                     |
| NADAA-062 | MVP Sprint 5                   | Implement Shelter And Recovery Support Module               | Done   | Codex | main                                                       | NADAA-020, NADAA-022                                          | 2026-07-06  | Shelter-service nearby lookup, recovery support locations, protected occupancy updates, citizen shelter map/list and recovery panel, authority capacity update view, shared types, docs, CI/Docker wiring, smoke, and tests added.                                                                                                                                                                                                                                                                   |
| NADAA-080 | MVP Sprint 6                   | Define Agency Integration Contracts                         | Done   | Codex | main                                                       | NADAA-002                                                     | 2026-07-06  | Integration matrix, inbound weather/hydrology contracts, outbound incident/alert sync contracts, mock integration service, shared types, docs, CI/staging wiring, smoke script, Dockerfile, and tests added.                                                                                                                                                                                                                                                                                         |
| NADAA-081 | MVP Sprint 6                   | Implement Weather And Hydrology Import Skeleton             | Done   | Codex | main                                                       | NADAA-020, NADAA-080                                          | 2026-07-06  | Integration-service now imports fixture rainfall/water-level observations into a `weather_observations`-aligned store, logs import jobs, supports retryable failures, exposes imported observations, includes a scheduled importer hook, and has shared types, docs, smoke, and tests.                                                                                                                                                                                                               |
| NADAA-070 | MVP Sprint 6                   | Create Flood Risk Dataset And Feature Pipeline              | Done   | Codex | main                                                       | NADAA-020, NADAA-002                                          | 2026-07-06  | Versioned flood-risk fixture inputs, 44-column schema, deterministic generation/validation scripts, generated JSON/CSV outputs, manifest checksums, candidate source notes, ML/data docs, QA status, dashboard record, and verification added.                                                                                                                                                                                                                                                       |
| NADAA-071 | MVP Sprint 6                   | Train Baseline Flood Risk Model                             | Done   | Codex | main                                                       | NADAA-070                                                     | 2026-07-06  | Deterministic logistic-regression trainer over `flood-risk-features.v1`, model metadata, sample predictions, evaluation JSON/report, medium fixture confidence, false-positive/false-negative review process, validation script, ML-service docs, dashboard record, and verification added.                                                                                                                                                                                                          |
| NADAA-072 | MVP Sprint 6                   | Serve ML Predictions Through Risk API                       | Done   | Codex | main                                                       | NADAA-071, NADAA-021                                          | 2026-07-06  | ML service prediction endpoint over baseline artifacts, in-memory prediction logs aligned to `ml_predictions`, risk-service decision-support enrichment, no-auto-publish safety flags, smoke/staging/CI/Docker wiring, docs, dashboard record, and verification added.                                                                                                                                                                                                                               |
| NADAA-073 | MVP Sprint 6                   | Add Authority ML Review View                                | Done   | Codex | main                                                       | NADAA-072, NADAA-050                                          | 2026-07-06  | Dispatcher-web ML prediction review map/list, explanation panel, live ML-service loading with fixture fallback, reviewed alert-draft action, structured `sourcePrediction` audit metadata, smoke/docs/dashboard records, and verification added.                                                                                                                                                                                                                                                     |
| NADAA-092 | MVP Sprint 7                   | Security Review And Hardening                               | Done   | Codex | main                                                       | MVP feature completion                                        | 2026-07-07  | Runtime HTTP hardening, `NADAA_ALLOWED_ORIGINS` CORS allowlist support, defensive API headers, security scan automation, non-root web containers, dependency scan, residual-risk register, docs, dashboard record, and verification completed.                                                                                                                                                                                                                                                       |
| NADAA-050 | MVP Sprint 4                   | Implement Alert Creation And Approval Workflow              | Done   | Codex | main                                                       | NADAA-011, NADAA-012, NADAA-020                               | 2026-07-06  | Alert-service draft/update/submit/approve/reject/emergency override API, RBAC/MFA/audit hooks, authority dashboard form/queue, shared types, docs, smoke, CI, tests, and Docker build added.                                                                                                                                                                                                                                                                                                         |
| NADAA-051 | MVP Sprint 4                   | Implement Geofenced Alert Targeting                         | Done   | Codex | main                                                       | NADAA-050                                                     | 2026-07-06  | Alert-service target geometry/preview endpoint/list filtering, district/radius/custom validation, authority dashboard geofence controls/preview, shared types, docs, smoke, tests, and Docker build added.                                                                                                                                                                                                                                                                                           |
| NADAA-052 | MVP Sprint 4                   | Implement In-App Alert Feed And Push/SMS Abstraction        | Done   | Codex | main                                                       | NADAA-050, NADAA-051                                          | 2026-07-06  | Notification-service citizen feed API, mock/disabled push/SMS providers, delivery logs, delivery-log schema, citizen current/expired feed UI, shared types, docs, smoke, tests, CI, and Docker build added.                                                                                                                                                                                                                                                                                          |
| NADAA-041 | MVP Sprint 3                   | Implement Verification And Status Workflow                  | Done   | Codex | main                                                       | NADAA-040, NADAA-012                                          | 2026-07-06  | Incident-service status transition rules, verification endpoint, RBAC/MFA gates, audit events, authority dashboard status controls, shared types, docs, smoke, tests, and Docker build added.                                                                                                                                                                                                                                                                                                        |
| NADAA-042 | MVP Sprint 3                   | Implement Agency Assignment And Incident Timeline           | Done   | Codex | main                                                       | NADAA-041                                                     | 2026-07-06  | Incident-service assignment endpoint/model, timeline events, assigned-agency filtering, RBAC/MFA gates, authority dashboard assignment controls, shared types, docs, smoke, tests, and Docker build added.                                                                                                                                                                                                                                                                                           |
| NADAA-043 | MVP Sprint 3                   | Implement Duplicate Merge Review                            | Done   | Codex | main                                                       | NADAA-033, NADAA-041                                          | 2026-07-06  | Incident-service duplicate review and merge endpoints, merge trace fields, audit/timeline events, authority dashboard side-by-side review UI, shared types, docs, smoke, tests, and Docker build added.                                                                                                                                                                                                                                                                                              |
| NADAA-091 | MVP Sprint 3                   | Implement Abuse, Spam, And False Report Handling            | Done   | Codex | main                                                       | NADAA-030, NADAA-041                                          | 2026-07-06  | Incident-service suspicious report signals, abuse review endpoint, false-report review workflow, authority dashboard moderation controls, shared types, docs, smoke, tests, and Docker build added.                                                                                                                                                                                                                                                                                                  |
| NADAA-044 | MVP Sprint 3/7                 | Create Dedicated Dispatcher Web Command Console             | Done   | Codex | main                                                       | NADAA-040, NADAA-041, NADAA-042, NADAA-043, NADAA-091         | 2026-07-06  | Scaffolded `apps/dispatcher-web` with `src/app/` shell and `features/dispatch-command/`, preserved command workflows, kept `authority-dashboard` as compatibility shell, and wired workspace scripts, smoke checks, staging docs, CI Docker build, and verification.                                                                                                                                                                                                                                 |
| NADAA-014 | MVP Sprint 7                   | Build Admin Web Governance Console                          | Done   | Codex | main                                                       | NADAA-011, NADAA-012, NADAA-080                               | 2026-07-06  | Scaffolded modular `apps/admin-web` with admin RBAC/MFA shell, agency/user/role management views, MFA support, audit logs, data-source and alert-rule views, Outfit typography across web apps, scripts, docs, CI/staging smoke, Docker build, and verification.                                                                                                                                                                                                                                     |
| NADAA-015 | MVP Sprint 7/Phase 2 Sprint 11 | Build Public Marketing Website                              | Done   | Codex | main                                                       | NADAA-004, NADAA-002                                          | 2026-07-07  | `apps/marketing-web` public marketing site completed with NADAA about, features, platform lanes, services, benefits, research-backed Ghana disaster context, emergency 112 contact, partner contact, real brand assets, workspace dev/smoke/staging scripts, dedicated smoke-marketing script, CI Docker matrix entry, docs, dashboard record, and verification.                                                                                                                                     |
| NADAA-090 | MVP Sprint 2                   | Implement Location Privacy And Anonymous Reporting Controls | Done   | Codex | main                                                       | NADAA-030, NADAA-040                                          | 2026-07-06  | Incident-service now requires authority context for incident lists, sanitizes reporter/contact data across authority views, returns privacy metadata, and surfaces privacy state in dispatcher/authority/citizen UI with docs, tests, and smoke coverage.                                                                                                                                                                                                                                            |
| NADAA-102 | MVP Sprint 7                   | Conduct UAT, Beta, And Production Readiness                 | Done   | Codex | main                                                       | NADAA-100, NADAA-101, MVP feature completion                  | 2026-07-07  | UAT scripts, feedback capture process, beta metrics dashboard definition, release notes template, user/training guide, acceptance checklist, hypercare checklist, release-readiness validation, dashboard record, and verification completed.                                                                                                                                                                                                                                                        |
| NADAA-110 | Phase 2 Sprint 8               | SMS/USSD Emergency Access                                   | Done   | Codex | main                                                       | NADAA-010, NADAA-030, NADAA-052                               | 2026-07-07  | Notification-service SMS/USSD provider webhooks, language menu, current alert summaries, basic report intake, 112 guidance, profile-link consent, structured INFO/WARN/ERROR runtime logging, optional incident-service handoff, shared types, smoke/docs/dashboard updates, and verification completed.                                                                                                                                                                                             |
| NADAA-111 | Phase 2 Sprint 8/9             | WhatsApp Emergency Chatbot                                  | Done   | Codex | main                                                       | NADAA-030, NADAA-052, NADAA-060                               | 2026-07-07  | Notification-service WhatsApp webhook aliases, alerts/risk/report/shelter/guide/112 intents, multi-message report state, location/media incident handoff, privacy-safe transcript summaries, shared types, smoke/docs/dashboard updates, structured logs, and verification completed.                                                                                                                                                                                                                |
| NADAA-112 | Phase 2 Sprint 9               | Multilingual Voice Alerts                                   | Done   | Codex | main                                                       | NADAA-050, NADAA-052, NADAA-111                               | 2026-07-07  | Notification-service voice alert asset generation, English/Twi/Ga/Ewe/Dagbani/Hausa low-literacy templates, sandbox TTS/recorded-audio metadata, review/approval gate, approved voice delivery logs, shared types, smoke/docs/dashboard updates, staging env guidance, and verification completed.                                                                                                                                                                                                   |
| NADAA-113 | Phase 2 Sprint 8/9             | Build Citizen Mobile App Foundation                         | Done   | Codex | main                                                       | NADAA-010, NADAA-021, NADAA-030, NADAA-052, NADAA-060         | 2026-07-07  | `apps/citizen-mobile` Expo/React Native foundation, NADAA logo/brand theme, native navigation/session shell, alerts/risk/report/guides/support screens, offline guide cache and report draft primitives, permission copy, sandbox push registration, smoke/docs/dashboard updates, and verification completed.                                                                                                                                                                                       |
| NADAA-120 | Phase 2 Sprint 10              | Community Volunteer App                                     | Done   | Codex | main                                                       | NADAA-011, NADAA-042                                          | 2026-07-07  | Incident-service volunteer registration, verification, district/community group membership, incident-linked task assignment, status/observation updates, safety/escalation rules, incident timeline/audit events, structured INFO/WARN/ERROR logs, shared contracts, citizen-mobile Community tab, offline task cache, smoke/docs/dashboard updates, and verification completed.                                                                                                                     |
| NADAA-121 | Phase 2 Sprint 10              | Hospital Capacity Tracker                                   | Done   | Codex | main                                                       | NADAA-020, NADAA-080                                          | 2026-07-07  | Shelter-service hospital/facility capacity schema, stale-data flags, manual authority updates, fixture adapter imports, source tracking, dispatcher-web capacity filters/list, shared contracts, smoke/docs/dashboard updates, structured INFO/WARN/ERROR logs, and verification completed.                                                                                                                                                                                                          |
| NADAA-124 | Phase 2 Sprint 10              | Build Dispatcher Mobile Triage App                          | Done   | Codex | feature/NADAA-124-dispatcher-mobile-triage                 | NADAA-041, NADAA-042, NADAA-091, NADAA-044                    | 2026-07-07  | `apps/dispatcher-mobile` Expo/RN foundation, agency session/MFA shell, incident queue-detail-action-capacity-profile screens, status/assignment/timeline-note actions, hospital capacity lookup, offline/stale indicators, sandbox push registration, smoke script, docs, dashboard record, and verification completed.                                                                                                                                                                              |
| NADAA-125 | Phase 2 Sprints 10/11          | Build Agency Web Operations Portal                          | Done   | Codex | main                                                       | NADAA-011, NADAA-042, NADAA-062, NADAA-121                    | 2026-07-07  | `apps/agency-web` React/Vite operations portal with agency session/MFA shell, assigned-incident dashboard, responder status/timeline-note updates, shelter/hospital capacity context, occupancy/capacity update forms, smoke-web/staging-smoke integration, Dockerfile, CI Docker matrix, docs/dashboard updates, and verification.                                                                                                                                                                  |
| NADAA-131 | Phase 2 Sprint 11              | Road Closure Integration                                    | Done   | Codex | main                                                       | NADAA-020, NADAA-080                                          | 2026-07-07  | `services/road-closure-service` Go API with list/create/update/adapter-import, database migration, shared TypeScript contracts, dispatcher-web map polyline layer, agency-web nearby closure context, citizen-web closure cards, integration-service adapter import endpoint, smoke script, CI Docker matrix, docs, dashboard record, and verification.                                                                                                                                              |
| NADAA-122 | Phase 2 Sprint 11              | Relief Distribution Tracking                                | Done   | Codex | main                                                       | NADAA-062, NADAA-012                                          | 2026-07-07  | Shelter-service relief point list/nearby/create/update/stock-history APIs, shared TypeScript contracts, agency web relief management tab, citizen web nearby relief display, dispatcher-web relief map markers and panel, smoke/docs/dashboard updates, structured INFO/WARN/ERROR logs, and verification completed.                                                                                                                                                                                 |
| NADAA-123 | Phase 2 Sprint 11              | Donation And Aid Coordination                               | Done   | Codex | main                                                       | NADAA-122, NADAA-012                                          | 2026-07-07  | Shelter-service aid request/pledge APIs, PostGIS migration, public approved-needs list, authority review, donor pledge flow, agency Aid tab, CSV export, smoke/docs/dashboard updates, structured INFO/WARN/ERROR logs, and verification completed.                                                                                                                                                                                                                                                  |
| NADAA-171 | Phase 2 Sprint 11              | Refactor Go Services Into Modular Packages                  | Done   | Codex | feature/NADAA-171-go-service-modularization                | All Go services                                               | 2026-07-07  | Split every Go service `main.go` monolith into `cmd/server`, `internal/config`, `internal/models`, `internal/store`, `internal/utils`, and `internal/handlers`; moved `apps/admin-web` to modular `api/`, `auth/`, `components/`, `data/`, `lib/`, `pages/` layout; updated Dockerfiles, tests, docs, and full verification passed.                                                                                                                                                                  |
| NADAA-172 | Phase 2 Sprint 11              | Go Service Quality & Graceful Shutdown                      | Done   | Kimi  | feature/NADAA-172-go-service-quality-and-graceful-shutdown | NADAA-171                                                     | 2026-07-07  | Completed graceful `http.Server` shutdown with timeouts and signal handling in all 10 Go services; added `doc.go` package comments; fixed exported-symbol docs, constants, context propagation, and lint issues across services; added `.golangci.yml` v2 baseline; expanded `docs/user-guide.md` and `docs/architecture.md`; full verification passed.                                                                                                                                              |
| NADAA-173 | Phase 2 Sprint 11              | Cross-Platform UI/UX Redesign And Design-System Hardening   | Done   | Kimi  | feature/NADAA-173-ui-ux-redesign                           | All frontend apps, NADAA-014, NADAA-044, NADAA-062, NADAA-125 | 2026-07-08  | Extended `packages/brand` with semantic tokens, spacing, typography, elevation, breakpoints, and accessible severity/hazard roles. Added `createNadaaTheme` canonical MUI theme factory and `brand.css`. Refactored `citizen-web`, `dispatcher-web`, `authority-dashboard`, `agency-web`, `admin-web`, and `marketing-web` to consume shared tokens, normalize CSS, add skip links/focus rings, and use icon+text severity chips. Added `docs/design-system.md`. Full workspace verification passed. |
| NADAA-174 | Phase 2 Sprint 11              | Extend Design System To Mobile Apps                         | Done   | Kimi  | feature/NADAA-173-ui-ux-redesign                           | NADAA-173                                                     | 2026-07-08  | Extended `@nadaa/brand` with a React Native `nativeTheme`, `severityBadgeFor`, and `hazardBadgeFor`. Rewrote `citizen-mobile` and `dispatcher-mobile` themes to consume shared native tokens. Replaced hard-coded colors/spacing with tokens and added accessible icon+text severity/hazard/urgency/capacity badges. Ensured touch targets ≥44 px and WCAG 2.1 AA contrast. `pnpm smoke:citizen-mobile` and `pnpm smoke:dispatcher-mobile` passed. Full workspace verification passed.               |
| NADAA-130 | Phase 2 Sprint 12              | Evacuation Route Planner                                    | Done   | Kimi  | feature/NADAA-130-evacuation-route-planner                 | NADAA-021, NADAA-062, NADAA-131                               | 2026-07-07  | services/route-service Go API with /routes/plan, /routes/options, /health, shelter/higher-ground/manual waypoints, closure/risk avoidance, structured logging, tests, Dockerfile, README, dashboard record, CI Docker matrix entry, and smoke script. Citizen-web route planner card and authority-dashboard route planner panel with Leaflet polyline, accessible forms, and decision-support disclaimer. Full workspace verification passed.                                                       |

### Agent Handoff Log

| Date       | Agent  | Item                | Status      | Handoff Notes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ---------- | ------ | ------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-07-09 | Claude | NADAA-151           | Done        | Took over from Kimi's in-progress first pass. Verified `services/incident-service go test ./...`, `go vet ./...`, `go build ./cmd/server`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/dispatcher-web typecheck` and `build`, live `pnpm smoke:ai-triage` plus regression `pnpm smoke:incident-workflow`, `smoke:incident-merge`, `smoke:incident-abuse`, and `smoke:incident-assignment` on incident-service `:18084`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm go:vet`, `pnpm security:scan`, and `git diff --check`. Kimi's first pass added incident-service triage endpoints (`GET /incidents/{id}/triage`, `POST /incidents/{id}/triage-review`) with explainable rules-based suggestions and accept/override timeline + audit logging, dispatcher-web AI triage panel, shared types, smoke script, api.md docs, and dashboard record. Claude fixed the triage-review response parsing (`{ incident }` envelope), removed the unsupported `severe` override severity option, made Accept log `accepted: true` through the API so acceptances are auditable, added the ml.md triage model + bias/error review process section, applied gofmt, and ran the full verification above plus an adversarial multi-agent diff review before commit.                                                                                                                                                                                                                                                                          |
| 2026-07-09 | Kimi  | NADAA-132           | Done   | Verified `services/missing-person-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/citizen-web typecheck` and `build`, `pnpm --filter @nadaa/authority-dashboard typecheck` and `build`, live `pnpm smoke:missing-person`, `pnpm validate:dashboard`, `pnpm validate:database`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm go:lint`, `docker build -f services/missing-person-service/Dockerfile -t nadaa/missing-person-service:local .`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added `services/missing-person-service` Go API with sensitive intake, authority review/public visibility, closure/reunification audit trail, approved public search, PostGIS migration, shared TypeScript contracts, citizen-web missing-person intake/search, authority-dashboard review/closure/audit panel, smoke/docs/dashboard updates, CI Docker matrix entry, and structured INFO/WARN/ERROR logs. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| 2026-07-08 | Kimi  | NADAA-123           | Done        | Verified `services/donation-service go test ./...` and `go build ./cmd/server`, live `DONATION_API_URL=http://127.0.0.1:8100/api/v1 pnpm smoke:donation`, `docker build -f services/donation-service/Dockerfile -t nadaa/donation-service:local .`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/authority-dashboard typecheck` and `build`, `pnpm --filter @nadaa/citizen-web typecheck` and `build`, `pnpm validate:docs`, `pnpm validate:database`, `pnpm validate:release`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm go:lint`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Rebased `feature/NADAA-123-donation-and-aid-coordination` onto current `main`, refactored `services/donation-service` into modular `cmd/server` + `internal/*` packages with graceful shutdown, preserved donors/aid-catalog/aid-requests/pledges/allocate endpoints, wired citizen-web donor portal and authority-dashboard donation panel, resolved shared-type conflicts with shelter aid types by renaming donation aid types, updated scripts/CI/docs/dashboard. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                           |
| 2026-07-07 | Codex  | NADAA-125           | Done        | Verified `pnpm --filter @nadaa/agency-web typecheck`, `pnpm --filter @nadaa/agency-web build`, `LOCAL_AGENCY_URL=http://127.0.0.1:5177/ pnpm smoke:web`, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `docker build -f apps/agency-web/Dockerfile -t nadaa/agency-web:local .`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added `apps/agency-web` React/Vite operations portal with agency session/MFA shell, assigned-incident dashboard, responder status/timeline-note updates, shelter/hospital capacity context, occupancy/capacity update forms, smoke-web/staging-smoke integration, Dockerfile, CI Docker matrix entry, dev script, docs, dashboard record, and plan updates. `pnpm audit --audit-level moderate` still reports one transitive Expo CLI `uuid` advisory via `xcode`; high-threshold audit passes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| 2026-07-07 | Codex  | NADAA-131           | Done        | Verified `services/road-closure-service go test ./...`, `services/integration-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/dispatcher-web typecheck`, `pnpm --filter @nadaa/dispatcher-web build`, `pnpm --filter @nadaa/agency-web typecheck`, `pnpm --filter @nadaa/agency-web build`, `pnpm --filter @nadaa/citizen-web typecheck`, `pnpm --filter @nadaa/citizen-web build`, live `pnpm smoke:road-closure` and `pnpm smoke:integration` with integration-service forwarding to road-closure-service, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `docker build -f services/road-closure-service/Dockerfile -t nadaa/road-closure-service:local .`, `docker build -f services/integration-service/Dockerfile -t nadaa/integration-service:local .`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added `services/road-closure-service` Go API with list/create/update/adapter-import, PostGIS migration, shared TypeScript contracts, dispatcher-web map polyline layer, agency-web nearby closure context, citizen-web closure cards, integration-service adapter import endpoint that forwards to `NADAA_ROAD_CLOSURE_SERVICE_URL` (default `http://localhost:8095`), smoke scripts, CI Docker matrix entries, docs, dashboard record, and plan updates. `pnpm audit --audit-level moderate` still reports one transitive Expo CLI `uuid` advisory via `xcode`; high-threshold audit passes. |
| 2026-07-07 | Codex  | NADAA-122           | Done        | Verified `services/shelter-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/citizen-web typecheck`, `pnpm --filter @nadaa/citizen-web build`, `pnpm --filter @nadaa/agency-web typecheck`, `pnpm --filter @nadaa/agency-web build`, `pnpm --filter @nadaa/dispatcher-web typecheck`, `pnpm --filter @nadaa/dispatcher-web build`, live `pnpm smoke:relief` and `pnpm smoke:shelter` on shelter-service `:8093`, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added shelter-service relief point list/nearby/create/update/stock-history APIs, shared TypeScript contracts, citizen web nearby relief display, agency web relief management tab with create/update form and stock history, dispatcher-web relief map markers and panel, smoke/docs/dashboard updates, and structured INFO/WARN/ERROR logs for relief reads, authority denials, creates, updates, history reads, and response encoding failures. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                                                          |
| 2026-07-07 | Codex  | NADAA-123           | Done        | Verified `services/shelter-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/agency-web typecheck`, `pnpm --filter @nadaa/agency-web build`, live `pnpm smoke:aid`, live `pnpm smoke:shelter`, live `pnpm smoke:relief` on shelter-service `:8093`, `pnpm validate:dashboard`, `pnpm validate:database`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `docker build -f services/shelter-service/Dockerfile -t nadaa/shelter-service:local .`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added shelter-service aid request and pledge APIs with public/partner filters, authority create/review gates, donor pledge intake/review, active pledge summaries, CSV export, anti-fraud/review notes, PostGIS aid coordination migration, shared TypeScript contracts, agency web Aid tab, smoke/docs/dashboard updates, and structured INFO/WARN/ERROR logs for aid reads, authority denials, creates, reviews, pledges, exports, and response errors. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                                                                                             |
| 2026-07-07 | Codex  | NADAA-171           | Done        | Verified `pnpm go:test` for all 10 Go services, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Refactored every Go service from single `main.go` monoliths into `cmd/server` + `internal/config`, `internal/models`, `internal/store`, `internal/utils`, and `internal/handlers`; moved tests into `internal/handlers` under `package handlers`; updated Dockerfiles to build `./cmd/server`. Also finalized the pending `apps/admin-web` modular refactor into `api/`, `auth/`, `components/`, `data/`, `lib/`, `pages/`, and `routes.tsx`. Preserved all env vars, routes, CORS, security headers, and behavior. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| 2026-07-07 | Kimi   | NADAA-172           | Done        | Verified `pnpm go:lint` (0 issues across all 10 services), `pnpm go:test`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm validate:docs`, `pnpm validate:database`, `pnpm validate:release`, `pnpm security:scan`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added `http.Server` graceful shutdown with `ReadTimeout: 10s`, `WriteTimeout: 30s`, `IdleTimeout: 120s`, signal handling for `os.Interrupt` and `syscall.SIGTERM`, and `doc.go` comments for every service package. Added `.golangci.yml` v2 baseline with `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`, `revive`, `gocognit` (min 30), and `misspell`; resolved all issues except two shelter-service legacy functions flagged with `//nolint:gocognit` for future refactor. Expanded `docs/user-guide.md` and `docs/architecture.md` and preserved release-readiness tokens.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| 2026-07-07 | Kimi   | NADAA-173           | Ready       | Claimed cross-platform UI/UX redesign story. Scope: audit current web app surfaces, consolidate shared design tokens and theme, apply consistent mobile-first layouts, accessibility improvements, and updated components across citizen-web, dispatcher-web, agency-web, admin-web, marketing-web, and authority-dashboard. Implementation plan required before coding begins.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| 2026-07-08 | Kimi   | NADAA-173           | Done        | Verified `pnpm --filter @nadaa/{citizen-web,dispatcher-web,authority-dashboard,agency-web,admin-web,marketing-web} typecheck` and `build`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm validate:docs`, `pnpm validate:database`, `pnpm validate:release`, `pnpm security:scan`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Extended `packages/brand` with tokens, `createNadaaTheme`, and `brand.css`. Refactored all six web apps to use the canonical theme, CSS token variables, skip links, visible focus rings, accessible severity/hazard chips, and responsive table wrappers. Added `docs/design-system.md`. `pnpm smoke:marketing` passed. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| 2026-07-08 | Kimi   | NADAA-174           | Done        | Verified `pnpm --filter @nadaa/citizen-mobile typecheck` and `build`, `pnpm --filter @nadaa/dispatcher-mobile typecheck` and `build`, `pnpm smoke:citizen-mobile`, `pnpm smoke:dispatcher-mobile`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm validate:docs`, `pnpm validate:database`, `pnpm validate:release`, `pnpm security:scan`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Extended `packages/brand` with `@nadaa/brand/native` (`nativeTheme`, `severityBadgeFor`, `hazardBadgeFor`). Rewrote `citizen-mobile` and `dispatcher-mobile` themes to consume shared native tokens, replaced hard-coded colors/spacing, and added accessible icon+text severity/hazard/urgency/capacity badges. Ensured touch targets ≥44 px and WCAG 2.1 AA contrast. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 2026-07-08 | Kimi   | NADAA-130           | Done        | Verified `services/route-service go test ./...` and `go build ./cmd/server`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/citizen-web typecheck` and `build`, `pnpm --filter @nadaa/authority-dashboard typecheck` and `build`, live `pnpm smoke:route`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm validate:docs`, `pnpm validate:database`, `pnpm validate:release`, `pnpm security:scan`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added `services/route-service` Go API with in-memory route planning, closure/risk avoidance, shelter/higher-ground/manual waypoints, structured logging, tests, Dockerfile, README, dashboard record, CI Docker matrix entry, and `smoke:route`. Added shared route TypeScript types, citizen-web route planner card, and authority-dashboard route planner panel with Leaflet polyline. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| 2026-07-07 | Codex  | NADAA-015           | Done        | Verified `pnpm --filter @nadaa/marketing-web typecheck`, `pnpm --filter @nadaa/marketing-web build`, `pnpm smoke:marketing` on localhost:5172, `pnpm smoke:web` with explicit local NADAA URLs, `pnpm smoke:staging` with explicit local NADAA staging URLs, `docker build -f apps/marketing-web/Dockerfile -t nadaa/marketing-web:local .`, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added `apps/marketing-web` public marketing site with NADAA about, features, platform lanes, services, benefits, research-backed Ghana disaster context, emergency 112 contact, partner contact, real brand assets, workspace dev/smoke/staging scripts, dedicated smoke-marketing script, CI Docker matrix entry, docs, dashboard record, and plan updates. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| 2026-07-07 | Codex  | NADAA-121           | Done        | Verified `services/shelter-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/dispatcher-web typecheck`, live `pnpm smoke:shelter` on shelter-service `:18093`, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added shelter-service hospital/facility capacity records, stale-data flags, distance/service/capacity/min-bed filters, manual authority updates, fixture adapter imports, source tracking, structured INFO/WARN/ERROR logs, shared TypeScript contracts, dispatcher-web hospital capacity panel, docs, smoke coverage, and dashboard record. `pnpm audit --audit-level high` exits clean while reporting one moderate advisory.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| 2026-07-07 | Codex  | NADAA-120           | Done        | Verified `services/incident-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/citizen-mobile typecheck`, `pnpm smoke:citizen-mobile`, live `pnpm smoke:community-volunteers`, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added incident-service volunteer profiles, verification, district/community response groups, volunteer task assignment, task status/field-observation endpoints, safety/escalation validation, incident timeline/audit events, structured INFO/WARN/ERROR logs, shared TypeScript contracts, citizen-mobile Community tab with offline profile/task cache, live smoke script, docs, and dashboard record. `pnpm audit --audit-level moderate` still reports one transitive Expo CLI `uuid` advisory via `xcode`; high-threshold audit passes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| 2026-07-07 | Codex  | NADAA-113           | Done        | Verified `pnpm --filter @nadaa/citizen-mobile typecheck`, `pnpm --filter @nadaa/citizen-mobile build`, `pnpm smoke:citizen-mobile`, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, and `git diff --check`. Added `apps/citizen-mobile` Expo/React Native foundation with shared NADAA brand/types/config, real logo asset, native shell and bottom navigation, session state, alert/risk/report/guide/support screens, offline guide cache and report draft primitives, location/media/camera/push permission copy, sandbox push registration, mobile smoke script, docs, dashboard record, and lockfile workspace integration. `pnpm audit --audit-level moderate` reports one transitive Expo CLI `uuid` advisory via `xcode`; high-threshold audit passes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| 2026-07-07 | Codex  | NADAA-112           | Done        | Verified `services/notification-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, live `pnpm smoke:voice-alerts`, live `pnpm smoke:notification`, live `pnpm smoke:sms-ussd`, live `pnpm smoke:whatsapp`, supported-file `pnpm exec prettier --check`, and `git diff --check`. Added notification-service voice asset generation/review/delivery endpoints, multilingual low-literacy templates, sandbox TTS/recorded-audio metadata, approval-gated voice delivery attempts, delivery log filters for `channel=voice`, shared types, smoke script, docs/dashboard/staging updates, and structured INFO/WARN/ERROR logs.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| 2026-07-07 | Codex  | NADAA-111           | Done        | Verified `services/notification-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, live `pnpm smoke:notification`, live `pnpm smoke:sms-ussd`, live `pnpm smoke:whatsapp`, supported-file `pnpm exec prettier --write`, and `git diff --check`. Added notification-service WhatsApp webhook aliases, provider/sandbox contract, alerts/risk/report/shelter/guides/112 intents, multi-message report conversation state, location and media ingestion, optional incident-service handoff, privacy-safe transcript summaries with 90-day retention, shared types, smoke/docs/dashboard updates, and structured INFO/WARN/ERROR logs.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| 2026-07-07 | Codex  | NADAA-110 logging   | Done        | Verified `services/notification-service go test ./...`, live `pnpm smoke:sms-ussd`, live `pnpm smoke:notification`, `pnpm validate:docs`, `pnpm validate:dashboard`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --write`, and `git diff --check`. Added structured `INFO`, `WARN`, and `ERROR` runtime logs across notification-service startup, citizen alert reads, delivery attempts, SMS/USSD webhook receipt, validation failures, menu transitions, provider errors, report creation, incident-service handoff, queue fallback, and inclusive access log/report persistence. Logs use `phoneRef`, command names, path depth, counts, statuses, provider IDs, and stored record IDs instead of raw phone numbers, full SMS bodies, or full report details.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| 2026-07-07 | Codex  | NADAA-110           | Done        | Verified `services/notification-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm validate:dashboard`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, live `pnpm smoke:notification`, live `pnpm smoke:sms-ussd`, supported-file `pnpm exec prettier --write`, and `git diff --check`. Added SMS/USSD provider webhooks, language menu, current alert summaries, basic report intake, 112 guidance, profile-link consent, provider error logging, optional incident-service handoff, shared types, smoke/docs/dashboard updates.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| 2026-07-07 | Codex  | NADAA-102           | Done        | Verified `pnpm validate:release`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm audit --audit-level high`, supported-file `pnpm exec prettier --write`, and `git diff --check`. Added UAT scripts and feedback capture, release-readiness gates and release notes template, user/training guide, beta monitoring metrics, hypercare checklist, readiness validation script, README/QA/deployment links, dashboard record, and plan updates.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| 2026-07-07 | Codex  | NADAA-092           | Done        | Verified `pnpm security:scan`, `pnpm audit --audit-level high`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, representative `docker build -f apps/citizen-web/Dockerfile -t nadaa/citizen-web:security-local .`, supported-file `pnpm exec prettier --write`, and `git diff --check`. Added API CORS allowlist and defensive headers across Go services, alert-service middleware regression test, non-root web image runtimes, CI security scan, security review/residual-risk report, staging allowlist documentation, QA/deployment docs, and dashboard record.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| 2026-07-06 | Codex  | NADAA-073           | Done        | Verified `services/alert-service go test ./...`, `pnpm --filter @nadaa/dispatcher-web typecheck`, `pnpm --filter @nadaa/dispatcher-web build`, `pnpm --filter @nadaa/shared-types typecheck`, live `pnpm smoke:ml-review` with ML service on `:18094` and alert-service on `:18089`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, and `git diff --check`. Added dispatcher ML review map/list, explanation panel, reviewed alert draft action, structured source-prediction metadata, docs, smoke, QA, and dashboard record.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| 2026-07-06 | Codex  | NADAA-072           | Done        | Verified `services/ml-service go test ./...`, `services/risk-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, live `pnpm smoke:ml` on ML service `:18094`, live `RISK_EXPECT_ML=true pnpm smoke:risk` with risk-service connected to ML service, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, ML service Docker build, and `git diff --check`. Added ML service prediction/log endpoints, risk-service ML enrichment, shared contracts, docs, smoke, CI/staging wiring, and no-auto-publish safety flags.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| 2026-07-06 | Codex  | NADAA-071           | Done        | Verified `pnpm ml:flood:train`, `pnpm validate:ml`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, and `git diff --check`. Added deterministic logistic-regression training over the NADAA-070 feature set, model metadata, sample predictions, evaluation JSON/report, medium fixture confidence, false-positive/false-negative review process, validator and package scripts, ML-service/docs/QA/dashboard records.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| 2026-07-06 | Codex  | NADAA-070           | Done        | Verified `pnpm features:flood`, `pnpm validate:features`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, and `git diff --check`. Added versioned flood-risk source fixtures, 44-column feature schema, deterministic generator, JSON/CSV outputs, manifest checksums, validation script, candidate production source notes, ML/QA/deployment/product/architecture docs, and dashboard record.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| 2026-07-06 | Codex  | NADAA-062           | Done        | Verified `services/shelter-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/citizen-web typecheck`, `pnpm --filter @nadaa/authority-dashboard typecheck`, live `pnpm smoke:shelter` on shelter-service `:8093`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, shelter-service Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| 2026-07-06 | Codex  | NADAA-062           | In Progress | Claimed shelter and recovery support module. Scope: add dedicated shelter-service with nearby lookup and authority occupancy update, seed shelter/recovery support fixtures, add shared shelter/recovery types, expose citizen shelter/recovery map/list using the service with fallback, add authority shelter update view, wire smoke/docs/dashboard/CI/Docker, and verify.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| 2026-07-06 | Codex  | NADAA-090           | Done        | Verified `services/incident-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm --filter @nadaa/dispatcher-web typecheck`, `pnpm --filter @nadaa/authority-dashboard typecheck`, `pnpm --filter @nadaa/citizen-web typecheck`, local `smoke-incident-workflow`, `smoke-incident-abuse`, `smoke-incident-assignment`, and `smoke-incident-merge` against incident-service on `127.0.0.1:18084`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| 2026-07-06 | Codex  | NADAA-090           | In Progress | Claimed location privacy and anonymous reporting controls. Scope: enforce reporter/contact/location minimization in incident-service authority list and duplicate-review payloads, surface privacy policy state in dispatcher/authority command views, update citizen location/privacy copy, shared types, API/security/QA docs, dashboard record, and verification.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| 2026-07-06 | Codex  | NADAA-014           | Done        | Verified `pnpm --filter @nadaa/admin-web typecheck`, `pnpm --filter @nadaa/admin-web build`, `pnpm validate:docs`, `pnpm typecheck`, `pnpm lint`, `pnpm test`, `pnpm build`, `pnpm go:test`, `LOCAL_ADMIN_URL=http://127.0.0.1:5180/ pnpm smoke:web`, local `pnpm smoke:staging` with citizen/authority/dispatcher/admin URLs, admin-web Docker build, and `git diff --check`. Local smoke used admin port `5180` because unrelated Xnaplooks servers already owned `127.0.0.1:5176` and `127.0.0.1:5177`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| 2026-07-06 | Codex  | NADAA-014           | In Progress | Claimed admin web governance console. Scope: scaffold modular `apps/admin-web`, enforce system-admin/session/MFA shell, expose agency/user/role, MFA support, audit log, data-source, alert-rule, and platform configuration views using safe fixtures/API fallbacks, apply Outfit typography across web apps, and wire scripts, CI, smoke checks, deployment docs, dashboard records, and verification notes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| 2026-07-06 | Codex  | NADAA-044           | Done        | Verified `pnpm --filter @nadaa/dispatcher-web typecheck`, `pnpm --filter @nadaa/dispatcher-web build`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, local `pnpm smoke:staging` with local citizen/authority/dispatcher URLs, dispatcher-web Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| 2026-07-06 | Codex  | NADAA-044           | In Progress | Claimed dispatcher web split. Scope: scaffold `apps/dispatcher-web`, preserve existing command-center incident map/list, filters, status, assignment, duplicate merge, abuse review, alert workflow, keep authority dashboard compatibility, and wire workspace scripts, smoke checks, staging docs, CI Docker build, dashboard sample, and verification notes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| 2026-07-06 | Codex  | NADAA-004           | Done        | Added a target platform portfolio and platform build board covering citizens web/mobile, dispatchers web/mobile, agency web, and admin web. Added platform delivery breakdown, ownership boundaries, shared dependencies, release gates, build sequencing, and backlog stories for admin web, dispatcher web, citizen mobile, dispatcher mobile, and agency web so future agents can build role-specific apps instead of expanding one authority dashboard.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| 2026-07-06 | Codex  | NADAA-081           | Done        | Verified `services/integration-service go test ./...`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm smoke:integration`, `pnpm smoke:web`, local `pnpm smoke:staging`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, integration-service Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| 2026-07-06 | Codex  | NADAA-081           | In Progress | Claimed integration-service weather/hydrology fixture import job, imported observation store aligned to `weather_observations`, import status logging, retryable failed imports, scheduled importer hook, shared integration types, docs, smoke, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| 2026-07-06 | Codex  | Web modularization  | Done        | Refactored both web apps so root `src/App.tsx` files are thin delegates. Authority dashboard now separates `src/app/` config/session/theme from `features/command-center/` app, components, data, types, and utilities. Citizen web now separates `src/app/` config/theme from `features/citizen/` app, data, types, and utilities. Verified `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm validate:docs`, `pnpm smoke:web`, `pnpm smoke:guide`, `pnpm smoke:citizen-guides`, local `pnpm smoke:staging`, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| 2026-07-06 | Codex  | NADAA-061           | Done        | Verified `pnpm --filter @nadaa/citizen-web typecheck`, `pnpm --filter @nadaa/shared-types typecheck`, `pnpm smoke:guide` with local guide-service on `:18086`, `pnpm smoke:citizen-guides` with local guide-service on `:18086`, `pnpm smoke:web`, local `pnpm smoke:staging`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| 2026-07-06 | Codex  | NADAA-061           | In Progress | Claimed citizen offline-first guidance UI, guide-service integration, localStorage cache fallback, hazard/stage/language filters, mobile-first guide browsing, visible 112 call-to-action, shared guide types, docs, smoke, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| 2026-07-06 | Codex  | NADAA-052           | Done        | Verified `services/notification-service go test ./...`, `pnpm smoke:notification` on `:8091`, `pnpm smoke:web`, local `pnpm smoke:staging`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, notification-service Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| 2026-07-06 | Codex  | NADAA-052           | In Progress | Claimed citizen current/expired alert feed, notification provider abstraction, development/mock push and SMS providers, delivery attempt logs, shared notification/alert-feed types, docs, smoke, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 2026-07-06 | Codex  | NADAA-051           | Done        | Verified `services/alert-service go test ./...`, `pnpm smoke:alert-geofence`, `pnpm smoke:alert`, `pnpm smoke:web`, local `pnpm smoke:staging`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, alert-service Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| 2026-07-06 | Codex  | NADAA-051           | In Progress | Claimed alert-service target geometry model, district/radius/custom geometry validation, target preview response, authority dashboard geofence controls and preview, shared alert types, docs, smoke, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| 2026-07-06 | Codex  | NADAA-091           | Done        | Verified `services/incident-service go test ./...`, `pnpm smoke:incident-abuse`, `pnpm smoke:incident-workflow`, `pnpm smoke:incident-assignment`, `pnpm smoke:incident-merge`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, incident-service Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| 2026-07-06 | Codex  | NADAA-091           | In Progress | Claimed incident-service suspicious report signal model, abuse review endpoint, false-report review workflow, authority dashboard moderation controls, shared types, docs, smoke, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 2026-07-06 | Codex  | NADAA-043           | Done        | Verified `services/incident-service go test ./...`, `pnpm smoke:incident-merge`, `pnpm smoke:incident-assignment`, `pnpm smoke:incident-workflow`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, incident-service Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| 2026-07-06 | Codex  | NADAA-043           | In Progress | Claimed incident-service duplicate review endpoint, merge endpoint, duplicate candidate traceability, merge audit events, authority dashboard side-by-side review controls, shared types, docs, smoke, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| 2026-07-06 | Codex  | NADAA-042           | Done        | Verified `services/incident-service go test ./...`, `pnpm smoke:incident-assignment`, `pnpm smoke:incident-workflow`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, incident-service Docker build, and live assigned-agency filtering on `localhost:8084`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 2026-07-06 | Codex  | NADAA-042           | In Progress | Claimed incident-service assignment model, incident timeline model, assignment endpoint, assigned-agency incident filtering, assignment permission tests, authority dashboard assignment controls, shared types, docs, smoke, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| 2026-07-06 | Codex  | NADAA-041           | Done        | Verified `services/incident-service go test ./...`, `pnpm --filter @nadaa/authority-dashboard typecheck`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, live `pnpm smoke:incident-workflow` on `localhost:8084`, local `pnpm smoke:staging`, incident-service Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| 2026-07-06 | Codex  | NADAA-041           | In Progress | Claimed incident-service status transition rules, verification endpoint, audited status changes, closure/false-report resolution-note validation, authority dashboard status controls, shared types, docs, smoke, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 2026-07-06 | Codex  | NADAA-050           | Done        | Verified `services/alert-service go test ./...`, `pnpm --filter @nadaa/authority-dashboard typecheck`, `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, live `pnpm smoke:alert` on `localhost:8089`, local `pnpm smoke:staging`, alert-service Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| 2026-07-06 | Codex  | NADAA-050           | In Progress | Claimed alert-service draft/update/submit/approve/reject/emergency override API, RBAC/MFA/audit hooks, authority dashboard alert form/queue, shared types, docs, smoke, CI wiring, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| 2026-07-06 | Codex  | NADAA-080           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, live `pnpm smoke:integration` on `localhost:8088`, local `pnpm smoke:staging`, integration-service Docker build, and `git diff --check`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| 2026-07-06 | Codex  | NADAA-080           | In Progress | Claimed integration matrix, inbound weather/hydrology contract, outbound incident/alert sync contract, mock integration service, docs, CI wiring, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 2026-07-06 | Codex  | NADAA-060           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, live `pnpm smoke:guide` on `localhost:8087`, live `pnpm smoke:risk` on `localhost:8081`, local `pnpm smoke:staging`, `git diff --check`, and disposable PostGIS migration/seed apply.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| 2026-07-06 | Codex  | NADAA-060           | In Progress | Claimed guide-service API, guide content fixtures, seed expansion, shared guide types, docs, and lookup tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| 2026-07-06 | Codex  | NADAA-040           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live incident command API smoke on `localhost:8084`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| 2026-07-06 | Codex  | NADAA-040           | In Progress | Claimed authority-dashboard incident command map, API-backed list fallback, map/list sync, filters, role-protected framing, and UI states.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| 2026-07-06 | Codex  | NADAA-012           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live auth/audit HTTP smoke on `localhost:8082`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| 2026-07-06 | Codex  | NADAA-012           | In Progress | Claimed auth-service audit event model, in-memory audit store/helper, auth/admin event wiring, tests, shared types, and retention docs.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| 2026-07-06 | Codex  | NADAA-011           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live agency auth HTTP smoke on `localhost:8082`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| 2026-07-06 | Codex  | NADAA-011           | In Progress | Claimed agency user creation, role catalog, mock MFA setup/verification, and RBAC middleware/tests in auth-service; shared auth contracts and docs may be updated.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| 2026-07-06 | Codex  | NADAA-022           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live `pnpm smoke:risk` on `localhost:8081`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| 2026-07-06 | Codex  | NADAA-022           | In Progress | Claimed citizen risk checker UI; depends on local/staging `VITE_RISK_API_URL` or default risk-service URL.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| 2026-07-06 | Codex  | NADAA-021           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live risk HTTP smoke on `localhost:8081`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| 2026-07-06 | Codex  | NADAA-021           | In Progress | Claimed area risk API; using seed-aligned fixtures until PostGIS persistence is wired into services.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| 2026-07-06 | Codex  | NADAA-101           | Done        | Verified `pnpm validate:docs`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, local `pnpm smoke:staging`, and local Docker builds for five deployable images.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| 2026-07-06 | Codex  | NADAA-101           | In Progress | Claimed CI/CD and staging foundation; credentials and provider deployment remain environment-owned, not committed.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| 2026-07-06 | Codex  | NADAA-033           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live duplicate candidate HTTP smoke on `localhost:8084`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| 2026-07-06 | Codex  | NADAA-033           | In Progress | Claimed incident deduplication baseline; focusing on candidate scoring/storage without merging or deleting reports.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| 2026-07-06 | Codex  | NADAA-032           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live citizen incident UI/API smoke on `localhost:8084`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 2026-07-06 | Codex  | NADAA-032           | In Progress | Claimed citizen incident reporting UI; integrating report submission with media upload initiation and existing incident-service API.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| 2026-07-06 | Codex  | NADAA-031           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live media upload/link smoke on `localhost:8084`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| 2026-07-06 | Codex  | NADAA-031           | In Progress | Claimed media upload flow; using controlled dev upload URLs and private in-memory metadata for this slice.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| 2026-07-06 | Codex  | NADAA-030           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live incident HTTP smoke on `localhost:8084`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| 2026-07-06 | Codex  | NADAA-030           | In Progress | Claimed incident reporting API; media upload storage and deduplication remain NADAA-031/NADAA-033.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| 2026-07-06 | Codex  | NADAA-010           | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and live auth HTTP smoke on `localhost:8082`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| 2026-07-06 | Codex  | NADAA-010           | In Progress | Claimed citizen authentication service slice; agency MFA remains NADAA-011.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| 2026-07-06 | Codex  | NADAA-100/NADAA-020 | Done        | Verified `pnpm validate:docs`, `pnpm typecheck`, `pnpm build`, `pnpm go:test`, `pnpm smoke:web`, and PostGIS migration/seed on `localhost:55432`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| 2026-07-06 | Codex  | NADAA-100/NADAA-020 | In Progress | Claimed QA matrix and PostGIS geospatial foundation.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| 2026-07-06 | Codex  | NADAA-002/NADAA-003 | Done        | Verified `pnpm validate:dashboard`, `pnpm typecheck`, `pnpm build`, and `pnpm go:test`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| 2026-07-06 | Codex  | NADAA-002/NADAA-003 | In Progress | Claimed documentation expansion and delivery dashboard contract.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| 2026-07-06 | Codex  | NADAA-001           | Done        | Verified `pnpm typecheck`, `pnpm build`, `pnpm go:test`, and app HTTP smoke checks on ports 5173 and 5174.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| 2026-07-06 | Codex  | NADAA-001           | In Progress | Claimed repository foundation and initial scaffold.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| 2026-07-06 | Codex  | agent_plan.md       | Done        | Initial plan created and expanded for multi-agent coordination.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |

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
  - `apps/citizen-mobile` - planned React Native citizen app for push alerts, GPS reporting, offline guides, shelters, and recovery support.
  - `apps/dispatcher-web` - dedicated dispatcher command console for incident intake, verification, duplicate review, assignment, status, maps, and ML review.
  - `apps/dispatcher-mobile` - planned React Native dispatcher app for urgent triage, assignment/status updates, and shift-based incident monitoring.
  - `apps/agency-web` - agency-scoped operations portal for assigned incidents, responder updates, capacity/shelter updates, relief context, and agency notes.
  - `apps/admin-web` - system/admin governance console for agencies, roles, users, MFA support, audit logs, data sources, alert rules, and platform configuration.
  - `apps/authority-dashboard` - current MVP compatibility shell; new authority work should migrate into `apps/dispatcher-web`, `apps/agency-web`, or `apps/admin-web` by role.
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

#### NADAA-004: Define Multi-Platform App Portfolio And Build Lanes

- User story: As the delivery team, we need clear app boundaries for citizens, dispatchers, agencies, and admins so each platform can be built without bloating a shared dashboard.
- Business value: Keeps role-specific workflows modular, makes multi-agent ownership clearer, and prevents web/mobile work from becoming tangled.
- Acceptance criteria:
  - Target platforms are documented as citizen web, citizen mobile, dispatcher web, dispatcher mobile, agency web, and admin web.
  - Each platform has a target app path, phase/sprint lane, and driving stories.
  - Current MVP authority-dashboard work has a documented migration path into dedicated dispatcher, agency, and admin apps.
- Tasks:
  - Add platform portfolio and platform build board to `agent_plan.md`.
  - Add platform delivery breakdown with ownership boundaries, shared dependencies, first build stories, and release gates.
  - Add platform build sequencing so app extraction and mobile work happen in the right order.
  - Update app workstream list with target app paths.
  - Add platform-specific stories to the MVP and Phase 2 trackers.
  - Update ready queue and ledger with the new build lanes.
- Estimate: 3 points.
- Dependencies: NADAA-001.

#### NADAA-015: Build Public Marketing Website

- User story: As a public stakeholder, partner, donor, agency leader, or citizen, I need a clear public website that explains what NADAA is, who it serves, what platforms exist, what services it provides, and how to get in touch.
- Business value: Gives NADAA a credible public entry point for awareness, stakeholder buy-in, partner onboarding, and platform education without exposing operational dashboards.
- Acceptance criteria:
  - `apps/marketing-web` exists as a dedicated React/Vite app with a thin root entrypoint and modular feature files.
  - The homepage summarizes the about story, benefits, features, services, platform portfolio, emergency contact guidance, partner contact path, and researched Ghana disaster context.
  - The design uses the real NADAA logo/brand assets, Outfit typography, and the approved navy, green, red, gold, slate, white, mist, and ink palette.
  - Public CTAs point to emergency `112`, citizen app, dispatcher/agency/admin platform lanes, and partner contact.
  - The website is wired into workspace scripts, smoke checks, staging smoke expectations, CI Docker build matrix, deployment docs, QA docs, and dashboard records.
  - No private operational data, fake live incident counts, or protected workflow controls are exposed.
- Tasks:
  - Research Ghana emergency/disaster context from official and credible public sources.
  - Scaffold `apps/marketing-web` with modular app, data, components, and styles.
  - Build the public website sections for hero, about, features, services, platforms, benefits, research context, and contact.
  - Add public assets, Dockerfile, workspace scripts, and smoke/staging checks.
  - Update README, deployment, QA, dashboard records, and plan ledger.
  - Verify typecheck, build, smoke, docs validation, formatting, and diff hygiene.
- Estimate: 5 points.
- Dependencies: NADAA-004, NADAA-002.

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

#### NADAA-014: Build Admin Web Governance Console

- User story: As a system/admin user, I need a dedicated web console so I can manage agencies, users, roles, MFA support, audit logs, data sources, alert rules, and platform configuration without entering dispatcher operations.
- Business value: Separates governance work from incident command, reduces accidental operational changes, and gives the platform a controlled administration surface.
- Acceptance criteria:
  - `apps/admin-web` exists as a dedicated React/Vite app with its own app shell, routing, session guard, and feature modules.
  - Only authorized `system_admin` and permitted admin roles can access the console.
  - Admin users can view/manage agencies, agency users, role assignments, MFA support state, audit logs, configured data sources, and alert-rule configuration where APIs exist.
  - Admin actions surface audit context and avoid exposing secrets or sensitive citizen report data.
  - Empty, loading, error, unauthorized, and success states are covered.
- Tasks:
  - Scaffold `apps/admin-web` using shared config, theme tokens, and shared types.
  - Build admin navigation and role-protected route/session shell.
  - Add agency/user/role management views from auth-service contracts.
  - Add audit log and data-source configuration views.
  - Add smoke/typecheck coverage and update deployment docs/scripts for the new app.
- Estimate: 13 points.
- Dependencies: NADAA-011, NADAA-012, NADAA-080.

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

#### NADAA-044: Create Dedicated Dispatcher Web Command Console

- User story: As a dispatcher, I need a dedicated web console for incident command so operational triage is separate from agency and admin workflows.
- Business value: Keeps high-pressure dispatch work focused, reduces accidental access to governance features, and creates a clean target for dispatcher mobile parity later.
- Acceptance criteria:
  - `apps/dispatcher-web` exists as a dedicated React/Vite app with thin `src/App.tsx`, `src/app/` shell code, and `src/features/dispatch-command/` modules.
  - Existing command-center capabilities from `apps/authority-dashboard` are available in dispatcher web: incident map/list, filters, verification/status, assignment, timeline, duplicate merge review, and abuse/false-report handling.
  - Dispatcher web uses dispatcher-appropriate RBAC and does not expose system-admin configuration or broad agency administration.
  - Shared UI/data/type modules are extracted only when reuse with agency/admin apps is real and bounded.
  - Smoke/typecheck/build coverage includes the new dispatcher web app.
- Tasks:
  - Scaffold `apps/dispatcher-web` and workspace scripts.
  - Move or reuse command-center feature modules from `apps/authority-dashboard`.
  - Keep `apps/authority-dashboard` as a compatibility shell or remove it only after routes/scripts/docs are updated.
  - Update docs, environment examples, smoke tests, and staging script coverage.
  - Add regression coverage for incident selection, verification/status, assignment, duplicate review, and abuse review flows.
- Estimate: 13 points.
- Dependencies: NADAA-040, NADAA-041, NADAA-042, NADAA-043, NADAA-091.

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
  - Create Go ML service.
  - Add model loading and prediction endpoint.
  - Add service-to-service contract.
  - Log predictions with `ml_predictions`-compatible metadata.
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
- NADAA-004 Define Multi-Platform App Portfolio And Build Lanes.
- NADAA-100 Build Test Strategy And QA Matrix.

Deliverables:

- Monorepo scaffold.
- Project docs.
- Initial API/security/ML/deployment plans.
- QA matrix.
- Platform portfolio and role-specific app boundaries.
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
- NADAA-044 Create Dedicated Dispatcher Web Command Console, initial extraction if capacity allows.
- NADAA-091 Implement Abuse, Spam, And False Report Handling, first pass.

Deliverables:

- Authority dashboard shell.
- Live incident map/list.
- Verification and status workflow.
- Agency assignment and timeline.
- Duplicate merge workflow.
- Dedicated dispatcher web app path and migration plan.

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
- NADAA-014 Build Admin Web Governance Console.
- NADAA-100 Build Test Strategy And QA Matrix, final pass.
- NADAA-101 Set Up CI/CD And Staging Environment, final pass.
- NADAA-102 Conduct UAT, Beta, And Production Readiness.

Deliverables:

- Security review and risk register.
- Dedicated admin web governance console.
- Staging release candidate.
- UAT scripts and sign-off package.
- Beta monitoring plan.
- Release notes, user guide, and hypercare checklist.

## Phase 2 Detailed Plan

Phase 2 extends the MVP from web-first emergency reporting and alerts into inclusive access, citizen/dispatcher mobile apps, agency-scoped operations, field coordination, recovery logistics, and richer operational data.

### EPIC 11: Inclusive Warning And Access Channels

Goal: reach citizens who do not have smartphones, stable internet, or high literacy, while keeping all public warning workflows auditable and authority-approved.

#### NADAA-110: SMS/USSD Emergency Access

- Outcome: citizens can check alerts, report basic incidents, and receive safety instructions through SMS/USSD.
- Acceptance criteria:
  - USSD menu supports language selection, current alerts, report emergency, shelter lookup, and 112 guidance.
  - SMS fallback can send alert summaries and basic report confirmations.
  - Messages are linked to citizen profiles when consent and phone identity are available.
  - Delivery failures and provider errors are logged.
  - Runtime logs expose `INFO`, `WARN`, and `ERROR` visibility for request receipt, validation, provider outcomes, menu transitions, report queueing/submission, and stored access log/report IDs without raw phone numbers or full message bodies.
- Tasks:
  - Select SMS/USSD provider abstraction.
  - Define USSD menu tree and message templates.
  - Implement inbound SMS/USSD webhook handling.
  - Map SMS/USSD reports into incident intake.
  - Add provider sandbox tests.
  - Add structured runtime logs for operational visibility and privacy-safe troubleshooting.
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

#### NADAA-113: Build Citizen Mobile App Foundation

- Outcome: citizens can use a native mobile app for alerts, GPS reporting, risk checks, offline guides, shelters, and recovery support.
- Acceptance criteria:
  - `apps/citizen-mobile` exists as a React Native app using shared types, config, design tokens, and API contracts from citizen web.
  - App supports authentication/session handling, current alert feed, incident report draft/submission, risk lookup, offline guide cache, shelter/recovery lookup, and push registration where provider configuration exists.
  - Offline and interrupted-network behavior is explicit for guides and report drafts.
  - Mobile permissions for location, media, and push notifications are requested with clear product copy.
  - Basic mobile smoke/build checks are documented and runnable locally/CI where possible.
- Tasks:
  - Choose React Native/Expo structure and workspace integration.
  - Scaffold `apps/citizen-mobile` with shared theme/config/types.
  - Build mobile navigation and session shell.
  - Port citizen web flows into native screens with offline draft/cache handling.
  - Wire push registration abstraction and permission copy.
  - Add build/typecheck/smoke documentation.
- Estimate: 21 points.
- Dependencies: NADAA-010, NADAA-021, NADAA-030, NADAA-052, NADAA-060, NADAA-061.

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

#### NADAA-124: Build Dispatcher Mobile Triage App

- Outcome: dispatchers can monitor urgent incidents, review core details, and make time-sensitive updates from a mobile device during shifts or field operations.
- Acceptance criteria:
  - `apps/dispatcher-mobile` exists as a React Native app with dispatcher authentication, MFA-aware session handling, and role-protected navigation.
  - App supports incident queue, selected incident details, status update, assignment handoff, critical alert/incident push notifications, and timeline notes.
  - Mobile app does not expose system-admin configuration, broad data-source management, or unrelated agency administration.
  - Offline refresh behavior is explicit, with clear stale-data indicators.
  - Mobile smoke/build checks and dispatcher safety test cases are documented.
- Tasks:
  - Scaffold `apps/dispatcher-mobile` with shared auth/config/types.
  - Build incident queue and selected incident views.
  - Add status, assignment, and timeline-note actions.
  - Wire push notification registration for critical incident escalation.
  - Add stale/offline indicators and mobile QA checklist.
- Estimate: 21 points.
- Dependencies: NADAA-041, NADAA-042, NADAA-091, NADAA-044.

#### NADAA-125: Build Agency Web Operations Portal

- Outcome: agency users can manage their assigned incidents, responder updates, capacity/shelter updates, and agency-scoped notes without entering dispatcher or admin consoles.
- Acceptance criteria:
  - `apps/agency-web` exists as a dedicated React/Vite app with agency-scoped route guards and feature modules.
  - Agency users can see only incidents assigned to their agency unless their role grants broader visibility.
  - Agency responders can update response status, add operational notes, and view relevant timeline/history.
  - Shelter, hospital capacity, relief, and road context appear when the corresponding APIs are available.
  - Agency web has empty/loading/error/unauthorized states and smoke/typecheck/build coverage.
- Tasks:
  - Scaffold `apps/agency-web` using shared config, theme tokens, and shared types.
  - Build agency session shell and assigned-incident dashboard.
  - Add responder status/timeline update views.
  - Add capacity/shelter/relief context modules as APIs become available.
  - Update deployment docs, smoke scripts, and RBAC tests for agency-scoped access.
- Estimate: 21 points.
- Dependencies: NADAA-011, NADAA-042, NADAA-062, NADAA-121.

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
- NADAA-111 WhatsApp Emergency Chatbot.
- NADAA-113 Build Citizen Mobile App Foundation, first pass.

Deliverables:

- SMS/USSD provider abstraction.
- USSD menu and inbound report flow.
- WhatsApp chatbot provider contract, conversation state, privacy-safe transcripts, and core intents.
- Citizen mobile app shell, shared contracts, and core navigation.

#### Sprint 9: Multilingual Alerts And Chat Completion

- NADAA-112 Multilingual Voice Alerts.
- NADAA-113 Build Citizen Mobile App Foundation, completion.

Deliverables:

- WhatsApp incident/report/shelter flows.
- Voice alert review and delivery workflow.
- Multilingual alert template library.
- Citizen mobile alerts, risk, reporting, offline guides, and push registration foundation.

#### Sprint 10: Field And Health Operations

- NADAA-120 Community Volunteer App.
- NADAA-121 Hospital Capacity Tracker.
- NADAA-124 Build Dispatcher Mobile Triage App.
- NADAA-125 Build Agency Web Operations Portal, first pass.

Deliverables:

- Volunteer registration, verification, and assignment.
- Hospital capacity map/list and update workflow.
- Dispatcher mobile triage queue and urgent update flow.
- Agency web shell and assigned-incident dashboard.

#### Sprint 11: Relief, Aid, And Mobility Data

- NADAA-122 Relief Distribution Tracking.
- NADAA-123 Donation And Aid Coordination.
- NADAA-125 Build Agency Web Operations Portal, completion.
- NADAA-131 Road Closure Integration.

Deliverables:

- Relief point management.
- Aid request and pledge tracking.
- Agency web capacity, shelter, relief, and road context modules.
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

| ID        | Phase | Sprint                     | Story                                                       | Status | Owner | Branch/PR | Notes                                                                                                                                                                                                                                     |
| --------- | ----- | -------------------------- | ----------------------------------------------------------- | ------ | ----- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| NADAA-001 | MVP   | Sprint 0                   | Create Repository And Monorepo Foundation                   | Done   | Codex | main      | Monorepo foundation, starter apps, risk service, infra, and docs created.                                                                                                                                                                 |
| NADAA-002 | MVP   | Sprint 0                   | Define Product, API, Security, ML, And Deployment Docs      | Done   | Codex | main      | Docs expanded and linked from README.                                                                                                                                                                                                     |
| NADAA-003 | MVP   | Sprint 0                   | Create Delivery Dashboard Data Contract                     | Done   | Codex | main      | Schema, sample records, and validation script added.                                                                                                                                                                                      |
| NADAA-004 | MVP   | Sprint 0                   | Define Multi-Platform App Portfolio And Build Lanes         | Done   | Codex | main      | Planned citizen web/mobile, dispatcher web/mobile, agency web, and admin web lanes with target paths, phase placement, driving stories, ownership boundaries, dependencies, sequence, and gates.                                          |
| NADAA-010 | MVP   | Sprint 1                   | Implement Citizen Authentication                            | Done   | Codex | main      | Citizen register/login/profile API and tests added in auth-service.                                                                                                                                                                       |
| NADAA-011 | MVP   | Sprint 1                   | Implement Agency Users, Roles, And MFA                      | Done   | Codex | main      | Agency user creation, authority role catalog, mock MFA setup/verification, agency login, MFA-aware tokens, shared types, docs, and RBAC tests added.                                                                                      |
| NADAA-012 | MVP   | Sprint 1                   | Implement Audit Logging Foundation                          | Done   | Codex | main      | Auth-service audit model, helper, auth/admin event wiring, system-admin audit read endpoint, shared types, retention docs, and tests added.                                                                                               |
| NADAA-014 | MVP   | Sprint 7                   | Build Admin Web Governance Console                          | Done   | Codex | main      | Dedicated `apps/admin-web` with admin RBAC/MFA shell, governance views, safe API fixture fallbacks, Outfit typography, scripts, docs, CI/staging smoke, and Docker build verification.                                                    |
| NADAA-015 | MVP   | Sprint 7/Phase 2 Sprint 11 | Build Public Marketing Website                              | Done   | Codex | main      | Public `apps/marketing-web` completed for about, platform summary, features, services, benefits, Ghana disaster context, contact, workspace scripts, smoke/docs/dashboard/CI/Docker wiring, and verification.                             |
| NADAA-020 | MVP   | Sprint 1                   | Set Up PostGIS And Core Geospatial Models                   | Done   | Codex | main      | Migration and seed verified against local PostGIS on port 55432.                                                                                                                                                                          |
| NADAA-021 | MVP   | Sprint 5                   | Implement Area Risk API                                     | Done   | Codex | main      | Fixture-backed API returns low/high/severe flood scoring, nearby shelters, nearby facilities, recommended actions, and coordinate validation.                                                                                             |
| NADAA-022 | MVP   | Sprint 5                   | Build Citizen Risk Checker UI                               | Done   | Codex | main      | Risk checker UI integrates the risk API, area presets, coordinate entry, GPS lookup, shelters, facilities, recommended actions, loading/error/permission/empty states, and risk smoke coverage.                                           |
| NADAA-030 | MVP   | Sprint 2                   | Implement Incident Reporting API                            | Done   | Codex | main      | Incident intake API and tests added in incident-service.                                                                                                                                                                                  |
| NADAA-031 | MVP   | Sprint 2                   | Implement Media Upload Flow                                 | Done   | Codex | main      | Controlled upload initiation and private media linkage added in incident-service.                                                                                                                                                         |
| NADAA-032 | MVP   | Sprint 2                   | Build Citizen Incident Reporting UI                         | Done   | Codex | main      | Citizen incident form integrated with media upload initiation, GPS/manual location, validation, offline retry messaging, and API success/error states.                                                                                    |
| NADAA-033 | MVP   | Sprint 2                   | Add Incident Deduplication Baseline                         | Done   | Codex | main      | Same-hazard duplicate candidates are scored by location distance, report time, and description similarity; reports remain reviewable and are never automatically merged or deleted.                                                       |
| NADAA-040 | MVP   | Sprint 3                   | Build Incident Command Map                                  | Done   | Codex | main      | Leaflet command map, API-backed incident feed with fixture fallback, map/list sync, filters, selected-incident detail, role-protected framing, docs, and UI states added.                                                                 |
| NADAA-041 | MVP   | Sprint 3                   | Implement Verification And Status Workflow                  | Done   | Codex | main      | Incident-service transition rules, verification/status endpoints, audited status changes, closure/false-report notes, authority dashboard controls, shared types, docs, smoke, tests, and Docker build added.                             |
| NADAA-042 | MVP   | Sprint 3                   | Implement Agency Assignment And Incident Timeline           | Done   | Codex | main      | Incident-service assignment/timeline models, assignment endpoint, assigned-agency filtering, authority dashboard assignment controls, docs, smoke, tests, and Docker build added.                                                         |
| NADAA-043 | MVP   | Sprint 3                   | Implement Duplicate Merge Review                            | Done   | Codex | main      | Duplicate review and merge endpoints, traceability fields, audit/timeline events, side-by-side authority dashboard UI, shared types, docs, smoke, tests, and Docker build added.                                                          |
| NADAA-044 | MVP   | Sprint 3/7                 | Create Dedicated Dispatcher Web Command Console             | Done   | Codex | main      | Dedicated `apps/dispatcher-web` app added with dispatch-command modules, workspace scripts, smoke/staging checks, CI Docker build, docs, dashboard record, and compatibility shell retained.                                              |
| NADAA-050 | MVP   | Sprint 4                   | Implement Alert Creation And Approval Workflow              | Done   | Codex | main      | Alert-service workflow API, emergency override path, RBAC/MFA/audit hooks, authority dashboard alert form/queue, shared types, docs, smoke, CI, tests, and Docker build added.                                                            |
| NADAA-051 | MVP   | Sprint 4                   | Implement Geofenced Alert Targeting                         | Done   | Codex | main      | Alert-service target geometry handling, preview endpoint, target query filters, district/radius/custom validation, authority dashboard selectors/preview, shared types, docs, smoke, tests, and Docker build added.                       |
| NADAA-052 | MVP   | Sprint 4                   | Implement In-App Alert Feed And Push/SMS Abstraction        | Done   | Codex | main      | Notification-service citizen feed API, mock/disabled push/SMS providers, delivery logs, delivery-log schema, citizen current/expired feed UI, shared types, docs, smoke, tests, CI, and Docker build added.                               |
| NADAA-060 | MVP   | Sprint 5                   | Implement Emergency Guide Content Model                     | Done   | Codex | main      | Guide-service API, guide content fixtures, seed expansion, shared guide types, docs, Docker/CI wiring, staging smoke wiring, and lookup tests added.                                                                                      |
| NADAA-061 | MVP   | Sprint 5                   | Build Offline-First Citizen Guidance UI                     | Done   | Codex | main      | Citizen guide browser integrates guide-service, local offline guide cache, service-worker guide/app-shell caching, hazard/stage/language filters, English fallback, visible 112 CTA, docs, smoke, and tests.                              |
| NADAA-062 | MVP   | Sprint 5                   | Implement Shelter And Recovery Support Module               | Done   | Codex | main      | Shelter-service nearby lookup, recovery support locations, protected capacity updates, citizen shelter map/list, authority capacity update view, shared types, docs, smoke, tests, CI, and Docker build added.                            |
| NADAA-070 | MVP   | Sprint 6                   | Create Flood Risk Dataset And Feature Pipeline              | Done   | Codex | main      | Versioned fixture inputs, 44-column schema, deterministic generation/validation scripts, JSON/CSV outputs, manifest checksums, docs, dashboard record, and verification added.                                                            |
| NADAA-071 | MVP   | Sprint 6                   | Train Baseline Flood Risk Model                             | Done   | Codex | main      | Deterministic logistic-regression trainer, model metadata, sample predictions, evaluation report, medium fixture confidence, validation, docs, dashboard record, and verification added.                                                  |
| NADAA-072 | MVP   | Sprint 6                   | Serve ML Predictions Through Risk API                       | Done   | Codex | main      | ML service prediction/log endpoints, risk-service decision-support enrichment, shared contracts, no-auto-publish safety flags, smoke/staging/CI/Docker wiring, docs, dashboard record, and verification added.                            |
| NADAA-073 | MVP   | Sprint 6                   | Add Authority ML Review View                                | Done   | Codex | main      | Dispatcher-web ML prediction review map/list, explanation panel, reviewed alert-draft action, structured source-prediction metadata, smoke/docs/dashboard records, and verification added.                                                |
| NADAA-080 | MVP   | Sprint 6                   | Define Agency Integration Contracts                         | Done   | Codex | main      | Integration matrix, inbound weather/hydrology contracts, outbound incident/alert sync contracts, mock integration service, shared types, docs, CI/staging wiring, smoke script, Dockerfile, and tests added.                              |
| NADAA-081 | MVP   | Sprint 6                   | Implement Weather And Hydrology Import Skeleton             | Done   | Codex | main      | Integration-service fixture importer, imported observation store, import status logs, retry flow, scheduled hook, shared types, docs, smoke, and tests added.                                                                             |
| NADAA-090 | MVP   | Sprint 2                   | Implement Location Privacy And Anonymous Reporting Controls | Done   | Codex | main      | Authority-only incident list, server-side privacy metadata, reporter/contact sanitization, duplicate/merge response minimization, command UI indicators, citizen privacy copy, docs, tests, and smoke coverage.                           |
| NADAA-091 | MVP   | Sprint 3                   | Implement Abuse, Spam, And False Report Handling            | Done   | Codex | main      | Suspicious report signals, abuse review endpoint, false-report workflow, authority dashboard moderation controls, shared types, docs, smoke, tests, and Docker build added.                                                               |
| NADAA-092 | MVP   | Sprint 7                   | Security Review And Hardening                               | Done   | Codex | main      | Runtime HTTP hardening, CORS allowlist support, defensive API headers, security scan automation, non-root web containers, dependency scan, residual-risk register, docs, dashboard record, and verification added.                        |
| NADAA-100 | MVP   | Sprint 0/7                 | Build Test Strategy And QA Matrix                           | Done   | Codex | main      | Sprint 0 QA matrix complete; final pass remains in Sprint 7.                                                                                                                                                                              |
| NADAA-101 | MVP   | Sprint 1/7                 | Set Up CI/CD And Staging Environment                        | Done   | Codex | main      | CI workflow, Docker build validation, manual staging smoke workflow/script, staging env template, and runbook added; registry push/deploy credentials remain environment-owned.                                                           |
| NADAA-102 | MVP   | Sprint 7                   | Conduct UAT, Beta, And Production Readiness                 | Done   | Codex | main      | UAT scripts, feedback capture process, beta metrics dashboard definition, release notes template, user/training guide, acceptance checklist, hypercare checklist, release-readiness validation, dashboard record, and verification added. |

### Phase 2 Tracker

| ID        | Phase   | Sprint       | Story                                      | Status | Owner      | Branch/PR                                   | Notes                                                                                                                                                                                                                                                                                                                                                            |
| --------- | ------- | ------------ | ------------------------------------------ | ------ | ---------- | ------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| NADAA-110 | Phase 2 | Sprint 8     | SMS/USSD Emergency Access                  | Done   | Codex      | main                                        | Notification-service SMS/USSD provider webhooks, language menu, alert summaries, basic report intake, 112 guidance, profile-link consent, structured INFO/WARN/ERROR runtime logging, optional incident-service handoff, shared types, smoke/docs/dashboard updates, and verification added.                                                                     |
| NADAA-111 | Phase 2 | Sprint 8/9   | WhatsApp Emergency Chatbot                 | Done   | Codex      | main                                        | Notification-service WhatsApp webhook aliases, alerts/risk/report/shelter/guide/112 intents, multi-message report state, location/media incident handoff, privacy-safe transcript summaries, shared types, smoke/docs/dashboard updates, structured logs, and verification added.                                                                                |
| NADAA-112 | Phase 2 | Sprint 9     | Multilingual Voice Alerts                  | Done   | Codex      | main                                        | Notification-service voice asset generation, multilingual low-literacy templates, review/approval workflow, approved voice delivery logs, shared types, smoke/docs/dashboard/staging updates, structured logs, and verification added.                                                                                                                           |
| NADAA-113 | Phase 2 | Sprint 8/9   | Build Citizen Mobile App Foundation        | Done   | Codex      | main                                        | `apps/citizen-mobile` Expo/React Native foundation, shared contracts/theme/config, mobile session/navigation, citizen flows, offline draft/cache primitives, permission copy, smoke/docs/dashboard updates, lockfile integration, and verification added.                                                                                                        |
| NADAA-120 | Phase 2 | Sprint 10    | Community Volunteer App                    | Done   | Codex      | main                                        | Incident-service volunteer profile/verification model, response groups, volunteer task assignment API, mobile Community task view, safety/escalation rules, audit/timeline events, shared contracts, smoke/docs/dashboard updates, and structured logs added.                                                                                                    |
| NADAA-121 | Phase 2 | Sprint 10    | Hospital Capacity Tracker                  | Done   | Codex      | main                                        | Shelter-service hospital/facility capacity schema, manual and fixture update workflow, dispatcher capacity filters/list, stale-data warnings, source-tracked updates, smoke/docs/dashboard updates, structured logs, and verification completed.                                                                                                                 |
| NADAA-124 | Phase 2 | Sprint 10    | Build Dispatcher Mobile Triage App         | Done   | Codex      | feature/NADAA-124-dispatcher-mobile-triage  | `apps/dispatcher-mobile` Expo/RN foundation, agency session/MFA shell, incident queue/detail/action/capacity/profile screens, status/assignment/timeline-note actions, hospital capacity lookup, offline/stale indicators, sandbox push registration, smoke script, docs, dashboard record, and verification completed.                                          |
| NADAA-125 | Phase 2 | Sprint 10/11 | Build Agency Web Operations Portal         | Done   | Codex      | main                                        | `apps/agency-web` React/Vite operations portal with agency session/MFA shell, assigned-incident dashboard, responder status/timeline-note updates, shelter/hospital capacity context, occupancy/capacity update forms, smoke-web/staging-smoke integration, Dockerfile, CI Docker matrix, docs/dashboard updates, and verification completed.                    |
| NADAA-122 | Phase 2 | Sprint 11    | Relief Distribution Tracking               | Done   | Codex      | main                                        | Shelter-service relief point list/nearby/create/update/stock-history APIs, agency web relief management tab, citizen web nearby relief display, smoke/docs/dashboard updates, structured INFO/WARN/ERROR logs, and verification completed.                                                                                                                       |
| NADAA-123 | Phase 2 | Sprint 11    | Donation And Aid Coordination              | Done   | Codex      | main                                        | Shelter-service aid request/pledge APIs, PostGIS migration, agency web Aid tab, public approved-needs listing, donor pledge flow, CSV export, smoke/docs/dashboard updates, structured INFO/WARN/ERROR logs, and verification completed.                                                                                                                         |
| NADAA-171 | Phase 2 | Sprint 11    | Refactor Go Services Into Modular Packages | Done   | Codex      | feature/NADAA-171-go-service-modularization | Split every Go service `main.go` monolith into `cmd/server`, `internal/config`, `internal/models`, `internal/store`, `internal/utils`, and `internal/handlers`; moved tests into `internal/handlers`; updated Dockerfiles; also finalized pending `apps/admin-web` modular refactor. Full verification passed.                                                   |
| NADAA-130 | Phase 2 | Sprint 12    | Evacuation Route Planner                   | Done   | Kimi       | feature/NADAA-130-evacuation-route-planner  | services/route-service Go API with in-memory route planning, closure/risk avoidance, shelter-service integration, decision-support disclaimer, tests, Dockerfile, README, dashboard record, CI Docker matrix entry, and smoke script. Citizen-web and authority-dashboard route planner UI with accessible forms and Leaflet polyline. Full verification passed. |
| NADAA-131 | Phase 2 | Sprint 11    | Road Closure Integration                   | Done   | Codex      | main                                        | `services/road-closure-service` Go API with list/create/update/adapter-import, database migration, shared TypeScript contracts, dispatcher-web map polyline layer, agency-web nearby closure context, citizen-web closure cards, integration-service adapter import endpoint, smoke script, CI Docker matrix, docs, dashboard record, and verification.          |
| NADAA-132 | Phase 2 | Sprint 12    | Missing Persons Module                     | Done   | Kimi       | main                                        | Completed missing-person-service private intake, authority review/public visibility, reunification/closure audit trail, public approved search, shared contracts, citizen/authority UI, migration, smoke/docs/dashboard updates, structured logs, Docker/CI wiring, and verification.                                                                             |
| NADAA-133 | Phase 2 | Sprint 12    | Insurance And Property Damage Claim Export | Todo   | Unassigned | TBD                                         | Recovery case export.                                                                                                                                                                                                                                                                                                                                            |
| NADAA-140 | Phase 2 | Sprint 12    | Drone And Satellite Image Ingestion        | Todo   | Unassigned | TBD                                         | Supports verification and ML.                                                                                                                                                                                                                                                                                                                                    |

### Phase 3 Tracker

| ID        | Phase   | Sprint       | Story                                                 | Status | Owner      | Branch/PR                            | Notes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| --------- | ------- | ------------ | ----------------------------------------------------- | ------ | ---------- | ------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| NADAA-150 | Phase 3 | Sprint 13/14 | Real-Time Flood Simulation                            | Todo   | Unassigned | TBD                                  | Spans two sprints.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| NADAA-151 | Phase 3 | Sprint 13    | AI Incident Triage                                    | Done   | Claude     | feature/NADAA-151-ai-incident-triage | Human-supervised only. Explainable rules-based suggestions (severity, duplicate likelihood, affected population, agency routing) via incident-service `GET /incidents/{id}/triage` and `POST /incidents/{id}/triage-review` with accept/override timeline + audit logging, dispatcher-web AI triage panel with editable override form, shared types, `pnpm smoke:ai-triage`, api.md endpoint docs, and ml.md triage model + bias/error review process. Claude took over from Kimi's first pass, fixed review-response parsing, override severity options, and accept-path logging, and ran full workspace verification. |
| NADAA-152 | Phase 3 | Sprint 14    | Computer Vision For Flood And Fire Image Verification | Todo   | Unassigned | TBD                                  | Supports verification, not auto-escalation.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| NADAA-153 | Phase 3 | Sprint 14    | Predictive Ambulance And Fire Station Positioning     | Todo   | Unassigned | TBD                                  | Agency decision support.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| NADAA-160 | Phase 3 | Sprint 15    | School Emergency Preparedness Module                  | Todo   | Unassigned | TBD                                  | District readiness.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| NADAA-161 | Phase 3 | Sprint 15    | Public Disaster Education Campaigns                   | Todo   | Unassigned | TBD                                  | Seasonal campaigns.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| NADAA-162 | Phase 3 | Sprint 15/16 | National Open Disaster Data Portal                    | Todo   | Unassigned | TBD                                  | Requires privacy governance.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| NADAA-163 | Phase 3 | Sprint 16    | Telecom Cell Broadcast Integration                    | Todo   | Unassigned | TBD                                  | Depends on official telecom path.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| NADAA-170 | Phase 3 | Sprint 16    | National-Scale Resilience And Operations Hardening    | Todo   | Unassigned | TBD                                  | Load, observability, DR.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |

## Initial Ready Queue

Start here:

1. NADAA-130 Evacuation Route Planner.

## Key Risks And Early Decisions

- Official agency data access may lag development. Use fixture/mock adapters early and isolate integration contracts.
- Flood model quality depends on data availability. Start with transparent rule-based scoring and baseline ML before advanced models.
- Public alert misuse is high risk. Require RBAC, MFA, approval workflow, audit logs, and emergency override controls.
- Citizen reporting can attract spam or false reports. Use rate limits, duplicate detection, suspicious report flags, and human review.
- Offline and low-literacy needs are important. Keep PWA/offline guidance in MVP and plan voice/USSD/WhatsApp for phase 2.
- Geospatial targeting must be correct. Prioritize PostGIS indexes, district boundaries, geometry validation, and map QA.
- Role-specific apps can drift if they fork too much shared behavior. Keep shells separate but share tokens, types, API clients, and domain utilities only where reuse is real.

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

| 2026-07-08 | Completed NADAA-174 extension of the design system to mobile apps. Added `@nadaa/brand/native` with `nativeTheme`, `severityBadgeFor`, and `hazardBadgeFor`. Rewrote `citizen-mobile` and `dispatcher-mobile` themes to consume shared tokens, replaced hard-coded colors/spacing, and added accessible icon+text badges. Ensured WCAG 2.1 AA contrast and 44 px touch targets. Full workspace verification passed. | Kimi | Complete |
| 2026-07-08 | Completed NADAA-130 evacuation route planner. Added `services/route-service` Go API with /routes/plan, /routes/options, /health, shelter/higher-ground/manual waypoints, closure/risk avoidance, decision-support disclaimer, tests, Dockerfile, README, dashboard record, CI Docker matrix entry, and smoke script. Added shared TypeScript route types, citizen-web route planner UI, and authority-dashboard route planner panel with Leaflet polyline. Full workspace verification passed. | Kimi | Complete |
| 2026-07-07 | Completed NADAA-172 Go service quality and graceful shutdown pass. Added `http.Server` graceful shutdown with timeouts and signal handling to all 10 services, package `doc.go` comments, `.golangci.yml` v2 quality baseline, exported-symbol and lint fixes, expanded `docs/user-guide.md` and `docs/architecture.md`, and full workspace verification. | Kimi | Complete |
| 2026-07-08 | Completed NADAA-173 cross-platform UI/UX redesign and design-system hardening. Extended `packages/brand` with semantic tokens, typography, spacing, elevation, breakpoints, severity/hazard roles, `createNadaaTheme`, and `brand.css`. Refactored citizen-web, dispatcher-web, authority-dashboard, agency-web, admin-web, and marketing-web to use shared tokens, canonical MUI theme, consistent CSS, skip links, focus-visible outlines, accessible icon+text chips, form/table ARIA associations, and responsive layouts. Added `docs/design-system.md`. Full workspace verification passed. | Kimi | Complete |

| Date       | Update                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | Owner | Status   |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----- | -------- |
| 2026-07-07 | Completed NADAA-123 with shelter-service aid request/pledge APIs, PostGIS aid coordination migration, public approved-needs listing, authority create/review gates, donor pledge intake/review, active pledge summaries, CSV export, anti-fraud notes, shared TypeScript contracts, agency web Aid tab, smoke/docs/dashboard updates, structured INFO/WARN/ERROR logs, live aid/shelter/relief smoke coverage, Docker build, and full verification.                                                                                                                                                                                                                                                                                                                         | Codex | Complete |
| 2026-07-07 | Completed NADAA-171 by refactoring all 10 Go services from single `main.go` monoliths into `cmd/server` plus `internal/config`, `internal/models`, `internal/store`, `internal/utils`, and `internal/handlers` packages; moved tests into `internal/handlers` under `package handlers`; updated Dockerfiles to build `./cmd/server`; preserved env vars, routes, CORS, security headers, and behavior. Also finalized the pending `apps/admin-web` modular refactor into `api/`, `auth/`, `components/`, `data/`, `lib/`, `pages/`, and `routes.tsx`. Verified `pnpm go:test`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm audit --audit-level high`, `pnpm exec prettier --check`, and `git diff --check`. | Codex | Complete |
| 2026-07-07 | Completed NADAA-122 with shelter-service relief point list/nearby/create/update/stock-history APIs, shared TypeScript contracts, citizen web nearby relief distribution display, agency web relief management tab with create/update/history workflow, smoke/docs/dashboard/integration updates, structured INFO/WARN/ERROR logs, and verification.                                                                                                                                                                                                                                                                                                                                                                                                                         | Codex | Complete |
| 2026-07-07 | Completed NADAA-125 with `apps/agency-web` React/Vite operations portal, agency session/MFA shell, assigned-incident dashboard, responder status/timeline-note updates, shelter/hospital capacity context, occupancy/capacity update forms, smoke-web/staging-smoke integration, Dockerfile, CI Docker matrix entry, `dev:agency` script, docs/dashboard/plan updates, formatting fixes, and verification.                                                                                                                                                                                                                                                                                                                                                                  | Codex | Complete |
| 2026-07-07 | Completed NADAA-121 with shelter-service hospital/facility capacity records, stale-data flags, distance/service/capacity/min-bed filters, manual authority updates, fixture adapter imports, source tracking, structured INFO/WARN/ERROR logs, shared TypeScript contracts, dispatcher-web hospital capacity panel, API/integration/deployment/QA docs, dashboard record, live smoke coverage, and verification.                                                                                                                                                                                                                                                                                                                                                            | Codex | Complete |
| 2026-07-06 | Completed NADAA-062 with dedicated shelter-service nearby lookup and recovery support locations, protected authority occupancy updates, shared shelter/recovery types, citizen shelter map/list and recovery support panel, authority shelter capacity update view, API/security/architecture/deployment/QA docs, dashboard record, smoke script, CI/Docker wiring, tests, and verification.                                                                                                                                                                                                                                                                                                                                                                                | Codex | Complete |
| 2026-07-06 | Completed NADAA-090 with authority-only incident list access, server-side reporter/contact privacy metadata and sanitization across list, duplicate, merge, verification, status, abuse, and assignment responses, dispatcher/authority privacy indicators, citizen report privacy copy, API/security/QA docs, dashboard record, focused tests, and incident smoke coverage.                                                                                                                                                                                                                                                                                                                                                                                                | Codex | Complete |
| 2026-07-06 | Completed NADAA-014 with modular `apps/admin-web`, system-admin/MFA access shell, agency/user/role management views, MFA support queue, audit log view, integration data-source view, alert-rule view, safe fixture fallbacks for unavailable APIs, Outfit typography across web apps, workspace scripts, CI/staging smoke wiring, deployment/QA/architecture docs, dashboard record, Docker build, and verification.                                                                                                                                                                                                                                                                                                                                                       | Codex | Complete |
| 2026-07-06 | Completed NADAA-044 with dedicated `apps/dispatcher-web`, dispatcher session/header/theme shell, `features/dispatch-command` modules, preserved incident command/alert workflows, authority-dashboard compatibility, workspace scripts, CI Docker build, local/staging smoke wiring, docs, dashboard record, lockfile update, and verification.                                                                                                                                                                                                                                                                                                                                                                                                                             | Codex | Complete |
| 2026-07-06 | Added the target platform portfolio for citizen web/mobile, dispatcher web/mobile, agency web, and admin web; added platform build board rows, delivery breakdown, ownership boundaries, dependencies, release gates, build sequence, MVP stories for admin and dispatcher web, Phase 2 stories for citizen mobile, dispatcher mobile, and agency web, plus updated sprint plans and ready queue.                                                                                                                                                                                                                                                                                                                                                                           | Codex | Complete |
| 2026-07-06 | Completed NADAA-041 with incident-service status transition rules, verification endpoint, RBAC/MFA gates, audited before/after status changes, terminal closure/false-report resolution notes, shared workflow types, authority dashboard controls, docs, smoke script, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | Codex | Complete |
| 2026-07-07 | Completed NADAA-131 with road-closure-service Go API (list/create/update/adapter-import), PostGIS migration, shared TypeScript contracts, dispatcher-web map polyline layer, agency-web nearby closure context, citizen-web closure cards, integration-service road-closure adapter import endpoint that forwards to `NADAA_ROAD_CLOSURE_SERVICE_URL` (default `http://localhost:8095`), smoke scripts, CI Docker matrix entries, docs, dashboard record, and plan updates. Verifications passed.                                                                                                                                                                                                                                                                           | Codex | Complete |
| 2026-07-06 | Completed NADAA-040 with authority-dashboard Leaflet incident command map, API-backed incident feed with fixture fallback, map/list synchronization, hazard/district/severity/status/time filters, selected-incident detail, role-protected framing, docs, and UI states.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | Codex | Complete |
| 2026-07-06 | Completed NADAA-012 with auth-service audit event model, in-memory audit store/helper, system-admin audit read endpoint, auth/admin event wiring, metadata capture, sanitized before/after snapshots, shared audit types, retention docs, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Codex | Complete |
| 2026-07-06 | Completed NADAA-011 with auth-service agency user creation, authority role catalog, mock MFA setup/verification, agency login, MFA-aware bearer tokens, shared agency auth types, API/security docs, and RBAC tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Codex | Complete |
| 2026-07-06 | Completed NADAA-022 with citizen risk checker API integration, area presets, coordinate entry, GPS lookup, shelters/facilities rendering, loading/error/permission/empty states, and risk smoke coverage.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | Codex | Complete |
| 2026-07-06 | Completed NADAA-021 with fixture-backed risk-service area lookup, baseline flood scoring, nearby shelter/facility aggregation, coordinate validation, shared response types, API docs, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Codex | Complete |
| 2026-07-06 | Completed NADAA-101 with GitHub Actions CI, Docker build validation for deployable apps/services, manual staging smoke workflow, staging smoke script, staging environment template, and staging runbook.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | Codex | Complete |
| 2026-07-06 | Completed NADAA-031 with controlled media upload initiation, private metadata, content-type and size validation, incident media linkage, shared media types, API docs, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Codex | Complete |
| 2026-07-06 | Completed NADAA-030 with incident-service report intake API, validation, anonymous/contact-permission behavior, media references, priority review flag, rate limiting, shared types, API docs, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | Codex | Complete |
| 2026-07-06 | Completed NADAA-010 with auth-service citizen registration, mock OTP login, signed bearer token profile access, shared auth types, API docs, and tests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Codex | Complete |
| 2026-07-06 | Completed NADAA-100 and NADAA-020 with QA strategy, MVP test matrix, smoke script, PostGIS schema, geospatial indexes, seed data, database docs, configurable local ports, and database validation.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Codex | Complete |
| 2026-07-06 | Completed NADAA-002 and NADAA-003 with expanded product/API/security/ML/deployment docs, dashboard tracking schema, sample records, and validation script.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Codex | Complete |
| 2026-07-06 | Completed NADAA-001 initial monorepo foundation with React/Vite citizen and authority apps, shared brand/types packages, Go risk service, local infra, docs, dependency lockfile, and Git remote setup.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Codex | Complete |
| 2026-07-06 | Expanded Phase 2 and Phase 3 into detailed epics, stories, tasks, sprint plans, and master status trackers.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Codex | Complete |
| 2026-07-06 | Created initial agent plan from repository documents.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | Codex | Complete |

| 2026-07-07 | Started NADAA-172 Go service quality and graceful shutdown pass. Phase 1 complete: all 10 services use `http.Server` with timeouts and signal-based shutdown. Added package `doc.go` comments to every service package. Restored `guide-service` quality template with context propagation and named constants. Reverted the partial `packages/go-common` extraction so shared-package work stays in its own story. Full workspace verification passed: `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm go:test`, `pnpm validate:docs`, `pnpm security:scan`, `pnpm audit --audit-level high`, `pnpm exec prettier --check`, and `git diff --check`. Remaining: exported-symbol docs, named constants, context propagation, and golangci-lint fixes. | Kimi | In Progress |
