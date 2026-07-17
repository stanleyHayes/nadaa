import { hazardRoles, severityRoles } from "@nadaa/brand";
import type {
  AlertSeverity,
  AreaRiskResponse,
  CitizenAlertFeedItem,
  EmergencyGuideRecord,
  GuideStage,
  HazardType,
  IncidentMediaContentType,
  InitiateMediaUploadRequest,
  MediaUploadResponse,
  ReliefStockCategory,
  RiskLevel,
  ShelterStatus,
} from "@nadaa/shared-types";
import { GUIDE_CACHE_KEY, INCIDENT_API_BASE } from "../../app/config";
import { areaPresets, guideLanguageOptions, supportedMediaTypes } from "./data";
import type { GuideCachePayload, GuideFilters } from "./types";

export function resolveAreaLookup(value: string) {
  const normalized = value.trim().toLowerCase();
  const preset = areaPresets.find(
    (item) => item.label.toLowerCase() === normalized,
  );
  if (preset) {
    return preset;
  }

  const partialPreset = areaPresets.find((item) =>
    item.label.toLowerCase().includes(normalized),
  );
  if (partialPreset && normalized.length >= 4) {
    return partialPreset;
  }

  const [latText, lngText] = value.split(",").map((part) => part.trim());
  const lat = Number(latText);
  const lng = Number(lngText);
  if (
    Number.isFinite(lat) &&
    Number.isFinite(lng) &&
    lat >= -90 &&
    lat <= 90 &&
    lng >= -180 &&
    lng <= 180
  ) {
    return { label: `${lat.toFixed(4)}, ${lng.toFixed(4)}`, lat, lng };
  }

  return null;
}

export function alertSeverityTone(
  severity: AlertSeverity,
  status: CitizenAlertFeedItem["status"],
): "success" | "warning" | "error" | "info" {
  if (status === "expired") {
    return "info";
  }
  if (severity === "emergency" || severity === "severe_warning") {
    return "error";
  }
  if (severity === "warning" || severity === "watch") {
    return "warning";
  }
  return "info";
}

export function alertSeverityLabel(severity: AlertSeverity): string {
  return severity
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function alertStatusLabel(
  status: CitizenAlertFeedItem["status"],
): string {
  return status.charAt(0).toUpperCase() + status.slice(1);
}

export function hazardLabel(hazard: HazardType): string {
  return hazard
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function severityRoleFor(
  level: RiskLevel | AlertSeverity,
): keyof typeof severityRoles {
  if (level === "moderate" || level === "advisory") return "medium";
  if (level === "emergency" || level === "severe_warning") return "severe";
  if (level === "watch" || level === "warning") return "high";
  if (level === "low") return "low";
  if (level === "high") return "high";
  if (level === "severe") return "severe";
  return "info";
}

export function hazardRoleFor(hazard: HazardType): keyof typeof hazardRoles {
  switch (hazard) {
    case "flood":
    case "blocked_drain":
    case "tidal_wave":
      return "flood";
    case "fire":
      return "fire";
    case "road_crash":
      return "road";
    case "medical_emergency":
      return "medical";
    case "building_collapse":
    case "landslide":
      return "geological";
    case "disease_outbreak":
      return "disease";
    case "storm":
      return "storm";
    default:
      return "default";
  }
}

export { hazardRoles, severityRoles };

export function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("en-GH", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function filterGuides(
  guides: EmergencyGuideRecord[],
  filters: GuideFilters,
): EmergencyGuideRecord[] {
  const stageAndHazardMatches = guides.filter((guide) => {
    if (filters.hazard !== "all" && guide.hazardType !== filters.hazard) {
      return false;
    }
    if (filters.stage !== "all" && guide.stage !== filters.stage) {
      return false;
    }
    return true;
  });

  const languageMatches = stageAndHazardMatches.filter(
    (guide) => guide.language === filters.language,
  );
  const fallbackMatches =
    languageMatches.length > 0 || filters.language === "en"
      ? languageMatches
      : stageAndHazardMatches.filter((guide) => guide.language === "en");

  return [...fallbackMatches].sort((a, b) => {
    if (a.sortOrder === b.sortOrder) {
      return a.title.localeCompare(b.title);
    }
    return a.sortOrder - b.sortOrder;
  });
}

const GUIDE_CACHE_DB_NAME = "nadaa-citizen";
const GUIDE_CACHE_DB_STORE = "guideCache";

/**
 * Open the guide-cache IndexedDB, creating the object store on first use.
 * Resolves to null when IDB is unavailable (private/restricted modes, older
 * browsers) so callers can fall back to localStorage.
 */
function openGuideCacheDb(): Promise<IDBDatabase | null> {
  if (typeof window === "undefined" || !("indexedDB" in window)) {
    return Promise.resolve(null);
  }
  return new Promise((resolve) => {
    let request: IDBOpenDBRequest;
    try {
      request = window.indexedDB.open(GUIDE_CACHE_DB_NAME, 1);
    } catch {
      resolve(null);
      return;
    }
    request.onupgradeneeded = () => {
      request.result.createObjectStore(GUIDE_CACHE_DB_STORE);
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => resolve(null);
    request.onblocked = () => resolve(null);
  });
}

function idbReadGuideCache(db: IDBDatabase): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const request = db
      .transaction(GUIDE_CACHE_DB_STORE, "readonly")
      .objectStore(GUIDE_CACHE_DB_STORE)
      .get(GUIDE_CACHE_KEY);
    request.onsuccess = () => resolve(request.result);
    request.onerror = () =>
      reject(request.error ?? new Error("guide cache read failed"));
  });
}

function idbWriteGuideCache(
  db: IDBDatabase,
  payload: GuideCachePayload,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const transaction = db.transaction(GUIDE_CACHE_DB_STORE, "readwrite");
    transaction.objectStore(GUIDE_CACHE_DB_STORE).put(payload, GUIDE_CACHE_KEY);
    transaction.oncomplete = () => resolve();
    transaction.onerror = () =>
      reject(transaction.error ?? new Error("guide cache write failed"));
    transaction.onabort = () =>
      reject(transaction.error ?? new Error("guide cache write aborted"));
  });
}

function readGuideCacheFromLocalStorage(): GuideCachePayload | null {
  try {
    const raw = window.localStorage.getItem(GUIDE_CACHE_KEY);
    if (!raw) {
      return null;
    }
    const payload = JSON.parse(raw) as GuideCachePayload;
    if (!isGuideCachePayload(payload)) {
      return null;
    }
    return payload;
  } catch {
    return null;
  }
}

/**
 * Read the offline guide cache. IndexedDB is the primary store; localStorage
 * is the fallback when IDB is unavailable and still holds caches written
 * before the IDB migration (the GUIDE_CACHE_KEY contract is unchanged).
 */
export async function readGuideCache(): Promise<GuideCachePayload | null> {
  if (typeof window === "undefined") {
    return null;
  }

  const db = await openGuideCacheDb();
  if (db) {
    try {
      const cached = await idbReadGuideCache(db);
      if (isGuideCachePayload(cached)) {
        return cached;
      }
    } catch {
      // Fall through to the localStorage copy below.
    } finally {
      db.close();
    }
  }

  return readGuideCacheFromLocalStorage();
}

/**
 * Persist the offline-available guides. Writes go to IndexedDB when available,
 * falling back to localStorage otherwise; a successful IDB write drops the
 * legacy localStorage copy so there is a single source of truth.
 */
export async function writeGuideCache(
  guides: EmergencyGuideRecord[],
  language: string,
  cachedAt: string,
): Promise<void> {
  if (typeof window === "undefined") {
    return;
  }

  const offlineGuides = guides.filter((guide) => guide.offlineAvailable);
  if (!offlineGuides.length) {
    return;
  }
  const payload: GuideCachePayload = { guides: offlineGuides, language, cachedAt };

  const db = await openGuideCacheDb();
  if (db) {
    try {
      await idbWriteGuideCache(db, payload);
      try {
        window.localStorage.removeItem(GUIDE_CACHE_KEY);
      } catch {
        // Storage can be unavailable in private or restricted browser modes.
      }
      return;
    } catch {
      // Fall through to the localStorage fallback below.
    } finally {
      db.close();
    }
  }

  try {
    window.localStorage.setItem(GUIDE_CACHE_KEY, JSON.stringify(payload));
  } catch {
    // Local storage can be unavailable in private or restricted browser modes.
  }
}

export function registerCitizenServiceWorker() {
  if (
    typeof window === "undefined" ||
    !("serviceWorker" in navigator) ||
    !import.meta.env.PROD
  ) {
    return;
  }

  navigator.serviceWorker.register("/sw.js").catch(() => undefined);
}

export function isGuideCachePayload(
  value: unknown,
): value is GuideCachePayload {
  return (
    typeof value === "object" &&
    value !== null &&
    Array.isArray((value as GuideCachePayload).guides) &&
    typeof (value as GuideCachePayload).cachedAt === "string" &&
    typeof (value as GuideCachePayload).language === "string" &&
    (value as GuideCachePayload).guides.every(
      (guide) =>
        typeof guide.id === "string" &&
        typeof guide.title === "string" &&
        typeof guide.body === "string" &&
        typeof guide.language === "string",
    )
  );
}

export function guideStageLabel(stage: GuideStage): string {
  return stage.charAt(0).toUpperCase() + stage.slice(1);
}

export function guideLanguageLabel(language: string): string {
  return (
    guideLanguageOptions.find((option) => option.value === language)?.label ??
    language.toUpperCase()
  );
}

export function formatOccupancy(shelter: {
  capacity?: number;
  currentOccupancy?: number;
  status?:
    ShelterStatus | AreaRiskResponse["nearestShelters"][number]["status"];
}): string {
  if (
    typeof shelter.currentOccupancy === "number" &&
    typeof shelter.capacity === "number"
  ) {
    return `${shelter.currentOccupancy}/${shelter.capacity} occupied`;
  }

  return shelter.status ? shelter.status : "Shelter status unavailable";
}

export function formatSupportType(value: string): string {
  return value
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function formatListLabel(values: string[]): string {
  return values.map(formatSupportType).join(" · ");
}

export function formatReliefStock(values: ReliefStockCategory[]): string {
  if (!values.length) {
    return "Stock details pending";
  }

  return values
    .slice(0, 3)
    .map(
      (item) =>
        `${formatSupportType(item.category)}: ${item.quantity.toLocaleString("en-GH")} ${item.unit}`,
    )
    .join(" · ");
}

export function formatDistance(meters: number): string {
  if (meters < 1000) {
    return `${Math.max(1, Math.round(meters))} m`;
  }

  return `${(meters / 1000).toFixed(1)} km`;
}

export async function initiateMediaUploads(files: File[]): Promise<string[]> {
  const mediaIds: string[] = [];

  for (const file of files) {
    if (!supportedMediaTypes.includes(file.type as IncidentMediaContentType)) {
      throw new Error(`${file.name} is not a supported media type.`);
    }

    const payload: InitiateMediaUploadRequest = {
      purpose: "incident_media",
      fileName: file.name,
      contentType: file.type as IncidentMediaContentType,
      sizeBytes: file.size,
      // uploadedBy is optional server-side and deliberately omitted: there is
      // no real session identity, so sending a fabricated id would corrupt
      // attribution data.
    };

    const response = await fetch(`${INCIDENT_API_BASE}/media/uploads`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }

    const upload = (await response.json()) as MediaUploadResponse;

    // PUT the file bytes to the returned upload URL. The backend's URL is a
    // dev stub with nothing real listening behind it, so a failed byte
    // transfer is tolerated — the media record (and its id) is registered.
    try {
      const byteResponse = await fetch(upload.uploadUrl, {
        method: upload.method,
        headers: upload.headers,
        body: file,
      });
      if (!byteResponse.ok) {
        console.warn(
          `Media byte upload for ${file.name} returned ${byteResponse.status}; continuing with the registered media id.`,
        );
      }
    } catch (error) {
      console.warn(
        `Media byte upload for ${file.name} failed; continuing with the registered media id.`,
        error,
      );
    }

    mediaIds.push(upload.mediaId);
  }

  return mediaIds;
}

export async function extractAPIError(response: Response): Promise<string> {
  try {
    const payload = (await response.json()) as { error?: { message?: string } };
    return (
      payload.error?.message ?? `Request failed with status ${response.status}`
    );
  } catch {
    return `Request failed with status ${response.status}`;
  }
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024 * 1024) {
    return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  }

  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
