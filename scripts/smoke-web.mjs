const targets = [
  [
    "marketing-web",
    localURL("LOCAL_MARKETING_URL", "http://127.0.0.1:5200/"),
    "NADAA —",
  ],
  [
    "citizen-web",
    localURL("LOCAL_CITIZEN_URL", "http://127.0.0.1:5201/"),
    "NADAA Citizen",
  ],
  [
    "authority-dashboard",
    localURL("LOCAL_AUTHORITY_URL", "http://127.0.0.1:5202/"),
    "NADAA Authority Dashboard",
  ],
  [
    "dispatcher-web",
    localURL("LOCAL_DISPATCHER_URL", "http://127.0.0.1:5203/"),
    "NADAA Dispatch Command",
  ],
  [
    "admin-web",
    localURL("LOCAL_ADMIN_URL", "http://127.0.0.1:5204/"),
    "NADAA Admin Console",
  ],
  [
    "agency-web",
    localURL("LOCAL_AGENCY_URL", "http://127.0.0.1:5205/"),
    "NADAA Agency Operations",
  ],
];

for (const [name, url, expectedTitle] of targets) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(
      `${name} smoke check failed: ${response.status} ${response.statusText}`,
    );
  }

  const html = await response.text();
  // Match a stable title prefix only: apps append SEO suffixes to <title>.
  if (!html.includes(`<title>${expectedTitle}`)) {
    throw new Error(`${name} smoke check reached the wrong app at ${url}`);
  }

  console.log(`${name} OK ${response.status}`);
}

function localURL(envKey, fallback) {
  const value = process.env[envKey]?.trim();
  return value || fallback;
}
