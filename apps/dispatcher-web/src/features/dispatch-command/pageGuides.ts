import type { ViewId } from "./navigation";

/**
 * Contextual-help registry for the dispatch console. Keyed by the shell's
 * active view (the six nav views plus the synthetic "settings" and "guide"
 * views), each entry powers the page-help popover in the topbar and one card in
 * the user guide. Copy is written from a dispatcher's perspective.
 */
export type GuideKey = ViewId | "settings" | "guide";

export type GuideSection =
  | "Command"
  | "Response"
  | "Intelligence"
  | "Resources"
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
    description: "At-a-glance dispatch posture across Greater Accra.",
    steps: [
      "Scan the four status tiles first: active incidents, pending alerts, available hospital beds, and ML flood signals waiting on review.",
      "Read the live triage board to see how active incidents band from Emergency down to Low.",
      "Use the alert pipeline and hospital capacity cards, or the Jump to links, to open the desk that needs action.",
    ],
  },
  {
    key: "incidents",
    section: "Response",
    title: "Incident queue",
    description: "Command map, filters, incident queue, and response routing.",
    steps: [
      "Narrow the queue with the hazard, district, severity, status, and time filters, then read it top-down.",
      "Open the command map or click a queue row to review the report, its location, and any duplicates before you act.",
      "Verify the incident, assign a response team, and advance the status as crews move from en route to on scene.",
      "Merge duplicate reports and flag abuse so the queue stays clean.",
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
    key: "triage",
    section: "Intelligence",
    title: "AI triage",
    description: "Review, accept, or override the AI triage suggestion.",
    steps: [
      "Pick the incident to triage from the focused-incident selector at the top.",
      "Read the AI severity and priority suggestion together with the reasoning it gives.",
      "Accept the suggestion when it matches your read, or override it with your own severity, population, and reason.",
    ],
  },
  {
    key: "ml-review",
    section: "Intelligence",
    title: "ML flood review",
    description: "Human-in-the-loop review of ML flood predictions.",
    steps: [
      "Open a flood prediction to see its confidence, affected zone, and model detail.",
      "Add a review note and confirm the signal before it can inform any alert; the model never decides alone.",
      "Create an alert draft from a reviewed prediction to hand it to the broadcast desk.",
    ],
  },
  {
    key: "capacity",
    section: "Resources",
    title: "Capacity & relief",
    description: "Hospital bed capacity and relief distribution points.",
    steps: [
      "Filter hospital facilities by service, minimum free beds, and staleness to find where casualties can go.",
      "Watch occupancy against capacity; facilities near full are highlighted so you can reroute early.",
      "Publish and refresh relief distribution points so citizens know where to collect supplies.",
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
    description: "Step-by-step help for every dispatch-console desk.",
    steps: [
      "Browse the cards by section to find the desk you are working in.",
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
  "Resources",
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
