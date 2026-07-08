import { nadaaBrand } from "@nadaa/brand";
import type {
  AgencyType,
  AidRequestCategory,
  AidRequestPriority,
  AidRequestRecord,
  HazardType,
  HospitalCapacityRecord,
  HospitalCapacityStatus,
  IncidentRecord,
  IncidentStatus,
  ReliefPointRecord,
  ReliefPointStatus,
  ReliefPointType,
  RiskLevel,
  ShelterRecord,
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

export const fallbackIncidents: IncidentRecord[] = [
  {
    id: "inc_accra_flood_0241",
    reference: "INC-0241",
    type: "flood",
    severity: "severe",
    status: "assigned",
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
    duplicateCandidates: [],
    mergedIncidentIds: [],
    assignments: [
      {
        id: "asg_fixture_001",
        agencyId: "00000000-0000-0000-0000-000000000101",
        agencyName: "NADMO Accra Metro",
        agencyType: "nadmo" as AgencyType,
        priority: "urgent",
        instructions: "Coordinate evacuation and traffic control.",
        responderLead: "NADMO Duty Officer",
        status: "active",
        assignedBy: "usr_dispatcher_fixture",
        assignedAt: new Date().toISOString(),
      },
    ],
    timeline: [
      {
        id: "tline_001",
        type: "incident.reported",
        message: "Citizen report received with photo evidence",
        createdAt: new Date().toISOString(),
      },
      {
        id: "tline_002",
        type: "incident.assigned",
        message: "Assigned to NADMO Accra Metro",
        actorUserId: "usr_dispatcher_fixture",
        actorAgencyId: "00000000-0000-0000-0000-000000000101",
        actorRole: "dispatcher",
        createdAt: new Date().toISOString(),
      },
    ],
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: "inc_tema_crash_0239",
    reference: "INC-0239",
    type: "road_crash",
    severity: "high",
    status: "response_en_route",
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
        id: "asg_fixture_002",
        agencyId: "00000000-0000-0000-0000-000000000101",
        agencyName: "NADMO Accra Metro",
        agencyType: "nadmo" as AgencyType,
        priority: "high",
        instructions: "Support ambulance and police on scene.",
        responderLead: "NADMO Duty Officer",
        status: "active",
        assignedBy: "usr_dispatcher_fixture",
        assignedAt: new Date().toISOString(),
      },
    ],
    timeline: [
      {
        id: "tline_003",
        type: "incident.reported",
        message: "Crash reported on Tema motorway shoulder",
        createdAt: new Date().toISOString(),
      },
    ],
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const fallbackShelters: ShelterRecord[] = [
  {
    id: "shelter-ama-001",
    name: "Accra Metro Assembly Shelter",
    type: "evacuation_shelter",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    address: "Accra Metropolitan Assembly Hall",
    location: { lat: 5.56, lng: -0.2 },
    capacity: 450,
    currentOccupancy: 116,
    status: "open",
    contact: nadaaBrand.supportLine,
    facilities: ["water", "first_aid", "accessible_entry", "family_area"],
    notes: "Primary flood evacuation shelter for central Accra.",
    updatedAt: new Date().toISOString(),
  },
];

export const fallbackHospitals: HospitalCapacityRecord[] = [
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
    updatedAt: new Date().toISOString(),
  },
];

export const fallbackReliefPoints: ReliefPointRecord[] = [
  {
    id: "relief_ama_food_001",
    name: "AMA Food Relief Point",
    type: "food",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    address: "Independence Avenue relief desk",
    location: { lat: 5.558, lng: -0.197 },
    contact: nadaaBrand.supportLine,
    operatingHours: "08:00-20:00",
    eligibility: "Flood-affected households with district registration.",
    schedule: "Daily distribution until response stands down.",
    stockCategories: [
      {
        category: "rice_kg",
        quantity: 420,
        unit: "kg",
        lastUpdated: new Date().toISOString(),
      },
      {
        category: "water_sachets",
        quantity: 1800,
        unit: "sachets",
        lastUpdated: new Date().toISOString(),
      },
    ],
    status: "open",
    source: "fixture",
    sourceRef: "agency-fixture",
    createdBy: "usr_fixture",
    updatedBy: "usr_fixture",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const fallbackAidRequests: AidRequestRecord[] = [
  {
    id: "aid_ama_hygiene_001",
    title: "Hygiene kits for displaced households",
    category: "hygiene",
    priority: "high",
    status: "open",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    location: { lat: 5.56, lng: -0.2 },
    receivingOrganization: "AMA Central Food Distribution",
    contact: nadaaBrand.supportLine,
    quantityNeeded: 300,
    quantityUnit: "kits",
    quantityPledged: 80,
    description:
      "Hygiene kits for families temporarily staying around the central distribution point.",
    neededBy: new Date(Date.now() + 72 * 60 * 60 * 1000).toISOString(),
    visibility: "public",
    sourceReliefPointId: "relief_ama_food_001",
    createdBy: "usr_fixture",
    approvedBy: "usr_fixture",
    approvalNotes: "Verified by district relief desk.",
    antiFraudNotes: "Receiving organization and point contact confirmed.",
    pledges: [
      {
        id: "pledge_hygiene_001",
        aidRequestId: "aid_ama_hygiene_001",
        donorName: "Accra Mutual Aid Network",
        donorType: "ngo",
        contact: "aiddesk@example.org",
        quantity: 80,
        unit: "kits",
        note: "Delivery available within 24 hours after acceptance.",
        status: "accepted",
        reviewStatus: "cleared",
        fraudReviewNotes: "Known partner; contact verified.",
        reviewedBy: "usr_fixture",
        pledgedAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ],
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];
