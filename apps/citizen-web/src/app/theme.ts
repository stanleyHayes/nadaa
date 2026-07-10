import { createNadaaTheme, type ThemeMode } from "@nadaa/brand/theme";

/**
 * Build the citizen theme for a given appearance mode. Called from `CitizenApp`
 * inside a memo so the MUI palette tracks the light/dark toggle alongside the
 * `--nadaa-*` CSS token flip.
 */
export function createCitizenTheme(
  options: { mode?: ThemeMode; reducedMotion?: boolean } = {},
) {
  return createNadaaTheme({ accent: "public", ...options });
}

/** Default light theme, retained for any static consumer. */
export const citizenTheme = createCitizenTheme();
