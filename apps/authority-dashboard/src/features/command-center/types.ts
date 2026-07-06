import type {
  AgencyType,
  AlertSeverity,
  AlertTargetType,
  HazardType,
  IncidentAbuseReviewDecision,
  IncidentAssignmentPriority,
  IncidentRecord,
  IncidentStatus,
  RiskLevel,
  ShelterStatus,
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

export type ShelterFormState = {
  shelterId: string;
  capacity: string;
  currentOccupancy: string;
  status: ShelterStatus;
  notes: string;
};

export type AssignmentAgencyOption = {
  id: string;
  name: string;
  type: AgencyType;
  responderLead: string;
};
