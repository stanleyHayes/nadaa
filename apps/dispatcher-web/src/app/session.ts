import { useSyncExternalStore } from "react";
import type { AgencyUserRole } from "@nadaa/shared-types";
import { AUTH_API_BASE } from "./config";

/**
 * Roles allowed to reach the dispatch console. Field responders and viewers are
 * intentionally excluded: the console assigns teams and drafts alerts, which the
 * dispatch desk and its supervising officers own.
 */
export const commandRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
];

/**
 * Roles alert-service allows to approve or reject submitted alerts. Mirrors
 * ApprovalRoles in services/alert-service/internal/utils/utils.go — the server
 * enforces this; the console hides actions the server would reject.
 */
export const alertApprovalRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
];

/**
 * Roles alert-service allows to perform an emergency override. Mirrors
 * OverrideRoles in services/alert-service/internal/utils/utils.go.
 */
export const alertOverrideRoles: AgencyUserRole[] = [
  "system_admin",
  "nadmo_officer",
];

export type DispatcherSession = {
  id: string;
  name: string;
  role: AgencyUserRole;
  agencyId: string;
  agency: string;
  district: string;
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
  /** Bearer token returned by auth-service agency login. */
  accessToken?: string;
  /** ISO expiry of the access token (12h TTL from auth-service). */
  tokenExpiresAt?: string;
};

/** Patch accepted by {@link updateDispatcherProfile}. */
export type DispatcherProfilePatch = {
  name?: string;
  department?: string;
};

/** Result of the {@link changeDispatcherPassword} action. */
export type PasswordChangeResult = {
  ok: boolean;
  message: string;
};

/**
 * Browser-scoped account preferences. These are presentation choices only and
 * never gate access, so they live beside the session but persist separately.
 */
export type DispatcherAccountPreferences = {
  inAppAlerts: boolean;
  criticalSms: boolean;
  approvalEmail: boolean;
  compactTables: boolean;
  reducedMotion: boolean;
};

/** Bundle returned by {@link useDispatcherSession}. */
export type DispatcherSessionState = {
  session: DispatcherSession | null;
  preferences: DispatcherAccountPreferences;
  updateProfile: (patch: DispatcherProfilePatch) => void;
  updatePreferences: (patch: Partial<DispatcherAccountPreferences>) => void;
  setMfaEnabled: (enabled: boolean) => void;
  changePassword: (current: string, next: string) => Promise<PasswordChangeResult>;
};

/**
 * Default agency identity per dispatch role. Used to pre-fill the sign-in
 * screen; a live deployment would resolve this from the directory service.
 */
export const agencyByRole: Record<AgencyUserRole, string> = {
  system_admin: "NADAA National Command",
  agency_admin: "NADMO National Dispatch",
  nadmo_officer: "NADMO Accra Metro",
  district_officer: "Accra Metropolitan Assembly",
  dispatcher: "NADMO Accra Dispatch Desk",
  responder: "Accra Regional Response Unit",
  agency_viewer: "Partner Agency Liaison",
};

/** Default operating unit per dispatch role. */
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
  dispatcher: "Dispatch controller",
  responder: "Field responder",
  agency_viewer: "Agency viewer",
};

const STORAGE_KEY = "nadaa.dispatcher.session.v1";
const PREFERENCES_KEY = "nadaa.dispatcher.account-settings.v1";
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
  return `${slug || "dispatcher"}@nadmo.gov.gh`;
}

function normalizeStatus(value: unknown): "active" | "suspended" {
  return value === "suspended" ? "suspended" : "active";
}

/**
 * Fill the optional account fields from the required identity so any session —
 * including the bare object the sign-in screen builds — renders completely.
 */
function normalizeSession(base: DispatcherSession): DispatcherSession {
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
 * before a controller signs in. The UI never renders console surfaces without
 * an authenticated session, so this keeps API clients resilient in dev.
 */
const fallbackSession: DispatcherSession = normalizeSession({
  id: "usr_dispatch_accra",
  name: "Accra Dispatcher",
  role: "dispatcher",
  agencyId: DEFAULT_AGENCY_ID,
  agency: "NADMO Accra Dispatch Desk",
  district: DEFAULT_DISTRICT,
  mfaCompleted: true,
});

function readStoredSession(): DispatcherSession | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }
    const parsed = JSON.parse(raw) as Partial<DispatcherSession>;
    if (
      !parsed ||
      typeof parsed.id !== "string" ||
      typeof parsed.role !== "string" ||
      !commandRoles.includes(parsed.role as AgencyUserRole)
    ) {
      return null;
    }
    const role = parsed.role as AgencyUserRole;
    const session = normalizeSession({
      id: parsed.id,
      name: parsed.name ?? roleLabels[role],
      role,
      agencyId: parsed.agencyId ?? DEFAULT_AGENCY_ID,
      agency: parsed.agency ?? agencyByRole[role],
      district: parsed.district ?? DEFAULT_DISTRICT,
      mfaCompleted: Boolean(parsed.mfaCompleted),
      email: typeof parsed.email === "string" ? parsed.email : undefined,
      department:
        typeof parsed.department === "string" ? parsed.department : undefined,
      status: normalizeStatus(parsed.status),
      mfaEnabled:
        typeof parsed.mfaEnabled === "boolean" ? parsed.mfaEnabled : undefined,
      lastLoginAt:
        typeof parsed.lastLoginAt === "string" ? parsed.lastLoginAt : undefined,
      accessToken:
        typeof parsed.accessToken === "string" ? parsed.accessToken : undefined,
      tokenExpiresAt:
        typeof parsed.tokenExpiresAt === "string"
          ? parsed.tokenExpiresAt
          : undefined,
    });
    // An expired token cannot authorize anything; force a fresh sign-in.
    if (
      session.tokenExpiresAt &&
      new Date(session.tokenExpiresAt).getTime() <= Date.now()
    ) {
      return null;
    }
    return session;
  } catch {
    return null;
  }
}

export const defaultAccountPreferences: DispatcherAccountPreferences = {
  inAppAlerts: true,
  criticalSms: true,
  approvalEmail: true,
  compactTables: false,
  reducedMotion: false,
};

function isAccountPreferences(
  value: unknown,
): value is Partial<DispatcherAccountPreferences> {
  if (!value || typeof value !== "object") {
    return false;
  }
  return (
    Object.keys(defaultAccountPreferences) as Array<
      keyof DispatcherAccountPreferences
    >
  ).every((key) => {
    const candidate = (value as Record<string, unknown>)[key];
    return candidate === undefined || typeof candidate === "boolean";
  });
}

function readStoredPreferences(): DispatcherAccountPreferences {
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

let currentSession: DispatcherSession | null = readStoredSession();
let currentPreferences: DispatcherAccountPreferences = readStoredPreferences();
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

function persist(session: DispatcherSession | null) {
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

function persistPreferences(preferences: DispatcherAccountPreferences) {
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

function getSnapshot(): DispatcherSession | null {
  return currentSession;
}

function getPreferencesSnapshot(): DispatcherAccountPreferences {
  return currentPreferences;
}

export function getDispatcherSession(): DispatcherSession | null {
  return currentSession;
}

export function getDispatcherPreferences(): DispatcherAccountPreferences {
  return currentPreferences;
}

export function signInDispatcher(session: DispatcherSession) {
  currentSession = normalizeSession(session);
  persist(currentSession);
  emit();
}

export function signOutDispatcher() {
  currentSession = null;
  persist(null);
  emit();
}

/**
 * Update the signed-in controller's editable profile fields. Email stays fixed —
 * it is presented as administrator-managed — so only name and department move.
 * TODO: wire to real profile API.
 */
export function updateDispatcherProfile(patch: DispatcherProfilePatch) {
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
export function updateDispatcherPreferences(
  patch: Partial<DispatcherAccountPreferences>,
) {
  currentPreferences = { ...currentPreferences, ...patch };
  persistPreferences(currentPreferences);
  emitPreferences();
}

/**
 * Record whether an authenticator is enrolled, as confirmed by auth-service.
 * The SecurityTab calls this only after the MFA verify endpoint answers — it
 * never drives enrolment itself. This never clears the current session's
 * completed-MFA gate, so the flag cannot lock the controller out mid shift.
 */
export function setDispatcherMfaEnabled(enabled: boolean) {
  if (!currentSession) {
    return;
  }
  currentSession = { ...currentSession, mfaEnabled: enabled };
  persist(currentSession);
  emit();
}

/**
 * Change the signed-in controller's password through auth-service
 * (`POST /auth/agency/password`). The server enforces the current-password
 * check (401), payload validation (400), and rate limiting (429); only the
 * cheap format checks run locally.
 */
export async function changeDispatcherPassword(
  current: string,
  next: string,
): Promise<PasswordChangeResult> {
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
  if (!currentSession?.accessToken) {
    return { ok: false, message: "Sign in again to change your password." };
  }
  try {
    const response = await fetch(`${AUTH_API_BASE}/auth/agency/password`, {
      method: "POST",
      headers: dispatcherHeaders(),
      body: JSON.stringify({ currentPassword: current, newPassword: next }),
    });
    if (response.ok) {
      return {
        ok: true,
        message: "Password updated. Use it the next time you sign in.",
      };
    }
    if (response.status === 401) {
      return { ok: false, message: "Your current password is incorrect." };
    }
    if (response.status === 429) {
      return {
        ok: false,
        message: "Too many attempts. Try again in a few minutes.",
      };
    }
    const body = (await response.json().catch(() => null)) as {
      error?: { message?: string };
      message?: string;
    } | null;
    return {
      ok: false,
      message:
        body?.error?.message ??
        body?.message ??
        `Password change failed (${response.status}).`,
    };
  } catch {
    return {
      ok: false,
      message:
        "Auth service unavailable. Check your connection and try again.",
    };
  }
}

/**
 * Reactive access to the signed-in dispatcher plus account preferences and the
 * store actions. `session` is null until a controller completes sign-in + MFA.
 */
export function useDispatcherSession(): DispatcherSessionState {
  const session = useSyncExternalStore(subscribe, getSnapshot, () => null);
  const preferences = useSyncExternalStore(
    subscribePreferences,
    getPreferencesSnapshot,
    () => defaultAccountPreferences,
  );
  return {
    session,
    preferences,
    updateProfile: updateDispatcherProfile,
    updatePreferences: updateDispatcherPreferences,
    setMfaEnabled: setDispatcherMfaEnabled,
    changePassword: changeDispatcherPassword,
  };
}

export function hasCommandAccess(session: DispatcherSession | null): boolean {
  return Boolean(
    session && session.mfaCompleted && commandRoles.includes(session.role),
  );
}

export function dispatcherHeaders() {
  const session = currentSession ?? fallbackSession;
  return {
    "Content-Type": "application/json",
    // Real bearer auth; services require this for every authority endpoint.
    ...(session.accessToken
      ? { Authorization: `Bearer ${session.accessToken}` }
      : {}),
    // Mock-actor headers are honored only by dev services running with
    // NADAA_AUTH_ALLOW_MOCK_ACTORS=true; they mirror the token user.
    "X-NADAA-Actor-ID": session.id,
    "X-NADAA-Actor-Role": session.role,
    "X-NADAA-Agency-ID": session.agencyId,
    "X-NADAA-Actor-District": session.district,
    "X-NADAA-MFA-Completed": session.mfaCompleted ? "true" : "false",
    "X-NADAA-Request-ID": `dispatcher-web-${Date.now()}`,
  };
}
