const baseURL =
  process.env.INCIDENT_API_URL?.trim() || "http://127.0.0.1:8084/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_dispatcher",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-incident-merge",
};

const primaryReport = {
  type: "flood",
  description: "Smoke merge primary report for flood water blocking a road.",
  location: { lat: 5.579, lng: -0.212 },
  peopleAffected: 4,
  injuriesReported: false,
  urgency: "high",
  anonymous: false,
  contactPermission: true,
  accessibilityNeeds: "",
  media: [],
  reporter: { userId: "usr_merge_primary", phone: "+233200000097" },
};

const duplicateReport = {
  ...primaryReport,
  description: "Smoke merge duplicate report for the same flooded road.",
  reporter: { userId: "usr_merge_duplicate", phone: "+233200000096" },
};

async function createIncident(report) {
  const response = await fetch(`${baseURL}/incidents`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(report),
  });
  if (response.status !== 201) {
    throw new Error(`incident create expected 201, got ${response.status}`);
  }
  return response.json();
}

const primary = await createIncident(primaryReport);
const duplicate = await createIncident(duplicateReport);
console.log(`incident merge create OK ${primary.id} ${duplicate.id}`);

const review = await fetch(`${baseURL}/incidents/${primary.id}/duplicates`, {
  headers: authorityHeaders,
});
if (!review.ok) {
  throw new Error(`duplicate review smoke failed: ${review.status}`);
}
const reviewPayload = await review.json();
if (
  reviewPayload.incident?.id !== primary.id ||
  !reviewPayload.candidates?.some(
    (candidate) => candidate.incident?.id === duplicate.id,
  )
) {
  throw new Error("duplicate review smoke did not return side-by-side match");
}
console.log(`duplicate review OK ${reviewPayload.candidates.length}`);

const merge = await fetch(`${baseURL}/incidents/${primary.id}/merge`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    duplicateIncidentIds: [duplicate.id],
    note: "Smoke merge confirmed both reports describe the same flooded road.",
  }),
});
if (!merge.ok) {
  throw new Error(`incident merge smoke failed: ${merge.status}`);
}
const mergePayload = await merge.json();
if (
  !mergePayload.incident?.mergedIncidentIds?.includes(duplicate.id) ||
  mergePayload.mergedIncidents?.[0]?.mergedIntoId !== primary.id ||
  mergePayload.mergedIncidents?.[0]?.status !== "closed"
) {
  throw new Error("incident merge smoke returned invalid trace payload");
}
console.log("incident merge OK trace fields");

const postReview = await fetch(
  `${baseURL}/incidents/${primary.id}/duplicates`,
  {
    headers: authorityHeaders,
  },
);
if (!postReview.ok) {
  throw new Error(`post-merge review failed: ${postReview.status}`);
}
const postReviewPayload = await postReview.json();
if (
  postReviewPayload.candidates?.some(
    (candidate) => candidate.incident?.id === duplicate.id,
  )
) {
  throw new Error("merged duplicate should not remain an open candidate");
}
console.log("post-merge review OK no open duplicate");

const audit = await fetch(`${baseURL}/incidents/audit?limit=5`, {
  headers: authorityHeaders,
});
if (!audit.ok) {
  throw new Error(`incident audit smoke failed: ${audit.status}`);
}
const auditPayload = await audit.json();
if (!auditPayload.logs?.some((log) => log.action === "incident.merged")) {
  throw new Error("incident merge smoke expected incident.merged audit log");
}
console.log("incident merge audit OK");
