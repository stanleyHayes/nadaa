# User Guide And Training

This guide prepares UAT, beta, and production training for the NADAA MVP. It is not a replacement for live emergency policy training; it explains how the MVP surfaces should be used and where human approval remains mandatory.

## Training Audiences

| Audience              | Product Surface          | Training Outcome                                                                  |
| --------------------- | ------------------------ | --------------------------------------------------------------------------------- |
| Citizens              | `apps/citizen-web`       | Check risk, report incidents, view alerts, access guides, and find shelters.      |
| Dispatchers           | `apps/dispatcher-web`    | Triage reports, verify incidents, manage status, assign agencies, review signals. |
| Officers/approvers    | Authority alert workflow | Draft, submit, approve, reject, and override public alerts safely.                |
| System administrators | `apps/admin-web`         | Review governance data, users, roles, MFA support, audit, and data sources.       |
| Support team          | Docs and dashboards      | Capture defects, triage severity, monitor beta health, and escalate incidents.    |

## Citizen Quick Guide

1. Open the citizen web app.
2. Check area flood risk with an area preset, manual coordinates, or device location.
3. Read risk guidance, nearby shelters, facilities, and emergency steps.
4. Report an incident with hazard, location, urgency, people affected, injuries, accessibility needs, and optional media.
5. Choose whether to report anonymously and whether responders may contact you.
6. Review the reference number after submission.
7. Open the alert feed for current and expired alerts.
8. Use emergency guides and shelter/recovery support information when connectivity is limited.

Citizen training emphasis:

- Call local emergency services for immediate life-threatening situations.
- Only share contact permission when willing to be contacted.
- Anonymous reports can still help responders, but may limit follow-up.
- Media should show the incident clearly and avoid unnecessary private details.

## Dispatcher Quick Guide

1. Open dispatcher web with an authorized authority session.
2. Review the incident map, queue, filters, severity, status, and privacy indicators.
3. Select an incident and verify it only when evidence supports authority action.
4. Review duplicate candidates and merge only when the same event is clearly represented.
5. Review suspicious report signals without suppressing urgent reports only because of a score.
6. Assign verified incidents to the appropriate response agency.
7. Update incident status and add timeline notes as response progresses.
8. Review ML flood predictions as decision support, not as automatic alert authority.

Dispatcher training emphasis:

- Privacy indicators explain what reporter information can be viewed.
- False-report closure requires reason notes.
- Assignment and status changes are audited.
- ML recommendations can create alert drafts but cannot publish alerts.

## Officer And Approver Quick Guide

1. Create or review an alert draft.
2. Confirm issuing agency, hazard, severity, target area, expiry, and public wording.
3. Submit the alert for approval.
4. Approver reviews the draft, target preview, and safety context.
5. Approver approves or rejects with a note.
6. Use emergency override only for urgent life-safety warnings and only with a reason.
7. Confirm alert feed visibility and delivery-log state after approval.

Alert training emphasis:

- Public alerts cannot bypass human approval.
- Non-system approvers cannot approve their own draft.
- Emergency override is restricted and audited.
- Expiry is mandatory.

## Admin Quick Guide

1. Open admin web with an authorized system/admin session.
2. Review agencies, users, roles, and MFA support state.
3. Review audit logs for sensitive activity.
4. Review integration data sources and configured alert rules.
5. Avoid entering dispatcher operations from the admin surface.
6. Do not enter or expose real secrets in the UI or issue tracker.

Admin training emphasis:

- Admin work is separate from dispatcher work.
- Secrets belong in environment or secret-manager configuration, not docs or screenshots.
- Audit records may contain sensitive operational metadata.

## Training Session Outline

| Session                       | Duration | Audience                    | Materials                                            |
| ----------------------------- | -------- | --------------------------- | ---------------------------------------------------- |
| MVP overview and safety model | 30 min   | All stakeholders            | Product scope, security review, UAT plan.            |
| Citizen reporting and alerts  | 45 min   | Product, QA, support        | Citizen web, sample incidents, alert feed.           |
| Dispatcher incident command   | 60 min   | Dispatchers, QA, operations | Dispatcher web, incident workflow, duplicate review. |
| Alert approval and ML review  | 45 min   | Officers, approvers, QA     | Alert workflow, ML review, audit expectations.       |
| Admin governance              | 45 min   | System admins, support      | Admin web, users, roles, MFA support, audit.         |
| UAT feedback and beta support | 30 min   | QA, product, support        | UAT log, severity model, hypercare process.          |

## Training Completion Checklist

- Each participant can identify the surface they should use.
- Each authority participant understands MFA and role limits.
- Dispatchers can explain privacy indicators and false-report handling.
- Alert approvers can explain approval, rejection, expiry, and emergency override.
- Admins can explain why governance is separated from incident command.
- QA and support can capture feedback with scenario IDs and severity.
