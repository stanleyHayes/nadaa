import { nadaaBrand } from "@nadaa/brand";
import type {
  AgencyType,
  HazardType,
  HospitalCapacityRecord,
  IncidentRecord,
  IncidentStatus,
  RiskLevel,
} from "@nadaa/shared-types";
import type {
  AssignmentFormState,
  AuthFormState,
  CapacityFilterState,
  DispatcherPermissionState,
  DispatcherSession,
  IncidentFilterState,
  StatusFormState,
  TimelineNoteFormState,
} from "./types";

const generatedAt = new Date().toISOString();

export const fixtureDispatcherSession: DispatcherSession = {
  agencyId: "00000000-0000-0000-0000-000000000101",
  agencyName: "NADMO Accra Dispatch",
  mfaCompleted: true,
  role: "dispatcher",
  userId: "usr_dispatch_accra",
  userName: "Accra Dispatcher",
};

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

export const fallbackIncidents: IncidentRecord[] = [
  {
    id: "inc_accra_flood_0241",
    reference: "INC-0241",
    type: "flood",
    severity: "severe",
    status: "under_review",
    description:
      "Water is rising near a low-lying road and vehicles are trapped.",
    location: { lat: 5.579, lng: -0.212 },
    peopleAffected: 28,
    injuriesReported: false,
    urgency: "life_threatening",
    anonymous: false,
    contactPermission: true,
    media: ["media_flood_photo_001"],
    priorityReview: true,
    abuseSignals: [],
    abuseScore: 0,
    abuseReviewRequired: false,
    duplicateCandidates: [
      {
        incidentId: "inc_accra_flood_0237",
        reference: "INC-0237",
        score: 0.82,
        distanceMeters: 214,
        minutesApart: 16,
        reasons: ["same_hazard", "nearby_location", "recent_report"],
      },
    ],
    mergedIncidentIds: [],
    assignments: [],
    timeline: [
      {
        id: "tline_0241_001",
        type: "incident.reported",
        message: "Citizen report received with photo evidence",
        actorUserId: "usr_ama",
        createdAt: generatedAt,
      },
    ],
    reportedBy: { userId: "usr_ama", phone: "+233200000003" },
    createdAt: generatedAt,
    updatedAt: generatedAt,
  },
  {
    id: "inc_tema_crash_0239",
    reference: "INC-0239",
    type: "road_crash",
    severity: "high",
    status: "assigned",
    description: "Three-vehicle crash blocking the Tema motorway shoulder.",
    location: { lat: 5.642, lng: -0.028 },
    peopleAffected: 11,
    injuriesReported: true,
    urgency: "high",
    anonymous: false,
    contactPermission: true,
    media: ["media_crash_photo_002"],
    priorityReview: true,
    abuseSignals: [],
    abuseScore: 0,
    abuseReviewRequired: false,
    duplicateCandidates: [],
    mergedIncidentIds: [],
    assignments: [
      {
        id: "asg_fixture_tema",
        agencyId: "00000000-0000-0000-0000-000000000202",
        agencyName: "National Ambulance Service",
        agencyType: "ambulance",
        priority: "high",
        instructions: "Attend crash scene and coordinate casualty transport.",
        responderLead: "Ambulance Control Lead",
        status: "active",
        assignedBy: "usr_dispatcher_fixture",
        assignedAt: generatedAt,
      },
    ],
    timeline: [
      {
        id: "tline_0239_001",
        type: "incident.reported",
        message: "Crash reported on Tema motorway shoulder",
        createdAt: generatedAt,
      },
      {
        id: "tline_0239_002",
        type: "incident.assigned",
        message: "Assigned to National Ambulance Service",
        actorUserId: "usr_dispatcher_fixture",
        actorAgencyId: "00000000-0000-0000-0000-000000000101",
        actorRole: "dispatcher",
        createdAt: generatedAt,
      },
    ],
    createdAt: generatedAt,
    updatedAt: generatedAt,
  },
  {
    id: "inc_korle_fire_0232",
    reference: "INC-0232",
    type: "fire",
    severity: "high",
    status: "response_en_route",
    description: "Electrical fire reported behind a market stall.",
    location: { lat: 5.544, lng: -0.213 },
    peopleAffected: 8,
    injuriesReported: false,
    urgency: "high",
    anonymous: false,
    contactPermission: true,
    media: ["media_fire_photo_003"],
    priorityReview: true,
    abuseSignals: [],
    abuseScore: 0,
    abuseReviewRequired: false,
    duplicateCandidates: [],
    mergedIncidentIds: [],
    assignments: [
      {
        id: "asg_fixture_fire",
        agencyId: "00000000-0000-0000-0000-000000000201",
        agencyName: "Ghana National Fire Service",
        agencyType: "fire",
        priority: "urgent",
        instructions: "Dispatch engine crew and secure hydrant access.",
        responderLead: "Station Officer Mensah",
        status: "active",
        assignedBy: "usr_dispatcher_fixture",
        assignedAt: generatedAt,
      },
    ],
    timeline: [
      {
        id: "tline_0232_001",
        type: "incident.reported",
        message: "Fire reported behind market stall",
        createdAt: generatedAt,
      },
    ],
    createdAt: generatedAt,
    updatedAt: generatedAt,
  },
];

export const fallbackHospitalFacilities: HospitalCapacityRecord[] = [
  {
    address: "Korle Bu emergency entrance",
    ambulancesAvailable: 3,
    availableBeds: 46,
    contact: "0302665401",
    district: "Accra Metropolitan",
    distanceMeters: 4200,
    emergencyCapacity: "available",
    emergencyUnitStatus: "open",
    icuBedsAvailable: 4,
    id: "hospital_001",
    isolationBedsAvailable: 3,
    location: { lat: 5.536, lng: -0.227 },
    maternityBedsAvailable: 9,
    name: "Korle Bu Teaching Hospital",
    notes: "Major referral facility for Accra emergency transfers.",
    oxygenAvailable: true,
    pediatricBedsAvailable: 5,
    region: "Greater Accra",
    services: [
      "emergency",
      "trauma",
      "icu",
      "maternity",
      "pediatric",
      "oxygen",
    ],
    source: "fixture",
    sourceRef: "hospital-capacity-feed",
    stale: false,
    totalBeds: 820,
    type: "teaching_hospital",
    updatedAt: generatedAt,
  },
  {
    address: "Ridge Hospital emergency unit",
    ambulancesAvailable: 1,
    availableBeds: 12,
    contact: "0302425201",
    district: "Accra Metropolitan",
    distanceMeters: 1100,
    emergencyCapacity: "limited",
    emergencyUnitStatus: "busy",
    icuBedsAvailable: 1,
    id: "hospital_002",
    isolationBedsAvailable: 0,
    location: { lat: 5.563, lng: -0.191 },
    maternityBedsAvailable: 3,
    name: "Greater Accra Regional Hospital",
    notes: "Emergency unit busy; use for critical stabilization.",
    oxygenAvailable: true,
    pediatricBedsAvailable: 2,
    region: "Greater Accra",
    services: ["emergency", "trauma", "oxygen", "ambulance"],
    source: "fixture",
    sourceRef: "hospital-capacity-feed",
    stale: false,
    totalBeds: 420,
    type: "regional_hospital",
    updatedAt: generatedAt,
  },
  {
    address: "Tema Community 12",
    ambulancesAvailable: 0,
    availableBeds: 0,
    contact: "0303202231",
    district: "Tema Metropolitan",
    distanceMeters: 18200,
    emergencyCapacity: "full",
    emergencyUnitStatus: "divert",
    icuBedsAvailable: 0,
    id: "hospital_003",
    isolationBedsAvailable: 0,
    location: { lat: 5.669, lng: -0.016 },
    maternityBedsAvailable: 1,
    name: "Tema General Hospital",
    notes: "Capacity stale; confirm before transfer.",
    oxygenAvailable: false,
    pediatricBedsAvailable: 0,
    region: "Greater Accra",
    services: ["emergency", "maternity", "pediatric", "ambulance"],
    source: "fixture",
    sourceRef: "hospital-capacity-feed",
    stale: true,
    staleReason: "capacity update older than 30 minutes",
    totalBeds: 310,
    type: "general_hospital",
    updatedAt: generatedAt,
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
