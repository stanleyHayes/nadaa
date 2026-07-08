import { nativeTheme } from "@nadaa/brand/native";

/**
 * Dispatcher mobile theme.
 *
 * Re-exports the canonical React Native theme from `@nadaa/brand/native` and
 * layers a small set of app-specific aliases so existing imports continue to
 * work. Prefer the `nativeTheme` tokens for new code.
 */
export const mobileTheme = {
  ...nativeTheme,
  colors: {
    ...nativeTheme.colors,
    // Backward-compatible aliases used across dispatcher-mobile screens/components.
    background: nativeTheme.colors.background,
    border: nativeTheme.colors.border,
    card: nativeTheme.colors.surface,
    danger: nativeTheme.colors.danger,
    gold: nativeTheme.colors.warning,
    green: nativeTheme.colors.success,
    ink: nativeTheme.colors.textPrimary,
    muted: nativeTheme.colors.textSecondary,
    navy: nativeTheme.colors.primary,
    softBlue: nativeTheme.colors.softBlue,
    softGreen: nativeTheme.colors.softGreen,
    softRed: nativeTheme.colors.softRed,
    white: nativeTheme.colors.surface,
  },
} as const;

export type MobileTheme = typeof mobileTheme;

/** Convert a 6-digit hex color from the theme to an RGBA string. */
export function withAlpha(color: string, alpha: number): string {
  const hex = color.replace("#", "");
  const r = parseInt(hex.slice(0, 2), 16);
  const g = parseInt(hex.slice(2, 4), 16);
  const b = parseInt(hex.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}
