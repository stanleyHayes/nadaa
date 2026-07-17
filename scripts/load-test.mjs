#!/usr/bin/env node
// NADAA-170 national-scale load harness.
//
// Drives closed-loop read load against the public endpoints that dominate a
// national flood event — citizens checking area risk while reports spike,
// alert delivery, dashboard/geospatial maps, and the notification queue — and
// reports throughput, latency percentiles, and error rate per scenario against
// a per-profile SLO budget. Every scenario URL must be publicly readable:
// authority feeds require a verified bearer token, so hitting them here would
// measure 401s, not load.
//
// Read-only by design so it is safe to run repeatedly against a live instance.
// Usage:
//   LOAD_PROFILE=baseline|elevated|surge node scripts/load-test.mjs
// Optional overrides: LOAD_DURATION_MS, LOAD_CONCURRENCY, LOAD_STRICT=true,
// LOAD_ALLOW_UNREACHABLE=true to skip scenarios whose service is down (by
// default an unreachable scenario fails the run), and the per-service
// *_API_URL variables shared with the smoke scripts.

const PROFILES = {
  // concurrency + duration model the closed-loop demand; the budgets are the
  // SLO gate the run is checked against.
  baseline: {
    concurrency: 8,
    durationMs: 4000,
    p95BudgetMs: 150,
    errorBudget: 0.01,
  },
  elevated: {
    concurrency: 32,
    durationMs: 6000,
    p95BudgetMs: 400,
    errorBudget: 0.02,
  },
  surge: {
    concurrency: 96,
    durationMs: 8000,
    p95BudgetMs: 1200,
    errorBudget: 0.05,
  },
};

const profileName = (
  process.env.LOAD_PROFILE?.trim() || "baseline"
).toLowerCase();
const base = PROFILES[profileName];
if (!base) {
  console.error(
    `unknown LOAD_PROFILE "${profileName}"; expected one of ${Object.keys(PROFILES).join(", ")}`,
  );
  process.exit(2);
}
const profile = {
  ...base,
  concurrency: Number(process.env.LOAD_CONCURRENCY) || base.concurrency,
  durationMs: Number(process.env.LOAD_DURATION_MS) || base.durationMs,
};
const strict = process.env.LOAD_STRICT?.trim() === "true";
// Unreachable scenarios fail the run by default; LOAD_ALLOW_UNREACHABLE=true
// opts back into skipping them (LOAD_STRICT still fails them).
const allowUnreachable =
  process.env.LOAD_ALLOW_UNREACHABLE?.trim() === "true" && !strict;

const services = {
  notification:
    process.env.NOTIFICATION_API_URL?.trim() || "http://127.0.0.1:8090/api/v1",
  risk: process.env.RISK_API_URL?.trim() || "http://127.0.0.1:8081/api/v1",
  shelter:
    process.env.SHELTER_API_URL?.trim() || "http://127.0.0.1:8093/api/v1",
};

const scenarios = [
  {
    // The incident list feed is authority-only; the public read that spikes
    // alongside report intake is citizens checking risk for their area.
    key: "report_spike",
    label: "Report spike — public area-risk lookups",
    urls: [`${services.risk}/risk?lat=5.5600&lng=-0.1870`],
  },
  {
    key: "alert_delivery",
    label: "Alert delivery — citizen alert feed",
    urls: [`${services.notification}/notifications/alerts?includeExpired=true`],
  },
  {
    key: "dashboard_maps",
    label: "Dashboard maps — geospatial risk + shelters",
    urls: [
      `${services.risk}/risk?lat=5.5600&lng=-0.1870`,
      `${services.shelter}/shelters/nearby?lat=5.5600&lng=-0.2000`,
    ],
  },
  {
    key: "notification_queue",
    label: "Notification queue — delivery logs",
    urls: [`${services.notification}/notifications/delivery-logs`],
  },
];

function percentile(sorted, p) {
  if (sorted.length === 0) return 0;
  const index = Math.min(
    sorted.length - 1,
    Math.ceil((p / 100) * sorted.length) - 1,
  );
  return sorted[Math.max(0, index)];
}

async function runScenario(scenario) {
  const deadline = Date.now() + profile.durationMs;
  const latencies = [];
  let ok = 0;
  let httpErrors = 0;
  let connErrors = 0;
  let cursor = 0;

  async function worker() {
    while (Date.now() < deadline) {
      const url = scenario.urls[cursor++ % scenario.urls.length];
      const start = performance.now();
      try {
        const res = await fetch(url, { method: "GET" });
        await res.arrayBuffer().catch(() => {});
        latencies.push(performance.now() - start);
        if (res.ok) ok += 1;
        else httpErrors += 1;
      } catch {
        latencies.push(performance.now() - start);
        connErrors += 1;
      }
    }
  }

  await Promise.all(
    Array.from({ length: profile.concurrency }, () => worker()),
  );

  const total = ok + httpErrors + connErrors;
  const unreachable = total > 0 && connErrors === total;
  latencies.sort((a, b) => a - b);
  const errorRate = total === 0 ? 1 : (httpErrors + connErrors) / total;
  const p95 = percentile(latencies, 95);
  const rps = total / (profile.durationMs / 1000);
  const pass =
    !unreachable &&
    errorRate <= profile.errorBudget &&
    p95 <= profile.p95BudgetMs;

  return {
    scenario,
    total,
    ok,
    httpErrors,
    connErrors,
    unreachable,
    errorRate,
    rps,
    p50: percentile(latencies, 50),
    p95,
    p99: percentile(latencies, 99),
    max: latencies.at(-1) ?? 0,
    pass,
  };
}

function fmt(ms) {
  return `${ms.toFixed(1)}ms`;
}

console.log(
  `NADAA load test — profile=${profileName} concurrency=${profile.concurrency} duration=${profile.durationMs}ms ` +
    `budget(p95<=${profile.p95BudgetMs}ms, err<=${(profile.errorBudget * 100).toFixed(0)}%)`,
);

const results = [];
for (const scenario of scenarios) {
  // Scenarios run sequentially so each measures an isolated demand window.
  results.push(await runScenario(scenario));
}

let failed = 0;
let unreachable = 0;
for (const r of results) {
  if (r.unreachable) {
    unreachable += 1;
    console.log(
      `  UNREACHABLE  ${r.scenario.label} — no service responded at ${r.scenario.urls[0]}`,
    );
    continue;
  }
  const verdict = r.pass ? "PASS" : "FAIL";
  if (!r.pass) failed += 1;
  console.log(
    `  ${verdict}  ${r.scenario.label}\n` +
      `        req=${r.total} rps=${r.rps.toFixed(0)} err=${(r.errorRate * 100).toFixed(2)}% ` +
      `p50=${fmt(r.p50)} p95=${fmt(r.p95)} p99=${fmt(r.p99)} max=${fmt(r.max)}`,
  );
}

if (unreachable > 0 && unreachable === results.length) {
  console.error(
    `\nAll ${unreachable} scenarios unreachable. Start the services (see docs/deployment.md) or set the *_API_URL env vars.`,
  );
  process.exit(allowUnreachable ? 2 : 1);
}
if (failed > 0) {
  console.error(
    `\n${failed} scenario(s) breached the ${profileName} SLO budget.`,
  );
  process.exit(1);
}
if (unreachable > 0 && !allowUnreachable) {
  console.error(
    `\n${unreachable} scenario(s) unreachable — start the services, set the *_API_URL env vars, or opt out with LOAD_ALLOW_UNREACHABLE=true.`,
  );
  process.exit(1);
}
console.log(
  `\nLoad test OK — ${results.length - unreachable} scenario(s) within ${profileName} SLO budget.`,
);
