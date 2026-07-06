const baseURL =
  process.env.INCIDENT_API_URL?.trim() || "http://127.0.0.1:8084/api/v1";

const fireAgencyId = "00000000-0000-0000-0000-000000000201";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_dispatcher",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-incident-assignment",
};

const report = {
  type: "fire",
  description: "Smoke test report for a small electrical fire behind a kiosk.",
  location: { lat: 5.544, lng: -0.213 },
  peopleAffected: 3,
  injuriesReported: false,
  urgency: "high",
  anonymous: false,
  contactPermission: true,
  accessibilityNeeds: "",
  media: [],
  reporter: { userId: "usr_assignment_smoke", phone: "+233200000098" },
};

const assignment = {
  agencyId: fireAgencyId,
  agencyName: "Ghana National Fire Service",
  agencyType: "fire",
  priority: "urgent",
  instructions: "Dispatch engine crew and confirm hydrant access.",
  responderLead: "Station Officer Mensah",
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
console.log(`incident assignment create OK ${incident.id}`);

const verify = await fetch(`${baseURL}/incidents/${incident.id}/verify`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({ note: "Smoke dispatcher verified assignment case." }),
});
if (!verify.ok) {
  throw new Error(`incident verify smoke failed: ${verify.status}`);
}
console.log("incident assignment verify OK verified");

const assign = await fetch(`${baseURL}/incidents/${incident.id}/assignments`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify(assignment),
});
if (assign.status !== 201) {
  throw new Error(`incident assign smoke expected 201, got ${assign.status}`);
}
const assigned = await assign.json();
if (
  assigned.status !== "assigned" ||
  !assigned.assignments?.some((item) => item.agencyId === fireAgencyId) ||
  !assigned.timeline?.some((item) => item.type === "incident.assigned")
) {
  throw new Error("incident assign smoke returned invalid assignment payload");
}
console.log(`incident assignment OK ${assigned.assignments.length}`);

const assignedToMe = await fetch(`${baseURL}/incidents?assignedToMe=true`, {
  headers: {
    ...authorityHeaders,
    "X-NADAA-Actor-ID": "usr_smoke_responder",
    "X-NADAA-Actor-Role": "responder",
    "X-NADAA-Agency-ID": fireAgencyId,
  },
});
if (!assignedToMe.ok) {
  throw new Error(`assigned-to-me smoke failed: ${assignedToMe.status}`);
}
const assignedToMePayload = await assignedToMe.json();
const assignedToMeIncident = assignedToMePayload.incidents?.find(
  (item) => item.id === incident.id,
);
if (!assignedToMeIncident) {
  throw new Error("assigned-to-me smoke did not return assigned incident");
}
if (
  assignedToMeIncident.reportedBy ||
  assignedToMeIncident.privacy?.reporterIdentityVisible ||
  assignedToMeIncident.privacy?.reporterContactVisible
) {
  throw new Error("assigned-to-me responder view should hide reporter details");
}
console.log(`assigned-to-me OK ${assignedToMePayload.incidents.length}`);

const otherAgency = await fetch(`${baseURL}/incidents?assignedToMe=true`, {
  headers: {
    ...authorityHeaders,
    "X-NADAA-Actor-ID": "usr_other_responder",
    "X-NADAA-Actor-Role": "responder",
    "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000999",
  },
});
if (!otherAgency.ok) {
  throw new Error(`other-agency assigned-to-me failed: ${otherAgency.status}`);
}
const otherAgencyPayload = await otherAgency.json();
if (otherAgencyPayload.incidents?.some((item) => item.id === incident.id)) {
  throw new Error("other agency should not see the assigned incident");
}
console.log("assigned-to-me agency boundary OK");

const forbiddenAgencyAdmin = await fetch(
  `${baseURL}/incidents/${incident.id}/assignments`,
  {
    method: "POST",
    headers: {
      ...authorityHeaders,
      "X-NADAA-Actor-ID": "usr_smoke_agency_admin",
      "X-NADAA-Actor-Role": "agency_admin",
      "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
    },
    body: JSON.stringify(assignment),
  },
);
if (forbiddenAgencyAdmin.status !== 403) {
  throw new Error(
    `cross-agency assignment expected 403, got ${forbiddenAgencyAdmin.status}`,
  );
}
console.log("cross-agency assignment gate OK 403");
