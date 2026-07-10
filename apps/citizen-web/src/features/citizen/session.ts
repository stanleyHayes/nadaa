import { useSyncExternalStore } from "react";

/**
 * Shared citizen session store. Viewing the platform (risk, alerts, shelters,
 * guides, open data) is public and needs no account. SUBMITTING anything —
 * incident reports, damage claims, aid pledges, missing-person reports — now
 * requires a signed-in citizen; anonymous submissions are not allowed. The
 * session is a single global store so every page/panel sees the same state
 * reactively, and any submission surface can open the sign-in dialog.
 *
 * Signed-in citizens also get a small account area (dashboard, report history,
 * notifications, profile, security, preferences). The editable profile — and the
 * multi-factor toggle — live on the session itself; preferences and the mock
 * notifications feed are persisted under their own versioned keys. Everything
 * here is local-only — there is no backend auth yet, so `changePassword` and
 * `setMfaEnabled` are mocks that only flip local state.
 */

/** How the citizen prefers to be reached about their reports. */
export type ContactChannel = "sms" | "call" | "whatsapp" | "email";

export type CitizenSession = {
  name: string;
  phone: string;
  region: string;
  language: string;
  since: string;
  /** Optional profile fields the citizen can add from the account area. */
  email?: string;
  contactChannel?: ContactChannel;
  /** Whether multi-factor authentication is on for this account (default off). */
  mfaEnabled?: boolean;
};

/** Editable profile fields (everything on the session except the join date). */
export type CitizenProfilePatch = Partial<Omit<CitizenSession, "since">>;

export type SavedReport = {
  reference: string;
  hazard: string;
  urgency: string;
  priorityReview: boolean;
  at: string;
};

/** Which channels the citizen wants official alerts delivered on. */
export type AlertChannels = {
  sms: boolean;
  email: boolean;
  push: boolean;
};

export type QuietHours = {
  enabled: boolean;
  /** 24h "HH:MM". */
  start: string;
  end: string;
};

export type CitizenPreferences = {
  /** Preferred language for guidance and alerts. */
  language: string;
  alertChannels: AlertChannels;
  /** Region the citizen wants to watch for alerts (defaults to their region). */
  regionOfInterest: string;
  quietHours: QuietHours;
};

export type NotificationCategory = "alert" | "report" | "shelter" | "system";

export type CitizenNotification = {
  id: string;
  category: NotificationCategory;
  title: string;
  body: string;
  /** ISO timestamp. */
  at: string;
  read: boolean;
};

/** Result of the mock password change — success or a human-readable reason. */
export type ChangePasswordResult = { ok: true } | { ok: false; error: string };

const SESSION_KEY = "nadaa.citizen.session.v1";
const REPORTS_KEY = "nadaa.citizen.savedReports.v1";
const PREFERENCES_KEY = "nadaa.citizen.preferences.v1";
const NOTIFICATIONS_KEY = "nadaa.citizen.notifications.v1";

export const signInRegions = [
  "Greater Accra",
  "Ashanti",
  "Western",
  "Central",
  "Eastern",
  "Volta",
  "Northern",
  "Upper East",
  "Upper West",
  "Bono",
  "Ahafo",
  "Oti",
  "Savannah",
  "North East",
  "Bono East",
  "Western North",
] as const;

function readJSON<T>(
  key: string,
  guard: (value: unknown) => value is T,
): T | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    const raw = window.localStorage.getItem(key);
    if (!raw) {
      return null;
    }
    const parsed = JSON.parse(raw) as unknown;
    return guard(parsed) ? parsed : null;
  } catch {
    return null;
  }
}

function writeJSON(key: string, value: unknown) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    window.localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // localStorage can be unavailable in private/restricted browser modes.
  }
}

function isSession(value: unknown): value is CitizenSession {
  return (
    typeof value === "object" &&
    value !== null &&
    typeof (value as CitizenSession).name === "string" &&
    typeof (value as CitizenSession).phone === "string" &&
    typeof (value as CitizenSession).region === "string" &&
    typeof (value as CitizenSession).language === "string"
  );
}

function isSavedReports(value: unknown): value is SavedReport[] {
  return (
    Array.isArray(value) &&
    value.every(
      (item) =>
        typeof item === "object" &&
        item !== null &&
        typeof (item as SavedReport).reference === "string",
    )
  );
}

function isPreferences(value: unknown): value is CitizenPreferences {
  if (typeof value !== "object" || value === null) {
    return false;
  }
  const pref = value as CitizenPreferences;
  return (
    typeof pref.language === "string" &&
    typeof pref.regionOfInterest === "string" &&
    typeof pref.alertChannels === "object" &&
    pref.alertChannels !== null &&
    typeof pref.alertChannels.sms === "boolean" &&
    typeof pref.alertChannels.email === "boolean" &&
    typeof pref.alertChannels.push === "boolean" &&
    typeof pref.quietHours === "object" &&
    pref.quietHours !== null &&
    typeof pref.quietHours.enabled === "boolean"
  );
}

function isNotifications(value: unknown): value is CitizenNotification[] {
  return (
    Array.isArray(value) &&
    value.every(
      (item) =>
        typeof item === "object" &&
        item !== null &&
        typeof (item as CitizenNotification).id === "string" &&
        typeof (item as CitizenNotification).title === "string" &&
        typeof (item as CitizenNotification).read === "boolean",
    )
  );
}

function defaultPreferences(session: CitizenSession | null): CitizenPreferences {
  return {
    language: session?.language ?? "en",
    alertChannels: { sms: true, email: false, push: true },
    regionOfInterest: session?.region ?? "Greater Accra",
    quietHours: { enabled: false, start: "22:00", end: "06:00" },
  };
}

/** Seed a short, believable notifications feed the first time the app loads. */
function seedNotifications(): CitizenNotification[] {
  const now = Date.now();
  const hoursAgo = (hours: number) =>
    new Date(now - hours * 60 * 60 * 1000).toISOString();
  return [
    {
      id: "ntf_flood_watch",
      category: "alert",
      title: "Flood watch issued for Greater Accra",
      body: "NADMO has issued a flood watch for low-lying areas near the Odaw river. Keep drains clear and be ready to move to higher ground.",
      at: hoursAgo(3),
      read: false,
    },
    {
      id: "ntf_report_verified",
      category: "report",
      title: "Your report is being reviewed",
      body: "Thank you. A NADMO officer has picked up your latest incident report and is verifying it before any public alert.",
      at: hoursAgo(20),
      read: false,
    },
    {
      id: "ntf_shelter_open",
      category: "shelter",
      title: "New shelter open near you",
      body: "Kaneshie Community Centre is now open and accepting families. Capacity and directions are on the Shelters page.",
      at: hoursAgo(52),
      read: true,
    },
    {
      id: "ntf_welcome",
      category: "system",
      title: "Welcome to your NADAA account",
      body: "Your dashboard keeps your reports, alerts and preferences in one place. Update how we reach you under Settings.",
      at: hoursAgo(96),
      read: true,
    },
  ];
}

type StoreState = {
  session: CitizenSession | null;
  savedReports: SavedReport[];
  preferences: CitizenPreferences;
  notifications: CitizenNotification[];
  signInOpen: boolean;
};

const initialSession = readJSON(SESSION_KEY, isSession);
const persistedNotifications = readJSON(NOTIFICATIONS_KEY, isNotifications);
const initialNotifications = persistedNotifications ?? seedNotifications();
if (!persistedNotifications) {
  // Persist the seed so read/unread state survives reloads.
  writeJSON(NOTIFICATIONS_KEY, initialNotifications);
}

let state: StoreState = {
  session: initialSession,
  savedReports: readJSON(REPORTS_KEY, isSavedReports) ?? [],
  preferences:
    readJSON(PREFERENCES_KEY, isPreferences) ??
    defaultPreferences(initialSession),
  notifications: initialNotifications,
  signInOpen: false,
};

const listeners = new Set<() => void>();

function emit() {
  for (const listener of listeners) {
    listener();
  }
}

function setState(patch: Partial<StoreState>) {
  state = { ...state, ...patch };
  emit();
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getSnapshot() {
  return state;
}

/** Sign a citizen in (persists) and close the sign-in dialog. */
export function signInCitizen(details: Omit<CitizenSession, "since">) {
  const session: CitizenSession = {
    ...details,
    since: new Date().toISOString(),
  };
  writeJSON(SESSION_KEY, session);
  setState({ session, signInOpen: false });
}

/** Sign out and clear the persisted session. Preferences and the local
 * notifications feed stay on the device (they are not tied to the sign-in). */
export function signOutCitizen() {
  if (typeof window !== "undefined") {
    window.localStorage.removeItem(SESSION_KEY);
  }
  setState({ session: null });
}

/** Persist a report reference for the signed-in citizen's "My reports" list. */
export function saveCitizenReport(report: SavedReport) {
  const exists = state.savedReports.some(
    (item) => item.reference === report.reference,
  );
  const savedReports = exists
    ? state.savedReports
    : [report, ...state.savedReports].slice(0, 12);
  writeJSON(REPORTS_KEY, savedReports);
  setState({ savedReports });
}

/** Update the signed-in citizen's editable profile fields (persists). */
export function updateCitizenProfile(patch: CitizenProfilePatch) {
  if (!state.session) {
    return;
  }
  const session: CitizenSession = { ...state.session, ...patch };
  writeJSON(SESSION_KEY, session);
  setState({ session });
}

/** Merge a preferences patch (nested channel/quiet-hours objects are merged). */
export function updateCitizenPreferences(patch: Partial<CitizenPreferences>) {
  const preferences: CitizenPreferences = {
    ...state.preferences,
    ...patch,
    alertChannels: {
      ...state.preferences.alertChannels,
      ...(patch.alertChannels ?? {}),
    },
    quietHours: {
      ...state.preferences.quietHours,
      ...(patch.quietHours ?? {}),
    },
  };
  writeJSON(PREFERENCES_KEY, preferences);
  setState({ preferences });
}

/**
 * Mock password change. There is no real credential store yet, so this only
 * validates the inputs and reports success/failure the account UI can surface.
 */
export function changeCitizenPassword(
  current: string,
  next: string,
): ChangePasswordResult {
  if (!state.session) {
    return { ok: false, error: "Sign in before changing your password." };
  }
  if (current.trim().length === 0) {
    return { ok: false, error: "Enter your current password." };
  }
  if (next.trim().length < 8) {
    return { ok: false, error: "New password must be at least 8 characters." };
  }
  if (next === current) {
    return {
      ok: false,
      error: "Choose a password different from your current one.",
    };
  }
  return { ok: true };
}

/**
 * Enable or disable multi-factor authentication for the signed-in citizen. The
 * flag is stored on the session (persisted under SESSION_KEY, like the profile
 * fields) and every subscriber is notified. This is a mock like
 * `changeCitizenPassword` — there is no authenticator backend yet, so it only
 * flips local state so the account UI can reflect it.
 * // TODO: wire to real MFA API
 */
export function setCitizenMfaEnabled(enabled: boolean) {
  if (!state.session) {
    return;
  }
  const session: CitizenSession = { ...state.session, mfaEnabled: enabled };
  writeJSON(SESSION_KEY, session);
  setState({ session });
}

/** Mark a single notification as read (persists). */
export function markCitizenNotificationRead(id: string) {
  const notifications = state.notifications.map((item) =>
    item.id === id ? { ...item, read: true } : item,
  );
  writeJSON(NOTIFICATIONS_KEY, notifications);
  setState({ notifications });
}

/** Mark every notification as read (persists). */
export function markAllCitizenNotificationsRead() {
  const notifications = state.notifications.map((item) =>
    item.read ? item : { ...item, read: true },
  );
  writeJSON(NOTIFICATIONS_KEY, notifications);
  setState({ notifications });
}

/** Open the sign-in dialog — used when a signed-out citizen tries to submit. */
export function requestSignIn() {
  setState({ signInOpen: true });
}

/** Close the sign-in dialog. */
export function closeSignIn() {
  setState({ signInOpen: false });
}

/**
 * Subscribe to the shared citizen session. Any component can read `session`,
 * gate a submission on it, and call `requestSignIn()` to open the dialog. The
 * account area additionally reads `preferences` / `notifications` / `mfaEnabled`
 * and calls the profile / preferences / password / MFA / notification actions.
 */
export function useCitizenSession() {
  const snapshot = useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
  return {
    session: snapshot.session,
    savedReports: snapshot.savedReports,
    preferences: snapshot.preferences,
    notifications: snapshot.notifications,
    /** Multi-factor state for the signed-in citizen (false when signed out). */
    mfaEnabled: snapshot.session?.mfaEnabled ?? false,
    signInOpen: snapshot.signInOpen,
    signIn: signInCitizen,
    signOut: signOutCitizen,
    saveReport: saveCitizenReport,
    updateProfile: updateCitizenProfile,
    updatePreferences: updateCitizenPreferences,
    changePassword: changeCitizenPassword,
    setMfaEnabled: setCitizenMfaEnabled,
    markNotificationRead: markCitizenNotificationRead,
    markAllRead: markAllCitizenNotificationsRead,
    requestSignIn,
    closeSignIn,
  };
}
