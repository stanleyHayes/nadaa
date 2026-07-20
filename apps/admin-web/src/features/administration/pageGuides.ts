import type { ViewId } from "./navigation";

/**
 * Contextual-help registry for the admin console. Keyed by the shell's active
 * view (the eight nav views plus the synthetic "settings" and "guide" views),
 * each entry powers the page-help popover in the topbar and one card in the
 * user guide. Copy is written from a platform-administrator's perspective.
 */
export type GuideKey = ViewId | "settings" | "guide";

export type GuideSection =
  | "Command"
  | "Directory"
  | "Access & policy"
  | "Governance"
  | "Account"
  | "Help & onboarding";

export type PageGuide = {
  key: GuideKey;
  section: GuideSection;
  title: string;
  description: string;
  steps: string[];
};

export const PAGE_GUIDES: PageGuide[] = [
  {
    key: "overview",
    section: "Command",
    title: "Overview",
    description: "Governance posture across agencies, access, and audit.",
    steps: [
      "Scan the four tiles first: registered agencies, MFA coverage, users awaiting MFA, and connected data sources.",
      "Check MFA readiness by agency, and read the recent audit activity to see what changed on the platform.",
      "Use the Governance posture panel and its Jump to links to open the desk that needs attention.",
    ],
  },
  {
    key: "agencies",
    section: "Directory",
    title: "Agencies",
    description: "Registered agencies, operating scope, and MFA coverage.",
    steps: [
      "Review each registered agency, its status, and the operating scope it is authorised for.",
      "Read the MFA coverage bar per agency to find teams that have not enrolled two-step verification.",
      "Follow up with any agency below full coverage before granting new access.",
    ],
  },
  {
    key: "users",
    section: "Directory",
    title: "Users",
    description: "Authority users, roles, and access provisioning.",
    steps: [
      "Search by name or email, or filter by role and MFA state, to find a user in the directory.",
      "Use Create user to provision authority access; new accounts start with MFA setup pending.",
      "Confirm each user's role and agency are correct before they sign in.",
    ],
  },
  {
    key: "roles",
    section: "Access & policy",
    title: "Roles & access",
    description: "Admin, alert, and operational permission matrix.",
    steps: [
      "Read the matrix to see which roles hold admin, alert-approval, and operational permissions.",
      "Confirm that sensitive actions stay limited to the roles your governance policy allows.",
      "Use it as the reference when you assign a role on the Users desk.",
    ],
  },
  {
    key: "mfa",
    section: "Access & policy",
    title: "MFA readiness",
    description: "Two-step verification coverage and users awaiting setup.",
    steps: [
      "Review overall two-step verification coverage and the users still awaiting setup.",
      "Prioritise accounts with elevated roles, since setup must complete before they can sign in.",
      "Follow up with each user's agency to close the remaining gaps.",
    ],
  },
  {
    key: "audit",
    section: "Governance",
    title: "Audit trail",
    description: "Sensitive-action trace across the platform.",
    steps: [
      "Read the trail top-down; the most recent sensitive actions are listed first.",
      "Open an entry to see the actor, the target, and the before-and-after snapshot, with secrets redacted.",
      "Use the trail to verify who changed what when you investigate an incident or access request.",
    ],
  },
  {
    key: "integrations",
    section: "Governance",
    title: "Data sources",
    description: "Integration contracts and safe secret scopes.",
    steps: [
      "Review each integration contract, its status, cadence, and the data it exchanges.",
      "Confirm the secret scope and PII handling match the contract before you rely on the feed.",
      "Flag any contract without a safe scope or a manual fallback for review.",
    ],
  },
  {
    key: "alertRules",
    section: "Governance",
    title: "Alert rules",
    description: "Approval, override, targeting, and audit posture derived from live alerts.",
    steps: [
      "Remember this is a read-only view: each card is derived from a live alert, not an editable rule.",
      "Review each card's approver roles, emergency-override roles, and targeting scope.",
      "Confirm MFA is required and the audit action is set, since every alert stays human-approved.",
    ],
  },
  {
    key: "settings",
    section: "Account",
    title: "Settings",
    description: "Manage your profile, security, notifications, and preferences.",
    steps: [
      "Update your profile details and confirm your agency and role are correct.",
      "Harden your account from Security: enable multi-factor authentication and change your password.",
      "Tune notification channels and preferences, including reduced motion, for this browser.",
    ],
  },
  {
    key: "guide",
    section: "Help & onboarding",
    title: "User guide",
    description: "Step-by-step help for every admin-console workspace.",
    steps: [
      "Browse the cards by section to find the workspace you are working in.",
      "Follow the numbered steps for a page, then use Open page to jump straight there.",
      "Press the help button in any page header, then Listen, to hear the steps read aloud.",
    ],
  },
];

const GUIDE_BY_KEY = Object.fromEntries(
  PAGE_GUIDES.map((guide) => [guide.key, guide]),
) as Record<GuideKey, PageGuide>;

/** Resolve the guide for the shell's current view. */
export function getPageGuide(key: GuideKey): PageGuide {
  return GUIDE_BY_KEY[key];
}

const SECTION_ORDER: GuideSection[] = [
  "Command",
  "Directory",
  "Access & policy",
  "Governance",
  "Account",
  "Help & onboarding",
];

/** Guides grouped by section, in display order, for the user-guide view. */
export function groupedPageGuides(): Array<{
  section: GuideSection;
  guides: PageGuide[];
}> {
  return SECTION_ORDER.map((section) => ({
    section,
    guides: PAGE_GUIDES.filter((guide) => guide.section === section),
  })).filter((group) => group.guides.length > 0);
}
