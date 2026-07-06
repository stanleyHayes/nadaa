const baseURL =
  process.env.SHELTER_API_URL?.trim() || "http://127.0.0.1:8093/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_shelter_operator",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000204",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-shelter",
};

const nearby = await fetch(`${baseURL}/shelters/nearby?lat=5.5600&lng=-0.2000`);
if (!nearby.ok) {
  throw new Error(`nearby shelter smoke failed: ${nearby.status}`);
}
const nearbyPayload = await nearby.json();
if (
  !Array.isArray(nearbyPayload.shelters) ||
  nearbyPayload.shelters.length < 2 ||
  !Array.isArray(nearbyPayload.recoverySupport) ||
  nearbyPayload.recoverySupport.length === 0
) {
  throw new Error(
    "nearby shelter smoke expected shelters and recovery support",
  );
}
if (
  nearbyPayload.shelters[0].distanceMeters >
  nearbyPayload.shelters[1].distanceMeters
) {
  throw new Error("nearby shelter smoke expected distance-sorted shelters");
}
console.log(`nearby shelters OK ${nearbyPayload.shelters.length}`);

const recovery = await fetch(
  `${baseURL}/recovery-support/nearby?lat=5.5600&lng=-0.2000`,
);
if (!recovery.ok) {
  throw new Error(`recovery support smoke failed: ${recovery.status}`);
}
const recoveryPayload = await recovery.json();
if (
  !recoveryPayload.recoverySupport?.some((item) => item.type === "relief_point")
) {
  throw new Error("recovery support smoke expected relief point");
}
console.log(`recovery support OK ${recoveryPayload.recoverySupport.length}`);

const missingAuthority = await fetch(
  `${baseURL}/shelters/${nearbyPayload.shelters[0].id}/occupancy`,
  {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ currentOccupancy: 120 }),
  },
);
if (missingAuthority.status !== 401) {
  throw new Error(
    `shelter authority smoke expected 401, got ${missingAuthority.status}`,
  );
}
console.log("shelter authority gate OK 401");

const update = await fetch(
  `${baseURL}/shelters/${nearbyPayload.shelters[0].id}/occupancy`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({
      capacity: 450,
      currentOccupancy: 180,
      status: "open",
      notes: "Smoke update confirms district shelter occupancy workflow.",
    }),
  },
);
if (!update.ok) {
  throw new Error(`shelter occupancy update smoke failed: ${update.status}`);
}
const updatePayload = await update.json();
if (
  updatePayload.shelter?.currentOccupancy !== 180 ||
  updatePayload.shelter?.updatedBy !== "usr_smoke_shelter_operator"
) {
  throw new Error("shelter occupancy update smoke returned invalid payload");
}
console.log("shelter occupancy update OK");

const invalid = await fetch(`${baseURL}/shelters/nearby?lat=91&lng=-0.2000`);
if (invalid.status !== 400) {
  throw new Error(
    `invalid-coordinate shelter smoke expected 400, got ${invalid.status}`,
  );
}
console.log("invalid-coordinate shelter OK 400");
