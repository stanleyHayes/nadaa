const baseURL =
  process.env.ROAD_CLOSURE_API_URL?.trim() || "http://127.0.0.1:8095/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_road_closure_officer",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000204",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-road-closure",
};

const list = await fetch(`${baseURL}/road-closures?lat=5.570&lng=-0.200`);
if (!list.ok) {
  throw new Error(`road closure list smoke failed: ${list.status}`);
}
const listPayload = await list.json();
if (!Array.isArray(listPayload.closures) || listPayload.closures.length < 2) {
  throw new Error("road closure list smoke expected at least two closures");
}
console.log(`road closure list OK ${listPayload.closures.length}`);

const bbox = await fetch(`${baseURL}/road-closures?bbox=-0.30,5.50,-0.15,5.60`);
if (!bbox.ok) {
  throw new Error(`road closure bbox smoke failed: ${bbox.status}`);
}
const bboxPayload = await bbox.json();
if (!Array.isArray(bboxPayload.closures) || bboxPayload.closures.length < 2) {
  throw new Error("road closure bbox smoke expected closures in bbox");
}
console.log(`road closure bbox OK ${bboxPayload.closures.length}`);

const missingAuthority = await fetch(`${baseURL}/road-closures`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ roadName: "Unauthorized Road" }),
});
if (missingAuthority.status !== 401) {
  throw new Error(
    `road closure authority smoke expected 401, got ${missingAuthority.status}`,
  );
}
console.log("road closure authority gate OK 401");

const create = await fetch(`${baseURL}/road-closures`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    roadName: "Smoke Test Road",
    reason: "Flooding",
    status: "active",
    severity: "high",
    geometry: {
      type: "LineString",
      coordinates: [
        [-0.205, 5.57],
        [-0.19, 5.58],
      ],
    },
    detourNote: "Use smoke test detour.",
  }),
});
if (!create.ok) {
  throw new Error(`road closure create smoke failed: ${create.status}`);
}
const createPayload = await create.json();
if (
  createPayload.closure?.roadName !== "Smoke Test Road" ||
  createPayload.closure?.createdBy !== "usr_smoke_road_closure_officer"
) {
  throw new Error("road closure create smoke returned invalid payload");
}
console.log(`road closure create OK ${createPayload.closure.id}`);

const invalidGeometry = await fetch(`${baseURL}/road-closures`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    roadName: "Bad Geometry Road",
    status: "active",
    geometry: {
      type: "Polygon",
      coordinates: [[-0.205, 5.57]],
    },
  }),
});
if (invalidGeometry.status !== 400) {
  throw new Error(
    `road closure invalid geometry smoke expected 400, got ${invalidGeometry.status}`,
  );
}
console.log("road closure invalid geometry OK 400");

const update = await fetch(
  `${baseURL}/road-closures/${createPayload.closure.id}`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({ status: "lifted", reason: "Water receded" }),
  },
);
if (!update.ok) {
  throw new Error(`road closure update smoke failed: ${update.status}`);
}
const updatePayload = await update.json();
if (
  updatePayload.closure?.status !== "lifted" ||
  updatePayload.closure?.updatedBy !== "usr_smoke_road_closure_officer"
) {
  throw new Error("road closure update smoke returned invalid payload");
}
console.log("road closure update OK");

const adapterImport = await fetch(`${baseURL}/road-closures/imports/adapter`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    source: "ghana-police",
    sourceRef: "smoke-road-closure",
    roadName: "Police Feed Road",
    status: "active",
    reason: "Flooding",
    geometry: "LINESTRING(-0.20 5.56, -0.19 5.57)",
    validFrom: new Date().toISOString(),
    detour: "Use alternate route",
  }),
});
if (!adapterImport.ok) {
  throw new Error(
    `road closure adapter import smoke failed: ${adapterImport.status}`,
  );
}
const adapterImportPayload = await adapterImport.json();
if (
  adapterImportPayload.imported !== 1 ||
  adapterImportPayload.closures[0]?.geometry.type !== "LineString"
) {
  throw new Error("road closure adapter import smoke returned invalid payload");
}
console.log("road closure adapter import OK");

const invalidAdapter = await fetch(`${baseURL}/road-closures/imports/adapter`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    source: "ghana-police",
    roadName: "Bad WKT Road",
    status: "active",
    geometry: "POINT(0 0)",
    validFrom: new Date().toISOString(),
  }),
});
if (invalidAdapter.status !== 400) {
  throw new Error(
    `road closure invalid adapter smoke expected 400, got ${invalidAdapter.status}`,
  );
}
console.log("road closure invalid adapter OK 400");
