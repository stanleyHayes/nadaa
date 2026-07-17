// Expects an externally started notification-service (see
// scripts/dev-citizen-backends.sh) running with NADAA_ENV=development,
// NADAA_AUTH_ALLOW_MOCK_ACTORS=true and
// NADAA_AUTH_TOKEN_SECRET=dev-secret-change-me so the mock X-NADAA-Actor-*
// headers below satisfy the authority gate on delivery endpoints.
const baseURL =
  process.env.NOTIFICATION_API_URL?.trim() || "http://127.0.0.1:8090/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_notification_officer",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-notification",
};

const feed = await fetch(`${baseURL}/notifications/alerts?includeExpired=true`);
if (!feed.ok) {
  throw new Error(
    `notification alert feed failed: ${feed.status} ${feed.statusText}`,
  );
}

const feedPayload = await feed.json();
if (!Array.isArray(feedPayload.alerts) || feedPayload.alerts.length < 2) {
  throw new Error("notification alert feed returned too few alerts");
}

const currentAlert = feedPayload.alerts.find(
  (alert) => alert.status === "current",
);
const expiredAlert = feedPayload.alerts.find(
  (alert) => alert.status === "expired",
);
if (!currentAlert || !expiredAlert) {
  throw new Error(
    "notification alert feed must include current and expired alerts",
  );
}
console.log(
  `notification alert feed OK current=${currentAlert.id} expired=${expiredAlert.id}`,
);

const deliver = await fetch(
  `${baseURL}/notifications/alerts/${currentAlert.id}/deliver`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      recipientId: "usr_demo_citizen",
      phone: "+233200000000",
      pushToken: "ExponentPushToken-smoke",
      channels: ["push", "sms"],
    }),
  },
);
if (deliver.status !== 202) {
  throw new Error(`notification delivery expected 202, got ${deliver.status}`);
}

const deliverPayload = await deliver.json();
if (
  !Array.isArray(deliverPayload.attempts) ||
  deliverPayload.attempts.length !== 2 ||
  !deliverPayload.attempts.every(
    (attempt) =>
      attempt.status === "delivered" &&
      ["mock_push", "mock_sms"].includes(attempt.provider),
  )
) {
  throw new Error(
    "notification delivery did not log mock push and SMS attempts",
  );
}
console.log("notification delivery OK mock push+sms");

const logs = await fetch(
  `${baseURL}/notifications/delivery-logs?alertId=${currentAlert.id}`,
);
if (!logs.ok) {
  throw new Error(
    `notification delivery logs failed: ${logs.status} ${logs.statusText}`,
  );
}
const logsPayload = await logs.json();
if (
  !Array.isArray(logsPayload.logs) ||
  logsPayload.logs.length < 2 ||
  !logsPayload.logs.some((log) => log.channel === "push") ||
  !logsPayload.logs.some((log) => log.channel === "sms")
) {
  throw new Error("notification delivery logs missing push or SMS entries");
}
console.log(`notification delivery logs OK ${logsPayload.logs.length}`);

const invalid = await fetch(
  `${baseURL}/notifications/alerts/${currentAlert.id}/deliver`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      recipientId: "usr_demo_citizen",
      channels: ["email"],
    }),
  },
);
if (invalid.status !== 400) {
  throw new Error(
    `invalid notification channel expected 400, got ${invalid.status}`,
  );
}
console.log("invalid-channel notification OK 400");
