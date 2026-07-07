# Beta Monitoring

Beta monitoring confirms that the NADAA MVP is safe, usable, and operationally stable with selected users and agencies before production release.

## Monitoring Cadence

| Period              | Cadence                               | Review Focus                                                                   |
| ------------------- | ------------------------------------- | ------------------------------------------------------------------------------ |
| First 24 hours      | Every 2 hours during staffed coverage | Availability, incident intake, alert approval, privacy defects, critical bugs. |
| Days 2 to 7         | Twice daily                           | Workflow completion, support volume, beta feedback, smoke failures.            |
| Days 8 to 14        | Daily                                 | Trend health, unresolved defects, enhancement backlog, training gaps.          |
| Production go/no-go | Final beta review                     | Exit criteria, accepted risks, support readiness, stakeholder sign-off.        |

## Metrics Dashboard Definition

| Metric                                | Target / Alert Threshold                                            | Source                                                           | Owner            |
| ------------------------------------- | ------------------------------------------------------------------- | ---------------------------------------------------------------- | ---------------- |
| Web app availability                  | 99 percent during staffed beta window; alert on failed smoke run.   | `pnpm smoke:staging`, uptime checks, web health endpoints.       | Operations lead  |
| API health                            | All configured services return `/healthz`; alert on any failure.    | Staging smoke checks and service logs.                           | Engineering lead |
| Incident report creation rate         | Baseline trend reviewed daily; spike reviewed by dispatcher lead.   | Incident-service report count and audit/timeline records.        | Dispatcher lead  |
| Report-to-verification time           | Median under agreed stakeholder target; outliers reviewed daily.    | Incident status timestamps and audit logs.                       | Operations lead  |
| Verification-to-assignment time       | Median under agreed stakeholder target; outliers reviewed daily.    | Assignment records and timeline events.                          | Dispatcher lead  |
| Alert draft-to-approval time          | Reviewed per alert; urgent delays escalated immediately.            | Alert status timestamps and audit logs.                          | Product owner    |
| Unauthorized alert attempts           | Zero tolerated; any attempt triggers security review.               | Alert-service RBAC/audit records.                                | Security lead    |
| Privacy exposure defects              | Zero critical/high tolerated.                                       | UAT/beta feedback, support tickets, audit review.                | Security lead    |
| False-report decisions with reason    | 100 percent.                                                        | Incident abuse review records.                                   | QA lead          |
| ML recommendations published directly | Zero tolerated.                                                     | Alert source-prediction metadata and alert approvals.            | Engineering lead |
| Notification provider failures        | Reviewed per provider; high failure rate escalates.                 | Notification delivery logs and provider sandbox/live dashboards. | Operations lead  |
| Support ticket volume                 | Trend reviewed daily; repeated theme becomes training/backlog item. | UAT/beta feedback tracker and support log.                       | Support lead     |
| User training completion              | 100 percent for beta authority users.                               | Training attendance and completion checklist.                    | Training lead    |

## Daily Beta Review

Each review should answer:

1. Are public alert, incident reporting, privacy, and authority workflows safe to continue?
2. Did any critical or high defect appear in the last review window?
3. Did monitoring detect service, notification, or data import instability?
4. Are support tickets showing a training or UX gap?
5. Are any residual risks no longer acceptable for the next beta window?
6. Should beta continue, pause, roll back, or exit to production readiness?

## Escalation Triggers

Pause beta and escalate when:

- Unauthorized public alert behavior is observed.
- Sensitive citizen information is exposed to unauthorized roles.
- Incident submission fails for more than one staffed review window.
- Dispatcher verification or assignment is blocked.
- Notification delivery sends incorrect public content or target area.
- The same high-severity defect recurs after a fix.
