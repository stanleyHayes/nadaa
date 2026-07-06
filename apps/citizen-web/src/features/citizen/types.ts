import type {
  EmergencyGuideRecord,
  GuideStage,
  HazardType,
  IncidentUrgency,
} from "@nadaa/shared-types";

export type ReportForm = {
  hazard: HazardType;
  lat: string;
  lng: string;
  description: string;
  peopleAffected: string;
  injuriesReported: boolean;
  urgency: IncidentUrgency;
  anonymous: boolean;
  contactPermission: boolean;
  accessibilityNeeds: string;
  files: File[];
};

export type ReportState =
  | { status: "idle" }
  | { status: "loading"; message: string }
  | { status: "success"; reference: string; priorityReview: boolean }
  | { status: "error"; message: string };

export type RiskState =
  | { status: "idle"; message?: string }
  | { status: "loading"; message: string }
  | { status: "error"; message: string }
  | { status: "permission-denied"; message: string };

export type AlertFeedView = "current" | "expired" | "all";

export type AlertFeedState =
  | { status: "idle"; message?: string }
  | { status: "loading"; message: string }
  | { status: "error"; message: string };

export type GuideHazardFilter = "all" | HazardType;
export type GuideStageFilter = "all" | GuideStage;

export type GuideFilters = {
  hazard: GuideHazardFilter;
  stage: GuideStageFilter;
  language: string;
};

export type GuideState =
  | { status: "idle"; message?: string }
  | { status: "loading"; message: string }
  | { status: "offline"; message: string }
  | { status: "error"; message: string };

export type GuideCachePayload = {
  guides: EmergencyGuideRecord[];
  cachedAt: string;
  language: string;
};

export type GuideCacheInfo = {
  cachedAt: string;
  source: "cache" | "fixture" | "network";
  language: string;
};
