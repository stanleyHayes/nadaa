import { useEffect, useMemo } from "react";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { hasCommandAccess, useAuthoritySession } from "@/app/session";
import { createAuthorityTheme } from "@/app/theme";
import { useThemeMode } from "@/app/theme-mode";
import { CommandCenterShell } from "./CommandCenterShell";
import { SignInScreen } from "./components/SignInScreen";

/**
 * Feature root. Provides the theme and gates the entire command center behind
 * an authenticated authority session. Until an operator signs in and completes
 * MFA, only the sign-in screen renders; the shell (sidebar + views) mounts and
 * loads its data once access is granted.
 */
function CommandCenterApp() {
  const { session, preferences } = useAuthoritySession();
  const authorized = hasCommandAccess(session);
  const mode = useThemeMode();

  // Rebuild the MUI theme whenever the appearance mode or reduced-motion
  // preference changes, so dialogs/menus/inputs match the CSS token flip.
  const theme = useMemo(
    () =>
      createAuthorityTheme({ mode, reducedMotion: preferences.reducedMotion }),
    [mode, preferences.reducedMotion],
  );

  // Apply the operator's reduced-motion preference across the whole app.
  useEffect(() => {
    const root = document.documentElement;
    if (preferences.reducedMotion) {
      root.setAttribute("data-nadaa-reduced-motion", "reduce");
    } else {
      root.removeAttribute("data-nadaa-reduced-motion");
    }
  }, [preferences.reducedMotion]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {authorized && session ? (
        <CommandCenterShell session={session} />
      ) : (
        <>
          <a href="#main-content" className="skip-link">
            Skip to main content
          </a>
          <SignInScreen />
        </>
      )}
    </ThemeProvider>
  );
}

export default CommandCenterApp;
