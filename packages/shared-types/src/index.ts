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

export type UserRole =
  | "citizen"
  | "agency_viewer"
  | "dispatcher"
  | "responder"
  | "nadmo_officer"
  | "district_officer"
  | "agency_admin"
  | "system_admin";

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

export type IncidentUrgency = "low" | "moderate" | "high" | "life_threatening";

export type AlertSeverity = "advisory" | "watch" | "warning" | "severe_warning" | "emergency";

export interface Coordinates {
  lat: number;
  lng: number;
}

export interface CitizenProfile {
  id: string;
  name: string;
  phone: string;
  role: "citizen";
  preferredLanguage: string;
  homeLocation?: Coordinates;
  contactPermission: boolean;
  createdAt: string;
}

export interface RegisterCitizenRequest {
  name: string;
  phone: string;
  preferredLanguage?: string;
  homeLocation?: Coordinates;
  contactPermission: boolean;
}

export interface RegisterCitizenResponse {
  userId: string;
  phone: string;
  challengeId: string;
  otpDelivery: "mock" | "sms" | "voice" | "whatsapp";
  devOtp?: string;
}

export interface LoginCitizenRequest {
  phone: string;
  otp: string;
}

export interface LoginCitizenResponse {
  accessToken: string;
  tokenType: "Bearer";
  expiresAt: string;
  user: CitizenProfile;
}

export interface IncidentReporterRef {
  userId: string;
  phone?: string;
}

export interface CreateIncidentRequest {
  type: HazardType;
  description: string;
  location: Coordinates;
  peopleAffected: number;
  injuriesReported: boolean;
  urgency: IncidentUrgency;
  anonymous: boolean;
  contactPermission: boolean;
  accessibilityNeeds?: string;
  media: string[];
  reporter?: IncidentReporterRef;
}

export interface IncidentRecord {
  id: string;
  reference: string;
  type: HazardType;
  severity: RiskLevel;
  status: IncidentStatus;
  description: string;
  location: Coordinates;
  peopleAffected: number;
  injuriesReported: boolean;
  urgency: IncidentUrgency;
  anonymous: boolean;
  contactPermission: boolean;
  accessibilityNeeds?: string;
  media: string[];
  priorityReview: boolean;
  reportedBy?: IncidentReporterRef;
  createdAt: string;
  updatedAt: string;
}

export interface CreateIncidentResponse {
  id: string;
  reference: string;
  status: "reported";
  severity: RiskLevel;
  priorityReview: boolean;
  duplicateCandidates: string[];
}

export interface IncidentListResponse {
  incidents: IncidentRecord[];
}

export type IncidentMediaPurpose = "incident_media";

export type IncidentMediaContentType =
  | "image/jpeg"
  | "image/png"
  | "image/webp"
  | "video/mp4"
  | "video/quicktime"
  | "audio/mpeg"
  | "audio/mp4"
  | "audio/wav";

export interface InitiateMediaUploadRequest {
  purpose: IncidentMediaPurpose;
  fileName: string;
  contentType: IncidentMediaContentType;
  sizeBytes: number;
  uploadedBy?: string;
}

export interface MediaUploadResponse {
  mediaId: string;
  uploadUrl: string;
  method: "PUT";
  headers: Record<string, string>;
  expiresAt: string;
  maxSizeBytes: number;
  access: "private";
}

export interface IncidentMediaRecord {
  id: string;
  purpose: IncidentMediaPurpose;
  fileName: string;
  contentType: IncidentMediaContentType;
  sizeBytes: number;
  uploadedBy?: string;
  incidentId?: string;
  access: "private";
  status: "pending_upload" | "linked";
  uploadUrl: string;
  expiresAt: string;
  createdAt: string;
  linkedAt?: string;
}

export interface MediaListResponse {
  media: IncidentMediaRecord[];
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
