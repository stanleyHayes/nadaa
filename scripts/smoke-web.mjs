const targets = [
  ["citizen-web", "http://localhost:5173/"],
  ["authority-dashboard", "http://localhost:5174/"]
];

for (const [name, url] of targets) {
  const response = await fetch(url, { method: "HEAD" });
  if (!response.ok) {
    throw new Error(`${name} smoke check failed: ${response.status} ${response.statusText}`);
  }

  console.log(`${name} OK ${response.status}`);
}

