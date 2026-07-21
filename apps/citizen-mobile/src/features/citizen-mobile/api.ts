import type {
  AreaRiskResponse,
  CitizenAlertFeedItem,
  CitizenAlertFeedResponse,
  CreateIncidentRequest,
  CreateIncidentResponse,
  EmergencyGuideRecord,
  GuideListResponse,
  NearbyShelterResponse,
  RegisterVolunteerRequest,
  VolunteerObservationRequest,
  VolunteerProfileResponse,
  VolunteerTaskListResponse,
  VolunteerTaskRecord,
  VolunteerTaskStatusRequest,
} from "@nadaa/shared-types";
import * as Notifications from "expo-notifications";
import {
  AUTH_API_BASE,
  GUIDE_API_BASE,
  INCIDENT_API_BASE,
  NOTIFICATION_API_BASE,
  PUSH_PROVIDER,
  RISK_API_BASE,
  SHELTER_API_BASE,
} from "../../app/config";
import { buildFallbackGuides, GUEST_PLACEHOLDER_PHONE } from "./data";
import type {
  CitizenLoginResult,
  CitizenOtpChallenge,
  MobileSession,
  PushRegistrationState,
  ReportDraft,
  VolunteerRegistrationDraft,
} from "./types";

/** API failure that keeps the HTTP status so callers can react to 401s. */
export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export function isAuthError(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}

export async function registerCitizen(request: {
  contactPermission: boolean;
  name: string;
  phone: string;
  preferredLanguage: string;
}): Promise<CitizenOtpChallenge> {
  const response = await fetch(`${AUTH_API_BASE}/auth/citizens/register`, {
    body: JSON.stringify(request),
    headers: { "Content-Type": "application/json" },
    method: "POST",
  });
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as CitizenOtpChallenge;
}

/** Returns null when the phone is not registered yet (404 phone_not_registered). */
export async function requestCitizenLoginOtp(
  phone: string,
): Promise<CitizenOtpChallenge | null> {
  const response = await fetch(`${AUTH_API_BASE}/auth/citizens/login/otp`, {
    body: JSON.stringify({ phone }),
    headers: { "Content-Type": "application/json" },
    method: "POST",
  });
  if (response.status === 404) {
    return null;
  }
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as CitizenOtpChallenge;
}

export async function loginCitizen(
  phone: string,
  otp: string,
): Promise<CitizenLoginResult> {
  const response = await fetch(`${AUTH_API_BASE}/auth/citizens/login`, {
    body: JSON.stringify({ otp, phone }),
    headers: { "Content-Type": "application/json" },
    method: "POST",
  });
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as CitizenLoginResult;
}

export async function fetchAreaRisk(
  lat: number,
  lng: number,
): Promise<AreaRiskResponse> {
  const response = await fetch(
    `${RISK_API_BASE}/risk?lat=${encodeURIComponent(lat)}&lng=${encodeURIComponent(lng)}`,
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as AreaRiskResponse;
}

/**
 * Live alert feed. Never substitutes fixtures: an empty feed returns [] and a
 * network/HTTP failure throws so the UI can show the offline state. Fixture-
 * sourced items are filtered out so they can never enter the live path.
 */
export async function fetchAlertFeed(): Promise<CitizenAlertFeedItem[]> {
  const response = await fetch(
    `${NOTIFICATION_API_BASE}/notifications/alerts?includeExpired=true`,
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  const payload = (await response.json()) as CitizenAlertFeedResponse;
  return payload.alerts.filter((alert) => alert.source !== "fixture");
}

export type GuideFetchResult = {
  guides: EmergencyGuideRecord[];
  /** False when the guide service was unreachable or empty — guides are the
   * bundled offline cache, not fresh content, and the UI must say so. */
  live: boolean;
};

/**
 * Guides degrade to the bundled offline cache instead of throwing: they are
 * safety content that must always render. `live` tells the caller whether the
 * fetch actually succeeded so it never reports a false "refreshed".
 */
export async function fetchGuides(language: string): Promise<GuideFetchResult> {
  try {
    const params = new URLSearchParams({ language, offline: "true" });
    const response = await fetch(`${GUIDE_API_BASE}/guides?${params}`);
    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }
    const payload = (await response.json()) as GuideListResponse;
    return payload.guides.length > 0
      ? { guides: payload.guides, live: true }
      : { guides: buildFallbackGuides(), live: false };
  } catch {
    return { guides: buildFallbackGuides(), live: false };
  }
}

/** Nearby shelters are live data only — failures throw so no fixture is shown. */
export async function fetchNearbyShelters(
  lat: number,
  lng: number,
): Promise<NearbyShelterResponse> {
  const response = await fetch(
    `${SHELTER_API_BASE}/shelters/nearby?lat=${encodeURIComponent(lat)}&lng=${encodeURIComponent(lng)}`,
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as NearbyShelterResponse;
}

function isRealPhone(phone: string): boolean {
  const trimmed = phone.trim();
  return trimmed.length > 0 && trimmed !== GUEST_PLACEHOLDER_PHONE;
}

export async function submitIncidentDraft(
  draft: ReportDraft,
  session: MobileSession,
  requestKind: "incident_report" | "distress_request" = "incident_report",
): Promise<CreateIncidentResponse> {
  const lat = Number(draft.lat);
  const lng = Number(draft.lng);
  const peopleAffected = Number(draft.peopleAffected || 0);
  if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
    throw new Error("Report needs valid latitude and longitude.");
  }
  if (draft.description.trim().length < 5) {
    throw new Error("Add a short description before submitting.");
  }

  // Responder follow-up requires a real phone number; the guest placeholder is
  // never sent as a callback contact.
  const reporterPhone =
    !draft.anonymous && draft.contactPermission && isRealPhone(session.phone)
      ? session.phone.trim()
      : undefined;
  const payload: CreateIncidentRequest = {
    requestKind,
    anonymous: draft.anonymous,
    contactPermission: draft.anonymous
      ? false
      : draft.contactPermission && Boolean(reporterPhone),
    description: draft.description.trim(),
    injuriesReported: draft.injuriesReported,
    location: { lat, lng },
    media: draft.mediaRefs,
    peopleAffected:
      requestKind === "distress_request"
        ? Math.max(1, Number.isFinite(peopleAffected) ? peopleAffected : 0)
        : Number.isFinite(peopleAffected)
          ? peopleAffected
          : 0,
    reporter: draft.anonymous
      ? undefined
      : {
          phone: reporterPhone,
          userId: session.userId,
        },
    type: draft.hazard,
    urgency:
      requestKind === "distress_request" ? "life_threatening" : draft.urgency,
  };

  const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
    body: JSON.stringify(payload),
    headers: { "Content-Type": "application/json" },
    method: "POST",
  });
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as CreateIncidentResponse;
}

/**
 * Push registration. Acquires the REAL Expo push token. notification-service
 * has no device-registration endpoint yet, so instead of pretending a sandbox
 * registration succeeded we report an honest not-configured state — alerts are
 * still delivered locally on refresh.
 */
export async function registerPushToken(
  granted: boolean,
): Promise<PushRegistrationState> {
  if (PUSH_PROVIDER === "disabled") {
    return {
      status: "not_configured",
      message: "Push registration is disabled for this environment.",
    };
  }
  if (!granted) {
    return {
      status: "permission_needed",
      message: "Allow notifications to receive urgent NADAA warnings.",
    };
  }
  try {
    await Notifications.getExpoPushTokenAsync();
  } catch (error) {
    return {
      status: "failed",
      message: `Could not get an Expo push token: ${
        error instanceof Error ? error.message : "unknown error"
      }`,
    };
  }
  return {
    status: "not_configured",
    message:
      "Remote push is not configured on the server yet. Alerts still arrive when you open or refresh the app.",
  };
}

/**
 * Register the signed-in citizen as a community volunteer. The incident-service
 * requires the citizen bearer token (it derives citizenUserId from the token
 * claims) and is idempotent per citizenUserId — a repeat registration returns
 * the existing profile. Community/district/region and skills come from the
 * registration form, never from client-side constants.
 */
export async function registerVolunteerProfile(
  session: MobileSession,
  registration: VolunteerRegistrationDraft,
): Promise<VolunteerProfileResponse> {
  const skills = registration.skills
    .split(",")
    .map((skill) => skill.trim())
    .filter((skill) => skill.length > 0);
  const payload: RegisterVolunteerRequest = {
    availabilityStatus: "available",
    citizenUserId: session.userId,
    community: registration.community.trim(),
    district: registration.district.trim(),
    languages: [session.preferredLanguage || "en"],
    name: session.name,
    phone: session.phone,
    region: registration.region.trim(),
    skills,
  };
  const response = await fetch(`${INCIDENT_API_BASE}/volunteers`, {
    body: JSON.stringify(payload),
    headers: bearerHeaders(session.accessToken),
    method: "POST",
  });
  if (!response.ok) {
    throw new ApiError(response.status, await extractAPIError(response));
  }
  return (await response.json()) as VolunteerProfileResponse;
}

export async function fetchVolunteerTasks(
  volunteerId: string,
  accessToken?: string,
): Promise<VolunteerTaskRecord[]> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/volunteers/${encodeURIComponent(volunteerId)}/tasks`,
    { headers: bearerHeaders(accessToken) },
  );
  if (!response.ok) {
    throw new ApiError(response.status, await extractAPIError(response));
  }
  const payload = (await response.json()) as VolunteerTaskListResponse;
  return payload.tasks;
}

/**
 * Volunteer write-path: errors THROW so the UI can show a retryable error.
 * Success is only ever reported when the service confirms it.
 */
export async function updateVolunteerTaskStatus(
  taskId: string,
  payload: VolunteerTaskStatusRequest,
  accessToken?: string,
): Promise<VolunteerTaskRecord> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/volunteer-tasks/${encodeURIComponent(taskId)}/status`,
    {
      body: JSON.stringify(payload),
      headers: bearerHeaders(accessToken),
      method: "PATCH",
    },
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as VolunteerTaskRecord;
}

/** Same contract as updateVolunteerTaskStatus: never fabricate success. */
export async function submitVolunteerObservation(
  taskId: string,
  payload: VolunteerObservationRequest,
  accessToken?: string,
): Promise<VolunteerTaskRecord> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/volunteer-tasks/${encodeURIComponent(taskId)}/observations`,
    {
      body: JSON.stringify(payload),
      headers: bearerHeaders(accessToken),
      method: "POST",
    },
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as VolunteerTaskRecord;
}

export function filterOfflineGuides(guides: EmergencyGuideRecord[]) {
  return guides.filter((guide) => guide.offlineAvailable);
}

function bearerHeaders(accessToken?: string): Record<string, string> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  return headers;
}

async function extractAPIError(response: Response) {
  try {
    const payload = (await response.json()) as {
      error?: { message?: string };
    };
    return (
      payload.error?.message ?? `${response.status} ${response.statusText}`
    );
  } catch {
    return `${response.status} ${response.statusText}`;
  }
}
