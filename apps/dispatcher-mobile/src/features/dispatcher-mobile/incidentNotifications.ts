import { Platform } from "react-native";
import * as Notifications from "expo-notifications";
import type { IncidentRecord, IncidentUrgency } from "@nadaa/shared-types";

/**
 * Device notifications for dispatchers when new incidents hit their queue.
 *
 * Do-Not-Disturb is handled natively via notification channels: the default
 * "incidents" channel is high-importance but respects the device's DND, while
 * a life-threatening incident uses the "critical-incidents" channel (bypassDnd
 * on Android) and an iOS critical alert — so a life-safety call breaks through
 * silent/DND for on-duty responders, while routine queue updates stay quiet
 * during DND. Same rule as the citizen app, keyed on incident urgency.
 *
 * IMPORTANT: verify on a physical device before relying on this operationally.
 * iOS critical alerts also require Apple's critical-alerts entitlement; without
 * it, life-threatening incidents degrade to a normal high-priority alert.
 */

const CRITICAL_URGENCY: IncidentUrgency = "life_threatening";
const DEFAULT_CHANNEL = "incidents";
const CRITICAL_CHANNEL = "critical-incidents";

/** Life-threatening incidents override Do-Not-Disturb; others respect it. */
export function isCriticalIncident(urgency: IncidentUrgency): boolean {
  return urgency === CRITICAL_URGENCY;
}

/** Foreground handler + Android channels. Call once on startup. */
export async function configureIncidentNotifications(): Promise<void> {
  Notifications.setNotificationHandler({
    handleNotification: async () => ({
      shouldShowAlert: true,
      shouldShowBanner: true,
      shouldShowList: true,
      shouldPlaySound: true,
      shouldSetBadge: true,
    }),
  });

  if (Platform.OS === "android") {
    await Notifications.setNotificationChannelAsync(DEFAULT_CHANNEL, {
      name: "Incident queue",
      importance: Notifications.AndroidImportance.HIGH,
      sound: "default",
      lockscreenVisibility: Notifications.AndroidNotificationVisibility.PUBLIC,
      lightColor: "#F4C20D",
    });
    await Notifications.setNotificationChannelAsync(CRITICAL_CHANNEL, {
      name: "Life-threatening incidents",
      importance: Notifications.AndroidImportance.MAX,
      sound: "default",
      bypassDnd: true,
      lockscreenVisibility: Notifications.AndroidNotificationVisibility.PUBLIC,
      vibrationPattern: [0, 250, 250, 250, 250, 250],
      lightColor: "#E53935",
    });
  }
}

/** Ask for notification permission, including iOS critical alerts. */
export async function requestIncidentPermission(): Promise<boolean> {
  const current = await Notifications.getPermissionsAsync();
  if (current.granted) {
    return true;
  }
  const result = await Notifications.requestPermissionsAsync({
    ios: {
      allowAlert: true,
      allowSound: true,
      allowBadge: true,
      allowCriticalAlerts: true,
    },
  });
  return result.granted;
}

/** Present a device notification for one incident on the right channel. */
export async function presentIncidentNotification(
  incident: IncidentRecord,
): Promise<void> {
  const critical = isCriticalIncident(incident.urgency);
  await Notifications.scheduleNotificationAsync({
    content: {
      title: critical ? "Life-threatening incident" : "New incident in queue",
      body: incident.description,
      sound: "default",
      interruptionLevel: critical ? "critical" : "timeSensitive",
      data: { id: incident.id, urgency: incident.urgency },
    },
    trigger: { channelId: critical ? CRITICAL_CHANNEL : DEFAULT_CHANNEL },
  });
}

/**
 * Notify for newly-arrived, still-active incidents, adding them to `seen` so
 * none fire twice. The OS channel handles DND; life-threatening overrides it.
 */
export async function notifyNewIncidents(
  incidents: IncidentRecord[],
  seen: Set<string>,
): Promise<IncidentRecord[]> {
  const fresh = incidents.filter(
    (incident) =>
      !seen.has(incident.id) &&
      incident.status !== "closed" &&
      incident.status !== "false_report",
  );
  for (const incident of fresh) {
    seen.add(incident.id);
    await presentIncidentNotification(incident);
  }
  return fresh;
}
