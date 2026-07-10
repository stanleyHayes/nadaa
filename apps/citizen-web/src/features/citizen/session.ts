import { useSyncExternalStore } from "react";

/**
 * Shared citizen session store. Viewing the platform (risk, alerts, shelters,
 * guides, open data) is public and needs no account. SUBMITTING anything —
 * incident reports, damage claims, aid pledges, missing-person reports — now
 * requires a signed-in citizen; anonymous submissions are not allowed. The
 * session is a single global store so every page/panel sees the same state
 * reactively, and any submission surface can open the sign-in dialog. Persisted
 * to localStorage; there is no backend auth here yet.
 */
export type CitizenSession = {
  name: string;
  phone: string;
  region: string;
  language: string;
  since: string;
};

export type SavedReport = {
  reference: string;
  hazard: string;
  urgency: string;
  priorityReview: boolean;
  at: string;
};

const SESSION_KEY = "nadaa.citizen.session.v1";
const REPORTS_KEY = "nadaa.citizen.savedReports.v1";

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

type StoreState = {
  session: CitizenSession | null;
  savedReports: SavedReport[];
  signInOpen: boolean;
};

let state: StoreState = {
  session: readJSON(SESSION_KEY, isSession),
  savedReports: readJSON(REPORTS_KEY, isSavedReports) ?? [],
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

/** Sign out and clear the persisted session. */
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
 * gate a submission on it, and call `requestSignIn()` to open the dialog.
 */
export function useCitizenSession() {
  const snapshot = useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
  return {
    session: snapshot.session,
    savedReports: snapshot.savedReports,
    signInOpen: snapshot.signInOpen,
    signIn: signInCitizen,
    signOut: signOutCitizen,
    saveReport: saveCitizenReport,
    requestSignIn,
    closeSignIn,
  };
}
