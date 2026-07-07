import type {
  AreaRiskResponse,
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
import {
  GUIDE_API_BASE,
  INCIDENT_API_BASE,
  NOTIFICATION_API_BASE,
  PUSH_PROVIDER,
  RISK_API_BASE,
  SHELTER_API_BASE,
} from "../../app/config";
import {
  buildFallbackAlerts,
  buildFallbackGuides,
  sampleRisk,
  sampleShelters,
  sampleVolunteerProfile,
  sampleVolunteerTasks,
} from "./data";
import type {
  MobileSession,
  PushRegistrationState,
  ReportDraft,
} from "./types";

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

export async function fetchAlertFeed() {
  try {
    const response = await fetch(
      `${NOTIFICATION_API_BASE}/notifications/alerts?includeExpired=true`,
    );
    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }
    const payload = (await response.json()) as CitizenAlertFeedResponse;
    return payload.alerts.length > 0 ? payload.alerts : buildFallbackAlerts();
  } catch {
    return buildFallbackAlerts();
  }
}

export async function fetchGuides(language: string) {
  try {
    const params = new URLSearchParams({ language, offline: "true" });
    const response = await fetch(`${GUIDE_API_BASE}/guides?${params}`);
    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }
    const payload = (await response.json()) as GuideListResponse;
    return payload.guides.length > 0 ? payload.guides : buildFallbackGuides();
  } catch {
    return buildFallbackGuides();
  }
}

export async function fetchNearbyShelters(
  lat: number,
  lng: number,
): Promise<NearbyShelterResponse> {
  try {
    const response = await fetch(
      `${SHELTER_API_BASE}/shelters/nearby?lat=${encodeURIComponent(lat)}&lng=${encodeURIComponent(lng)}`,
    );
    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }
    return (await response.json()) as NearbyShelterResponse;
  } catch {
    return sampleShelters;
  }
}

export async function submitIncidentDraft(
  draft: ReportDraft,
  session: MobileSession,
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

  const payload: CreateIncidentRequest = {
    anonymous: draft.anonymous,
    contactPermission: draft.anonymous ? false : draft.contactPermission,
    description: draft.description.trim(),
    injuriesReported: draft.injuriesReported,
    location: { lat, lng },
    media: draft.mediaRefs,
    peopleAffected: Number.isFinite(peopleAffected) ? peopleAffected : 0,
    reporter: draft.anonymous
      ? undefined
      : {
          phone: draft.contactPermission ? session.phone : undefined,
          userId: session.userId,
        },
    type: draft.hazard,
    urgency: draft.urgency,
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
  return {
    provider: PUSH_PROVIDER,
    status: "registered",
    token: `expo_push_${PUSH_PROVIDER}_sandbox`,
  };
}

export async function registerVolunteerProfile(
  session: MobileSession,
): Promise<VolunteerProfileResponse> {
  const payload: RegisterVolunteerRequest = {
    availabilityStatus: "available",
    citizenUserId: session.userId,
    community: "Jamestown",
    district: "Accra Metropolitan",
    languages: [session.preferredLanguage || "en"],
    name: session.name,
    phone: session.phone,
    region: "Greater Accra",
    skills: ["first aid", "community alerts"],
  };
  try {
    const response = await fetch(`${INCIDENT_API_BASE}/volunteers`, {
      body: JSON.stringify(payload),
      headers: { "Content-Type": "application/json" },
      method: "POST",
    });
    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }
    return (await response.json()) as VolunteerProfileResponse;
  } catch {
    return { volunteer: sampleVolunteerProfile };
  }
}

export async function fetchVolunteerTasks(
  volunteerId: string,
): Promise<VolunteerTaskRecord[]> {
  try {
    const response = await fetch(
      `${INCIDENT_API_BASE}/volunteers/${encodeURIComponent(volunteerId)}/tasks`,
    );
    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }
    const payload = (await response.json()) as VolunteerTaskListResponse;
    return payload.tasks.length > 0 ? payload.tasks : sampleVolunteerTasks;
  } catch {
    return sampleVolunteerTasks;
  }
}

export async function updateVolunteerTaskStatus(
  taskId: string,
  payload: VolunteerTaskStatusRequest,
): Promise<VolunteerTaskRecord> {
  try {
    const response = await fetch(
      `${INCIDENT_API_BASE}/volunteer-tasks/${encodeURIComponent(taskId)}/status`,
      {
        body: JSON.stringify(payload),
        headers: { "Content-Type": "application/json" },
        method: "PATCH",
      },
    );
    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }
    return (await response.json()) as VolunteerTaskRecord;
  } catch {
    return {
      ...sampleVolunteerTasks[0],
      status: payload.status,
      updatedAt: new Date().toISOString(),
    };
  }
}

export async function submitVolunteerObservation(
  taskId: string,
  payload: VolunteerObservationRequest,
): Promise<VolunteerTaskRecord> {
  try {
    const response = await fetch(
      `${INCIDENT_API_BASE}/volunteer-tasks/${encodeURIComponent(taskId)}/observations`,
      {
        body: JSON.stringify(payload),
        headers: { "Content-Type": "application/json" },
        method: "POST",
      },
    );
    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }
    return (await response.json()) as VolunteerTaskRecord;
  } catch {
    return {
      ...sampleVolunteerTasks[0],
      escalationRequired:
        payload.escalationRequested ||
        payload.safetyStatus === "unsafe" ||
        payload.safetyStatus === "needs_authority",
      status:
        payload.escalationRequested ||
        payload.safetyStatus === "unsafe" ||
        payload.safetyStatus === "needs_authority"
          ? "needs_escalation"
          : sampleVolunteerTasks[0].status,
      updatedAt: new Date().toISOString(),
    };
  }
}

export function fallbackRisk() {
  return sampleRisk;
}

export function filterOfflineGuides(guides: EmergencyGuideRecord[]) {
  return guides.filter((guide) => guide.offlineAvailable);
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
