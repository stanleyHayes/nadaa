// Expects an externally started notification-service (see
// scripts/dev-citizen-backends.sh) running with NADAA_ENV=development,
// NADAA_AUTH_ALLOW_MOCK_ACTORS=true and
// NADAA_AUTH_TOKEN_SECRET=dev-secret-change-me so the mock X-NADAA-Actor-*
// headers below satisfy the authority gate on the deliver endpoint.
const baseURL =
  process.env.NOTIFICATION_API_URL?.trim() || "http://127.0.0.1:8090/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_voice_officer",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-voice-alerts",
};

async function requestJSON(path, options, expectedStatus) {
  const response = await fetch(`${baseURL}${path}`, options);
  if (response.status !== expectedStatus) {
    throw new Error(
      `${path} expected ${expectedStatus}, got ${response.status} ${response.statusText}`,
    );
  }
  return response.json();
}

const created = await requestJSON(
  "/notifications/voice-alerts",
  {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      alertId: "alert_feed_current_flood",
      languages: ["en", "tw", "ha"],
      workflowRequestedBy: "voice-smoke",
      source: "tts_sandbox",
    }),
  },
  201,
);

if (
  !created.asset?.id ||
  created.asset.reviewStatus !== "pending_review" ||
  created.asset.variants?.length !== 3
) {
  throw new Error("voice alert generation did not create pending variants");
}
if (
  !created.asset.variants.every(
    (variant) =>
      variant.audioUrl &&
      variant.durationSeconds > 0 &&
      variant.messageText?.includes("112"),
  )
) {
  throw new Error(
    "voice alert variants missing audio metadata or 112 guidance",
  );
}
console.log(`voice alert generated OK ${created.asset.id}`);

const reviewed = await requestJSON(
  `/notifications/voice-alerts/${created.asset.id}/review`,
  {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      action: "approve",
      reviewer: "voice-smoke-reviewer",
      note: "Smoke script approved low-literacy voice variants.",
    }),
  },
  200,
);

if (
  reviewed.asset?.status !== "approved" ||
  reviewed.asset.reviewStatus !== "approved"
) {
  throw new Error("voice alert review did not approve the asset");
}
console.log("voice alert review OK approved");

const delivered = await requestJSON(
  `/notifications/voice-alerts/${created.asset.id}/deliver`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      recipients: [
        { phone: "+233200000300", language: "en" },
        {
          recipientId: "usr_voice_smoke_002",
          phone: "+233200000301",
          language: "tw",
        },
      ],
    }),
  },
  202,
);

if (
  delivered.attempts?.length !== 2 ||
  !delivered.attempts.every(
    (attempt) =>
      attempt.channel === "voice" &&
      attempt.provider === "mock_voice" &&
      attempt.status === "delivered" &&
      attempt.voiceAssetId === created.asset.id &&
      attempt.audioUrl,
  )
) {
  throw new Error("voice alert delivery did not log delivered attempts");
}
console.log("voice alert delivery OK mock_voice");

const logs = await fetch(
  `${baseURL}/notifications/delivery-logs?channel=voice&alertId=${created.asset.alertId}`,
);
if (!logs.ok) {
  throw new Error(
    `voice delivery logs failed: ${logs.status} ${logs.statusText}`,
  );
}
const logsPayload = await logs.json();
if (
  !Array.isArray(logsPayload.logs) ||
  logsPayload.logs.length < 2 ||
  !logsPayload.logs.some((log) => log.voiceAssetId === created.asset.id)
) {
  throw new Error("voice delivery logs missing expected voice asset attempts");
}
console.log(`voice delivery logs OK ${logsPayload.logs.length}`);
