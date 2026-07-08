import { nativeTheme } from "@nadaa/brand/native";

/**
 * Convert a hex color to an rgba string.
 *
 * Keeps alpha values derived from the brand palette instead of hard-coding
 * translucent values in screens and components.
 */
export function hexToRgba(hex: string, alpha: number): string {
  const sanitized = hex.replace("#", "");
  const full =
    sanitized.length === 3
      ? sanitized
          .split("")
          .map((char) => char + char)
          .join("")
      : sanitized;
  const bigint = Number.parseInt(full, 16);
  const r = (bigint >> 16) & 255;
  const g = (bigint >> 8) & 255;
  const b = bigint & 255;
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

/**
 * NADAA citizen-mobile theme.
 *
 * Re-exports and extends the shared React Native theme from `@nadaa/brand/native`
 * so the mobile app consumes the same design tokens as the web apps. Legacy
 * aliases such as `navy`, `ink`, and `muted` are preserved so existing imports
 * keep working.
 */
export const mobileTheme = {
  ...nativeTheme,
  colors: {
    ...nativeTheme.colors,
    background: nativeTheme.colors.background,
    border: hexToRgba(nativeTheme.colors.primary, 0.12),
    card: nativeTheme.colors.surface,
    danger: nativeTheme.colors.danger,
    gold: nativeTheme.colors.accent,
    green: nativeTheme.colors.success,
    ink: nativeTheme.colors.textPrimary,
    muted: nativeTheme.colors.textSecondary,
    navy: nativeTheme.colors.primary,
    softBlue: nativeTheme.colors.softBlue,
    softGreen: nativeTheme.colors.softGreen,
    softRed: nativeTheme.colors.softRed,
    white: nativeTheme.colors.surface,
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
} as const;

export type MobileTheme = typeof mobileTheme;

export {
  hazardBadgeFor,
  nativeTheme,
  severityBadgeFor,
  type NativeTheme,
} from "@nadaa/brand/native";
