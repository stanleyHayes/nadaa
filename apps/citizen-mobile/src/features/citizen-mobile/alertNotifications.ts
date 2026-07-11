import { Platform } from "react-native";
import * as Notifications from "expo-notifications";
import type { AlertSeverity, CitizenAlertFeedItem } from "@nadaa/shared-types";

/**
 * Device notifications for incoming citizen warnings.
 *
 * Do-Not-Disturb is handled the correct, native way — through notification
 * channels/importance — rather than an in-app guess (which is all the web can
 * do). The default "alerts" channel is high-importance but still respects the
 * device's DND/silent, while the "emergency" channel sets `bypassDnd` on Android
 * and level-5 (`emergency`) alerts are presented as iOS critical alerts. So a
 * level-5 warning breaks through silent/DND, and everything below it stays quiet
 * during DND — mirroring the web alert-sound rule.
 *
 * IMPORTANT: this must be verified on a physical device before it is relied on
 * for emergencies. iOS critical alerts also require the critical-alerts
 * entitlement from Apple; without it, level-5 alerts degrade to a normal
 * high-priority alert (still sound, but does not override the mute switch).
 */

const LEVEL_5: AlertSeverity = "emergency";
const DEFAULT_CHANNEL = "alerts";
const EMERGENCY_CHANNEL = "emergency";

/** Level-5 alerts override Do-Not-Disturb; everything else respects it. */
export function isEmergencyAlert(severity: AlertSeverity): boolean {
  return severity === LEVEL_5;
}

/** Foreground handler + Android channels. Call once on startup. */
export async function configureAlertNotifications(): Promise<void> {
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
      name: "Alerts",
      importance: Notifications.AndroidImportance.HIGH,
      sound: "default",
      lockscreenVisibility: Notifications.AndroidNotificationVisibility.PUBLIC,
      lightColor: "#F4C20D",
    });
    // Emergencies bypass Do-Not-Disturb.
    await Notifications.setNotificationChannelAsync(EMERGENCY_CHANNEL, {
      name: "Emergency warnings",
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
export async function requestAlertPermission(): Promise<boolean> {
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

/** Present a device notification for one alert on the appropriate channel. */
export async function presentAlertNotification(
  alert: CitizenAlertFeedItem,
): Promise<void> {
  const emergency = isEmergencyAlert(alert.severity);
  await Notifications.scheduleNotificationAsync({
    content: {
      title: alert.title,
      body: alert.message,
      sound: emergency ? "defaultCritical" : "default",
      interruptionLevel: emergency ? "critical" : "timeSensitive",
      data: { id: alert.id, severity: alert.severity },
    },
    trigger: { channelId: emergency ? EMERGENCY_CHANNEL : DEFAULT_CHANNEL },
  });
}

/**
 * Present notifications for newly-arrived current alerts, adding them to `seen`
 * so none fire twice. The OS channel handles DND; level-5 overrides it. Returns
 * the alerts that were notified.
 */
export async function notifyNewAlerts(
  alerts: CitizenAlertFeedItem[],
  seen: Set<string>,
): Promise<CitizenAlertFeedItem[]> {
  const fresh = alerts.filter(
    (alert) => alert.status === "current" && !seen.has(alert.id),
  );
  const delivered: CitizenAlertFeedItem[] = [];
  for (const alert of fresh) {
    try {
      await presentAlertNotification(alert);
      seen.add(alert.id);
      delivered.push(alert);
    } catch {
      // Leave un-seen so a transient failure retries on the next refresh.
    }
  }
  return delivered;
}
