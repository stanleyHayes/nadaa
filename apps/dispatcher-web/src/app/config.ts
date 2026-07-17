export const AUTH_API_BASE =
  import.meta.env.VITE_AUTH_API_URL ?? "http://localhost:8080/api/v1";
export const INCIDENT_API_BASE =
  import.meta.env.VITE_INCIDENT_API_URL ?? "http://localhost:8084/api/v1";
export const ALERT_API_BASE =
  import.meta.env.VITE_ALERT_API_URL ?? "http://localhost:8089/api/v1";
export const ML_API_BASE =
  import.meta.env.VITE_ML_API_URL ?? "http://localhost:8094/api/v1";
export const SHELTER_API_BASE =
  import.meta.env.VITE_SHELTER_API_URL ?? "http://localhost:8093/api/v1";
export const ROAD_CLOSURE_API_BASE =
  import.meta.env.VITE_ROAD_CLOSURE_API_URL ?? "http://localhost:8095/api/v1";

/**
 * Dev-only escape hatch: baseline fixture predictions/suggestions may fill the
 * console when a backend is down during local development. In production builds
 * fixtures never reach the UI — fabricated probabilities must never drive a
 * real alert draft.
 */
export const FIXTURE_DATA_ENABLED = import.meta.env.DEV;
