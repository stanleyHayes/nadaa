/**
 * NADAA design tokens.
 *
 * These tokens are framework-agnostic and can be consumed by MUI themes,
 * plain CSS custom properties, or marketing-web hand-written styles.
 */

import { palette } from "./palettes.js";

/**
 * Raw palette for the active brand.
 *
 * The keys are historical slots, not colour descriptions — `navy` is the
 * brand-ink slot and has held a navy, a near-black and now SUBVENIO's deep
 * navy. They keep these names because the 78 importers and ~5,000
 * `--nadaa-*` consumers resolve through them; palettes.ts re-points the
 * values per brand instead.
 */
export const colors = {
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
} as const;

export type NadaaColor = keyof typeof colors;

export const semantic = {
  surface: colors.mist,
  surfaceElevated: colors.white,
  border: palette.border,
  divider: palette.divider,
  textPrimary: colors.ink,
  textSecondary: colors.slate,
  textInverse: colors.white,
  primary: colors.navy,
  // Functional safety colors are preserved through the rebrand: green (safety),
  // gold (warning/accent — pops on black+silver), red (danger), flood-blue info.
  secondary: colors.green,
  accent: colors.gold,
  info: palette.info,
  success: colors.green,
  warning: colors.gold,
  danger: colors.red,
} as const;

export const spacing = {
  0: "0px",
  0.5: "2px",
  1: "4px",
  1.5: "6px",
  2: "8px",
  2.5: "10px",
  3: "12px",
  4: "16px",
  5: "20px",
  6: "24px",
  7: "28px",
  8: "32px",
  9: "36px",
  10: "40px",
  12: "48px",
  14: "56px",
  16: "64px",
  20: "80px",
  24: "96px",
} as const;

export const typography = {
  fontFamily: '"Outfit", "Helvetica Neue", Arial, sans-serif',
  weights: {
    regular: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
    extrabold: 800,
  },
  sizes: {
    xs: { fontSize: "0.75rem", lineHeight: 1.5 },
    sm: { fontSize: "0.875rem", lineHeight: 1.5 },
    base: { fontSize: "1rem", lineHeight: 1.5 },
    lg: { fontSize: "1.125rem", lineHeight: 1.4 },
    xl: { fontSize: "1.25rem", lineHeight: 1.3 },
    "2xl": { fontSize: "1.5rem", lineHeight: 1.25 },
    "3xl": { fontSize: "2rem", lineHeight: 1.2 },
    "4xl": { fontSize: "2.5rem", lineHeight: 1.15 },
  },
} as const;

/**
 * Elevation. Tinted with the brand ink so shadows sit in the palette's family
 * rather than reading as neutral grey over a coloured surface.
 */
export const shadows = {
  none: "none",
  sm: `0 1px 2px color-mix(in srgb, ${palette.ink} 6%, transparent)`,
  md: `0 4px 12px color-mix(in srgb, ${palette.ink} 8%, transparent)`,
  lg: `0 8px 24px color-mix(in srgb, ${palette.ink} 10%, transparent)`,
  xl: `0 18px 48px color-mix(in srgb, ${palette.ink} 12%, transparent)`,
} as const;

export const breakpoints = {
  values: {
    xs: 0,
    sm: 600,
    md: 900,
    lg: 1200,
    xl: 1536,
  },
} as const;

export const radii = {
  none: "0px",
  sm: "4px",
  md: "8px",
  lg: "12px",
  xl: "16px",
  full: "9999px",
} as const;

/**
 * Hazard and severity color pairs with accessible foreground colors.
 * Background + foreground combinations must maintain WCAG 2.1 AA contrast.
 */
/**
 * Hazard and severity chips.
 *
 * These are SEMANTIC, not brand: a flood reads blue and a fire reads red under
 * any identity, which is why they do not follow ACTIVE_BRAND. The guide is
 * explicit that red in particular is semantic and never the master brand
 * colour.
 *
 * The foregrounds are the chip's LABEL TEXT, so each one is picked to clear
 * 4.5:1 against its own background rather than inherited from the brand
 * palette. They used to reference colors.red / colors.green / colors.slate,
 * which quietly broke when the palette moved: the backgrounds are fixed light
 * tints computed for the previous foregrounds, so a lighter brand green or
 * slate dropped the pair below AA. Measured before this change: medical and
 * low 2.91:1, medium 3.00:1, fire and severe 3.43:1, high 3.46:1, storm
 * 3.14:1, default 4.21:1.
 *
 * If you change a background here, re-check the pair. Severity must also be
 * carried by label and icon, never by colour alone.
 */
export const hazardRoles = {
  flood: { background: "#E8F4FC", foreground: "#0B6FB8", border: "#B8DDF3" },
  fire: { background: "#FDECEC", foreground: "#D31E24", border: "#F5B3B3" },
  medical: {
    background: "#E8F6EE",
    foreground: "#117C51",
    border: "#B8E4CC",
  },
  geological: {
    background: "#F3E9DE",
    foreground: "#9A5A23",
    border: "#DDC4AD",
  },
  road: { background: "#F0F1F3", foreground: "#4C5563", border: "#D0D4DA" },
  storm: { background: "#E8F4FC", foreground: "#2A71AC", border: "#B8DDF3" },
  disease: { background: "#F2EDFD", foreground: "#7C3AED", border: "#D6C7FB" },
  default: {
    background: "#F0F1F3",
    foreground: "#5E6D83",
    border: "#D0D4DA",
  },
} as const;

/** Severity chips. Same rule as hazardRoles above: semantic, and each
 *  foreground clears 4.5:1 on its own background as label text. */
export const severityRoles = {
  low: {
    background: "#E8F6EE",
    foreground: "#117C51",
    border: "#B8E4CC",
    icon: "CheckCircle2",
  },
  medium: {
    background: "#FEF9E7",
    foreground: "#906B00",
    border: "#F7E28D",
    icon: "AlertTriangle",
  },
  high: {
    background: "#FFF3E0",
    foreground: "#C14400",
    border: "#FFCC80",
    icon: "AlertTriangle",
  },
  severe: {
    background: "#FDECEC",
    foreground: "#D31E24",
    border: "#F5B3B3",
    icon: "AlertOctagon",
  },
  info: {
    background: "#E8F4FC",
    foreground: "#0B6FB8",
    border: "#B8DDF3",
    icon: "Info",
  },
} as const;

export type Severity = keyof typeof severityRoles;
export type Hazard = keyof typeof hazardRoles;

/** Public vs operational app accent colors. */
export const appAccent = {
  public: colors.gold,
  operational: colors.green,
} as const;
