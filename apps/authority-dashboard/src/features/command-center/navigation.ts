import {
  Camera,
  CloudRain,
  GraduationCap,
  HeartHandshake,
  LayoutDashboard,
  LifeBuoy,
  RadioTower,
  Siren,
  type LucideIcon,
} from "lucide-react";

export type ViewId =
  | "overview"
  | "incidents"
  | "alerts"
  | "shelters"
  | "forecasting"
  | "evidence"
  | "recovery"
  | "preparedness";

export type BadgeKey = "openIncidents" | "pendingAlerts" | "sheltersCritical";

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

export type NavGroup = {
  id: string;
  label: string;
  items: NavItem[];
};

export const navGroups: NavGroup[] = [
  {
    id: "command",
    label: "Command",
    items: [
      {
        id: "overview",
        label: "Overview",
        description: "At-a-glance triage across Greater Accra operations",
        icon: LayoutDashboard,
      },
    ],
  },
  {
    id: "response",
    label: "Response",
    items: [
      {
        id: "incidents",
        label: "Incidents",
        description: "Command map, queue, incident detail, and response routing",
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
      {
        id: "shelters",
        label: "Shelters & relief",
        description: "Update shelter capacity and publish relief distribution",
        icon: LifeBuoy,
        badgeKey: "sheltersCritical",
        badgeTone: "gold",
      },
    ],
  },
  {
    id: "intelligence",
    label: "Intelligence",
    items: [
      {
        id: "forecasting",
        label: "Forecasting",
        description: "Resource positioning, flood simulation, and ML review",
        icon: CloudRain,
      },
      {
        id: "evidence",
        label: "Evidence",
        description: "Imagery capture and computer-vision review",
        icon: Camera,
      },
    ],
  },
  {
    id: "recovery-readiness",
    label: "Recovery & readiness",
    items: [
      {
        id: "recovery",
        label: "Recovery",
        description: "Damage claims, donations, and missing persons",
        icon: HeartHandshake,
      },
      {
        id: "preparedness",
        label: "Preparedness",
        description: "Awareness campaigns and school readiness",
        icon: GraduationCap,
      },
    ],
  },
];

export const navItems: NavItem[] = navGroups.flatMap((group) => group.items);

export function navItemById(id: ViewId): NavItem {
  return navItems.find((item) => item.id === id) ?? navItems[0];
}

export function groupLabelForView(id: ViewId): string {
  const group = navGroups.find((candidate) =>
    candidate.items.some((item) => item.id === id),
  );
  return group?.label ?? "Command";
}

export const DEFAULT_VIEW: ViewId = "overview";

export function isViewId(value: string | null): value is ViewId {
  return Boolean(value && navItems.some((item) => item.id === value));
}
