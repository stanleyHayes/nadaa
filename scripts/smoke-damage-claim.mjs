const baseURL =
  process.env.DAMAGE_CLAIM_API_URL?.trim() || "http://127.0.0.1:8098";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_damage_claim_operator",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000204",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-damage-claim",
};

const health = await fetch(`${baseURL}/health`);
if (!health.ok) {
  throw new Error(`damage-claim health smoke failed: ${health.status}`);
}
console.log("damage-claim health OK");

const create = await fetch(`${baseURL}/claims`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    incidentId: "inc_smoke_001",
    reporter: {
      name: "Smoke Reporter",
      phone: "+233000000001",
      email: "smoke@example.com",
    },
    damageType: "flood",
    damageDescription: "Flooded ground floor and furniture",
    estimatedLossAmount: "4500.00",
    location: { lat: 5.56, lng: -0.2, address: "Accra smoke address" },
    damagePhotos: ["https://example.com/photo1.jpg"],
    privacyConsent: true,
  }),
});
if (!create.ok) {
  throw new Error(`damage-claim intake smoke failed: ${create.status}`);
}
const created = await create.json();
if (!created.id || created.verificationStatus !== "pending") {
  throw new Error("damage-claim intake expected pending record");
}
console.log(`damage-claim intake OK ${created.reference}`);

const list = await fetch(`${baseURL}/claims`, { headers: authorityHeaders });
if (!list.ok) {
  throw new Error(`damage-claim authority list smoke failed: ${list.status}`);
}
const listPayload = await list.json();
if (!Array.isArray(listPayload.claims)) {
  throw new Error("damage-claim authority list expected claims array");
}
console.log(`damage-claim authority list OK ${listPayload.claims.length}`);

const verify = await fetch(`${baseURL}/claims/${created.id}/verify`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({ verificationStatus: "verified", notes: "Smoke verified" }),
});
if (!verify.ok) {
  throw new Error(`damage-claim verify smoke failed: ${verify.status}`);
}
const verified = await verify.json();
if (verified.verificationStatus !== "verified") {
  throw new Error("damage-claim verify expected verified status");
}
console.log("damage-claim verify OK");

const csvExport = await fetch(`${baseURL}/claims/${created.id}/export?format=csv`, {
  headers: authorityHeaders,
});
if (!csvExport.ok) {
  throw new Error(`damage-claim CSV export smoke failed: ${csvExport.status}`);
}
const csvText = await csvExport.text();
if (!csvText.includes(created.reference)) {
  throw new Error("damage-claim CSV export expected reference");
}
console.log("damage-claim CSV export OK");

const pdfExport = await fetch(`${baseURL}/claims/${created.id}/export?format=pdf`, {
  headers: authorityHeaders,
});
if (!pdfExport.ok) {
  throw new Error(`damage-claim PDF export smoke failed: ${pdfExport.status}`);
}
const pdfBuffer = await pdfExport.arrayBuffer();
const pdfHeader = new TextDecoder().decode(pdfBuffer.slice(0, 8));
if (!pdfHeader.startsWith("%PDF-1.4")) {
  throw new Error("damage-claim PDF export expected PDF-1.4 header");
}
console.log("damage-claim PDF export OK");

const close = await fetch(`${baseURL}/claims/${created.id}/close`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({ reason: "Smoke closure" }),
});
if (!close.ok) {
  throw new Error(`damage-claim close smoke failed: ${close.status}`);
}
const closed = await close.json();
if (closed.status !== "closed") {
  throw new Error("damage-claim close expected closed status");
}
console.log("damage-claim close OK");

const invalidVerify = await fetch(`${baseURL}/claims/${created.id}/verify`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({ verificationStatus: "rejected", notes: "Invalid after close" }),
});
if (invalidVerify.status !== 400) {
  throw new Error(`damage-claim invalid verify smoke expected 400 got ${invalidVerify.status}`);
}
console.log("damage-claim invalid verify OK 400");
