/**
 * React Native compatible NADAA theme values.
 *
 * Import from `@nadaa/brand/native` in Expo/React Native apps.
 */

import {
  colors,
  semantic,
  spacing,
  typography,
  shadows,
  radii,
  hazardRoles,
  severityRoles,
} from "./tokens.js";

export const nativeTheme = {
  colors: {
    primary: colors.navy,
    secondary: colors.green,
    accent: colors.gold,
    danger: colors.red,
    warning: colors.gold,
    success: colors.green,
    info: semantic.info,
    background: semantic.surface,
    surface: semantic.surfaceElevated,
    border: semantic.border,
    divider: semantic.divider,
    textPrimary: semantic.textPrimary,
    textSecondary: semantic.textSecondary,
    textInverse: semantic.textInverse,
    // Soft tints for cards/badges
    softBlue: hazardRoles.flood.background,
    softGreen: hazardRoles.medical.background,
    softRed: hazardRoles.fire.background,
    softGold: severityRoles.medium.background,
  },
  font: {
    family: "Outfit",
    regular: "Outfit_400Regular",
    medium: "Outfit_500Medium",
    semibold: "Outfit_600SemiBold",
    bold: "Outfit_800ExtraBold",
  },
  spacing: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
    "2xl": 32,
    "3xl": 48,
  },
  radius: {
    sm: 4,
    md: 8,
    lg: 12,
    xl: 16,
  },
  shadows: {
    sm: shadows.sm,
    md: shadows.md,
    lg: shadows.lg,
    xl: shadows.xl,
  },
} as const;

export type NativeTheme = typeof nativeTheme;

/** Accessible severity badge styles for React Native. */
export function severityBadgeFor(severity: string) {
  const key = severity.toLowerCase() as keyof typeof severityRoles;
  const role = severityRoles[key] ?? severityRoles.info;
  return {
    background: role.background,
    color: role.foreground,
    border: role.border,
    icon: role.icon,
  };
}

/** Accessible hazard badge styles for React Native. */
export function hazardBadgeFor(hazard: string) {
  const key = hazard.toLowerCase() as keyof typeof hazardRoles;
  const role = (hazardRoles[key] ??
    hazardRoles.default) as (typeof hazardRoles)["flood"];
  return {
    background: role.background,
    color: role.foreground,
    border: role.border,
  };
}
