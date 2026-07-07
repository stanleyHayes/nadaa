# Hypercare

Hypercare covers the first controlled beta window and the first production window after launch. It keeps product, engineering, operations, security, QA, and support aligned while usage patterns and defects are still fresh.

## Hypercare Roles

| Role               | Responsibility                                                                      |
| ------------------ | ----------------------------------------------------------------------------------- |
| Incident commander | Owns launch-room decisions, go/no-go calls, and escalation priority.                |
| Engineering lead   | Owns technical triage, rollback recommendation, and fix coordination.               |
| Operations lead    | Owns service monitoring, smoke checks, provider status, and deployment health.      |
| Security lead      | Owns alert misuse, privacy exposure, secret exposure, and access-control incidents. |
| QA lead            | Owns defect intake quality, reproduction, severity, and retest coordination.        |
| Product owner      | Owns scope decisions, stakeholder communications, and enhancement prioritization.   |
| Support lead       | Owns user communications, known issues, training gaps, and support handoff.         |

## Severity Response

| Severity | Examples                                                                      | Initial Response                               |
| -------- | ----------------------------------------------------------------------------- | ---------------------------------------------- |
| Critical | Unauthorized public alert, sensitive data exposure, blocked emergency intake. | Stop rollout, assemble leads, decide rollback. |
| High     | Core workflow broken, MFA/access issue, alert approval unusable.              | Fix before next rollout window or pause beta.  |
| Medium   | Workflow degraded with workaround, confusing UI state, slow manual process.   | Track, assign owner, target next patch window. |
| Low      | Copy issue, minor visual bug, non-blocking training gap.                      | Track for backlog or scheduled polish.         |

## Launch Room Checklist

Before opening beta or production:

- Confirm release candidate SHA and image tags.
- Confirm staging smoke and release-candidate command results.
- Confirm `NADAA_ALLOWED_ORIGINS` and environment variables.
- Confirm notification provider mode and target audience.
- Confirm support rota, escalation channel, and decision owners.
- Confirm rollback owner and last known good version.
- Confirm no critical or high UAT defects are unresolved without explicit acceptance.
- Confirm known limitations are included in release notes.

## Daily Hypercare Checklist

- Review smoke-check results and service health.
- Review incident intake, verification, assignment, alert approval, and notification logs.
- Review UAT/beta feedback, support tickets, and unresolved blockers.
- Review security and privacy events.
- Review beta metrics and threshold breaches.
- Decide continue, patch, pause, rollback, or exit hypercare.
- Publish a short stakeholder update with status, blockers, and next action.

## Support Handoff

At the end of hypercare, support should have:

- Known issues list with severity, workaround, and owner.
- User guide and training completion notes.
- Escalation contacts for product, engineering, security, and operations.
- Monitoring dashboard links or command references.
- Release notes and rollback record.
- Backlog of accepted enhancements and deferred non-critical defects.
