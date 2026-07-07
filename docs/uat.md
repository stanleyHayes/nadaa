# UAT Plan

NADAA UAT validates that the MVP can be accepted by stakeholders before beta. It should run in staging with non-production data, sandbox notification providers, explicit `NADAA_ALLOWED_ORIGINS`, and the security residual risks from [Security Review](security-review.md) included in the sign-off packet.

## Participants

| Group                   | Role In UAT                                                                   |
| ----------------------- | ----------------------------------------------------------------------------- |
| Product owner           | Confirms MVP scope, acceptance decisions, and enhancement backlog ownership.  |
| NADMO/district officers | Validate alert approval, incident command, shelter visibility, and workflows. |
| Dispatchers             | Validate triage, duplicate review, abuse handling, assignment, and status.    |
| System administrators   | Validate admin governance, agency/user/MFA views, audit, and data sources.    |
| QA lead                 | Captures pass/fail evidence, defects, severity, and retest results.           |
| Engineering lead        | Confirms release gates, known risks, rollback plan, and technical blockers.   |

## Entry Criteria

- CI is passing on the release candidate.
- Staging smoke checks pass for citizen web, authority compatibility shell, dispatcher web, admin web, and configured services.
- `pnpm validate:docs`, `pnpm security:scan`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`, and `pnpm go:test` pass.
- Staging has explicit `NADAA_ALLOWED_ORIGINS`.
- No real citizen data, production credentials, or live emergency notifications are used.
- Security residual risks are reviewed and accepted for UAT only.

## Feedback Capture

Every UAT observation must become one of:

- `Defect`: expected MVP behavior fails or regresses.
- `Enhancement`: useful request outside the accepted MVP behavior.
- `Question`: policy, workflow, copy, or data decision required.
- `Training`: user education or support content gap.

Use this log structure in the project tracker:

| Field           | Required Value                                                                  |
| --------------- | ------------------------------------------------------------------------------- |
| Feedback ID     | `UAT-YYYYMMDD-###`                                                              |
| Scenario ID     | Matching script below, for example `UAT-004`.                                   |
| Type            | `Defect`, `Enhancement`, `Question`, or `Training`.                             |
| Severity        | `Critical`, `High`, `Medium`, or `Low` using [QA Strategy](qa.md).              |
| Reporter        | Stakeholder name and role.                                                      |
| Environment     | Staging URL, browser/device, and service version or commit SHA.                 |
| Steps           | Minimal steps to reproduce or review.                                           |
| Expected Result | What should have happened.                                                      |
| Actual Result   | What happened.                                                                  |
| Evidence        | Screenshot, screen recording, request id, audit id, alert id, or incident id.   |
| Decision        | `Fix before UAT exit`, `Accept for beta`, `Move to backlog`, or `Not a defect`. |
| Owner           | Product, QA, Engineering, Security, Operations, or Training.                    |
| Retest Result   | `Pending`, `Passed`, `Failed`, or `Not required`.                               |

## UAT Scripts

| ID      | Flow                       | Actor                    | Script                                                                                                                                                                            | Acceptance Result                                                                                                     |
| ------- | -------------------------- | ------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| UAT-001 | Citizen risk check         | Citizen                  | Open citizen web, select Accra Central or enter coordinates, review flood risk, shelters, facilities, and recommended actions.                                                    | Risk result loads, location is understandable, advice is actionable, and empty/error states are clear.                |
| UAT-002 | Citizen incident report    | Citizen                  | Submit a flood report with location, urgency, people affected, injuries, accessibility need, contact permission, and media metadata.                                              | Incident reference is shown, report reaches authority queue, private fields respect privacy settings.                 |
| UAT-003 | Anonymous/privacy behavior | Citizen and dispatcher   | Submit an anonymous report without contact permission, then view it from dispatcher and authority views.                                                                          | Reporter identity and phone are not exposed to standard views; privacy state is visible.                              |
| UAT-004 | Dispatcher triage workflow | Dispatcher               | Open dispatcher web, filter incidents, select a report, verify it, update status, assign an agency, and add operational notes.                                                    | Status, assignment, and timeline changes are visible and audited.                                                     |
| UAT-005 | Duplicate and abuse review | Dispatcher               | Review duplicate candidates and suspicious report signals; accept a merge candidate or mark a report as false only with required reason.                                          | No report is silently discarded; decision reason and audit trace exist.                                               |
| UAT-006 | Alert approval workflow    | Officer and approver     | Draft a district-targeted flood alert, submit it, approve or reject it as a different authorized approver, and review public feed output.                                         | Public alert cannot bypass approval; emergency override is restricted and audited.                                    |
| UAT-007 | ML-assisted alert drafting | Dispatcher and officer   | Review flood ML prediction, inspect confidence/explanation factors, create an alert draft from the recommendation, and verify no auto-publish occurs.                             | Prediction is transparent, source metadata is traceable, and draft still requires human approval.                     |
| UAT-008 | Citizen alert and guidance | Citizen                  | View current/expired alerts, open offline-first emergency guides, switch language where available, and check nearby shelters/recovery support.                                    | Citizen can understand alert status, access guide content, and find shelter/recovery information.                     |
| UAT-009 | Admin governance           | System administrator     | Open admin web, review agency/user/role/MFA support views, audit log, integration data sources, and alert-rule configuration.                                                     | Admin surface is separate from dispatch operations and does not expose secrets or unnecessary citizen report details. |
| UAT-010 | Integration readiness      | Admin and engineering    | Review integration contracts, run weather/hydrology fixture import, inspect imported observation status, and confirm failed imports are retryable.                                | Data-source ownership, cadence, payload, status, and retry behavior are understandable.                               |
| UAT-011 | Staging smoke              | QA lead                  | Run staging smoke checks with configured web and service URLs, then capture command output and any failing URL.                                                                   | Smoke suite passes or each failing check is logged with owner and severity.                                           |
| UAT-012 | Sign-off review            | Product, QA, engineering | Review defect list, unresolved questions, security residual risks, release notes, beta metrics, rollback plan, support coverage, training outline, and production no-go criteria. | Stakeholders sign UAT pass, conditional pass, or fail with explicit blockers.                                         |

## Exit Criteria

- All `Critical` and `High` UAT defects are fixed and retested, or explicitly accepted by product, security, and operations for beta only.
- No open defect can send an unauthorized public alert, expose sensitive citizen data, or block emergency reporting.
- Beta metrics and alert thresholds are agreed.
- Release notes, user guide, training outline, acceptance checklist, rollback plan, and hypercare checklist are complete.
- Product owner, QA lead, engineering lead, and operations lead sign the release-readiness checklist.
