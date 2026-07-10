# Resilience & Operations Runbook (NADAA-170)

**Owner:** NADAA Platform Operations · **Applies to:** all `services/*` and the
web apps · **Companion docs:** [observability.md](../observability.md),
[deployment.md](../deployment.md), [security.md](../security.md).

This runbook prepares NADAA for high-volume national events: it defines the load
profiles the platform is sized against, how to run load and queue tests, the
disaster-recovery and failover procedure, the incident-response flow, and the
backup/restore validation checklist.

> **Be Aware. Be Prepared. Be Safe.** During a national event the platform's job
> is to keep citizen alerts and dispatcher decisions flowing. Availability of the
> alert path takes priority over every non-critical feature.

---

## 1. National event load profiles

Three profiles model demand from calm to a national flood emergency. They are
encoded in `scripts/load-test.mjs` and drive the SLO gate.

| Profile    | Scenario demand                    | Concurrency | p95 budget | Error budget |
| ---------- | ---------------------------------- | ----------- | ---------- | ------------ |
| `baseline` | Normal operations                  | 8           | 150 ms     | 1%           |
| `elevated` | Active watch/warning in one region | 32          | 400 ms     | 2%           |
| `surge`    | National multi-region event        | 96          | 1200 ms    | 5%           |

Each profile exercises the four demand drivers of a flood event:

1. **Report spike** — citizen incident intake feed (`incident-service`).
2. **Alert delivery** — citizen alert feed + delivery (`notification-service`).
3. **Dashboard maps** — geospatial risk + shelter reads (`risk`/`shelter`).
4. **Notification queue** — delivery-log depth (`notification-service`).

Sizing intent: the platform must sustain `surge` with graceful latency
degradation (no errors above budget), and recover to `baseline` latency within 5
minutes of demand returning to normal.

---

## 2. Running load and queue tests

Start the target services (see `deployment.md`), then:

```bash
pnpm load:baseline     # or load:elevated / load:surge / load:test
```

Point at a non-local environment with the shared `*_API_URL` variables:

```bash
INCIDENT_API_URL=https://staging.example/incident/api/v1 \
NOTIFICATION_API_URL=https://staging.example/notify/api/v1 \
RISK_API_URL=https://staging.example/risk/api/v1 \
SHELTER_API_URL=https://staging.example/shelter/api/v1 \
LOAD_PROFILE=surge pnpm load:test
```

The harness reports throughput, `p50/p95/p99`, and error rate per scenario and
exits non-zero when a scenario breaches the profile budget (SLO gate for CI /
release readiness). Unreachable services are reported explicitly; use
`LOAD_STRICT=true` to treat unreachable as a failure. The harness is read-only and
safe to run repeatedly.

**Queue test:** during a `surge` run, watch the notification queue-depth dashboard
(observability Dashboard 3). Delivery must drain to within the profile target and
the backlog age must not grow unbounded.

---

## 3. Disaster recovery & failover

**Objectives:** RTO ≤ 30 min, RPO ≤ 5 min for approved alerts and incident state.

### Failure modes and response

| Failure                      | Detection                            | Response                                                                                                  |
| ---------------------------- | ------------------------------------ | --------------------------------------------------------------------------------------------------------- |
| Single service instance down | `/healthz` red, 5xx spike            | Restart/redeploy the instance; traffic shifts to healthy replicas.                                        |
| Whole service down           | All scenarios for that service error | Roll back to last-good image; if persistent, degrade gracefully (below).                                  |
| Region/zone loss             | Regional health red                  | Fail over to the standby region; repoint DNS/load balancer; verify `/healthz` across services.            |
| Data store loss              | Write errors, restore needed         | Restore from the latest validated backup (§5); replay since RPO checkpoint.                               |
| Telecom/provider outage      | Delivery success drops               | Adapters record `skipped`/`failed` and fall back (SMS/push/voice); never silently drop an approved alert. |

### Graceful degradation order

Shed load in this order to protect the alert path:

1. Non-essential dashboards and analytics reads.
2. Open-data portal and campaign browsing.
3. Rich map layers (fall back to list views).
4. **Never** shed: citizen alert feed, approved-alert delivery, incident intake,
   112 guidance.

### Failover procedure

1. Declare the incident (§4) and assign an incident commander.
2. Confirm scope with `/healthz` across services and the observability dashboards.
3. Execute the relevant response from the table above.
4. Verify the alert path end-to-end: alert feed loads, a test delivery logs, the
   dispatcher dashboard renders.
5. Communicate status; keep the audit trail.

---

## 4. Incident response

**Severity:** SEV1 = citizen alert path degraded/down (national impact); SEV2 =
single-region or single-service degradation; SEV3 = non-critical feature.

1. **Detect** — alert fires (observability) or report received.
2. **Declare** — open an incident channel, assign an incident commander and scribe,
   set severity.
3. **Assess** — identify blast radius from the dashboards; check recent deploys
   (`git log`) and config changes first (a signal that pattern-matches a known
   failure may have a different cause — confirm before restarting/rolling back).
4. **Mitigate** — apply the DR/failover response; degrade gracefully if needed.
5. **Verify** — confirm the golden signals return within SLO and the alert path is
   healthy end-to-end.
6. **Communicate** — regular status updates to NADMO and stakeholders.
7. **Review** — blameless post-incident review; file follow-ups; update this
   runbook and add regression coverage.

Escalation: SEV1 pages the on-call platform lead and the NADMO duty officer
immediately. Keep public-safety decisions human-approved throughout.

---

## 5. Backup & restore validation

**Cadence:** continuous/near-real-time backups of durable state; a full
restore-and-verify drill at least once per release and before forecast high-risk
seasons.

### Restore drill checklist

- [ ] Identify the latest backup and confirm it is within the RPO window.
- [ ] Restore into an isolated environment (never over production).
- [ ] Verify integrity: record counts and checksums match the backup manifest.
- [ ] Bring services up against the restored data; `/healthz` green across all.
- [ ] Functional check: approved alerts present, incident state intact, audit logs
      continuous, delivery logs consistent.
- [ ] Run `pnpm load:baseline` against the restored stack — latency/error within
      budget.
- [ ] Record RTO achieved, RPO gap, and any data loss; file follow-ups.
- [ ] Tear down the isolated environment.

> Note: the MVP services use in-memory stores; the durable equivalents are the
> core database schema and audit tables. This drill targets those durable stores
> in a deployed environment. Validate that a restore preserves the audit trail —
> public-safety accountability depends on it.

---

## 6. Pre-event readiness checklist

- [ ] Observability dashboards and alerts live ([observability.md](../observability.md)).
- [ ] `pnpm load:surge` passes against the target environment.
- [ ] DR/failover procedure rehearsed; standby region verified.
- [ ] Latest backup restore-validated (§5).
- [ ] On-call roster, escalation path, and NADMO duty officer confirmed.
- [ ] Provider/telecom fallbacks confirmed (SMS/push/voice, cell broadcast per its
      [compliance runbook](cell-broadcast-compliance.md)).
- [ ] Capacity headroom reviewed against the `surge` profile.
