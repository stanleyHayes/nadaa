import * as ImagePicker from "expo-image-picker";
import * as Location from "expo-location";
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

type OSPermissionAnswer = {
  canAskAgain?: boolean;
  granted: boolean;
};

function resolveOSPermission(answer: OSPermissionAnswer): PermissionStatus {
  if (answer.granted) {
    return "granted";
  }
  return answer.canAskAgain === false ? "blocked" : "denied";
}

/**
 * Request the REAL OS permission for camera, location, or media library and
 * report the actual OS answer — these are device permissions, not in-app
 * preferences. Push notifications are handled separately (alertNotifications).
 */
export async function requestOSPermission(
  key: Exclude<keyof MobilePermissionState, "push">,
): Promise<PermissionStatus> {
  try {
    switch (key) {
      case "camera":
        return resolveOSPermission(
          await ImagePicker.requestCameraPermissionsAsync(),
        );
      case "location":
        return resolveOSPermission(
          await Location.requestForegroundPermissionsAsync(),
        );
      case "media":
        return resolveOSPermission(
          await ImagePicker.requestMediaLibraryPermissionsAsync(),
        );
    }
  } catch {
    // Permission module unavailable (e.g. unsupported platform): keep "unknown".
    return "unknown";
  }
}

export function permissionMessage(
  key: keyof MobilePermissionState,
  status: PermissionStatus,
) {
  return permissionCopy[key][status === "unknown" ? "prompt" : status];
}
