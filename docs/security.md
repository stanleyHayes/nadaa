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
- Agency tokens include user type, role, agency id, and MFA-completed state.
- Agency-user creation is limited to `system_admin` and `agency_admin`; agency admins are scoped to their own agency.
- Agency users cannot log in until MFA setup and verification are complete.
- Mock MFA is acceptable only for local development and automated tests until the production MFA provider is connected.

## MVP Controls

- Role-based access control for authority workflows.
- MFA for authority users.
- Audit logs for alert, incident, assignment, status, and admin actions.
- Explicit contact permission for citizen reports.
- Anonymous report support where policy allows.
- Rate limits for public incident intake.
- Private media storage with controlled access.
- Approval workflow for mass alerts.
- Emergency override restricted to authorized roles and fully audited.

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

## Data Classification

| Class | Examples | Default Handling |
| --- | --- | --- |
| Public | Approved alerts, public guidance, approved shelter listings | Cacheable, no personal data |
| Internal | Incident status, assignments, operational notes | Authority-only |
| Sensitive | Citizen phone, exact home location, report contact permission, private media | Need-to-know access |
| Restricted | Admin credentials, provider tokens, Jira/GitHub/cloud secrets | Secret manager only |
| Open Data Candidate | Aggregated incident counts, anonymized risk zones | Requires privacy review |

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
- Surface suspicion flags to dispatchers.
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
- ML predictions can create drafts but cannot publish alerts.
- Alert expiry is mandatory.
- Alerts must keep issuing agency, approver, target geometry, and delivery logs.

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
