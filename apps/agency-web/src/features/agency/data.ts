import { nadaaBrand } from "@nadaa/brand";
import type {
  AidRequestCategory,
  AidRequestPriority,
  HazardType,
  HospitalCapacityStatus,
  IncidentStatus,
  ReliefPointStatus,
  ReliefPointType,
  RiskLevel,
} from "@nadaa/shared-types";
import type {
  HospitalCapacityFormState,
  AidRequestFormState,
  IncidentFilterState,
  ReliefPointFormState,
  ShelterOccupancyFormState,
  StatusFormState,
} from "./types";

export const defaultFilters: IncidentFilterState = {
  hazard: "all",
  severity: "all",
  status: "all",
};

export const initialStatusForm: StatusFormState = {
  status: "reported",
  note: "",
  resolutionNotes: "",
};

export const initialShelterOccupancyForm: ShelterOccupancyFormState = {
  capacity: "",
  currentOccupancy: "",
  status: "open",
  notes: "",
};

export const initialHospitalCapacityForm: HospitalCapacityFormState = {
  totalBeds: "",
  availableBeds: "",
  icuBedsAvailable: "",
  maternityBedsAvailable: "",
  pediatricBedsAvailable: "",
  isolationBedsAvailable: "",
  ambulancesAvailable: "",
  emergencyCapacity: "available",
  emergencyUnitStatus: "open",
  oxygenAvailable: true,
  notes: "",
};

export const initialReliefPointForm: ReliefPointFormState = {
  name: "",
  type: "mixed",
  region: "Greater Accra",
  district: "Accra Metropolitan",
  address: "",
  lat: "5.5600",
  lng: "-0.2000",
  contact: nadaaBrand.supportLine,
  operatingHours: "08:00-18:00",
  eligibility: "",
  schedule: "",
  status: "open",
  stockCategories: "rice_kg:100:kg, water_sachets:500:sachets",
};

export const initialAidRequestForm: AidRequestFormState = {
  title: "",
  category: "food",
  priority: "high",
  region: "Greater Accra",
  district: "Accra Metropolitan",
  lat: "5.5600",
  lng: "-0.2000",
  receivingOrganization: "NADMO Accra Metro",
  contact: nadaaBrand.supportLine,
  quantityNeeded: "100",
  quantityUnit: "units",
  description: "",
  neededBy: new Date(Date.now() + 72 * 60 * 60 * 1000)
    .toISOString()
    .slice(0, 16),
  visibility: "public",
  sourceReliefPointId: "",
};

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

export const severityColors: Record<RiskLevel, string> = {
  emergency: "#7F1D1D",
  severe: nadaaBrand.colors.red,
  high: "#D97706",
  moderate: nadaaBrand.colors.gold,
  low: nadaaBrand.colors.green,
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

export const shelterStatusOptions = ["open", "full", "closed"];

export const hospitalCapacityOptions: HospitalCapacityStatus[] = [
  "available",
  "limited",
  "full",
];

export const hospitalUnitStatusOptions = ["open", "busy", "divert", "closed"];

export const reliefPointStatusOptions: ReliefPointStatus[] = [
  "open",
  "limited",
  "paused",
  "closed",
];

export const reliefPointTypeOptions: ReliefPointType[] = [
  "food",
  "water",
  "medical",
  "hygiene",
  "blankets",
  "cash",
  "mixed",
];

export const aidRequestCategoryOptions: AidRequestCategory[] = [
  "food",
  "water",
  "medical",
  "hygiene",
  "shelter",
  "logistics",
  "cash",
  "equipment",
  "volunteers",
  "other",
];

export const aidRequestPriorityOptions: AidRequestPriority[] = [
  "urgent",
  "high",
  "medium",
  "low",
];

export function statusLabel(status: IncidentStatus) {
  return status
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function hazardLabel(hazard: HazardType) {
  return hazard
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function severityLabel(severity: RiskLevel) {
  return severity.charAt(0).toUpperCase() + severity.slice(1);
}

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

export function requiresResolutionNotes(status: IncidentStatus) {
  return status === "closed" || status === "false_report";
}
