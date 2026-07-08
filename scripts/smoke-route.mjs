const baseURL =
  process.env.ROUTE_SERVICE_URL?.trim() || "http://127.0.0.1:8096";

const health = await fetch(`${baseURL}/health`);
if (!health.ok) {
  throw new Error(`route-service health smoke failed: ${health.status}`);
}
const healthPayload = await health.json();
if (
  healthPayload.status !== "ok" ||
  healthPayload.service !== "route-service"
) {
  throw new Error("route-service health smoke returned invalid payload");
}
console.log("route-service health OK");

const options = await fetch(`${baseURL}/routes/options`);
if (!options.ok) {
  throw new Error(`route-service options smoke failed: ${options.status}`);
}
const optionsPayload = await options.json();
if (
  !Array.isArray(optionsPayload.waypointTypes) ||
  !optionsPayload.waypointTypes.includes("shelter")
) {
  throw new Error("route-service options smoke returned invalid waypointTypes");
}
console.log(
  `route-service options OK ${optionsPayload.waypointTypes.join(", ")}`,
);

const plan = await fetch(`${baseURL}/routes/plan`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    origin: { lat: 5.56, lng: -0.2 },
    waypointType: "higher_ground",
    avoidRiskLevels: ["severe", "emergency"],
    closureBufferMeters: 100,
  }),
});
if (!plan.ok) {
  throw new Error(`route-service plan smoke failed: ${plan.status}`);
}
const planPayload = await plan.json();
if (
  !Array.isArray(planPayload.route) ||
  planPayload.route.length < 2 ||
  typeof planPayload.distanceMeters !== "number" ||
  typeof planPayload.estimatedDurationMinutes !== "number" ||
  !Array.isArray(planPayload.segments) ||
  typeof planPayload.disclaimer !== "string" ||
  planPayload.decisionSupport !== true
) {
  throw new Error("route-service plan smoke returned invalid payload");
}
console.log(
  `route-service plan OK ${planPayload.distanceMeters}m ${planPayload.estimatedDurationMinutes}min`,
);

const manualPlan = await fetch(`${baseURL}/routes/plan`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    origin: { lat: 5.56, lng: -0.2 },
    destination: { lat: 5.58, lng: -0.18 },
    waypointType: "manual",
  }),
});
if (!manualPlan.ok) {
  throw new Error(
    `route-service manual plan smoke failed: ${manualPlan.status}`,
  );
}
const manualPlanPayload = await manualPlan.json();
if (
  !Array.isArray(manualPlanPayload.route) ||
  manualPlanPayload.route.length < 2
) {
  throw new Error("route-service manual plan smoke returned invalid payload");
}
console.log(
  `route-service manual plan OK ${manualPlanPayload.distanceMeters}m`,
);

const invalidPlan = await fetch(`${baseURL}/routes/plan`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    origin: { lat: 91, lng: -0.2 },
    waypointType: "manual",
  }),
});
if (invalidPlan.status !== 400) {
  throw new Error(
    `route-service invalid plan smoke expected 400, got ${invalidPlan.status}`,
  );
}
console.log("route-service invalid plan OK 400");
