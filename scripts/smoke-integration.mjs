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

const importJob = await fetch(
  `${baseURL}/integrations/weather-hydrology/import-jobs`,
  {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      metric: "rainfall_mm",
      requestedBy: "smoke-test",
      correlationId: "corr_import_smoke_001",
    }),
  },
);
if (importJob.status !== 202) {
  throw new Error(
    `integration import smoke expected 202, got ${importJob.status}`,
  );
}
const importPayload = await importJob.json();
if (
  importPayload.status !== "succeeded" ||
  importPayload.trigger !== "manual" ||
  importPayload.importedCount < 1
) {
  throw new Error(
    "integration import smoke did not store fixture observations",
  );
}
console.log(`integration import OK ${importPayload.importedCount}`);

const importedObservations = await fetch(
  `${baseURL}/integrations/weather-hydrology/observations?metric=rainfall_mm`,
);
if (!importedObservations.ok) {
  throw new Error(
    `integration imported observations smoke failed: ${importedObservations.status} ${importedObservations.statusText}`,
  );
}
const importedPayload = await importedObservations.json();
if (
  !Array.isArray(importedPayload.observations) ||
  importedPayload.observations.length < importPayload.importedCount ||
  !importedPayload.observations.every(
    (observation) =>
      observation.importJobId &&
      observation.storageTarget === "weather_observations" &&
      observation.rainfallMm !== undefined,
  )
) {
  throw new Error(
    "integration imported observations smoke returned invalid records",
  );
}
console.log(
  `integration imported observations OK ${importedPayload.observations.length}`,
);

const failedImport = await fetch(
  `${baseURL}/integrations/weather-hydrology/import-jobs`,
  {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      metric: "water_level_m",
      simulateFailure: true,
      failureMessage: "smoke retry check",
      correlationId: "corr_import_smoke_retry",
    }),
  },
);
if (failedImport.status !== 202) {
  throw new Error(
    `integration failed import smoke expected 202, got ${failedImport.status}`,
  );
}
const failedImportPayload = await failedImport.json();
if (
  failedImportPayload.status !== "failed" ||
  !failedImportPayload.retryable ||
  !failedImportPayload.nextRetryAt
) {
  throw new Error(
    "integration failed import smoke did not log retryable failure",
  );
}

const retryImport = await fetch(
  `${baseURL}/integrations/weather-hydrology/import-jobs/${failedImportPayload.id}/retry`,
  { method: "POST" },
);
if (retryImport.status !== 202) {
  throw new Error(
    `integration import retry smoke expected 202, got ${retryImport.status}`,
  );
}
const retryPayload = await retryImport.json();
if (
  retryPayload.status !== "succeeded" ||
  retryPayload.trigger !== "retry" ||
  retryPayload.attempts !== 2 ||
  retryPayload.importedCount < 1
) {
  throw new Error("integration import retry smoke did not succeed");
}
console.log("integration import retry OK");

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
