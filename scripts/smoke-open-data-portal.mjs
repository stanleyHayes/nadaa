const baseURL =
  process.env.OPEN_DATA_SERVICE_URL?.trim() || "http://127.0.0.1:8102/api/v1";

const adminHeaders = {
  "Content-Type": "application/json",
  "X-NADAA-Actor-ID": "usr_smoke_admin",
  "X-NADAA-Agency-ID": "00000000-0000-0000-0000-000000000101",
  "X-NADAA-Actor-Role": "system_admin",
  "X-NADAA-MFA-Completed": "true",
  "X-NADAA-Request-ID": "smoke-open-data",
};

const health = await fetch(`${baseURL.replace("/api/v1", "")}/healthz`);
if (!health.ok) {
  throw new Error(`open-data-service health smoke failed: ${health.status}`);
}
const healthPayload = await health.json();
if (
  healthPayload.status !== "ok" ||
  healthPayload.service !== "open-data-service"
) {
  throw new Error("open-data-service health smoke returned invalid payload");
}
console.log("open-data-service health OK");

const catalog = await fetch(`${baseURL}/open-data/datasets`);
if (!catalog.ok) {
  throw new Error(`open-data-service catalog smoke failed: ${catalog.status}`);
}
const catalogPayload = await catalog.json();
if (
  !Array.isArray(catalogPayload.datasets) ||
  catalogPayload.datasets.length === 0
) {
  throw new Error("open-data-service catalog smoke returned invalid payload");
}
const approvedDataset = catalogPayload.datasets.find(
  (dataset) => dataset.privacyReviewStatus === "approved",
);
if (!approvedDataset) {
  throw new Error(
    "open-data-service catalog smoke returned no approved datasets",
  );
}
console.log(
  `open-data-service catalog OK ${catalogPayload.datasets.length} datasets`,
);

const detail = await fetch(
  `${baseURL}/open-data/datasets/${approvedDataset.id}`,
);
if (!detail.ok) {
  throw new Error(`open-data-service detail smoke failed: ${detail.status}`);
}
const detailPayload = await detail.json();
if (
  typeof detailPayload.dataset !== "object" ||
  detailPayload.dataset.id !== approvedDataset.id
) {
  throw new Error("open-data-service detail smoke returned invalid payload");
}
console.log(`open-data-service detail OK ${approvedDataset.id}`);

const restrictedDataset = catalogPayload.datasets.find(
  (dataset) => dataset.privacyReviewStatus === "pending_review",
);
if (restrictedDataset) {
  const restrictedDownload = await fetch(
    `${baseURL}/open-data/datasets/${restrictedDataset.id}/download`,
  );
  if (restrictedDownload.status !== 403) {
    throw new Error(
      `open-data-service restricted download smoke expected 403, got ${restrictedDownload.status}`,
    );
  }
  console.log(
    `open-data-service restricted download OK 403 ${restrictedDataset.id}`,
  );
}

const download = await fetch(
  `${baseURL}/open-data/datasets/${approvedDataset.id}/download?format=csv`,
);
if (!download.ok) {
  throw new Error(
    `open-data-service download smoke failed: ${download.status}`,
  );
}
const contentDisposition = download.headers.get("content-disposition") ?? "";
if (
  !contentDisposition.includes("attachment") ||
  !contentDisposition.includes(`${approvedDataset.id}.csv`)
) {
  throw new Error(
    "open-data-service download smoke returned no attachment disposition",
  );
}
const csvBody = await download.text();
const csvLines = csvBody.trim().split("\n");
if (csvLines.length < 2 || csvLines[0].length === 0) {
  throw new Error(
    "open-data-service download smoke returned no real CSV bytes",
  );
}
if (download.headers.get("x-nadaa-audit-logged") !== "true") {
  throw new Error(
    "open-data-service download smoke expected X-NADAA-Audit-Logged true",
  );
}
if (
  Number.isNaN(Number(download.headers.get("x-ratelimit-remaining") ?? ""))
) {
  throw new Error(
    "open-data-service download smoke returned no rate limit headers",
  );
}
console.log(
  `open-data-service download OK ${approvedDataset.id}.csv ${csvLines.length - 1} rows`,
);

const requestBody = {
  datasetId: restrictedDataset ? restrictedDataset.id : approvedDataset.id,
  requesterInfo: {
    name: "Smoke Test",
    organization: "NADAA QA",
    email: "smoke@nadaa.gov.gh",
    useCase: "automated smoke testing",
  },
  purpose: "Automated smoke test verifying the open data access request flow.",
};
const requestAccess = await fetch(`${baseURL}/open-data/requests`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(requestBody),
});
if (!requestAccess.ok) {
  throw new Error(
    `open-data-service request smoke failed: ${requestAccess.status}`,
  );
}
const requestPayload = await requestAccess.json();
if (
  typeof requestPayload.request !== "object" ||
  requestPayload.request.status !== "pending"
) {
  throw new Error("open-data-service request smoke returned invalid payload");
}
console.log(`open-data-service request OK ${requestPayload.request.id}`);

const adminNoAuth = await fetch(`${baseURL}/open-data/requests`);
if (adminNoAuth.status !== 401) {
  throw new Error(
    `open-data-service admin without auth expected 401, got ${adminNoAuth.status}`,
  );
}
console.log("open-data-service admin auth gate OK 401");

const adminList = await fetch(`${baseURL}/open-data/requests`, {
  headers: adminHeaders,
});
if (!adminList.ok) {
  throw new Error(
    `open-data-service admin list smoke failed: ${adminList.status}`,
  );
}
const adminListPayload = await adminList.json();
if (!Array.isArray(adminListPayload.requests)) {
  throw new Error(
    "open-data-service admin list smoke returned invalid payload",
  );
}
console.log(
  `open-data-service admin list OK ${adminListPayload.requests.length} requests`,
);

const review = await fetch(
  `${baseURL}/open-data/requests/${requestPayload.request.id}/approve`,
  {
    method: "POST",
    headers: adminHeaders,
    body: JSON.stringify({
      reviewer: "smoke@nadaa.gov.gh",
      approved: true,
      note: "Approved by smoke test.",
    }),
  },
);
if (!review.ok) {
  throw new Error(`open-data-service review smoke failed: ${review.status}`);
}
const reviewPayload = await review.json();
if (reviewPayload.request.status !== "approved") {
  throw new Error("open-data-service review smoke returned invalid payload");
}
console.log(`open-data-service review OK ${reviewPayload.request.status}`);

const reReview = await fetch(
  `${baseURL}/open-data/requests/${requestPayload.request.id}/approve`,
  {
    method: "POST",
    headers: adminHeaders,
    body: JSON.stringify({ approved: false, note: "Attempted re-review." }),
  },
);
if (reReview.status !== 409) {
  throw new Error(
    `open-data-service re-review smoke expected 409, got ${reReview.status}`,
  );
}
console.log("open-data-service re-review conflict OK 409");
