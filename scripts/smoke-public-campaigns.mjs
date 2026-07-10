const baseURL =
  process.env.CAMPAIGN_API_URL?.trim() || "http://127.0.0.1:8103/api/v1";

const authorityHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_test",
  "X-NADAA-Actor-Role": "nadmo_officer",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": `smoke-${Date.now()}`,
};

const health = await fetch(`${baseURL.replace("/api/v1", "")}/healthz`);
if (!health.ok) {
  throw new Error(`campaign-service health smoke failed: ${health.status}`);
}
const healthPayload = await health.json();
if (
  healthPayload.status !== "ok" ||
  healthPayload.service !== "campaign-service"
) {
  throw new Error("campaign-service health smoke returned invalid payload");
}
console.log("campaign-service health OK");

const templatesResponse = await fetch(`${baseURL}/campaign-templates`);
if (!templatesResponse.ok) {
  throw new Error(
    `campaign-templates smoke failed: ${templatesResponse.status}`,
  );
}
const templatesPayload = await templatesResponse.json();
if (
  !Array.isArray(templatesPayload.templates) ||
  !templatesPayload.templates.length
) {
  throw new Error("campaign-templates smoke returned no templates");
}
console.log(`campaign-templates OK ${templatesPayload.templates.length}`);

const publicList = await fetch(`${baseURL}/campaigns?region=Greater+Accra`);
if (!publicList.ok) {
  throw new Error(`public campaigns list smoke failed: ${publicList.status}`);
}
const publicListPayload = await publicList.json();
if (!Array.isArray(publicListPayload.campaigns)) {
  throw new Error("public campaigns list smoke returned invalid payload");
}
console.log(`public campaigns list OK ${publicListPayload.campaigns.length}`);

const createBody = {
  title: "Smoke test preparedness campaign",
  hazardType: "flood",
  targetRegions: ["Greater Accra"],
  languages: ["en"],
  contentBlocks: [
    {
      type: "article",
      title: "Stay informed",
      body: "Listen to official updates.",
    },
    {
      type: "checklist",
      title: "Readiness checklist",
      items: ["Know your shelter", "Charge devices"],
    },
  ],
  publishWindow: {
    startsAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    endsAt: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
  },
  status: "published",
  linkedGuideIds: ["guide_flood_before_en"],
  linkedAlertIds: ["alert_feed_current_flood"],
};

const createResponse = await fetch(`${baseURL}/campaigns`, {
  method: "POST",
  headers: authorityHeaders,
  body: JSON.stringify(createBody),
});
if (!createResponse.ok) {
  const body = await createResponse.text();
  throw new Error(
    `campaign create smoke failed: ${createResponse.status} ${body}`,
  );
}
const createPayload = await createResponse.json();
if (!createPayload.campaign?.id) {
  throw new Error("campaign create smoke returned invalid payload");
}
const campaignId = createPayload.campaign.id;
console.log(`campaign create OK ${campaignId}`);

const getResponse = await fetch(`${baseURL}/campaigns/${campaignId}`);
if (!getResponse.ok) {
  throw new Error(`campaign get smoke failed: ${getResponse.status}`);
}
const getPayload = await getResponse.json();
if (getPayload.campaign?.id !== campaignId) {
  throw new Error("campaign get smoke returned wrong campaign");
}
console.log("campaign get OK");

const metricsResponse = await fetch(
  `${baseURL}/campaigns/${campaignId}/metrics`,
);
if (!metricsResponse.ok) {
  throw new Error(`campaign metrics smoke failed: ${metricsResponse.status}`);
}
const metricsPayload = await metricsResponse.json();
if (!Array.isArray(metricsPayload.metrics) || !metricsPayload.metrics.length) {
  throw new Error("campaign metrics smoke returned invalid payload");
}
console.log(`campaign metrics OK ${metricsPayload.metrics.length}`);

const updateResponse = await fetch(`${baseURL}/campaigns/${campaignId}`, {
  method: "PUT",
  headers: authorityHeaders,
  body: JSON.stringify({
    title: "Updated smoke test campaign",
    status: "archived",
  }),
});
if (!updateResponse.ok) {
  throw new Error(`campaign update smoke failed: ${updateResponse.status}`);
}
const updatePayload = await updateResponse.json();
if (updatePayload.campaign?.status !== "archived") {
  throw new Error("campaign update smoke did not apply status change");
}
console.log("campaign update OK");

const filteredList = await fetch(
  `${baseURL}/campaigns?hazard=flood&language=en`,
);
if (!filteredList.ok) {
  throw new Error(
    `filtered campaigns list smoke failed: ${filteredList.status}`,
  );
}
const filteredPayload = await filteredList.json();
if (!Array.isArray(filteredPayload.campaigns)) {
  throw new Error("filtered campaigns list smoke returned invalid payload");
}
console.log(`filtered campaigns list OK ${filteredPayload.campaigns.length}`);

const invalidResponse = await fetch(`${baseURL}/campaigns?hazard=unknown`);
if (invalidResponse.status !== 400) {
  throw new Error(
    `invalid campaign hazard smoke expected 400, got ${invalidResponse.status}`,
  );
}
console.log("invalid campaign hazard OK 400");
