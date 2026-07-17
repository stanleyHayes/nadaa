# Security

NADAA handles emergency reports, location data, media, authority actions, public warnings, and ML predictions. Security controls must be treated as product requirements, not later hardening.

## Security Goals

- Prevent unauthorized public alerts.
- Protect citizen identity, location, contact permission, media, and anonymous reporting choices.
- Ensure authority actions are attributable through audit logs.
- Keep official agency data source metadata intact.
- Prevent automated systems from suppressing urgent life-safety reports without human review.
- Keep secrets out of source control and client bundles.

## Roles

Initial roles:

- `citizen`
- `agency_viewer`
- `dispatcher`
- `responder`
- `nadmo_officer`
- `district_officer`
- `agency_admin`
- `system_admin`

## Sensitive Actions

Sensitive actions require authority authentication, RBAC, MFA where applicable, and audit logging.

- Create, submit, approve, reject, publish, expire, or override alerts.
- Verify, assign, merge, close, or mark incidents as false reports.
- View non-public citizen contact details.
- View private incident media.
- Update shelter capacity or hospital capacity.
- Create agency users and change roles.
- Export damage, missing persons, or open data records.
- Create alert drafts from ML predictions.

## Authority Authentication Baseline

- Agency users must belong to an agency and use one of the authority roles.
- Agency users log in via `POST /api/v1/auth/agency/login` (`{email, password, mfaCode}`); auth-service issues 12-hour HMAC-signed bearer tokens (`nadaa.<payload>.<sig>`).
- Agency tokens include user type, role, agency id, district, and MFA-completed state.
- Every service verifies the bearer token with the shared `NADAA_AUTH_TOKEN_SECRET`; authority endpoints reject unauthenticated or under-privileged callers.
- Agency-user creation is limited to `system_admin` and `agency_admin`; agency admins are scoped to their own agency.
- Agency users cannot log in until MFA setup and verification are complete.
- Self-asserted `X-NADAA-Actor-*` identity headers are honored only when a service runs with `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` (local development and smoke tests); deployed environments must set it to `false`.
- Mock MFA is acceptable only for local development and automated tests until the production MFA provider is connected.

## MVP Controls

- Role-based access control for authority workflows.
- MFA for authority users.
- Audit logs for alert, incident, assignment, status, and admin actions.
- Authority workflow endpoints on every service require a verified auth-service bearer token; self-asserted actor/request-id headers are accepted only in local development (`NADAA_AUTH_ALLOW_MOCK_ACTORS=true`).
- Explicit contact permission for citizen reports.
- Anonymous report support where policy allows.
- Rate limits for public incident intake.
- Private media storage with controlled access.
- Approval workflow for mass alerts.
- Emergency override restricted to authorized roles and fully audited.
- Runtime API CORS must use `NADAA_ALLOWED_ORIGINS` in staging, beta, and production.
- Runtime API responses should include defensive security headers and no-store cache headers.

## Runtime HTTP Hardening

All Go APIs use a shared local pattern for CORS and defensive response headers:

- `NADAA_ALLOWED_ORIGINS` accepts a comma-separated list of approved browser origins.
- Empty `NADAA_ALLOWED_ORIGINS` or `*` keeps wildcard CORS only for local development and fixture smoke testing.
- Staging, beta, and production must set explicit citizen, dispatcher, agency/admin, and authority origins.
- CORS responses vary by `Origin` when an allowlist is active.
- API responses include `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Content-Security-Policy`, `Strict-Transport-Security`, and `Cache-Control: no-store`.

Security review notes and residual risks are tracked in [Security Review](security-review.md).

## Audit Log Minimum Fields

- `id`
- `actorUserId`
- `actorAgencyId`
- `actorRole`
- `action`
- `targetType`
- `targetId`
- `requestId`
- `ipAddress`
- `userAgent`
- `before`
- `after`
- `createdAt`

## Audit Retention Assumptions

- MVP audit logs are append-only and should be retained for at least 24 months in production unless Ghana legal or agency policy requires a longer period.
- Production storage should be tamper-evident, backed up, and restricted to authorized system administrators and auditors.
- Audit records may include internal user identifiers, agency identifiers, request metadata, and sanitized before/after snapshots.
- Audit records must not store passwords, OTPs, MFA codes, bearer tokens, object-storage credentials, provider API keys, or raw private media.
- Exports of audit logs are restricted actions and must create their own audit events once export workflows are implemented.

## Data Classification

| Class               | Examples                                                                     | Default Handling            |
| ------------------- | ---------------------------------------------------------------------------- | --------------------------- |
| Public              | Approved alerts, public guidance, approved shelter listings                  | Cacheable, no personal data |
| Internal            | Incident status, assignments, operational notes                              | Authority-only              |
| Sensitive           | Citizen phone, exact home location, report contact permission, private media | Need-to-know access         |
| Restricted          | Admin credentials, provider tokens, Jira/GitHub/cloud secrets                | Secret manager only         |
| Open Data Candidate | Aggregated incident counts, anonymized risk zones                            | Requires privacy review     |

## Incident Privacy Controls

- Public incident intake may accept reporter metadata, but anonymous reports do not retain `reportedBy`.
- Reports without contact permission must not expose reporter phone or identity in authority incident views.
- Authority incident list, duplicate-review, merge, verify, status, abuse-review, and assignment responses apply server-side incident sanitization before returning records.
- Reporter identity and contact visibility are limited to MFA-verified `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher` actors when the citizen has granted contact permission.
- `responder` and `agency_viewer` roles receive standard operational incident views without reporter identity or phone.
- Exact incident location is available only through MFA-verified authority incident endpoints and is used for emergency response routing, duplicate detection, assignment, and verified authority coordination.
- Command UIs must surface privacy state so operators understand whether reporter identity, contact, and location use are restricted.

## Secret Handling

Never commit:

- API tokens.
- Client or citizen data.
- Database credentials.
- GitHub secrets.
- Jira credentials.
- SMS, WhatsApp, push, email, or cloud provider credentials.

Use environment variables and deployment secret stores.

## Incident Intake Abuse Controls

- Rate limit anonymous and authenticated report submissions.
- Track repeated reports from the same phone/device/IP when available.
- Detect near-duplicate reports by hazard, time, distance, and description similarity.
- Surface transparent suspicion flags, scores, and review state to dispatchers.
- Require reason notes for false-report closure.
- Do not silently discard life-threatening reports based only on automated suspicion.

## Media Security

- Validate content type and file size before upload.
- Store media in object storage, not the relational database.
- Keep incident media private by default.
- Generate short-lived signed URLs for authorized viewing.
- Retain source, uploader, incident id, checksum, and created timestamp.
- Apply retention policy once legal and operational requirements are confirmed.

## Alert Safety

- Alert drafts do not reach citizens until approved.
- Mass alerts require approval.
- Emergency override is restricted, audited, and visible in review reports.
- Alert-service write endpoints require a verified authority bearer token with MFA completed (mock actor headers only in local development).
- Non-system approvers cannot approve their own draft; emergency override is the audited exception path for urgent public warnings.
- ML predictions can create drafts but cannot publish alerts.
- Alert expiry is mandatory.
- Alerts must keep issuing agency, approver, target geometry, and delivery logs.

## Incident Workflow Safety

- Verification is limited to `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`.
- Operational status updates are limited to authority workflow roles and require completed MFA.
- Duplicate merge review is limited to `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`; accepted merges must keep primary and duplicate audit trails.
- Abuse and false-report review is limited to `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`; accepted decisions must create audit events.
- Agency assignment is limited to `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`; agency admins can assign only to their own agency.
- Assigned-agency queue filters require an authority reader role and completed MFA.
- Shelter capacity and occupancy updates are limited to MFA-verified `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher` actors.
- `closed` and `false_report` are terminal incident states.
- `resolutionNotes` are mandatory for `closed` and `false_report`.
- Accepted status changes create before/after audit events.

## AI/ML Safety

- ML predictions must not automatically issue public alerts.
- Model outputs must show confidence and explanation factors.
- Model version and prediction inputs must be logged.
- False positives and false negatives must be reviewable.
- Authority dashboard must make model uncertainty visible.

## Open Questions

- Final Ghana data protection and retention requirements for emergency reports.
- Official approval policy for anonymous reports and identity disclosure.
- Telecom and government requirements for future cell broadcast.
- Agency-by-agency access boundaries for shared incidents.
