import type { ViewId } from "./navigation";

/**
 * Contextual-help registry for the command center. Keyed by the shell's active
 * view (the eight nav views plus the synthetic "settings" and "guide" views),
 * each entry powers the page-help popover in the topbar and one card in the
 * user guide. Copy is written from a duty-officer's perspective.
 */
export type GuideKey = ViewId | "settings" | "guide";

export type GuideSection =
  | "Command"
  | "Response"
  | "Intelligence"
  | "Recovery & readiness"
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
    description: "Live triage across every Greater Accra flood operation.",
    steps: [
      "Scan the four status tiles first: active incidents, pending alerts, shelter occupancy, and teams en route.",
      "Read the live triage board to see how active incidents band from Emergency down to Low.",
      "Use the alert pipeline and shelter cards, or the Jump to links, to open the desk that needs action.",
    ],
  },
  {
    key: "incidents",
    section: "Response",
    title: "Incidents",
    description: "Command map, queue, incident detail, and response routing.",
    steps: [
      "Work the queue top-down; Emergency and severe reports are sorted to the top.",
      "Open an incident to review its report, its location on the command map, and any evidence before you act.",
      "Assign a response team and advance the status as crews move from en route to on scene.",
      "Flag anything that needs a second opinion for priority review.",
    ],
  },
  {
    key: "alerts",
    section: "Response",
    title: "Alerts & broadcast",
    description: "Draft, review, and publish approved public-safety alerts.",
    steps: [
      "Start a draft or open a submitted alert, then check the hazard, target area, and message wording.",
      "Every alert needs human approval; confirm the details are accurate before you approve it.",
      "Publish approved alerts to push them to citizens, and clear stale drafts from the pipeline.",
    ],
  },
  {
    key: "shelters",
    section: "Response",
    title: "Shelters & relief",
    description: "Update shelter capacity and publish relief distribution.",
    steps: [
      "Review each shelter's occupancy against capacity; rows near full are highlighted.",
      "Update the current headcount as people arrive or leave so the public map stays accurate.",
      "Publish relief-distribution points so citizens know where to collect supplies.",
    ],
  },
  {
    key: "forecasting",
    section: "Intelligence",
    title: "Forecasting",
    description: "Resource positioning, flood simulation, and ML review.",
    steps: [
      "Run the flood simulation to see which zones are projected to be affected next.",
      "Use resource positioning to pre-stage teams and supplies ahead of the forecast.",
      "Treat ML risk scores as guidance only; a person makes the final call on any alert.",
    ],
  },
  {
    key: "evidence",
    section: "Intelligence",
    title: "Evidence",
    description: "Imagery capture and computer-vision review.",
    steps: [
      "Open a captured image to see the computer-vision assessment beside the original photo.",
      "Confirm or correct the model's read before it informs an incident or an alert.",
      "Attach verified imagery to the incident it supports.",
    ],
  },
  {
    key: "recovery",
    section: "Recovery & readiness",
    title: "Recovery",
    description: "Damage claims, donations, and missing persons.",
    steps: [
      "Move between damage claims, donations, and missing persons using the section tabs.",
      "Verify each damage claim against its evidence before approving it.",
      "Match missing-person reports and log donations so the recovery record stays complete.",
    ],
  },
  {
    key: "preparedness",
    section: "Recovery & readiness",
    title: "Preparedness",
    description: "Awareness campaigns and school readiness.",
    steps: [
      "Plan and schedule public awareness campaigns from the campaign manager.",
      "Track school readiness so each site has a current flood plan and drill record.",
      "Publish preparedness material ahead of the forecast flood season.",
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
    description: "Step-by-step help for every command-center workspace.",
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
  "Response",
  "Intelligence",
  "Recovery & readiness",
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
