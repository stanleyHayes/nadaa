import {
  BellRing,
  Building2,
  DatabaseZap,
  Gauge,
  KeyRound,
  Landmark,
  LayoutDashboard,
  LockKeyhole,
  Network,
  ScrollText,
  ShieldCheck,
  UsersRound,
  type LucideIcon,
} from "lucide-react";

export type ViewId =
  | "overview"
  | "agencies"
  | "users"
  | "roles"
  | "mfa"
  | "audit"
  | "integrations"
  | "alertRules";

export type BadgeKey = "agencies" | "usersAwaitingMfa";

export type NavItem = {
  id: ViewId;
  label: string;
  /** Admin-facing subtitle shown under the view title in the topbar. */
  description: string;
  icon: LucideIcon;
  badgeKey?: BadgeKey;
  /** Tone for the badge; unfinished work reads as a warning, not info. */
  badgeTone?: "gold" | "green" | "red";
};

export type NavGroup = {
  id: string;
  label: string;
  icon: LucideIcon;
  items: NavItem[];
};

export const navGroups: NavGroup[] = [
  {
    id: "command",
    label: "Command",
    icon: Gauge,
    items: [
      {
        id: "overview",
        label: "Overview",
        description: "Governance posture across agencies, access, and audit",
        icon: LayoutDashboard,
      },
    ],
  },
  {
    id: "directory",
    label: "Directory",
    icon: Network,
    items: [
      {
        id: "agencies",
        label: "Agencies",
        description: "Registered agencies, operating scope, and MFA coverage",
        icon: Building2,
        badgeKey: "agencies",
        badgeTone: "green",
      },
      {
        id: "users",
        label: "Users",
        description: "Authority users, roles, and access provisioning",
        icon: UsersRound,
      },
    ],
  },
  {
    id: "access",
    label: "Access & policy",
    icon: KeyRound,
    items: [
      {
        id: "roles",
        label: "Roles & access",
        description: "Admin, alert, and operational permission matrix",
        icon: LockKeyhole,
      },
      {
        id: "mfa",
        label: "MFA readiness",
        description: "Two-step verification coverage and users awaiting setup",
        icon: ShieldCheck,
        badgeKey: "usersAwaitingMfa",
        badgeTone: "gold",
      },
    ],
  },
  {
    id: "governance",
    label: "Governance",
    icon: Landmark,
    items: [
      {
        id: "audit",
        label: "Audit trail",
        description: "Sensitive-action trace across the platform",
        icon: ScrollText,
      },
      {
        id: "integrations",
        label: "Data sources",
        description: "Integration contracts and safe secret scopes",
        icon: DatabaseZap,
      },
      {
        id: "alertRules",
        label: "Alert rules",
        description: "Approval, override, targeting, and audit rules",
        icon: BellRing,
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
