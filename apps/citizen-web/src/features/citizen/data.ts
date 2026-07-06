import type {
  AreaRiskResponse,
  CitizenAlertFeedItem,
  EmergencyGuideRecord,
  HazardType,
  IncidentMediaContentType,
  IncidentUrgency,
  NearbyShelterResponse,
  RiskLevel,
} from "@nadaa/shared-types";
import type { GuideHazardFilter, GuideStageFilter, ReportForm } from "./types";

export const riskTone: Record<
  RiskLevel,
  "success" | "warning" | "error" | "info"
> = {
  low: "success",
  moderate: "info",
  high: "warning",
  severe: "error",
  emergency: "error",
};

export const sampleRisk: AreaRiskResponse = {
  location: "Accra Central",
  overallRisk: "high",
  risks: [
    {
      type: "flood",
      level: "severe",
      probability: 0.82,
      reason:
        "Heavy rainfall forecast, low elevation, and historical flood reports nearby.",
    },
    {
      type: "fire",
      level: "moderate",
      probability: 0.34,
      reason:
        "Dense market activity and recent dry periods increase localized risk.",
    },
  ],
  nearestShelters: [
    {
      id: "shelter-ama-001",
      name: "Accra Metro Assembly Shelter",
      location: { lat: 5.56, lng: -0.2 },
      capacity: 450,
      currentOccupancy: 116,
      contact: "112",
    },
    {
      id: "shelter-osu-002",
      name: "Osu Community Hall",
      location: { lat: 5.55, lng: -0.18 },
      capacity: 220,
      currentOccupancy: 34,
      contact: "112",
    },
  ],
  nearbyFacilities: [
    {
      id: "agency-nadmo-ama",
      name: "NADMO Accra Metro",
      type: "nadmo",
      location: { lat: 5.56, lng: -0.2 },
      contact: "112",
    },
  ],
  recommendedActions: [
    "Avoid low-lying roads and open drains.",
    "Move valuables above ground level.",
    "Prepare an evacuation route to the nearest safe shelter.",
  ],
};

const sampleGeneratedAt = new Date().toISOString();

export const sampleShelterResponse: NearbyShelterResponse = {
  generatedAt: sampleGeneratedAt,
  shelters: [
    {
      id: "00000000-0000-0000-0000-000000000301",
      name: "Accra Metro Assembly Shelter",
      type: "evacuation_shelter",
      region: "Greater Accra",
      district: "Accra Metropolitan",
      address: "Accra Metropolitan Assembly Hall",
      location: { lat: 5.56, lng: -0.2 },
      capacity: 450,
      currentOccupancy: 116,
      status: "open",
      contact: "112",
      facilities: ["water", "first_aid", "accessible_entry", "family_area"],
      notes: "Primary flood evacuation shelter for central Accra.",
      distanceMeters: 0,
      updatedAt: sampleGeneratedAt,
    },
    {
      id: "00000000-0000-0000-0000-000000000302",
      name: "Osu Community Hall",
      type: "temporary_shelter",
      region: "Greater Accra",
      district: "Korle Klottey",
      address: "Osu Community Hall",
      location: { lat: 5.55, lng: -0.18 },
      capacity: 220,
      currentOccupancy: 34,
      status: "open",
      contact: "112",
      facilities: ["water", "first_aid", "family_area"],
      notes: "Short-term shelter and reunification point.",
      distanceMeters: 2480,
      updatedAt: sampleGeneratedAt,
    },
  ],
  recoverySupport: [
    {
      id: "recovery_ama_relief_001",
      name: "AMA Relief Distribution Point",
      type: "relief_point",
      region: "Greater Accra",
      district: "Accra Metropolitan",
      address: "Independence Avenue recovery desk",
      location: { lat: 5.558, lng: -0.197 },
      contact: "112",
      services: ["food", "water", "blankets", "family_reunification"],
      hours: "08:00-20:00",
      status: "open",
      distanceMeters: 420,
      updatedAt: sampleGeneratedAt,
    },
    {
      id: "recovery_osu_registration_001",
      name: "Osu Recovery Registration Desk",
      type: "recovery_registration",
      region: "Greater Accra",
      district: "Korle Klottey",
      address: "Osu Community Hall annex",
      location: { lat: 5.551, lng: -0.181 },
      contact: "112",
      services: ["needs_registration", "damage_reporting", "case_follow_up"],
      hours: "08:00-18:00",
      status: "open",
      distanceMeters: 2300,
      updatedAt: sampleGeneratedAt,
    },
  ],
};

export const areaPresets = [
  { label: "Accra Metropolitan", lat: 5.6037, lng: -0.187 },
  { label: "Accra flood zone", lat: 5.56, lng: -0.2 },
  { label: "Kumasi area", lat: 6.6885, lng: -1.6244 },
];

export function buildFallbackAlerts(): CitizenAlertFeedItem[] {
  const now = new Date();
  return [
    {
      id: "alert_feed_current_flood",
      title: "Severe flood warning",
      hazardType: "flood",
      severity: "severe_warning",
      message:
        "Heavy rainfall and rising drains may flood low-lying parts of Accra Metro and Tema.",
      target: {
        type: "district",
        ids: ["accra-metropolitan", "tema-metropolitan"],
        label: "Accra Metro and Tema",
      },
      targetLabel: "Accra Metro and Tema",
      startsAt: new Date(now.getTime() - 30 * 60 * 1000).toISOString(),
      expiresAt: new Date(now.getTime() + 5 * 60 * 60 * 1000).toISOString(),
      status: "current",
      recommendedAction:
        "Move away from drains, avoid flooded roads, and prepare to go to a shelter if directed.",
      evacuationRequired: true,
      shelterIds: ["shelter-ama-001", "shelter-osu-002"],
      source: "fixture",
      updatedAt: new Date(now.getTime() - 20 * 60 * 1000).toISOString(),
    },
    {
      id: "alert_feed_current_fire",
      title: "Market fire watch",
      hazardType: "fire",
      severity: "watch",
      message:
        "Responders are monitoring dense market areas after smoke reports near electrical kiosks.",
      target: {
        type: "community",
        ids: ["accra-central"],
        label: "Accra Central",
      },
      targetLabel: "Accra Central",
      startsAt: new Date(now.getTime() - 20 * 60 * 1000).toISOString(),
      expiresAt: new Date(now.getTime() + 3 * 60 * 60 * 1000).toISOString(),
      status: "current",
      recommendedAction:
        "Keep access lanes open, avoid overloaded sockets, and call 112 if you see flames or heavy smoke.",
      evacuationRequired: false,
      shelterIds: [],
      source: "fixture",
      updatedAt: new Date(now.getTime() - 15 * 60 * 1000).toISOString(),
    },
    {
      id: "alert_feed_expired_road",
      title: "Road hazard resolved",
      hazardType: "road_crash",
      severity: "advisory",
      message:
        "Earlier congestion near Kaneshie Market Road has cleared after responders reopened the lane.",
      target: {
        type: "radius",
        ids: ["kaneshie-market-road"],
        label: "Kaneshie Market Road",
        center: { lat: 5.566, lng: -0.242 },
        radiusMeters: 1500,
      },
      targetLabel: "Kaneshie Market Road",
      startsAt: new Date(now.getTime() - 8 * 60 * 60 * 1000).toISOString(),
      expiresAt: new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString(),
      status: "expired",
      recommendedAction:
        "Continue to drive carefully and give way to emergency vehicles.",
      evacuationRequired: false,
      shelterIds: [],
      source: "fixture",
      updatedAt: new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString(),
    },
  ];
}

export const guideHazardOptions: { label: string; value: GuideHazardFilter }[] =
  [
    { label: "All hazards", value: "all" },
    { label: "Flood", value: "flood" },
    { label: "Fire", value: "fire" },
    { label: "Road crash", value: "road_crash" },
    { label: "Electrical", value: "electrical_hazard" },
    { label: "Disease", value: "disease_outbreak" },
    { label: "General", value: "other" },
  ];

export const guideStageOptions: { label: string; value: GuideStageFilter }[] = [
  { label: "All stages", value: "all" },
  { label: "Before", value: "before" },
  { label: "During", value: "during" },
  { label: "After", value: "after" },
  { label: "Recovery", value: "recovery" },
];

export const guideLanguageOptions = [
  { label: "English", value: "en" },
  { label: "Twi", value: "tw" },
  { label: "Ga", value: "ga" },
];

export function buildFallbackGuides(): EmergencyGuideRecord[] {
  const now = new Date().toISOString();
  return [
    {
      id: "guide_flood_before_en",
      hazardType: "flood",
      stage: "before",
      title: "Prepare before flooding",
      body: "Know your nearest shelter, keep documents dry, clear drains safely, prepare drinking water, and agree on a family meeting point.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 10,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "guide_flood_during_en",
      hazardType: "flood",
      stage: "during",
      title: "Stay safe during flooding",
      body: "Move to higher ground, avoid walking or driving through floodwater, turn off electricity only if safe, and call 112 for life-threatening danger.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 20,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "guide_fire_during_en",
      hazardType: "fire",
      stage: "during",
      title: "Fire safety response",
      body: "Leave immediately, warn people nearby, stay low under smoke, never use lifts, and call 112 for Ghana National Fire Service support.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 40,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "guide_evacuation_during_en",
      hazardType: "other",
      stage: "during",
      title: "Safe evacuation",
      body: "Take only essentials, follow official routes, help children and elderly people first, avoid floodwater or smoke, and tell relatives where you are going.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 80,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "guide_112_during_en",
      hazardType: "other",
      stage: "during",
      title: "Calling 112",
      body: "Call 112 for life-threatening emergencies. Share the hazard, exact location, people affected, injuries, and a safe callback number if available.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 110,
      createdAt: now,
      updatedAt: now,
    },
  ];
}

export const hazardOptions: { label: string; value: HazardType }[] = [
  { label: "Flood", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Road crash", value: "road_crash" },
  { label: "Medical emergency", value: "medical_emergency" },
  { label: "Building collapse", value: "building_collapse" },
  { label: "Blocked drain", value: "blocked_drain" },
  { label: "Other", value: "other" },
];

export const urgencyOptions: { label: string; value: IncidentUrgency }[] = [
  { label: "Moderate", value: "moderate" },
  { label: "High", value: "high" },
  { label: "Life threatening", value: "life_threatening" },
  { label: "Low", value: "low" },
];

export const supportedMediaTypes: IncidentMediaContentType[] = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "video/mp4",
  "video/quicktime",
  "audio/mpeg",
  "audio/mp4",
  "audio/wav",
];

export const mediaSizeLimits: Record<IncidentMediaContentType, number> = {
  "image/jpeg": 10 * 1024 * 1024,
  "image/png": 10 * 1024 * 1024,
  "image/webp": 10 * 1024 * 1024,
  "video/mp4": 100 * 1024 * 1024,
  "video/quicktime": 100 * 1024 * 1024,
  "audio/mpeg": 25 * 1024 * 1024,
  "audio/mp4": 25 * 1024 * 1024,
  "audio/wav": 25 * 1024 * 1024,
};

export const initialReportForm: ReportForm = {
  hazard: "flood",
  lat: "5.579",
  lng: "-0.212",
  description: "",
  peopleAffected: "0",
  injuriesReported: false,
  urgency: "moderate",
  anonymous: false,
  contactPermission: true,
  accessibilityNeeds: "",
  files: [],
};
