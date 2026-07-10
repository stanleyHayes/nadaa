import {
  CloudRain,
  HeartPulse,
  LayoutDashboard,
  Radar,
  RadioTower,
  Siren,
  Sparkles,
  type LucideIcon,
} from "lucide-react";

export type ViewId =
  | "overview"
  | "incidents"
  | "triage"
  | "ml-review"
  | "alerts"
  | "capacity";

export type BadgeKey = "openIncidents" | "pendingAlerts" | "mlNeedsReview";

export type NavItem = {
  id: ViewId;
  label: string;
  /** Operator-facing subtitle shown under the view title in the topbar. */
  description: string;
  icon: LucideIcon;
  badgeKey?: BadgeKey;
  /** Tone for the badge; critical counts read as warnings, not info. */
  badgeTone?: "gold" | "green" | "red";
};

export type GroupAccent = "navy" | "red" | "gold" | "green";

export type NavGroup = {
  id: string;
  label: string;
  /** Leading icon rendered inside the group-heading chip. */
  icon: LucideIcon;
  /** Accent used for the group chip and the active connector branch. */
  accent: GroupAccent;
  items: NavItem[];
};

export const navGroups: NavGroup[] = [
  {
    id: "command",
    label: "Command",
    icon: LayoutDashboard,
    accent: "navy",
    items: [
      {
        id: "overview",
        label: "Overview",
        description: "At-a-glance dispatch posture across Greater Accra",
        icon: LayoutDashboard,
      },
    ],
  },
  {
    id: "response",
    label: "Response",
    icon: Siren,
    accent: "red",
    items: [
      {
        id: "incidents",
        label: "Incident queue",
        description: "Live command map, incident queue, and response routing",
        icon: Siren,
        badgeKey: "openIncidents",
        badgeTone: "red",
      },
      {
        id: "alerts",
        label: "Alerts & broadcast",
        description: "Draft, review, and publish public safety alerts",
        icon: RadioTower,
        badgeKey: "pendingAlerts",
        badgeTone: "gold",
      },
    ],
  },
  {
    id: "intelligence",
    label: "Intelligence",
    icon: Radar,
    accent: "gold",
    items: [
      {
        id: "triage",
        label: "AI triage",
        description: "Review, accept, or override the AI triage suggestion",
        icon: Sparkles,
      },
      {
        id: "ml-review",
        label: "ML flood review",
        description: "Human-in-the-loop review of ML flood predictions",
        icon: CloudRain,
        badgeKey: "mlNeedsReview",
        badgeTone: "gold",
      },
    ],
  },
  {
    id: "resources",
    label: "Resources",
    icon: HeartPulse,
    accent: "green",
    items: [
      {
        id: "capacity",
        label: "Capacity & relief",
        description: "Hospital bed capacity and relief distribution points",
        icon: HeartPulse,
      },
    ],
  },
];

export const navItems: NavItem[] = navGroups.flatMap((group) => group.items);

export function navItemById(id: ViewId): NavItem {
  return navItems.find((item) => item.id === id) ?? navItems[0];
}

export function groupForView(id: ViewId): NavGroup | undefined {
  return navGroups.find((candidate) =>
    candidate.items.some((item) => item.id === id),
  );
}

export function groupLabelForView(id: ViewId): string {
  return groupForView(id)?.label ?? "Command";
}

export function groupIdForView(id: ViewId): string {
  return groupForView(id)?.id ?? navGroups[0].id;
}

export const DEFAULT_VIEW: ViewId = "overview";

export function isViewId(value: string | null): value is ViewId {
  return Boolean(value && navItems.some((item) => item.id === value));
}
