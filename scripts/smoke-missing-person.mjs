import { spawn } from "node:child_process";

const baseURL =
  process.env.MISSING_PERSON_API_URL?.trim() || "http://127.0.0.1:8101/api/v1";
const rootURL = baseURL.replace("/api/v1", "");
const serviceDir = "services/missing-person-service";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_missing_officer",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000204",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-missing-person",
};

let serviceChild = null;

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function waitForHealth(maxMs = 30000) {
  const deadline = Date.now() + maxMs;
  let lastErr;
  while (Date.now() < deadline) {
    try {
      const res = await fetch(`${rootURL}/healthz`);
      if (res.ok) return;
    } catch (err) {
      lastErr = err;
    }
    await sleep(500);
  }
  throw new Error(
    `missing-person-service did not become healthy: ${lastErr?.message}`,
  );
}

async function ensureService() {
  try {
    const res = await fetch(`${rootURL}/healthz`);
    if (res.ok) return;
  } catch {
    // service is not running; start it below
  }
  console.log("starting missing-person-service for smoke test...");
  serviceChild = spawn("go", ["run", "./cmd/server"], {
    cwd: serviceDir,
    stdio: "inherit",
    detached: false,
  });
  serviceChild.on("error", (err) => {
    console.error("failed to start missing-person-service:", err.message);
  });
  await waitForHealth();
}

function cleanup() {
  if (serviceChild && !serviceChild.killed) {
    serviceChild.kill("SIGTERM");
  }
}

process.on("exit", cleanup);
process.on("SIGINT", () => {
  cleanup();
  process.exit(130);
});
process.on("SIGTERM", cleanup);

await ensureService();

const health = await fetch(`${rootURL}/healthz`);
if (!health.ok) {
  throw new Error(`missing-person healthz smoke failed: ${health.status}`);
}
const healthPayload = await health.json();
if (healthPayload.service !== "missing-person-service") {
  throw new Error("missing-person healthz returned unexpected service");
}
console.log("missing-person healthz OK");

const publicList = await fetch(`${baseURL}/missing-persons`);
if (!publicList.ok) {
  throw new Error(
    `missing-person public list smoke failed: ${publicList.status}`,
  );
}
const publicPayload = await publicList.json();
if (
  !Array.isArray(publicPayload.records) ||
  publicPayload.records.length !== 1
) {
  throw new Error("missing-person public list expected only approved fixture");
}
console.log(`missing-person public list OK ${publicPayload.records.length}`);

const authorityGate = await fetch(`${baseURL}/authority/missing-persons`);
if (authorityGate.status !== 401) {
  throw new Error(
    `missing-person authority gate expected 401, got ${authorityGate.status}`,
  );
}
console.log("missing-person authority gate OK 401");

const create = await fetch(`${baseURL}/missing-persons`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    personName: "Smoke Missing Person",
    age: 15,
    gender: "unknown",
    description: "Last seen near the smoke-test shelter desk.",
    photoUrl: "https://example.test/smoke-missing.jpg",
    lastSeenAt: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
    lastSeenLocation: {
      label: "Osu Community Hall",
      region: "Greater Accra",
      district: "Korle Klottey",
      lat: 5.55,
      lng: -0.18,
    },
    relatedIncidentId: "inc_accra_flood_0241",
    reporter: {
      name: "Smoke Reporter",
      phone: "+233200000333",
      email: "smoke.reporter@example.com",
      relationship: "guardian",
      consentToContact: true,
      consentToPublicShare: true,
    },
  }),
});
if (!create.ok) {
  throw new Error(`missing-person create smoke failed: ${create.status}`);
}
const created = await create.json();
if (
  created.status !== "pending_review" ||
  created.publicVisibility !== "private" ||
  !created.reporter?.phone
) {
  throw new Error(
    "missing-person create smoke returned invalid private record",
  );
}
console.log(`missing-person create OK ${created.id}`);

const hidden = await fetch(`${baseURL}/missing-persons/${created.id}`);
if (hidden.status !== 404) {
  throw new Error(
    `missing-person unapproved public lookup expected 404, got ${hidden.status}`,
  );
}
console.log("missing-person public privacy OK 404");

const review = await fetch(
  `${baseURL}/authority/missing-persons/${created.id}/review`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({
      decision: "approve_public",
      publicSummary: "Smoke case approved for safe public search.",
      reviewNotes: "Reporter consent verified during smoke test.",
    }),
  },
);
if (!review.ok) {
  throw new Error(`missing-person review smoke failed: ${review.status}`);
}
const reviewed = await review.json();
if (
  reviewed.reviewStatus !== "approved" ||
  reviewed.publicVisibility !== "public" ||
  reviewed.status !== "active"
) {
  throw new Error("missing-person review smoke returned invalid record");
}
console.log("missing-person review OK");

const close = await fetch(
  `${baseURL}/authority/missing-persons/${created.id}/close`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({
      closureType: "reunited",
      closureNotes: "Smoke family reunification completed.",
      reunitedWithFamily: true,
    }),
  },
);
if (!close.ok) {
  throw new Error(`missing-person close smoke failed: ${close.status}`);
}
const closed = await close.json();
if (closed.status !== "reunited" || closed.publicVisibility !== "private") {
  throw new Error("missing-person close smoke returned invalid record");
}
console.log("missing-person close OK");

const audit = await fetch(
  `${baseURL}/authority/missing-persons/${created.id}/audit`,
  { headers: authorityHeaders },
);
if (!audit.ok) {
  throw new Error(`missing-person audit smoke failed: ${audit.status}`);
}
const auditPayload = await audit.json();
if (!Array.isArray(auditPayload.entries) || auditPayload.entries.length < 3) {
  throw new Error("missing-person audit smoke expected workflow audit entries");
}
console.log(`missing-person audit OK ${auditPayload.entries.length}`);
cleanup();
