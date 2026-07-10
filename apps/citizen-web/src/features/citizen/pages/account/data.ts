import { Bell, FileText, LayoutDashboard, Settings } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { ContactChannel, NotificationCategory } from "../../session";

/** Left-hand sub-navigation for the account area (also drives the mobile row). */
export type AccountNavItem = {
  to: string;
  label: string;
  icon: LucideIcon;
  /** Exact-match highlighting for the index route. */
  end: boolean;
};

export const accountNavItems: AccountNavItem[] = [
  { to: "/account", label: "Overview", icon: LayoutDashboard, end: true },
  { to: "/account/reports", label: "My reports", icon: FileText, end: false },
  {
    to: "/account/notifications",
    label: "Notifications",
    icon: Bell,
    end: false,
  },
  { to: "/account/settings", label: "Settings", icon: Settings, end: false },
];

/** How the citizen prefers to be reached about their reports. */
export const contactChannelOptions: { label: string; value: ContactChannel }[] =
  [
    { label: "SMS text message", value: "sms" },
    { label: "Phone call", value: "call" },
    { label: "WhatsApp", value: "whatsapp" },
    { label: "Email", value: "email" },
  ];

/** Chip tone per notification category, reusing the brand colour roles. */
export const notificationTone: Record<
  NotificationCategory,
  "error" | "info" | "success" | "default"
> = {
  alert: "error",
  report: "info",
  shelter: "success",
  system: "default",
};

export const notificationCategoryLabel: Record<NotificationCategory, string> = {
  alert: "Alert",
  report: "Report",
  shelter: "Shelter",
  system: "System",
};
