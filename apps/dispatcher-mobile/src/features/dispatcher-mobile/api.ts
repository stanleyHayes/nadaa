import type {
  AgencyUserProfile,
  AssignIncidentRequest,
  HospitalCapacityResponse,
  IncidentAbuseReviewRequest,
  IncidentListResponse,
  IncidentRecord,
  IncidentStatusUpdateRequest,
  LoginAgencyRequest,
  LoginAgencyResponse,
  MergeIncidentsRequest,
} from "@nadaa/shared-types";
import * as Notifications from "expo-notifications";
import {
  AUTH_API_BASE,
  INCIDENT_API_BASE,
  PUSH_PROVIDER,
  SHELTER_API_BASE,
} from "../../app/config";
import type { DispatcherSession, PushRegistrationState } from "./types";

async function extractAPIError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as {
      message?: string;
      error?: string;
    };
    return body.message ?? body.error ?? `Request failed (${response.status})`;
  } catch {
    return `Request failed (${response.status})`;
  }
}

export function authorityHeaders(session: DispatcherSession) {
  return {
    "Content-Type": "application/json",
    // Authority endpoints require a verified Bearer token. The X-NADAA-* actor
    // headers are kept for mock-actors dev mode; services ignore them once a
    // real token is verified.
    ...(session.accessToken
      ? { Authorization: `Bearer ${session.accessToken}` }
      : {}),
    "X-NADAA-Actor-ID": session.userId,
    "X-NADAA-Actor-Role": session.role,
    "X-NADAA-Agency-ID": session.agencyId,
    "X-NADAA-MFA-Completed": session.mfaCompleted ? "true" : "false",
    "X-NADAA-Request-ID": `dispatcher-mobile-${Date.now()}`,
  };
}

export async function agencyLogin(
  credentials: LoginAgencyRequest,
): Promise<LoginAgencyResponse & { user: AgencyUserProfile }> {
  const response = await fetch(`${AUTH_API_BASE}/auth/agency/login`, {
    body: JSON.stringify(credentials),
    headers: { "Content-Type": "application/json" },
    method: "POST",
  });
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as LoginAgencyResponse & {
    user: AgencyUserProfile;
  };
}

export async function fetchMyProfile(
  session: DispatcherSession,
): Promise<AgencyUserProfile> {
  const response = await fetch(`${AUTH_API_BASE}/auth/me`, {
    headers: {
      Authorization: `Bearer ${session.accessToken ?? ""}`,
      "Content-Type": "application/json",
    },
  });
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as AgencyUserProfile;
}

export async function fetchIncidentQueue(
  session: DispatcherSession,
): Promise<IncidentRecord[]> {
  const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
    headers: authorityHeaders(session),
  });
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  const payload = (await response.json()) as IncidentListResponse;
  // An empty queue is a real answer, and failures must surface to the caller —
  // substituting fixture incidents would let fixtures reach device
  // notifications and overwrite the offline cache.
  return payload.incidents;
}

export async function verifyIncident(
  session: DispatcherSession,
  incidentId: string,
  note?: string,
): Promise<IncidentRecord> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/incidents/${encodeURIComponent(incidentId)}/verify`,
    {
      body: JSON.stringify({ note } satisfies { note?: string }),
      headers: authorityHeaders(session),
      method: "POST",
    },
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as IncidentRecord;
}

export async function updateIncidentStatus(
  session: DispatcherSession,
  incidentId: string,
  request: IncidentStatusUpdateRequest,
): Promise<IncidentRecord> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/incidents/${encodeURIComponent(incidentId)}/status`,
    {
      body: JSON.stringify(request),
      headers: authorityHeaders(session),
      method: "PATCH",
    },
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as IncidentRecord;
}

export async function assignIncident(
  session: DispatcherSession,
  incidentId: string,
  request: AssignIncidentRequest,
): Promise<IncidentRecord> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/incidents/${encodeURIComponent(incidentId)}/assignments`,
    {
      body: JSON.stringify(request),
      headers: authorityHeaders(session),
      method: "POST",
    },
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as IncidentRecord;
}

export async function reviewAbuse(
  session: DispatcherSession,
  incidentId: string,
  request: IncidentAbuseReviewRequest,
): Promise<IncidentRecord> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/incidents/${encodeURIComponent(incidentId)}/abuse-review`,
    {
      body: JSON.stringify(request),
      headers: authorityHeaders(session),
      method: "POST",
    },
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as IncidentRecord;
}

export async function mergeIncidents(
  session: DispatcherSession,
  incidentId: string,
  request: MergeIncidentsRequest,
): Promise<IncidentRecord> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/incidents/${encodeURIComponent(incidentId)}/merge`,
    {
      body: JSON.stringify(request),
      headers: authorityHeaders(session),
      method: "POST",
    },
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  return (await response.json()) as IncidentRecord;
}

export async function fetchHospitalCapacity(
  session: DispatcherSession,
  lat: number,
  lng: number,
): Promise<HospitalCapacityResponse> {
  const params = new URLSearchParams({
    includeStale: "true",
    lat: lat.toString(),
    limit: "6",
    lng: lng.toString(),
  });
  const response = await fetch(
    `${SHELTER_API_BASE}/hospitals/capacity?${params}`,
    {
      headers: authorityHeaders(session),
    },
  );
  if (!response.ok) {
    throw new Error(await extractAPIError(response));
  }
  // Failures must surface to the caller, which keeps the cached capacity with
  // an offline indicator — fixture facilities must not pose as live data.
  return (await response.json()) as HospitalCapacityResponse;
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
      message: "Enable push notifications to receive critical escalations.",
    };
  }
  try {
    // A real Expo push token — never a fabricated sandbox string.
    await Notifications.getExpoPushTokenAsync();
  } catch {
    return {
      status: "failed",
      message: "This device could not provide an Expo push token.",
    };
  }
  // notification-service exposes no device-token registration endpoint, so a
  // real backend registration cannot complete — report that honestly instead
  // of faking a "registered" state. The queue polls while the app is
  // foregrounded, so new incidents still surface on the device.
  return {
    status: "not_configured",
    message:
      "Push backend registration is not configured; the queue polls for new incidents while the app is open.",
  };
}
