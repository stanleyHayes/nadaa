import { nadaaBrand } from "@nadaa/brand";
import type {
  AreaRiskResponse,
  EmergencyGuideRecord,
  HazardType,
  NearbyShelterResponse,
} from "@nadaa/shared-types";
import type {
  MobilePermissionState,
  MobileSession,
  ReportDraft,
  SignInDraft,
} from "./types";

const generatedAt = new Date().toISOString();

/** Placeholder phone for anonymous guest sessions — never sent as a real contact. */
export const GUEST_PLACEHOLDER_PHONE = "+233200000000";

export const initialSession: MobileSession = {
  contactPermission: true,
  isGuest: true,
  name: "Guest citizen",
  phone: GUEST_PLACEHOLDER_PHONE,
  preferredLanguage: "en",
  userId: "usr_mobile_guest",
};

export const initialSignIn: SignInDraft = {
  name: "",
  otp: "",
  phone: "",
};

export const emptyShelters: NearbyShelterResponse = {
  generatedAt: "",
  recoverySupport: [],
  shelters: [],
};

export const initialPermissions: MobilePermissionState = {
  camera: "unknown",
  location: "unknown",
  media: "unknown",
  push: "unknown",
};

export const initialReportDraft: ReportDraft = {
  anonymous: false,
  contactPermission: true,
  description: "",
  hazard: "flood",
  injuriesReported: false,
  lat: "5.603700",
  lng: "-0.187000",
  mediaRefs: [],
  peopleAffected: "0",
  urgency: "moderate",
};

export const mobileAreaPresets = [
  { label: "Accra Metropolitan", lat: 5.6037, lng: -0.187 },
  { label: "Accra flood zone", lat: 5.56, lng: -0.2 },
  { label: "Kumasi area", lat: 6.6885, lng: -1.6244 },
];

export const hazardOptions: Array<{ label: string; value: HazardType }> = [
  { label: "Flood", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Medical emergency", value: "medical_emergency" },
  { label: "Road crash", value: "road_crash" },
  { label: "Building collapse", value: "building_collapse" },
  { label: "Other", value: "other" },
];

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
      contact: nadaaBrand.supportLine,
    },
  ],
  nearbyFacilities: [
    {
      id: "agency-nadmo-ama",
      name: "NADMO Accra Metro",
      type: "nadmo",
      location: { lat: 5.56, lng: -0.2 },
      contact: nadaaBrand.supportLine,
    },
  ],
  recommendedActions: [
    "Avoid low-lying roads and open drains.",
    "Move valuables above ground level.",
    "Prepare an evacuation route to the nearest safe shelter.",
  ],
};

export function buildFallbackGuides(): EmergencyGuideRecord[] {
  return [
    {
      id: "guide_flood_during_en",
      hazardType: "flood",
      stage: "during",
      title: "During flooding",
      body: "Move to higher ground. Do not walk or drive through floodwater. Call 112 if anyone is trapped or injured.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 10,
      createdAt: generatedAt,
      updatedAt: generatedAt,
    },
    {
      id: "guide_fire_during_en",
      hazardType: "fire",
      stage: "during",
      title: "During a fire",
      body: "Leave the area quickly. Keep access routes open for responders. Call 112 if you see flames or heavy smoke.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 20,
      createdAt: generatedAt,
      updatedAt: generatedAt,
    },
    {
      id: "guide_flood_before_tw",
      hazardType: "flood",
      stage: "before",
      title: "Flood readiness",
      body: "Keep documents dry, charge your phone, and know your nearest shelter before heavy rain.",
      language: "tw",
      offlineAvailable: true,
      sortOrder: 30,
      createdAt: generatedAt,
      updatedAt: generatedAt,
    },
  ];
}
