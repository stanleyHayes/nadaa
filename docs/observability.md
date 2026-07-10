# Observability (NADAA-170)

This document specifies the dashboards and alerts that keep NADAA operable during
a national flood event. It is the contract between the services (what they must
emit) and the ops stack (what it must render and alert on). Panels are grouped
into five dashboards matching the NADAA-170 acceptance criteria: latency, errors,
queue depth, provider delivery, and geospatial query performance.

Every NADAA Go service emits structured `INFO`/`WARN`/`ERROR` logs with stable
keys (actor/agency/request ids, alert/incident ids, channel, provider, status,
counts). Until a metrics backend is provisioned, these logs are the source of
truth; the panels below are expressed as the log fields / derived metrics they
read, so they can be wired to either a log-based (Loki/CloudWatch) or metric-based
(Prometheus/Grafana) backend without changing the service contract.

## Golden signals and SLOs

| Signal                     | Target (baseline) | Target (surge) | Source                                 |
| -------------------------- | ----------------- | -------------- | -------------------------------------- |
| Request latency p95        | ≤ 150 ms          | ≤ 1200 ms      | request logs / `http_request_duration` |
| Error rate (5xx + refused) | ≤ 1%              | ≤ 5%           | request logs by status                 |
| Notification queue depth   | ≤ 1k pending      | ≤ 25k pending  | delivery-log backlog                   |
| Provider delivery success  | ≥ 99%             | ≥ 95%          | delivery attempts by `status`          |
| Geospatial query p95       | ≤ 200 ms          | ≤ 800 ms       | risk/shelter read logs                 |

The load harness (`pnpm load:test`, see the resilience runbook) checks the latency
and error budgets directly, per profile.

## Dashboard 1 — Latency

- **Request latency p50/p95/p99 by service** — one series per service (alert,
  notification, incident, risk, shelter, ml, route, …).
- **Latency by route** — top 10 slowest routes; watch alert delivery, incident
  intake, and the geospatial reads.
- **Latency heatmap over time** — surface long-tail regressions during surge.
- Alert: p95 above the profile budget for 5 minutes.

## Dashboard 2 — Errors

- **Error rate by service and status class** (4xx vs 5xx vs connection refused).
- **Top error codes** — e.g. `invalid_channel`, `mfa_required`, `not_found`,
  `rate_limit_exceeded` — from the `code` field of `WARN`/`ERROR` logs.
- **Rejected-write rate** — validation failures on incident/report intake.
- Alert: 5xx rate > profile error budget for 5 minutes; any sustained spike in
  `ERROR` logs.

## Dashboard 3 — Queue depth

- **Notification queue depth** — pending vs delivered vs skipped/failed delivery
  attempts over time (`GET /notifications/delivery-logs` by `status`).
- **Delivery throughput** — attempts/sec by channel (push, sms, voice,
  cell_broadcast).
- **Backlog age** — oldest undelivered attempt.
- Alert: queue depth above the profile target, or backlog age > 2 minutes during a
  live alert.

## Dashboard 4 — Provider delivery

- **Delivery success rate by channel** — push / sms / voice / cell_broadcast.
- **Provider outcome breakdown** — delivered / skipped / failed with `reason`.
- **Cell broadcast dispatches** — by channel identifier (4370/4371/4373), adapter
  (`sandbox_cbc`/`disabled_cbc`/live), and status (broadcast/simulated/skipped).
- **Voice/asset review latency** — generate → approve → deliver.
- Alert: success rate below the SLO for any channel; any unexpected live cell
  broadcast dispatch.

## Dashboard 5 — Geospatial query performance

- **Risk query latency** — `GET /risk?lat&lng` p50/p95/p99.
- **Shelter/relief nearby latency** — `GET /shelters/nearby`, `/relief-points/nearby`.
- **Geofence evaluation latency** — alert targeting/geofence checks.
- **Map read volume** — requests/sec from the authority and dispatcher dashboards.
- Alert: geospatial p95 above target during surge (early indicator of map-driven
  overload).

## Operational panels (cross-cutting)

- **Service health** — `/healthz` up/down per service.
- **Saturation** — CPU, memory, goroutines, open connections per instance.
- **Approval funnel** — human-approval steps (alerts, voice, cell broadcast) so
  operators can see review latency during an event.

## Wiring checklist

- [ ] Ship structured logs to the log backend with the stable field set retained.
- [ ] Derive the five dashboards above (log-based or metric-based).
- [ ] Configure the alerts noted per dashboard against the SLO table.
- [ ] Add `/healthz` uptime checks for every service.
- [ ] Link this spec from the on-call runbook (`docs/runbooks/resilience-and-operations.md`).
