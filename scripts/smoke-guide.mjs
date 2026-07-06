const baseURL =
  process.env.GUIDE_API_URL?.trim() || "http://127.0.0.1:8086/api/v1";

const checks = [
  {
    name: "flood before English",
    path: "/guides?hazard=flood&stage=before&language=en",
    expected: { hazardType: "flood", stage: "before", language: "en" },
  },
  {
    name: "fire language fallback",
    path: "/guides?hazard=fire&stage=during&language=ga",
    expected: { hazardType: "fire", stage: "during", language: "en" },
  },
  {
    name: "offline guides",
    path: "/guides?offline=true",
    expected: { offlineAvailable: true },
  },
];

for (const check of checks) {
  const response = await fetch(`${baseURL}${check.path}`);
  if (!response.ok) {
    throw new Error(
      `${check.name} guide smoke failed: ${response.status} ${response.statusText}`,
    );
  }

  const payload = await response.json();
  if (!Array.isArray(payload.guides) || payload.guides.length === 0) {
    throw new Error(`${check.name} guide smoke returned no guides`);
  }

  for (const [field, expected] of Object.entries(check.expected)) {
    const matches = payload.guides.every((guide) => guide[field] === expected);
    if (!matches) {
      throw new Error(
        `${check.name} guide smoke expected ${field}=${expected}`,
      );
    }
  }

  console.log(`${check.name} guide OK ${payload.guides.length}`);
}

const invalid = await fetch(`${baseURL}/guides?stage=panic`);
if (invalid.status !== 400) {
  throw new Error(
    `invalid guide stage smoke expected 400, got ${invalid.status}`,
  );
}
console.log("invalid-stage guide OK 400");
