export const INCIDENT_API_BASE =
  import.meta.env.VITE_INCIDENT_API_URL ?? "http://localhost:8084/api/v1";
export const RISK_API_BASE =
  import.meta.env.VITE_RISK_API_URL ?? "http://localhost:8081/api/v1";
export const NOTIFICATION_API_BASE =
  import.meta.env.VITE_NOTIFICATION_API_URL ?? "http://localhost:8090/api/v1";
export const GUIDE_API_BASE =
  import.meta.env.VITE_GUIDE_API_URL ?? "http://localhost:8086/api/v1";
export const SHELTER_API_BASE =
  import.meta.env.VITE_SHELTER_API_URL ?? "http://localhost:8093/api/v1";
export const ROAD_CLOSURE_API_BASE =
  import.meta.env.VITE_ROAD_CLOSURE_API_URL ?? "http://localhost:8095/api/v1";
export const ROUTE_API_BASE =
  import.meta.env.VITE_ROUTE_SERVICE_URL ?? "http://localhost:8096";
export const DONATION_API_BASE =
  import.meta.env.VITE_DONATION_API_URL ?? "http://localhost:8100/api/v1";
export const GUIDE_CACHE_KEY = "nadaa.citizen.guides.v1";
