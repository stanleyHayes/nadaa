# NADAA-092 Security Review

Date: 2026-07-07

Scope: MVP security, privacy, and safety hardening before UAT readiness. This review covers public citizen flows, authority workflows, Go APIs, CI checks, Docker runtime posture, environment handling, alert approval paths, ML-assisted alert drafting, media privacy assumptions, audit coverage, and known residual risks.

## Threat Model Summary

High-impact misuse paths for the MVP:

- Unauthorized public alert creation or publication.
- Citizen identity, phone, location, contact-permission, or private media exposure.
- Abuse of public incident intake through spam, duplicate reports, or intentionally false reports.
- ML predictions being treated as authority-approved alerts.
- Cross-origin browser access from unapproved web origins.
- Root runtime containers widening impact after an image or runtime compromise.
- Secrets or production environment files being committed to the repository.

## Reviewed Controls

| Area                 | Review Result                                                                                                                                                                          |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Auth and MFA         | Authority auth, role catalog, MFA setup/verification, and MFA-aware agency tokens are implemented in the auth service. Mock MFA remains local/test-only until a provider is connected. |
| RBAC                 | Sensitive authority workflows require actor, role, agency, MFA-completed, and request-id context until shared bearer-token middleware is wired across services.                        |
| Alert approval       | Alert drafts require submit/approve or emergency override. Non-system approvers cannot approve their own draft. Emergency override is restricted and audited.                          |
| Audit logging        | Auth, alert, incident workflow, assignment, duplicate merge, abuse review, shelter capacity, and admin actions have MVP audit events.                                                  |
| Public intake abuse  | Incident intake includes rate limits, suspicious report signals, false-report closure workflow, and human review for urgent reports.                                                   |
| Privacy              | Anonymous and contact-permission controls sanitize authority incident views unless the actor has the required authority role and completed MFA.                                        |
| Media storage        | Media metadata is private by default, content type and size are validated, and object-storage signed URL behavior is documented for production hardening.                              |
| ML safety            | Flood predictions include confidence, explanation factors, model version, and no-auto-publish flags. ML review can draft alerts but cannot publish them.                               |
| Runtime HTTP posture | Go APIs now apply defensive HTTP headers and support `NADAA_ALLOWED_ORIGINS` CORS allowlists. Empty or `*` keeps local-development wildcard behavior.                                  |
| Container posture    | Go services already run as non-root `appuser`. Web app images now use the unprivileged nginx image and explicit non-root `USER 101`.                                                   |
| Environment handling | `.gitignore` blocks real `.env` files while allowing `.env.example` templates. Staging now documents `NADAA_ALLOWED_ORIGINS`.                                                          |

## Findings

| Severity | Status   | Finding                                                                                       | Resolution                                                                                                                                                                      |
| -------- | -------- | --------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| High     | Resolved | Go APIs sent wildcard CORS responses without a runtime allowlist and lacked security headers. | Added `NADAA_ALLOWED_ORIGINS`, origin-varying CORS responses, `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, CSP, HSTS, and `Cache-Control: no-store` headers. |
| High     | Resolved | Web app containers did not explicitly declare a non-root runtime user.                        | Switched citizen, authority, dispatcher, and admin web images to `nginxinc/nginx-unprivileged:1.27-alpine` and `USER 101`.                                                      |
| Medium   | Resolved | Security checks were manual and could drift across services or Dockerfiles.                   | Added `pnpm security:scan` and CI coverage for API hardening tokens, non-root containers, env-file guardrails, and review documentation.                                        |
| Medium   | Resolved | Staging environment templates did not document approved browser origins for API CORS.         | Added `NADAA_ALLOWED_ORIGINS` to `infra/staging/staging.env.example` and deployment docs.                                                                                       |

## Scan And Verification Commands

Run these before UAT or after any security-sensitive change:

```bash
pnpm security:scan
pnpm audit --audit-level high
pnpm validate:docs
pnpm lint
pnpm typecheck
pnpm test
pnpm build
pnpm go:test
```

Container build validation remains in CI through the Docker matrix. The static security scan checks for non-root runtime declarations, but image vulnerability scanning should be added with registry or CI tooling when the deployment target is selected.

## Residual Risks

| Risk                                                 | Severity | Acceptance / Next Action                                                                                                                                             |
| ---------------------------------------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Header-based authority context in Go services        | High     | Accepted for MVP fixtures only. Replace with shared bearer-token middleware and centralized policy enforcement before production.                                    |
| In-memory service stores                             | High     | Accepted for MVP implementation slices. Persist production data in Postgres/PostGIS/object storage with backups, retention policy, and migration validation.         |
| Mock MFA                                             | High     | Accepted only for local/staging fixture use. Connect a production MFA provider and disable exposed OTP behavior before beta or production.                           |
| Fixture ML model                                     | Medium   | Accepted for decision-support contract testing. Do not use as production alerting intelligence until official data, model governance, monitoring, and review pass.   |
| External image and dependency vulnerability scanning | Medium   | `pnpm audit` and repo static checks are available now. Add Trivy, registry scanning, or equivalent container/SBOM scanning in CI/CD once registry credentials exist. |
| Final legal/data-retention policy                    | Medium   | Ghana data-protection and agency retention requirements still need stakeholder confirmation before production launch.                                                |

## UAT Gate Recommendation

NADAA can proceed toward UAT only if the verification commands pass, staging uses an explicit `NADAA_ALLOWED_ORIGINS` allowlist, no real secrets are committed, and the residual risks above are included in the UAT sign-off package.
