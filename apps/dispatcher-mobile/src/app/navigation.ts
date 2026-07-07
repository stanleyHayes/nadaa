export type DispatcherTab =
  "queue" | "detail" | "action" | "capacity" | "profile";

export const dispatcherTabs: Array<{
  id: DispatcherTab;
  icon: string;
  label: string;
}> = [
  { id: "queue", icon: "list", label: "Queue" },
  { id: "detail", icon: "file-text", label: "Detail" },
  { id: "action", icon: "edit-3", label: "Action" },
  { id: "capacity", icon: "activity", label: "Capacity" },
  { id: "profile", icon: "shield", label: "Profile" },
];
