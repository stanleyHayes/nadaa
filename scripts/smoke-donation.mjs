import { spawn } from "node:child_process";

const baseURL =
  process.env.DONATION_API_URL?.trim() || "http://127.0.0.1:8100/api/v1";
const rootURL = baseURL.replace("/api/v1", "");
const serviceDir = "services/donation-service";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_donation_officer",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000204",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-donation",
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
    `donation-service did not become healthy: ${lastErr?.message}`,
  );
}

async function ensureService() {
  try {
    const res = await fetch(`${rootURL}/healthz`);
    if (res.ok) return;
  } catch {
    // service is not running; start it below
  }
  console.log("starting donation-service for smoke test...");
  serviceChild = spawn("go", ["run", "."], {
    cwd: serviceDir,
    stdio: "inherit",
    detached: false,
  });
  serviceChild.on("error", (err) => {
    console.error("failed to start donation-service:", err.message);
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
  throw new Error(`donation healthz smoke failed: ${health.status}`);
}
const healthPayload = await health.json();
if (healthPayload.service !== "donation-service") {
  throw new Error("donation healthz returned unexpected service");
}
console.log("donation healthz OK");

const catalog = await fetch(`${baseURL}/aid-catalog`);
if (!catalog.ok) {
  throw new Error(`donation catalog smoke failed: ${catalog.status}`);
}
const catalogPayload = await catalog.json();
if (!Array.isArray(catalogPayload.items) || catalogPayload.items.length < 2) {
  throw new Error("donation catalog smoke expected at least two items");
}
console.log(`donation catalog OK ${catalogPayload.items.length}`);

const requests = await fetch(`${baseURL}/aid-requests`);
if (!requests.ok) {
  throw new Error(`donation aid request list smoke failed: ${requests.status}`);
}
const requestsPayload = await requests.json();
if (
  !Array.isArray(requestsPayload.requests) ||
  requestsPayload.requests.length < 1
) {
  throw new Error(
    "donation aid request list smoke expected at least one request",
  );
}
console.log(`donation aid request list OK ${requestsPayload.requests.length}`);

const missingAuthority = await fetch(`${baseURL}/donors`, {
  headers: { "Content-Type": "application/json" },
});
if (missingAuthority.status !== 401) {
  throw new Error(
    `donation authority gate smoke expected 401, got ${missingAuthority.status}`,
  );
}
console.log("donation authority gate OK 401");

const donor = await fetch(`${baseURL}/donors`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    name: "Smoke Test Donor",
    type: "organization",
    contactName: "Smoke Contact",
    contactEmail: "smoke@example.com",
    contactPhone: "0302000000",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    itemsOffered: ["food_parcel", "water_liter"],
  }),
});
if (!donor.ok) {
  throw new Error(`donation donor create smoke failed: ${donor.status}`);
}
const donorPayload = await donor.json();
if (donorPayload.name !== "Smoke Test Donor" || !donorPayload.id) {
  throw new Error("donation donor create smoke returned invalid payload");
}
console.log(`donation donor create OK ${donorPayload.id}`);

const donorUpdate = await fetch(`${baseURL}/donors/${donorPayload.id}`, {
  method: "PATCH",
  headers: authorityHeaders,
  body: JSON.stringify({ status: "inactive", notes: "Smoke test note" }),
});
if (!donorUpdate.ok) {
  throw new Error(`donation donor update smoke failed: ${donorUpdate.status}`);
}
const donorUpdatePayload = await donorUpdate.json();
if (
  donorUpdatePayload.status !== "inactive" ||
  donorUpdatePayload.notes !== "Smoke test note"
) {
  throw new Error("donation donor update smoke returned invalid payload");
}
console.log("donation donor update OK");

const aidRequest = await fetch(`${baseURL}/aid-requests`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    title: "Smoke Test Aid Request",
    description: "Smoke test aid request description.",
    category: "food",
    itemCode: "food_parcel",
    quantityNeeded: 100,
    unit: "parcels",
    priority: "high",
    locationLabel: "Smoke Test Location",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    beneficiaryCount: 500,
  }),
});
if (!aidRequest.ok) {
  throw new Error(
    `donation aid request create smoke failed: ${aidRequest.status}`,
  );
}
const aidRequestPayload = await aidRequest.json();
if (
  aidRequestPayload.title !== "Smoke Test Aid Request" ||
  !aidRequestPayload.id
) {
  throw new Error("donation aid request create smoke returned invalid payload");
}
console.log(`donation aid request create OK ${aidRequestPayload.id}`);

const aidRequestUpdate = await fetch(
  `${baseURL}/aid-requests/${aidRequestPayload.id}`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({ quantityNeeded: 150 }),
  },
);
if (!aidRequestUpdate.ok) {
  throw new Error(
    `donation aid request update smoke failed: ${aidRequestUpdate.status}`,
  );
}
const aidRequestUpdatePayload = await aidRequestUpdate.json();
if (aidRequestUpdatePayload.quantityNeeded !== 150) {
  throw new Error("donation aid request update smoke returned invalid payload");
}
console.log("donation aid request update OK");

const aidRequestClose = await fetch(
  `${baseURL}/aid-requests/${aidRequestPayload.id}`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({ status: "closed" }),
  },
);
if (!aidRequestClose.ok) {
  throw new Error(
    `donation aid request close smoke failed: ${aidRequestClose.status}`,
  );
}
const aidRequestClosePayload = await aidRequestClose.json();
if (aidRequestClosePayload.status !== "closed") {
  throw new Error("donation aid request close smoke returned invalid payload");
}
console.log("donation aid request close OK");

const pledge = await fetch(
  `${baseURL}/aid-requests/${aidRequestPayload.id}/pledges`,
  {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      donorId: donorPayload.id,
      donorName: "Smoke Test Donor",
      quantityPledged: 50,
      contactEmail: "smoke@example.com",
    }),
  },
);
if (!pledge.ok) {
  throw new Error(`donation pledge create smoke failed: ${pledge.status}`);
}
const pledgePayload = await pledge.json();
if (
  pledgePayload.quantityPledged !== 50 ||
  pledgePayload.status !== "pledged" ||
  !pledgePayload.id
) {
  throw new Error("donation pledge create smoke returned invalid payload");
}
console.log(`donation pledge create OK ${pledgePayload.id}`);

const pledgeUpdate = await fetch(`${baseURL}/pledges/${pledgePayload.id}`, {
  method: "PATCH",
  headers: authorityHeaders,
  body: JSON.stringify({
    status: "delivered",
    quantityDelivered: 50,
    deliveryNote: "Smoke test delivery",
  }),
});
if (!pledgeUpdate.ok) {
  throw new Error(
    `donation pledge update smoke failed: ${pledgeUpdate.status}`,
  );
}
const pledgeUpdatePayload = await pledgeUpdate.json();
if (
  pledgeUpdatePayload.status !== "delivered" ||
  pledgeUpdatePayload.quantityDelivered !== 50
) {
  throw new Error("donation pledge update smoke returned invalid payload");
}
console.log("donation pledge update OK");

const allocate = await fetch(
  `${baseURL}/aid-requests/${aidRequestPayload.id}/allocate`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      pledgeId: pledgePayload.id,
      quantity: 50,
    }),
  },
);
if (!allocate.ok) {
  throw new Error(`donation pledge allocate smoke failed: ${allocate.status}`);
}
const allocatePayload = await allocate.json();
if (
  allocatePayload.status !== "delivered" ||
  allocatePayload.quantityDelivered !== 50
) {
  throw new Error("donation pledge allocate smoke returned invalid payload");
}
console.log("donation pledge allocate OK");

console.log("donation smoke passed");
cleanup();
process.exit(0);
