import type { DispatcherPermissionState, PermissionStatus } from "./types";

export const permissionCopy: Record<
  keyof DispatcherPermissionState,
  {
    blocked: string;
    denied: string;
    granted: string;
    prompt: string;
    title: string;
  }
> = {
  camera: {
    blocked:
      "Camera is blocked. You can still update incidents with text notes.",
    denied: "Camera was not allowed. Scene photos can be added later.",
    granted: "Camera ready for scene photos.",
    prompt: "Use camera to capture scene evidence.",
    title: "Camera",
  },
  location: {
    blocked:
      "Location is blocked. Incident and capacity lookups still work from the report location.",
    denied:
      "Location was not allowed. Use incident coordinates for nearby context.",
    granted: "Location ready for incident and capacity context.",
    prompt: "Use location to find nearby incidents and hospital capacity.",
    title: "Location",
  },
  push: {
    blocked:
      "Notifications are blocked. Check the queue regularly for critical incidents.",
    denied:
      "Notifications were not allowed. Critical escalations appear in the incident queue.",
    granted: "Push alerts ready for critical incident escalation.",
    prompt: "Allow urgent escalation notifications from NADAA Dispatcher.",
    title: "Push alerts",
  },
};

export function nextPermissionStatus(
  status: PermissionStatus,
): PermissionStatus {
  switch (status) {
    case "unknown":
      return "granted";
    case "denied":
      return "blocked";
    case "blocked":
      return "unknown";
    case "granted":
      return "denied";
  }
}

export function permissionMessage(
  key: keyof DispatcherPermissionState,
  status: PermissionStatus,
) {
  return permissionCopy[key][status === "unknown" ? "prompt" : status];
}
