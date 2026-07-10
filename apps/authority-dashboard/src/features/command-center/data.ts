import { nadaaBrand } from "@nadaa/brand";
import type {
  AlertSeverity,
  AlertTargetType,
  Coordinates,
  IncidentStatus,
  RiskLevel,
} from "@nadaa/shared-types";
import type { AssignmentAgencyOption, FilterState } from "./types";

export const assignmentAgencyOptions: AssignmentAgencyOption[] = [
  {
    id: "00000000-0000-0000-0000-000000000101",
    name: "NADMO Accra Metro",
    type: "nadmo",
    responderLead: "NADMO Duty Officer",
  },
  {
    id: "00000000-0000-0000-0000-000000000201",
    name: "Ghana National Fire Service",
    type: "fire",
    responderLead: "Station Officer Mensah",
  },
  {
    id: "00000000-0000-0000-0000-000000000202",
    name: "National Ambulance Service",
    type: "ambulance",
    responderLead: "Ambulance Control Lead",
  },
  {
    id: "00000000-0000-0000-0000-000000000203",
    name: "Ghana Police Service",
    type: "police",
    responderLead: "Motor Traffic Lead",
  },
  {
    id: "00000000-0000-0000-0000-000000000204",
    name: "Accra Metropolitan Assembly",
    type: "district_assembly",
    responderLead: "Metro Works Supervisor",
  },
];

export const defaultFilters: FilterState = {
  hazard: "all",
  regionDistrict: "all",
  severity: "all",
  status: "all",
  time: "all",
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

export const alertSeverityOptions: AlertSeverity[] = [
  "advisory",
  "watch",
  "warning",
  "severe_warning",
  "emergency",
];

export const alertTargetTypeOptions: AlertTargetType[] = [
  "district",
  "radius",
  "custom",
  "region",
  "community",
  "national",
];

export type AlertTargetCatalogItem = {
  id: string;
  type: AlertTargetType;
  label: string;
  center: Coordinates;
  radiusMeters: number;
  areaSqKm: number;
  estimatedPopulation: number;
};

export const alertTargetCatalog: Record<string, AlertTargetCatalogItem> = {
  "region:greater-accra": {
    id: "greater-accra",
    type: "region",
    label: "Greater Accra Region",
    center: { lat: 5.75, lng: -0.11 },
    radiusMeters: 52000,
    areaSqKm: 3245,
    estimatedPopulation: 5455000,
  },
  "district:accra-metropolitan": {
    id: "accra-metropolitan",
    type: "district",
    label: "Accra Metropolitan",
    center: { lat: 5.56, lng: -0.2 },
    radiusMeters: 9000,
    areaSqKm: 61,
    estimatedPopulation: 284000,
  },
  "district:tema-metropolitan": {
    id: "tema-metropolitan",
    type: "district",
    label: "Tema Metropolitan",
    center: { lat: 5.642, lng: -0.028 },
    radiusMeters: 12000,
    areaSqKm: 565,
    estimatedPopulation: 402000,
  },
  "district:ablekuma-west": {
    id: "ablekuma-west",
    type: "district",
    label: "Ablekuma West",
    center: { lat: 5.601, lng: -0.286 },
    radiusMeters: 7000,
    areaSqKm: 15,
    estimatedPopulation: 220000,
  },
  "community:accra-central": {
    id: "accra-central",
    type: "community",
    label: "Accra Central",
    center: { lat: 5.556, lng: -0.202 },
    radiusMeters: 3000,
    areaSqKm: 8,
    estimatedPopulation: 75000,
  },
};

export const incidentStatusOptions: IncidentStatus[] = [
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
