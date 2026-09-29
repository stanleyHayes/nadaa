import { ACTIVE_BRAND, brand, type BrandId } from "./identity.js";
import { palette } from "./palettes.js";

/**
 * Legacy alias for the active brand identity.
 *
 * Kept under this name because 78 modules import it. New code should prefer
 * `brand` from identity.js and `palette` from palettes.js, which say what they
 * are; this object stays as the compatibility surface and now resolves through
 * whichever identity ACTIVE_BRAND selects rather than hard-coding NADAA.
 *
 * `slogan` is retained as the historical key name; identity.ts calls the same
 * value `tagline`.
 */
/**
 * What each palette slot signifies, per brand. Rendered on the brand sheet, so
 * it has to track whichever palette is active: the same `navy` slot means
 * "Authority & Focus" as Onyx black and "Foundation & Security" as SUBVENIO's
 * deep navy, and `gold` shifts from optimism to SUBVENIO's human urgency.
 */
const MEANINGS: Record<BrandId, Record<string, string>> = {
  nadaa: {
    navy: "Authority & Focus",
    charcoal: "Depth & Structure",
    silver: "Clarity & Precision",
    green: "Safety & Growth",
    red: "Alert & Urgency",
    gold: "Hope & Optimism",
    slate: "Stability & Strength",
  },
  subvenio: {
    navy: "Foundation & Security",
    charcoal: "Depth & Structure",
    silver: "Support & Calm",
    green: "Confirmed Safe",
    red: "Critical Danger",
    gold: "Urgency & Human Action",
    slate: "Clarity & Metadata",
  },
};

const colourMeanings = MEANINGS[ACTIVE_BRAND];

export const nadaaBrand = {
  name: brand.name,
  fullName: brand.fullName,
  country: brand.country,
  slogan: brand.tagline,
  supportLine: brand.supportLine,
  assets: brand.assets,
  colors: {
    navy: palette.navy,
    charcoal: palette.charcoal,
    silver: palette.silver,
    lightSilver: palette.lightSilver,
    green: palette.green,
    red: palette.red,
    gold: palette.gold,
    slate: palette.slate,
    white: palette.white,
    mist: palette.mist,
    ink: palette.ink,
  },
  meanings: colourMeanings,
} as const;

export const featurePillars = [
  {
    title: "Know Your Risk",
    description: "Check risks in any area anytime",
    accent: nadaaBrand.colors.green,
  },
  {
    title: "Get Alerts",
    description: "Receive timely warnings that save lives",
    accent: nadaaBrand.colors.red,
  },
  {
    title: "Report Incidents",
    description: "Report disasters and accidents",
    accent: nadaaBrand.colors.gold,
  },
  {
    title: "Get Help",
    description: "Connect to emergency services fast",
    accent: "#42A5F5",
  },
  {
    title: "Stay Informed",
    description: "Learn how to prepare, respond, and recover",
    accent: nadaaBrand.colors.green,
  },
  {
    title: "Stronger Together",
    description: "Building resilient communities",
    accent: nadaaBrand.colors.red,
  },
] as const;

export const hazardPalette = {
  flood: "#0B6FB8",
  fire: nadaaBrand.colors.red,
  medical: nadaaBrand.colors.green,
  geological: "#9A5A23",
  road: "#4C5563",
  storm: "#3E8ED0",
  disease: "#7C3AED",
} as const;
