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
import {
  AUTH_API_BASE,
  INCIDENT_API_BASE,
  SHELTER_API_BASE,
} from "../../app/config";
import { fallbackHospitalFacilities, fallbackIncidents } from "./data";
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
  try {
    const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
      headers: authorityHeaders(session),
    });
    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }
    const payload = (await response.json()) as IncidentListResponse;
    return payload.incidents.length > 0 ? payload.incidents : fallbackIncidents;
  } catch {
    return fallbackIncidents;
  }
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
  try {
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
    return (await response.json()) as HospitalCapacityResponse;
  } catch {
    return {
      facilities: fallbackHospitalFacilities,
      generatedAt: new Date().toISOString(),
      staleThresholdMinutes: 30,
    };
  }
}

export async function registerPushToken(
  granted: boolean,
): Promise<PushRegistrationState> {
  if (!granted) {
    return {
      status: "permission_needed",
      message: "Enable push notifications to receive critical escalations.",
    };
  }
  return {
    status: "registered",
    provider: "sandbox",
    token: `sandbox_dispatcher_${Date.now()}`,
  };
}
