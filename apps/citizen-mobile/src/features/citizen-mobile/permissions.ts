import type { MobilePermissionState, PermissionStatus } from "./types";

export const permissionCopy: Record<
  keyof MobilePermissionState,
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
      "Camera is blocked. You can still submit a report, then add media later from settings.",
    denied: "Camera was not allowed. Reports still work without photos.",
    granted: "Camera ready for incident photos.",
    prompt: "Use camera to capture incident evidence for responders.",
    title: "Camera",
  },
  location: {
    blocked:
      "Location is blocked. Type coordinates or choose a nearby area instead.",
    denied: "Location was not allowed. You can still report manually.",
    granted: "Location ready for risk checks, shelters, and report routing.",
    prompt: "Use location to find risk, shelters, and report accurately.",
    title: "Location",
  },
  media: {
    blocked: "Media library is blocked. You can still send text reports.",
    denied: "Media library was not allowed. Text reports still work.",
    granted: "Media library ready for report attachments.",
    prompt:
      "Attach existing photos or videos to help responders verify reports.",
    title: "Media",
  },
  push: {
    blocked:
      "Notifications are blocked. Use SMS, WhatsApp, USSD, or the alert feed as backup.",
    denied:
      "Notifications were not allowed. You can still check alerts in the app.",
    granted: "Push alerts ready for urgent warnings.",
    prompt: "Allow urgent warning notifications from NADAA.",
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
  key: keyof MobilePermissionState,
  status: PermissionStatus,
) {
  return permissionCopy[key][status === "unknown" ? "prompt" : status];
}
