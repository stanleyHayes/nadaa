import type {
  CitizenAlertFeedItem,
  EmergencyGuideRecord,
  HazardType,
  IncidentUrgency,
} from "@nadaa/shared-types";

export type MobileSession = {
  accessToken?: string;
  contactPermission: boolean;
  isGuest: boolean;
  name: string;
  phone: string;
  preferredLanguage: string;
  userId: string;
};

export type PermissionStatus = "unknown" | "granted" | "denied" | "blocked";

export type MobilePermissionState = {
  camera: PermissionStatus;
  location: PermissionStatus;
  media: PermissionStatus;
  push: PermissionStatus;
};

export type ReportDraft = {
  anonymous: boolean;
  contactPermission: boolean;
  description: string;
  hazard: HazardType;
  injuriesReported: boolean;
  lat: string;
  lng: string;
  mediaRefs: string[];
  peopleAffected: string;
  savedAt?: string;
  urgency: IncidentUrgency;
};

export type GuideCachePayload = {
  cachedAt: string;
  guides: EmergencyGuideRecord[];
  language: string;
};

export type MobileLoadState =
  | { status: "idle"; message?: string }
  | { status: "loading"; message: string }
  | { status: "offline"; message: string }
  | { status: "error"; message: string }
  | { status: "success"; message: string };

export type PushRegistrationState =
  | { status: "not_configured"; message: string }
  | { status: "permission_needed"; message: string }
  | { status: "registered"; provider: string; token: string }
  | { status: "failed"; message: string };

export type AlertView = "current" | "expired" | "all";

export type VolunteerObservationDraft = {
  escalationRequested: boolean;
  note: string;
  safetyStatus: "safe" | "caution" | "unsafe" | "needs_authority";
};

export type CitizenMobileSnapshot = {
  alertFeed: CitizenAlertFeedItem[];
  guideCache: GuideCachePayload;
  permissions: MobilePermissionState;
  reportDraft: ReportDraft;
  session: MobileSession;
};
