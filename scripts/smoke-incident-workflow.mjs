const baseURL =
  process.env.INCIDENT_API_URL?.trim() || "http://127.0.0.1:8084/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_dispatcher",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-incident-workflow",
};

const report = {
  type: "flood",
  description: "Smoke test report for flood water blocking a road.",
  location: { lat: 5.579, lng: -0.212 },
  peopleAffected: 4,
  injuriesReported: false,
  urgency: "high",
  anonymous: false,
  contactPermission: true,
  accessibilityNeeds: "",
  media: [],
  reporter: { userId: "usr_smoke_citizen", phone: "+233200000099" },
};

const create = await fetch(`${baseURL}/incidents`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(report),
});
if (create.status !== 201) {
  throw new Error(`incident create smoke expected 201, got ${create.status}`);
}
const incident = await create.json();
if (incident.status !== "reported") {
  throw new Error(
    `incident create smoke expected reported, got ${incident.status}`,
  );
}
console.log(`incident create OK ${incident.id}`);

const verify = await fetch(`${baseURL}/incidents/${incident.id}/verify`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({ note: "Smoke dispatcher verified the report." }),
});
if (!verify.ok) {
  throw new Error(`incident verify smoke failed: ${verify.status}`);
}
const verified = await verify.json();
if (
  verified.status !== "verified" ||
  verified.verifiedBy !== "usr_smoke_dispatcher"
) {
  throw new Error("incident verify smoke returned invalid verified incident");
}
console.log("incident verify OK verified");

for (const status of [
  "assigned",
  "response_en_route",
  "on_scene",
  "contained",
  "recovery_ongoing",
]) {
  const response = await fetch(`${baseURL}/incidents/${incident.id}/status`, {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({
      status,
      note: `Smoke transition to ${status}.`,
    }),
  });
  if (!response.ok) {
    throw new Error(`incident ${status} smoke failed: ${response.status}`);
  }
  const payload = await response.json();
  if (payload.status !== status) {
    throw new Error(`incident smoke expected ${status}, got ${payload.status}`);
  }
  console.log(`incident status OK ${status}`);
}

const missingResolution = await fetch(
  `${baseURL}/incidents/${incident.id}/status`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({ status: "closed", note: "Missing closure notes." }),
  },
);
if (missingResolution.status !== 400) {
  throw new Error(
    `incident closure notes smoke expected 400, got ${missingResolution.status}`,
  );
}
console.log("incident closure notes gate OK 400");

const close = await fetch(`${baseURL}/incidents/${incident.id}/status`, {
  method: "PATCH",
  headers: authorityHeaders,
  body: JSON.stringify({
    status: "closed",
    note: "Smoke test response complete.",
    resolutionNotes: "Flood smoke test incident closed after response review.",
  }),
});
if (!close.ok) {
  throw new Error(`incident close smoke failed: ${close.status}`);
}
const closed = await close.json();
if (closed.status !== "closed" || !closed.resolutionNotes) {
  throw new Error("incident close smoke returned invalid closed incident");
}
console.log("incident close OK closed");

const audit = await fetch(`${baseURL}/incidents/audit?limit=10`, {
  headers: authorityHeaders,
});
if (!audit.ok) {
  throw new Error(`incident audit smoke failed: ${audit.status}`);
}
const auditPayload = await audit.json();
if (!Array.isArray(auditPayload.logs) || auditPayload.logs.length < 7) {
  throw new Error("incident audit smoke expected workflow audit logs");
}
console.log(`incident audit OK ${auditPayload.logs.length}`);

const invalidMFA = await fetch(`${baseURL}/incidents/${incident.id}/verify`, {
  method: "POST",
  headers: { ...authorityHeaders, "X-NADAA-MFA-Completed": "false" },
  body: JSON.stringify({ note: "Missing MFA." }),
});
if (invalidMFA.status !== 403) {
  throw new Error(`incident MFA smoke expected 403, got ${invalidMFA.status}`);
}
console.log("incident MFA gate OK 403");
