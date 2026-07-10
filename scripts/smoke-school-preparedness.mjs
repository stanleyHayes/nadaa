const baseURL =
  process.env.SCHOOL_API_URL?.trim() || "http://127.0.0.1:8097/api/v1";

const headers = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_test",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000204",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Actor-District": "Accra Metropolitan",
};

const adminHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_admin",
  "X-NADAA-Actor-Role": "system_admin",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000001",
  "X-NADAA-MFA-Completed": "true",
};

const health = await fetch(`${baseURL.replace("/api/v1", "")}/healthz`);
if (!health.ok) {
  throw new Error(`school-service health smoke failed: ${health.status}`);
}
const healthPayload = await health.json();
if (
  healthPayload.status !== "ok" ||
  healthPayload.service !== "school-service"
) {
  throw new Error("school-service health smoke returned invalid payload");
}
console.log("school-service health OK");

const list = await fetch(`${baseURL}/schools`, { headers });
if (!list.ok) {
  throw new Error(`school-service list smoke failed: ${list.status}`);
}
const listPayload = await list.json();
if (!Array.isArray(listPayload.schools)) {
  throw new Error("school-service list smoke returned invalid payload");
}
console.log(`school-service list OK ${listPayload.schools.length} schools`);

const adminList = await fetch(`${baseURL}/schools`, { headers: adminHeaders });
const adminListPayload = await adminList.json();
if (
  !Array.isArray(adminListPayload.schools) ||
  adminListPayload.schools.length < 3
) {
  throw new Error("school-service admin list smoke expected all schools");
}
console.log(
  `school-service admin list OK ${adminListPayload.schools.length} schools`,
);

const create = await fetch(`${baseURL}/schools`, {
  method: "POST",
  headers,
  body: JSON.stringify({
    name: "Smoke Test School",
    location: { lat: 5.55, lng: -0.19 },
    region: "Greater Accra",
    district: "Accra Metropolitan",
    address: "Smoke Test Road",
    studentPopulation: 250,
    emergencyContacts: [
      {
        name: "Smoke Head",
        role: "headteacher",
        phone: "+233200000999",
        isPrimary: true,
      },
    ],
    hazards: ["flood"],
    evacuationPoints: [
      {
        label: "Smoke Point",
        location: { lat: 5.551, lng: -0.191 },
        capacity: 300,
      },
    ],
  }),
});
if (!create.ok) {
  throw new Error(`school-service create smoke failed: ${create.status}`);
}
const created = await create.json();
if (!created.id || created.name !== "Smoke Test School") {
  throw new Error("school-service create smoke returned invalid payload");
}
console.log(`school-service create OK ${created.id}`);

const get = await fetch(`${baseURL}/schools/${created.id}`, { headers });
if (!get.ok) {
  throw new Error(`school-service get smoke failed: ${get.status}`);
}
const getPayload = await get.json();
if (getPayload.school.id !== created.id) {
  throw new Error("school-service get smoke returned invalid payload");
}
console.log(`school-service get OK ${getPayload.school.id}`);

const update = await fetch(`${baseURL}/schools/${created.id}`, {
  method: "PUT",
  headers,
  body: JSON.stringify({ studentPopulation: 300 }),
});
if (!update.ok) {
  throw new Error(`school-service update smoke failed: ${update.status}`);
}
const updated = await update.json();
if (updated.studentPopulation !== 300) {
  throw new Error("school-service update smoke returned invalid payload");
}
console.log("school-service update OK 300 students");

const drill = await fetch(`${baseURL}/schools/${created.id}/drills`, {
  method: "POST",
  headers,
  body: JSON.stringify({
    date: new Date().toISOString(),
    type: "fire",
    participants: 240,
    notes: "Smoke test drill.",
    completed: true,
  }),
});
if (!drill.ok) {
  throw new Error(`school-service drill create smoke failed: ${drill.status}`);
}
const drillRecord = await drill.json();
if (drillRecord.schoolId !== created.id) {
  throw new Error("school-service drill create smoke returned invalid payload");
}
console.log(`school-service drill create OK ${drillRecord.id}`);

const drills = await fetch(`${baseURL}/schools/${created.id}/drills`, {
  headers,
});
if (!drills.ok) {
  throw new Error(`school-service drills list smoke failed: ${drills.status}`);
}
const drillsPayload = await drills.json();
if (!Array.isArray(drillsPayload.drills) || drillsPayload.drills.length === 0) {
  throw new Error("school-service drills list smoke returned invalid payload");
}
console.log(
  `school-service drills list OK ${drillsPayload.drills.length} drills`,
);

const readiness = await fetch(`${baseURL}/schools/${created.id}/readiness`, {
  method: "POST",
  headers,
  body: JSON.stringify({
    checkDate: new Date().toISOString(),
    riskLevel: "high",
    areaRiskRef: "risk_smoke_001",
    overallStatus: "ready",
    checklistItems: [
      { label: "Contacts updated", checked: true, category: "admin" },
    ],
    notes: "Smoke test readiness.",
  }),
});
if (!readiness.ok) {
  throw new Error(
    `school-service readiness create smoke failed: ${readiness.status}`,
  );
}
const readinessRecord = await readiness.json();
if (readinessRecord.schoolId !== created.id) {
  throw new Error(
    "school-service readiness create smoke returned invalid payload",
  );
}
console.log(`school-service readiness create OK ${readinessRecord.id}`);

const getReadiness = await fetch(`${baseURL}/schools/${created.id}/readiness`, {
  headers,
});
if (!getReadiness.ok) {
  throw new Error(
    `school-service readiness get smoke failed: ${getReadiness.status}`,
  );
}
const readinessPayload = await getReadiness.json();
if (
  !readinessPayload.readiness ||
  readinessPayload.readiness.id !== readinessRecord.id
) {
  throw new Error(
    "school-service readiness get smoke returned invalid payload",
  );
}
console.log("school-service readiness get OK");

const scoped = await fetch(`${baseURL}/schools/school_002`, { headers });
if (scoped.status !== 403) {
  throw new Error(
    `school-service district scope smoke expected 403, got ${scoped.status}`,
  );
}
console.log("school-service district scope OK 403");

const invalid = await fetch(`${baseURL}/schools`, {
  method: "POST",
  headers,
  body: JSON.stringify({ name: "" }),
});
if (invalid.status !== 400) {
  throw new Error(
    `school-service invalid create smoke expected 400, got ${invalid.status}`,
  );
}
console.log("school-service invalid create OK 400");

console.log("school-preparedness smoke passed");
