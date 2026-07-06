const baseURL =
  process.env.ML_API_URL?.trim() || "http://127.0.0.1:8094/api/v1";

const predictionResponse = await fetch(`${baseURL}/ml/flood/predictions`, {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    location: { lat: 5.56, lng: -0.2 },
    requestedBy: "smoke-ml",
    correlationId: "smoke-ml-accra-central",
  }),
});

if (!predictionResponse.ok) {
  throw new Error(
    `ML prediction smoke failed: ${predictionResponse.status} ${predictionResponse.statusText}`,
  );
}

const predictionPayload = await predictionResponse.json();
const prediction = predictionPayload.prediction;
if (prediction.modelVersion !== "flood-logistic-baseline-0.1.0") {
  throw new Error(
    `expected baseline model version, got ${prediction.modelVersion}`,
  );
}
if (prediction.hazardType !== "flood" || prediction.severity !== "severe") {
  throw new Error(
    `expected severe flood prediction, got ${prediction.severity}`,
  );
}
if (
  !prediction.humanReviewRequired ||
  prediction.autoPublishAllowed !== false
) {
  throw new Error(
    "ML prediction must require human review and disallow auto-publish",
  );
}
if (
  !predictionPayload.log?.modelVersion ||
  predictionPayload.log.storageTarget !== "ml_predictions"
) {
  throw new Error("ML prediction response must include a prediction log");
}

console.log(
  `ML prediction OK ${prediction.modelVersion} ${prediction.severity}`,
);

const logsResponse = await fetch(`${baseURL}/ml/prediction-logs`);
if (!logsResponse.ok) {
  throw new Error(
    `ML log smoke failed: ${logsResponse.status} ${logsResponse.statusText}`,
  );
}

const logsPayload = await logsResponse.json();
if (!Array.isArray(logsPayload.logs) || logsPayload.logs.length === 0) {
  throw new Error("ML log smoke expected at least one logged prediction");
}

console.log(`ML logs OK ${logsPayload.logs.length}`);
