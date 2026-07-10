import type { ViewId } from "./navigation";

/**
 * Contextual-help registry for the agency workspace. Keyed by the shell's active
 * view (the five nav views plus the synthetic "settings" and "guide" views),
 * each entry powers the page-help popover in the topbar and one card in the user
 * guide. Copy is written from a field-desk officer's perspective.
 */
export type GuideKey = ViewId | "settings" | "guide";

export type GuideSection =
  | "Operations"
  | "Response"
  | "Relief & aid"
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
    section: "Operations",
    title: "Overview",
    description: "Response posture across the operations assigned to your desk.",
    steps: [
      "Read the four status tiles first: assigned to your desk, en route, on scene, and priority review.",
      "Scan the response-posture ladder to see how your active incidents band across each response stage.",
      "Use the capacity and relief cards, or the Jump to links, to open the desk that needs action.",
    ],
  },
  {
    key: "incidents",
    section: "Response",
    title: "Assigned incidents",
    description: "Triage, review, and update the incidents on your desk.",
    steps: [
      "Filter the queue to focus the incidents you own; priority and severe reports need eyes first.",
      "Open an incident to review its report and full detail before you act.",
      "Advance the status as your crews move from en route to on scene and into recovery.",
    ],
  },
  {
    key: "capacity",
    section: "Response",
    title: "Nearby capacity",
    description: "Shelters, hospital beds, and road closures around a scene.",
    steps: [
      "Open an assigned incident first to recentre this page on its scene.",
      "Update shelter occupancy and hospital beds as people and patients arrive so the public map stays accurate.",
      "Check nearby road closures and relief points before you route a team.",
    ],
  },
  {
    key: "relief",
    section: "Relief & aid",
    title: "Relief distribution",
    description: "Publish distribution points and keep stock levels current.",
    steps: [
      "Publish a new distribution point, or open one to edit its location and eligibility notes.",
      "Update stock levels as supplies move so citizens see accurate availability.",
      "Review the stock history to confirm what was distributed and when.",
    ],
  },
  {
    key: "aid",
    section: "Relief & aid",
    title: "Aid & donations",
    description: "Review verified aid needs and coordinate partner pledges.",
    steps: [
      "Create an aid need, then approve, pause, or close its public listing after review.",
      "Track partner pledges against each need so coordination stays in one place.",
      "Export the coordination report as CSV to share with partners; this never changes incident status.",
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
    description: "Step-by-step help for every agency workspace.",
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
  "Operations",
  "Response",
  "Relief & aid",
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
