const baseURL =
  process.env.SHELTER_API_URL?.trim() || "http://127.0.0.1:8093/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_aid_officer",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000204",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-aid",
};

const publicList = await fetch(`${baseURL}/aid-requests`);
if (!publicList.ok) {
  throw new Error(`aid request public list smoke failed: ${publicList.status}`);
}
const publicPayload = await publicList.json();
if (
  !Array.isArray(publicPayload.aidRequests) ||
  publicPayload.aidRequests.length < 1
) {
  throw new Error(
    "aid request public list expected at least one approved need",
  );
}
console.log(`aid request public list OK ${publicPayload.aidRequests.length}`);

const unauthorizedCreate = await fetch(`${baseURL}/aid-requests`, {
  body: JSON.stringify({ title: "Unauthorized aid need" }),
  headers: { "Content-Type": "application/json" },
  method: "POST",
});
if (unauthorizedCreate.status !== 401) {
  throw new Error(
    `aid request authority gate expected 401, got ${unauthorizedCreate.status}`,
  );
}
console.log("aid request authority gate OK 401");

const create = await fetch(`${baseURL}/aid-requests`, {
  body: JSON.stringify({
    title: "Smoke Test Medical Kits",
    category: "medical",
    priority: "urgent",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    location: { lat: 5.56, lng: -0.2 },
    receivingOrganization: "Smoke Test Relief Desk",
    contact: "112",
    quantityNeeded: 40,
    quantityUnit: "kits",
    description: "Medical kits for smoke-tested donation coordination.",
    neededBy: new Date(Date.now() + 48 * 60 * 60 * 1000).toISOString(),
    visibility: "public",
  }),
  headers: authorityHeaders,
  method: "POST",
});
if (!create.ok) {
  throw new Error(`aid request create smoke failed: ${create.status}`);
}
const created = await create.json();
if (
  created.status !== "pending_review" ||
  created.createdBy !== authorityHeaders["X-NADAA-Actor-ID"]
) {
  throw new Error("aid request create returned invalid payload");
}
console.log(`aid request create OK ${created.id}`);

const review = await fetch(`${baseURL}/aid-requests/${created.id}/review`, {
  body: JSON.stringify({
    status: "approved",
    approvalNotes: "Smoke test approval after receiving organization check.",
    antiFraudNotes: "Smoke test anti-fraud review note.",
  }),
  headers: authorityHeaders,
  method: "PATCH",
});
if (!review.ok) {
  throw new Error(`aid request review smoke failed: ${review.status}`);
}
const reviewed = await review.json();
if (
  reviewed.status !== "approved" ||
  reviewed.approvedBy !== authorityHeaders["X-NADAA-Actor-ID"]
) {
  throw new Error("aid request review returned invalid payload");
}
console.log("aid request review OK");

const pledge = await fetch(`${baseURL}/aid-requests/${created.id}/pledges`, {
  body: JSON.stringify({
    donorName: "Smoke Test Donor",
    donorType: "business",
    contact: "donor@example.org",
    quantity: 10,
    unit: "kits",
    note: "Smoke pledge ready for pickup.",
  }),
  headers: { "Content-Type": "application/json" },
  method: "POST",
});
if (!pledge.ok) {
  throw new Error(`aid pledge create smoke failed: ${pledge.status}`);
}
const pledgePayload = await pledge.json();
if (
  pledgePayload.status !== "pledged" ||
  pledgePayload.reviewStatus !== "pending_review"
) {
  throw new Error("aid pledge create returned invalid payload");
}
console.log(`aid pledge create OK ${pledgePayload.id}`);

const pledgeReview = await fetch(
  `${baseURL}/aid-requests/${created.id}/pledges/${pledgePayload.id}/review`,
  {
    body: JSON.stringify({
      status: "accepted",
      reviewStatus: "cleared",
      fraudReviewNotes: "Smoke donor contact checked.",
    }),
    headers: authorityHeaders,
    method: "PATCH",
  },
);
if (!pledgeReview.ok) {
  throw new Error(`aid pledge review smoke failed: ${pledgeReview.status}`);
}
console.log("aid pledge review OK");

const exportResponse = await fetch(`${baseURL}/aid-requests/report.csv`, {
  headers: authorityHeaders,
});
if (!exportResponse.ok) {
  throw new Error(`aid export smoke failed: ${exportResponse.status}`);
}
const csv = await exportResponse.text();
if (
  !csv.includes("Smoke Test Medical Kits") ||
  !csv.includes("quantityPledged")
) {
  throw new Error("aid export smoke missing expected CSV content");
}
console.log("aid export OK");
