const baseURL =
  process.env.SIMULATION_API_URL?.trim() || "http://127.0.0.1:8094";

const health = await fetch(`${baseURL}/healthz`);
if (!health.ok) {
  throw new Error(`ml-service health smoke failed: ${health.status}`);
}
console.log("ml-service health OK");

const create = await fetch(`${baseURL}/api/v1/ml/flood/simulations`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    name: "Smoke flood simulation",
    rainfallMmOverride: 40,
    waterLevelTrendCmOverride: 12,
    durationHours: 3,
    timeStepHours: 1,
  }),
});
if (!create.ok) {
  throw new Error(`flood simulation create smoke failed: ${create.status}`);
}
const created = await create.json();
if (!created.simulation?.id || created.simulation.status !== "completed") {
  throw new Error("flood simulation create expected completed job");
}
if (created.simulation.frames.length !== 3) {
  throw new Error(
    `flood simulation expected 3 frames got ${created.simulation.frames.length}`,
  );
}
if (
  created.simulation.safety.humanReviewRequired !== true ||
  created.simulation.safety.autoPublishAllowed !== false
) {
  throw new Error("flood simulation expected restrictive safety policy");
}
console.log(
  `flood simulation create OK ${created.simulation.reference} (${created.simulation.frames.length} frames)`,
);

const list = await fetch(`${baseURL}/api/v1/ml/flood/simulations`);
if (!list.ok) {
  throw new Error(`flood simulation list smoke failed: ${list.status}`);
}
const listPayload = await list.json();
if (!Array.isArray(listPayload.simulations)) {
  throw new Error("flood simulation list expected simulations array");
}
console.log(`flood simulation list OK ${listPayload.simulations.length}`);

const get = await fetch(
  `${baseURL}/api/v1/ml/flood/simulations/${created.simulation.id}`,
);
if (!get.ok) {
  throw new Error(`flood simulation get smoke failed: ${get.status}`);
}
const got = await get.json();
if (got.simulation.id !== created.simulation.id) {
  throw new Error("flood simulation get returned wrong job");
}
console.log("flood simulation get OK");

const invalid = await fetch(`${baseURL}/api/v1/ml/flood/simulations`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ durationHours: 2 }),
});
if (invalid.status !== 400) {
  throw new Error(
    `flood simulation invalid request expected 400 got ${invalid.status}`,
  );
}
console.log("flood simulation invalid request OK 400");
