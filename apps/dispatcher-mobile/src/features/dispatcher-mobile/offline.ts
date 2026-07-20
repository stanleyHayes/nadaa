import {
  CAPACITY_CACHE_KEY,
  INCIDENT_CACHE_KEY,
  SELECTED_INCIDENT_KEY,
  SESSION_KEY,
} from "../../app/config";
import { defaultCapacityFilters, defaultFilters } from "./data";
import type {
  CapacityCachePayload,
  CapacityFilterState,
  DispatcherSession,
  IncidentCachePayload,
  IncidentFilterState,
} from "./types";

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

/**
 * Read the persisted session. Empty or corrupt storage means signed out —
 * fixture sessions must never pose as an authenticated dispatcher.
 */
export async function readSession(
  storage: KeyValueStorage,
): Promise<DispatcherSession | null> {
  const raw = await storage.getItem(SESSION_KEY);
  if (!raw) {
    return null;
  }
  try {
    const payload = JSON.parse(raw) as DispatcherSession;
    if (
      typeof payload.userId === "string" &&
      typeof payload.agencyId === "string" &&
      typeof payload.role === "string"
    ) {
      return payload;
    }
  } catch {
    await storage.removeItem(SESSION_KEY);
  }
  return null;
}

export async function writeSession(
  storage: KeyValueStorage,
  session: DispatcherSession,
) {
  await storage.setItem(SESSION_KEY, JSON.stringify(session));
}

export async function clearSession(storage: KeyValueStorage) {
  await storage.removeItem(SESSION_KEY);
}

export async function readIncidentCache(
  storage: KeyValueStorage,
): Promise<IncidentCachePayload> {
  const raw = await storage.getItem(INCIDENT_CACHE_KEY);
  if (!raw) {
    return {
      cachedAt: new Date().toISOString(),
      incidents: [],
    };
  }
  try {
    const payload = JSON.parse(raw) as IncidentCachePayload;
    if (
      Array.isArray(payload.incidents) &&
      typeof payload.cachedAt === "string"
    ) {
      return payload;
    }
  } catch {
    await storage.removeItem(INCIDENT_CACHE_KEY);
  }
  // Corrupt cache: an honest empty queue, never fixture incidents.
  return {
    cachedAt: new Date().toISOString(),
    incidents: [],
  };
}

export async function writeIncidentCache(
  storage: KeyValueStorage,
  payload: IncidentCachePayload,
) {
  await storage.setItem(INCIDENT_CACHE_KEY, JSON.stringify(payload));
}

export async function readSelectedIncidentId(
  storage: KeyValueStorage,
): Promise<string | null> {
  const raw = await storage.getItem(SELECTED_INCIDENT_KEY);
  if (!raw) {
    return null;
  }
  try {
    const payload = JSON.parse(raw) as { id: string };
    if (typeof payload.id === "string") {
      return payload.id;
    }
  } catch {
    await storage.removeItem(SELECTED_INCIDENT_KEY);
  }
  return null;
}

export async function writeSelectedIncidentId(
  storage: KeyValueStorage,
  id: string | null,
) {
  if (id == null) {
    await storage.removeItem(SELECTED_INCIDENT_KEY);
    return;
  }
  await storage.setItem(SELECTED_INCIDENT_KEY, JSON.stringify({ id }));
}

export async function readCapacityCache(
  storage: KeyValueStorage,
): Promise<CapacityCachePayload> {
  const raw = await storage.getItem(CAPACITY_CACHE_KEY);
  if (!raw) {
    return {
      cachedAt: new Date().toISOString(),
      facilities: [],
    };
  }
  try {
    const payload = JSON.parse(raw) as CapacityCachePayload;
    if (
      Array.isArray(payload.facilities) &&
      typeof payload.cachedAt === "string"
    ) {
      return payload;
    }
  } catch {
    await storage.removeItem(CAPACITY_CACHE_KEY);
  }
  // Corrupt cache: an honest empty list, never fixture facilities.
  return {
    cachedAt: new Date().toISOString(),
    facilities: [],
  };
}

export async function writeCapacityCache(
  storage: KeyValueStorage,
  payload: CapacityCachePayload,
) {
  await storage.setItem(CAPACITY_CACHE_KEY, JSON.stringify(payload));
}

export function readFiltersFromCache(
  storage: KeyValueStorage,
): Promise<IncidentFilterState> {
  return Promise.resolve(defaultFilters);
}

export function writeFiltersToCache(
  storage: KeyValueStorage,
  filters: IncidentFilterState,
): Promise<void> {
  return Promise.resolve();
}

export function readCapacityFiltersFromCache(
  storage: KeyValueStorage,
): Promise<CapacityFilterState> {
  return Promise.resolve(defaultCapacityFilters);
}

export function writeCapacityFiltersToCache(
  storage: KeyValueStorage,
  filters: CapacityFilterState,
): Promise<void> {
  return Promise.resolve();
}
