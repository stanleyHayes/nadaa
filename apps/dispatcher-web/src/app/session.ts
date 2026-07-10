import { useSyncExternalStore } from "react";
import type { AgencyUserRole } from "@nadaa/shared-types";

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

export type DispatcherSession = {
  id: string;
  name: string;
  role: AgencyUserRole;
  agencyId: string;
  agency: string;
  district: string;
  mfaCompleted: boolean;
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
const DEFAULT_AGENCY_ID = "00000000-0000-0000-0000-000000000101";
const DEFAULT_DISTRICT = "Accra Metropolitan";

/**
 * Fallback identity used only for request headers if a call somehow fires
 * before a controller signs in. The UI never renders console surfaces without
 * an authenticated session, so this keeps API clients resilient in dev.
 */
const fallbackSession: DispatcherSession = {
  id: "usr_dispatch_accra",
  name: "Accra Dispatcher",
  role: "dispatcher",
  agencyId: DEFAULT_AGENCY_ID,
  agency: "NADMO Accra Dispatch Desk",
  district: DEFAULT_DISTRICT,
  mfaCompleted: true,
};

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
    return {
      id: parsed.id,
      name: parsed.name ?? roleLabels[role],
      role,
      agencyId: parsed.agencyId ?? DEFAULT_AGENCY_ID,
      agency: parsed.agency ?? agencyByRole[role],
      district: parsed.district ?? DEFAULT_DISTRICT,
      mfaCompleted: Boolean(parsed.mfaCompleted),
    };
  } catch {
    return null;
  }
}

let currentSession: DispatcherSession | null = readStoredSession();
const listeners = new Set<() => void>();

function emit() {
  for (const listener of listeners) {
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

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getSnapshot(): DispatcherSession | null {
  return currentSession;
}

export function getDispatcherSession(): DispatcherSession | null {
  return currentSession;
}

export function signInDispatcher(session: DispatcherSession) {
  currentSession = session;
  persist(session);
  emit();
}

export function signOutDispatcher() {
  currentSession = null;
  persist(null);
  emit();
}

/**
 * Reactive access to the signed-in dispatcher. Returns null until a controller
 * completes the sign-in + MFA flow.
 */
export function useDispatcherSession(): DispatcherSession | null {
  return useSyncExternalStore(subscribe, getSnapshot, () => null);
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
    "X-NADAA-Actor-ID": session.id,
    "X-NADAA-Actor-Role": session.role,
    "X-NADAA-Agency-ID": session.agencyId,
    "X-NADAA-Actor-District": session.district,
    "X-NADAA-MFA-Completed": session.mfaCompleted ? "true" : "false",
    "X-NADAA-Request-ID": `dispatcher-web-${Date.now()}`,
  };
}
