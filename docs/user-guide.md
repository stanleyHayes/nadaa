# User Guide And Training

This guide prepares UAT, beta, and production training for the NADAA MVP. It is not a replacement for live emergency policy training; it explains how the MVP surfaces should be used and where human approval remains mandatory.

NADAA stands for National Disaster Alert & Response Platform.

Slogan: **Be Aware. Be Prepared. Be Safe.**

---

## Training Audiences

| Audience              | Product Surface          | Training Outcome                                                                  |
| --------------------- | ------------------------ | --------------------------------------------------------------------------------- |
| Citizens              | `apps/citizen-web`       | Check risk, report incidents, view alerts, access guides, and find shelters.      |
| Dispatchers           | `apps/dispatcher-web`    | Triage reports, verify incidents, manage status, assign agencies, review signals. |
| Officers/approvers    | Authority alert workflow | Draft, submit, approve, reject, and override public alerts safely.                |
| System administrators | `apps/admin-web`         | Review governance data, users, roles, MFA support, audit, and data sources.       |
| Support team          | Docs and dashboards      | Capture defects, triage severity, monitor beta health, and escalate incidents.    |

---

## Citizen Quick Guide

### How to check area risk

1. Open the citizen web app.
2. Check area flood risk with an area preset, manual coordinates, or device location.
3. Read risk guidance, nearby shelters, facilities, and emergency steps.

### How to report an incident

1. Tap **Report an incident**.
2. Select the hazard type, for example flood, fire, road crash, or disease outbreak.
3. Set the location using GPS, a map pin, or manual coordinates.
4. Describe the incident: urgency, people affected, injuries, accessibility needs, and optional media.
5. Choose whether to report anonymously and whether responders may contact you.
6. Submit the report and save the reference number.

> **Important:** For life-threatening emergencies, call **112** first. NADAA supplements, but does not replace, emergency services.

### How to view alerts and use guides

1. Open the **Alerts** tab for current and recent alerts.
2. Open the **Guides** tab and choose a hazard and stage: **before**, **during**, **after**, or **recovery**.
3. Guides that have already been loaded remain available offline.

### How to find shelters and recovery support

1. Open the **Shelters** or **Recovery** tab.
2. View nearby shelters, hospitals, relief distribution points, and aid services.
3. Check opening status, capacity, and eligibility notes before travelling.

### Citizen training emphasis

- Call local emergency services for immediate life-threatening situations.
- Only share contact permission when willing to be contacted.
- Anonymous reports can still help responders, but may limit follow-up.
- Media should show the incident clearly and avoid unnecessary private details.

---

## Dispatcher Quick Guide

### How to use the incident queue

1. Open dispatcher web with an authorized authority session.
2. Review the incident map, queue, filters, severity, status, and privacy indicators.
3. Select an incident and verify it only when evidence supports authority action.
4. Review duplicate candidates and merge only when the same event is clearly represented.
5. Review suspicious report signals without suppressing urgent reports only because of a score.
6. Assign verified incidents to the appropriate response agency.
7. Update incident status and add timeline notes as response progresses.
8. Review ML flood predictions as decision support, not as automatic alert authority.

### Dispatcher training emphasis

- Privacy indicators explain what reporter information can be viewed.
- False-report closure requires reason notes.
- Assignment and status changes are audited.
- ML recommendations can create alert drafts but cannot publish alerts.

---

## Officer And Approver Quick Guide

### How to create and approve an alert

1. Create or review an alert draft.
2. Confirm issuing agency, hazard, severity, target area, expiry, and public wording.
3. Submit the alert for approval.
4. Approver reviews the draft, target preview, and safety context.
5. Approver approves or rejects with a note.
6. Use emergency override only for urgent life-safety warnings and only with a reason.
7. Confirm alert feed visibility and delivery-log state after approval.

### Alert training emphasis

- Public alerts cannot bypass human approval.
- Non-system approvers cannot approve their own draft.
- Emergency override is restricted and audited.
- Expiry is mandatory.

---

## Admin Quick Guide

### How to manage the platform

1. Open admin web with an authorized system/admin session.
2. Review agencies, users, roles, and MFA support state.
3. Review audit logs for sensitive activity.
4. Review integration data sources and configured alert rules.
5. Avoid entering dispatcher operations from the admin surface.
6. Do not enter or expose real secrets in the UI or issue tracker.

### Admin training emphasis

- Admin work is separate from dispatcher work.
- Secrets belong in environment or secret-manager configuration, not docs or screenshots.
- Audit records may contain sensitive operational metadata.

---

## Safety, Privacy, And Data Rules

### Public alerts

- A public alert cannot be sent without human approval.
- ML predictions can create drafts but cannot publish alerts.
- Emergency override is restricted, audited, and must include a reason.

### Citizen privacy

- Anonymous reports do not retain a reporter identity.
- Contact details are only shown to authorized responders when permission is given.
- Private media is only visible to authorized viewers through signed, short-lived URLs.
- Exact incident locations are only available to authorized authority endpoints.

### Authority accountability

- Every sensitive action creates an audit record.
- Audit records include actor, agency, role, action, target, request ID, and before/after snapshots.
- Audit logs must not store passwords, OTPs, MFA codes, tokens, or raw private media.

### Data classification

| Class      | Examples                                                         | Default Handling            |
| ---------- | ---------------------------------------------------------------- | --------------------------- |
| Public     | Approved alerts, public guidance, public shelter listings        | Cacheable, no personal data |
| Internal   | Incident status, assignments, operational notes                  | Authority-only              |
| Sensitive  | Citizen phone, exact location, contact permission, private media | Need-to-know access         |
| Restricted | Admin credentials, provider tokens, cloud secrets                | Secret manager only         |

---

## Training Session Outline

| Session                       | Duration | Audience                    | Materials                                            |
| ----------------------------- | -------- | --------------------------- | ---------------------------------------------------- |
| MVP overview and safety model | 30 min   | All stakeholders            | Product scope, security review, UAT plan.            |
| Citizen reporting and alerts  | 45 min   | Product, QA, support        | Citizen web, sample incidents, alert feed.           |
| Dispatcher incident command   | 60 min   | Dispatchers, QA, operations | Dispatcher web, incident workflow, duplicate review. |
| Alert approval and ML review  | 45 min   | Officers, approvers, QA     | Alert workflow, ML review, audit expectations.       |
| Admin governance              | 45 min   | System admins, support      | Admin web, users, roles, MFA support, audit.         |
| UAT feedback and beta support | 30 min   | QA, product, support        | UAT log, severity model, hypercare process.          |

---

## Training Completion Checklist

- [ ] Each participant can identify the surface they should use.
- [ ] Each authority participant understands MFA and role limits.
- [ ] Dispatchers can explain privacy indicators and false-report handling.
- [ ] Alert approvers can explain approval, rejection, expiry, and emergency override.
- [ ] Admins can explain why governance is separated from incident command.
- [ ] QA and support can capture feedback with scenario IDs and severity.
