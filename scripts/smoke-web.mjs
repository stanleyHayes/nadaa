const targets = [
  [
    "marketing-web",
    localURL("LOCAL_MARKETING_URL", "http://127.0.0.1:5172/"),
    "NADAA Marketing",
  ],
  [
    "citizen-web",
    localURL("LOCAL_CITIZEN_URL", "http://127.0.0.1:5173/"),
    "NADAA Citizen",
  ],
  [
    "authority-dashboard",
    localURL("LOCAL_AUTHORITY_URL", "http://127.0.0.1:5174/"),
    "NADAA Authority Dashboard",
  ],
  [
    "dispatcher-web",
    localURL("LOCAL_DISPATCHER_URL", "http://127.0.0.1:5175/"),
    "NADAA Dispatch Command",
  ],
  [
    "admin-web",
    localURL("LOCAL_ADMIN_URL", "http://127.0.0.1:5176/"),
    "NADAA Admin Console",
  ],
  [
    "agency-web",
    localURL("LOCAL_AGENCY_URL", "http://127.0.0.1:5177/"),
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
  if (!html.includes(`<title>${expectedTitle}</title>`)) {
    throw new Error(`${name} smoke check reached the wrong app at ${url}`);
  }

  console.log(`${name} OK ${response.status}`);
}

function localURL(envKey, fallback) {
  const value = process.env[envKey]?.trim();
  return value || fallback;
}
