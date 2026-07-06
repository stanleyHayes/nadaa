export const INCIDENT_API_BASE =
  import.meta.env.VITE_INCIDENT_API_URL ?? "http://localhost:8084/api/v1";
export const RISK_API_BASE =
  import.meta.env.VITE_RISK_API_URL ?? "http://localhost:8081/api/v1";
export const NOTIFICATION_API_BASE =
  import.meta.env.VITE_NOTIFICATION_API_URL ?? "http://localhost:8090/api/v1";
export const GUIDE_API_BASE =
  import.meta.env.VITE_GUIDE_API_URL ?? "http://localhost:8086/api/v1";
export const GUIDE_CACHE_KEY = "nadaa.citizen.guides.v1";
