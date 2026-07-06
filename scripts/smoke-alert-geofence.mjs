const baseURL =
  process.env.ALERT_API_URL?.trim() || "http://127.0.0.1:8089/api/v1";

const drafterHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_drafter",
  "X-NADAA-Actor-Role": "district_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-alert-geofence-draft",
};

const approverHeaders = {
  ...drafterHeaders,
  "X-NADAA-Actor-ID": "usr_smoke_approver",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Request-ID": "smoke-alert-geofence-approve",
};

const now = Date.now();
const targetId = `accra-central-radius-${now}`;

const preview = await fetch(`${baseURL}/alerts/targets/preview`, {
  method: "POST",
  headers: drafterHeaders,
  body: JSON.stringify({
    type: "district",
    ids: ["accra-metropolitan"],
    label: "Accra Metropolitan",
  }),
});
if (!preview.ok) {
  throw new Error(`alert target preview failed: ${preview.status}`);
}
const previewPayload = await preview.json();
if (
  !previewPayload.target?.geometry ||
  !previewPayload.target?.center ||
  !previewPayload.target?.areaSqKm
) {
  throw new Error("alert target preview did not return geometry metadata");
}
console.log("alert target preview OK district geometry");

const body = {
  title: "Smoke radius flood warning",
  hazardType: "flood",
  severity: "warning",
  message: "Smoke test warning for flood-prone roads near Accra Central.",
  target: {
    type: "radius",
    ids: [targetId],
    label: "Accra Central 5km radius",
    center: { lat: 5.556, lng: -0.202 },
    radiusMeters: 5000,
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
  throw new Error(`alert radius create expected 201, got ${create.status}`);
}
const draft = await create.json();
if (
  draft.target?.type !== "radius" ||
  !draft.target?.center ||
  draft.target?.radiusMeters !== 5000 ||
  !draft.target?.areaSqKm
) {
  throw new Error("alert radius create did not store target metadata");
}
console.log(`alert radius create OK ${draft.id}`);

const override = await fetch(
  `${baseURL}/alerts/${draft.id}/emergency-override`,
  {
    method: "POST",
    headers: approverHeaders,
    body: JSON.stringify({ reason: "Smoke test immediate geofenced warning." }),
  },
);
if (!override.ok) {
  throw new Error(`alert radius override failed: ${override.status}`);
}
console.log("alert radius override OK approved");

const filtered = await fetch(
  `${baseURL}/alerts?current=true&targetType=radius&targetId=${targetId}`,
);
if (!filtered.ok) {
  throw new Error(`alert radius target query failed: ${filtered.status}`);
}
const filteredPayload = await filtered.json();
if (
  !Array.isArray(filteredPayload.alerts) ||
  !filteredPayload.alerts.some((alert) => alert.id === draft.id)
) {
  throw new Error(
    "alert radius target query did not return the approved alert",
  );
}
console.log("alert radius query OK targetType targetId");

const invalidCustom = await fetch(`${baseURL}/alerts`, {
  method: "POST",
  headers: drafterHeaders,
  body: JSON.stringify({
    ...body,
    target: {
      type: "custom",
      ids: ["bad-polygon"],
      label: "Bad Polygon",
      geometry: {
        type: "Polygon",
        coordinates: [
          [
            [-0.22, 5.55],
            [-0.18, 5.55],
            [-0.18, 5.59],
          ],
        ],
      },
    },
  }),
});
if (invalidCustom.status !== 400) {
  throw new Error(
    `alert invalid custom geometry expected 400, got ${invalidCustom.status}`,
  );
}
console.log("alert custom geometry gate OK 400");
