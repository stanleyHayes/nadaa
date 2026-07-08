/**
 * NADAA design tokens.
 *
 * These tokens are framework-agnostic and can be consumed by MUI themes,
 * plain CSS custom properties, or marketing-web hand-written styles.
 */

export const colors = {
  navy: "#0D1B3D",
  green: "#118D4E",
  red: "#E53935",
  gold: "#F4C20D",
  slate: "#555B66",
  white: "#FFFFFF",
  mist: "#F5F8FC",
  ink: "#101828",
} as const;

export type NadaaColor = keyof typeof colors;

export const semantic = {
  surface: colors.mist,
  surfaceElevated: colors.white,
  border: "#DFEAF2",
  divider: "#E4E7EC",
  textPrimary: colors.ink,
  textSecondary: colors.slate,
  textInverse: colors.white,
  primary: colors.navy,
  secondary: colors.green,
  accent: colors.gold,
  info: "#0B6FB8",
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

export const shadows = {
  none: "none",
  sm: "0 1px 2px rgba(13, 27, 61, 0.06)",
  md: "0 4px 12px rgba(13, 27, 61, 0.08)",
  lg: "0 8px 24px rgba(13, 27, 61, 0.10)",
  xl: "0 18px 48px rgba(13, 27, 61, 0.12)",
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
export const hazardRoles = {
  flood: { background: "#E8F4FC", foreground: "#0B6FB8", border: "#B8DDF3" },
  fire: { background: "#FDECEC", foreground: colors.red, border: "#F5B3B3" },
  medical: {
    background: "#E8F6EE",
    foreground: colors.green,
    border: "#B8E4CC",
  },
  geological: {
    background: "#F3E9DE",
    foreground: "#9A5A23",
    border: "#DDC4AD",
  },
  road: { background: "#F0F1F3", foreground: "#4C5563", border: "#D0D4DA" },
  storm: { background: "#E8F4FC", foreground: "#3E8ED0", border: "#B8DDF3" },
  disease: { background: "#F2EDFD", foreground: "#7C3AED", border: "#D6C7FB" },
  default: {
    background: "#F0F1F3",
    foreground: colors.slate,
    border: "#D0D4DA",
  },
} as const;

export const severityRoles = {
  low: {
    background: "#E8F6EE",
    foreground: colors.green,
    border: "#B8E4CC",
    icon: "CheckCircle2",
  },
  medium: {
    background: "#FEF9E7",
    foreground: "#B98900",
    border: "#F7E28D",
    icon: "AlertTriangle",
  },
  high: {
    background: "#FFF3E0",
    foreground: "#E65100",
    border: "#FFCC80",
    icon: "AlertTriangle",
  },
  severe: {
    background: "#FDECEC",
    foreground: colors.red,
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
