import type {
  HazardType,
  HospitalCapacityRecord,
  HospitalCapacityStatus,
  IncidentRecord,
  IncidentStatus,
  ReliefPointStatus,
  ReliefPointType,
  RiskLevel,
  ShelterRecord,
} from "@nadaa/shared-types";

export type AgencyTab = "dashboard" | "incident" | "capacity" | "relief";

export type IncidentLoadState =
  "loading" | "ready" | "fallback" | "empty" | "error";
export type CapacityLoadState =
  "loading" | "ready" | "fallback" | "empty" | "error";
export type UpdateLoadState = "idle" | "loading" | "success" | "error";

export type IncidentFilterState = {
  hazard: "all" | HazardType;
  severity: "all" | RiskLevel;
  status: "all" | IncidentStatus;
};

export type StatusFormState = {
  status: IncidentStatus;
  note: string;
  resolutionNotes: string;
};

export type ShelterOccupancyFormState = {
  capacity: string;
  currentOccupancy: string;
  status: string;
  notes: string;
};

export type HospitalCapacityFormState = {
  totalBeds: string;
  availableBeds: string;
  icuBedsAvailable: string;
  maternityBedsAvailable: string;
  pediatricBedsAvailable: string;
  isolationBedsAvailable: string;
  ambulancesAvailable: string;
  emergencyCapacity: HospitalCapacityStatus;
  emergencyUnitStatus: string;
  oxygenAvailable: boolean;
  notes: string;
};

export type ReliefPointFormState = {
  name: string;
  type: ReliefPointType;
  region: string;
  district: string;
  address: string;
  lat: string;
  lng: string;
  contact: string;
  operatingHours: string;
  eligibility: string;
  schedule: string;
  status: ReliefPointStatus;
  stockCategories: string;
};

export type AgencyIncident = IncidentRecord & {
  source: "api" | "fixture";
};
