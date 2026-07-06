const baseURL =
  process.env.RISK_API_URL?.trim() || "http://127.0.0.1:8081/api/v1";

const checks = [
  ["severe", "lat=5.5600&lng=-0.2000", "severe"],
  ["high", "lat=5.6037&lng=-0.1870", "high"],
  ["low", "lat=6.6885&lng=-1.6244", "low"],
];

for (const [name, query, expectedRisk] of checks) {
  const response = await fetch(`${baseURL}/risk?${query}`);
  if (!response.ok) {
    throw new Error(
      `${name} risk smoke failed: ${response.status} ${response.statusText}`,
    );
  }

  const payload = await response.json();
  if (payload.overallRisk !== expectedRisk) {
    throw new Error(
      `${name} risk smoke expected ${expectedRisk}, got ${payload.overallRisk}`,
    );
  }
  if (!Array.isArray(payload.risks) || payload.risks.length === 0) {
    throw new Error(`${name} risk smoke returned no hazard risks`);
  }
  if (
    !Array.isArray(payload.nearestShelters) ||
    !Array.isArray(payload.nearbyFacilities)
  ) {
    throw new Error(
      `${name} risk smoke returned an incomplete resource payload`,
    );
  }

  console.log(`${name} risk OK ${payload.overallRisk}`);
}

const invalid = await fetch(`${baseURL}/risk?lat=91&lng=-0.1870`);
if (invalid.status !== 400) {
  throw new Error(
    `invalid coordinate smoke expected 400, got ${invalid.status}`,
  );
}
console.log("invalid-coordinate risk OK 400");
