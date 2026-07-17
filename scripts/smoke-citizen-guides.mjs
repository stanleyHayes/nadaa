const guideURL =
  process.env.GUIDE_API_URL?.trim() || "http://127.0.0.1:8086/api/v1";
const citizenURL =
  process.env.CITIZEN_WEB_URL?.trim() || "http://127.0.0.1:5201";

const offlineGuides = await fetch(
  `${guideURL}/guides?offline=true&language=en`,
);
if (!offlineGuides.ok) {
  throw new Error(
    `offline guide feed failed: ${offlineGuides.status} ${offlineGuides.statusText}`,
  );
}

const offlinePayload = await offlineGuides.json();
if (
  !Array.isArray(offlinePayload.guides) ||
  offlinePayload.guides.length < 5 ||
  !offlinePayload.guides.every((guide) => guide.offlineAvailable)
) {
  throw new Error("offline guide feed did not return cached guide candidates");
}

if (!offlinePayload.guides.some((guide) => guide.title.includes("112"))) {
  throw new Error("offline guide feed must include visible 112 guidance");
}
console.log(`offline guide feed OK ${offlinePayload.guides.length}`);

const languageFallback = await fetch(
  `${guideURL}/guides?hazard=fire&stage=during&language=ga`,
);
if (!languageFallback.ok) {
  throw new Error(
    `guide language fallback failed: ${languageFallback.status} ${languageFallback.statusText}`,
  );
}
const fallbackPayload = await languageFallback.json();
if (
  !Array.isArray(fallbackPayload.guides) ||
  fallbackPayload.guides.length !== 1 ||
  fallbackPayload.guides[0].language !== "en"
) {
  throw new Error("guide language fallback did not return English guidance");
}
console.log("guide language fallback OK");

const app = await fetch(`${citizenURL}/`);
if (!app.ok) {
  throw new Error(`citizen app failed: ${app.status} ${app.statusText}`);
}
const html = await app.text();
// The title carries an SEO suffix after the stable "NADAA Citizen" prefix.
if (!html.includes("<title>NADAA Citizen")) {
  throw new Error("citizen guide smoke reached the wrong web app");
}

const serviceWorker = await fetch(`${citizenURL}/sw.js`);
if (!serviceWorker.ok) {
  throw new Error(
    `citizen service worker failed: ${serviceWorker.status} ${serviceWorker.statusText}`,
  );
}
const serviceWorkerSource = await serviceWorker.text();
if (!serviceWorkerSource.includes("nadaa-citizen-guides")) {
  throw new Error("citizen service worker is missing guide cache logic");
}
console.log("citizen guide UI shell OK");
