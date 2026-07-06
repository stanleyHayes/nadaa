const baseURL =
  process.env.ALERT_API_URL?.trim() || "http://127.0.0.1:8089/api/v1";

const drafterHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_drafter",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-alert-draft",
};

const approverHeaders = {
  ...drafterHeaders,
  "X-NADAA-Actor-ID": "usr_smoke_approver",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Request-ID": "smoke-alert-approve",
};

const now = Date.now();
const body = {
  title: "Smoke flood warning",
  hazardType: "flood",
  severity: "warning",
  message: "Smoke test warning for flood-prone roads near Accra Central.",
  target: {
    type: "district",
    ids: ["accra-metropolitan"],
    label: "Accra Metropolitan",
  },
  startsAt: new Date(now - 5 * 60 * 1000).toISOString(),
  expiresAt: new Date(now + 6 * 60 * 60 * 1000).toISOString(),
  recommendedAction: "Avoid flooded roads and follow NADMO instructions.",
  evacuationRequired: false,
  shelterIds: ["00000000-0000-0000-0000-000000000301"],
};

const create = await fetch(`${baseURL}/alerts`, {
  method: "POST",
  headers: drafterHeaders,
  body: JSON.stringify(body),
});
if (create.status !== 201) {
  throw new Error(`alert create smoke expected 201, got ${create.status}`);
}
const draft = await create.json();
if (draft.status !== "draft") {
  throw new Error(`alert create smoke expected draft, got ${draft.status}`);
}
console.log(`alert create OK ${draft.id}`);

const submit = await fetch(`${baseURL}/alerts/${draft.id}/submit`, {
  method: "POST",
  headers: drafterHeaders,
});
if (!submit.ok) {
  throw new Error(`alert submit smoke failed: ${submit.status}`);
}
const submitted = await submit.json();
if (submitted.status !== "submitted") {
  throw new Error(
    `alert submit smoke expected submitted, got ${submitted.status}`,
  );
}
console.log("alert submit OK submitted");

const approve = await fetch(`${baseURL}/alerts/${draft.id}/approve`, {
  method: "POST",
  headers: approverHeaders,
  body: JSON.stringify({ note: "Smoke approver review complete." }),
});
if (!approve.ok) {
  throw new Error(`alert approve smoke failed: ${approve.status}`);
}
const approved = await approve.json();
if (
  approved.status !== "approved" ||
  approved.approvedBy !== "usr_smoke_approver"
) {
  throw new Error("alert approve smoke returned invalid approved alert");
}
console.log("alert approve OK approved");

const publicList = await fetch(`${baseURL}/alerts?current=true`);
if (!publicList.ok) {
  throw new Error(`alert public list smoke failed: ${publicList.status}`);
}
const publicPayload = await publicList.json();
if (
  !Array.isArray(publicPayload.alerts) ||
  !publicPayload.alerts.some((alert) => alert.id === draft.id)
) {
  throw new Error(
    "alert public list smoke did not include approved current alert",
  );
}
console.log("alert public list OK current approved");

const audit = await fetch(`${baseURL}/alerts/audit?limit=10`, {
  headers: approverHeaders,
});
if (!audit.ok) {
  throw new Error(`alert audit smoke failed: ${audit.status}`);
}
const auditPayload = await audit.json();
if (!Array.isArray(auditPayload.logs) || auditPayload.logs.length < 3) {
  throw new Error("alert audit smoke expected workflow audit logs");
}
console.log(`alert audit OK ${auditPayload.logs.length}`);

const invalid = await fetch(`${baseURL}/alerts`, {
  method: "POST",
  headers: { ...drafterHeaders, "X-NADAA-MFA-Completed": "false" },
  body: JSON.stringify(body),
});
if (invalid.status !== 403) {
  throw new Error(`alert MFA smoke expected 403, got ${invalid.status}`);
}
console.log("alert MFA gate OK 403");
