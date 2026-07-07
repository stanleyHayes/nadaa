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

const reliefList = await fetch(`${baseURL}/relief-points?status=open`);
if (!reliefList.ok) {
  throw new Error(`relief point list smoke failed: ${reliefList.status}`);
}
const reliefListPayload = await reliefList.json();
if (
  !Array.isArray(reliefListPayload.reliefPoints) ||
  reliefListPayload.reliefPoints.length < 2 ||
  reliefListPayload.reliefPoints.some((point) => point.status !== "open")
) {
  throw new Error("relief point list smoke expected open relief points");
}
console.log(`relief point list OK ${reliefListPayload.reliefPoints.length}`);

const nearbyRelief = await fetch(
  `${baseURL}/relief-points/nearby?lat=5.5600&lng=-0.2000`,
);
if (!nearbyRelief.ok) {
  throw new Error(`nearby relief point smoke failed: ${nearbyRelief.status}`);
}
const nearbyReliefPayload = await nearbyRelief.json();
if (
  !Array.isArray(nearbyReliefPayload.reliefPoints) ||
  nearbyReliefPayload.reliefPoints.length < 2 ||
  nearbyReliefPayload.reliefPoints[0].distanceMeters >
    nearbyReliefPayload.reliefPoints[1].distanceMeters
) {
  throw new Error("nearby relief point smoke expected distance-sorted points");
}
console.log(
  `nearby relief points OK ${nearbyReliefPayload.reliefPoints.length}`,
);

const missingReliefAuthority = await fetch(`${baseURL}/relief-points`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    name: "Unauthorized Relief Point",
    type: "mixed",
    location: { lat: 5.56, lng: -0.2 },
  }),
});
if (missingReliefAuthority.status !== 401) {
  throw new Error(
    `relief authority smoke expected 401, got ${missingReliefAuthority.status}`,
  );
}
console.log("relief authority gate OK 401");

const reliefCreate = await fetch(`${baseURL}/relief-points`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify({
    address: "Smoke relief desk, Accra",
    contact: "112",
    district: "Accra Metropolitan",
    eligibility: "Smoke-test affected households with district registration.",
    location: { lat: 5.55, lng: -0.19 },
    name: "Smoke Relief Point",
    operatingHours: "08:00-18:00",
    region: "Greater Accra",
    schedule: "Daily smoke distribution.",
    source: "manual",
    sourceRef: "smoke-shelter",
    status: "open",
    stockCategories: [
      {
        category: "rice_kg",
        quantity: 120,
        unit: "kg",
      },
    ],
    type: "mixed",
  }),
});
if (!reliefCreate.ok) {
  throw new Error(`relief point create smoke failed: ${reliefCreate.status}`);
}
const reliefCreatePayload = await reliefCreate.json();
if (
  !reliefCreatePayload.id ||
  reliefCreatePayload.status !== "open" ||
  reliefCreatePayload.createdBy !== "usr_smoke_shelter_operator"
) {
  throw new Error("relief point create smoke returned invalid payload");
}
console.log(`relief point create OK ${reliefCreatePayload.id}`);

const reliefUpdate = await fetch(
  `${baseURL}/relief-points/${reliefCreatePayload.id}`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({
      schedule: "Morning and evening smoke distribution.",
      sourceRef: "smoke-shelter",
      status: "limited",
      stockCategories: [
        {
          category: "rice_kg",
          quantity: 90,
          unit: "kg",
        },
        {
          category: "water_sachets",
          quantity: 500,
          unit: "sachets",
        },
      ],
    }),
  },
);
if (!reliefUpdate.ok) {
  throw new Error(`relief point update smoke failed: ${reliefUpdate.status}`);
}
const reliefUpdatePayload = await reliefUpdate.json();
if (
  reliefUpdatePayload.status !== "limited" ||
  reliefUpdatePayload.updatedBy !== "usr_smoke_shelter_operator" ||
  reliefUpdatePayload.stockCategories?.length !== 2
) {
  throw new Error("relief point update smoke returned invalid payload");
}
console.log("relief point update OK");

const reliefHistory = await fetch(
  `${baseURL}/relief-points/${reliefCreatePayload.id}/stock-history`,
);
if (!reliefHistory.ok) {
  throw new Error(
    `relief point stock history smoke failed: ${reliefHistory.status}`,
  );
}
const reliefHistoryPayload = await reliefHistory.json();
if (
  reliefHistoryPayload.reliefPointId !== reliefCreatePayload.id ||
  !reliefHistoryPayload.history?.some(
    (entry) => entry.changedBy === "usr_smoke_shelter_operator",
  )
) {
  throw new Error("relief point stock history smoke returned invalid payload");
}
console.log(
  `relief point stock history OK ${reliefHistoryPayload.history.length}`,
);

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

const hospitalCapacity = await fetch(
  `${baseURL}/hospitals/capacity?lat=5.5600&lng=-0.2000&service=emergency&includeStale=true`,
);
if (!hospitalCapacity.ok) {
  throw new Error(`hospital capacity smoke failed: ${hospitalCapacity.status}`);
}
const hospitalCapacityPayload = await hospitalCapacity.json();
if (
  !Array.isArray(hospitalCapacityPayload.facilities) ||
  hospitalCapacityPayload.facilities.length < 2 ||
  !hospitalCapacityPayload.facilities.some((facility) => facility.stale)
) {
  throw new Error(
    "hospital capacity smoke expected emergency facilities with stale markers",
  );
}
if (
  hospitalCapacityPayload.facilities[0].distanceMeters >
  hospitalCapacityPayload.facilities[1].distanceMeters
) {
  throw new Error(
    "hospital capacity smoke expected distance-sorted facilities",
  );
}
console.log(
  `hospital capacity list OK ${hospitalCapacityPayload.facilities.length}`,
);

const hospitalUpdateTarget =
  hospitalCapacityPayload.facilities.find(
    (facility) => facility.id === "hospital_001",
  ) ?? hospitalCapacityPayload.facilities[0];
const hospitalUpdate = await fetch(
  `${baseURL}/hospitals/${hospitalUpdateTarget.id}/capacity`,
  {
    method: "PATCH",
    headers: authorityHeaders,
    body: JSON.stringify({
      ambulancesAvailable: 2,
      availableBeds: 37,
      emergencyCapacity: "available",
      emergencyUnitStatus: "open",
      icuBedsAvailable: 3,
      notes: "Smoke update confirms hospital capacity workflow.",
      oxygenAvailable: true,
      source: "manual",
      sourceRef: "smoke-shelter",
    }),
  },
);
if (!hospitalUpdate.ok) {
  throw new Error(
    `hospital capacity update smoke failed: ${hospitalUpdate.status}`,
  );
}
const hospitalUpdatePayload = await hospitalUpdate.json();
if (
  hospitalUpdatePayload.facility?.availableBeds !== 37 ||
  hospitalUpdatePayload.facility?.updatedBy !== "usr_smoke_shelter_operator" ||
  hospitalUpdatePayload.facility?.source !== "manual"
) {
  throw new Error("hospital capacity update smoke returned invalid payload");
}
console.log("hospital capacity update OK");

const fixtureImport = await fetch(
  `${baseURL}/hospitals/capacity/imports/fixture`,
  {
    method: "POST",
    headers: authorityHeaders,
    body: JSON.stringify({}),
  },
);
if (!fixtureImport.ok) {
  throw new Error(
    `hospital capacity fixture import smoke failed: ${fixtureImport.status}`,
  );
}
const fixtureImportPayload = await fixtureImport.json();
if (
  fixtureImportPayload.imported < 2 ||
  fixtureImportPayload.source !== "fixture_adapter" ||
  !fixtureImportPayload.facilities?.every(
    (facility) => facility.source === "fixture_adapter",
  )
) {
  throw new Error("hospital capacity fixture import returned invalid payload");
}
console.log(
  `hospital capacity fixture import OK ${fixtureImportPayload.imported}`,
);

const filteredHospitalCapacity = await fetch(
  `${baseURL}/hospitals/capacity?includeStale=false&minAvailableBeds=20`,
);
if (!filteredHospitalCapacity.ok) {
  throw new Error(
    `hospital capacity filtered smoke failed: ${filteredHospitalCapacity.status}`,
  );
}
const filteredHospitalPayload = await filteredHospitalCapacity.json();
if (
  !filteredHospitalPayload.facilities?.length ||
  filteredHospitalPayload.facilities.some((facility) => facility.stale) ||
  filteredHospitalPayload.facilities.some(
    (facility) => facility.availableBeds < 20,
  )
) {
  throw new Error(
    "hospital capacity filtered smoke expected fresh facilities with enough beds",
  );
}
console.log(
  `hospital capacity filtered OK ${filteredHospitalPayload.facilities.length}`,
);

const invalid = await fetch(`${baseURL}/shelters/nearby?lat=91&lng=-0.2000`);
if (invalid.status !== 400) {
  throw new Error(
    `invalid-coordinate shelter smoke expected 400, got ${invalid.status}`,
  );
}
console.log("invalid-coordinate shelter OK 400");
