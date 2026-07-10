import type {
  HazardType,
  IncidentMediaContentType,
  IncidentUrgency,
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

export const areaPresets = [
  { label: "Accra Metropolitan", lat: 5.6037, lng: -0.187 },
  { label: "Accra flood zone", lat: 5.56, lng: -0.2 },
  { label: "Kumasi area", lat: 6.6885, lng: -1.6244 },
];

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
