const targets = [
  ["citizen-web", "http://127.0.0.1:5173/", "NADAA Citizen"],
  ["authority-dashboard", "http://127.0.0.1:5174/", "NADAA Authority Dashboard"]
];

for (const [name, url, expectedTitle] of targets) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`${name} smoke check failed: ${response.status} ${response.statusText}`);
  }

  const html = await response.text();
  if (!html.includes(`<title>${expectedTitle}</title>`)) {
    throw new Error(`${name} smoke check reached the wrong app at ${url}`);
  }

  console.log(`${name} OK ${response.status}`);
}
