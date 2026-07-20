const mlBaseURL =
  process.env.ML_API_URL?.trim() || "http://127.0.0.1:8094/api/v1";
const alertBaseURL =
  process.env.ALERT_API_URL?.trim() || "http://127.0.0.1:8089/api/v1";

// ml-service gates every non-health endpoint when NADAA_INTERNAL_SERVICE_TOKEN
// is configured, so send the shared service token on every ML call.
const serviceTokenHeaders = {
  "X-NADAA-Service-Token":
    process.env.NADAA_INTERNAL_SERVICE_TOKEN || "dev-internal-service-token",
};

const drafterHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_ml_reviewer",
  "X-NADAA-Actor-Role": "dispatcher",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-ml-review-draft",
};

const approverHeaders = {
  ...drafterHeaders,
  "X-NADAA-Actor-ID": "usr_smoke_ml_auditor",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Request-ID": "smoke-ml-review-audit",
};

const predictionResponse = await fetch(`${mlBaseURL}/ml/flood/predictions`, {
  method: "POST",
  headers: { "Content-Type": "application/json", ...serviceTokenHeaders },
  body: JSON.stringify({
    location: { lat: 5.56, lng: -0.2 },
    requestedBy: "smoke-ml-review",
    correlationId: "smoke-ml-review-accra-central",
  }),
});

if (!predictionResponse.ok) {
  throw new Error(
    `ML review prediction smoke failed: ${predictionResponse.status}`,
  );
}

const predictionPayload = await predictionResponse.json();
const prediction = predictionPayload.prediction;
if (
  !prediction.humanReviewRequired ||
  prediction.autoPublishAllowed !== false
) {
  throw new Error("ML review prediction safety flags are invalid");
}

const now = Date.now();
const draftResponse = await fetch(`${alertBaseURL}/alerts`, {
  method: "POST",
  headers: drafterHeaders,
  body: JSON.stringify({
    title: `ML reviewed ${prediction.community} flood alert`,
    hazardType: prediction.hazardType,
    severity: prediction.severity === "severe" ? "severe_warning" : "warning",
    message: `Reviewed ML prediction estimates ${
      Math.round(prediction.probability * 1000) / 10
    }% flood probability for ${prediction.community}.`,
    target: {
      type: "custom",
      ids: [prediction.cellId],
      label: `${prediction.community} prediction cell`,
      geometry: prediction.geometry,
    },
    startsAt: new Date(now - 5 * 60 * 1000).toISOString(),
    expiresAt: new Date(now + 8 * 60 * 60 * 1000).toISOString(),
    recommendedAction:
      "Avoid low-lying roads and follow official NADMO instructions.",
    evacuationRequired: prediction.severity === "severe",
    shelterIds: ["00000000-0000-0000-0000-000000000301"],
    sourcePrediction: {
      predictionId: prediction.id,
      predictionLogId: predictionPayload.log.id,
      modelVersion: prediction.modelVersion,
      inputFeatureSetVersion: prediction.inputFeatureSetVersion,
      probability: prediction.probability,
      severity: prediction.severity,
      confidence: prediction.confidence,
      humanReviewRequired: prediction.humanReviewRequired,
      autoPublishAllowed: prediction.autoPublishAllowed,
      reviewNote: "Smoke reviewer checked explanation factors.",
    },
  }),
});

if (draftResponse.status !== 201) {
  throw new Error(
    `ML review draft smoke expected 201, got ${draftResponse.status}`,
  );
}

const draft = await draftResponse.json();
if (
  draft.status !== "draft" ||
  draft.sourcePrediction?.predictionId !== prediction.id
) {
  throw new Error("ML review draft did not preserve source prediction trace");
}
console.log(`ML review draft OK ${draft.id}`);

const auditResponse = await fetch(`${alertBaseURL}/alerts/audit?limit=10`, {
  headers: approverHeaders,
});
if (!auditResponse.ok) {
  throw new Error(`ML review audit smoke failed: ${auditResponse.status}`);
}

const auditPayload = await auditResponse.json();
const creationLog = auditPayload.logs?.find(
  (log) =>
    log.targetId === draft.id &&
    log.action === "alert.created" &&
    log.after?.sourcePrediction?.predictionId === prediction.id,
);
if (!creationLog) {
  throw new Error("ML review audit smoke did not find source prediction trace");
}
console.log(`ML review audit OK ${creationLog.id}`);
