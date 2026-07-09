import { nadaaBrand } from "@nadaa/brand";
import type {
  AgencyType,
  AlertSeverity,
  AlertTargetType,
  AuthorityAlertRecord,
  Coordinates,
  HospitalCapacityRecord,
  IncidentStatus,
  IncidentTriageSeverity,
  RiskLevel,
} from "@nadaa/shared-types";
import type {
  AssignmentAgencyOption,
  CommandIncident,
  FilterState,
  MLPredictionReview,
  MLPredictionReviewPoint,
  TriageSuggestionReview,
} from "./types";

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

export const defaultHospitalCapacityFilters = {
  emergencyCapacity: "all",
  includeStale: true,
  minAvailableBeds: "0",
  service: "all",
} as const;

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
    updatedAt: "2026-07-06T18:55:00Z",
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
    updatedAt: "2026-07-06T18:48:00Z",
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
    updatedAt: "2026-07-06T18:10:00Z",
  },
];

export const fallbackIncidents: CommandIncident[] = [
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
    abuseSignals: [
      {
        code: "reporter_burst",
        label: "Reporter burst",
        detail: "Reporter has submitted multiple nearby reports today.",
        weight: 0.55,
      },
    ],
    abuseScore: 0.55,
    abuseReviewRequired: true,
    abuseReviewReason: "Review requested: Reporter burst",
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
    timeline: [],
    reportedBy: { userId: "usr_ama", phone: "+233200000003" },
    createdAt: "2026-07-06T18:42:00Z",
    updatedAt: "2026-07-06T18:48:00Z",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    locality: "Accra Central",
    assignedAgency: "Unassigned",
    responderEta: "7 min",
    timelineEntries: [
      "Citizen report received with photo evidence",
      "Duplicate reports grouped near Accra Central",
      "NADMO AMA dispatcher reviewing severity",
    ],
    source: "fixture",
  },
  {
    id: "inc_accra_flood_0237",
    reference: "INC-0237",
    type: "flood",
    severity: "high",
    status: "reported",
    description:
      "Resident reports the same market road flooding with stranded taxis.",
    location: { lat: 5.579, lng: -0.212 },
    peopleAffected: 9,
    injuriesReported: false,
    urgency: "high",
    anonymous: false,
    contactPermission: true,
    media: [],
    priorityReview: true,
    abuseSignals: [],
    abuseScore: 0,
    abuseReviewRequired: false,
    duplicateCandidates: [
      {
        incidentId: "inc_accra_flood_0241",
        reference: "INC-0241",
        score: 0.82,
        distanceMeters: 214,
        minutesApart: 16,
        reasons: ["same_hazard", "nearby_location", "recent_report"],
      },
    ],
    mergedIncidentIds: [],
    assignments: [],
    timeline: [],
    reportedBy: { userId: "usr_kofi", phone: "+233200000004" },
    createdAt: "2026-07-06T18:26:00Z",
    updatedAt: "2026-07-06T18:43:00Z",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    locality: "Accra Central",
    assignedAgency: "Unassigned",
    responderEta: "12 min",
    timelineEntries: [
      "Citizen report received near market road",
      "Matched as likely duplicate of INC-0241",
      "Awaiting dispatcher merge review",
    ],
    source: "fixture",
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
        assignedAt: "2026-07-06T18:33:00Z",
      },
    ],
    timeline: [],
    createdAt: "2026-07-06T18:25:00Z",
    updatedAt: "2026-07-06T18:39:00Z",
    region: "Greater Accra",
    district: "Tema Metropolitan",
    locality: "Tema Motorway",
    assignedAgency: "National Ambulance Service",
    responderEta: "12 min",
    timelineEntries: [
      "Dispatcher verified multiple injured persons",
      "Ambulance and police units assigned",
      "Motorway patrol requested lane control",
    ],
    source: "fixture",
  },
  {
    id: "inc_ablekuma_drain_0236",
    reference: "INC-0236",
    type: "blocked_drain",
    severity: "moderate",
    status: "verified",
    description: "Blocked drain backing water into a residential street.",
    location: { lat: 5.601, lng: -0.286 },
    peopleAffected: 14,
    injuriesReported: false,
    urgency: "moderate",
    anonymous: true,
    contactPermission: false,
    media: [],
    priorityReview: false,
    abuseSignals: [],
    abuseScore: 0,
    abuseReviewRequired: false,
    duplicateCandidates: [],
    mergedIncidentIds: [],
    assignments: [],
    timeline: [],
    createdAt: "2026-07-06T17:58:00Z",
    updatedAt: "2026-07-06T18:12:00Z",
    region: "Greater Accra",
    district: "Ablekuma West",
    locality: "Dansoman",
    assignedAgency: "Ready for assignment",
    responderEta: "31 min",
    timelineEntries: [
      "District officer verified blocked drain",
      "Sanitation crew notified",
      "Resident contact hidden due anonymous report",
    ],
    source: "fixture",
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
        assignedAt: "2026-07-06T18:05:00Z",
      },
    ],
    timeline: [],
    createdAt: "2026-07-06T17:41:00Z",
    updatedAt: "2026-07-06T18:19:00Z",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    locality: "Korle Gonno",
    assignedAgency: "Ghana National Fire Service",
    responderEta: "5 min",
    timelineEntries: [
      "Fire service call confirmed smoke visible",
      "Hydrant access checked by dispatcher",
      "Engine crew en route",
    ],
    source: "fixture",
  },
];

export const fallbackAlerts: AuthorityAlertRecord[] = [
  {
    id: "alert_fixture_submitted",
    title: "Accra flood watch",
    hazardType: "flood",
    severity: "warning",
    message: "Heavy rainfall may cause flooding in low-lying communities.",
    target: {
      type: "district",
      ids: ["accra-metropolitan"],
      label: "Accra Metropolitan",
    },
    startsAt: "2026-07-06T19:30:00Z",
    expiresAt: "2026-07-07T07:00:00Z",
    recommendedAction:
      "Avoid flooded roads and prepare to move to higher ground.",
    evacuationRequired: false,
    shelterIds: ["00000000-0000-0000-0000-000000000301"],
    issuingAgencyId: "00000000-0000-0000-0000-000000000101",
    issuedBy: "usr_dispatcher_fixture",
    status: "submitted",
    emergencyOverride: false,
    createdAt: "2026-07-06T18:15:00Z",
    updatedAt: "2026-07-06T18:45:00Z",
    submittedAt: "2026-07-06T18:45:00Z",
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

export const predictionReviewPoints: MLPredictionReviewPoint[] = [
  {
    id: "accra-central",
    label: "Accra Central",
    location: { lat: 5.56, lng: -0.2 },
  },
  {
    id: "accra-north",
    label: "Accra North",
    location: { lat: 5.6037, lng: -0.187 },
  },
  {
    id: "osu",
    label: "Osu",
    location: { lat: 5.55, lng: -0.18 },
  },
  {
    id: "tema-community-one",
    label: "Tema Community One",
    location: { lat: 5.669, lng: -0.016 },
  },
  {
    id: "adum",
    label: "Adum",
    location: { lat: 6.6885, lng: -1.6244 },
  },
];

export const fallbackMLPredictions: MLPredictionReview[] = [
  {
    id: "pred_grid-accra-central-001",
    modelVersion: "flood-logistic-baseline-0.1.0",
    hazardType: "flood",
    predictionTime: "2026-07-06T10:00:00Z",
    targetTime: "2026-07-06T12:00:00Z",
    cellId: "grid-accra-central-001",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    community: "Accra Central",
    location: { lat: 5.56, lng: -0.2 },
    geometry: {
      type: "Polygon",
      coordinates: [
        [
          [-0.21, 5.55],
          [-0.19, 5.55],
          [-0.19, 5.57],
          [-0.21, 5.57],
          [-0.21, 5.55],
        ],
      ],
    },
    probability: 0.9993,
    severity: "severe",
    expectedOnset: "within_24h",
    confidence: "medium",
    explanationFactors: [
      {
        feature: "exposure_score",
        label: "population and land-use exposure",
        value: 0.8103,
        contribution: 0.7004,
        direction: "increases_risk",
      },
      {
        feature: "population_density_per_sq_km",
        label: "population density",
        value: 14600,
        contribution: 0.696,
        direction: "increases_risk",
      },
      {
        feature: "water_level_trend_cm",
        label: "water-level trend",
        value: 28,
        contribution: 0.6573,
        direction: "increases_risk",
      },
      {
        feature: "low_lying_area",
        label: "low-lying area",
        value: true,
        contribution: 0.6048,
        direction: "increases_risk",
      },
    ],
    inputFeatureSetVersion: "flood-risk-features.v1",
    humanReviewRequired: true,
    autoPublishAllowed: false,
    source: "baseline_fixture_model",
    reviewStatus: "needs_review",
  },
  {
    id: "pred_grid-accra-north-002",
    modelVersion: "flood-logistic-baseline-0.1.0",
    hazardType: "flood",
    predictionTime: "2026-07-06T10:00:00Z",
    targetTime: "2026-07-06T12:00:00Z",
    cellId: "grid-accra-north-002",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    community: "Accra North",
    location: { lat: 5.6037, lng: -0.187 },
    geometry: {
      type: "Polygon",
      coordinates: [
        [
          [-0.197, 5.5937],
          [-0.177, 5.5937],
          [-0.177, 5.6137],
          [-0.197, 5.6137],
          [-0.197, 5.5937],
        ],
      ],
    },
    probability: 0.9828,
    severity: "high",
    expectedOnset: "24_to_48h",
    confidence: "medium",
    explanationFactors: [
      {
        feature: "vulnerable_population_pct",
        label: "vulnerable population share",
        value: 21,
        contribution: 0.8183,
        direction: "increases_risk",
      },
      {
        feature: "low_lying_area",
        label: "low-lying area",
        value: true,
        contribution: 0.6048,
        direction: "increases_risk",
      },
      {
        feature: "rainfall_intensity_score",
        label: "rainfall intensity",
        value: 0.7134,
        contribution: 0.3522,
        direction: "increases_risk",
      },
    ],
    inputFeatureSetVersion: "flood-risk-features.v1",
    humanReviewRequired: true,
    autoPublishAllowed: false,
    source: "baseline_fixture_model",
    reviewStatus: "needs_review",
  },
  {
    id: "pred_grid-osu-003",
    modelVersion: "flood-logistic-baseline-0.1.0",
    hazardType: "flood",
    predictionTime: "2026-07-06T10:00:00Z",
    targetTime: "2026-07-06T12:00:00Z",
    cellId: "grid-osu-003",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    community: "Osu",
    location: { lat: 5.55, lng: -0.18 },
    geometry: {
      type: "Polygon",
      coordinates: [
        [
          [-0.19, 5.54],
          [-0.17, 5.54],
          [-0.17, 5.56],
          [-0.19, 5.56],
          [-0.19, 5.54],
        ],
      ],
    },
    probability: 0.9906,
    severity: "high",
    expectedOnset: "24_to_48h",
    confidence: "medium",
    explanationFactors: [
      {
        feature: "low_lying_area",
        label: "low-lying area",
        value: true,
        contribution: 0.6048,
        direction: "increases_risk",
      },
      {
        feature: "exposure_score",
        label: "population and land-use exposure",
        value: 0.7446,
        contribution: 0.4668,
        direction: "increases_risk",
      },
      {
        feature: "impervious_surface_pct",
        label: "impervious surface share",
        value: 74,
        contribution: 0.4301,
        direction: "increases_risk",
      },
    ],
    inputFeatureSetVersion: "flood-risk-features.v1",
    humanReviewRequired: true,
    autoPublishAllowed: false,
    source: "baseline_fixture_model",
    reviewStatus: "needs_review",
  },
  {
    id: "pred_grid-tema-004",
    modelVersion: "flood-logistic-baseline-0.1.0",
    hazardType: "flood",
    predictionTime: "2026-07-06T10:00:00Z",
    targetTime: "2026-07-06T12:00:00Z",
    cellId: "grid-tema-004",
    region: "Greater Accra",
    district: "Tema Metropolitan",
    community: "Tema Community One",
    location: { lat: 5.669, lng: -0.016 },
    geometry: {
      type: "Polygon",
      coordinates: [
        [
          [-0.026, 5.659],
          [-0.006, 5.659],
          [-0.006, 5.679],
          [-0.026, 5.679],
          [-0.026, 5.659],
        ],
      ],
    },
    probability: 0.0264,
    severity: "low",
    expectedOnset: "not_expected_in_72h",
    confidence: "medium",
    explanationFactors: [
      {
        feature: "low_lying_area",
        label: "low-lying area",
        value: false,
        contribution: -0.9072,
        direction: "reduces_risk",
      },
      {
        feature: "vulnerable_population_pct",
        label: "vulnerable population share",
        value: 13,
        contribution: -0.6048,
        direction: "reduces_risk",
      },
    ],
    inputFeatureSetVersion: "flood-risk-features.v1",
    humanReviewRequired: true,
    autoPublishAllowed: false,
    source: "baseline_fixture_model",
    reviewStatus: "needs_review",
  },
  {
    id: "pred_grid-kumasi-005",
    modelVersion: "flood-logistic-baseline-0.1.0",
    hazardType: "flood",
    predictionTime: "2026-07-06T10:00:00Z",
    targetTime: "2026-07-06T12:00:00Z",
    cellId: "grid-kumasi-005",
    region: "Ashanti",
    district: "Kumasi Metropolitan",
    community: "Adum",
    location: { lat: 6.6885, lng: -1.6244 },
    geometry: {
      type: "Polygon",
      coordinates: [
        [
          [-1.6344, 6.6785],
          [-1.6144, 6.6785],
          [-1.6144, 6.6985],
          [-1.6344, 6.6985],
          [-1.6344, 6.6785],
        ],
      ],
    },
    probability: 0.0008,
    severity: "low",
    expectedOnset: "not_expected_in_72h",
    confidence: "medium",
    explanationFactors: [
      {
        feature: "rainfall_intensity_score",
        label: "rainfall intensity",
        value: 0.1946,
        contribution: -0.7818,
        direction: "reduces_risk",
      },
      {
        feature: "rainfall_forecast_24h_mm",
        label: "24-hour rainfall forecast",
        value: 18.5,
        contribution: -0.7448,
        direction: "reduces_risk",
      },
    ],
    inputFeatureSetVersion: "flood-risk-features.v1",
    humanReviewRequired: true,
    autoPublishAllowed: false,
    source: "baseline_fixture_model",
    reviewStatus: "needs_review",
  },
];

// Triage severities the incident-service accepts for overrides (no "severe").
export const triageSeverityOptions: IncidentTriageSeverity[] = [
  "low",
  "moderate",
  "high",
  "emergency",
];

export function fallbackTriageSuggestion(
  incident: CommandIncident,
): TriageSuggestionReview {
  const severity: IncidentTriageSeverity =
    incident.urgency === "life_threatening" || incident.injuriesReported
      ? "emergency"
      : incident.urgency === "high" || incident.peopleAffected >= 20
        ? "high"
        : incident.urgency === "moderate" || incident.peopleAffected >= 5
          ? "moderate"
          : "low";

  const duplicateLikelihood = incident.duplicateCandidates.length
    ? Math.max(
        ...incident.duplicateCandidates.map((candidate) => candidate.score),
      )
    : 0;

  const topDuplicateIncidentIds = incident.duplicateCandidates
    .slice(0, 3)
    .map((candidate) => candidate.incidentId);

  // Mirrors the incident-service triageSuggestedAgency rules.
  const agency =
    incident.type === "fire" ||
    incident.type === "electrical_hazard" ||
    incident.type === "building_collapse"
      ? {
          agencyType: "fire" as AgencyType,
          agencyId: "00000000-0000-0000-0000-000000000201",
          name: "Ghana National Fire Service",
          reason: "Primary responder for fire and structural collapse incidents.",
        }
      : incident.type === "road_crash" || incident.type === "security_incident"
        ? {
            agencyType: "police" as AgencyType,
            agencyId: "00000000-0000-0000-0000-000000000203",
            name: "Ghana Police Service",
            reason:
              incident.type === "road_crash"
                ? "Traffic and scene control for road crashes."
                : "Law enforcement lead for security incidents.",
          }
        : incident.type === "medical_emergency" ||
            incident.type === "disease_outbreak"
          ? {
              agencyType: "ambulance" as AgencyType,
              agencyId: "00000000-0000-0000-0000-000000000202",
              name: "National Ambulance Service",
              reason: "Primary responder for medical and health incidents.",
            }
          : incident.type === "blocked_drain"
            ? {
                agencyType: "district_assembly" as AgencyType,
                agencyId: "00000000-0000-0000-0000-000000000204",
                name: "Accra Metropolitan Assembly",
                reason: "Local drainage and sanitation works responsibility.",
              }
            : {
                agencyType: "nadmo" as AgencyType,
                agencyId: "00000000-0000-0000-0000-000000000101",
                name: "NADMO Accra Metro",
                reason: "NADMO coordinates multi-hazard disaster response.",
              };

  return {
    suggestionId: "",
    severity,
    duplicateLikelihood: Math.round(duplicateLikelihood * 100) / 100,
    topDuplicateIncidentIds,
    affectedPopulation: incident.peopleAffected || 1,
    suggestedAgency: agency,
    confidence:
      incident.peopleAffected > 0 && incident.urgency !== "low"
        ? "high"
        : incident.peopleAffected > 0 || incident.urgency !== "low"
          ? "medium"
          : "low",
    modelVersion: "incident-triage-rules-0.1.0-fixture",
    featureSetVersion: "incident-features.v1",
    explanationFactors: [
      {
        feature: "urgency",
        label: "Reported urgency",
        value: incident.urgency,
        contribution:
          incident.urgency === "life_threatening"
            ? 0.9
            : incident.urgency === "high"
              ? 0.6
              : incident.urgency === "moderate"
                ? 0.3
                : 0.1,
        direction: "increases_risk",
      },
      {
        feature: "people_affected",
        label: "People directly affected",
        value: incident.peopleAffected,
        contribution:
          incident.peopleAffected >= 20
            ? 0.5
            : incident.peopleAffected >= 5
              ? 0.3
              : 0.1,
        direction: "increases_risk",
      },
      {
        feature: "hazard_type",
        label: "Hazard type",
        value: incident.type,
        contribution: 0.3,
        direction: "increases_risk",
      },
      {
        feature: "duplicate_candidates",
        label: "Duplicate report candidates",
        value: incident.duplicateCandidates.length,
        contribution: duplicateLikelihood,
        direction: "increases_risk",
      },
    ],
    humanReviewRequired: true,
    autoPublishAllowed: false,
    incidentId: incident.id,
  };
}

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
