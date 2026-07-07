const baseURL =
  process.env.RELIEF_API_URL?.trim() || "http://127.0.0.1:8093/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_relief_officer",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000204",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-relief",
};

const list = await fetch(`${baseURL}/relief-points?status=open`);
if (!list.ok) {
  throw new Error(`relief point list smoke failed: ${list.status}`);
}
const listPayload = await list.json();
if (!Array.isArray(listPayload.reliefPoints) || listPayload.reliefPoints.length < 2) {
  throw new Error("relief point list smoke expected at least two points");
}
console.log(`relief point list OK ${listPayload.reliefPoints.length}`);

const nearby = await fetch(
  `${baseURL}/relief-points/nearby?lat=5.560&lng=-0.200`,
);
if (!nearby.ok) {
  throw new Error(`relief point nearby smoke failed: ${nearby.status}`);
}
const nearbyPayload = await nearby.json();
if (!Array.isArray(nearbyPayload.reliefPoints) || nearbyPayload.reliefPoints.length < 2) {
  throw new Error("relief point nearby smoke expected at least two points");
}
console.log(`relief point nearby OK ${nearbyPayload.reliefPoints.length}`);

const missingAuthority = await fetch(`${baseURL}/relief-points`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ name: "Unauthorized Relief Point", type: "food" }),
});
if (missingAuthority.status !== 401) {
  throw new Error(
    `relief point authority smoke expected 401, got ${missingAuthority.status}`,
  );
}
console.log("relief point authority gate OK 401");

const create = await fetch(`${baseURL}/relief-points`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    name: "Smoke Test Relief Point",
    type: "food",
    location: { lat: 5.55, lng: -0.19 },
    stockCategories: [
      { category: "rice_kg", quantity: 100, unit: "kg" },
    ],
  }),
});
if (!create.ok) {
  throw new Error(`relief point create smoke failed: ${create.status}`);
}
const createPayload = await create.json();
if (
  createPayload.name !== "Smoke Test Relief Point" ||
  createPayload.createdBy !== "usr_smoke_relief_officer"
) {
  throw new Error("relief point create smoke returned invalid payload");
}
console.log(`relief point create OK ${createPayload.id}`);

const invalidGeometry = await fetch(`${baseURL}/relief-points`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    name: "Invalid Point",
    type: "food",
    location: { lat: 95, lng: -0.19 },
  }),
});
if (invalidGeometry.status !== 400) {
  throw new Error(
    `relief point invalid geometry smoke expected 400, got ${invalidGeometry.status}`,
  );
}
console.log("relief point invalid geometry OK 400");

const update = await fetch(
  `${baseURL}/relief-points/${createPayload.id}`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({
      status: "limited",
      stockCategories: [
        { category: "rice_kg", quantity: 50, unit: "kg" },
        { category: "water_bottles", quantity: 100, unit: "bottles" },
      ],
    }),
  },
);
if (!update.ok) {
  throw new Error(`relief point update smoke failed: ${update.status}`);
}
const updatePayload = await update.json();
if (updatePayload.status !== "limited") {
  throw new Error("relief point update smoke returned invalid payload");
}
console.log("relief point update OK");

const history = await fetch(
  `${baseURL}/relief-points/${createPayload.id}/stock-history`,
);
if (!history.ok) {
  throw new Error(`relief point stock history smoke failed: ${history.status}`);
}
const historyPayload = await history.json();
if (!Array.isArray(historyPayload.history) || historyPayload.history.length !== 1) {
  throw new Error("relief point stock history smoke expected one entry");
}
console.log("relief point stock history OK");
