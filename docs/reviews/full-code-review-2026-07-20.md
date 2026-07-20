# NADAA Monorepo Full Code Review

**Date:** 2026-07-20
**Branch:** `main` (HEAD `74a04e7`, clean tree at review start)
**Scope:** 18 Go services, 8 frontend/mobile apps, 2 shared packages, 48 scripts, database migrations/seeds, infra & CI
**Reviewer:** Automated codebase review — 32 parallel area reviewers covering every module, the NADAA-183 delta, and a re-verification pass over the 2026-07-17 remediation claims; fixes applied by 12 scoped fix agents plus the lead reviewer
**Previous review:** `docs/reviews/full-code-review-2026-07-17.md` (b768e6d + remediation dd7ef78). This review covers the delta since (NADAA-183) and re-examines the whole tree.

---

## Executive Summary

The 2026-07-17 remediation (dd7ef78) largely held up: every one of the 10 spot-checked headline claims (bearer-token verification everywhere, route-service `/api/v1` paths, citizen re-login, PBKDF2, ID-collision fixes, command alarm, fixture clocks, security-scan coverage, donation transient-error handling, WhatsApp full-phone conversation keys) verified as correctly implemented. Automated gates were green at HEAD (go test/vet 18/18, golangci-lint 0 issues, tsc typecheck 10/10, security-scan 197 checks).

This sweep found a **new class of defects: incomplete remediations and remediation-introduced regressions**, plus fresh bugs in NADAA-183 and previously unreviewed corners. **1 Critical, 29 High, ~70 Medium, 108 Low** (212 total, deduplicated).

**Top risks found:**

1. **MFA was unenrollable in production (Critical):** the enrollment code only left the server via a dev-gated field, the "secret" was a random ID, and the stored second factor was a static, never-rotating 6-digit code — while the consoles told users to use an authenticator app. Every provisioned user would have been locked out forever; the bootstrap admin's "one-time" code was actually its permanent MFA secret.
2. **Authz gaps on token type:** any self-issued _citizen_ token (same signing secret) unlocked the full authority alert list including drafts and ML `sourcePrediction` (alert-service), and all of ml-service including CPU-bound simulations.
3. **Relief-integrity:** anonymous unvetted pledges flipped aid requests to `fulfilled` and erased them from the donor feed (shelter-service, proven end-to-end), while donation-service's new email-match check 403'd every authority-dashboard pledge; the public aid list leaked donor PII and anti-fraud notes.
4. **Fixture/fabricated data still reached live surfaces:** fixture alerts were deliverable via real push/SMS and always present in the citizen feed (notification-service server side); location-less incident reports plotted at (0,0); dispatcher-web invented districts (Kumasi → "Ablekuma West") that became real alert targets; dispatcher-mobile hydrated a fake life_threatening queue on every cold start; citizen-web seeded fake flood-watch notifications.
5. **Broken-contract regressions from dd7ef78:** ml-service's new auth gate 401'd all three authority-dashboard ML panels (they send no credentials); open-data audit forwarding POSTed to an endpoint that never existed, with no service auth; route-service's risk URL pointed at a dead port; school evacuation-point saves 400'd by default from the dashboard.

## Remediation outcome

**All Critical, High, and Medium findings were fixed in the working tree** (uncommitted): **227 files changed, +8,932 / −1,828** across all 18 services, 8 apps, 2 packages, scripts, database, and infra. Every Go fix ships with regression tests; every service re-verified `go test` + `go vet` + `golangci-lint` (0 issues); every app re-verified `tsc --noEmit`. 108 Low findings are documented as backlog (not fixed).

---

## Key Fixes

### Critical

- **Real TOTP MFA (auth-service + all four authority consoles).** RFC 6238 TOTP implemented stdlib-only (proven against the RFC Appendix-B vectors): setup returns `{secret, otpauthUrl}`, enrollment verifies the authenticator's current code, login verifies the current window — no static/reusable codes anywhere. Bootstrap admin takes `NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_SECRET` (base32 seed); when unset a random secret's otpauth URL is logged once. The retired `NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE` is ignored with a loud startup WARN. Consoles (admin/agency/authority/dispatcher) replaced fake MFA toggles with the real two-step enrollment flow (password confirm → secret + otpauth URL → live code verify); non-functional "Disable MFA" buttons were removed (no disable endpoint exists; authority MFA is mandatory).

### High (services)

- **alert-service:** authority detection now requires `typ=="agency"` — citizen tokens get the public-only alert view (drafts + `sourcePrediction` no longer leak).
- **notification-service:** fixture alerts are excluded from delivery eligibility and the citizen feed unless `NADAA_ENV=development` or explicit opt-in (`NADAA_NOTIFICATION_ALLOW_FIXTURE_ALERTS=true`); upstream failure now yields an explicit `degraded` feed status instead of silent fixture substitution.
- **incident-service:** `Location` is now `*Coordinates` — `(0,0)` rejected (`invalid_location`), absent location stored as null (USSD reports work, no more Gulf-of-Guinea pins); `false_report` via PATCH /status restricted to abuse-review roles; volunteer registration requires a citizen token (identity from claims, rate-limited, idempotent per `citizenUserId`); volunteer-task audit events attribute the verified actor; media bytes now really upload (`PUT/GET /api/v1/media/{id}/content`, disk storage, absolute `uploadUrl` via `NADAA_INCIDENT_PUBLIC_BASE_URL`).
- **donation-service:** authority callers skip the pledge email-match (dashboard pledges work again); fulfillment counts **delivered** quantities only — fake pledges can no longer fulfill requests; `AllocatePledge` accumulates tranches with over-allocation guard; payment references can't re-issue after a same-day restart; `POST /donations` is rate-limited.
- **shelter-service:** anonymous aid-request list/get returns a public view model (no pledge contacts, no anti-fraud/review internals); only `cleared`+accepted pledges count toward fulfillment; pledge creation rate-limited; hospital sub-capacities bounded by `totalBeds`; G706 path-value log injection closed in all handlers; pause/close preserves original approval metadata.
- **ml-service:** access gate fails closed (no token configured → closed; citizen tokens → 403); new `PATCH /api/v1/cv/results/{id}/review` persists human CV review; FIFO caps (500 prediction logs / 100 simulations / 500 CV results) + limit/offset pagination end unbounded memory growth; invalid simulation params rejected instead of coerced.
- **route-service:** `RISK_SERVICE_URL` default fixed to `:8081`; upstream closure/risk failures now set `degraded:true` + `enrichmentStatus` (no more silent all-clear routes); the final polyline is re-sampled against hazards after detour insertion; risk probes run concurrently within budget; 1 MiB body cap.
- **open-data-service:** audit forwarding now sends `X-NADAA-Service-Token` to auth-service's new `POST /api/v1/audit/logs` ingest (constant-time compare, fail-closed); downloads serve real serialized CSV/JSON bytes with real sizes; review decisions are audit-logged and re-review returns 409.
- **risk-service:** `RecommendedActions` now uses the ML-elevated level. **road-closure-service:** in-window scheduled closures surface as active. **school-service:** evacuation-point coordinates/capacity validated. **missing-person-service:** re-review no longer resets `located`→`active`; public intake rate-limited + capped. **imagery-service:** retention lifecycle deletes files, expired downloads refused (410); geometry size/position caps. **campaign/incident/route/shelter/missing-person:** request body caps.
- **All 17 mock-actor-supporting services** now fail closed: `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` is rejected at startup unless `NADAA_ENV=development` (guide-service has no mock-actor path — verified N/A).

### High (apps)

- **authority-dashboard:** the three ML panels send `authorityHeaders()` (no more 401s); school evacuation points emit the nested `{label, location:{lat,lng}}` contract; CV review calls the real ml-service endpoint; approval queue no longer capped at 4; fake aid-request delete removed.
- **admin-web:** MFA enrollment displays the real secret/otpauth URL; agency_admin gets a usable console (agency locked to session, 403-scoped views instead of console-wide outage); users directory loads from the new `GET /auth/agency-users`; fabricated agency metrics replaced with computed/null-honest ones.
- **citizen-mobile:** no fixture risk on cold start (auto-refresh, honest empty/loading states); location grant now reads the real device position for risk/shelters/report prefill (hardcoded Accra coordinates gone); volunteer registration sends the bearer token and collects real community/skills; fabricated registration constants gone.
- **citizen-web:** fabricated seeded notifications deleted; media byte upload goes to the absolute service URL and fails visibly instead of silently; service worker caches the app shell (offline cold start works); pledge identity fields lock to the registered donor.
- **dispatcher-mobile:** Refresh button no longer passes the gesture event as a session (manual refresh works); real AsyncStorage replaces memory-only persistence; cold start shows honest empty/sign-in states instead of a fixture queue with a fake life-threatening incident; 401/403 routes to an auth-expired state; password/MFA fields use `secureTextEntry`.
- **dispatcher-web:** `districtFromCoordinates` no longer fabricates districts — outside the catalog shows "District unavailable" and leaves alert targets unselected; fabricated responder ETA/agency replaced with real assignment data or "—".
- **agency-web / authority-dashboard / dispatcher-web / admin-web:** all four password "change" mocks now call the real `POST /api/v1/auth/agency/password` (lockout + complexity enforced server-side).

### Scripts / infra / database

- Three ml-service smokes (`smoke-ml-review`, `smoke-cv-verification`, `smoke-predictive-positioning`) send `X-NADAA-Service-Token`; `smoke-donation` spawns a built binary (no more orphaned `go run` server); both dev scripts export the cross-service `*_SERVICE_URL` values for their 94xx topology and the renamed `NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_SECRET` (with a documented dev TOTP seed).
- `render.yaml`: ML URL comment now requires the `/api/v1` suffix; dead school-service `RISK_SERVICE_URL` removed; `nadaa-internal-service` group attached to open-data-service (audit forwarding can authenticate); MFA key renamed. `services/auth-service/Dockerfile`: `EXPOSE 8082` → `8080`.
- **New `database/migrations/008_fix_remaining_id_alignment.sql`** (idempotent): converts `users.id`, `incidents.id` + actor/merge columns, `audit_logs.actor_user_id`, and the road-closure/relief-point/aid tables from UUID to TEXT (dropping defaults and FKs the services' prefixed string IDs could never satisfy), and adds the incident abuse-review/status-lifecycle columns incident-service persists. Seed agency IDs realigned with incident-service triage fixtures. `database/README.md` lists all 8 migrations.

---

## Notable deviations (deliberate, documented in code)

- incident-service: `closed` was **not** restricted to abuse-review roles (responders legitimately close after response); only `false_report` was gated.
- missing-person-service: first approval still defaults `pending_review`→`active` (the documented intake flow); only _re-review_ preserves status.
- donation-service: reference monotonicity holds across restarts that reload the day's records (the documented persistence path); a format change was rejected to preserve the `GIFT-<date>-<n>` contract.
- Consoles: "Disable MFA" controls removed rather than wired — no disable endpoint exists and authority MFA is mandatory.
- admin-web: any 401 (including a wrong-current-password 401) trips the global sign-out guard — pre-existing platform behavior, flagged for a future pass.

## Backlog (108 Low findings, not fixed)

Representative themes: imagery CORS/Host edge cases, remaining per-service log-field hygiene, alert-service stale workflow fields after reject→edit, non-constant-time compares in legacy paths, CSV export polish, mobile double-notify races, Expo peer-dependency packaging, `@nadaa/brand` src-vs-dist publishing, shared-types file size, frontend behavioral test absence (app `test` = `tsc --noEmit` only), in-memory store / per-process rate-limiter architecture (accepted MVP trade-off). Full list retained in the review working notes; none affect the safety invariants.

## Safety invariants — status

| Invariant                                                 | Status                                                                                                                                                                            |
| --------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ML predictions never auto-send public alerts              | Holds (re-verified; ml-service gate hardened further)                                                                                                                             |
| Mass alerts require authority approval + audit            | **Strengthened** — fixture alerts can no longer bypass approval at delivery; emergency override now has separation-of-duties; audit snapshots capture full citizen-facing content |
| Authority users require RBAC + MFA                        | **Now real** — TOTP MFA is enrollable and rotating; agency-viewer/viewer role gaps closed in school/incident/volunteer paths                                                      |
| Anonymous citizen reports preserve privacy                | **Improved** — location-less reports no longer geo-fabricated; pledge/donor PII leaks closed                                                                                      |
| Life-threatening reports never hidden by automated scores | **Improved** — `false_report` restricted to abuse-review roles on the status endpoint                                                                                             |

## Verification (executed at the final tree)

```bash
go test ./... && go vet ./...          # 18/18 services PASS (incl. new regression suites)
golangci-lint run ./...                # 0 issues on all 18 services
pnpm -r typecheck                      # PASS (10 projects)
node scripts/security-scan.mjs         # PASS — 197 checks, 18 modules
pnpm validate:docs                     # PASS (dashboard contract, database assets, flood-risk, release docs)
node scripts/smoke-donation.mjs        # PASS end-to-end (live service, new pledge semantics)
node scripts/smoke-community-volunteers.mjs  # PASS (citizen-token registration flow)
# Plus per-agent live proofs: RFC 6238 test vectors, TOTP setup→verify→login flow,
# media initiate→PUT→GET round-trip, fixture-alert 404 in production mode,
# anonymous aid-list PII redaction, pledge-fulfillment regression tests.
```

## Follow-ups requiring human action

1. **Review and commit** the 227-file working-tree change (nothing was committed by the review).
2. Any environment still setting `NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE` must migrate to `NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_SECRET` (base32 seed) — the old var is warn-and-ignored.
3. Render: first deploy with `NADAA_INCIDENT_MEDIA_STORAGE_PATH=/app/uploads/media` is ephemeral — attach a disk if media must survive redeploys.
4. Consider a citizen-accessible `GET /volunteers/me` endpoint so mobile clients can re-fetch profiles without re-registering (currently idempotent re-register is the workaround).
5. The global 401→sign-out guard in the web consoles also fires on wrong-current-password 401s from the password endpoint — worth scoping per-endpoint in a future pass.
