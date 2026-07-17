import { useSyncExternalStore } from "react";
import type { AgencyUserRole } from "@nadaa/shared-types";

export const commandRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
  "responder",
  "agency_viewer",
];

export type AuthoritySession = {
  id: string;
  name: string;
  role: AgencyUserRole;
  agencyId: string;
  agency: string;
  district: string;
  mfaCompleted: boolean;
  /** Bearer token issued by auth-service agency login. */
  accessToken?: string;
  /** ISO timestamp when {@link accessToken} stops being accepted. */
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

/** Patch accepted by {@link updateAuthorityProfile}. */
export type AuthorityProfilePatch = {
  name?: string;
  department?: string;
};

/** Result of the mock {@link changeAuthorityPassword} action. */
export type PasswordChangeResult = {
  ok: boolean;
  message: string;
};

/**
 * Browser-scoped account preferences. These are presentation choices only and
 * never gate access, so they live beside the session but persist separately.
 */
export type AuthorityAccountPreferences = {
  inAppAlerts: boolean;
  criticalSms: boolean;
  approvalEmail: boolean;
  compactTables: boolean;
  reducedMotion: boolean;
};

/** Bundle returned by {@link useAuthoritySession}. */
export type AuthoritySessionState = {
  session: AuthoritySession | null;
  preferences: AuthorityAccountPreferences;
  updateProfile: (patch: AuthorityProfilePatch) => void;
  updatePreferences: (patch: Partial<AuthorityAccountPreferences>) => void;
  setMfaEnabled: (enabled: boolean) => void;
  changePassword: (current: string, next: string) => PasswordChangeResult;
};

/**
 * Default agency identity per command role. Used to pre-fill the sign-in
 * screen; a live deployment would resolve this from the directory service.
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

/** Default operating unit per command role. */
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

const STORAGE_KEY = "nadaa.authority.session.v1";
const PREFERENCES_KEY = "nadaa.authority.account-settings.v1";
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
function normalizeSession(base: AuthoritySession): AuthoritySession {
  return {
    ...base,
    email: base.email ?? deriveEmail(base.name),
    department: base.department ?? departmentByRole[base.role],
    status: normalizeStatus(base.status),
    mfaEnabled: base.mfaEnabled ?? Boolean(base.mfaCompleted),
    lastLoginAt: base.lastLoginAt ?? DEFAULT_LAST_LOGIN,
  };
}

/**
 * Fallback identity used only for request headers if a call somehow fires
 * before an operator signs in. The UI never renders command surfaces without
 * an authenticated session, so this keeps API clients resilient in dev.
 */
const fallbackSession: AuthoritySession = normalizeSession({
  id: "usr_nadmo_accra",
  name: "NADMO Officer",
  role: "nadmo_officer",
  agencyId: DEFAULT_AGENCY_ID,
  agency: "NADMO Accra Metro",
  district: DEFAULT_DISTRICT,
  mfaCompleted: true,
});

function readStoredSession(): AuthoritySession | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }
    const parsed = JSON.parse(raw) as Partial<AuthoritySession>;
    if (
      !parsed ||
      typeof parsed.id !== "string" ||
      typeof parsed.role !== "string" ||
      !commandRoles.includes(parsed.role as AgencyUserRole)
    ) {
      return null;
    }
    const role = parsed.role as AgencyUserRole;
    const tokenExpiresAt =
      typeof parsed.tokenExpiresAt === "string"
        ? parsed.tokenExpiresAt
        : undefined;
    // An expired token can only produce 401s — force a fresh sign-in instead.
    if (
      tokenExpiresAt &&
      !Number.isNaN(Date.parse(tokenExpiresAt)) &&
      Date.parse(tokenExpiresAt) <= Date.now()
    ) {
      return null;
    }
    return normalizeSession({
      id: parsed.id,
      name: parsed.name ?? roleLabels[role],
      role,
      agencyId: parsed.agencyId ?? DEFAULT_AGENCY_ID,
      agency: parsed.agency ?? agencyByRole[role],
      district: parsed.district ?? DEFAULT_DISTRICT,
      mfaCompleted: Boolean(parsed.mfaCompleted),
      accessToken:
        typeof parsed.accessToken === "string" ? parsed.accessToken : undefined,
      tokenExpiresAt,
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

export const defaultAccountPreferences: AuthorityAccountPreferences = {
  inAppAlerts: true,
  criticalSms: true,
  approvalEmail: true,
  compactTables: false,
  reducedMotion: false,
};

function isAccountPreferences(
  value: unknown,
): value is Partial<AuthorityAccountPreferences> {
  if (!value || typeof value !== "object") {
    return false;
  }
  return (
    Object.keys(defaultAccountPreferences) as Array<
      keyof AuthorityAccountPreferences
    >
  ).every((key) => {
    const candidate = (value as Record<string, unknown>)[key];
    return candidate === undefined || typeof candidate === "boolean";
  });
}

function readStoredPreferences(): AuthorityAccountPreferences {
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

let currentSession: AuthoritySession | null = readStoredSession();
let currentPreferences: AuthorityAccountPreferences = readStoredPreferences();
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

function persist(session: AuthoritySession | null) {
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

function persistPreferences(preferences: AuthorityAccountPreferences) {
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

function getSnapshot(): AuthoritySession | null {
  return currentSession;
}

function getPreferencesSnapshot(): AuthorityAccountPreferences {
  return currentPreferences;
}

export function getAuthoritySession(): AuthoritySession | null {
  return currentSession;
}

export function getAuthorityPreferences(): AuthorityAccountPreferences {
  return currentPreferences;
}

export function signInAuthority(session: AuthoritySession) {
  currentSession = normalizeSession(session);
  persist(currentSession);
  emit();
}

export function signOutAuthority() {
  currentSession = null;
  persist(null);
  emit();
}

/**
 * Update the signed-in operator's editable profile fields. Email stays fixed —
 * it is presented as administrator-managed — so only name and department move.
 * TODO: wire to real profile/password/MFA API.
 */
export function updateAuthorityProfile(patch: AuthorityProfilePatch) {
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
export function updateAuthorityPreferences(
  patch: Partial<AuthorityAccountPreferences>,
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
export function setAuthorityMfaEnabled(enabled: boolean) {
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
export function changeAuthorityPassword(
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
 * Reactive access to the signed-in authority plus account preferences and the
 * store actions. `session` is null until an operator completes sign-in + MFA.
 */
export function useAuthoritySession(): AuthoritySessionState {
  const session = useSyncExternalStore(subscribe, getSnapshot, () => null);
  const preferences = useSyncExternalStore(
    subscribePreferences,
    getPreferencesSnapshot,
    () => defaultAccountPreferences,
  );
  return {
    session,
    preferences,
    updateProfile: updateAuthorityProfile,
    updatePreferences: updateAuthorityPreferences,
    setMfaEnabled: setAuthorityMfaEnabled,
    changePassword: changeAuthorityPassword,
  };
}

export function hasCommandAccess(session: AuthoritySession | null): boolean {
  return Boolean(
    session && session.mfaCompleted && commandRoles.includes(session.role),
  );
}

export function authorityHeaders(): Record<string, string> {
  const session = currentSession ?? fallbackSession;
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "X-NADAA-Actor-ID": session.id,
    "X-NADAA-Actor-Role": session.role,
    "X-NADAA-Agency-ID": session.agencyId,
    "X-NADAA-Actor-District": session.district,
    "X-NADAA-MFA-Completed": session.mfaCompleted ? "true" : "false",
    "X-NADAA-Request-ID": `authority-ui-${Date.now()}`,
  };
  // Real bearer token from agency login; services require it on authority
  // endpoints. The X-NADAA-* actor headers above stay for mock-actors dev mode.
  if (session.accessToken) {
    headers.Authorization = `Bearer ${session.accessToken}`;
  }
  return headers;
}
