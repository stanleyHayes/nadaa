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

export type AgencyUserRole = Exclude<UserRole, "citizen">;

export type AgencyType =
  | "nadmo"
  | "district_assembly"
  | "police"
  | "fire"
  | "ambulance"
  | "meteorological"
  | "hydrological"
  | "hospital"
  | "utility"
  | "ngo"
  | "other";

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

export type AlertSeverity =
  "advisory" | "watch" | "warning" | "severe_warning" | "emergency";

export type AlertStatus =
  | "draft"
  | "submitted"
  | "approved"
  | "rejected"
  | "published"
  | "expired"
  | "cancelled";

export type AlertTargetType =
  "national" | "region" | "district" | "radius" | "community" | "custom";

export type GuideStage = "before" | "during" | "after" | "recovery";

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

export interface AgencySummary {
  id: string;
  name: string;
  type: AgencyType;
  region: string;
  district: string;
  contactNumber?: string;
}

export interface AgencyUserProfile {
  id: string;
  name: string;
  email: string;
  phone: string;
  role: AgencyUserRole;
  agency: AgencySummary;
  mfaRequired: boolean;
  mfaEnabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateAgencyUserRequest {
  name: string;
  email: string;
  phone: string;
  agencyId: string;
  role: AgencyUserRole;
}

export interface CreateAgencyUserResponse {
  user: AgencyUserProfile;
  temporaryPassword: string;
  mfaSetupRequired: true;
}

export interface AgencyMFASetupRequest {
  email: string;
  temporaryPassword: string;
}

export interface AgencyMFASetupResponse {
  userId: string;
  challengeId: string;
  method: "mock_totp";
  secret: string;
  expiresAt: string;
  devCode?: string;
}

export interface AgencyMFAVerifyRequest {
  email: string;
  temporaryPassword: string;
  code: string;
}

export interface AgencyMFAVerifyResponse {
  user: AgencyUserProfile;
}

export interface LoginAgencyRequest {
  email: string;
  password: string;
  mfaCode?: string;
}

export interface LoginAgencyResponse {
  accessToken: string;
  tokenType: "Bearer";
  expiresAt: string;
  user: AgencyUserProfile;
}

export interface AuditLogRecord {
  id: string;
  actorUserId?: string;
  actorAgencyId?: string;
  actorRole?: UserRole;
  action: string;
  targetType: string;
  targetId?: string;
  requestId?: string;
  ipAddress?: string;
  userAgent?: string;
  before?: Record<string, unknown>;
  after?: Record<string, unknown>;
  createdAt: string;
}

export interface AuditLogListResponse {
  logs: AuditLogRecord[];
}

export interface AlertTarget {
  type: AlertTargetType;
  ids: string[];
  label: string;
  center?: Coordinates;
  radiusMeters?: number;
  geometry?: AlertTargetGeometry;
  areaSqKm?: number;
  estimatedPopulation?: number;
}

export interface AlertTargetGeometry {
  type: "Polygon";
  coordinates: number[][][];
}

export interface AlertTargetPreviewResponse {
  target: AlertTarget;
  summary: string;
  warnings: string[];
}

export interface CreateAlertRequest {
  title: string;
  hazardType: HazardType;
  severity: AlertSeverity;
  message: string;
  target: AlertTarget;
  startsAt: string;
  expiresAt: string;
  recommendedAction: string;
  evacuationRequired: boolean;
  shelterIds: string[];
}

export interface AlertWorkflowRequest {
  note?: string;
  reason?: string;
}

export interface AuthorityAlertRecord {
  id: string;
  title: string;
  hazardType: HazardType;
  severity: AlertSeverity;
  message: string;
  target: AlertTarget;
  startsAt: string;
  expiresAt: string;
  recommendedAction: string;
  evacuationRequired: boolean;
  shelterIds: string[];
  issuingAgencyId: string;
  issuedBy: string;
  approvedBy?: string;
  rejectedBy?: string;
  status: AlertStatus;
  emergencyOverride: boolean;
  statusReason?: string;
  createdAt: string;
  updatedAt: string;
  submittedAt?: string;
  approvedAt?: string;
  rejectedAt?: string;
}

export interface AlertListResponse {
  alerts: AuthorityAlertRecord[];
}

export interface EmergencyGuideRecord {
  id: string;
  hazardType: HazardType;
  stage: GuideStage;
  title: string;
  body: string;
  language: string;
  offlineAvailable: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface GuideListResponse {
  guides: EmergencyGuideRecord[];
}

export type IntegrationDirection = "inbound" | "outbound" | "bidirectional";

export type IntegrationDomain =
  | "weather"
  | "hydrology"
  | "incident_sync"
  | "alert_sync"
  | "road_closure"
  | "hospital_capacity"
  | "utility_outage"
  | "shelter_status";

export interface IntegrationAuthentication {
  mode: "none" | "api_key" | "oauth2" | "mtls" | "signed_webhook" | "sftp";
  requiredHeaders?: string[];
  secretScope?: string;
}

export interface IntegrationPayloadContract {
  name: string;
  contentType: "application/json";
  requiredFields: string[];
  optionalFields?: string[];
  pii: "none" | "minimal_operational" | "aggregate_only";
  geometry?: string;
  exampleRef: string;
}

export interface IntegrationFailureBehavior {
  retryable: boolean;
  maxAttempts: number;
  backoffSeconds: number[];
  deadLetterQueue: string;
  manualFallback: string;
}

export interface IntegrationContract {
  id: string;
  partner: string;
  partnerType: AgencyType;
  domain: IntegrationDomain;
  direction: IntegrationDirection;
  dataOwner: string;
  cadence: string;
  authentication: IntegrationAuthentication;
  payloads: IntegrationPayloadContract[];
  failureBehavior: IntegrationFailureBehavior;
  sourceOfTruth: "originating_partner" | "nadaa" | "field_specific";
  freshnessWindowMinutes: number;
  contactPoint: string;
  status: "mock_contract" | "pilot" | "production";
  notes: string;
  updatedAt: string;
}

export interface IntegrationContractListResponse {
  contracts: IntegrationContract[];
}

export interface WeatherHydrologyObservation {
  id: string;
  source: string;
  metric: "rainfall_mm" | "water_level_m";
  value: number;
  unit: "mm" | "m";
  stationId: string;
  location: Coordinates;
  observedAt: string;
  validFrom: string;
  validTo: string;
  quality: string;
  generatedBy: "mock_adapter" | "partner_adapter";
}

export interface IntegrationObservationListResponse {
  observations: WeatherHydrologyObservation[];
}

export interface IntegrationSyncRequest {
  type: "incident" | "alert";
  sourceId: string;
  reference: string;
  hazardType: HazardType;
  status: string;
  severity: string;
  title?: string;
  summary?: string;
  message?: string;
  location?: Coordinates;
  targetLabel?: string;
  targetAgencyIds: string[];
  correlationId: string;
  occurredAt?: string;
}

export interface IntegrationSyncEvent {
  id: string;
  type: "incident" | "alert";
  sourceId: string;
  reference: string;
  targetAgencyIds: string[];
  correlationId: string;
  status: "accepted";
  adapterId: string;
  queuedAt: string;
  retryable: boolean;
}

export interface IntegrationSyncEventListResponse {
  events: IntegrationSyncEvent[];
}

export interface IncidentReporterRef {
  userId: string;
  phone?: string;
}

export interface DuplicateIncidentCandidate {
  incidentId: string;
  reference: string;
  score: number;
  distanceMeters: number;
  minutesApart: number;
  reasons: string[];
}

export interface IncidentAbuseSignal {
  code: string;
  label: string;
  detail: string;
  weight: number;
}

export type IncidentAbuseReviewDecision = "clear" | "monitor" | "false_report";

export type IncidentAssignmentPriority = "low" | "normal" | "high" | "urgent";

export interface IncidentAssignmentRecord {
  id: string;
  agencyId: string;
  agencyName: string;
  agencyType: AgencyType;
  priority: IncidentAssignmentPriority;
  instructions: string;
  responderLead?: string;
  status: "active" | "completed" | "cancelled";
  assignedBy: string;
  assignedAt: string;
}

export interface IncidentTimelineEvent {
  id: string;
  type: string;
  message: string;
  actorUserId?: string;
  actorAgencyId?: string;
  actorRole?: AgencyUserRole;
  metadata?: Record<string, string>;
  createdAt: string;
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
  abuseSignals: IncidentAbuseSignal[];
  abuseScore: number;
  abuseReviewRequired: boolean;
  abuseReviewReason?: string;
  abuseReviewDecision?: IncidentAbuseReviewDecision;
  abuseReviewedBy?: string;
  abuseReviewedAt?: string;
  duplicateCandidates: DuplicateIncidentCandidate[];
  mergedIncidentIds: string[];
  assignments: IncidentAssignmentRecord[];
  timeline: IncidentTimelineEvent[];
  mergedIntoId?: string;
  mergedBy?: string;
  mergedAt?: string;
  mergeReason?: string;
  reportedBy?: IncidentReporterRef;
  verifiedBy?: string;
  verifiedAt?: string;
  statusUpdatedBy?: string;
  statusReason?: string;
  resolutionNotes?: string;
  closedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateIncidentResponse {
  id: string;
  reference: string;
  status: "reported";
  severity: RiskLevel;
  priorityReview: boolean;
  abuseSignals: IncidentAbuseSignal[];
  abuseScore: number;
  abuseReviewRequired: boolean;
  duplicateCandidates: DuplicateIncidentCandidate[];
}

export interface IncidentListResponse {
  incidents: IncidentRecord[];
}

export interface DuplicateReviewCandidate {
  candidate: DuplicateIncidentCandidate;
  incident: IncidentRecord;
}

export interface DuplicateReviewResponse {
  incident: IncidentRecord;
  candidates: DuplicateReviewCandidate[];
}

export interface IncidentWorkflowRequest {
  note?: string;
  resolutionNotes?: string;
}

export interface MergeIncidentsRequest {
  duplicateIncidentIds: string[];
  note: string;
}

export interface MergeIncidentsResponse {
  incident: IncidentRecord;
  mergedIncidents: IncidentRecord[];
}

export interface IncidentStatusUpdateRequest extends IncidentWorkflowRequest {
  status: IncidentStatus;
}

export interface IncidentAbuseReviewRequest extends IncidentWorkflowRequest {
  decision: IncidentAbuseReviewDecision;
}

export interface AssignIncidentRequest {
  agencyId: string;
  agencyName: string;
  agencyType: AgencyType;
  priority?: IncidentAssignmentPriority;
  instructions: string;
  responderLead?: string;
}

export interface IncidentAuditEvent {
  id: string;
  actorUserId: string;
  actorAgencyId: string;
  actorRole: AgencyUserRole;
  action: string;
  targetType: "incident";
  targetId: string;
  requestId?: string;
  before?: Record<string, unknown>;
  after?: Record<string, unknown>;
  createdAt: string;
}

export interface IncidentAuditListResponse {
  logs: IncidentAuditEvent[];
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
  distanceMeters?: number;
  status?: "open" | "full" | "closed" | "unknown";
  facilities?: string[];
}

export interface EmergencyFacilitySummary {
  id: string;
  name: string;
  type: string;
  location: Coordinates;
  region?: string;
  district?: string;
  contact?: string;
  distanceMeters?: number;
}

export interface AreaRiskResponse {
  location: string;
  overallRisk: RiskLevel;
  risks: RiskSummary[];
  nearestShelters: ShelterSummary[];
  nearbyFacilities: EmergencyFacilitySummary[];
  recommendedActions: string[];
}
