const baseURL =
  process.env.FORECAST_API_URL?.trim() || "http://127.0.0.1:8094/api/v1";
const healthURL = `${baseURL.replace(/\/api\/v1$/, "")}/healthz`;

const health = await fetch(healthURL);
if (!health.ok) {
  throw new Error(`ml-service health smoke failed: ${health.status}`);
}
console.log("ml-service health OK");

const forecastsResponse = await fetch(`${baseURL}/forecasts`);
if (!forecastsResponse.ok) {
  throw new Error(`forecasts smoke failed: ${forecastsResponse.status}`);
}
const forecastsPayload = await forecastsResponse.json();
if (
  !Array.isArray(forecastsPayload.forecasts) ||
  forecastsPayload.forecasts.length === 0
) {
  throw new Error("forecasts smoke expected a non-empty forecasts array");
}
const topForecast = forecastsPayload.forecasts[0];
if (
  typeof topForecast.predictedIncidentCount !== "number" ||
  !topForecast.riskLevel ||
  !topForecast.confidence ||
  !Array.isArray(topForecast.factors) ||
  topForecast.factors.length === 0
) {
  throw new Error("forecast is missing explainability fields");
}
// Deterministic ordering: forecasts are sorted by predicted demand descending.
for (let i = 1; i < forecastsPayload.forecasts.length; i += 1) {
  if (
    forecastsPayload.forecasts[i - 1].predictedIncidentCount <
    forecastsPayload.forecasts[i].predictedIncidentCount
  ) {
    throw new Error("forecasts are not sorted by predicted demand");
  }
}
console.log(
  `forecasts OK ${forecastsPayload.forecasts.length} district(s), top ${topForecast.district} ${topForecast.predictedIncidentCount} incident(s)`,
);

const regionResponse = await fetch(`${baseURL}/forecasts/Greater%20Accra`);
if (!regionResponse.ok) {
  throw new Error(`forecast-by-region smoke failed: ${regionResponse.status}`);
}
const regionPayload = await regionResponse.json();
if (
  !regionPayload.forecast ||
  regionPayload.forecast.region !== "Greater Accra"
) {
  throw new Error("forecast-by-region returned an unexpected region");
}
console.log("forecast-by-region OK");

const missingRegion = await fetch(`${baseURL}/forecasts/Nowhere`);
if (missingRegion.status !== 404) {
  throw new Error(
    `forecast-by-region unknown expected 404 got ${missingRegion.status}`,
  );
}
console.log("forecast-by-region unknown OK 404");

const stagingResponse = await fetch(
  `${baseURL}/staging-suggestions?agencyType=ambulance`,
);
if (!stagingResponse.ok) {
  throw new Error(`staging smoke failed: ${stagingResponse.status}`);
}
const stagingPayload = await stagingResponse.json();
if (
  !Array.isArray(stagingPayload.suggestions) ||
  stagingPayload.suggestions.length === 0
) {
  throw new Error("staging smoke expected a non-empty suggestions array");
}
for (const suggestion of stagingPayload.suggestions) {
  if (suggestion.agencyType !== "ambulance") {
    throw new Error("staging agency filter leaked a non-ambulance suggestion");
  }
  if (
    suggestion.recommendedUnits < 1 ||
    suggestion.radiusMeters <= 0 ||
    !Array.isArray(suggestion.operationalConstraints) ||
    suggestion.operationalConstraints.length === 0
  ) {
    throw new Error("staging suggestion is missing operational fields");
  }
}
console.log(
  `staging OK ${stagingPayload.suggestions.length} ambulance base(s)`,
);

const compareResponse = await fetch(`${baseURL}/forecasts/compare`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ historicalWeight: 2, timeWindowHours: 48 }),
});
if (!compareResponse.ok) {
  throw new Error(`compare smoke failed: ${compareResponse.status}`);
}
const comparePayload = await compareResponse.json();
if (
  !Array.isArray(comparePayload.scenarios) ||
  comparePayload.scenarios.length !== 2 ||
  comparePayload.scenarios[0].name !== "Current conditions" ||
  comparePayload.scenarios[1].name !== "Adjusted scenario"
) {
  throw new Error("compare smoke expected baseline and adjusted scenarios");
}
if (
  comparePayload.scenarios[1].summary.totalPredictedIncidents <
  comparePayload.scenarios[0].summary.totalPredictedIncidents
) {
  throw new Error("adjusted scenario demand should be >= baseline");
}
console.log("compare OK baseline vs adjusted");

const invalidCompare = await fetch(`${baseURL}/forecasts/compare`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ riskLevel: "catastrophic" }),
});
if (invalidCompare.status !== 400) {
  throw new Error(
    `compare invalid request expected 400 got ${invalidCompare.status}`,
  );
}
console.log("compare invalid request OK 400");
