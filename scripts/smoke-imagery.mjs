const baseURL =
  process.env.IMAGERY_API_URL?.trim() || "http://127.0.0.1:8099/api/v1";

const authorityHeaders = {
  "X-NADAA-Actor-ID": "usr_smoke_imagery_operator",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000205",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-imagery",
};

const minPNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
  "base64",
);

const health = await fetch(`${baseURL.replace("/api/v1", "")}/healthz`);
if (!health.ok) {
  throw new Error(`imagery health check failed: ${health.status}`);
}
console.log("imagery health OK");

const unauthorizedList = await fetch(`${baseURL}/imagery`);
if (unauthorizedList.status !== 401) {
  throw new Error(
    `imagery authority gate expected 401, got ${unauthorizedList.status}`,
  );
}
console.log("imagery authority gate OK 401");

const form = new FormData();
form.append(
  "file",
  new Blob([minPNG], { type: "image/png" }),
  "smoke-imagery.png",
);
form.append("source", "drone");
form.append("captureTime", new Date().toISOString());
form.append(
  "geometry",
  JSON.stringify({
    type: "Polygon",
    coordinates: [
      [
        [-0.22, 5.56],
        [-0.19, 5.56],
        [-0.19, 5.59],
        [-0.22, 5.59],
        [-0.22, 5.56],
      ],
    ],
  }),
);
form.append("coverageAreaKm2", "4.5");
form.append("resolutionMeters", "0.25");
form.append("license", "CC-BY-4.0");
form.append("relatedIncidentId", "incident_smoke_140");

const create = await fetch(`${baseURL}/imagery`, {
  method: "POST",
  headers: authorityHeaders,
  body: form,
});
if (!create.ok) {
  const text = await create.text().catch(() => "");
  throw new Error(`imagery create smoke failed: ${create.status} ${text}`);
}
const created = await create.json();
if (
  !created.id ||
  created.source !== "drone" ||
  created.status !== "active" ||
  !created.storagePath
) {
  throw new Error("imagery create smoke returned invalid payload");
}
console.log(`imagery create OK ${created.id}`);

const list = await fetch(`${baseURL}/imagery?source=drone`, {
  headers: authorityHeaders,
});
if (!list.ok) {
  throw new Error(`imagery list smoke failed: ${list.status}`);
}
const listPayload = await list.json();
if (
  !Array.isArray(listPayload.imagery) ||
  !listPayload.imagery.some((record) => record.id === created.id)
) {
  throw new Error("imagery list smoke did not include created record");
}
console.log(`imagery list OK ${listPayload.imagery.length}`);

const geojson = await fetch(`${baseURL}/imagery/geojson`);
if (!geojson.ok) {
  throw new Error(`imagery geojson smoke failed: ${geojson.status}`);
}
const geojsonPayload = await geojson.json();
if (
  geojsonPayload.type !== "FeatureCollection" ||
  !Array.isArray(geojsonPayload.features) ||
  !geojsonPayload.features.some(
    (feature) => feature.properties?.id === created.id,
  )
) {
  throw new Error("imagery geojson smoke returned invalid payload");
}
console.log(`imagery geojson OK ${geojsonPayload.features.length}`);

const lifecycle = await fetch(`${baseURL}/imagery/lifecycle/run`, {
  method: "POST",
  headers: authorityHeaders,
});
if (!lifecycle.ok) {
  throw new Error(`imagery lifecycle smoke failed: ${lifecycle.status}`);
}
const lifecyclePayload = await lifecycle.json();
if (typeof lifecyclePayload.expiredCount !== "number") {
  throw new Error("imagery lifecycle smoke returned invalid payload");
}
console.log(`imagery lifecycle OK expired=${lifecyclePayload.expiredCount}`);

const detail = await fetch(`${baseURL}/imagery/${created.id}`, {
  headers: authorityHeaders,
});
if (!detail.ok) {
  throw new Error(`imagery detail smoke failed: ${detail.status}`);
}
const detailPayload = await detail.json();
if (detailPayload.id !== created.id) {
  throw new Error("imagery detail smoke returned invalid payload");
}
console.log(`imagery detail OK ${detailPayload.status}`);

const download = await fetch(`${baseURL}/imagery/${created.id}/download`, {
  headers: authorityHeaders,
});
if (!download.ok) {
  throw new Error(`imagery download smoke failed: ${download.status}`);
}
if (download.headers.get("content-type") !== "image/png") {
  throw new Error(
    `imagery download smoke expected image/png, got ${download.headers.get("content-type")}`,
  );
}
console.log("imagery download OK");

const remove = await fetch(`${baseURL}/imagery/${created.id}`, {
  method: "DELETE",
  headers: authorityHeaders,
});
if (!remove.ok) {
  throw new Error(`imagery delete smoke failed: ${remove.status}`);
}
console.log("imagery delete OK");

console.log("imagery smoke tests passed");
