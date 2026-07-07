# Release Readiness

This document is the NADAA MVP release gate for UAT, beta, and production. It links the technical checks, stakeholder acceptance, release notes, rollout plan, and no-go criteria needed before the MVP leaves staging.

## Release Candidate Gate

| Gate                  | Required Evidence                                                                                                           | Owner            |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| Scope confirmation    | MVP stories in `agent_plan.md` are Done or explicitly deferred.                                                             | Product owner    |
| CI and tests          | CI passes plus local release-candidate commands pass.                                                                       | Engineering lead |
| Security              | [Security Review](security-review.md) is complete and residual risks are accepted for the target environment.               | Security lead    |
| UAT                   | [UAT Plan](uat.md) scenarios are executed and defects are triaged.                                                          | QA lead          |
| Staging deployment    | Staging smoke checks pass against configured URLs.                                                                          | Release engineer |
| Beta monitoring       | [Beta Monitoring](beta-monitoring.md) metrics, owners, thresholds, and cadence are confirmed.                               | Operations lead  |
| User readiness        | [User Guide And Training](user-guide.md) is reviewed by stakeholder representatives.                                        | Training lead    |
| Support readiness     | [Hypercare](hypercare.md) coverage, escalation paths, severity handling, and contact rota are confirmed.                    | Support lead     |
| Data and legal review | Production data retention, citizen privacy, agency access boundaries, and public alert policy are approved by stakeholders. | Product owner    |

## Release Candidate Commands

Run and attach the command output to the release sign-off packet:

```bash
pnpm validate:docs
pnpm security:scan
pnpm audit --audit-level high
pnpm lint
pnpm typecheck
pnpm test
pnpm build
pnpm go:test
pnpm smoke:staging
```

Use `STAGING_*` environment variables from [Deployment](deployment.md) for `pnpm smoke:staging`.

## Acceptance Checklist

| Item                                      | Required Before UAT Exit | Required Before Beta | Required Before Production |
| ----------------------------------------- | ------------------------ | -------------------- | -------------------------- |
| UAT scripts executed                      | Yes                      | Yes                  | Yes                        |
| Critical and high UAT defects resolved    | Yes                      | Yes                  | Yes                        |
| Security residual risks accepted          | Yes                      | Yes                  | No unaccepted high risks   |
| Explicit CORS allowlist configured        | Yes                      | Yes                  | Yes                        |
| Mock MFA restrictions understood          | Yes                      | Yes                  | Production MFA connected   |
| Staging smoke checks passed               | Yes                      | Yes                  | Yes                        |
| Rollback plan reviewed                    | Yes                      | Yes                  | Yes                        |
| Support rota confirmed                    | Yes                      | Yes                  | Yes                        |
| Beta monitoring thresholds confirmed      | No                       | Yes                  | Yes                        |
| Production data retention policy approved | No                       | Conditional          | Yes                        |
| Live notification provider approvals      | No                       | Conditional          | Yes                        |

## Release Notes Template

Use this template for each release candidate and final launch note.

```md
# NADAA MVP Release Notes

Release candidate:
Commit SHA:
Release date:
Prepared by:

## Summary

- Citizen-facing changes:
- Dispatcher-facing changes:
- Admin/governance changes:
- API/service changes:
- Security and privacy changes:
- Known limitations:

## Verification

- CI run:
- Staging smoke result:
- UAT result:
- Security review:

## Rollback

- Rollback owner:
- Last known good version:
- Database migration impact:
- Notification/provider impact:

## Sign-Off

- Product:
- QA:
- Engineering:
- Security:
- Operations:
```

## Rollback Plan

1. Pause public rollout communications and notify the release channel.
2. Disable live notification provider delivery if alert safety is affected.
3. Revert web app image tags to the last known good version.
4. Revert service image tags to the last known good version.
5. Do not roll back database migrations until the migration owner confirms data impact and backup state.
6. Run staging or production health checks after rollback.
7. Record the incident, root cause, decision owner, and follow-up action.

## No-Go Criteria

Do not enter beta or production when any of these are true:

- A known path can send an unauthorized public alert.
- Sensitive citizen identity, contact details, exact private location, or media can be exposed to unauthorized users.
- Incident reporting, dispatcher triage, alert approval, or citizen alert feed is unavailable in staging.
- Security review has unaccepted critical or high findings for the target environment.
- Production secrets or real citizen data are committed or exposed.
- Support coverage, escalation, and rollback ownership are not confirmed.
