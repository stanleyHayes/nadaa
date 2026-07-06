export type HazardType =
  | "flood"
  | "fire"
  | "road_crash"
  | "building_collapse"
  | "medical_emergency"
  | "security_incident"
  | "disease_outbreak"
  | "electrical_hazard"
  | "blocked_drain"
  | "landslide"
  | "marine_accident"
  | "storm"
  | "tidal_wave"
  | "other";

export type RiskLevel = "low" | "moderate" | "high" | "severe" | "emergency";

export type IncidentStatus =
  | "reported"
  | "under_review"
  | "verified"
  | "assigned"
  | "response_en_route"
  | "on_scene"
  | "contained"
  | "recovery_ongoing"
  | "closed"
  | "false_report";

export type AlertSeverity = "advisory" | "watch" | "warning" | "severe_warning" | "emergency";

export interface Coordinates {
  lat: number;
  lng: number;
}

export interface RiskSummary {
  type: HazardType;
  level: RiskLevel;
  probability?: number;
  reason: string;
}

export interface ShelterSummary {
  id: string;
  name: string;
  location: Coordinates;
  capacity?: number;
  currentOccupancy?: number;
  contact?: string;
}

export interface AreaRiskResponse {
  location: string;
  overallRisk: RiskLevel;
  risks: RiskSummary[];
  nearestShelters: ShelterSummary[];
  recommendedActions: string[];
}

