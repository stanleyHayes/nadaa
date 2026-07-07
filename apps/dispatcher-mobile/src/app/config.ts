export const AUTH_API_BASE =
  process.env.EXPO_PUBLIC_AUTH_API_URL ?? "http://localhost:8080/api/v1";
export const INCIDENT_API_BASE =
  process.env.EXPO_PUBLIC_INCIDENT_API_URL ?? "http://localhost:8084/api/v1";
export const SHELTER_API_BASE =
  process.env.EXPO_PUBLIC_SHELTER_API_URL ?? "http://localhost:8093/api/v1";
export const PUSH_PROVIDER = process.env.EXPO_PUBLIC_PUSH_PROVIDER ?? "sandbox";

export const SESSION_KEY = "nadaa.dispatcher.session.v1";
export const INCIDENT_CACHE_KEY = "nadaa.dispatcher.incidents.v1";
export const CAPACITY_CACHE_KEY = "nadaa.dispatcher.capacity.v1";
export const SELECTED_INCIDENT_KEY = "nadaa.dispatcher.selected-incident.v1";
