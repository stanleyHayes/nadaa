import { createNadaaTheme, type ThemeMode } from "@nadaa/brand/theme";

/**
 * Build the agency theme for a given appearance mode. Called from the reactive
 * `AgencyThemeProvider` inside a memo so the MUI palette tracks the light/dark
 * toggle alongside the `--nadaa-*` CSS token flip.
 */
export function createAgencyTheme(options: {
  mode?: ThemeMode;
  reducedMotion?: boolean;
} = {}) {
  return createNadaaTheme({ accent: "operational", ...options });
}

/** Default light theme, retained for any static consumer. */
export const agencyTheme = createAgencyTheme();
