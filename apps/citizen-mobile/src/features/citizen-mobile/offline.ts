import {
  GUIDE_CACHE_KEY,
  REPORT_DRAFT_KEY,
  SESSION_KEY,
  VOLUNTEER_PROFILE_KEY,
  VOLUNTEER_TASKS_KEY,
} from "../../app/config";
import {
  buildFallbackGuides,
  initialReportDraft,
  initialSession,
  sampleVolunteerProfile,
  sampleVolunteerTasks,
} from "./data";
import type { GuideCachePayload, MobileSession, ReportDraft } from "./types";
import type {
  VolunteerProfile,
  VolunteerTaskRecord,
} from "@nadaa/shared-types";

export type KeyValueStorage = {
  getItem(key: string): Promise<string | null>;
  removeItem(key: string): Promise<void>;
  setItem(key: string, value: string): Promise<void>;
};

export function createMemoryStorage(seed: Record<string, string> = {}) {
  const store = new Map(Object.entries(seed));
  return {
    async getItem(key: string) {
      return store.get(key) ?? null;
    },
    async removeItem(key: string) {
      store.delete(key);
    },
    async setItem(key: string, value: string) {
      store.set(key, value);
    },
  } satisfies KeyValueStorage;
}

export async function readGuideCache(
  storage: KeyValueStorage,
): Promise<GuideCachePayload> {
  const raw = await storage.getItem(GUIDE_CACHE_KEY);
  if (!raw) {
    return {
      cachedAt: new Date().toISOString(),
      guides: buildFallbackGuides(),
      language: "en",
    };
  }

  try {
    const payload = JSON.parse(raw) as GuideCachePayload;
    if (Array.isArray(payload.guides) && typeof payload.cachedAt === "string") {
      return payload;
    }
  } catch {
    await storage.removeItem(GUIDE_CACHE_KEY);
  }

  return {
    cachedAt: new Date().toISOString(),
    guides: buildFallbackGuides(),
    language: "en",
  };
}

export async function writeGuideCache(
  storage: KeyValueStorage,
  payload: GuideCachePayload,
) {
  const guides = payload.guides.filter((guide) => guide.offlineAvailable);
  await storage.setItem(
    GUIDE_CACHE_KEY,
    JSON.stringify({ ...payload, guides }),
  );
}

export async function readReportDraft(
  storage: KeyValueStorage,
): Promise<ReportDraft> {
  const raw = await storage.getItem(REPORT_DRAFT_KEY);
  if (!raw) {
    return initialReportDraft;
  }
  try {
    const payload = JSON.parse(raw) as ReportDraft;
    if (typeof payload.description === "string" && payload.hazard) {
      return payload;
    }
  } catch {
    await storage.removeItem(REPORT_DRAFT_KEY);
  }
  return initialReportDraft;
}

export async function writeReportDraft(
  storage: KeyValueStorage,
  draft: ReportDraft,
) {
  await storage.setItem(
    REPORT_DRAFT_KEY,
    JSON.stringify({ ...draft, savedAt: new Date().toISOString() }),
  );
}

export async function readSession(
  storage: KeyValueStorage,
): Promise<MobileSession> {
  const raw = await storage.getItem(SESSION_KEY);
  if (!raw) {
    return initialSession;
  }
  try {
    const payload = JSON.parse(raw) as MobileSession;
    if (
      typeof payload.userId === "string" &&
      typeof payload.phone === "string"
    ) {
      return payload;
    }
  } catch {
    await storage.removeItem(SESSION_KEY);
  }
  return initialSession;
}

export async function writeSession(
  storage: KeyValueStorage,
  session: MobileSession,
) {
  await storage.setItem(SESSION_KEY, JSON.stringify(session));
}

export async function readVolunteerProfile(
  storage: KeyValueStorage,
): Promise<VolunteerProfile> {
  const raw = await storage.getItem(VOLUNTEER_PROFILE_KEY);
  if (!raw) {
    return sampleVolunteerProfile;
  }
  try {
    const payload = JSON.parse(raw) as VolunteerProfile;
    if (typeof payload.id === "string" && typeof payload.groupId === "string") {
      return payload;
    }
  } catch {
    await storage.removeItem(VOLUNTEER_PROFILE_KEY);
  }
  return sampleVolunteerProfile;
}

export async function writeVolunteerProfile(
  storage: KeyValueStorage,
  profile: VolunteerProfile,
) {
  await storage.setItem(VOLUNTEER_PROFILE_KEY, JSON.stringify(profile));
}

export async function readVolunteerTasks(
  storage: KeyValueStorage,
): Promise<VolunteerTaskRecord[]> {
  const raw = await storage.getItem(VOLUNTEER_TASKS_KEY);
  if (!raw) {
    return sampleVolunteerTasks;
  }
  try {
    const payload = JSON.parse(raw) as VolunteerTaskRecord[];
    if (Array.isArray(payload)) {
      return payload;
    }
  } catch {
    await storage.removeItem(VOLUNTEER_TASKS_KEY);
  }
  return sampleVolunteerTasks;
}

export async function writeVolunteerTasks(
  storage: KeyValueStorage,
  tasks: VolunteerTaskRecord[],
) {
  await storage.setItem(VOLUNTEER_TASKS_KEY, JSON.stringify(tasks));
}
