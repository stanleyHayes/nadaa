const baseURL =
  process.env.INCIDENT_API_URL?.trim() || "http://127.0.0.1:8084/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_dispatcher",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-ai-triage",
};

const reporterSuffix = Date.now();

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

const first = await createIncident({
  type: "flood",
  description: `Smoke AI triage baseline report at the market road (${reporterSuffix}).`,
  location: { lat: 5.579, lng: -0.212 },
  peopleAffected: 28,
  injuriesReported: false,
  urgency: "high",
  anonymous: false,
  contactPermission: true,
  accessibilityNeeds: "",
  media: [],
  reporter: {
    userId: `usr_smoke_triage_${reporterSuffix}`,
    phone: "+233200000099",
  },
});
console.log(`ai triage incident create OK ${first.id}`);

const second = await createIncident({
  type: "flood",
  description: `Smoke AI triage duplicate report near the same market road (${reporterSuffix}).`,
  location: { lat: 5.5791, lng: -0.2121 },
  peopleAffected: 9,
  injuriesReported: false,
  urgency: "high",
  anonymous: false,
  contactPermission: true,
  accessibilityNeeds: "",
  media: [],
  reporter: {
    userId: `usr_smoke_triage_dup_${reporterSuffix}`,
    phone: "+233200000098",
  },
});
console.log(`ai triage duplicate incident create OK ${second.id}`);

const suggestionResponse = await fetch(
  `${baseURL}/incidents/${second.id}/triage`,
  {
    headers: authorityHeaders,
  },
);
if (!suggestionResponse.ok) {
  throw new Error(`ai triage suggestion failed: ${suggestionResponse.status}`);
}
const suggestionPayload = await suggestionResponse.json();
const suggestion = suggestionPayload.suggestion;

if (
  !suggestion ||
  !suggestion.humanReviewRequired ||
  suggestion.autoPublishAllowed !== false
) {
  throw new Error("ai triage suggestion missing safety flags");
}
if (!["low", "moderate", "high", "emergency"].includes(suggestion.severity)) {
  throw new Error(
    `ai triage suggestion invalid severity: ${suggestion.severity}`,
  );
}
if (typeof suggestion.duplicateLikelihood !== "number") {
  throw new Error("ai triage suggestion missing duplicate likelihood");
}
if (
  !Array.isArray(suggestion.topDuplicateIncidentIds) ||
  !suggestion.topDuplicateIncidentIds.includes(first.id)
) {
  throw new Error("ai triage suggestion missing expected duplicate candidate");
}
if (!suggestion.suggestedAgency?.agencyType) {
  throw new Error("ai triage suggestion missing suggested agency");
}
if (
  !Array.isArray(suggestion.explanationFactors) ||
  suggestion.explanationFactors.length === 0
) {
  throw new Error("ai triage suggestion missing explanation factors");
}
if (!suggestion.suggestionId?.startsWith("trs_")) {
  throw new Error("ai triage suggestion missing trs_ suggestion id");
}
console.log("ai triage suggestion OK", {
  suggestionId: suggestion.suggestionId,
  severity: suggestion.severity,
  duplicateLikelihood: suggestion.duplicateLikelihood,
  confidence: suggestion.confidence,
});

const missingReason = await fetch(
  `${baseURL}/incidents/${second.id}/triage-review`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      accepted: false,
      overriddenFields: { severity: "emergency" },
    }),
  },
);
if (missingReason.status !== 400) {
  throw new Error(
    `ai triage override reason gate expected 400, got ${missingReason.status}`,
  );
}
console.log("ai triage override reason gate OK 400");

const emptyOverride = await fetch(
  `${baseURL}/incidents/${second.id}/triage-review`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      accepted: false,
      overriddenFields: {},
      reason: "Empty override objects must be rejected.",
    }),
  },
);
if (emptyOverride.status !== 400) {
  throw new Error(
    `ai triage empty override gate expected 400, got ${emptyOverride.status}`,
  );
}
console.log("ai triage empty override gate OK 400");

const unknownSuggestion = await fetch(
  `${baseURL}/incidents/${second.id}/triage-review`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      accepted: true,
      suggestionId: "trs_does_not_exist",
    }),
  },
);
if (unknownSuggestion.status !== 400) {
  throw new Error(
    `ai triage unknown suggestion gate expected 400, got ${unknownSuggestion.status}`,
  );
}
console.log("ai triage unknown suggestion gate OK 400");

const overrideResponse = await fetch(
  `${baseURL}/incidents/${second.id}/triage-review`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      accepted: false,
      suggestionId: suggestion.suggestionId,
      overriddenFields: {
        severity: "emergency",
        affectedPopulation: 60,
        suggestedAgencyType: "fire",
        suggestedAgencyId: "00000000-0000-0000-0000-000000000201",
      },
      reason:
        "Smoke dispatcher upgraded severity after callback confirmed entrapment.",
    }),
  },
);
if (!overrideResponse.ok) {
  throw new Error(`ai triage override failed: ${overrideResponse.status}`);
}
const overridePayload = await overrideResponse.json();
if (
  !overridePayload.incident?.timeline?.some(
    (event) => event.type === "incident.triage_overridden",
  )
) {
  throw new Error("ai triage override missing timeline event");
}
console.log("ai triage override OK");

const firstSuggestionResponse = await fetch(
  `${baseURL}/incidents/${first.id}/triage`,
  {
    headers: authorityHeaders,
  },
);
if (!firstSuggestionResponse.ok) {
  throw new Error(
    `ai triage first suggestion failed: ${firstSuggestionResponse.status}`,
  );
}
const firstSuggestion = (await firstSuggestionResponse.json()).suggestion;

const acceptResponse = await fetch(
  `${baseURL}/incidents/${first.id}/triage-review`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({
      accepted: true,
      suggestionId: firstSuggestion.suggestionId,
    }),
  },
);
if (!acceptResponse.ok) {
  throw new Error(`ai triage accept failed: ${acceptResponse.status}`);
}
const acceptPayload = await acceptResponse.json();
if (
  !acceptPayload.incident?.timeline?.some(
    (event) => event.type === "incident.triage_accepted",
  )
) {
  throw new Error("ai triage accept missing timeline event");
}
console.log("ai triage accept OK");

const audit = await fetch(`${baseURL}/incidents/audit?limit=20`, {
  headers: authorityHeaders,
});
if (!audit.ok) {
  throw new Error(`ai triage audit smoke failed: ${audit.status}`);
}
const auditPayload = await audit.json();
const actions = new Set(auditPayload.logs?.map((log) => log.action) ?? []);
if (
  !actions.has("incident.triage_suggested") ||
  !actions.has("incident.triage_overridden") ||
  !actions.has("incident.triage_accepted")
) {
  throw new Error(
    "ai triage audit smoke expected suggested, override, and accept audit logs",
  );
}
console.log(`ai triage audit OK ${auditPayload.logs.length}`);
