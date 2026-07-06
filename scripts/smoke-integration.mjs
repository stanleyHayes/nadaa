const baseURL =
  process.env.INTEGRATION_API_URL?.trim() || "http://127.0.0.1:8088/api/v1";

const contracts = await fetch(
  `${baseURL}/integrations/contracts?domain=weather&direction=inbound`,
);
if (!contracts.ok) {
  throw new Error(
    `integration contracts smoke failed: ${contracts.status} ${contracts.statusText}`,
  );
}

const contractPayload = await contracts.json();
if (
  !Array.isArray(contractPayload.contracts) ||
  contractPayload.contracts.length !== 1
) {
  throw new Error("integration contracts smoke expected one weather contract");
}
const weatherContract = contractPayload.contracts[0];
if (
  weatherContract.partner !== "Ghana Meteorological Agency" ||
  weatherContract.authentication?.mode !== "api_key" ||
  !Array.isArray(weatherContract.payloads) ||
  weatherContract.payloads.length === 0
) {
  throw new Error("integration contracts smoke returned incomplete contract");
}
console.log("integration contracts OK");

const observations = await fetch(
  `${baseURL}/integrations/mock/weather-hydrology/observations?metric=rainfall_mm`,
);
if (!observations.ok) {
  throw new Error(
    `integration observations smoke failed: ${observations.status} ${observations.statusText}`,
  );
}

const observationPayload = await observations.json();
if (
  !Array.isArray(observationPayload.observations) ||
  observationPayload.observations.length === 0 ||
  !observationPayload.observations.every(
    (observation) =>
      observation.metric === "rainfall_mm" &&
      observation.generatedBy === "mock_adapter",
  )
) {
  throw new Error("integration observations smoke returned invalid fixtures");
}
console.log(
  `integration observations OK ${observationPayload.observations.length}`,
);

const sync = await fetch(`${baseURL}/integrations/mock/sync-events`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    type: "incident",
    sourceId: "inc_smoke_001",
    reference: "INC-SMOKE-001",
    hazardType: "flood",
    status: "verified",
    severity: "high",
    summary: "Smoke-test incident sync payload",
    location: { lat: 5.6037, lng: -0.187 },
    targetAgencyIds: ["00000000-0000-0000-0000-000000000101"],
    correlationId: "corr_smoke_001",
  }),
});
if (sync.status !== 202) {
  throw new Error(`integration sync smoke expected 202, got ${sync.status}`);
}
const syncPayload = await sync.json();
if (
  syncPayload.status !== "accepted" ||
  syncPayload.adapterId !== "mock-incident-sync-adapter"
) {
  throw new Error("integration sync smoke returned invalid acceptance payload");
}
console.log("integration sync OK accepted");

const invalid = await fetch(`${baseURL}/integrations/contracts?domain=aliens`);
if (invalid.status !== 400) {
  throw new Error(
    `invalid integration domain smoke expected 400, got ${invalid.status}`,
  );
}
console.log("invalid-domain integration OK 400");
