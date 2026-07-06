const baseURL =
  process.env.NOTIFICATION_API_URL?.trim() || "http://127.0.0.1:8090/api/v1";

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
    headers: { "Content-Type": "application/json" },
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
    headers: { "Content-Type": "application/json" },
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
