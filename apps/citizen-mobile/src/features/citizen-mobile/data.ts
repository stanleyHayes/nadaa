import { nadaaBrand } from "@nadaa/brand";
import type {
  AreaRiskResponse,
  CitizenAlertFeedItem,
  EmergencyGuideRecord,
  HazardType,
  NearbyShelterResponse,
  VolunteerProfile,
  VolunteerTaskRecord,
} from "@nadaa/shared-types";
import type {
  MobilePermissionState,
  MobileSession,
  ReportDraft,
} from "./types";

const generatedAt = new Date().toISOString();

export const initialSession: MobileSession = {
  contactPermission: true,
  isGuest: true,
  name: "Guest citizen",
  phone: "+233200000000",
  preferredLanguage: "en",
  userId: "usr_mobile_guest",
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

export const sampleVolunteerProfile: VolunteerProfile = {
  availabilityStatus: "available",
  citizenUserId: "usr_mobile_guest",
  community: "Jamestown",
  createdAt: generatedAt,
  district: "Accra Metropolitan",
  groupId: "grp_greater-accra_accra-metropolitan_jamestown",
  id: "vol_mobile_guest",
  languages: ["en", "tw"],
  name: "Guest volunteer",
  phone: "+233200000000",
  region: "Greater Accra",
  safetyNotes: [
    "Stay in public, safe areas and never enter floodwater, fire zones, collapsed structures, or violent scenes.",
    "Call 112 and request authority escalation for injuries, trapped people, unsafe crowds, or blocked emergency access.",
    "Share observations only when doing so does not delay evacuation or personal safety.",
  ],
  skills: ["first aid", "community alerts"],
  updatedAt: generatedAt,
  verificationStatus: "verified",
  verifiedAt: generatedAt,
  verifiedBy: "usr_demo_district_officer",
};

export const sampleVolunteerTasks: VolunteerTaskRecord[] = [
  {
    assignedAt: generatedAt,
    assignedBy: "usr_demo_dispatcher",
    escalationRequired: false,
    groupId: sampleVolunteerProfile.groupId,
    id: "vtask_mobile_001",
    incidentId: "inc_mobile_community_001",
    incidentReference: "INC-000142",
    instructions:
      "Check whether households near the shelter approach need water, transport, or accessible support. Stay on the public road.",
    locationLabel: "Jamestown shelter approach",
    priority: "high",
    safetyRules: sampleVolunteerProfile.safetyNotes,
    status: "assigned",
    type: "welfare_check",
    updatedAt: generatedAt,
    updates: [],
    volunteerId: sampleVolunteerProfile.id,
    volunteerName: sampleVolunteerProfile.name,
  },
  {
    assignedAt: generatedAt,
    assignedBy: "usr_demo_dispatcher",
    escalationRequired: false,
    groupId: sampleVolunteerProfile.groupId,
    id: "vtask_mobile_002",
    incidentId: "inc_mobile_community_002",
    incidentReference: "INC-000143",
    instructions:
      "Share approved shelter information with households near the community centre and report blocked access routes.",
    locationLabel: "Jamestown community centre",
    priority: "normal",
    safetyRules: sampleVolunteerProfile.safetyNotes,
    status: "accepted",
    type: "community_alerting",
    updatedAt: generatedAt,
    updates: [
      {
        createdAt: generatedAt,
        createdBy: sampleVolunteerProfile.id,
        escalationRequested: false,
        id: "vtup_mobile_001",
        note: "Accepted and moving toward the community centre.",
        safetyStatus: "safe",
        status: "accepted",
        type: "status",
      },
    ],
    volunteerId: sampleVolunteerProfile.id,
    volunteerName: sampleVolunteerProfile.name,
  },
];

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

export const sampleShelters: NearbyShelterResponse = {
  generatedAt,
  recoverySupport: [
    {
      id: "recovery_ama_relief_001",
      name: "AMA Relief Distribution Point",
      type: "relief_point",
      region: "Greater Accra",
      district: "Accra Metropolitan",
      address: "Independence Avenue recovery desk",
      location: { lat: 5.558, lng: -0.197 },
      contact: nadaaBrand.supportLine,
      services: ["food", "water", "blankets", "family_reunification"],
      hours: "08:00-20:00",
      status: "open",
      distanceMeters: 420,
      updatedAt: generatedAt,
    },
  ],
  shelters: [
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
      distanceMeters: 0,
      updatedAt: generatedAt,
    },
    {
      id: "shelter-osu-002",
      name: "Osu Community Hall",
      type: "temporary_shelter",
      region: "Greater Accra",
      district: "Korle Klottey",
      address: "Osu Community Hall",
      location: { lat: 5.55, lng: -0.18 },
      capacity: 220,
      currentOccupancy: 34,
      status: "open",
      contact: nadaaBrand.supportLine,
      facilities: ["water", "first_aid", "family_area"],
      notes: "Short-term shelter and reunification point.",
      distanceMeters: 2480,
      updatedAt: generatedAt,
    },
  ],
};

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
  ];
}

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
