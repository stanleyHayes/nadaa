/**
 * Planned domain map.
 *
 * NOT YET REGISTERED. usesubvenio.com has not been purchased at the time of
 * writing, so nothing here resolves. It is recorded now so the apex and the
 * subdomain split live in one place rather than being invented per app, and so
 * the cutover is an edit here plus the env vars below.
 *
 * Nothing reads these at runtime in production: every deployed app takes its
 * origin from an environment variable and refuses to start without one (see
 * apps/marketing-web/src/app/config.ts). These are the DEFAULTS and the
 * documentation of intent.
 *
 * The previous origin was nadaa.gov.gh. That is a Ghanaian government domain
 * and would belong to NADMO, not to the registered operator of this product —
 * see the operator note in PrivacyPage.tsx. Any remaining nadaa.gov.gh address
 * should be treated as a placeholder to be cut over, not as a working address.
 */

/** Apex — marketing site and canonical origin. */
export const APEX = "usesubvenio.com";

export const domains = {
  /** Marketing site. */
  site: `https://${APEX}`,
  /** Citizen web application. */
  app: `https://app.${APEX}`,
  /** Administration console. */
  admin: `https://admin.${APEX}`,
  /** Responder portal. Provisional — the responder module is not built yet. */
  respond: `https://respond.${APEX}`,
  /** Documentation and API reference. */
  docs: `https://docs.${APEX}`,
  /** Service status page. */
  status: `https://status.${APEX}`,
} as const;

/**
 * Mailboxes. Also pending the domain purchase — until it completes these do not
 * receive mail, so do not present them to users as live contact routes.
 */
export const mailboxes = {
  partnerships: `partnerships@${APEX}`,
  privacy: `privacy@${APEX}`,
  support: `support@${APEX}`,
} as const;

export type DomainKey = keyof typeof domains;
