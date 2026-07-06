# QA Strategy

NADAA QA must validate product behavior, safety gates, and operational readiness. The test strategy starts lightweight in Sprint 0 and becomes stricter as public-safety workflows move toward UAT.

## Test Levels

| Level | Purpose | Owner | Runs |
| --- | --- | --- | --- |
| Type checks | Catch contract and TypeScript errors | Developers/agents | Every change |
| Unit tests | Validate service/domain behavior | Developers/agents | Every story |
| API tests | Validate request/response and authorization behavior | Backend agents/QA | Service stories |
| E2E/smoke tests | Validate critical user journeys | Frontend agents/QA | Release candidates |
| Security tests | Validate auth, RBAC, MFA, audit, rate limits, and secrets | Engineering lead/QA | Sensitive changes |
| Performance tests | Validate high-volume report and alert conditions | Platform/QA | Before beta/production |
| UAT scripts | Validate stakeholder acceptance | PM/QA/stakeholders | UAT |

## MVP Test Matrix

| Flow | Primary Stories | Test Type | Acceptance Focus | Status |
| --- | --- | --- | --- | --- |
| Citizen risk check | NADAA-021, NADAA-022 | API, E2E | Location lookup returns risk, shelters, guidance, loading/error states | MVP API/UI smoke covered |
| Agency auth, roles, and MFA | NADAA-011 | API, security | Agency membership, role catalog, MFA-required login, admin-only creation, forbidden role checks | MVP API/security tests covered |
| Citizen incident report | NADAA-030, NADAA-032 | API, E2E | GPS, hazard type, description, urgency, affected people, contact permission | MVP API/UI smoke covered |
| Media upload | NADAA-031 | API, security | File type/size validation, private storage, incident linkage | MVP API/UI smoke covered |
| Anonymous report privacy | NADAA-090 | API, security | Identity hidden where policy allows, contact permission honored | Todo |
| Duplicate detection | NADAA-033, NADAA-043 | Unit, API | Nearby/time-window duplicate candidates are reviewable, not deleted | MVP baseline covered; merge review pending |
| Authority incident map | NADAA-040 | E2E | Filters, map/list sync, role-protected access | Todo |
| Incident verification/status | NADAA-041 | API, E2E, audit | Valid transitions, invalid transition rejection, closure notes | Todo |
| Agency assignment/timeline | NADAA-042 | API, E2E, audit | Assignment permissions, timeline event creation | Todo |
| Alert draft/approval | NADAA-050 | API, E2E, security | Draft, submit, approve/reject, emergency override audit | Todo |
| Geofenced targeting | NADAA-051 | API, geospatial | District/radius/custom geometry stored and previewable | Todo |
| Alert feed/delivery logs | NADAA-052 | API, E2E | Current alerts visible, mock delivery attempts logged | Todo |
| Emergency guides offline | NADAA-060, NADAA-061 | E2E/PWA | Guides available, key content cached, language-ready | Todo |
| Shelter lookup/update | NADAA-062 | API, E2E | Nearby lookup, occupancy update permission, map/list display | Todo |
| Flood ML review | NADAA-070, NADAA-071, NADAA-072, NADAA-073 | Model, API, E2E | Confidence, model version, explanation, no auto-publish | Todo |
| Weather/hydrology import | NADAA-080, NADAA-081 | Integration | Fixture import, source metadata, retryable failures | Todo |

## Current Sprint 0 Commands

```bash
pnpm validate:docs
pnpm typecheck
pnpm build
pnpm go:test
pnpm smoke:web
pnpm smoke:risk
```

`pnpm smoke:web` expects the citizen app on port `5173` and the authority dashboard on port `5174`.
`pnpm smoke:risk` expects the risk service on port `8081`.
Agency auth, RBAC, and mock MFA coverage currently runs through `pnpm go:test`.

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
3. Dispatcher verifies the report and changes status to verified.
4. Dispatcher assigns NADMO/district response.
5. Officer drafts a district-targeted flood alert.
6. Approver reviews and approves the alert.
7. Citizen sees the alert in the app.
8. Citizen views nearest shelter and flood guidance.
9. Authority closes incident with resolution notes.

## Defect Severity

| Severity | Definition | Response |
| --- | --- | --- |
| Critical | Could send unauthorized public alert, expose sensitive data, or block emergency reporting | Stop release |
| High | Breaks a core MVP workflow or authority action | Fix before QA pass |
| Medium | Degrades workflow but has workaround | Fix before release candidate if feasible |
| Low | Cosmetic, copy, or minor non-blocking issue | Track and schedule |
