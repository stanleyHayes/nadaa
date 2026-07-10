import { useCallback, useEffect, useState } from "react";

/**
 * Light, optional citizen sign-in. The citizen PWA is public and mostly
 * anonymous: a session only unlocks convenience (saved reports and claims) and
 * NEVER gates life-safety features. Everything is persisted to localStorage —
 * there is no backend auth here.
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

function readJSON<T>(key: string, guard: (value: unknown) => value is T): T | null {
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

/**
 * Central store for the optional citizen session plus the reports saved while
 * signed in. Returns stable callbacks and keeps localStorage in sync.
 */
export function useCitizenSession() {
  const [session, setSession] = useState<CitizenSession | null>(() =>
    readJSON(SESSION_KEY, isSession),
  );
  const [savedReports, setSavedReports] = useState<SavedReport[]>(
    () => readJSON(REPORTS_KEY, isSavedReports) ?? [],
  );

  useEffect(() => {
    if (session) {
      writeJSON(SESSION_KEY, session);
    } else if (typeof window !== "undefined") {
      window.localStorage.removeItem(SESSION_KEY);
    }
  }, [session]);

  useEffect(() => {
    writeJSON(REPORTS_KEY, savedReports);
  }, [savedReports]);

  const signIn = useCallback(
    (details: Omit<CitizenSession, "since">) => {
      setSession({ ...details, since: new Date().toISOString() });
    },
    [],
  );

  const signOut = useCallback(() => {
    setSession(null);
  }, []);

  const saveReport = useCallback((report: SavedReport) => {
    setSavedReports((current) => {
      if (current.some((item) => item.reference === report.reference)) {
        return current;
      }
      return [report, ...current].slice(0, 12);
    });
  }, []);

  return { session, savedReports, signIn, signOut, saveReport };
}
