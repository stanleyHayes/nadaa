// Sibling app URLs are env-driven so the deployed marketing site can link to
// the production frontends (set VITE_*_URL in the Vercel project, see
// docs/render-vercel-deploy.md). Local dev falls back to the dev ports;
// production deploys must set the real URLs — the live site must never link
// citizens to localhost.
function siblingAppUrl(
  envValue: string | undefined,
  envName: string,
  devFallback: string,
): string {
  if (envValue) {
    return envValue;
  }
  if (import.meta.env.PROD) {
    throw new Error(
      `[nadaa] ${envName} is not set: production builds must define the deployed sibling app URLs (see docs/render-vercel-deploy.md).`,
    );
  }
  return devFallback;
}

export const marketingLinks = {
  citizenWeb: siblingAppUrl(
    import.meta.env.VITE_CITIZEN_WEB_URL,
    "VITE_CITIZEN_WEB_URL",
    "http://localhost:5201/",
  ),
  dispatcherWeb: siblingAppUrl(
    import.meta.env.VITE_DISPATCHER_WEB_URL,
    "VITE_DISPATCHER_WEB_URL",
    "http://localhost:5203/",
  ),
  agencyWeb: siblingAppUrl(
    import.meta.env.VITE_AGENCY_WEB_URL,
    "VITE_AGENCY_WEB_URL",
    "http://localhost:5205/",
  ),
  adminWeb: siblingAppUrl(
    import.meta.env.VITE_ADMIN_WEB_URL,
    "VITE_ADMIN_WEB_URL",
    "http://localhost:5204/",
  ),
  emergencyPhone: "tel:112",
  partnerMail:
    "mailto:partnerships@nadaa.gov.gh?subject=NADAA%20partnership%20request",
} as const;
