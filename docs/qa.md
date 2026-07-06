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

| Flow                         | Primary Stories                            | Test Type          | Acceptance Focus                                                                                  | Status                         |
| ---------------------------- | ------------------------------------------ | ------------------ | ------------------------------------------------------------------------------------------------- | ------------------------------ |
| Citizen risk check           | NADAA-021, NADAA-022                       | API, E2E           | Location lookup returns risk, shelters, guidance, loading/error states                            | MVP API/UI smoke covered       |
| Agency auth, roles, and MFA  | NADAA-011                                  | API, security      | Agency membership, role catalog, MFA-required login, admin-only creation, forbidden role checks   | MVP API/security tests covered |
| Audit logging foundation     | NADAA-012                                  | API, security      | Actor/action/target metadata, before/after snapshots, RBAC denial events, system-admin audit read | MVP API/security tests covered |
| Citizen incident report      | NADAA-030, NADAA-032                       | API, E2E           | GPS, hazard type, description, urgency, affected people, contact permission                       | MVP API/UI smoke covered       |
| Media upload                 | NADAA-031                                  | API, security      | File type/size validation, private storage, incident linkage                                      | MVP API/UI smoke covered       |
| Anonymous report privacy     | NADAA-090                                  | API, security      | Identity hidden where policy allows, contact permission honored                                   | Todo                           |
| Duplicate detection          | NADAA-033, NADAA-043                       | Unit, API          | Nearby/time-window duplicate candidates are reviewable, mergeable, and traceable                  | MVP API/UI smoke covered       |
| Authority incident map       | NADAA-040                                  | E2E                | Filters, map/list sync, role-protected access, loading/empty/error fallback                       | MVP UI smoke covered           |
| Incident verification/status | NADAA-041                                  | API, E2E, audit    | Valid transitions, invalid transition rejection, closure notes                                    | MVP API/UI smoke covered       |
| Agency assignment/timeline   | NADAA-042                                  | API, E2E, audit    | Assignment permissions, assigned-agency filtering, timeline event creation                        | MVP API/UI smoke covered       |
| Abuse and false reports      | NADAA-091                                  | API, E2E, audit    | Rate limits, visible suspicion signals, review decisions, false-report resolution notes           | MVP API/UI smoke covered       |
| Alert draft/approval         | NADAA-050                                  | API, E2E, security | Draft, submit, approve/reject, emergency override audit                                           | MVP API/UI smoke covered       |
| Geofenced targeting          | NADAA-051                                  | API, geospatial    | District/radius/custom geometry stored and previewable                                            | MVP API/UI smoke covered       |
| Alert feed/delivery logs     | NADAA-052                                  | API, E2E           | Current/expired alerts visible, mock delivery attempts logged                                     | MVP API/UI smoke covered       |
| Emergency guides offline     | NADAA-060, NADAA-061                       | API, E2E/PWA       | Guide API/content model, language fallback, citizen guide browsing, and offline cache             | MVP API/UI smoke covered       |
| Shelter lookup/update        | NADAA-062                                  | API, E2E           | Nearby lookup, occupancy update permission, map/list display                                      | Todo                           |
| Flood ML review              | NADAA-070, NADAA-071, NADAA-072, NADAA-073 | Model, API, E2E    | Confidence, model version, explanation, no auto-publish                                           | Todo                           |
| Agency integration contracts | NADAA-080                                  | API, Integration   | Partner matrix, ownership, cadence, payloads, auth, retry/dead-letter behavior, mock adapters     | Contract API covered           |
| Weather/hydrology import     | NADAA-081                                  | Integration        | Fixture import, source metadata, persisted observation shape, retryable failures, scheduled hook  | MVP API smoke covered          |

## Current Sprint 0 Commands

```bash
pnpm validate:docs
pnpm typecheck
pnpm build
pnpm go:test
pnpm smoke:web
pnpm smoke:citizen-guides
pnpm smoke:alert
pnpm smoke:alert-geofence
pnpm smoke:notification
pnpm smoke:incident-abuse
pnpm smoke:incident-assignment
pnpm smoke:incident-merge
pnpm smoke:incident-workflow
pnpm smoke:risk
pnpm smoke:guide
pnpm smoke:integration
```

`pnpm smoke:web` expects the citizen app on port `5173`, the authority dashboard compatibility shell on port `5174`, and the dispatcher command console on port `5175`.
`pnpm smoke:citizen-guides` expects the citizen app on port `5173` and guide service on port `8086`.
`pnpm smoke:alert` expects the alert service on port `8089`.
`pnpm smoke:alert-geofence` expects the alert service on port `8089`.
`pnpm smoke:notification` expects the notification service on port `8090`.
`pnpm smoke:incident-abuse` expects the incident service on port `8084`.
`pnpm smoke:incident-assignment` expects the incident service on port `8084`.
`pnpm smoke:incident-merge` expects the incident service on port `8084`.
`pnpm smoke:incident-workflow` expects the incident service on port `8084`.
`pnpm smoke:risk` expects the risk service on port `8081`.
`pnpm smoke:guide` expects the guide service on port `8086`.
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
