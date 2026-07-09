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

export interface AlertSourcePrediction {
  predictionId: string;
  predictionLogId?: string;
  modelVersion: string;
  inputFeatureSetVersion: string;
  probability: number;
  severity: RiskLevel;
  confidence: "low" | "medium" | "high";
  humanReviewRequired: boolean;
  autoPublishAllowed: false;
  reviewNote?: string;
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
  sourcePrediction?: AlertSourcePrediction;
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
  sourcePrediction?: AlertSourcePrediction;
  createdAt: string;
  updatedAt: string;
  submittedAt?: string;
  approvedAt?: string;
  rejectedAt?: string;
}

export interface AlertListResponse {
  alerts: AuthorityAlertRecord[];
}

export type CitizenAlertFeedStatus = "current" | "expired" | "upcoming";

export interface CitizenAlertFeedItem {
  id: string;
  title: string;
  hazardType: HazardType;
  severity: AlertSeverity;
  message: string;
  target: AlertTarget;
  targetLabel: string;
  startsAt: string;
  expiresAt: string;
  status: CitizenAlertFeedStatus;
  recommendedAction: string;
  evacuationRequired: boolean;
  shelterIds: string[];
  source: "alert-service" | "fixture";
  updatedAt: string;
}

export interface CitizenAlertFeedResponse {
  alerts: CitizenAlertFeedItem[];
  generatedAt: string;
  source: "alert-service" | "fixture" | "alert-service+fixture";
}

export type NotificationChannel = "push" | "sms" | "voice";
export type NotificationDeliveryStatus =
  "queued" | "delivered" | "failed" | "skipped";
export type VoiceLanguage = "en" | "tw" | "ga" | "ee" | "dag" | "ha";
export type VoiceAlertStatus = "generated" | "approved" | "rejected";
export type VoiceReviewStatus =
  "pending_review" | "partial_review" | "approved" | "rejected";
export type InclusiveAccessChannel = "sms" | "ussd" | "whatsapp";
export type InclusiveAccessStatus =
  "handled" | "failed" | "queued" | "submitted";
export type InclusiveAccessIntent =
  | "language_menu"
  | "main_menu"
  | "current_alerts"
  | "report_emergency"
  | "risk_check"
  | "emergency_guides"
  | "shelter_lookup"
  | "guidance_112"
  | "provider_error"
  | "invalid_selection";

export interface NotificationDeliveryRequest {
  recipientId?: string;
  phone?: string;
  pushToken?: string;
  language?: string;
  channels: NotificationChannel[];
  dryRun?: boolean;
}

export interface NotificationDeliveryAttempt {
  id: string;
  alertId: string;
  alertTitle: string;
  channel: NotificationChannel;
  provider: string;
  recipientRef: string;
  status: NotificationDeliveryStatus;
  reason?: string;
  messageId?: string;
  voiceAssetId?: string;
  language?: string;
  audioUrl?: string;
  attemptedAt: string;
}

export interface NotificationDeliveryResponse {
  attempts: NotificationDeliveryAttempt[];
}

export interface NotificationDeliveryLogListResponse {
  logs: NotificationDeliveryAttempt[];
}

export interface VoiceAlertRequest {
  alertId: string;
  languages?: VoiceLanguage[];
  workflowRequestedBy?: string;
  source?: "tts_sandbox" | "recorded_audio";
}

export interface VoiceVariant {
  id: string;
  language: VoiceLanguage;
  locale: string;
  voiceName: string;
  messageText: string;
  audioUrl: string;
  durationSeconds: number;
  status: VoiceAlertStatus;
  reviewStatus: VoiceReviewStatus;
  accessibilityChecks: string[];
  createdAt: string;
  updatedAt: string;
}

export interface VoiceAlertAsset {
  id: string;
  alertId: string;
  alertTitle: string;
  hazardType: HazardType;
  severity: AlertSeverity;
  targetLabel: string;
  status: VoiceAlertStatus;
  reviewStatus: VoiceReviewStatus;
  source: "tts_sandbox" | "recorded_audio";
  workflowRequestedBy?: string;
  reviewer?: string;
  reviewNote?: string;
  variants: VoiceVariant[];
  createdAt: string;
  updatedAt: string;
  reviewedAt?: string;
}

export interface VoiceAlertResponse {
  asset: VoiceAlertAsset;
}

export interface VoiceAlertListResponse {
  assets: VoiceAlertAsset[];
}

export interface VoiceReviewRequest {
  action: "approve" | "reject";
  reviewer: string;
  note?: string;
  languages?: VoiceLanguage[];
}

export interface VoiceRecipient {
  recipientId?: string;
  phone?: string;
  language: VoiceLanguage;
}

export interface VoiceDeliveryRequest {
  recipients: VoiceRecipient[];
  dryRun?: boolean;
}

export interface VoiceDeliveryResponse {
  attempts: NotificationDeliveryAttempt[];
}

export interface InclusiveAccessLog {
  id: string;
  channel: InclusiveAccessChannel;
  provider: string;
  providerMessageId?: string;
  sessionId?: string;
  phoneRef: string;
  profileId?: string;
  linkedProfile: boolean;
  language: string;
  intent: InclusiveAccessIntent;
  status: InclusiveAccessStatus;
  providerError?: string;
  incidentId?: string;
  incidentReference?: string;
  createdAt: string;
}

export interface InclusiveAccessReport {
  id: string;
  channel: InclusiveAccessChannel;
  type: HazardType;
  urgency: IncidentUrgency;
  description: string;
  location: Coordinates;
  locationLabel: string;
  phoneRef: string;
  profileId?: string;
  linkedProfile: boolean;
  status: "queued" | "submitted";
  media?: string[];
  incidentId?: string;
  incidentReference?: string;
  failureReason?: string;
  createdAt: string;
}

export interface USSDWebhookRequest {
  sessionId: string;
  phone: string;
  serviceCode?: string;
  text: string;
  language?: string;
  network?: string;
  provider?: string;
  providerMessageId?: string;
  providerError?: string;
  profileId?: string;
  linkProfile?: boolean;
  location?: Coordinates;
}

export interface USSDWebhookResponse {
  sessionId: string;
  action: "continue" | "end";
  message: string;
  language: string;
  log: InclusiveAccessLog;
  report?: InclusiveAccessReport;
}

export interface SMSInboundRequest {
  from: string;
  body: string;
  language?: string;
  provider?: string;
  providerMessageId?: string;
  providerError?: string;
  profileId?: string;
  linkProfile?: boolean;
  location?: Coordinates;
}

export interface SMSInboundResponse {
  message: string;
  log: InclusiveAccessLog;
  report?: InclusiveAccessReport;
}

export interface WhatsAppMedia {
  id?: string;
  url?: string;
  contentType?: string;
  caption?: string;
}

export interface WhatsAppInboundRequest {
  from: string;
  body: string;
  language?: string;
  provider?: string;
  providerMessageId?: string;
  providerError?: string;
  profileId?: string;
  linkProfile?: boolean;
  location?: Coordinates;
  media?: WhatsAppMedia[];
}

export type WhatsAppConversationState =
  | "idle"
  | "awaiting_report_hazard"
  | "awaiting_report_urgency"
  | "awaiting_report_location";

export interface WhatsAppConversation {
  id: string;
  channel: "whatsapp";
  phoneRef: string;
  profileId?: string;
  linkedProfile: boolean;
  language: string;
  intent: InclusiveAccessIntent;
  state: WhatsAppConversationState;
  hazard?: HazardType;
  urgency?: IncidentUrgency;
  lastMessageSummary?: string;
  lastMediaSummary?: string;
  startedAt: string;
  updatedAt: string;
  expiresAt: string;
  retentionUntil: string;
}

export interface WhatsAppTranscript {
  id: string;
  conversationId: string;
  provider: string;
  providerMessageId?: string;
  phoneRef: string;
  profileId?: string;
  linkedProfile: boolean;
  direction: "inbound" | "outbound";
  intent: InclusiveAccessIntent | "incoming";
  state: WhatsAppConversationState;
  messageSummary?: string;
  mediaSummary?: string;
  createdAt: string;
  retentionUntil: string;
}

export interface WhatsAppInboundResponse {
  message: string;
  conversation: WhatsAppConversation;
  log: InclusiveAccessLog;
  report?: InclusiveAccessReport;
  transcriptIds?: string[];
}

export interface InclusiveAccessLogListResponse {
  logs: InclusiveAccessLog[];
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

export interface ImportedWeatherHydrologyObservation extends WeatherHydrologyObservation {
  rainfallMm?: number;
  waterLevelM?: number;
  metadata: Record<string, string>;
  importJobId: string;
  importedAt: string;
  sourceRecord: string;
  storageTarget: "weather_observations";
}

export interface ImportedObservationListResponse {
  observations: ImportedWeatherHydrologyObservation[];
}

export interface ObservationImportRequest {
  adapterId?: string;
  metric?: "rainfall_mm" | "water_level_m";
  simulateFailure?: boolean;
  failureMessage?: string;
  requestedBy?: string;
  correlationId?: string;
}

export type ObservationImportStatus = "running" | "succeeded" | "failed";
export type ObservationImportTrigger = "manual" | "scheduled" | "retry";

export interface ObservationImportJob {
  id: string;
  adapterId: string;
  source: string;
  metric?: "rainfall_mm" | "water_level_m";
  status: ObservationImportStatus;
  trigger: ObservationImportTrigger;
  attempts: number;
  retryable: boolean;
  startedAt: string;
  finishedAt?: string;
  nextRetryAt?: string;
  importedCount: number;
  failedCount: number;
  error?: string;
  message: string;
  requestedBy?: string;
  correlationId?: string;
}

export interface ObservationImportJobListResponse {
  jobs: ObservationImportJob[];
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

export type IncidentLocationPrecision = "exact" | "approximate";

export interface IncidentPrivacyPolicy {
  reporterIdentityVisible: boolean;
  reporterContactVisible: boolean;
  locationPrecision: IncidentLocationPrecision;
  locationUse: "emergency_response";
  disclosure: string;
  notes: string[];
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
  privacy?: IncidentPrivacyPolicy;
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

export type VolunteerAvailabilityStatus = "available" | "busy" | "off_duty";
export type VolunteerVerificationStatus =
  "pending" | "verified" | "rejected" | "suspended";
export type VolunteerVerificationDecision = "verify" | "reject" | "suspend";
export type VolunteerTaskType =
  | "welfare_check"
  | "shelter_support"
  | "supply_distribution"
  | "damage_observation"
  | "route_observation"
  | "community_alerting";
export type VolunteerTaskStatus =
  | "assigned"
  | "accepted"
  | "en_route"
  | "on_scene"
  | "completed"
  | "cancelled"
  | "needs_escalation";
export type VolunteerSafetyStatus =
  "safe" | "caution" | "unsafe" | "needs_authority";

export interface VolunteerProfile {
  id: string;
  citizenUserId: string;
  name: string;
  phone?: string;
  region: string;
  district: string;
  community: string;
  groupId: string;
  skills: string[];
  languages: string[];
  availabilityStatus: VolunteerAvailabilityStatus;
  verificationStatus: VolunteerVerificationStatus;
  safetyNotes: string[];
  verifiedBy?: string;
  verifiedAt?: string;
  rejectionReason?: string;
  createdAt: string;
  updatedAt: string;
}

export interface RegisterVolunteerRequest {
  citizenUserId: string;
  name: string;
  phone: string;
  region: string;
  district: string;
  community: string;
  skills: string[];
  languages: string[];
  availabilityStatus?: VolunteerAvailabilityStatus;
}

export interface VolunteerProfileResponse {
  volunteer: VolunteerProfile;
}

export interface VolunteerListResponse {
  volunteers: VolunteerProfile[];
}

export interface VerifyVolunteerRequest {
  decision: VolunteerVerificationDecision;
  note: string;
}

export interface AssignVolunteerTaskRequest {
  volunteerId: string;
  type: VolunteerTaskType;
  priority?: IncidentAssignmentPriority;
  instructions: string;
  locationLabel: string;
}

export interface VolunteerTaskUpdate {
  id: string;
  type: "status" | "observation";
  status?: VolunteerTaskStatus;
  note: string;
  safetyStatus: VolunteerSafetyStatus;
  location?: Coordinates;
  escalationRequested: boolean;
  createdBy: string;
  createdAt: string;
}

export interface VolunteerTaskRecord {
  id: string;
  incidentId: string;
  incidentReference: string;
  volunteerId: string;
  volunteerName: string;
  groupId: string;
  type: VolunteerTaskType;
  priority: IncidentAssignmentPriority;
  instructions: string;
  locationLabel: string;
  status: VolunteerTaskStatus;
  safetyRules: string[];
  escalationRequired: boolean;
  assignedBy: string;
  assignedAt: string;
  updatedAt: string;
  acceptedAt?: string;
  completedAt?: string;
  updates: VolunteerTaskUpdate[];
}

export interface VolunteerTaskListResponse {
  tasks: VolunteerTaskRecord[];
}

export interface VolunteerTaskStatusRequest {
  volunteerId: string;
  status: Exclude<VolunteerTaskStatus, "assigned">;
  note?: string;
  safetyStatus?: VolunteerSafetyStatus;
  location?: Coordinates;
}

export interface VolunteerObservationRequest {
  volunteerId: string;
  observation: string;
  safetyStatus?: VolunteerSafetyStatus;
  location?: Coordinates;
  escalationRequested?: boolean;
  media?: string[];
}

export interface IncidentAuditEvent {
  id: string;
  actorUserId: string;
  actorAgencyId?: string;
  actorRole: UserRole | "volunteer";
  action: string;
  targetType: "incident" | "volunteer_profile" | "volunteer_task";
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

export interface MLExplanationFactor {
  feature: string;
  label: string;
  value: string | number | boolean;
  contribution: number;
  direction: "increases_risk" | "reduces_risk";
}

export interface MLPredictionSummary {
  id: string;
  modelVersion: string;
  hazardType: HazardType;
  predictionTime: string;
  targetTime: string;
  cellId: string;
  region: string;
  district: string;
  community: string;
  location?: Coordinates;
  geometry?: AlertTargetGeometry;
  distanceMeters?: number;
  probability: number;
  severity: RiskLevel;
  expectedOnset: string;
  confidence: "low" | "medium" | "high";
  explanationFactors: MLExplanationFactor[];
  inputFeatureSetVersion: string;
  predictionLogId?: string;
  humanReviewRequired: boolean;
  autoPublishAllowed: false;
  source: "baseline_fixture_model" | "ml-service";
}

export interface MLPredictionRequest {
  location: Coordinates;
  requestedBy?: string;
  correlationId?: string;
}

export interface MLPredictionLogRecord {
  id: string;
  predictionId: string;
  modelVersion: string;
  inputFeatureSetVersion: string;
  requestedBy?: string;
  correlationId?: string;
  location: Coordinates;
  storageTarget: "ml_predictions";
  humanReviewRequired: boolean;
  autoPublishAllowed: false;
  createdAt: string;
}

export interface MLPredictionResponse {
  prediction: MLPredictionSummary;
  log: MLPredictionLogRecord;
  safety: {
    humanReviewRequired: boolean;
    autoPublishAllowed: false;
    message: string;
  };
}

export interface MLPredictionLogListResponse {
  logs: MLPredictionLogRecord[];
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

export type ShelterStatus = "open" | "full" | "closed" | "unknown";

export interface ShelterRecord {
  id: string;
  name: string;
  type: "evacuation_shelter" | "temporary_shelter" | "relief_shelter";
  region: string;
  district: string;
  address: string;
  location: Coordinates;
  capacity: number;
  currentOccupancy: number;
  status: ShelterStatus;
  contact: string;
  facilities: string[];
  notes?: string;
  distanceMeters?: number;
  updatedBy?: string;
  updatedAt: string;
}

export type RecoverySupportType =
  | "relief_point"
  | "medical_support"
  | "recovery_registration"
  | "water_point"
  | "family_reunification";

export interface RecoverySupportLocation {
  id: string;
  name: string;
  type: RecoverySupportType;
  region: string;
  district: string;
  address: string;
  location: Coordinates;
  contact: string;
  services: string[];
  hours: string;
  status: ShelterStatus;
  distanceMeters?: number;
  updatedAt: string;
}

export interface ShelterListResponse {
  shelters: ShelterRecord[];
  generatedAt: string;
}

export interface NearbyShelterResponse {
  shelters: ShelterRecord[];
  recoverySupport: RecoverySupportLocation[];
  generatedAt: string;
}

export interface RecoverySupportResponse {
  recoverySupport: RecoverySupportLocation[];
  generatedAt: string;
}

export type ReliefPointStatus = "open" | "limited" | "closed" | "paused";

export type ReliefPointType =
  "food" | "water" | "medical" | "hygiene" | "blankets" | "cash" | "mixed";

export interface ReliefStockCategory {
  category: string;
  quantity: number;
  unit: string;
  lastUpdated: string;
}

export interface ReliefPointRecord {
  id: string;
  name: string;
  type: ReliefPointType;
  region: string;
  district: string;
  address: string;
  location: Coordinates;
  contact: string;
  operatingHours: string;
  eligibility: string;
  schedule: string;
  stockCategories: ReliefStockCategory[];
  status: ReliefPointStatus;
  source: string;
  sourceRef?: string;
  distanceMeters?: number;
  createdBy?: string;
  updatedBy?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ReliefPointListResponse {
  reliefPoints: ReliefPointRecord[];
  generatedAt: string;
}

export interface ReliefPointNearbyResponse {
  reliefPoints: ReliefPointRecord[];
  generatedAt: string;
}

export interface ReliefPointStockHistoryResponse {
  reliefPointId: string;
  history: ReliefStockHistoryEntry[];
  generatedAt: string;
}

export interface ReliefStockHistoryEntry {
  id: string;
  changedBy: string;
  changedAt: string;
  note?: string;
  stockCategories: ReliefStockCategory[];
}

export interface CreateReliefPointRequest {
  name: string;
  type: ReliefPointType;
  region?: string;
  district?: string;
  address?: string;
  location: Coordinates;
  contact?: string;
  operatingHours?: string;
  eligibility?: string;
  schedule?: string;
  stockCategories?: ReliefStockCategory[];
  status?: ReliefPointStatus;
  source?: string;
  sourceRef?: string;
}

export interface UpdateReliefPointRequest {
  name?: string;
  type?: ReliefPointType;
  region?: string;
  district?: string;
  address?: string;
  location?: Coordinates;
  contact?: string;
  operatingHours?: string;
  eligibility?: string;
  schedule?: string;
  stockCategories?: ReliefStockCategory[];
  status?: ReliefPointStatus;
  sourceRef?: string;
}

export type AidRequestCategory =
  | "food"
  | "water"
  | "medical"
  | "hygiene"
  | "shelter"
  | "logistics"
  | "cash"
  | "equipment"
  | "volunteers"
  | "other";

export type AidRequestPriority = "low" | "medium" | "high" | "urgent";

export type AidRequestStatus =
  | "pending_review"
  | "approved"
  | "open"
  | "partially_matched"
  | "fulfilled"
  | "paused"
  | "closed"
  | "rejected";

export type AidPledgeStatus =
  "pledged" | "accepted" | "received" | "cancelled" | "flagged";

export type AidPledgeReviewStatus = "pending_review" | "cleared" | "flagged";

export type AidDonorType =
  | "individual"
  | "business"
  | "ngo"
  | "faith_group"
  | "diaspora"
  | "government"
  | "other";

export interface AidPledgeRecord {
  id: string;
  aidRequestId: string;
  donorName: string;
  donorType: AidDonorType;
  contact: string;
  quantity: number;
  unit: string;
  note?: string;
  status: AidPledgeStatus;
  reviewStatus: AidPledgeReviewStatus;
  fraudReviewNotes?: string;
  reviewedBy?: string;
  pledgedAt: string;
  updatedAt: string;
}

export interface AidRequestRecord {
  id: string;
  title: string;
  category: AidRequestCategory;
  priority: AidRequestPriority;
  status: AidRequestStatus;
  region: string;
  district: string;
  location: Coordinates;
  receivingOrganization: string;
  contact: string;
  quantityNeeded: number;
  quantityUnit: string;
  quantityPledged: number;
  description: string;
  neededBy: string;
  visibility: "public" | "partners_only";
  sourceReliefPointId?: string;
  createdBy: string;
  approvedBy?: string;
  approvalNotes?: string;
  antiFraudNotes?: string;
  pledges: AidPledgeRecord[];
  createdAt: string;
  updatedAt: string;
}

export interface AidRequestListResponse {
  aidRequests: AidRequestRecord[];
  generatedAt: string;
}

export interface AidPledgeListResponse {
  aidRequestId: string;
  pledges: AidPledgeRecord[];
  generatedAt: string;
}

export interface CreateAidRequestRequest {
  title: string;
  category: AidRequestCategory;
  priority: AidRequestPriority;
  region?: string;
  district?: string;
  location: Coordinates;
  receivingOrganization: string;
  contact?: string;
  quantityNeeded: number;
  quantityUnit: string;
  description: string;
  neededBy: string;
  visibility?: "public" | "partners_only";
  sourceReliefPointId?: string;
}

export interface ReviewAidRequestRequest {
  status: "approved" | "open" | "paused" | "closed" | "rejected";
  approvalNotes?: string;
  antiFraudNotes?: string;
}

export interface CreateAidPledgeRequest {
  donorName: string;
  donorType: AidDonorType;
  contact: string;
  quantity: number;
  unit: string;
  note?: string;
}

export interface ReviewAidPledgeRequest {
  status?: AidPledgeStatus;
  reviewStatus?: AidPledgeReviewStatus;
  fraudReviewNotes?: string;
}

export interface ShelterOccupancyUpdateRequest {
  capacity?: number;
  currentOccupancy?: number;
  status?: ShelterStatus;
  notes?: string;
}

export interface ShelterUpdateResponse {
  shelter: ShelterRecord;
}

export type HospitalCapacityStatus =
  "available" | "limited" | "full" | "offline" | "unknown";

export type HospitalEmergencyUnitStatus =
  "open" | "busy" | "divert" | "closed" | "unknown";

export interface HospitalCapacityRecord {
  id: string;
  name: string;
  type: string;
  region: string;
  district: string;
  address: string;
  location: Coordinates;
  contact: string;
  services: string[];
  totalBeds: number;
  availableBeds: number;
  icuBedsAvailable: number;
  maternityBedsAvailable: number;
  pediatricBedsAvailable: number;
  isolationBedsAvailable: number;
  emergencyCapacity: HospitalCapacityStatus;
  emergencyUnitStatus: HospitalEmergencyUnitStatus;
  ambulancesAvailable: number;
  oxygenAvailable: boolean;
  notes?: string;
  source: "manual" | "fixture" | "fixture_adapter" | string;
  sourceRef?: string;
  updatedBy?: string;
  updatedAt: string;
  distanceMeters?: number;
  stale: boolean;
  staleReason?: string;
}

export interface HospitalCapacityResponse {
  facilities: HospitalCapacityRecord[];
  generatedAt: string;
  staleThresholdMinutes: number;
}

export interface HospitalCapacityUpdateRequest {
  totalBeds?: number;
  availableBeds?: number;
  icuBedsAvailable?: number;
  maternityBedsAvailable?: number;
  pediatricBedsAvailable?: number;
  isolationBedsAvailable?: number;
  emergencyCapacity?: HospitalCapacityStatus;
  emergencyUnitStatus?: HospitalEmergencyUnitStatus;
  ambulancesAvailable?: number;
  oxygenAvailable?: boolean;
  notes?: string;
  source?: "manual" | "fixture_adapter" | string;
  sourceRef?: string;
}

export interface HospitalCapacityUpdateResponse {
  facility: HospitalCapacityRecord;
}

export interface HospitalCapacityFixtureRecord {
  facilityId: string;
  availableBeds: number;
  icuBedsAvailable?: number;
  maternityBedsAvailable?: number;
  pediatricBedsAvailable?: number;
  isolationBedsAvailable?: number;
  emergencyCapacity: HospitalCapacityStatus;
  emergencyUnitStatus?: HospitalEmergencyUnitStatus;
  ambulancesAvailable?: number;
  oxygenAvailable?: boolean;
  notes?: string;
}

export interface HospitalCapacityImportRequest {
  source?: "fixture_adapter" | string;
  sourceRef?: string;
  records?: HospitalCapacityFixtureRecord[];
}

export interface HospitalCapacityImportResponse {
  imported: number;
  facilities: HospitalCapacityRecord[];
  generatedAt: string;
  source: string;
}

export type RoadClosureStatus = "active" | "scheduled" | "lifted" | "cancelled";

export type RoadClosureSeverity =
  "low" | "moderate" | "high" | "severe" | "emergency";

export interface RoadClosureLineStringGeometry {
  type: "LineString";
  coordinates: number[][];
}

export interface RoadClosureRecord {
  id: string;
  roadName: string;
  reason?: string;
  status: RoadClosureStatus;
  severity: RoadClosureSeverity;
  source: string;
  sourceRef?: string;
  geometry: RoadClosureLineStringGeometry;
  validFrom: string;
  validTo?: string;
  detourNote?: string;
  distanceMeters?: number;
  createdBy?: string;
  updatedBy?: string;
  createdAt: string;
  updatedAt: string;
}

export interface RoadClosureListResponse {
  closures: RoadClosureRecord[];
  generatedAt: string;
}

export interface RoadClosureResponse {
  closure: RoadClosureRecord;
}

export interface CreateRoadClosureRequest {
  roadName: string;
  reason?: string;
  status?: RoadClosureStatus;
  severity?: RoadClosureSeverity;
  source?: string;
  sourceRef?: string;
  geometry: RoadClosureLineStringGeometry;
  validFrom?: string;
  validTo?: string;
  detourNote?: string;
}

export interface UpdateRoadClosureRequest {
  roadName?: string;
  reason?: string;
  status?: RoadClosureStatus;
  severity?: RoadClosureSeverity;
  source?: string;
  sourceRef?: string;
  geometry?: RoadClosureLineStringGeometry;
  validFrom?: string;
  validTo?: string;
  detourNote?: string;
}

export interface RoadClosureAdapterImportRequest {
  source: string;
  sourceRef?: string;
  roadName: string;
  status: RoadClosureStatus;
  reason?: string;
  geometry: string;
  validFrom: string;
  validTo?: string;
  detour?: string;
}

export interface RoadClosureAdapterImportResponse {
  imported: number;
  closures: RoadClosureRecord[];
  generatedAt: string;
  source: string;
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
  mlPrediction?: MLPredictionSummary;
  nearestShelters: ShelterSummary[];
  nearbyFacilities: EmergencyFacilitySummary[];
  recommendedActions: string[];
}

export type RouteWaypointType = "shelter" | "higher_ground" | "manual";

export interface RouteCoordinates {
  lat: number;
  lng: number;
}

export interface RouteSegment {
  start: RouteCoordinates;
  end: RouteCoordinates;
  distanceMeters: number;
  mode: string;
}

export interface RouteTargetShelter {
  id: string;
  name: string;
  location: RouteCoordinates;
  status: string;
}

export interface RoutePlanRequest {
  origin: RouteCoordinates;
  destination?: RouteCoordinates;
  waypointType: RouteWaypointType;
  avoidRiskLevels?: string[];
  closureBufferMeters?: number;
}

export interface RoutePlanResponse {
  route: RouteCoordinates[];
  segments: RouteSegment[];
  distanceMeters: number;
  estimatedDurationMinutes: number;
  targetShelter?: RouteTargetShelter;
  avoidedClosures: string[];
  avoidedRiskZones: string[];
  disclaimer: string;
  decisionSupport: boolean;
  generatedAt: string;
}

export interface RouteOptionsResponse {
  waypointTypes: string[];
  generatedAt: string;
}

// NADAA-140 — Drone and satellite imagery ingestion

export type ImagerySource = "drone" | "satellite" | "other";

export type ImageryStatus = "active" | "expired";

export interface ImageryGeometry {
  type: "Polygon";
  coordinates: number[][][];
}

export interface ImageryRecord {
  id: string;
  reference: string;
  source: ImagerySource;
  captureTime: string;
  geometry: ImageryGeometry;
  coverageAreaKm2: number;
  resolutionMeters: number;
  license?: string;
  relatedIncidentId?: string;
  relatedRiskZoneId?: string;
  mlWorkflowId?: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
  storagePath: string;
  status: ImageryStatus;
  uploadedBy: string;
  createdAt: string;
  expiresAt: string;
}

export interface ImageryListResponse {
  imagery: ImageryRecord[];
  generatedAt: string;
}

export interface ImageryLifecycleResponse {
  expiredCount: number;
}

export interface ImageryGeoJSONFeatureProperties {
  id: string;
  reference: string;
  source: ImagerySource;
  captureTime: string;
  resolutionMeters: number;
  downloadUrl: string;
}

export interface ImageryGeoJSONFeature {
  type: "Feature";
  geometry: ImageryGeometry;
  properties: ImageryGeoJSONFeatureProperties;
}

export interface ImageryGeoJSONFeatureCollection {
  type: "FeatureCollection";
  features: ImageryGeoJSONFeature[];
}

export interface CreateImageryRequest {
  source: ImagerySource;
  captureTime: string;
  geometry: string;
  coverageAreaKm2: string;
  resolutionMeters: string;
  license?: string;
  relatedIncidentId?: string;
  relatedRiskZoneId?: string;
  mlWorkflowId?: string;
}
