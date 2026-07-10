import { nadaaBrand } from "@nadaa/brand";
import type {
  AgencyType,
  AlertSeverity,
  AlertTargetType,
  Coordinates,
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
          reason:
            "Primary responder for fire and structural collapse incidents.",
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
