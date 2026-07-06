const baseURL =
  process.env.INCIDENT_API_URL?.trim() || "http://127.0.0.1:8084/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_dispatcher",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-incident-abuse",
};

const reporterSuffix = Date.now();

const buildReport = (description, reporterId, urgency = "high") => ({
  type: "flood",
  description,
  location: { lat: 5.579, lng: -0.212 },
  peopleAffected: 2,
  injuriesReported: urgency === "life_threatening",
  urgency,
  anonymous: false,
  contactPermission: true,
  accessibilityNeeds: "",
  media: [],
  reporter: { userId: reporterId, phone: "+233200000099" },
});

const createIncident = async (report) => {
  const response = await fetch(`${baseURL}/incidents`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(report),
  });
  if (response.status !== 201) {
    throw new Error(`incident create expected 201, got ${response.status}`);
  }
  return response.json();
};

const first = await createIncident(
  buildReport(
    "Free money promo click here https://example.test emergency emergency emergency",
    `usr_smoke_abuse_${reporterSuffix}`,
    "life_threatening",
  ),
);
if (
  first.status !== "reported" ||
  !first.priorityReview ||
  !first.abuseReviewRequired ||
  !Array.isArray(first.abuseSignals) ||
  first.abuseSignals.length < 2
) {
  throw new Error(
    "abuse create smoke expected live suspicious priority report",
  );
}
console.log(`incident abuse create OK ${first.id}`);

const clear = await fetch(`${baseURL}/incidents/${first.id}/abuse-review`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    decision: "clear",
    note: "Smoke dispatcher confirmed this suspicious report is legitimate.",
  }),
});
if (!clear.ok) {
  throw new Error(`incident abuse clear failed: ${clear.status}`);
}
const cleared = await clear.json();
if (cleared.abuseReviewRequired || cleared.abuseReviewDecision !== "clear") {
  throw new Error("incident abuse clear returned invalid review metadata");
}
console.log("incident abuse review OK clear");

const second = await createIncident(
  buildReport(
    "Free money promo click here https://example.test",
    `usr_smoke_false_${reporterSuffix}`,
  ),
);

const missingResolution = await fetch(
  `${baseURL}/incidents/${second.id}/abuse-review`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      decision: "false_report",
      note: "Smoke dispatcher confirmed no emergency.",
    }),
  },
);
if (missingResolution.status !== 400) {
  throw new Error(
    `incident false report resolution smoke expected 400, got ${missingResolution.status}`,
  );
}
console.log("incident abuse false-report reason gate OK 400");

const falseReport = await fetch(
  `${baseURL}/incidents/${second.id}/abuse-review`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      decision: "false_report",
      note: "Smoke dispatcher confirmed no emergency at the reported location.",
      resolutionNotes:
        "District desk and caller callback confirmed there is no active incident.",
    }),
  },
);
if (!falseReport.ok) {
  throw new Error(`incident abuse false-report failed: ${falseReport.status}`);
}
const falseReportPayload = await falseReport.json();
if (
  falseReportPayload.status !== "false_report" ||
  falseReportPayload.abuseReviewRequired ||
  !falseReportPayload.resolutionNotes
) {
  throw new Error("incident abuse false-report returned invalid closure");
}
console.log("incident abuse review OK false_report");

const audit = await fetch(`${baseURL}/incidents/audit?limit=10`, {
  headers: authorityHeaders,
});
if (!audit.ok) {
  throw new Error(`incident abuse audit smoke failed: ${audit.status}`);
}
const auditPayload = await audit.json();
const actions = new Set(auditPayload.logs?.map((log) => log.action) ?? []);
if (
  !actions.has("incident.abuse_cleared") ||
  !actions.has("incident.false_reported")
) {
  throw new Error("incident abuse audit smoke expected moderation audit logs");
}
console.log(`incident abuse audit OK ${auditPayload.logs.length}`);
