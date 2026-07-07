import { nadaaBrand } from "@nadaa/brand";

export const mobileTheme = {
  colors: {
    background: "#F4F7FB",
    border: "rgba(13, 27, 61, 0.12)",
    card: nadaaBrand.colors.white,
    danger: nadaaBrand.colors.red,
    gold: nadaaBrand.colors.gold,
    green: nadaaBrand.colors.green,
    ink: nadaaBrand.colors.ink,
    muted: nadaaBrand.colors.slate,
    navy: nadaaBrand.colors.navy,
    softBlue: "#EEF5FF",
    softGreen: "#F0FAF4",
    softRed: "#FEF0EF",
    white: nadaaBrand.colors.white,
  },
  font: {
    family: "Outfit",
    regular: "Outfit_400Regular",
    medium: "Outfit_500Medium",
    semibold: "Outfit_600SemiBold",
    bold: "Outfit_800ExtraBold",
  },
  radius: {
    sm: 6,
    md: 8,
  },
  spacing: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
  },
} as const;

export type MobileTheme = typeof mobileTheme;
