import {
  Building2,
  ClipboardList,
  Gauge,
  HandHeart,
  HeartHandshake,
  LayoutDashboard,
  PackageCheck,
  Siren,
  type LucideIcon,
} from "lucide-react";

export type ViewId = "overview" | "incidents" | "capacity" | "relief" | "aid";

export type BadgeKey =
  | "openIncidents"
  | "sheltersCritical"
  | "reliefOpen"
  | "aidOpen";

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

export type GroupAccent = "navy" | "gold" | "green";

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
    id: "operations",
    label: "Operations",
    icon: Gauge,
    accent: "navy",
    items: [
      {
        id: "overview",
        label: "Overview",
        description: "Response posture across your assigned operations",
        icon: LayoutDashboard,
      },
    ],
  },
  {
    id: "response",
    label: "Response",
    icon: Siren,
    accent: "gold",
    items: [
      {
        id: "incidents",
        label: "Assigned incidents",
        description: "Triage, review, and update the incidents on your desk",
        icon: ClipboardList,
        badgeKey: "openIncidents",
        badgeTone: "red",
      },
      {
        id: "capacity",
        label: "Nearby capacity",
        description: "Shelters, hospital beds, and road closures around a scene",
        icon: Building2,
        badgeKey: "sheltersCritical",
        badgeTone: "gold",
      },
    ],
  },
  {
    id: "relief-aid",
    label: "Relief & aid",
    icon: HeartHandshake,
    accent: "green",
    items: [
      {
        id: "relief",
        label: "Relief distribution",
        description: "Publish distribution points and keep stock levels current",
        icon: PackageCheck,
        badgeKey: "reliefOpen",
        badgeTone: "green",
      },
      {
        id: "aid",
        label: "Aid & donations",
        description: "Review verified aid needs and coordinate partner pledges",
        icon: HandHeart,
        badgeKey: "aidOpen",
        badgeTone: "gold",
      },
    ],
  },
];

export const navItems: NavItem[] = navGroups.flatMap((group) => group.items);

export function navItemById(id: ViewId): NavItem {
  return navItems.find((item) => item.id === id) ?? navItems[0];
}

export function groupLabelForView(id: ViewId): string {
  return groupForView(id)?.label ?? "Operations";
}

export function groupForView(id: ViewId): NavGroup | undefined {
  return navGroups.find((candidate) =>
    candidate.items.some((item) => item.id === id),
  );
}

export function groupIdForView(id: ViewId): string {
  return groupForView(id)?.id ?? navGroups[0].id;
}

export const DEFAULT_VIEW: ViewId = "overview";

export function isViewId(value: string | null): value is ViewId {
  return Boolean(value && navItems.some((item) => item.id === value));
}
