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
const DEFAULT_AGENCY_ID = "00000000-0000-0000-0000-000000000101";
const DEFAULT_DISTRICT = "Accra Metropolitan";

/**
 * Fallback identity used only for request headers if a call somehow fires
 * before an operator signs in. The UI never renders agency surfaces without an
 * authenticated session, so this keeps API clients resilient in dev.
 */
const fallbackSession: AgencySession = {
  id: "usr_agency_responder_001",
  name: "NADMO Responder",
  role: "responder",
  agencyId: DEFAULT_AGENCY_ID,
  agency: "NADMO Accra Metro",
  district: DEFAULT_DISTRICT,
  token: "local-agency-token",
  mfaCompleted: true,
};

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
    return {
      id: parsed.id,
      name: parsed.name ?? roleLabels[role],
      role,
      agencyId: parsed.agencyId ?? DEFAULT_AGENCY_ID,
      agency: parsed.agency ?? agencyByRole[role],
      district: parsed.district ?? DEFAULT_DISTRICT,
      token: parsed.token ?? `agency-${role}`,
      mfaCompleted: Boolean(parsed.mfaCompleted),
    };
  } catch {
    return null;
  }
}

let currentSession: AgencySession | null = readStoredSession();
const listeners = new Set<() => void>();

function emit() {
  for (const listener of listeners) {
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

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getSnapshot(): AgencySession | null {
  return currentSession;
}

export function getAgencySession(): AgencySession | null {
  return currentSession;
}

export function signInAgency(session: AgencySession) {
  currentSession = session;
  persist(session);
  emit();
}

export function signOutAgency() {
  currentSession = null;
  persist(null);
  emit();
}

/**
 * Reactive access to the signed-in agency operator. Returns null until an
 * operator completes the sign-in + MFA flow.
 */
export function useAgencySession(): AgencySession | null {
  return useSyncExternalStore(subscribe, getSnapshot, () => null);
}

export function hasAgencyAccess(session: AgencySession | null): boolean {
  return Boolean(
    session && session.mfaCompleted && agencyRoles.includes(session.role),
  );
}

export function agencyHeaders() {
  const session = currentSession ?? fallbackSession;
  return {
    Authorization: `Bearer ${session.token}`,
    "Content-Type": "application/json",
    "X-NADAA-Actor-ID": session.id,
    "X-NADAA-Actor-Role": session.role,
    "X-NADAA-Agency-ID": session.agencyId,
    "X-NADAA-Actor-District": session.district,
    "X-NADAA-MFA-Completed": session.mfaCompleted ? "true" : "false",
    "X-NADAA-Request-ID": `agency-web-${Date.now()}`,
  };
}
