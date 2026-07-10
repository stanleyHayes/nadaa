import { useMemo, type ReactNode } from "react";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { useAgencySession } from "@/app/session";
import { createAgencyTheme } from "@/app/theme";
import { useThemeMode } from "@/app/theme-mode";

/**
 * Reactive MUI theme boundary for the agency desk. The agency app mounts its
 * `ThemeProvider` here (above `AgencyApp`) rather than inside the feature root,
 * so this wrapper owns the reactive theme build: it rebuilds the MUI palette
 * whenever the appearance mode (Sun/Moon toggle) or the operator's
 * reduced-motion preference changes, keeping dialogs, menus, and inputs in step
 * with the `--nadaa-*` CSS token flip.
 */
export function AgencyThemeProvider({ children }: { children: ReactNode }) {
  const mode = useThemeMode();
  const { preferences } = useAgencySession();

  const theme = useMemo(
    () => createAgencyTheme({ mode, reducedMotion: preferences.reducedMotion }),
    [mode, preferences.reducedMotion],
  );

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  );
}
