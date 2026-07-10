# NADAA Monorepo Full Code Review

**Date:** 2026-07-09  
**Branch:** `main` (HEAD `fcd8294`)  
**Scope:** Go services, frontend apps, shared packages, scripts/tooling, documentation  
**Reviewer:** Automated codebase review

---

## Executive Summary

**Overall health score: 6 / 10**

The NADAA monorepo is a well-organized, feature-rich MVP with consistent modular structure, passing CI gates, and strong documentation discipline. However, several **critical** and **high** severity issues remain that block production readiness:

1. **No real cross-service authentication:** every non-auth service trusts easily-forged `X-NADAA-*` headers instead of verifying bearer tokens.
2. **Weak password hashing:** agency passwords are stored as unsalted SHA-256 digests.
3. **Broken services in the tree:** `campaign-service`, `open-data-service`, and `school-service` have `go.mod` files but do not compile; root Go scripts will fail on them.
4. **Security scan is silently no-op:** it looks for `services/<name>/main.go`, but every service was moved to `cmd/server/main.go`.
5. **Frontend "tests" are only type checks:** no unit/integration test runner exists, and production UIs use hardcoded authority sessions.

Top 5 risks:

1. Header-spoofing trivially grants authority access across incident/alert/shelter/road-closure/donation/damage-claim/etc.
2. Stolen/leaked password database exposes agency credentials because of fast, unsalted hashes.
3. Broken modules pollute the workspace and break root `pnpm go:test/vet/lint`.
4. In-memory stores mean total data loss on restart and no concurrency isolation beyond a single process.
5. Hardcoded dispatcher/admin/agency sessions and fake bearer tokens in production bundles.

---

## Findings by Area

### Go Services

#### Auth Service (`services/auth-service`)

| Severity     | Finding                                                                                                                                                       | Recommendation                                                                       |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| **Critical** | Agency passwords are hashed with `sha256.Sum256` and no salt (`internal/utils/utils.go:315`). This is vulnerable to rainbow tables and fast offline cracking. | Replace with `bcrypt`/`argon2` and a per-credential salt.                            |
| High         | `NADAA_AUTH_TOKEN_SECRET` defaults to hardcoded `dev-secret-change-me` (`internal/config/config.go:29`).                                                      | Require a strong secret at startup; fail fast if unset in non-dev environments.      |
| High         | `NADAA_AUTH_EXPOSE_DEV_OTP` and fixed `NADAA_AUTH_MOCK_OTP=123456` can leak into production.                                                                  | Disallow these env vars unless an explicit `NADAA_ENV=development` flag is set.      |
| Medium       | Bootstrap admin password/MFA code come directly from env without complexity checks.                                                                           | Enforce minimum password length and MFA secret entropy.                              |
| Low          | `cmd/server/main.go`, `store`, `models`, `utils` packages have no tests; only `handlers` is tested.                                                           | Add unit tests for token signing/verification, password hashing, and config loading. |

#### Incident Service (`services/incident-service`)

| Severity     | Finding                                                                                                                                                                                | Recommendation                                                                           |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| **Critical** | `requireAuthority` reads `X-NADAA-*` headers and never validates the bearer token (`internal/handlers/middleware.go:78`). Any caller can impersonate any role/agency.                  | Add shared JWT/HMAC token middleware or call `auth-service` for every protected request. |
| High         | Volunteer task status/observation endpoints (`PATCH /volunteer-tasks/{id}/status`, `POST /volunteer-tasks/{id}/observations`) have **no authentication** (`volunteers.go:157`, `183`). | Require citizen or volunteer token; verify task ownership.                               |
| Medium       | Rate limiter is in-process and keyed by `ClientIdentifier` (best-effort IP), so it is ineffective behind load balancers and per-process only.                                          | Move rate limiting to Redis with a consistent key strategy.                              |
| Medium       | `ListIncidents` returns unsorted map iteration before `sort.Slice`; fine now but could regress if sort is removed.                                                                     | Keep sorted guarantee and add tests.                                                     |
| Low          | Some `kept := events[:0]` slice reuse in rate limiter is safe because it holds under lock.                                                                                             | Add a comment and test for the windowed eviction behavior.                               |

#### Alert Service (`services/alert-service`)

| Severity     | Finding                                                                                                                                             | Recommendation                                                                         |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| **Critical** | Same header-only auth as incident service (`internal/handlers/middleware.go:11`). Anyone can create, submit, approve, or emergency-override alerts. | Enforce bearer-token validation and integrate with auth-service before any write path. |
| High         | `emergencyOverrideHandler` transitions to `emergency_override` status; ensure this still triggers the same delivery gate as `approved`.             | Add an explicit end-to-end smoke test for emergency override → citizen feed.           |
| Medium       | `listAlertsHandler` exposes non-public alerts when authority headers are present (`hasAuthorityHeaders`), but it does not verify those headers.     | Make public/private determination depend on a verified token.                          |

#### Shelter / Road-Closure / Donation / Damage-Claim / Imagery / Missing-Person Services

| Severity     | Finding                                                                                                                                                                 | Recommendation                                                        |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| **Critical** | All use the same forgeable `X-NADAA-*` header pattern (`requireAuthority` / `authorityContextFromRequest`) and never verify tokens.                                     | Centralize auth middleware and roll it out to every service.          |
| High         | `damage-claim-service` fetches incident details over HTTP to a configurable `INCIDENT_API_URL` without timeouts or auth forwarding (`internal/handlers/claims.go:200`). | Add request timeouts and forward the caller's verified token/context. |
| Medium       | `donation-service` `createDonorHandler` allows public creation but optionally attaches any actor ID from unverified headers (`donors.go:62`).                           | Remove optional authority attribution until auth is enforced.         |
| Low          | Several services use identical `utils.go` helpers (CORS, JSON, validation) duplicated across modules.                                                                   | Extract a shared Go package once interfaces stabilize.                |

#### Route Service (`services/route-service`)

| Severity | Finding                                                                                                                    | Recommendation                                                                               |
| -------- | -------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Low      | Previously reported `errcheck` warning is no longer present; `golangci-lint` and `go vet` report 0 issues.                 | —                                                                                            |
| Medium   | Route planning uses a naive perpendicular detour algorithm that can route through blocked areas or out-of-bounds geometry. | Document as decision-support only (already done) and add tests for corner cases.             |
| Medium   | External service calls (`fetchJSON`) do not forward the caller's auth context and silently degrade on upstream failure.    | Add timeouts are present, but add circuit-breaker/fallback semantics and structured metrics. |

#### Risk / Guide / ML / Integration / Notification Services

| Severity | Finding                                                                                                                                                                                                        | Recommendation                                                                |
| -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| High     | `risk-service` enriches with `ml-service` but calls it without propagating auth; model artifact loading from disk in `ml-service` (`ml-service/internal/utils/utils.go:130`) has no integrity/signature check. | Verify model checksums/signatures and load from object storage in production. |
| Medium   | `notification-service` alert delivery endpoint accepts any `alertId` path value and does not verify the caller is authorized to trigger that alert.                                                            | Add token/RBAC check before attempting delivery.                              |
| Medium   | `integration-service` and others log actor IDs from unverified headers, creating false audit attribution.                                                                                                      | Do not trust actor metadata until the token is verified.                      |

#### Broken / Unfinished Services

| Severity     | Finding                                                                                                                                                    | Recommendation                                                                                        |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| **Critical** | `campaign-service`, `open-data-service`, and `school-service` contain `go.mod` but fail `go vet`/`golangci-lint` due to missing symbols/undefined methods. | Either complete and test these services or remove their `go.mod` and exclude them from the workspace. |
| High         | `dispatch-service` exists as an empty directory with no `go.mod`.                                                                                          | Remove the directory or document it clearly as a placeholder.                                         |

---

### Frontend Apps

#### Dispatcher Web (`apps/dispatcher-web`)

| Severity     | Finding                                                                                                                                           | Recommendation                                                                                       |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| **Critical** | `dispatcherSession` is hardcoded (`src/app/session.ts:11`) and the UI sends unverified `X-NADAA-*` headers; there is no bearer-token integration. | Implement real login flow, token storage, and send `Authorization: Bearer` to a token-aware gateway. |
| High         | `DispatcherCommandApp.tsx` is a single 1900-line component holding dozens of `useState` hooks and fallback fixtures.                              | Split into smaller feature components/hooks and remove fixture fallback for production.              |
| Medium       | Built bundle is ~719 KB JS (gzip 214 KB) without code splitting.                                                                                  | Add route-based lazy loading and reduce initial bundle.                                              |
| Low          | Map and chart libraries are bundled even if the current view does not need them.                                                                  | Lazy-load Leaflet/map components.                                                                    |

#### Admin Web / Agency Web / Authority Dashboard

| Severity     | Finding                                                                                                                                                                                         | Recommendation                                                       |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| **Critical** | Same hardcoded-session pattern: `adminSession.token = "local-admin-token"` and `agencySession.token = "local-agency-token"` (`admin-web/src/auth/session.ts`, `agency-web/src/app/session.ts`). | Replace with real auth-service integration.                          |
| High         | `authority-dashboard` is described as a compatibility shell but still builds and ships a large bundle; it duplicates dispatcher functionality.                                                  | Deprecate and redirect to `dispatcher-web`/`agency-web`/`admin-web`. |

#### Citizen Web / Citizen Mobile

| Severity | Finding                                                                                                               | Recommendation                                            |
| -------- | --------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| Medium   | Citizen web stores offline guides in `localStorage` (`GUIDE_CACHE_KEY`) which is synchronous and limited to ~5 MB.    | Use IndexedDB for larger/offline content.                 |
| Medium   | Mobile app has all Expo/React Native deps marked `peerDependencies` and `optional`; build reproducibility is fragile. | Pin required native dependencies as regular dependencies. |
| Low      | Several fallback fixtures are embedded in production bundles, increasing size.                                        | Serve fixtures from a dev-only module or lazy-load them.  |

#### Marketing Web

| Severity | Finding                                                                                        | Recommendation                                      |
| -------- | ---------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| Low      | No accessibility violations detected via static inspection, but no automated a11y tests exist. | Add `@axe-core/react` or similar to CI smoke tests. |

---

### Shared Packages

| Severity | Finding                                                                                                                            | Recommendation                                                                               |
| -------- | ---------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Medium   | `@nadaa/brand` exports point directly at `src/*.ts`; packages do not ship compiled JS.                                             | Add a `tsup`/`tsc` build step and publish `dist` outputs if external consumers are expected. |
| Low      | `@nadaa/config` exists in the workspace but many apps import config from their own `src/app/config.ts`, duplicating API base URLs. | Centralize runtime config contract in `@nadaa/config`.                                       |
| Low      | `shared-types` is large (2467 lines) and mixes domain types with API shapes; breaking changes are hard to track.                   | Consider versioning the package or splitting domain vs. API types.                           |

---

### Scripts and Tooling

| Severity     | Finding                                                                                                                                                                                          | Recommendation                                                                                        |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------- |
| **Critical** | `scripts/security-scan.mjs` checks `services/<name>/main.go` (line 58), but every Go service moved to `cmd/server/main.go` in NADAA-171. All service HTTP-hardening checks are silently skipped. | Fix the path to `cmd/server/main.go` and add a negative test that fails when no services are checked. |
| High         | Root `pnpm go:test`, `go:lint`, `go:vet` iterate `services/*/go.mod`, so they fail on broken `campaign-service`, `open-data-service`, `school-service`.                                          | Exclude broken modules or fix them before running these scripts.                                      |
| High         | Frontend `lint`/`test` scripts all run `tsc --noEmit`; there is no actual test runner (Vitest/Jest).                                                                                             | Add a test framework and behavioral tests for at least auth, API, and report flows.                   |
| Medium       | `.github/workflows/ci.yml` has a duplicate `marketing-web` Docker matrix entry (lines 101–104).                                                                                                  | Remove duplicate.                                                                                     |
| Medium       | `pnpm audit` reports one moderate `uuid` vulnerability via Expo dependencies (`GHSA-w5hq-g745-h8pq`).                                                                                            | Upgrade `uuid` transitive dependency or override to >= 11.1.1.                                        |
| Low          | `go:lint` is not run in CI; only `go test` is.                                                                                                                                                   | Add a CI job for `golangci-lint` to catch the broken modules.                                         |

---

### Documentation

| Severity | Finding                                                                                                                                                     | Recommendation                                         |
| -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| Medium   | `docs/security.md` and `docs/security-review.md` correctly flag header-based auth as a residual high risk, but the security scan claims this is "resolved." | Align scan checks with documented residual risks.      |
| Medium   | `docs/api.md` describes bearer-token auth for authority endpoints, but the actual services use header-based auth.                                           | Update docs or implement the documented contract.      |
| Low      | `docs/architecture.md` lists `dispatch-service` as a placeholder, matching the empty directory.                                                             | Keep consistent; remove once the directory is removed. |
| Low      | Several READMEs under services still reference pre-modularization paths.                                                                                    | Audit and update service READMEs.                      |

---

## Security Audit

| #   | Concern                                                                                                 | Mitigation / Status                                                                                                        |
| --- | ------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Forgeable authority headers** grant full write access across services.                                | Acknowledged in `docs/security-review.md` as a residual high risk, but not gated in code. Must be fixed before production. |
| 2   | **Unsalted SHA-256 password hashes** for agency accounts.                                               | Replace with `bcrypt`/`argon2`; schedule password reset on migration.                                                      |
| 3   | **Hardcoded dev secrets** (`dev-secret-change-me`, mock OTP) have safe defaults that could be deployed. | Make auth service fail startup when secrets are weak/default outside development.                                          |
| 4   | **Volunteer task endpoints unauthenticated**.                                                           | Add citizen/volunteer token checks.                                                                                        |
| 5   | **CORS wildcard** when `NADAA_ALLOWED_ORIGINS` is unset.                                                | Acceptable for local dev; staging/prod must set explicit allowlists.                                                       |
| 6   | **Media upload URL is returned but no real object-store integration**; all media is in-memory metadata. | Implement signed-URL generation and object storage before accepting real media.                                            |
| 7   | **No XSS vectors found** (`dangerouslySetInnerHTML`, `eval`, inline scripts are absent).                | Good. Continue forbidding these patterns.                                                                                  |
| 8   | **No secrets committed** to tracked files beyond dev defaults.                                          | Good; `.gitignore` blocks `.env` files.                                                                                    |
| 9   | **Security scan silently passes** because it looks at wrong paths.                                      | Fix immediately.                                                                                                           |
| 10  | **npm audit** reports one moderate transitive vulnerability.                                            | Patch or override.                                                                                                         |

---

## Performance & Scalability

| #   | Bottleneck                                                                                              | Impact                                                 | Recommendation                                        |
| --- | ------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ | ----------------------------------------------------- |
| 1   | All service stores are in-memory maps protected by a single `sync.RWMutex`.                             | Cannot scale beyond one process; data lost on restart. | Implement Postgres/PostGIS persistence as documented. |
| 2   | Rate limiter, duplicate detection, and abuse scoring are per-process.                                   | Inconsistent limits across replicas; easy to bypass.   | Move to Redis with Lua/Redlock semantics.             |
| 3   | Dispatcher/authority web bundles exceed 700 KB JS without code splitting.                               | Slow first load on low-bandwidth devices.              | Add route-based dynamic imports and manual chunks.    |
| 4   | `route-service` calls shelter/closure/risk services synchronously per plan request.                     | Latency stacks and failures cascade.                   | Add caching, parallel fetching, and circuit breakers. |
| 5   | `incident-service` recomputes duplicate candidates and abuse signals on every create; no spatial index. | O(n) scan per report will degrade as volume grows.     | Add spatial index (PostGIS) and incremental scoring.  |

---

## Maintainability & Style

- **Strengths:**
  - Consistent package layout (`cmd/server`, `internal/config`, `internal/models`, `internal/store`, `internal/utils`, `internal/handlers`) across services.
  - Structured logging with `INFO`/`WARN`/`ERROR` levels in most services.
  - Good use of `go vet`, `golangci-lint`, and graceful shutdown patterns.
  - TypeScript strictness is enforced; all workspace projects typecheck.
  - Clear separation of web apps by role.

- **Weaknesses:**
  - Massive single-file components in dispatcher and citizen web apps.
  - Heavy duplication of `utils.go` helpers across Go services (acknowledged as intentional, but now stable enough to share).
  - `campaign-service`, `open-data-service`, `school-service` are dead weight.
  - Magic strings for roles/statuses are duplicated in multiple services instead of a shared Go constant package.
  - Frontend fixture data is deeply embedded in production components.

---

## Testing & Verification

| Area            | Coverage                                                  | Gaps                                                                                                      |
| --------------- | --------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Go handlers     | Each service has at least one `_test.go` file (22 total). | `config`, `models`, `store`, `utils` packages are mostly untested. No integration tests between services. |
| Go lint         | `golangci-lint` passes on 15 services; 3 are broken.      | Broken services are not exercised in CI.                                                                  |
| Frontend        | `test` script runs `tsc --noEmit` only.                   | No behavioral, a11y, or E2E tests.                                                                        |
| Smoke tests     | 42 scripts cover major flows.                             | They rely on hardcoded authority headers, so they do not test real security boundaries.                   |
| Security scan   | Passes 31 checks.                                         | Most service checks are no-op due to wrong path.                                                          |
| Docs validation | Passes.                                                   | Does not validate doc-to-code consistency.                                                                |

---

## Documentation & DevEx

- **Accurate:** `docs/architecture.md`, `docs/security.md`, and `docs/security-review.md` accurately describe the current state, including known residual risks.
- **Incomplete:** `docs/api.md` still implies bearer-token auth for authority endpoints, which is not implemented outside `auth-service`.
- **Friction:**
  - New contributors will hit broken Go services when running root `pnpm go:test`.
  - `pnpm security:scan` gives a false sense of security.
  - No `README` in `docs/reviews/` directory; review artifacts are not standardized.

---

## Recommended Priority Actions

1. **Fix cross-service authentication immediately.** Add a shared Go middleware that verifies the `nadaa.` HMAC bearer token (or a JWT) in every protected service. Gate all `X-NADAA-*` headers on a verified token and reject requests without it.
2. **Replace SHA-256 password hashing with `bcrypt`** (or `argon2id`) and migrate existing fixture/test credentials.
3. **Remove or repair broken services.** Either delete `campaign-service`, `open-data-service`, and `school-service` (and the empty `dispatch-service` directory) or bring them to a compiling, tested state.
4. **Fix `scripts/security-scan.mjs`.** Change the service main-file path to `cmd/server/main.go` and add a check that at least one service was validated.
5. **Add a real frontend test runner** (Vitest) and write behavioral tests for login, report submission, and alert workflow; replace hardcoded sessions with real auth integration.
6. **Enforce non-default secrets.** Make `auth-service` fail fast when `NADAA_AUTH_TOKEN_SECRET` matches a known dev default or is too short.
7. **Add authentication to volunteer task endpoints** and verify task/volunteer ownership.
8. **Add a CI lint job** (`pnpm go:lint`) so broken modules cannot re-enter `main`.
9. **Patch or override the transitive `uuid` vulnerability** reported by `pnpm audit`.
10. **Remove duplicate `marketing-web` entry** in `.github/workflows/ci.yml` and reduce dispatcher/authority bundle sizes with code splitting.

---

## Verification Commands Run

```bash
# Go
go test ./...          # 15 services pass; 3 broken modules fail
go vet ./...           # same split
golangci-lint run ./... # same split

# Node / pnpm
pnpm -r typecheck      # passed
pnpm -r lint           # passed (tsc only)
pnpm -r test           # passed (tsc only)
pnpm -r build          # passed
pnpm validate:docs     # passed
pnpm security:scan     # passed (but silently skips services)
pnpm audit             # 1 moderate uuid advisory
```
