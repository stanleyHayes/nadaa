import { createHmac } from "node:crypto";

const baseURL =
  process.env.INCIDENT_API_URL?.trim() || "http://127.0.0.1:8084/api/v1";

// Citizen endpoints (volunteer registration, volunteer task self-access)
// require a verified citizen bearer token. Dev backends run with
// NADAA_AUTH_TOKEN_SECRET=dev-secret-change-me (see
// scripts/dev-citizen-backends.sh), so the smoke signs its own token.
const tokenSecret =
  process.env.NADAA_AUTH_TOKEN_SECRET?.trim() || "dev-secret-change-me";

function base64url(value) {
  return Buffer.from(value).toString("base64url");
}

function citizenToken(subject) {
  const payload = base64url(
    JSON.stringify({
      sub: subject,
      typ: "citizen",
      role: "citizen",
      exp: Math.floor(Date.now() / 1000) + 3600,
    }),
  );
  const signature = createHmac("sha256", tokenSecret)
    .update(payload)
    .digest("base64url");
  return `nadaa.${payload}.${signature}`;
}

const citizenHeaders = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${citizenToken("usr_smoke_volunteer_001")}`,
};

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_volunteer_dispatcher",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-community-volunteers",
};

const volunteerProfile = {
  availabilityStatus: "available",
  citizenUserId: "usr_smoke_volunteer_001",
  community: "Jamestown",
  district: "Accra Metropolitan",
  languages: ["en", "tw"],
  name: "Smoke Volunteer",
  phone: "+233200000111",
  region: "Greater Accra",
  skills: ["first aid", "community alerts"],
};

const report = {
  anonymous: false,
  contactPermission: true,
  description: "Smoke test flood report for volunteer field coordination.",
  injuriesReported: false,
  location: { lat: 5.56, lng: -0.2 },
  media: [],
  peopleAffected: 5,
  reporter: { phone: "+233200000098", userId: "usr_volunteer_smoke_reporter" },
  type: "flood",
  urgency: "high",
};

const volunteer = await postJSON("/volunteers", volunteerProfile, citizenHeaders);
if (
  !volunteer.volunteer?.id ||
  volunteer.volunteer.verificationStatus !== "pending"
) {
  throw new Error("community volunteer smoke returned invalid profile");
}
console.log(`community volunteer register OK ${volunteer.volunteer.id}`);

const verified = await postJSON(
  `/volunteers/${volunteer.volunteer.id}/verify`,
  {
    decision: "verify",
    note: "Smoke verification completed by district officer.",
  },
  authorityHeaders,
);
if (verified.volunteer?.verificationStatus !== "verified") {
  throw new Error("community volunteer smoke did not verify profile");
}
console.log("community volunteer verify OK");

const incident = await postJSON("/incidents", report, {
  "Content-Type": "application/json",
});
console.log(`community volunteer incident create OK ${incident.id}`);

await postJSON(
  `/incidents/${incident.id}/verify`,
  { note: "Smoke dispatcher verified volunteer coordination case." },
  authorityHeaders,
);
console.log("community volunteer incident verify OK");

const task = await postJSON(
  `/incidents/${incident.id}/volunteer-tasks`,
  {
    instructions:
      "Check whether households near the shelter approach need water or transport. Stay on public roads.",
    locationLabel: "Jamestown shelter approach",
    priority: "high",
    type: "welfare_check",
    volunteerId: volunteer.volunteer.id,
  },
  authorityHeaders,
);
if (task.status !== "assigned" || task.volunteerId !== volunteer.volunteer.id) {
  throw new Error("community volunteer task assignment payload invalid");
}
console.log(`community volunteer task assign OK ${task.id}`);

const accepted = await patchJSON(
  `/volunteer-tasks/${task.id}/status`,
  {
    location: { lat: 5.561, lng: -0.201 },
    note: "Accepted from smoke script.",
    safetyStatus: "safe",
    status: "accepted",
    volunteerId: volunteer.volunteer.id,
  },
  citizenHeaders,
);
if (accepted.status !== "accepted" || !accepted.acceptedAt) {
  throw new Error("community volunteer task status update invalid");
}
console.log("community volunteer task status OK accepted");

const observed = await postJSON(
  `/volunteer-tasks/${task.id}/observations`,
  {
    escalationRequested: true,
    location: { lat: 5.562, lng: -0.202 },
    observation:
      "Water is rising near the footbridge and families are waiting for authority transport.",
    safetyStatus: "needs_authority",
    volunteerId: volunteer.volunteer.id,
  },
  citizenHeaders,
);
if (observed.status !== "needs_escalation" || !observed.escalationRequired) {
  throw new Error("community volunteer observation did not escalate task");
}
console.log("community volunteer observation OK needs_escalation");

const tasks = await getJSON(
  `/volunteers/${volunteer.volunteer.id}/tasks`,
  citizenHeaders,
);
if (!tasks.tasks?.some((item) => item.id === task.id)) {
  throw new Error("community volunteer task list missing assigned task");
}
console.log(`community volunteer task list OK ${tasks.tasks.length}`);

const incidents = await getJSON("/incidents", authorityHeaders);
const updated = incidents.incidents?.find((item) => item.id === incident.id);
for (const expected of [
  "incident.volunteer_assigned",
  "incident.volunteer_status_updated",
  "incident.volunteer_observation",
  "incident.volunteer_escalation",
]) {
  if (!updated?.timeline?.some((item) => item.type === expected)) {
    throw new Error(`community volunteer timeline missing ${expected}`);
  }
}
console.log("community volunteer timeline OK");

async function getJSON(path, headers = {}) {
  const response = await fetch(`${baseURL}${path}`, { headers });
  if (!response.ok) {
    throw new Error(`${path} expected OK, got ${response.status}`);
  }
  return response.json();
}

async function postJSON(path, body, headers) {
  const response = await fetch(`${baseURL}${path}`, {
    body: JSON.stringify(body),
    headers,
    method: "POST",
  });
  if (!response.ok) {
    throw new Error(`${path} expected OK, got ${response.status}`);
  }
  return response.json();
}

async function patchJSON(path, body, headers) {
  const response = await fetch(`${baseURL}${path}`, {
    body: JSON.stringify(body),
    headers,
    method: "PATCH",
  });
  if (!response.ok) {
    throw new Error(`${path} expected OK, got ${response.status}`);
  }
  return response.json();
}
