import { nadaaBrand } from "@nadaa/brand";
import type {
  AgencyType,
  HazardType,
  IncidentStatus,
  RiskLevel,
} from "@nadaa/shared-types";
import type {
  AssignmentFormState,
  AuthFormState,
  CapacityFilterState,
  DispatcherPermissionState,
  IncidentFilterState,
  StatusFormState,
  TimelineNoteFormState,
} from "./types";

export const initialPermissions: DispatcherPermissionState = {
  camera: "unknown",
  location: "unknown",
  push: "unknown",
};

export const defaultFilters: IncidentFilterState = {
  hazard: "all",
  severity: "all",
  status: "all",
  time: "all",
};

export const defaultCapacityFilters: CapacityFilterState = {
  emergencyCapacity: "all",
  includeStale: true,
  minAvailableBeds: "0",
  service: "all",
};

export const initialAuthForm: AuthFormState = {
  email: "",
  mfaCode: "",
  password: "",
};

export const initialStatusForm: StatusFormState = {
  note: "",
  resolutionNotes: "",
  status: "reported",
};

export const initialAssignmentForm: AssignmentFormState = {
  agencyId: "",
  agencyName: "",
  agencyType: "nadmo",
  instructions: "",
  priority: "normal",
  responderLead: "",
};

export const initialTimelineNoteForm: TimelineNoteFormState = {
  note: "",
};

export const hazardOptions: Array<{ label: string; value: HazardType }> = [
  { label: "Flood", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Road crash", value: "road_crash" },
  { label: "Building collapse", value: "building_collapse" },
  { label: "Medical emergency", value: "medical_emergency" },
  { label: "Security incident", value: "security_incident" },
  { label: "Disease outbreak", value: "disease_outbreak" },
  { label: "Electrical hazard", value: "electrical_hazard" },
  { label: "Blocked drain", value: "blocked_drain" },
  { label: "Landslide", value: "landslide" },
  { label: "Marine accident", value: "marine_accident" },
  { label: "Storm", value: "storm" },
  { label: "Tidal wave", value: "tidal_wave" },
  { label: "Other", value: "other" },
];

export const statusOptions: IncidentStatus[] = [
  "reported",
  "under_review",
  "verified",
  "assigned",
  "response_en_route",
  "on_scene",
  "contained",
  "recovery_ongoing",
  "closed",
  "false_report",
];

export const incidentTransitionOptions: Record<
  IncidentStatus,
  IncidentStatus[]
> = {
  reported: ["under_review", "verified", "false_report"],
  under_review: ["verified", "false_report"],
  verified: ["assigned", "response_en_route", "false_report"],
  assigned: [
    "response_en_route",
    "on_scene",
    "contained",
    "recovery_ongoing",
    "closed",
    "false_report",
  ],
  response_en_route: [
    "on_scene",
    "contained",
    "recovery_ongoing",
    "closed",
    "false_report",
  ],
  on_scene: ["contained", "recovery_ongoing", "closed", "false_report"],
  contained: ["recovery_ongoing", "closed", "false_report"],
  recovery_ongoing: ["closed", "false_report"],
  closed: [],
  false_report: [],
};

export const severityOrder: Record<RiskLevel, number> = {
  emergency: 5,
  severe: 4,
  high: 3,
  moderate: 2,
  low: 1,
};

export const assignmentAgencyOptions: Array<{
  id: string;
  name: string;
  responderLead: string;
  type: AgencyType;
}> = [
  {
    id: "00000000-0000-0000-0000-000000000101",
    name: "NADMO Accra Metro",
    responderLead: "NADMO Duty Officer",
    type: "nadmo",
  },
  {
    id: "00000000-0000-0000-0000-000000000201",
    name: "Ghana National Fire Service",
    responderLead: "Station Officer Mensah",
    type: "fire",
  },
  {
    id: "00000000-0000-0000-0000-000000000202",
    name: "National Ambulance Service",
    responderLead: "Ambulance Control Lead",
    type: "ambulance",
  },
  {
    id: "00000000-0000-0000-0000-000000000203",
    name: "Ghana Police Service",
    responderLead: "Motor Traffic Lead",
    type: "police",
  },
  {
    id: "00000000-0000-0000-0000-000000000204",
    name: "Accra Metropolitan Assembly",
    responderLead: "Metro Works Supervisor",
    type: "district_assembly",
  },
];

export function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("en-GH", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function formatRelativeTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  const minutes = Math.floor((Date.now() - date.getTime()) / (1000 * 60));
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

export function statusLabel(status: IncidentStatus) {
  return status
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function severityTone(
  severity: RiskLevel,
): "danger" | "gold" | "green" | "navy" {
  switch (severity) {
    case "emergency":
    case "severe":
      return "danger";
    case "high":
      return "gold";
    case "moderate":
      return "navy";
    case "low":
    default:
      return "green";
  }
}

export const supportLine = nadaaBrand.supportLine;
