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

export function readGuideCache(): GuideCachePayload | null {
  if (typeof window === "undefined") {
    return null;
  }

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

export function writeGuideCache(
  guides: EmergencyGuideRecord[],
  language: string,
  cachedAt: string,
) {
  if (typeof window === "undefined") {
    return;
  }

  const offlineGuides = guides.filter((guide) => guide.offlineAvailable);
  if (!offlineGuides.length) {
    return;
  }

  try {
    window.localStorage.setItem(
      GUIDE_CACHE_KEY,
      JSON.stringify({ guides: offlineGuides, language, cachedAt }),
    );
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

export function formatOccupancy(
  shelter: AreaRiskResponse["nearestShelters"][number],
): string {
  if (
    typeof shelter.currentOccupancy === "number" &&
    typeof shelter.capacity === "number"
  ) {
    return `${shelter.currentOccupancy}/${shelter.capacity} occupied`;
  }

  return shelter.status ? shelter.status : "Shelter status unavailable";
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
      uploadedBy: "usr_demo_citizen",
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
