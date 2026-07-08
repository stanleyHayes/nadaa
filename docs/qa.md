# QA Strategy

NADAA QA must validate product behavior, safety gates, and operational readiness. The test strategy starts lightweight in Sprint 0 and becomes stricter as public-safety workflows move toward UAT.

## Test Levels

| Level             | Purpose                                                   | Owner               | Runs                   |
| ----------------- | --------------------------------------------------------- | ------------------- | ---------------------- |
| Type checks       | Catch contract and TypeScript errors                      | Developers/agents   | Every change           |
| Unit tests        | Validate service/domain behavior                          | Developers/agents   | Every story            |
| API tests         | Validate request/response and authorization behavior      | Backend agents/QA   | Service stories        |
| E2E/smoke tests   | Validate critical user journeys                           | Frontend agents/QA  | Release candidates     |
| Security tests    | Validate auth, RBAC, MFA, audit, rate limits, and secrets | Engineering lead/QA | Sensitive changes      |
| Performance tests | Validate high-volume report and alert conditions          | Platform/QA         | Before beta/production |
| UAT scripts       | Validate stakeholder acceptance                           | PM/QA/stakeholders  | UAT                    |

## MVP Test Matrix

| Flow                          | Primary Stories                            | Test Type             | Acceptance Focus                                                                                    | Status                                                                                                                |
| ----------------------------- | ------------------------------------------ | --------------------- | --------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| Citizen risk check            | NADAA-021, NADAA-022                       | API, E2E              | Location lookup returns risk, shelters, guidance, loading/error states                              | MVP API/UI smoke covered                                                                                              |
| Agency auth, roles, and MFA   | NADAA-011                                  | API, security         | Agency membership, role catalog, MFA-required login, admin-only creation, forbidden role checks     | MVP API/security tests covered                                                                                        |
| Audit logging foundation      | NADAA-012                                  | API, security         | Actor/action/target metadata, before/after snapshots, RBAC denial events, system-admin audit read   | MVP API/security tests covered                                                                                        |
| Citizen incident report       | NADAA-030, NADAA-032                       | API, E2E              | GPS, hazard type, description, urgency, affected people, contact permission                         | MVP API/UI smoke covered                                                                                              |
| Media upload                  | NADAA-031                                  | API, security         | File type/size validation, private storage, incident linkage                                        | MVP API/UI smoke covered                                                                                              |
| Anonymous report privacy      | NADAA-090                                  | API, security         | Identity hidden where policy allows, contact permission honored                                     | MVP API/UI smoke covered                                                                                              |
| Public marketing website      | NADAA-015                                  | Type, Smoke, E2E      | About, features, services, platform lanes, benefits, emergency contact, research-source links       | MVP marketing web typecheck/build/smoke covered                                                                       |
| Duplicate detection           | NADAA-033, NADAA-043                       | Unit, API             | Nearby/time-window duplicate candidates are reviewable, mergeable, and traceable                    | MVP API/UI smoke covered                                                                                              |
| Authority incident map        | NADAA-040                                  | E2E                   | Filters, map/list sync, role-protected access, loading/empty/error fallback                         | MVP UI smoke covered                                                                                                  |
| Incident verification/status  | NADAA-041                                  | API, E2E, audit       | Valid transitions, invalid transition rejection, closure notes                                      | MVP API/UI smoke covered                                                                                              |
| Agency assignment/timeline    | NADAA-042                                  | API, E2E, audit       | Assignment permissions, assigned-agency filtering, timeline event creation                          | MVP API/UI smoke covered                                                                                              |
| Abuse and false reports       | NADAA-091                                  | API, E2E, audit       | Rate limits, visible suspicion signals, review decisions, false-report resolution notes             | MVP API/UI smoke covered                                                                                              |
| Security hardening            | NADAA-092                                  | Security, CI          | CORS allowlist, defensive headers, non-root containers, env guardrails, residual risks              | MVP repo scan and review artifact covered                                                                             |
| Alert draft/approval          | NADAA-050                                  | API, E2E, security    | Draft, submit, approve/reject, emergency override audit                                             | MVP API/UI smoke covered                                                                                              |
| Geofenced targeting           | NADAA-051                                  | API, geospatial       | District/radius/custom geometry stored and previewable                                              | MVP API/UI smoke covered                                                                                              |
| Alert feed/delivery logs      | NADAA-052                                  | API, E2E              | Current/expired alerts visible, mock delivery attempts logged                                       | MVP API/UI smoke covered                                                                                              |
| SMS/USSD emergency access     | NADAA-110                                  | API, Integration      | Language menu, current alerts, basic reports, shelter/112 guidance, structured INFO/WARN/ERROR logs | Phase 2 API smoke covered                                                                                             |
| WhatsApp emergency chatbot    | NADAA-111                                  | API, Integration      | Alerts, risk check, report conversation state, media handoff, guide/shelter/112, transcript privacy | Phase 2 API smoke covered                                                                                             |
| Multilingual voice alerts     | NADAA-112                                  | API, Integration      | Voice variant generation, approval gate, low-literacy scripts, delivery logs, structured logs       | Phase 2 API smoke covered                                                                                             |
| Citizen mobile foundation     | NADAA-113                                  | Type, Smoke, Mobile   | Native shell, session, alerts, risk, report drafts, offline guides, shelters, permissions, push     | Phase 2 scaffold smoke/typecheck covered                                                                              |
| Community volunteers          | NADAA-120                                  | API, Mobile, Audit    | Verified profile, response group, task assignment, status/observation updates, escalation timeline  | Phase 2 API smoke, Go tests, mobile scaffold smoke covered                                                            |
| Hospital capacity tracker     | NADAA-121                                  | API, E2E, Integration | Hospital availability filters, manual updates, fixture imports, source tracking, stale warnings     | Phase 2 API smoke, Go tests, dispatcher web typecheck covered                                                         |
| Relief distribution tracking  | NADAA-122                                  | API, E2E              | Relief point list/nearby lookup, authority create/update, stock history, citizen/agency display     | Phase 2 API smoke, Go tests, citizen/agency/dispatcher web typechecks covered                                         |
| Donation and aid coordination | NADAA-123                                  | API, E2E, Audit       | Aid request creation/review, public partner listing, pledges, anti-fraud notes, CSV export          | Phase 2 API smoke, Go tests, shared-types and agency web typechecks covered                                           |
| Dispatcher mobile triage      | NADAA-124                                  | Type, Smoke, Mobile   | Agency session/MFA, incident queue, status/assignment/timeline actions, capacity, offline/stale     | Phase 2 scaffold smoke/typecheck covered                                                                              |
| Agency web operations portal  | NADAA-125                                  | Type, Smoke, E2E      | Agency session/MFA, assigned-incident scoping, responder status/timeline updates, capacity context  | Phase 2 smoke/typecheck/build/Docker covered                                                                          |
| Road closure integration      | NADAA-131                                  | API, E2E, Integration | Road closure geometry, status, severity, source, manual/adapter import, map layer, citizen context  | Phase 2 smoke/typecheck/build/Docker covered                                                                          |
| Emergency guides offline      | NADAA-060, NADAA-061                       | API, E2E/PWA          | Guide API/content model, language fallback, citizen guide browsing, and offline cache               | MVP API/UI smoke covered                                                                                              |
| Shelter lookup/update         | NADAA-062                                  | API, E2E              | Nearby lookup, occupancy update permission, map/list display                                        | MVP API/UI smoke covered                                                                                              |
| Flood ML review               | NADAA-070, NADAA-071, NADAA-072, NADAA-073 | Model, API, E2E       | Confidence, model version, explanation, no auto-publish                                             | Feature pipeline, baseline model, serving API, dispatcher review UI, source-prediction draft trace, and smoke covered |
| Agency integration contracts  | NADAA-080                                  | API, Integration      | Partner matrix, ownership, cadence, payloads, auth, retry/dead-letter behavior, mock adapters       | Contract API covered                                                                                                  |
| Weather/hydrology import      | NADAA-081                                  | Integration           | Fixture import, source metadata, persisted observation shape, retryable failures, scheduled hook    | MVP API smoke covered                                                                                                 |
| Admin governance console      | NADAA-014                                  | E2E, security         | Admin-only access shell, agency/user/MFA views, audit trace, data-source and alert-rule visibility  | MVP UI smoke covered                                                                                                  |

## Current Sprint 0 Commands

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
pnpm smoke:citizen-mobile
pnpm smoke:dispatcher-mobile
pnpm smoke:citizen-guides
pnpm smoke:alert
pnpm smoke:alert-geofence
pnpm smoke:notification
pnpm smoke:sms-ussd
pnpm smoke:whatsapp
pnpm smoke:voice-alerts
pnpm smoke:incident-abuse
pnpm smoke:incident-assignment
pnpm smoke:incident-merge
pnpm smoke:incident-workflow
pnpm smoke:ml
pnpm smoke:ml-review
pnpm smoke:risk
pnpm smoke:guide
pnpm smoke:shelter
pnpm smoke:integration
```

`pnpm smoke:web` expects the marketing website on port `5172`, the citizen app on port `5173`, the authority dashboard compatibility shell on port `5174`, the dispatcher command console on port `5175`, the admin governance console on port `5176`, and the agency operations portal on port `5177` by default. Override with `LOCAL_MARKETING_URL`, `LOCAL_CITIZEN_URL`, `LOCAL_AUTHORITY_URL`, `LOCAL_DISPATCHER_URL`, `LOCAL_ADMIN_URL`, or `LOCAL_AGENCY_URL` when a local port is already occupied.
`pnpm smoke:citizen-mobile` validates the `apps/citizen-mobile` Expo scaffold, required screen modules, NADAA logo asset, offline primitives, and package scripts.
`pnpm smoke:dispatcher-mobile` validates the `apps/dispatcher-mobile` Expo scaffold, incident queue/detail/action/capacity/profile screens, authority API helpers, offline primitives, and package scripts.
`pnpm smoke:citizen-guides` expects the citizen app on port `5173` and guide service on port `8086`.
`pnpm smoke:alert` expects the alert service on port `8089`.
`pnpm smoke:alert-geofence` expects the alert service on port `8089`.
`pnpm smoke:notification` expects the notification service on port `8090`.
`pnpm smoke:sms-ussd` expects the notification service on port `8090`.
`pnpm smoke:whatsapp` expects the notification service on port `8090`.
`pnpm smoke:voice-alerts` expects the notification service on port `8090`.
`pnpm smoke:incident-abuse` expects the incident service on port `8084`.
`pnpm smoke:incident-assignment` expects the incident service on port `8084`.
`pnpm smoke:incident-merge` expects the incident service on port `8084`.
`pnpm smoke:incident-workflow` expects the incident service on port `8084`.
`pnpm smoke:ml` expects the ML service on port `8094`.
`pnpm smoke:ml-review` expects the ML service on port `8094` and alert service on port `8089`.
`pnpm smoke:risk` expects the risk service on port `8081`.
`pnpm smoke:guide` expects the guide service on port `8086`.
`pnpm smoke:shelter` expects the shelter service on port `8093` and covers shelter/recovery lookup, hospital capacity list, manual update, fixture import, and stale/filter behavior.
`pnpm smoke:relief` expects the shelter service on port `8093` and covers relief point list/nearby lookup, authority create/update gates, invalid geometry rejection, and stock history.
`pnpm smoke:aid` expects the shelter service on port `8093` and covers public aid listing, authority request creation/review, public pledge creation, pledge review, and CSV export.
`pnpm smoke:integration` expects the integration service on port `8088`.
Agency auth, RBAC, mock MFA, and audit foundation coverage currently run through `pnpm go:test`.

## Release Candidate Checklist

- All changed packages pass type checks and tests.
- Critical API paths have success, validation, unauthorized, and forbidden coverage.
- User-facing flows have empty, loading, error, and success states.
- Sensitive authority actions create audit events.
- Public alerts cannot bypass approval.
- ML predictions cannot publish alerts automatically.
- Staging smoke tests pass.
- Known limitations and residual risks are documented.

## UAT Script Outline

The detailed UAT package lives in [UAT Plan](uat.md). Use the outline below as the release-candidate smoke path and the linked plan for stakeholder execution, feedback capture, severity, and sign-off.

1. Citizen checks flood risk for Accra Central.
2. Citizen reports a flood with location, description, affected people, and photo metadata.
3. Dispatcher reviews and merges confirmed duplicate reports.
4. Dispatcher clears or monitors any suspicious report signals without blocking urgent response.
5. Dispatcher verifies the report and changes status to verified.
6. Dispatcher assigns NADMO/district response.
7. Officer drafts a district-targeted flood alert.
8. Approver reviews and approves the alert.
9. Citizen sees the alert in the app.
10. Citizen views nearest shelter and flood guidance.
11. Authority closes incident with resolution notes.
12. Assigned agency can filter its incident queue with `assignedToMe=true`.

## Defect Severity

| Severity | Definition                                                                                | Response                                 |
| -------- | ----------------------------------------------------------------------------------------- | ---------------------------------------- |
| Critical | Could send unauthorized public alert, expose sensitive data, or block emergency reporting | Stop release                             |
| High     | Breaks a core MVP workflow or authority action                                            | Fix before QA pass                       |
| Medium   | Degrades workflow but has workaround                                                      | Fix before release candidate if feasible |
| Low      | Cosmetic, copy, or minor non-blocking issue                                               | Track and schedule                       |
