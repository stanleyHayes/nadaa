const baseURL =
  process.env.INCIDENT_API_URL?.trim() || "http://127.0.0.1:8084/api/v1";

const authorityHeaders = {
  "X-NADAA-Actor-ID": "usr_smoke_dispatcher",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-distress",
};

const create = await fetch(`${baseURL}/incidents`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    requestKind: "distress_request",
    type: "flood",
    description: "I am trapped by rising water and need rescue now.",
    location: { lat: 5.579, lng: -0.212 },
    peopleAffected: 1,
    injuriesReported: false,
    urgency: "low",
    anonymous: true,
    contactPermission: false,
    media: [],
  }),
});
if (create.status !== 201) {
  throw new Error(`distress create expected 201, got ${create.status}`);
}
const distress = await create.json();
if (
  !distress.reference?.startsWith("SOS-") ||
  distress.requestKind !== "distress_request" ||
  !distress.rescueRequested ||
  !distress.priorityReview ||
  distress.severity !== "emergency"
) {
  throw new Error("distress create returned an invalid rescue response");
}
console.log(`distress create OK ${distress.reference}`);

const list = await fetch(`${baseURL}/incidents`, { headers: authorityHeaders });
if (!list.ok) throw new Error(`distress list failed: ${list.status}`);
const listPayload = await list.json();
const queued = listPayload.incidents?.find(
  (incident) => incident.id === distress.id,
);
if (
  !queued?.rescueRequested ||
  queued.requestKind !== "distress_request" ||
  queued.timeline?.[0]?.type !== "incident.distress_requested"
) {
  throw new Error("dispatcher queue did not preserve distress metadata");
}
console.log("distress dispatcher queue OK emergency");

const audit = await fetch(`${baseURL}/incidents/audit?limit=10`, {
  headers: authorityHeaders,
});
if (!audit.ok) throw new Error(`distress audit failed: ${audit.status}`);
const auditPayload = await audit.json();
if (
  !auditPayload.logs?.some(
    (entry) =>
      entry.action === "incident.distress_requested" &&
      entry.targetId === distress.id,
  )
) {
  throw new Error("distress audit event was not recorded");
}
console.log("distress audit OK incident.distress_requested");

const missingLocation = await fetch(`${baseURL}/incidents`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    requestKind: "distress_request",
    type: "medical_emergency",
    description: "I need urgent rescue assistance.",
    peopleAffected: 1,
    urgency: "life_threatening",
    anonymous: true,
    contactPermission: false,
    media: [],
  }),
});
if (missingLocation.status !== 400) {
  throw new Error(
    `distress location guard expected 400, got ${missingLocation.status}`,
  );
}
console.log("distress location guard OK 400");
