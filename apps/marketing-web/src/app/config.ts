// Sibling app URLs are env-driven so the deployed marketing site can link to
// the production frontends (set VITE_*_URL in the Vercel project), while
// falling back to the local dev ports when unset.
export const marketingLinks = {
  citizenWeb: import.meta.env.VITE_CITIZEN_WEB_URL ?? "http://localhost:5201/",
  dispatcherWeb:
    import.meta.env.VITE_DISPATCHER_WEB_URL ?? "http://localhost:5203/",
  agencyWeb: import.meta.env.VITE_AGENCY_WEB_URL ?? "http://localhost:5205/",
  adminWeb: import.meta.env.VITE_ADMIN_WEB_URL ?? "http://localhost:5204/",
  emergencyPhone: "tel:112",
  partnerMail:
    "mailto:partnerships@nadaa.gov.gh?subject=NADAA%20partnership%20request",
} as const;
