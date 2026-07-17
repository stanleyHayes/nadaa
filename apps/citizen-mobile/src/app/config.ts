export const AUTH_API_BASE =
  process.env.EXPO_PUBLIC_AUTH_API_URL ?? "http://localhost:8080/api/v1";
export const INCIDENT_API_BASE =
  process.env.EXPO_PUBLIC_INCIDENT_API_URL ?? "http://localhost:8084/api/v1";
export const RISK_API_BASE =
  process.env.EXPO_PUBLIC_RISK_API_URL ?? "http://localhost:8081/api/v1";
export const NOTIFICATION_API_BASE =
  process.env.EXPO_PUBLIC_NOTIFICATION_API_URL ??
  "http://localhost:8090/api/v1";
export const GUIDE_API_BASE =
  process.env.EXPO_PUBLIC_GUIDE_API_URL ?? "http://localhost:8086/api/v1";
export const SHELTER_API_BASE =
  process.env.EXPO_PUBLIC_SHELTER_API_URL ?? "http://localhost:8093/api/v1";
export const PUSH_PROVIDER = process.env.EXPO_PUBLIC_PUSH_PROVIDER ?? "sandbox";

export const GUIDE_CACHE_KEY = "nadaa.mobile.guides.v1";
export const REPORT_DRAFT_KEY = "nadaa.mobile.report-draft.v1";
export const SESSION_KEY = "nadaa.mobile.session.v1";
export const VOLUNTEER_PROFILE_KEY = "nadaa.mobile.volunteer-profile.v1";
export const VOLUNTEER_TASKS_KEY = "nadaa.mobile.volunteer-tasks.v1";
