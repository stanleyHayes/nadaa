const url =
  process.env.MARKETING_URL?.trim() ||
  process.env.LOCAL_MARKETING_URL?.trim() ||
  "http://127.0.0.1:5200/";

const response = await fetch(url);
if (!response.ok) {
  throw new Error(
    `marketing smoke failed: ${response.status} ${response.statusText}`,
  );
}

const html = await response.text();

// The title carries an SEO suffix after the stable "NADAA —" prefix.
if (!html.includes("<title>NADAA —")) {
  throw new Error("marketing smoke reached the wrong page");
}

if (!html.includes('id="root"') || !html.includes("/src/main.tsx")) {
  throw new Error("marketing smoke missing Vite app shell");
}

console.log(`marketing OK ${response.status}`);
