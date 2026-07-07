const baseURL =
  process.env.NOTIFICATION_API_URL?.trim() || "http://127.0.0.1:8090/api/v1";

async function postJSON(path, body, expectedStatus) {
  const response = await fetch(`${baseURL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  if (response.status !== expectedStatus) {
    throw new Error(
      `${path} expected ${expectedStatus}, got ${response.status} ${response.statusText}`,
    );
  }

  return response.json();
}

const menu = await postJSON(
  "/notifications/ussd",
  {
    sessionId: "smoke-ussd-001",
    phone: "+233200000100",
    text: "",
  },
  200,
);
if (menu.action !== "continue" || !menu.message.includes("Select language")) {
  throw new Error("USSD language menu did not return a continue prompt");
}
console.log("USSD language menu OK");

const alert = await postJSON(
  "/notifications/ussd",
  {
    sessionId: "smoke-ussd-001",
    phone: "+233200000100",
    text: "1*1",
  },
  200,
);
if (alert.action !== "end" || !alert.message.includes("warning")) {
  throw new Error("USSD current alert summary did not return warning text");
}
console.log("USSD current alerts OK");

const ussdReport = await postJSON(
  "/notifications/ussd",
  {
    sessionId: "smoke-ussd-002",
    phone: "+233200000101",
    text: "1*2*1*3",
    profileId: "usr_smoke_ussd",
    linkProfile: true,
    location: { lat: 5.579, lng: -0.212 },
  },
  200,
);
if (
  !ussdReport.report ||
  !["queued", "submitted"].includes(ussdReport.report.status) ||
  !ussdReport.report.linkedProfile
) {
  throw new Error(
    "USSD report did not create a linked queued/submitted report",
  );
}
console.log(`USSD report OK ${ussdReport.report.status}`);

const smsAlerts = await postJSON(
  "/notifications/sms/inbound",
  {
    from: "+233200000102",
    body: "ALERTS",
    provider: "sms-smoke",
  },
  202,
);
if (!smsAlerts.message.includes("warning")) {
  throw new Error("SMS ALERTS did not return an alert summary");
}
console.log("SMS alerts OK");

const smsReport = await postJSON(
  "/notifications/sms/inbound",
  {
    from: "+233200000103",
    body: "REPORT FLOOD HIGH water rising near Circle",
    provider: "sms-smoke",
    profileId: "usr_smoke_sms",
    linkProfile: true,
    location: { lat: 5.566, lng: -0.242 },
  },
  202,
);
if (
  !smsReport.report ||
  !["queued", "submitted"].includes(smsReport.report.status)
) {
  throw new Error("SMS report did not create a queued/submitted report");
}
console.log(`SMS report OK ${smsReport.report.status}`);

const providerError = await postJSON(
  "/notifications/sms/inbound",
  {
    from: "+233200000104",
    provider: "sms-smoke",
    providerError: "sandbox provider signature failed",
  },
  202,
);
if (providerError.log.status !== "failed") {
  throw new Error("SMS provider error was not logged as failed");
}
console.log("SMS provider error log OK");

const logs = await fetch(`${baseURL}/notifications/access-logs?channel=sms`);
if (!logs.ok) {
  throw new Error(
    `inclusive access logs failed: ${logs.status} ${logs.statusText}`,
  );
}
const logsPayload = await logs.json();
if (!Array.isArray(logsPayload.logs) || logsPayload.logs.length < 3) {
  throw new Error("inclusive access logs missing SMS entries");
}
console.log(`inclusive access logs OK ${logsPayload.logs.length}`);
