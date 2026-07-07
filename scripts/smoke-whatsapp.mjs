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

const alerts = await postJSON(
  "/notifications/whatsapp/webhook",
  {
    from: "+233200000200",
    body: "ALERTS",
    provider: "whatsapp-smoke",
  },
  202,
);
if (!alerts.message.includes("warning")) {
  throw new Error("WhatsApp ALERTS did not return an alert summary");
}
console.log("WhatsApp alerts OK");

const risk = await postJSON(
  "/notifications/whatsapp/webhook",
  {
    from: "+233200000200",
    body: "RISK",
    provider: "whatsapp-smoke",
    location: { lat: 5.566, lng: -0.242 },
  },
  202,
);
if (!risk.message.includes("Location received")) {
  throw new Error("WhatsApp RISK did not acknowledge location");
}
console.log("WhatsApp risk OK");

const guide = await postJSON(
  "/notifications/whatsapp/webhook",
  {
    from: "+233200000200",
    body: "GUIDE FLOOD",
    provider: "whatsapp-smoke",
  },
  202,
);
if (!guide.message.includes("Flood guide")) {
  throw new Error("WhatsApp GUIDE FLOOD did not return guide text");
}
console.log("WhatsApp guide OK");

const startReport = await postJSON(
  "/notifications/whatsapp/webhook",
  {
    from: "+233200000201",
    body: "REPORT",
    provider: "whatsapp-smoke",
  },
  202,
);
if (
  startReport.conversation?.state !== "awaiting_report_hazard" ||
  !startReport.message.includes("What type")
) {
  throw new Error("WhatsApp REPORT did not start hazard prompt");
}
console.log("WhatsApp report start OK");

const reportHazard = await postJSON(
  "/notifications/whatsapp/webhook",
  {
    from: "+233200000201",
    body: "FLOOD",
    provider: "whatsapp-smoke",
  },
  202,
);
if (reportHazard.conversation?.state !== "awaiting_report_urgency") {
  throw new Error("WhatsApp report hazard did not advance to urgency");
}
console.log("WhatsApp report hazard OK");

const reportUrgency = await postJSON(
  "/notifications/whatsapp/webhook",
  {
    from: "+233200000201",
    body: "HIGH",
    provider: "whatsapp-smoke",
  },
  202,
);
if (reportUrgency.conversation?.state !== "awaiting_report_location") {
  throw new Error("WhatsApp report urgency did not advance to location");
}
console.log("WhatsApp report urgency OK");

const reportComplete = await postJSON(
  "/notifications/whatsapp/webhook",
  {
    from: "+233200000201",
    body: "water rising near Circle",
    provider: "whatsapp-smoke",
    profileId: "usr_smoke_whatsapp",
    linkProfile: true,
    location: { lat: 5.579, lng: -0.212 },
    media: [{ id: "wa_smoke_media_001", contentType: "image/jpeg" }],
  },
  202,
);
if (
  !reportComplete.report ||
  reportComplete.report.channel !== "whatsapp" ||
  !["queued", "submitted"].includes(reportComplete.report.status) ||
  reportComplete.conversation?.state !== "idle" ||
  reportComplete.report.media?.[0] !== "wa_smoke_media_001"
) {
  throw new Error("WhatsApp report did not create a completed report");
}
console.log(`WhatsApp report OK ${reportComplete.report.status}`);

const providerError = await postJSON(
  "/notifications/whatsapp/webhook",
  {
    from: "+233200000202",
    provider: "whatsapp-smoke",
    providerError: "sandbox provider signature failed",
  },
  202,
);
if (
  providerError.log?.status !== "failed" ||
  providerError.log?.intent !== "provider_error"
) {
  throw new Error("WhatsApp provider error was not logged as failed");
}
console.log("WhatsApp provider error log OK");

const logs = await fetch(
  `${baseURL}/notifications/access-logs?channel=whatsapp`,
);
if (!logs.ok) {
  throw new Error(
    `WhatsApp access logs failed: ${logs.status} ${logs.statusText}`,
  );
}
const logsPayload = await logs.json();
if (!Array.isArray(logsPayload.logs) || logsPayload.logs.length < 7) {
  throw new Error("WhatsApp access logs missing expected entries");
}
console.log(`WhatsApp access logs OK ${logsPayload.logs.length}`);
