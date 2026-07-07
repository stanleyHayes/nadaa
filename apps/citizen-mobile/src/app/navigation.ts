export type CitizenTab =
  "home" | "alerts" | "report" | "community" | "guides" | "support";

export const citizenTabs: Array<{
  id: CitizenTab;
  icon: string;
  label: string;
}> = [
  { id: "home", icon: "activity", label: "Home" },
  { id: "alerts", icon: "bell", label: "Alerts" },
  { id: "report", icon: "send", label: "Report" },
  { id: "community", icon: "users", label: "Community" },
  { id: "guides", icon: "book-open", label: "Guides" },
  { id: "support", icon: "map-pin", label: "Help" },
];
