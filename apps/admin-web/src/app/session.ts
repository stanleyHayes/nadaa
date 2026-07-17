import { useSyncExternalStore } from "react";
import type { AgencyUserRole } from "@nadaa/shared-types";

/**
 * Governance-console session. Identity (including role) always comes from the
 * auth-service agency login response; the bearer token is persisted beside it
 * so `adminHeaders` can authorize every governance API call.
 */
export type AdminSession = {
  id: string;
  name: string;
  role: AgencyUserRole;
  agencyId: string;
  agency: string;
  mfaCompleted: boolean;
  /** Bearer token issued by `POST /auth/agency/login` (12h TTL). */
  accessToken: string;
  /** ISO timestamp when {@link AdminSession.accessToken} expires. */
  tokenExpiresAt?: string;
  /** Work email; derived from the display name when not supplied. */
  email?: string;
  /** Operating unit within the agency. */
  department?: string;
  /** Directory account state. */
  status?: "active" | "suspended";
  /** Whether an authenticator is enrolled on the account. */
  mfaEnabled?: boolean;
  /** ISO timestamp of the previous successful sign-in. */
  lastLoginAt?: string;
};

/** Patch accepted by {@link updateAdminProfile}. */
export type AdminProfilePatch = {
  name?: string;
  department?: string;
};

/**
 * Browser-scoped account preferences. These are presentation choices only and
 * never gate access, so they live beside the session but persist separately.
 */
export type AdminAccountPreferences = {
  inAppAlerts: boolean;
  criticalSms: boolean;
  approvalEmail: boolean;
  compactTables: boolean;
  reducedMotion: boolean;
};

/** Bundle returned by {@link useAdminSession}. */
export type AdminSessionState = {
  session: AdminSession | null;
  preferences: AdminAccountPreferences;
  updateProfile: (patch: AdminProfilePatch) => void;
  updatePreferences: (patch: Partial<AdminAccountPreferences>) => void;
  setMfaEnabled: (enabled: boolean) => void;
};

/**
 * Default agency identity per admin role. Only used as a display fallback when
 * a stored session predates a field; the sign-in flow always takes the agency
 * from the auth-service profile.
 */
export const agencyByRole: Record<AgencyUserRole, string> = {
  system_admin: "NADMO National Operations",
  agency_admin: "NADMO Accra Metro",
  nadmo_officer: "NADMO Accra Metro",
  district_officer: "Accra Metropolitan Assembly",
  dispatcher: "NADMO Dispatch Desk",
  responder: "Accra Regional Response Unit",
  agency_viewer: "Partner Agency Liaison",
};

/** Default operating unit per admin role. */
const departmentByRole: Record<AgencyUserRole, string> = {
  system_admin: "Platform Administration",
  agency_admin: "Agency Administration",
  nadmo_officer: "Flood Operations",
  district_officer: "District Coordination",
  dispatcher: "Dispatch Operations",
  responder: "Field Response",
  agency_viewer: "Partner Liaison",
};

const roleLabels: Record<AgencyUserRole, string> = {
  system_admin: "System administrator",
  agency_admin: "Agency administrator",
  nadmo_officer: "NADMO officer",
  district_officer: "District officer",
  dispatcher: "Dispatcher",
  responder: "Field responder",
  agency_viewer: "Agency viewer",
};

const STORAGE_KEY = "nadaa.admin.session.v1";
const PREFERENCES_KEY = "nadaa.admin.account-settings.v1";
const DEFAULT_AGENCY_ID = "00000000-0000-0000-0000-000000000101";
/**
 * Fixed reference timestamp for "last sign in". A constant keeps snapshots
 * stable across renders and avoids reading the clock at module load.
 */
const DEFAULT_LAST_LOGIN = "2026-07-09T07:42:00.000Z";

/** Build a plausible work email from a display name. */
function deriveEmail(name: string): string {
  const slug = name
    .trim()
    .toLowerCase()
    .replace(/[^a-z\s]/g, "")
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .join(".");
  return `${slug || "admin"}@nadmo.gov.gh`;
}

function normalizeStatus(value: unknown): "active" | "suspended" {
  return value === "suspended" ? "suspended" : "active";
}

/**
 * Fill the optional account fields from the required identity so any session —
 * including the bare object the sign-in screen builds — renders completely.
 */
function normalizeSession(base: AdminSession): AdminSession {
  return {
    ...base,
    email: base.email ?? deriveEmail(base.name),
    department: base.department ?? departmentByRole[base.role],
    status: normalizeStatus(base.status),
    mfaEnabled: base.mfaEnabled ?? Boolean(base.mfaCompleted),
    lastLoginAt: base.lastLoginAt ?? DEFAULT_LAST_LOGIN,
  };
}

function readStoredSession(): AdminSession | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }
    const parsed = JSON.parse(raw) as Partial<AdminSession>;
    if (
      !parsed ||
      typeof parsed.id !== "string" ||
      typeof parsed.role !== "string" ||
      typeof parsed.accessToken !== "string" ||
      !parsed.accessToken
    ) {
      return null;
    }
    const role = parsed.role as AgencyUserRole;
    return normalizeSession({
      id: parsed.id,
      name: parsed.name ?? roleLabels[role] ?? "Administrator",
      role,
      agencyId: parsed.agencyId ?? DEFAULT_AGENCY_ID,
      agency: parsed.agency ?? agencyByRole[role] ?? "NADAA",
      mfaCompleted: Boolean(parsed.mfaCompleted),
      accessToken: parsed.accessToken,
      tokenExpiresAt:
        typeof parsed.tokenExpiresAt === "string"
          ? parsed.tokenExpiresAt
          : undefined,
      email: typeof parsed.email === "string" ? parsed.email : undefined,
      department:
        typeof parsed.department === "string" ? parsed.department : undefined,
      status: normalizeStatus(parsed.status),
      mfaEnabled:
        typeof parsed.mfaEnabled === "boolean" ? parsed.mfaEnabled : undefined,
      lastLoginAt:
        typeof parsed.lastLoginAt === "string" ? parsed.lastLoginAt : undefined,
    });
  } catch {
    return null;
  }
}

export const defaultAccountPreferences: AdminAccountPreferences = {
  inAppAlerts: true,
  criticalSms: true,
  approvalEmail: true,
  compactTables: false,
  reducedMotion: false,
};

function isAccountPreferences(
  value: unknown,
): value is Partial<AdminAccountPreferences> {
  if (!value || typeof value !== "object") {
    return false;
  }
  return (
    Object.keys(defaultAccountPreferences) as Array<
      keyof AdminAccountPreferences
    >
  ).every((key) => {
    const candidate = (value as Record<string, unknown>)[key];
    return candidate === undefined || typeof candidate === "boolean";
  });
}

function readStoredPreferences(): AdminAccountPreferences {
  if (typeof window === "undefined") {
    return defaultAccountPreferences;
  }
  try {
    const raw = window.localStorage.getItem(PREFERENCES_KEY);
    if (!raw) {
      return defaultAccountPreferences;
    }
    const parsed: unknown = JSON.parse(raw);
    if (!isAccountPreferences(parsed)) {
      return defaultAccountPreferences;
    }
    return { ...defaultAccountPreferences, ...parsed };
  } catch {
    return defaultAccountPreferences;
  }
}

let currentSession: AdminSession | null = readStoredSession();
let currentPreferences: AdminAccountPreferences = readStoredPreferences();
const listeners = new Set<() => void>();
const preferenceListeners = new Set<() => void>();

function emit() {
  for (const listener of listeners) {
    listener();
  }
}

function emitPreferences() {
  for (const listener of preferenceListeners) {
    listener();
  }
}

function persist(session: AdminSession | null) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    if (session) {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
    } else {
      window.localStorage.removeItem(STORAGE_KEY);
    }
  } catch {
    // Storage can be unavailable (private mode); session stays in memory.
  }
}

function persistPreferences(preferences: AdminAccountPreferences) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    window.localStorage.setItem(PREFERENCES_KEY, JSON.stringify(preferences));
  } catch {
    // Storage can be unavailable (private mode); preferences stay in memory.
  }
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function subscribePreferences(listener: () => void) {
  preferenceListeners.add(listener);
  return () => {
    preferenceListeners.delete(listener);
  };
}

function getSnapshot(): AdminSession | null {
  return currentSession;
}

function getPreferencesSnapshot(): AdminAccountPreferences {
  return currentPreferences;
}

export function getAdminSession(): AdminSession | null {
  return currentSession;
}

export function getAdminPreferences(): AdminAccountPreferences {
  return currentPreferences;
}

export function signInAdmin(session: AdminSession) {
  currentSession = normalizeSession(session);
  persist(currentSession);
  emit();
}

export function signOutAdmin() {
  currentSession = null;
  persist(null);
  emit();
}

/**
 * Update the signed-in admin's editable profile fields. Email stays fixed —
 * it is presented as administrator-managed — so only name and department move.
 * TODO: wire to real profile/password/MFA API.
 */
export function updateAdminProfile(patch: AdminProfilePatch) {
  if (!currentSession) {
    return;
  }
  const trimmedName = patch.name?.trim();
  currentSession = {
    ...currentSession,
    name: trimmedName ? trimmedName : currentSession.name,
    department: patch.department ?? currentSession.department,
  };
  persist(currentSession);
  emit();
}

/** Merge a preferences patch and persist it to the account-settings key. */
export function updateAdminPreferences(
  patch: Partial<AdminAccountPreferences>,
) {
  currentPreferences = { ...currentPreferences, ...patch };
  persistPreferences(currentPreferences);
  emitPreferences();
}

/**
 * Toggle whether an authenticator is enrolled. This never clears the current
 * session's completed-MFA gate, so disabling never locks the admin out mid
 * session. TODO: wire to real profile/MFA API.
 */
export function setAdminMfaEnabled(enabled: boolean) {
  if (!currentSession) {
    return;
  }
  currentSession = { ...currentSession, mfaEnabled: enabled };
  persist(currentSession);
  emit();
}

/**
 * Reactive access to the signed-in administrator plus account preferences and
 * the store actions. `session` is null until an admin completes sign-in + MFA.
 */
export function useAdminSession(): AdminSessionState {
  const session = useSyncExternalStore(subscribe, getSnapshot, () => null);
  const preferences = useSyncExternalStore(
    subscribePreferences,
    getPreferencesSnapshot,
    () => defaultAccountPreferences,
  );
  return {
    session,
    preferences,
    updateProfile: updateAdminProfile,
    updatePreferences: updateAdminPreferences,
    setMfaEnabled: setAdminMfaEnabled,
  };
}

/**
 * Headers for governance API calls. The bearer token from agency login is the
 * credential every NADAA service now requires; the X-NADAA actor headers ride
 * along only for local dev, where services may run with
 * NADAA_AUTH_ALLOW_MOCK_ACTORS=true. Without a signed-in session there is no
 * token and no identity to send.
 */
export function adminHeaders() {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "X-NADAA-Request-ID": `admin-web-${Date.now()}`,
  };
  const session = currentSession;
  if (!session) {
    return headers;
  }
  headers.Authorization = `Bearer ${session.accessToken}`;
  headers["X-NADAA-Actor-ID"] = session.id;
  headers["X-NADAA-Actor-Role"] = session.role;
  headers["X-NADAA-Agency-ID"] = session.agencyId;
  headers["X-NADAA-MFA-Completed"] = session.mfaCompleted ? "true" : "false";
  return headers;
}
