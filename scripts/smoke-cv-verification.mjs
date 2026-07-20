const baseURL = process.env.CV_API_URL?.trim() || "http://127.0.0.1:8094";

// ml-service gates every non-health endpoint when NADAA_INTERNAL_SERVICE_TOKEN
// is configured, so send the shared service token on every CV call.
const serviceTokenHeaders = {
  "X-NADAA-Service-Token":
    process.env.NADAA_INTERNAL_SERVICE_TOKEN || "dev-internal-service-token",
};

async function assertOk(response, label) {
  if (!response.ok) {
    const body = await response.text();
    throw new Error(`${label} failed: ${response.status} ${body}`);
  }
}

// Health check
const health = await fetch(`${baseURL}/healthz`);
await assertOk(health, "Health check");
const healthBody = await health.json();
if (healthBody.status !== "ok" || healthBody.service !== "ml-service") {
  throw new Error("Unexpected health response");
}
console.log("[smoke-cv] Health OK");

// Analyze flood image
const floodAnalyze = await fetch(`${baseURL}/api/v1/cv/analyze`, {
  method: "POST",
  headers: { "Content-Type": "application/json", ...serviceTokenHeaders },
  body: JSON.stringify({ imageId: "smoke_flood_001", imageName: "flood-scene.jpg" }),
});
await assertOk(floodAnalyze, "CV flood analyze");
const floodResult = await floodAnalyze.json();
if (!floodResult.result?.labels?.length) {
  throw new Error("Expected CV labels in flood analyze response");
}
if (floodResult.result.labels[0].label !== "flood_evidence") {
  throw new Error(`Expected flood_evidence label, got ${floodResult.result.labels[0].label}`);
}
if (floodResult.result.humanReviewRequired) {
  throw new Error("Expected no human review for high-confidence flood image");
}
if (floodResult.safety?.autoPublishAllowed) {
  throw new Error("Expected autoPublishAllowed=false for safety");
}
console.log("[smoke-cv] Flood analyze OK", floodResult.result.labels.map((l) => `${l.label}:${(l.confidence * 100).toFixed(0)}%`).join(", "));

// Analyze sensitive image
const sensitiveAnalyze = await fetch(`${baseURL}/api/v1/cv/analyze`, {
  method: "POST",
  headers: { "Content-Type": "application/json", ...serviceTokenHeaders },
  body: JSON.stringify({ imageId: "smoke_sensitive_001", imageName: "injured-person.jpg" }),
});
await assertOk(sensitiveAnalyze, "CV sensitive analyze");
const sensitiveResult = await sensitiveAnalyze.json();
if (!sensitiveResult.result.humanReviewRequired) {
  throw new Error("Expected human review required for sensitive image");
}
if (sensitiveResult.result.labels[0].label !== "sensitive") {
  throw new Error(`Expected sensitive label, got ${sensitiveResult.result.labels[0].label}`);
}
console.log("[smoke-cv] Sensitive analyze OK");

// Retrieve cached result
const getResult = await fetch(`${baseURL}/api/v1/cv/results/smoke_flood_001`, {
  headers: serviceTokenHeaders,
});
await assertOk(getResult, "CV get result");
const getBody = await getResult.json();
if (getBody.result.imageId !== "smoke_flood_001") {
  throw new Error("Expected cached result for smoke_flood_001");
}
console.log("[smoke-cv] Get result OK");

// List results
const listResults = await fetch(`${baseURL}/api/v1/cv/results`, {
  headers: serviceTokenHeaders,
});
await assertOk(listResults, "CV list results");
const listBody = await listResults.json();
if (!Array.isArray(listBody.results)) {
  throw new Error("Expected results array in list response");
}
if (listBody.results.length < 2) {
  throw new Error(`Expected at least 2 cached results, got ${listBody.results.length}`);
}
console.log("[smoke-cv] List results OK", listBody.results.length);

// Verify caching: re-analyze same image should return cached ID
const cachedAnalyze = await fetch(`${baseURL}/api/v1/cv/analyze`, {
  method: "POST",
  headers: { "Content-Type": "application/json", ...serviceTokenHeaders },
  body: JSON.stringify({ imageId: "smoke_flood_001", imageName: "flood-scene.jpg" }),
});
await assertOk(cachedAnalyze, "CV cached analyze");
const cachedResult = await cachedAnalyze.json();
if (cachedResult.result.id !== floodResult.result.id) {
  throw new Error("Expected cached result to have same ID");
}
console.log("[smoke-cv] Caching OK");

console.log("[smoke-cv] All CV verification checks passed.");
