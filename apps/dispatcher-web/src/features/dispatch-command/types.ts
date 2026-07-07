import type {
  AgencyType,
  AlertSeverity,
  AlertTargetType,
  Coordinates,
  HazardType,
  HospitalCapacityStatus,
  IncidentAbuseReviewDecision,
  IncidentAssignmentPriority,
  IncidentRecord,
  IncidentStatus,
  MLPredictionSummary,
  RiskLevel,
} from "@nadaa/shared-types";

export type CommandIncident = IncidentRecord & {
  region: string;
  district: string;
  locality: string;
  assignedAgency: string;
  responderEta: string;
  timelineEntries: string[];
  source: "api" | "fixture";
};

export type FilterState = {
  hazard: "all" | HazardType;
  regionDistrict: "all" | string;
  severity: "all" | RiskLevel;
  status: "all" | IncidentStatus;
  time: "all" | "1h" | "6h" | "24h";
};

export type IncidentLoadState =
  "loading" | "ready" | "fallback" | "empty" | "error";
export type AlertLoadState = "loading" | "ready" | "fallback" | "error";
export type MLReviewLoadState = "loading" | "ready" | "fallback" | "error";
export type CapacityLoadState = "loading" | "ready" | "fallback" | "empty";

export type HospitalCapacityFilterState = {
  emergencyCapacity: "all" | HospitalCapacityStatus;
  includeStale: boolean;
  minAvailableBeds: string;
  service:
    | "all"
    | "emergency"
    | "trauma"
    | "icu"
    | "maternity"
    | "pediatric"
    | "ambulance"
    | "oxygen";
};

export type MLPredictionReviewPoint = {
  id: string;
  label: string;
  location: Coordinates;
};

export type MLPredictionReview = MLPredictionSummary & {
  location: Coordinates;
  reviewStatus: "needs_review" | "draft_created";
};

export type AlertFormState = {
  title: string;
  severity: AlertSeverity;
  message: string;
  targetType: AlertTargetType;
  targetIds: string;
  targetLabel: string;
  targetLatitude: string;
  targetLongitude: string;
  targetRadiusMeters: string;
  targetGeometry: string;
  startsAt: string;
  expiresAt: string;
  recommendedAction: string;
  evacuationRequired: boolean;
  shelterIds: string;
};

export type IncidentStatusFormState = {
  status: IncidentStatus;
  note: string;
  resolutionNotes: string;
};

export type AbuseReviewFormState = {
  decision: IncidentAbuseReviewDecision;
  note: string;
  resolutionNotes: string;
};

export type AssignmentFormState = {
  agencyId: string;
  agencyName: string;
  agencyType: AgencyType;
  priority: IncidentAssignmentPriority;
  instructions: string;
  responderLead: string;
};

export type AssignmentAgencyOption = {
  id: string;
  name: string;
  type: AgencyType;
  responderLead: string;
};
