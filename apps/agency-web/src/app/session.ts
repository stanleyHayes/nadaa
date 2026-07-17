import { useSyncExternalStore } from "react";
import type { AgencyUserRole } from "@nadaa/shared-types";

export const agencyRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
  "responder",
  "agency_viewer",
];

export type AgencySession = {
  id: string;
  name: string;
  role: AgencyUserRole;
  agencyId: string;
  agency: string;
  district: string;
  token: string;
  mfaCompleted: boolean;
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

/** Patch accepted by {@link updateAgencyProfile}. */
export type AgencyProfilePatch = {
  name?: string;
  department?: string;
};

/** Result of the mock {@link changeAgencyPassword} action. */
export type PasswordChangeResult = {
  ok: boolean;
  message: string;
};

/**
 * Browser-scoped account preferences. These are presentation choices only and
 * never gate access, so they live beside the session but persist separately.
 */
export type AgencyAccountPreferences = {
  inAppAlerts: boolean;
  criticalSms: boolean;
  approvalEmail: boolean;
  compactTables: boolean;
  reducedMotion: boolean;
};

/** Bundle returned by {@link useAgencySession}. */
export type AgencySessionState = {
  session: AgencySession | null;
  preferences: AgencyAccountPreferences;
  updateProfile: (patch: AgencyProfilePatch) => void;
  updatePreferences: (patch: Partial<AgencyAccountPreferences>) => void;
  setMfaEnabled: (enabled: boolean) => void;
  changePassword: (current: string, next: string) => PasswordChangeResult;
};

/**
 * Default agency identity per role. Used to pre-fill the sign-in screen; a live
 * deployment would resolve this from the agency directory service.
 */
export const agencyByRole: Record<AgencyUserRole, string> = {
  system_admin: "NADAA National Command",
  agency_admin: "NADMO National Secretariat",
  nadmo_officer: "NADMO Accra Metro",
  district_officer: "Accra Metropolitan Assembly",
  dispatcher: "NADMO Dispatch Desk",
  responder: "Accra Regional Response Unit",
  agency_viewer: "Partner Agency Liaison",
};

/** Default operating unit per agency role. */
const departmentByRole: Record<AgencyUserRole, string> = {
  system_admin: "Platform Administration",
  agency_admin: "Agency Administration",
  nadmo_officer: "Flood Operations",
  district_officer: "District Coordination",
  dispatcher: "Dispatch Operations",
  responder: "Field Response",
  agency_viewer: "Partner Liaison",
};

export const roleLabels: Record<AgencyUserRole, string> = {
  system_admin: "System administrator",
  agency_admin: "Agency administrator",
  nadmo_officer: "NADMO officer",
  district_officer: "District officer",
  dispatcher: "Dispatcher",
  responder: "Field responder",
  agency_viewer: "Agency viewer",
};

const STORAGE_KEY = "nadaa.agency.session.v1";
const PREFERENCES_KEY = "nadaa.agency.account-settings.v1";
const DEFAULT_AGENCY_ID = "00000000-0000-0000-0000-000000000101";
const DEFAULT_DISTRICT = "Accra Metropolitan";
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
  return `${slug || "officer"}@nadmo.gov.gh`;
}

function normalizeStatus(value: unknown): "active" | "suspended" {
  return value === "suspended" ? "suspended" : "active";
}

/**
 * Fill the optional account fields from the required identity so any session —
 * including the bare object the sign-in screen builds — renders completely.
 */
function normalizeSession(base: AgencySession): AgencySession {
  return {
    ...base,
    email: base.email ?? deriveEmail(base.name),
    department: base.department ?? departmentByRole[base.role],
    status: normalizeStatus(base.status),
    mfaEnabled: base.mfaEnabled ?? Boolean(base.mfaCompleted),
    lastLoginAt: base.lastLoginAt ?? DEFAULT_LAST_LOGIN,
  };
}

function readStoredSession(): AgencySession | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }
    const parsed = JSON.parse(raw) as Partial<AgencySession>;
    if (
      !parsed ||
      typeof parsed.id !== "string" ||
      typeof parsed.role !== "string" ||
      !agencyRoles.includes(parsed.role as AgencyUserRole)
    ) {
      return null;
    }
    const role = parsed.role as AgencyUserRole;
    return normalizeSession({
      id: parsed.id,
      name: parsed.name ?? roleLabels[role],
      role,
      agencyId: parsed.agencyId ?? DEFAULT_AGENCY_ID,
      agency: parsed.agency ?? agencyByRole[role],
      district: parsed.district ?? DEFAULT_DISTRICT,
      token: parsed.token ?? `agency-${role}`,
      mfaCompleted: Boolean(parsed.mfaCompleted),
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

export const defaultAccountPreferences: AgencyAccountPreferences = {
  inAppAlerts: true,
  criticalSms: true,
  approvalEmail: true,
  compactTables: false,
  reducedMotion: false,
};

function isAccountPreferences(
  value: unknown,
): value is Partial<AgencyAccountPreferences> {
  if (!value || typeof value !== "object") {
    return false;
  }
  return (
    Object.keys(defaultAccountPreferences) as Array<
      keyof AgencyAccountPreferences
    >
  ).every((key) => {
    const candidate = (value as Record<string, unknown>)[key];
    return candidate === undefined || typeof candidate === "boolean";
  });
}

function readStoredPreferences(): AgencyAccountPreferences {
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

let currentSession: AgencySession | null = readStoredSession();
let currentPreferences: AgencyAccountPreferences = readStoredPreferences();
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

function persist(session: AgencySession | null) {
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

function persistPreferences(preferences: AgencyAccountPreferences) {
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

function getSnapshot(): AgencySession | null {
  return currentSession;
}

function getPreferencesSnapshot(): AgencyAccountPreferences {
  return currentPreferences;
}

export function getAgencySession(): AgencySession | null {
  return currentSession;
}

export function getAgencyPreferences(): AgencyAccountPreferences {
  return currentPreferences;
}

export function signInAgency(session: AgencySession) {
  currentSession = normalizeSession(session);
  persist(currentSession);
  emit();
}

export function signOutAgency() {
  currentSession = null;
  persist(null);
  emit();
}

/**
 * Update the signed-in operator's editable profile fields. Email stays fixed —
 * it is presented as administrator-managed — so only name and department move.
 * TODO: wire to real profile/password/MFA API.
 */
export function updateAgencyProfile(patch: AgencyProfilePatch) {
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
export function updateAgencyPreferences(
  patch: Partial<AgencyAccountPreferences>,
) {
  currentPreferences = { ...currentPreferences, ...patch };
  persistPreferences(currentPreferences);
  emitPreferences();
}

/**
 * Toggle whether an authenticator is enrolled. This never clears the current
 * session's completed-MFA gate, so disabling never locks the operator out mid
 * shift. TODO: wire to real profile/password/MFA API.
 */
export function setAgencyMfaEnabled(enabled: boolean) {
  if (!currentSession) {
    return;
  }
  currentSession = { ...currentSession, mfaEnabled: enabled };
  persist(currentSession);
  emit();
}

/**
 * Mock password change. Validates locally and returns a result; there is no
 * backend in this build. TODO: wire to real profile/password/MFA API.
 */
export function changeAgencyPassword(
  current: string,
  next: string,
): PasswordChangeResult {
  if (!current.trim()) {
    return { ok: false, message: "Enter your current password." };
  }
  if (next.length < 8) {
    return {
      ok: false,
      message: "New password must be at least 8 characters.",
    };
  }
  if (next === current) {
    return {
      ok: false,
      message: "New password must be different from your current password.",
    };
  }
  return {
    ok: true,
    message: "Password updated. Use it the next time you sign in.",
  };
}

/**
 * Reactive access to the signed-in agency operator plus account preferences and
 * the store actions. `session` is null until an operator completes sign-in +
 * MFA.
 */
export function useAgencySession(): AgencySessionState {
  const session = useSyncExternalStore(subscribe, getSnapshot, () => null);
  const preferences = useSyncExternalStore(
    subscribePreferences,
    getPreferencesSnapshot,
    () => defaultAccountPreferences,
  );
  return {
    session,
    preferences,
    updateProfile: updateAgencyProfile,
    updatePreferences: updateAgencyPreferences,
    setMfaEnabled: setAgencyMfaEnabled,
    changePassword: changeAgencyPassword,
  };
}

export function hasAgencyAccess(session: AgencySession | null): boolean {
  return Boolean(
    session && session.mfaCompleted && agencyRoles.includes(session.role),
  );
}

/**
 * Roles allowed to mutate shelter-service resources (shelter occupancy,
 * hospital capacity, relief points, aid requests) and export the aid CSV.
 * Mirrors `ShelterUpdateRoles` in services/shelter-service/internal/utils.
 */
export const shelterWriteRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
];

/** Roles allowed to advance an incident's status. Mirrors incident-service. */
export const incidentStatusRoles: AgencyUserRole[] = [
  ...shelterWriteRoles,
  "responder",
];

export function canManageShelterResources(
  session: AgencySession | null,
): boolean {
  return Boolean(session && shelterWriteRoles.includes(session.role));
}

export function canUpdateIncidentStatus(
  session: AgencySession | null,
): boolean {
  return Boolean(session && incidentStatusRoles.includes(session.role));
}

/**
 * Request headers for agency API calls. When no operator is signed in, no
 * identity headers are sent at all — a signed-out client must look anonymous
 * so expired or revoked sessions trip the 401 guard instead of silently
 * acting under a fallback identity.
 */
export function agencyHeaders() {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "X-NADAA-Request-ID": `agency-web-${Date.now()}`,
  };
  if (currentSession) {
    headers.Authorization = `Bearer ${currentSession.token}`;
  }
  return headers;
}
