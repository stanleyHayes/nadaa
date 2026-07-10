import { useSyncExternalStore } from "react";
import type { AgencyUserRole } from "@nadaa/shared-types";

/**
 * Roles allowed to reach the governance console. Kept in sync with the
 * feature-level RBAC gate (`features/administration/rbac.ts`); the sign-in
 * screen only offers these roles and the gate re-checks on every render.
 */
export const adminSignInRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
];

export type AdminSession = {
  id: string;
  name: string;
  role: AgencyUserRole;
  agencyId: string;
  agency: string;
  mfaCompleted: boolean;
};

/**
 * Default agency identity per admin role. Used to pre-fill the sign-in screen;
 * a live deployment would resolve this from the directory service.
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
const DEFAULT_AGENCY_ID = "00000000-0000-0000-0000-000000000101";

/**
 * Fallback identity used only for request headers if a call somehow fires
 * before an admin signs in. The UI never renders governance surfaces without
 * an authenticated session, so this keeps API clients resilient in dev.
 */
const fallbackSession: AdminSession = {
  id: "usr_system_admin",
  name: "NADAA System Admin",
  role: "system_admin",
  agencyId: DEFAULT_AGENCY_ID,
  agency: "NADMO National Operations",
  mfaCompleted: true,
};

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
      typeof parsed.role !== "string"
    ) {
      return null;
    }
    const role = parsed.role as AgencyUserRole;
    return {
      id: parsed.id,
      name: parsed.name ?? roleLabels[role] ?? "Administrator",
      role,
      agencyId: parsed.agencyId ?? DEFAULT_AGENCY_ID,
      agency: parsed.agency ?? agencyByRole[role] ?? "NADAA",
      mfaCompleted: Boolean(parsed.mfaCompleted),
    };
  } catch {
    return null;
  }
}

let currentSession: AdminSession | null = readStoredSession();
const listeners = new Set<() => void>();

function emit() {
  for (const listener of listeners) {
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

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getSnapshot(): AdminSession | null {
  return currentSession;
}

export function getAdminSession(): AdminSession | null {
  return currentSession;
}

export function signInAdmin(session: AdminSession) {
  currentSession = session;
  persist(session);
  emit();
}

export function signOutAdmin() {
  currentSession = null;
  persist(null);
  emit();
}

/**
 * Reactive access to the signed-in administrator. Returns null until an admin
 * completes the sign-in + MFA flow.
 */
export function useAdminSession(): AdminSession | null {
  return useSyncExternalStore(subscribe, getSnapshot, () => null);
}

export function adminHeaders() {
  const session = currentSession ?? fallbackSession;
  return {
    Authorization: `Bearer local-admin-token`,
    "Content-Type": "application/json",
    "X-NADAA-Actor-ID": session.id,
    "X-NADAA-Actor-Role": session.role,
    "X-NADAA-Agency-ID": session.agencyId,
    "X-NADAA-MFA-Completed": session.mfaCompleted ? "true" : "false",
    "X-NADAA-Request-ID": `admin-web-${Date.now()}`,
  };
}
